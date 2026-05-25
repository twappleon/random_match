package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"random-match/backend/internal/config"
	"random-match/backend/internal/model"
	"random-match/backend/internal/store"
)

type Server struct {
	cfg   config.Config
	db    *store.Mongo
	cache *store.Redis
	hub   *Hub
}

func New(cfg config.Config, db *store.Mongo, cache *store.Redis) *Server {
	return &Server{
		cfg:   cfg,
		db:    db,
		cache: cache,
		hub:   NewHub(),
	}
}

func (s *Server) Routes() http.Handler {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(s.requestLog(), s.cors(), gin.Recovery())

	router.GET("/health", s.health)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	v1.POST("/auth/anonymous", s.anonymousAuth)
	v1.GET("/stats", s.stats)
	v1.GET("/ws", s.ws)

	auth := v1.Group("")
	auth.Use(s.requireAuth())
	auth.GET("/auth/session", s.authSession)
	auth.GET("/me", s.getProfile)
	auth.PUT("/me", s.updateProfile)
	auth.POST("/match/join", s.joinMatch)
	auth.POST("/match/leave", s.leaveMatch)
	auth.POST("/match/snapshot", s.saveMatchSnapshot)
	auth.POST("/users/:id/block", s.blockUser)
	auth.POST("/users/:id/report", s.reportUser)
	auth.POST("/push/subscription", s.savePushSubscriptionHandler)
	auth.POST("/push/test", s.testPushHandler)

	return router
}

// health godoc
//
//	@Summary		Health check
//	@Description	Returns the backend service health status.
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	healthResponse
//	@Router			/health [get]
func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// authSession godoc
//
//	@Summary		Verify current session
//	@Description	Verifies the bearer token and returns the authenticated user id.
//	@Tags			auth
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	authSessionResponse
//	@Failure		401	{object}	errorResponse
//	@Router			/api/v1/auth/session [get]
func (s *Server) authSession(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"userId": userIDFromContext(c)})
}

// anonymousAuth godoc
//
//	@Summary		Create anonymous user
//	@Description	Creates an anonymous user and returns a JWT for API access.
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	anonymousAuthResponse
//	@Failure		500	{object}	errorResponse
//	@Router			/api/v1/auth/anonymous [post]
func (s *Server) anonymousAuth(c *gin.Context) {
	now := time.Now().UTC()
	user := model.User{
		ID:           newID(),
		DisplayName:  "星球旅人",
		AvatarURL:    defaultAvatarURL(),
		Interests:    []string{"聊天", "电影", "音乐"},
		AgeConfirmed: false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	_, err := s.db.DB.Collection("users").InsertOne(ctx, user)
	if err != nil {
		log.Printf("auth anonymous failed user_id=%s err=%v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create user failed"})
		return
	}

	token, err := s.signToken(user.ID)
	if err != nil {
		log.Printf("auth sign token failed user_id=%s err=%v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sign token failed"})
		return
	}

	log.Printf("auth anonymous ok user_id=%s", user.ID)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

// getProfile godoc
//
//	@Summary		Get current user profile
//	@Description	Returns the current user's anonymous social profile.
//	@Tags			profile
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	profileResponse
//	@Failure		401	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Router			/api/v1/me [get]
func (s *Server) getProfile(c *gin.Context) {
	userID := userIDFromContext(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	profile, err := s.userProfile(ctx, userID)
	if err != nil {
		log.Printf("profile get failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get profile failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": profile})
}

// updateProfile godoc
//
//	@Summary		Update current user profile
//	@Description	Updates display name, bio, interests, and age confirmation.
//	@Tags			profile
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		updateProfileRequest	true	"Profile fields"
//	@Success		200		{object}	profileResponse
//	@Failure		400		{object}	errorResponse
//	@Failure		401		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/api/v1/me [put]
func (s *Server) updateProfile(c *gin.Context) {
	userID := userIDFromContext(c)
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		displayName = "星球旅人"
	}
	if len([]rune(displayName)) > 24 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "displayName is too long"})
		return
	}
	bio := strings.TrimSpace(req.Bio)
	if len([]rune(bio)) > 120 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bio is too long"})
		return
	}
	interests := normalizeInterests(req.Interests)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	_, err := s.db.DB.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{
		"displayName":  displayName,
		"bio":          bio,
		"interests":    interests,
		"ageConfirmed": req.AgeConfirmed,
		"updatedAt":    time.Now().UTC(),
	}})
	if err != nil {
		log.Printf("profile update failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update profile failed"})
		return
	}
	profile, err := s.userProfile(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "get profile failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": profile})
}

