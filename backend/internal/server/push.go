package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const pushCooldown = 10 * time.Minute

var (
	errPushNotConfigured        = errors.New("push is not configured")
	errPushSubscriptionNotFound = errors.New("push subscription not found")
)

type pushSubscriptionDocument struct {
	UserID    string               `bson:"userId"`
	Endpoint  string               `bson:"endpoint"`
	Keys      pushSubscriptionKeys `bson:"keys"`
	CreatedAt time.Time            `bson:"createdAt"`
	UpdatedAt time.Time            `bson:"updatedAt"`
}

type pushDeviceTokenDocument struct {
	UserID    string    `bson:"userId"`
	Token     string    `bson:"token"`
	Platform  string    `bson:"platform"`
	CreatedAt time.Time `bson:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

var (
	fcmClientMu sync.Mutex
	fcmClient   *messaging.Client
	fcmInitErr  error
)

func (s *Server) pushSubscriptions() *mongo.Collection {
	return s.db.DB.Collection("push_subscriptions")
}

func (s *Server) pushDeviceTokens() *mongo.Collection {
	return s.db.DB.Collection("push_device_tokens")
}

func (s *Server) notifyOfflineUsers(ctx context.Context, actorUserID string) {
	payload, _ := json.Marshal(ginPushPayload{
		Title: "有人上线了",
		Body:  "现在有人在线，可以开始随机匹配。",
		URL:   "/",
	})
	s.notifyOfflineWebPushUsers(ctx, actorUserID, payload)
	s.notifyOfflineNativePushUsers(ctx, actorUserID)
}

func (s *Server) notifyOfflineWebPushUsers(ctx context.Context, actorUserID string, payload []byte) {
	if s.cfg.VAPIDPublicKey == "" || s.cfg.VAPIDPrivateKey == "" {
		log.Printf("push web skip reason=vapid_not_configured")
		return
	}

	cursor, err := s.pushSubscriptions().Find(ctx, bson.M{"userId": bson.M{"$ne": actorUserID}})
	if err != nil {
		log.Printf("push web find subscriptions failed err=%v", err)
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var sub pushSubscriptionDocument
		if err := cursor.Decode(&sub); err != nil {
			log.Printf("push web decode subscription failed err=%v", err)
			continue
		}
		if s.hub.IsOnline(sub.UserID) || !s.acquirePushCooldown(ctx, sub.UserID) {
			continue
		}
		_, _ = s.sendPush(ctx, sub, payload)
	}
	if err := cursor.Err(); err != nil {
		log.Printf("push web cursor failed err=%v", err)
	}
}

func (s *Server) notifyOfflineNativePushUsers(ctx context.Context, actorUserID string) {
	client, err := s.fcmMessagingClient(ctx)
	if err != nil {
		log.Printf("push native skip reason=fcm_not_configured err=%v", err)
		return
	}

	cursor, err := s.pushDeviceTokens().Find(ctx, bson.M{"userId": bson.M{"$ne": actorUserID}})
	if err != nil {
		log.Printf("push native find tokens failed err=%v", err)
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var token pushDeviceTokenDocument
		if err := cursor.Decode(&token); err != nil {
			log.Printf("push native decode token failed err=%v", err)
			continue
		}
		if s.hub.IsOnline(token.UserID) || !s.acquirePushCooldown(ctx, token.UserID) {
			continue
		}
		if err := s.sendNativePush(ctx, client, token, "有人上线了", "现在有人在线，可以开始随机匹配。"); err != nil {
			log.Printf("push native send failed user_id=%s platform=%s err=%v", token.UserID, token.Platform, err)
		}
	}
	if err := cursor.Err(); err != nil {
		log.Printf("push native cursor failed err=%v", err)
	}
}

type ginPushPayload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
}

func (s *Server) acquirePushCooldown(ctx context.Context, userID string) bool {
	key := "push:online:cooldown:" + userID
	ok, err := s.cache.Client.SetNX(ctx, key, "1", pushCooldown).Result()
	if err != nil && err != redis.Nil {
		log.Printf("push cooldown failed user_id=%s err=%v", userID, err)
		return false
	}
	return ok
}

func (s *Server) fcmMessagingClient(ctx context.Context) (*messaging.Client, error) {
	fcmClientMu.Lock()
	defer fcmClientMu.Unlock()
	if fcmClient != nil {
		return fcmClient, nil
	}
	if fcmInitErr != nil {
		return nil, fcmInitErr
	}
	if s.cfg.FirebaseProjectID == "" {
		fcmInitErr = errPushNotConfigured
		return nil, fcmInitErr
	}
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: s.cfg.FirebaseProjectID})
	if err != nil {
		fcmInitErr = err
		return nil, err
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		fcmInitErr = err
		return nil, err
	}
	fcmClient = client
	return fcmClient, nil
}

func (s *Server) sendNativePush(ctx context.Context, client *messaging.Client, doc pushDeviceTokenDocument, title, body string) error {
	_, err := client.Send(ctx, &messaging.Message{
		Token: doc.Token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: map[string]string{
			"type": "user-online",
			"url":  "/",
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	})
	if err != nil {
		if messaging.IsRegistrationTokenNotRegistered(err) {
			_, deleteErr := s.pushDeviceTokens().DeleteOne(ctx, bson.M{"token": doc.Token})
			if deleteErr != nil {
				log.Printf("push native delete expired failed user_id=%s err=%v", doc.UserID, deleteErr)
			}
		}
		return err
	}
	log.Printf("push native sent user_id=%s platform=%s", doc.UserID, doc.Platform)
	return nil
}

func (s *Server) sendPush(ctx context.Context, doc pushSubscriptionDocument, payload []byte) (int, error) {
	resp, err := webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: doc.Endpoint,
		Keys: webpush.Keys{
			Auth:   doc.Keys.Auth,
			P256dh: doc.Keys.P256dh,
		},
	}, &webpush.Options{
		Subscriber:      s.cfg.VAPIDSubject,
		VAPIDPublicKey:  s.cfg.VAPIDPublicKey,
		VAPIDPrivateKey: s.cfg.VAPIDPrivateKey,
		TTL:             300,
		Topic:           "user-online",
	})
	if err != nil {
		log.Printf("push send failed user_id=%s err=%v", doc.UserID, err)
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
		_, err := s.pushSubscriptions().DeleteOne(ctx, bson.M{"endpoint": doc.Endpoint})
		if err != nil {
			log.Printf("push delete expired failed user_id=%s status=%d err=%v", doc.UserID, resp.StatusCode, err)
		}
		return resp.StatusCode, fmt.Errorf("push subscription expired status=%d", resp.StatusCode)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("push send non_2xx user_id=%s status=%d", doc.UserID, resp.StatusCode)
		return resp.StatusCode, fmt.Errorf("push service returned status=%d", resp.StatusCode)
	}
	log.Printf("push sent user_id=%s status=%d", doc.UserID, resp.StatusCode)
	return resp.StatusCode, nil
}

func (s *Server) savePushSubscription(ctx context.Context, userID string, req pushSubscriptionRequest) error {
	now := time.Now().UTC()
	_, err := s.pushSubscriptions().UpdateOne(
		ctx,
		bson.M{"endpoint": req.Endpoint},
		bson.M{
			"$set": bson.M{
				"userId":    userID,
				"endpoint":  req.Endpoint,
				"keys":      req.Keys,
				"updatedAt": now,
			},
			"$setOnInsert": bson.M{"createdAt": now},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

func (s *Server) savePushDeviceToken(ctx context.Context, userID string, req pushDeviceTokenRequest) error {
	now := time.Now().UTC()
	_, err := s.pushDeviceTokens().UpdateOne(
		ctx,
		bson.M{"token": req.Token},
		bson.M{
			"$set": bson.M{
				"userId":    userID,
				"token":     req.Token,
				"platform":  req.Platform,
				"updatedAt": now,
			},
			"$setOnInsert": bson.M{"createdAt": now},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

func (s *Server) sendTestPush(ctx context.Context, userID string) error {
	webConfigured := s.cfg.VAPIDPublicKey != "" && s.cfg.VAPIDPrivateKey != ""
	nativeClient, nativeErr := s.fcmMessagingClient(ctx)
	if !webConfigured && nativeErr != nil {
		return errPushNotConfigured
	}

	var sub pushSubscriptionDocument
	err := s.pushSubscriptions().FindOne(
		ctx,
		bson.M{"userId": userID},
		options.FindOne().SetSort(bson.M{"updatedAt": -1}),
	).Decode(&sub)
	if err == nil && webConfigured {
		payload, _ := json.Marshal(ginPushPayload{
			Title: "服务器推送测试",
			Body:  "这是一则从后端发出的真实 Web Push。",
			URL:   "/",
		})
		_, err = s.sendPush(ctx, sub, payload)
		return err
	}
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if nativeErr != nil {
		return errPushSubscriptionNotFound
	}

	var token pushDeviceTokenDocument
	err = s.pushDeviceTokens().FindOne(
		ctx,
		bson.M{"userId": userID},
		options.FindOne().SetSort(bson.M{"updatedAt": -1}),
	).Decode(&token)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return errPushSubscriptionNotFound
	}
	if err != nil {
		return err
	}
	return s.sendNativePush(ctx, nativeClient, token, "服务器推送测试", "这是一则从后端发出的真实 App Push。")
}

// savePushDeviceTokenHandler godoc
//
//	@Summary		Save native app push token
//	@Description	Saves the current user's Firebase Cloud Messaging token for native app notifications.
//	@Tags			push
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		pushDeviceTokenRequest	true	"FCM device token"
//	@Success		200		{object}	pushDeviceTokenResponse
//	@Failure		400		{object}	errorResponse
//	@Failure		401		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/api/v1/push/device-token [post]
func (s *Server) savePushDeviceTokenHandler(c *gin.Context) {
	userID := userIDFromContext(c)
	var req pushDeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid device token"})
		return
	}
	req.Token = strings.TrimSpace(req.Token)
	req.Platform = strings.ToLower(strings.TrimSpace(req.Platform))
	if req.Token == "" || (req.Platform != "ios" && req.Platform != "android") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token and platform are required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := s.savePushDeviceToken(ctx, userID, req); err != nil {
		log.Printf("push device token save failed user_id=%s platform=%s err=%v", userID, req.Platform, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save device token failed"})
		return
	}
	log.Printf("push device token saved user_id=%s platform=%s", userID, req.Platform)
	c.JSON(http.StatusOK, gin.H{"status": "saved"})
}
