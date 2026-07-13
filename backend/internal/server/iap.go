package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"

	"random-match/backend/internal/model"
)

const appleIAPGracePeriod = 2 * time.Minute

type appleTransactionClaims struct {
	BundleID              string `json:"bundleId"`
	ProductID             string `json:"productId"`
	TransactionID         string `json:"transactionId"`
	OriginalTransactionID string `json:"originalTransactionId"`
	Environment           string `json:"environment"`
	ExpiresDate           any    `json:"expiresDate"`
	PurchaseDate          any    `json:"purchaseDate"`
	RevocationDate        any    `json:"revocationDate"`
}

type appleTransactionLookupResponse struct {
	SignedTransactionInfo string `json:"signedTransactionInfo"`
}

func (s *Server) verifyAppleIAPPurchase(c *gin.Context) {
	userID := userIDFromContext(c)
	var req appleIAPVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	req.ProductID = strings.TrimSpace(req.ProductID)
	req.PurchaseID = strings.TrimSpace(req.PurchaseID)
	req.VerificationData = strings.TrimSpace(req.VerificationData)
	if req.ProductID == "" {
		req.ProductID = s.cfg.AppleIAPProductID
	}
	if req.ProductID != s.cfg.AppleIAPProductID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported product"})
		return
	}
	if req.PurchaseID == "" && req.VerificationData == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing purchase verification"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 12*time.Second)
	defer cancel()

	claims, err := s.verifiedAppleTransaction(ctx, req)
	if err != nil {
		log.Printf("apple iap verify failed user_id=%s product_id=%s purchase_id=%s err=%v", userID, req.ProductID, req.PurchaseID, err)
		status := http.StatusBadRequest
		if errors.Is(err, errAppleIAPNotConfigured) {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	if err := s.validateAppleTransaction(claims); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	purchase, err := s.activateAppleMembership(ctx, userID, claims, req.Source)
	if err != nil {
		log.Printf("apple iap activate failed user_id=%s transaction_id=%s err=%v", userID, claims.TransactionID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "iap activation failed"})
		return
	}
	commerce, err := s.commerceStatusForUser(ctx, userID)
	if err != nil {
		log.Printf("apple iap commerce reload failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "iap activation failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"commerce": commerce, "purchase": purchase})
}

var errAppleIAPNotConfigured = errors.New("apple iap server api is not configured")

func (s *Server) verifiedAppleTransaction(ctx context.Context, req appleIAPVerifyRequest) (appleTransactionClaims, error) {
	if s.appleServerAPIConfigured() {
		transactionID := req.PurchaseID
		if transactionID == "" {
			claims, err := parseAppleSignedTransaction(req.VerificationData)
			if err != nil {
				return appleTransactionClaims{}, err
			}
			transactionID = claims.TransactionID
		}
		signed, err := s.lookupAppleTransaction(ctx, transactionID)
		if err != nil {
			return appleTransactionClaims{}, err
		}
		return parseAppleSignedTransaction(signed)
	}
	if s.cfg.AppleIAPAllowUnverified && strings.EqualFold(s.cfg.AppEnv, "development") {
		return parseAppleSignedTransaction(req.VerificationData)
	}
	return appleTransactionClaims{}, errAppleIAPNotConfigured
}

func (s *Server) appleServerAPIConfigured() bool {
	return s.cfg.AppleIAPIssuerID != "" &&
		s.cfg.AppleIAPKeyID != "" &&
		(s.cfg.AppleIAPPrivateKey != "" || s.cfg.AppleIAPPrivateKeyPath != "")
}

func (s *Server) lookupAppleTransaction(ctx context.Context, transactionID string) (string, error) {
	if transactionID == "" {
		return "", errors.New("missing apple transaction id")
	}
	token, err := s.appleServerToken()
	if err != nil {
		return "", err
	}
	baseURL := "https://api.storekit.itunes.apple.com"
	if strings.EqualFold(s.cfg.AppleIAPEnvironment, "Sandbox") {
		baseURL = "https://api.storekit-sandbox.itunes.apple.com"
	}
	url := fmt.Sprintf("%s/inApps/v1/transactions/%s", baseURL, transactionID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("apple transaction lookup failed: %s %s", res.Status, strings.TrimSpace(string(body)))
	}
	var payload appleTransactionLookupResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	if payload.SignedTransactionInfo == "" {
		return "", errors.New("apple transaction lookup returned empty signed transaction")
	}
	return payload.SignedTransactionInfo, nil
}