// stats godoc
//
//	@Summary		Runtime usage stats
//	@Description	Returns current online users, waiting users, and users in active chats.
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	statsResponse
//	@Failure		500	{object}	errorResponse
//	@Router			/api/v1/stats [get]
func (s *Server) stats(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	waiting, err := countWaitingUsers(ctx, s.cache.Client)
	if err != nil {
		log.Printf("stats waiting count failed err=%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "stats failed"})
		return
	}

	online, chatting := s.hub.Stats()
	c.JSON(http.StatusOK, gin.H{
		"online":   online,
		"waiting":  waiting,
		"chatting": chatting,
	})
}

// joinMatch godoc
//
//	@Summary		Join matchmaking queue
//	@Description	Adds the current user to the video matchmaking queue. Returns 202 while waiting or 200 when matched.
//	@Tags			match
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		joinMatchRequest	true	"Match preferences"
//	@Success		200		{object}	matchedResponse
//	@Success		202		{object}	waitingMatchResponse
//	@Failure		400		{object}	errorResponse
//	@Failure		401		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/api/v1/match/join [post]
func (s *Server) joinMatch(c *gin.Context) {
	userID := userIDFromContext(c)
	var req joinMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("match join invalid_json user_id=%s err=%v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if req.Mode != model.MatchModeVideo {
		log.Printf("match join invalid_mode user_id=%s mode=%q", userID, req.Mode)
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be video"})
		return
	}
	if strings.TrimSpace(req.Region) == "" {
		req.Region = "global"
	}

	queueKey := "match:queue:v2:" + string(req.Mode) + ":" + req.Region
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	self := model.MatchTicket{UserID: userID, Mode: req.Mode, Region: req.Region, CreatedAt: time.Now().UTC()}
	result, err := enqueueAndMatch(ctx, s.cache.Client, queueKey, self)
	if err != nil {
		log.Printf("match join failed user_id=%s mode=%s region=%s err=%v", userID, req.Mode, req.Region, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "match failed"})
		return
	}
	if result.waiting {
		log.Printf("match waiting user_id=%s mode=%s region=%s queue=%s", userID, req.Mode, req.Region, queueKey)
		c.JSON(http.StatusAccepted, gin.H{"status": "waiting"})
		return
	}

	partner := result.partner
	if !s.hub.IsOnline(partner.UserID) {
		log.Printf("match stale_partner user_id=%s peer_id=%s mode=%s region=%s", userID, partner.UserID, req.Mode, req.Region)
		self = model.MatchTicket{UserID: userID, Mode: req.Mode, Region: req.Region, CreatedAt: time.Now().UTC()}
		if err := s.cache.Client.RPush(ctx, queueKey, self.UserID).Err(); err != nil {
			log.Printf("match requeue failed user_id=%s mode=%s region=%s err=%v", userID, req.Mode, req.Region, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "match failed"})
			return
		}
		if err := s.cache.Client.SAdd(ctx, queueKey+":members", self.UserID).Err(); err != nil {
			log.Printf("match requeue member failed user_id=%s mode=%s region=%s err=%v", userID, req.Mode, req.Region, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "match failed"})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"status": "waiting"})
		return
	}
	blocked, err := s.isBlockedEither(ctx, userID, partner.UserID)
	if err != nil {
		log.Printf("match block check failed user_id=%s peer_id=%s err=%v", userID, partner.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "match failed"})
		return
	}
	if blocked {
		log.Printf("match blocked_pair user_id=%s peer_id=%s mode=%s region=%s", userID, partner.UserID, req.Mode, req.Region)
		if err := s.cache.Client.RPush(ctx, queueKey, partner.UserID).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "match failed"})
			return
		}
		if err := s.cache.Client.SAdd(ctx, queueKey+":members", partner.UserID).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "match failed"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "blocked match skipped, retry"})
		return
	}
	selfProfile, err := s.userProfile(ctx, userID)
	if err != nil {
		log.Printf("match self profile failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "match failed"})
		return
	}
	partnerProfile, err := s.userProfile(ctx, partner.UserID)
	if err != nil {
		log.Printf("match partner profile failed user_id=%s peer_id=%s err=%v", userID, partner.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "match failed"})
		return
	}
	roomID := newID()
	log.Printf("match paired room_id=%s user_id=%s peer_id=%s mode=%s region=%s", roomID, userID, partner.UserID, req.Mode, req.Region)
	s.hub.Pair(userID, partner.UserID)
	s.hub.Notify(partner.UserID, SignalMessage{Type: "matched", RoomID: roomID, PeerID: userID, PeerProfile: selfProfile, Mode: string(req.Mode), Initiator: false})
	s.hub.Notify(userID, SignalMessage{Type: "matched", RoomID: roomID, PeerID: partner.UserID, PeerProfile: partnerProfile, Mode: string(req.Mode), Initiator: true})
	c.JSON(http.StatusOK, gin.H{
		"status":      "matched",
		"roomId":      roomID,
		"peerId":      partner.UserID,
		"peerProfile": partnerProfile,
		"initiator":   true,
	})
}

