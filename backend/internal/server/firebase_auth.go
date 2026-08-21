package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	firebase "firebase.google.com/go/v4"
	firebaseauth "firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"random-match/backend/internal/model"
)

var (
	firebaseAuthClientMu sync.Mutex
	firebaseAuthClient   *firebaseauth.Client
	firebaseAuthInitErr  error
	errFirebaseAuthUnset = errors.New("firebase auth is not configured")
)

func (s *Server) firebaseAuthClientForAuth(ctx context.Context) (*firebaseauth.Client, error) {
	firebaseAuthClientMu.Lock()
	defer firebaseAuthClientMu.Unlock()
	if firebaseAuthClient != nil {
		return firebaseAuthClient, nil
	}
	if firebaseAuthInitErr != nil {
		return nil, firebaseAuthInitErr
	}
	if s.cfg.FirebaseProjectID == "" {
		firebaseAuthInitErr = errFirebaseAuthUnset
		return nil, firebaseAuthInitErr
	}
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: s.cfg.FirebaseProjectID})
	if err != nil {
		firebaseAuthInitErr = err
		return nil, err
	}
	client, err := app.Auth(ctx)
	if err != nil {
		firebaseAuthInitErr = err
		return nil, err
	}
	firebaseAuthClient = client
	return firebaseAuthClient, nil
}

func (s *Server) firebaseAuth(c *gin.Context) {
	var req firebaseAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.IDToken) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid firebase token"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
	defer cancel()
	client, err := s.firebaseAuthClientForAuth(ctx)
	if err != nil {
		log.Printf("firebase auth init failed err=%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "firebase auth is not configured"})
		return
	}
	token, err := client.VerifyIDToken(ctx, req.IDToken)
	if err != nil {
		log.Printf("firebase auth verify failed err=%v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid firebase token"})
		return
	}
	firebaseUser, err := client.GetUser(ctx, token.UID)
	if err != nil {
		log.Printf("firebase auth user fetch failed uid=%s err=%v", token.UID, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "firebase user not found"})
		return
	}

	user, err := s.upsertFirebaseUser(ctx, firebaseUser)
	if err != nil {
		log.Printf("firebase auth upsert failed uid=%s err=%v", token.UID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "firebase login failed"})
		return
	}
	apiToken, err := s.signToken(user.ID)
	if err != nil {
		log.Printf("firebase auth sign failed user_id=%s err=%v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sign token failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": apiToken,
		"user":  profileFromUser(user),
	})
}

func (s *Server) upsertFirebaseUser(ctx context.Context, firebaseUser *firebaseauth.UserRecord) (model.User, error) {
	users := s.db.DB.Collection("users")
	var user model.User
	if err := users.FindOne(ctx, bson.M{"firebaseUid": firebaseUser.UID}).Decode(&user); err == nil {
		update := bson.M{
			"firebaseUid": firebaseUser.UID,
			"updatedAt":   time.Now().UTC(),
		}
		if strings.TrimSpace(user.DisplayName) == "" {
			update["displayName"] = firebaseDisplayName(firebaseUser)
		}
		if strings.TrimSpace(user.AvatarURL) == "" && strings.TrimSpace(firebaseUser.PhotoURL) != "" {
			update["avatarUrl"] = strings.TrimSpace(firebaseUser.PhotoURL)
		}
		_, _ = users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": update})
		if update["displayName"] != nil {
			user.DisplayName = update["displayName"].(string)
		}
		if update["avatarUrl"] != nil {
			user.AvatarURL = update["avatarUrl"].(string)
		}
		return user, nil
	} else if err != mongo.ErrNoDocuments {
		return model.User{}, err
	}

	now := time.Now().UTC()
	user = model.User{
		ID:           newID(),
		DisplayName:  firebaseDisplayName(firebaseUser),
		AvatarURL:    firebaseAvatarURL(firebaseUser),
		Interests:    defaultUserInterests(firebaseUser.UID),
		Region:       "global",
		Gender:       "private",
		Language:     "zh",
		GemsBalance:  120,
		AgeConfirmed: false,
		FirebaseUID:  firebaseUser.UID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if _, err := users.InsertOne(ctx, user); err != nil {
		return model.User{}, err
	}
	return user, nil
}

func firebaseDisplayName(firebaseUser *firebaseauth.UserRecord) string {
	if name := strings.TrimSpace(firebaseUser.DisplayName); name != "" {
		return name
	}
	if firebaseUser.Email != "" {
		return strings.Split(firebaseUser.Email, "@")[0]
	}
	return defaultDisplayName(firebaseUser.UID)
}

func firebaseAvatarURL(firebaseUser *firebaseauth.UserRecord) string {
	if url := strings.TrimSpace(firebaseUser.PhotoURL); url != "" {
		return url
	}
	return defaultAvatarURL()
}
