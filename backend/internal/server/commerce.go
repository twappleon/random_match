package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"

	"random-match/backend/internal/model"
)

const (
	dailyFreeMatchLimit = 10
	premiumMonthlyPlan  = "premium_monthly"
)

var reserveMatchAttemptScript = redis.NewScript(`
local current = tonumber(redis.call('GET', KEYS[1]) or '0')
local limit = tonumber(ARGV[1])
if current >= limit then
  return {0, current}
end
current = redis.call('INCR', KEYS[1])
redis.call('EXPIRE', KEYS[1], tonumber(ARGV[2]))
return {1, current}
`)

type matchUsage struct {
	DailyLimit     int `json:"dailyLimit"`
	DailyUsed      int `json:"dailyUsed"`
	DailyRemaining int `json:"dailyRemaining"`
}

func (s *Server) commerceStatus(c *gin.Context) {
	userID := userIDFromContext(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	status, err := s.commerceStatusForUser(ctx, userID)
	if err != nil {
		log.Printf("commerce status failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "commerce status failed"})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (s *Server) commerceStatusForUser(ctx context.Context, userID string) (commerceStatusResponse, error) {
	user, err := s.userByID(ctx, userID)
	if err != nil {
		return commerceStatusResponse{}, err
	}
	member := isActiveMember(user, time.Now().UTC())
	usage, err := s.matchUsage(ctx, userID, member)
	if err != nil {
		return commerceStatusResponse{}, err
	}
	return commerceStatusResponse{
		IsMember:            member,
		MembershipPlan:      user.MembershipPlan,
		MembershipExpiresAt: user.MembershipExpiresAt,
		DailyLimit:          usage.DailyLimit,
		DailyUsed:           usage.DailyUsed,
		DailyRemaining:      usage.DailyRemaining,
		PriorityQueue:       member,
		GemsBalance:         user.GemsBalance,
	}, nil
}

func (s *Server) createPaymentOrder(c *gin.Context) {
	userID := userIDFromContext(c)
	var req createPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if req.Plan == "" {
		req.Plan = premiumMonthlyPlan
	}
	if req.Plan != premiumMonthlyPlan {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported plan"})
		return
	}

	order := model.PaymentOrder{
		ID:        newID(),
		UserID:    userID,
		Plan:      req.Plan,
		Amount:    699,
		Currency:  "USD",
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if _, err := s.db.DB.Collection("payment_orders").InsertOne(ctx, order); err != nil {
		log.Printf("payment order create failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create payment order failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"order": order})
}

func (s *Server) confirmPaymentOrder(c *gin.Context) {
	userID := userIDFromContext(c)
	orderID := c.Param("id")
	now := time.Now().UTC()

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var order model.PaymentOrder
	if err := s.db.DB.Collection("payment_orders").FindOne(ctx, bson.M{"_id": orderID, "userId": userID}).Decode(&order); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	if order.Status == "paid" {
		c.JSON(http.StatusOK, gin.H{"order": order})
		return
	}
	if order.Plan != premiumMonthlyPlan {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported plan"})
		return
	}

	user, err := s.userByID(ctx, userID)
	if err != nil {
		log.Printf("payment user load failed user_id=%s order_id=%s err=%v", userID, orderID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "confirm payment failed"})
		return
	}
	startAt := now
	if user.MembershipExpiresAt != nil && user.MembershipExpiresAt.After(now) {
		startAt = *user.MembershipExpiresAt
	}
	expiresAt := startAt.AddDate(0, 1, 0)

	_, err = s.db.DB.Collection("payment_orders").UpdateOne(ctx, bson.M{"_id": order.ID, "status": "pending"}, bson.M{"$set": bson.M{
		"status": "paid",
		"paidAt": now,
	}})
	if err != nil {
		log.Printf("payment order update failed user_id=%s order_id=%s err=%v", userID, orderID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "confirm payment failed"})
		return
	}
	_, err = s.db.DB.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{
		"membershipPlan":      premiumMonthlyPlan,
		"membershipExpiresAt": expiresAt,
		"updatedAt":           now,
	}, "$inc": bson.M{
		"gemsBalance": 300,
	}})
	if err != nil {
		log.Printf("payment membership update failed user_id=%s order_id=%s err=%v", userID, orderID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "confirm payment failed"})
		return
	}
	order.Status = "paid"
	order.PaidAt = &now
	c.JSON(http.StatusOK, gin.H{"order": order})
}

func (s *Server) userByID(ctx context.Context, userID string) (model.User, error) {
	var user model.User
	err := s.db.DB.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	return user, err
}

func isActiveMember(user model.User, now time.Time) bool {
	return user.MembershipExpiresAt != nil && user.MembershipExpiresAt.After(now)
}

func (s *Server) matchUsage(ctx context.Context, userID string, member bool) (matchUsage, error) {
	if member {
		return matchUsage{DailyLimit: 0, DailyUsed: 0, DailyRemaining: -1}, nil
	}
	used, err := s.cache.Client.Get(ctx, dailyQuotaKey(userID)).Result()
	if err == redis.Nil {
		return matchUsage{DailyLimit: dailyFreeMatchLimit, DailyUsed: 0, DailyRemaining: dailyFreeMatchLimit}, nil
	}
	if err != nil {
		return matchUsage{}, err
	}
	value, _ := strconv.Atoi(used)
	remaining := dailyFreeMatchLimit - value
	if remaining < 0 {
		remaining = 0
	}
	return matchUsage{DailyLimit: dailyFreeMatchLimit, DailyUsed: value, DailyRemaining: remaining}, nil
}

func (s *Server) reserveMatchAttempt(ctx context.Context, userID string, member bool) (matchUsage, bool, error) {
	if member {
		return matchUsage{DailyLimit: 0, DailyUsed: 0, DailyRemaining: -1}, true, nil
	}
	raw, err := reserveMatchAttemptScript.Run(ctx, s.cache.Client, []string{dailyQuotaKey(userID)}, dailyFreeMatchLimit, int(secondsUntilTomorrowUTC().Seconds())).Result()
	if err != nil {
		return matchUsage{}, false, err
	}
	items, ok := raw.([]any)
	if !ok || len(items) != 2 {
		return matchUsage{}, false, errors.New("unexpected quota script result")
	}
	allowed, _ := items[0].(int64)
	used64, _ := items[1].(int64)
	used := int(used64)
	remaining := dailyFreeMatchLimit - used
	if remaining < 0 {
		remaining = 0
	}
	return matchUsage{DailyLimit: dailyFreeMatchLimit, DailyUsed: used, DailyRemaining: remaining}, allowed == 1, nil
}

func dailyQuotaKey(userID string) string {
	return "match:quota:" + time.Now().UTC().Format("20060102") + ":" + userID
}

func secondsUntilTomorrowUTC() time.Duration {
	now := time.Now().UTC()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 2, 0, 0, 0, time.UTC)
	return tomorrow.Sub(now)
}