// leaveMatch godoc
//
//	@Summary		Leave matchmaking queue
//	@Description	Marks the current user as having left matchmaking.
//	@Tags			match
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	leaveMatchResponse
//	@Failure		401	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Router			/api/v1/match/leave [post]
func (s *Server) leaveMatch(c *gin.Context) {
	userID := userIDFromContext(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if err := removeUserFromMatchQueues(ctx, s.cache.Client, userID); err != nil {
		log.Printf("match leave queue cleanup failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "leave match failed"})
		return
	}

	if peerID := s.hub.Unpair(userID); peerID != "" {
		s.hub.Notify(peerID, SignalMessage{Type: "peer-left", PeerID: userID})
	}
	log.Printf("match leave user_id=%s", userID)
	c.JSON(http.StatusOK, gin.H{"status": "left"})
}

// blockUser godoc
//
//	@Summary		Block a user
//	@Description	Blocks another user and prevents future matches.
//	@Tags			safety
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string				true	"Target user id"
//	@Param			request	body	userActionRequest	false	"Reason"
//	@Success		200		{object}	userActionResponse
//	@Failure		401		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/api/v1/users/{id}/block [post]
func (s *Server) blockUser(c *gin.Context) {
	userID := userIDFromContext(c)
	targetID := strings.TrimSpace(c.Param("id"))
	if targetID == "" || targetID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target"})
		return
	}
	now := time.Now().UTC()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	block := model.UserBlock{
		ID:        userID + ":" + targetID,
		UserID:    userID,
		BlockedID: targetID,
		CreatedAt: now,
	}
	_, err := s.db.DB.Collection("user_blocks").UpdateOne(ctx, bson.M{"_id": block.ID}, bson.M{"$setOnInsert": block}, updateUpsert())
	if err != nil {
		log.Printf("user block failed user_id=%s target_id=%s err=%v", userID, targetID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "block failed"})
		return
	}
	if peerID := s.hub.Unpair(userID); peerID != "" {
		s.hub.Notify(peerID, SignalMessage{Type: "peer-left", PeerID: userID})
	}
	c.JSON(http.StatusOK, gin.H{"status": "blocked"})
}

