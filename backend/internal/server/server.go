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
	v1.GET("/ws", s.ws)

	auth := v1.Group("")
	auth.Use(s.requireAuth())
	auth.GET("/auth/session", s.authSession)
	auth.POST("/match/join", s.joinMatch)
	auth.POST("/match/leave", s.leaveMatch)

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
		ID:          newID(),
		DisplayName: "Guest",
		CreatedAt:   now,
		UpdatedAt:   now,
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

// joinMatch godoc
//
//	@Summary		Join matchmaking queue
//	@Description	Adds the current user to a voice or video matchmaking queue. Returns 202 while waiting or 200 when matched.
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
	roomID := newID()
	log.Printf("match paired room_id=%s user_id=%s peer_id=%s mode=%s region=%s", roomID, userID, partner.UserID, req.Mode, req.Region)
	s.hub.Notify(partner.UserID, SignalMessage{Type: "matched", RoomID: roomID, PeerID: userID, Mode: string(req.Mode), Initiator: false})
	s.hub.Notify(userID, SignalMessage{Type: "matched", RoomID: roomID, PeerID: partner.UserID, Mode: string(req.Mode), Initiator: true})
	c.JSON(http.StatusOK, gin.H{
		"status":    "matched",
		"roomId":    roomID,
		"peerId":    partner.UserID,
		"initiator": true,
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
//	@Router			/api/v1/match/leave [post]
func (s *Server) leaveMatch(c *gin.Context) {
	userID := userIDFromContext(c)
	// For production, store each user's active queue key separately and remove exactly that ticket.
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	_ = s.cache.Client.Set(ctx, "match:left:"+userID, "1", 2*time.Minute).Err()
	log.Printf("match leave user_id=%s", userID)
	c.JSON(http.StatusOK, gin.H{"status": "left"})
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
	defer func() {
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
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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