func (s *Server) appleServerToken() (string, error) {
	privateKey, err := s.applePrivateKeyPEM()
	if err != nil {
		return "", err
	}
	key, err := jwt.ParseECPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"iss": s.cfg.AppleIAPIssuerID,
		"iat": now.Unix(),
		"exp": now.Add(20 * time.Minute).Unix(),
		"aud": "appstoreconnect-v1",
		"bid": s.cfg.AppleBundleID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = s.cfg.AppleIAPKeyID
	return token.SignedString(key)
}

func (s *Server) applePrivateKeyPEM() (string, error) {
	value := strings.TrimSpace(s.cfg.AppleIAPPrivateKey)
	if value == "" && s.cfg.AppleIAPPrivateKeyPath != "" {
		raw, err := os.ReadFile(s.cfg.AppleIAPPrivateKeyPath)
		if err != nil {
			return "", err
		}
		value = string(raw)
	}
	if value == "" {
		return "", errAppleIAPNotConfigured
	}
	value = strings.ReplaceAll(value, `\n`, "\n")
	return value, nil
}

func parseAppleSignedTransaction(value string) (appleTransactionClaims, error) {
	parts := strings.Split(value, ".")
	if len(parts) < 2 {
		return appleTransactionClaims{}, errors.New("invalid apple signed transaction")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return appleTransactionClaims{}, err
	}
	var claims appleTransactionClaims
	if err := json.Unmarshal(raw, &claims); err != nil {
		return appleTransactionClaims{}, err
	}
	if claims.TransactionID == "" {
		return appleTransactionClaims{}, errors.New("apple transaction missing transaction id")
	}
	return claims, nil
}

func (s *Server) validateAppleTransaction(claims appleTransactionClaims) error {
	if claims.BundleID != "" && claims.BundleID != s.cfg.AppleBundleID {
		return errors.New("apple transaction bundle id mismatch")
	}
	if claims.ProductID != s.cfg.AppleIAPProductID {
		return errors.New("apple transaction product mismatch")
	}
	if unixMillis(claims.RevocationDate) > 0 {
		return errors.New("apple transaction was revoked")
	}
	expiresAt := millisTime(claims.ExpiresDate)
	if expiresAt.IsZero() {
		return errors.New("apple transaction missing expiration")
	}
	if expiresAt.Add(appleIAPGracePeriod).Before(time.Now().UTC()) {
		return errors.New("apple transaction is expired")
	}
	return nil
}

func (s *Server) activateAppleMembership(ctx context.Context, userID string, claims appleTransactionClaims, source string) (model.AppleIAPPurchase, error) {
	now := time.Now().UTC()
	expiresAt := millisTime(claims.ExpiresDate)
	purchase := model.AppleIAPPurchase{
		ID:                    "apple:" + claims.TransactionID,
		UserID:                userID,
		ProductID:             claims.ProductID,
		TransactionID:         claims.TransactionID,
		OriginalTransactionID: claims.OriginalTransactionID,
		Environment:           claims.Environment,
		ExpiresAt:             expiresAt,
		VerifiedAt:            now,
		Source:                source,
	}

	var existing model.AppleIAPPurchase
	err := s.db.DB.Collection("apple_iap_purchases").FindOne(ctx, bson.M{"_id": purchase.ID}).Decode(&existing)
	isNew := err != nil
	if isNew {
		if _, err := s.db.DB.Collection("apple_iap_purchases").InsertOne(ctx, purchase); err != nil {
			return purchase, err
		}
	} else {
		purchase = existing
	}

	user, err := s.userByID(ctx, userID)
	if err != nil {
		return purchase, err
	}
	nextExpiresAt := expiresAt
	if user.MembershipExpiresAt != nil && user.MembershipExpiresAt.After(nextExpiresAt) {
		nextExpiresAt = *user.MembershipExpiresAt
	}
	update := bson.M{"$set": bson.M{
		"membershipPlan":      premiumMonthlyPlan,
		"membershipExpiresAt": nextExpiresAt,
		"updatedAt":           now,
	}}
	if isNew {
		update["$inc"] = bson.M{"gemsBalance": 300}
	}
	if _, err := s.db.DB.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, update); err != nil {
		return purchase, err
	}
	return purchase, nil
}

func unixMillis(value any) int64 {
	switch v := value.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	case string:
		parsed, _ := strconv.ParseInt(v, 10, 64)
		return parsed
	default:
		return 0
	}
}

func millisTime(value any) time.Time {
	ms := unixMillis(value)
	if ms <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms).UTC()
}
