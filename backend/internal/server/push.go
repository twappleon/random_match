package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const pushCooldown = 10 * time.Minute

type pushSubscriptionDocument struct {
	UserID    string               `bson:"userId"`
	Endpoint  string               `bson:"endpoint"`
	Keys      pushSubscriptionKeys `bson:"keys"`
	CreatedAt time.Time            `bson:"createdAt"`
	UpdatedAt time.Time            `bson:"updatedAt"`
}

func (s *Server) pushSubscriptions() *mongo.Collection {
	return s.db.DB.Collection("push_subscriptions")
}

func (s *Server) notifyOfflineUsers(ctx context.Context, actorUserID string) {
	if s.cfg.VAPIDPublicKey == "" || s.cfg.VAPIDPrivateKey == "" {
		log.Printf("push skip reason=vapid_not_configured")
		return
	}

	cursor, err := s.pushSubscriptions().Find(ctx, bson.M{"userId": bson.M{"$ne": actorUserID}})
	if err != nil {
		log.Printf("push find subscriptions failed err=%v", err)
		return
	}
	defer cursor.Close(ctx)

	payload, _ := json.Marshal(ginPushPayload{
		Title: "有人上线了",
		Body:  "现在有人在线，可以开始随机匹配。",
		URL:   "/",
	})

	for cursor.Next(ctx) {
		var sub pushSubscriptionDocument
		if err := cursor.Decode(&sub); err != nil {
			log.Printf("push decode subscription failed err=%v", err)
			continue
		}
		if s.hub.IsOnline(sub.UserID) || !s.acquirePushCooldown(ctx, sub.UserID) {
			continue
		}
		s.sendPush(ctx, sub, payload)
	}
	if err := cursor.Err(); err != nil {
		log.Printf("push cursor failed err=%v", err)
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

func (s *Server) sendPush(ctx context.Context, doc pushSubscriptionDocument, payload []byte) {
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
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusGone || resp.StatusCode == http.StatusNotFound {
		_, err := s.pushSubscriptions().DeleteOne(ctx, bson.M{"endpoint": doc.Endpoint})
		if err != nil {
			log.Printf("push delete expired failed user_id=%s status=%d err=%v", doc.UserID, resp.StatusCode, err)
		}
		return
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("push send non_2xx user_id=%s status=%d", doc.UserID, resp.StatusCode)
		return
	}
	log.Printf("push sent user_id=%s status=%d", doc.UserID, resp.StatusCode)
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
