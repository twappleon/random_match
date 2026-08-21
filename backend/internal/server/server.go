package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"math"
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
	v1.POST("/auth/firebase", s.firebaseAuth)
	v1.GET("/stats", s.stats)
	v1.GET("/ws", s.ws)

	auth := v1.Group("")
	auth.Use(s.requireAuth())
	auth.GET("/auth/session", s.authSession)
	auth.GET("/me", s.getProfile)
	auth.PUT("/me", s.updateProfile)
	auth.PUT("/me/location", s.updateLocation)
	auth.GET("/discover/profiles", s.discoverProfiles)
	auth.POST("/match/join", s.joinMatch)
	auth.POST("/match/leave", s.leaveMatch)
	auth.POST("/match/snapshot", s.saveMatchSnapshot)
	auth.GET("/users/blocks", s.listBlockedUsers)
	auth.GET("/users/follows", s.listFollowedUsers)
	auth.POST("/users/:id/block", s.blockUser)
	auth.DELETE("/users/:id/block", s.unblockUser)
	auth.POST("/users/:id/report", s.reportUser)
	auth.POST("/users/:id/follow", s.followUser)
	auth.DELETE("/users/:id/follow", s.unfollowUser)
	auth.POST("/users/:id/messages", s.sendDirectMessage)
	auth.GET("/commerce/status", s.commerceStatus)
	auth.POST("/commerce/orders", s.createPaymentOrder)
	auth.POST("/commerce/orders/:id/confirm", s.confirmPaymentOrder)
	auth.POST("/iap/apple/verify", s.verifyAppleIAPPurchase)
	auth.POST("/push/subscription", s.savePushSubscriptionHandler)
	auth.POST("/push/device-token", s.savePushDeviceTokenHandler)
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
	userID := newID()
	user := model.User{
		ID:           userID,
		DisplayName:  defaultDisplayName(userID),
		AvatarURL:    defaultAvatarURL(),
		Interests:    defaultUserInterests(userID),
		Region:       "global",
		Gender:       "private",
		Language:     "zh",
		GemsBalance:  120,
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
		displayName = defaultDisplayName(userID)
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
	region := normalizeRegion(req.Region)
	gender := normalizeProfileGender(req.Gender)
	language := normalizeLanguage(req.Language)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	_, err := s.db.DB.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{
		"displayName":  displayName,
		"bio":          bio,
		"interests":    interests,
		"region":       region,
		"gender":       gender,
		"language":     language,
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

// updateLocation godoc
//
//	@Summary		Update current user location
//	@Description	Stores the current user's latest coarse device coordinates. Other users only receive computed distance, never raw coordinates.
//	@Tags			profile
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		updateLocationRequest	true	"Location fields"
//	@Success		200		{object}	updateLocationResponse
//	@Failure		400		{object}	errorResponse
//	@Failure		401		{object}	errorResponse
//	@Failure		500		{object}	errorResponse
//	@Router			/api/v1/me/location [put]
func (s *Server) updateLocation(c *gin.Context) {
	userID := userIDFromContext(c)
	var req updateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if !validCoordinates(req.Latitude, req.Longitude) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid location"})
		return
	}

	now := time.Now().UTC()
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	_, err := s.db.DB.Collection("users").UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": bson.M{
		"latitude":          req.Latitude,
		"longitude":         req.Longitude,
		"locationAccuracy":  math.Max(0, req.Accuracy),
		"locationUpdatedAt": now,
		"updatedAt":         now,
	}})
	if err != nil {
		log.Printf("location update failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update location failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "saved", "locationUpdatedAt": now})
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

// discoverProfiles godoc
//
//	@Summary		List discoverable profiles
//	@Description	Returns a small Lounge-style list of profiles excluding the current user and blocked users.
//	@Tags			discover
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	discoverProfilesResponse
//	@Failure		401	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Router			/api/v1/discover/profiles [get]
func (s *Server) discoverProfiles(c *gin.Context) {
	userID := userIDFromContext(c)
	region := normalizeRegion(c.Query("region"))
	gender := normalizeGenderPreference(c.Query("gender"))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var viewer model.User
	_ = s.db.DB.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&viewer)

	blockedIDs, err := s.blockedUserIDs(ctx, userID)
	if err != nil {
		log.Printf("discover blocks failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "discover failed"})
		return
	}
	excluded := append(blockedIDs, userID)
	filter := bson.M{
		"_id":          bson.M{"$nin": excluded},
		"ageConfirmed": true,
	}
	if region != "global" {
		filter["region"] = region
	}
	if gender != "everyone" {
		filter["gender"] = gender
	}

	cursor, err := s.db.DB.Collection("users").Find(
		ctx,
		filter,
		options.Find().SetSort(bson.M{"updatedAt": -1}).SetLimit(20),
	)
	if err != nil {
		log.Printf("discover query failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "discover failed"})
		return
	}
	defer cursor.Close(ctx)

	users := make([]model.UserProfile, 0, 20)
	for cursor.Next(ctx) {
		var user model.User
		if err := cursor.Decode(&user); err != nil {
			log.Printf("discover decode failed user_id=%s err=%v", userID, err)
			continue
		}
		users = append(users, profileFromUserForViewer(user, viewer))
	}
	if err := cursor.Err(); err != nil {
		log.Printf("discover cursor failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "discover failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
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
	if req.Mode != model.MatchModeVideo && req.Mode != model.MatchModeVoice {
		log.Printf("match join invalid_mode user_id=%s mode=%q", userID, req.Mode)
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be video or voice"})
		return
	}
	req.Region = normalizeRegion(req.Region)
	req.Gender = normalizeGenderPreference(req.Gender)
	req.Language = normalizeLanguage(req.Language)
	requestedInterests := normalizeInterests(req.Interests)

	queueKey := "match:queue:v3:" + string(req.Mode) + ":" + req.Region + ":" + req.Language
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	user, err := s.userByID(ctx, userID)
	if err != nil {
		log.Printf("match join user failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "match failed"})
		return
	}
	isMember := isActiveMember(user, time.Now().UTC())
	usage, allowed, err := s.reserveMatchAttempt(ctx, userID, isMember)
	if err != nil {
		log.Printf("match quota failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "match quota failed"})
		return
	}
	if !allowed {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error":          "daily free match limit reached",
			"dailyLimit":     usage.DailyLimit,
			"dailyUsed":      usage.DailyUsed,
			"dailyRemaining": usage.DailyRemaining,
		})
		return
	}

	self := model.MatchTicket{
		UserID:           userID,
		Mode:             req.Mode,
		Region:           req.Region,
		GenderPreference: req.Gender,
		Language:         req.Language,
		Interests:        requestedInterests,
		CreatedAt:        time.Now().UTC(),
	}
	result, err := enqueueAndMatch(ctx, s.cache.Client, queueKey, self, isMember, func(candidateTicket model.MatchTicket) (bool, error) {
		candidate, err := s.userByID(ctx, candidateTicket.UserID)
		if err != nil {
			return false, err
		}
		return compatibleMatch(user, self, candidate, candidateTicket), nil
	})
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
		self.CreatedAt = time.Now().UTC()
		if err := queueTicket(ctx, s.cache.Client, queueKey, self, isMember); err != nil {
			log.Printf("match requeue failed user_id=%s mode=%s region=%s err=%v", userID, req.Mode, req.Region, err)
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
		partnerPriority := false
		if partnerUser, err := s.userByID(ctx, partner.UserID); err == nil {
			partnerPriority = isActiveMember(partnerUser, time.Now().UTC())
		}
		if err := queueTicket(ctx, s.cache.Client, queueKey, partner, partnerPriority); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "match failed"})
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "blocked match skipped, retry"})
		return
	}
	selfUser, err := s.userByID(ctx, userID)
	if err != nil {
		log.Printf("match self profile failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "match failed"})
		return
	}
	partnerUser, err := s.userByID(ctx, partner.UserID)
	if err != nil {
		log.Printf("match partner profile failed user_id=%s peer_id=%s err=%v", userID, partner.UserID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "match failed"})
		return
	}
	selfProfile := profileFromUserForViewer(selfUser, partnerUser)
	partnerProfile := profileFromUserForViewer(partnerUser, selfUser)
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

// listBlockedUsers godoc
//
//	@Summary		List blocked users
//	@Description	Returns users blocked by the current user.
//	@Tags			safety
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	blockedUsersResponse
//	@Failure		401	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Router			/api/v1/users/blocks [get]
func (s *Server) listBlockedUsers(c *gin.Context) {
	userID := userIDFromContext(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	cursor, err := s.db.DB.Collection("user_blocks").Find(ctx, bson.M{"userId": userID}, options.Find().SetSort(bson.M{"createdAt": -1}))
	if err != nil {
		log.Printf("user blocks list failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list blocked users failed"})
		return
	}
	defer cursor.Close(ctx)

	items := make([]blockedUserItem, 0)
	for cursor.Next(ctx) {
		var block model.UserBlock
		if err := cursor.Decode(&block); err != nil {
			log.Printf("user block decode failed user_id=%s err=%v", userID, err)
			continue
		}
		profile, err := s.userProfile(ctx, block.BlockedID)
		if err != nil {
			log.Printf("blocked user profile missing user_id=%s blocked_id=%s err=%v", userID, block.BlockedID, err)
			continue
		}
		items = append(items, blockedUserItem{
			User:      profile,
			CreatedAt: block.CreatedAt,
		})
	}
	if err := cursor.Err(); err != nil {
		log.Printf("user blocks cursor failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list blocked users failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": items})
}

// unblockUser godoc
//
//	@Summary		Unblock a user
//	@Description	Removes a block created by the current user.
//	@Tags			safety
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Target user id"
//	@Success		200	{object}	userActionResponse
//	@Failure		401	{object}	errorResponse
//	@Failure		500	{object}	errorResponse
//	@Router			/api/v1/users/{id}/block [delete]
func (s *Server) unblockUser(c *gin.Context) {
	userID := userIDFromContext(c)
	targetID := strings.TrimSpace(c.Param("id"))
	if targetID == "" || targetID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	if _, err := s.db.DB.Collection("user_blocks").DeleteOne(ctx, bson.M{"_id": userID + ":" + targetID}); err != nil {
		log.Printf("user unblock failed user_id=%s target_id=%s err=%v", userID, targetID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unblock failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "unblocked"})
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

func (s *Server) followUser(c *gin.Context) {
	userID := userIDFromContext(c)
	targetID := strings.TrimSpace(c.Param("id"))
	if targetID == "" || targetID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target"})
		return
	}
	now := time.Now().UTC()
	follow := model.UserFollow{
		ID:        userID + ":" + targetID,
		UserID:    userID,
		FollowID:  targetID,
		CreatedAt: now,
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := s.ensureUserExists(ctx, targetID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target not found"})
		return
	}
	if _, err := s.db.DB.Collection("user_follows").UpdateOne(ctx, bson.M{"_id": follow.ID}, bson.M{"$setOnInsert": follow}, updateUpsert()); err != nil {
		log.Printf("user follow failed user_id=%s target_id=%s err=%v", userID, targetID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "follow failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "followed"})
}

func (s *Server) listFollowedUsers(c *gin.Context) {
	userID := userIDFromContext(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	var viewer model.User
	_ = s.db.DB.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&viewer)
	cursor, err := s.db.DB.Collection("user_follows").Find(
		ctx,
		bson.M{"userId": userID},
		options.Find().SetSort(bson.M{"createdAt": -1}).SetLimit(50),
	)
	if err != nil {
		log.Printf("list follows failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list follows failed"})
		return
	}
	defer cursor.Close(ctx)

	users := make([]model.UserProfile, 0, 20)
	for cursor.Next(ctx) {
		var follow model.UserFollow
		if err := cursor.Decode(&follow); err != nil {
			continue
		}
		user, err := s.userByID(ctx, follow.FollowID)
		if err != nil {
			continue
		}
		users = append(users, profileFromUserForViewer(user, viewer))
	}
	if err := cursor.Err(); err != nil {
		log.Printf("list follows cursor failed user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "list follows failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (s *Server) unfollowUser(c *gin.Context) {
	userID := userIDFromContext(c)
	targetID := strings.TrimSpace(c.Param("id"))
	if targetID == "" || targetID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target"})
		return
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if _, err := s.db.DB.Collection("user_follows").DeleteOne(ctx, bson.M{"_id": userID + ":" + targetID}); err != nil {
		log.Printf("user unfollow failed user_id=%s target_id=%s err=%v", userID, targetID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unfollow failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "unfollowed"})
}

func (s *Server) sendDirectMessage(c *gin.Context) {
	userID := userIDFromContext(c)
	targetID := strings.TrimSpace(c.Param("id"))
	if targetID == "" || targetID == userID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target"})
		return
	}
	var req directMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	text := strings.TrimSpace(req.Text)
	if text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is empty"})
		return
	}
	if len([]rune(text)) > 240 {
		text = string([]rune(text)[:240])
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	if err := s.ensureUserExists(ctx, targetID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "target not found"})
		return
	}
	message := model.DirectMessage{
		ID:         newID(),
		SenderID:   userID,
		ReceiverID: targetID,
		Text:       text,
		CreatedAt:  time.Now().UTC(),
	}
	if _, err := s.db.DB.Collection("direct_messages").InsertOne(ctx, message); err != nil {
		log.Printf("direct message failed user_id=%s target_id=%s err=%v", userID, targetID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "message failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": message})
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
	return profileFromUser(user), nil
}

func profileFromUser(user model.User) model.UserProfile {
	return profileFromUserForViewer(user, model.User{})
}

func profileFromUserForViewer(user model.User, viewer model.User) model.UserProfile {
	region := normalizeRegion(user.Region)
	gender := normalizeProfileGender(user.Gender)
	distanceKm := distanceBetweenUsersKm(viewer, user)
	return model.UserProfile{
		ID:                  user.ID,
		DisplayName:         user.DisplayName,
		AvatarURL:           user.AvatarURL,
		Bio:                 user.Bio,
		Interests:           user.Interests,
		Region:              region,
		DistanceKm:          distanceKm,
		LocationUpdatedAt:   user.LocationUpdatedAt,
		Gender:              gender,
		Language:            normalizeLanguage(user.Language),
		TrustBadge:          len(user.Interests) >= 3 && user.AgeConfirmed,
		AgeConfirmed:        user.AgeConfirmed,
		MembershipPlan:      user.MembershipPlan,
		MembershipExpiresAt: user.MembershipExpiresAt,
		IsMember:            isActiveMember(user, time.Now().UTC()),
	}
}

func validCoordinates(latitude, longitude float64) bool {
	return !math.IsNaN(latitude) &&
		!math.IsNaN(longitude) &&
		!math.IsInf(latitude, 0) &&
		!math.IsInf(longitude, 0) &&
		latitude >= -90 &&
		latitude <= 90 &&
		longitude >= -180 &&
		longitude <= 180
}

func distanceBetweenUsersKm(viewer, user model.User) *float64 {
	if viewer.ID == "" || user.ID == "" || viewer.ID == user.ID {
		return nil
	}
	if viewer.Latitude == nil || viewer.Longitude == nil || user.Latitude == nil || user.Longitude == nil {
		return nil
	}
	distance := haversineKm(*viewer.Latitude, *viewer.Longitude, *user.Latitude, *user.Longitude)
	rounded := math.Round(distance*10) / 10
	return &rounded
}

func haversineKm(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusKm = 6371.0
	toRad := func(value float64) float64 {
		return value * math.Pi / 180
	}
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	rLat1 := toRad(lat1)
	rLat2 := toRad(lat2)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(rLat1)*math.Cos(rLat2)*math.Sin(dLon/2)*math.Sin(dLon/2)
	a = math.Min(1, math.Max(0, a))
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
}

func (s *Server) isBlockedEither(ctx context.Context, userID, peerID string) (bool, error) {
	count, err := s.db.DB.Collection("user_blocks").CountDocuments(ctx, bson.M{"$or": []bson.M{
		{"userId": userID, "blockedId": peerID},
		{"userId": peerID, "blockedId": userID},
	}})
	return count > 0, err
}

func (s *Server) blockedUserIDs(ctx context.Context, userID string) ([]string, error) {
	cursor, err := s.db.DB.Collection("user_blocks").Find(ctx, bson.M{"$or": []bson.M{
		{"userId": userID},
		{"blockedId": userID},
	}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	ids := make([]string, 0)
	for cursor.Next(ctx) {
		var block model.UserBlock
		if err := cursor.Decode(&block); err != nil {
			continue
		}
		if block.UserID == userID {
			ids = append(ids, block.BlockedID)
		} else {
			ids = append(ids, block.UserID)
		}
	}
	return ids, cursor.Err()
}

func (s *Server) ensureUserExists(ctx context.Context, userID string) error {
	return s.db.DB.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Err()
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
		return []string{"聊天", "电影", "音乐"}
	}
	return out
}

func normalizeRegion(value string) string {
	switch strings.TrimSpace(value) {
	case "nearby", "asia", "europe", "america":
		return strings.TrimSpace(value)
	default:
		return "global"
	}
}

func normalizeGenderPreference(value string) string {
	switch strings.TrimSpace(value) {
	case "female", "male":
		return strings.TrimSpace(value)
	default:
		return "everyone"
	}
}

func normalizeProfileGender(value string) string {
	switch strings.TrimSpace(value) {
	case "female", "male":
		return strings.TrimSpace(value)
	default:
		return "private"
	}
}

func normalizeLanguage(value string) string {
	switch strings.TrimSpace(value) {
	case "zh", "en", "ja", "ko", "es":
		return strings.TrimSpace(value)
	default:
		return "zh"
	}
}

func compatibleMatch(self model.User, selfTicket model.MatchTicket, candidate model.User, candidateTicket model.MatchTicket) bool {
	if normalizeLanguage(selfTicket.Language) != normalizeLanguage(candidateTicket.Language) {
		return false
	}
	if !genderAccepted(selfTicket.GenderPreference, candidate.Gender) {
		return false
	}
	if !genderAccepted(candidateTicket.GenderPreference, self.Gender) {
		return false
	}
	return hasSharedInterest(selfTicket.Interests, candidateTicket.Interests)
}

func genderAccepted(preference, profileGender string) bool {
	preference = normalizeGenderPreference(preference)
	profileGender = normalizeProfileGender(profileGender)
	return preference == "everyone" || preference == profileGender
}

func hasSharedInterest(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	seen := map[string]bool{}
	for _, item := range a {
		value := strings.TrimSpace(item)
		if value != "" {
			seen[value] = true
		}
	}
	for _, item := range b {
		if seen[strings.TrimSpace(item)] {
			return true
		}
	}
	return false
}

func defaultAvatarURL() string {
	return ""
}

func defaultDisplayName(userID string) string {
	suffix := strings.ToUpper(userID)
	if len(suffix) > 4 {
		suffix = suffix[:4]
	}
	if suffix == "" {
		return "星球旅人"
	}
	return "星球旅人 " + suffix
}

var defaultInterestPool = []string{
	"聊天",
	"电影",
	"音乐",
	"旅行",
	"美食",
	"运动",
	"游戏",
	"动漫",
	"摄影",
	"宠物",
	"读书",
	"咖啡",
	"健身",
	"语言交换",
	"科技",
	"深夜电台",
}

func defaultUserInterests(userID string) []string {
	if userID == "" {
		return []string{"聊天", "电影", "音乐"}
	}
	start := int(userID[0]) % len(defaultInterestPool)
	step := 5
	out := make([]string, 0, 4)
	seen := map[string]bool{}
	for i := 0; len(out) < 4 && i < len(defaultInterestPool); i++ {
		item := defaultInterestPool[(start+i*step)%len(defaultInterestPool)]
		if seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
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