// reportUser godoc
//
//	@Summary		Report a user
//	@Description	Reports another user for moderation review.
//	@Tags			safety
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path	string				true	"Target user id"
//	@Param			request	body	userActionRequest	false	"Reason"
//	@Success		200		{object}	userActionResponse
//	@Failure		401		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/api/v1/users/{id}/report [post]
func (s *Server) reportUser(c *gin.Context) {
	userID := userIDFromContext(c)
	targetID := strings.TrimSpace(c.Param("id"))
	if targetID == "" || targetID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target"})
		return
	}
	var req userActionRequest
	_ = c.ShouldBindJSON(&req)
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		reason = "user report"
	}
	if len([]rune(reason)) > 200 {
		reason = string([]rune(reason)[:200])
	}
	report := model.UserReport{
		ID:         newID(),
		ReporterID: userID,
		TargetID:   targetID,
		Reason:     reason,
		CreatedAt:  time.Now().UTC(),
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if _, err := s.db.DB.Collection("user_reports").InsertOne(ctx, report); err != nil {
		log.Printf("user report failed user_id=%s target_id=%s err=%v", userID, targetID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "report failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "reported"})
}

// saveMatchSnapshot godoc
//
//	@Summary		Save match snapshot
//	@Description	Saves a camera snapshot on the server after a match is created.
//	@Tags			match
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		snapshotRequest	true	"Snapshot payload"
//	@Success		200		{object}	snapshotResponse
//	@Failure		400		{object}	errorResponse
//	@Failure		401		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/api/v1/match/snapshot [post]
func (s *Server) saveMatchSnapshot(c *gin.Context) {
	userID := userIDFromContext(c)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 3*1024*1024)

	var req snapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("snapshot invalid_json user_id=%s err=%v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid snapshot"})
		return
	}
	if strings.TrimSpace(req.RoomID) == "" || strings.TrimSpace(req.Image) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "roomId and image are required"})
		return
	}

	path, err := saveSnapshot(s.cfg.SnapshotDir, userID, req)
	if err != nil {
		log.Printf("snapshot save failed user_id=%s room_id=%s err=%v", userID, req.RoomID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save snapshot failed"})
		return
	}
	log.Printf("snapshot saved user_id=%s peer_id=%s room_id=%s path=%s", userID, req.PeerID, req.RoomID, path)
	c.JSON(http.StatusOK, gin.H{"status": "saved", "path": path})
}

// savePushSubscriptionHandler godoc
//
//	@Summary		Save browser push subscription
//	@Description	Saves the current user's browser Web Push subscription so the server can notify them while offline.
//	@Tags			push
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		pushSubscriptionRequest	true	"Push subscription"
//	@Success		200		{object}	pushSubscriptionResponse
//	@Failure		400		{object}	errorResponse
//	@Failure		401		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/api/v1/push/subscription [post]
func (s *Server) savePushSubscriptionHandler(c *gin.Context) {
	userID := userIDFromContext(c)
	var req pushSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("push subscription invalid_json user_id=%s err=%v", userID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription"})
		return
	}
	if strings.TrimSpace(req.Endpoint) == "" || strings.TrimSpace(req.Keys.Auth) == "" || strings.TrimSpace(req.Keys.P256dh) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "subscription endpoint and keys are required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := s.savePushSubscription(ctx, userID, req); err != nil {
		log.Printf("push subscription save failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save push subscription failed"})
		return
	}
	log.Printf("push subscription saved user_id=%s", userID)
	c.JSON(http.StatusOK, gin.H{"status": "saved"})
}

// testPushHandler godoc
//
//	@Summary		Send a test push notification
//	@Description	Sends a server-side Web Push notification to the current user's latest subscription.
//	@Tags			push
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	pushTestResponse
//	@Failure		401	{object}	errorResponse
//	@Failure		404	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Router			/api/v1/push/test [post]
func (s *Server) testPushHandler(c *gin.Context) {
	userID := userIDFromContext(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := s.sendTestPush(ctx, userID); err != nil {
		log.Printf("push test failed user_id=%s err=%v", userID, err)
		status := http.StatusInternalServerError
		if errors.Is(err, errPushNotConfigured) {
			status = http.StatusServiceUnavailable
		}
		if errors.Is(err, errPushSubscriptionNotFound) {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "sent"})
}

// ws godoc
//
//	@Summary		Connect signaling WebSocket
//	@Description	Upgrades to a WebSocket used for match and WebRTC signaling messages. Pass the JWT as the token query parameter.
//	@Tags			signaling
//	@Param			token	query	string	true	"JWT token"
//	@Success		101		{string}	string	"Switching Protocols"
//	@Failure		401		{object}	errorResponse
//	@Router			/api/v1/ws [get]
func (s *Server) ws(c *gin.Context) {
	token := c.Query("token")
	userID, err := s.verifyToken(token)
	if err != nil {
		log.Printf("ws auth failed origin=%s err=%v", c.Request.Header.Get("Origin"), err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		OriginPatterns: s.cfg.CORSOrigins,
	})
	if err != nil {
		log.Printf("ws accept failed user_id=%s origin=%s err=%v", userID, c.Request.Header.Get("Origin"), err)
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	log.Printf("ws connected user_id=%s origin=%s", userID, c.Request.Header.Get("Origin"))
	client := s.hub.Register(userID, conn)
	go func() {
		notifyCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		s.notifyOfflineUsers(notifyCtx, userID)
	}()
	defer func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := removeUserFromMatchQueues(cleanupCtx, s.cache.Client, userID); err != nil {
			log.Printf("ws queue cleanup failed user_id=%s err=%v", userID, err)
		}
		s.hub.Unregister(userID, client)
		log.Printf("ws disconnected user_id=%s", userID)
	}()
	client.ReadLoop(c.Request.Context(), s.hub)
}

func (s *Server) signToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Server) verifyToken(raw string) (string, error) {
	token, err := jwt.Parse(raw, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return "", errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("invalid claims")
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", errors.New("missing subject")
	}
	return sub, nil
}

func (s *Server) requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		userID, err := s.verifyToken(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		if err := s.db.DB.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Err(); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
}

func userIDFromContext(c *gin.Context) string {
	userID, _ := c.Get("userID")
	value, _ := userID.(string)
	return value
}

func (s *Server) userProfile(ctx context.Context, userID string) (model.UserProfile, error) {
	var user model.User
	if err := s.db.DB.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		return model.UserProfile{}, err
	}
	return model.UserProfile{
		ID:           user.ID,
		DisplayName:  user.DisplayName,
		AvatarURL:    user.AvatarURL,
		Bio:          user.Bio,
		Interests:    user.Interests,
		AgeConfirmed: user.AgeConfirmed,
	}, nil
}

func (s *Server) isBlockedEither(ctx context.Context, userID, peerID string) (bool, error) {
	count, err := s.db.DB.Collection("user_blocks").CountDocuments(ctx, bson.M{"$or": []bson.M{
		{"userId": userID, "blockedId": peerID},
		{"userId": peerID, "blockedId": userID},
	}})
	return count > 0, err
}

func normalizeInterests(items []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, 6)
	for _, item := range items {
		value := strings.TrimSpace(item)
		if value == "" || seen[value] {
			continue
		}
		if len([]rune(value)) > 12 {
			value = string([]rune(value)[:12])
		}
		seen[value] = true
		out = append(out, value)
		if len(out) == 6 {
			break
		}
	}
	if len(out) == 0 {
		return []string{"聊天"}
	}
	return out
}

func defaultAvatarURL() string {
	return ""
}

func updateUpsert() *options.UpdateOptions {
	return options.Update().SetUpsert(true)
}

func (s *Server) cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		for _, allowed := range s.cfg.CORSOrigins {
			if origin == allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				break
			}
		}
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func (s *Server) requestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		if strings.EqualFold(c.GetHeader("Upgrade"), "websocket") {
			log.Printf("http websocket start method=%s path=%s origin=%s remote=%s", c.Request.Method, c.Request.URL.Path, c.GetHeader("Origin"), c.ClientIP())
			c.Next()
			log.Printf("http websocket end method=%s path=%s duration=%s origin=%s remote=%s", c.Request.Method, c.Request.URL.Path, time.Since(start).Round(time.Millisecond), c.GetHeader("Origin"), c.ClientIP())
			return
		}
		c.Next()
		log.Printf("http method=%s path=%s status=%d duration=%s origin=%s remote=%s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start).Round(time.Millisecond), c.GetHeader("Origin"), c.ClientIP())
	}
}

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(b[:])
}
