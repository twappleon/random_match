package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
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
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("POST /api/v1/auth/anonymous", s.anonymousAuth)
	mux.HandleFunc("POST /api/v1/match/join", s.requireAuth(s.joinMatch))
	mux.HandleFunc("POST /api/v1/match/leave", s.requireAuth(s.leaveMatch))
	mux.HandleFunc("GET /api/v1/ws", s.ws)
	return s.cors(mux)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) anonymousAuth(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	user := model.User{
		ID:          newID(),
		DisplayName: "Guest",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	_, err := s.db.DB.Collection("users").InsertOne(ctx, user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create user failed")
		return
	}

	token, err := s.signToken(user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "sign token failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token": token,
		"user":  user,
	})
}

func (s *Server) joinMatch(w http.ResponseWriter, r *http.Request, userID string) {
	var req struct {
		Mode   model.MatchMode `json:"mode"`
		Region string          `json:"region"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if req.Mode != model.MatchModeVideo && req.Mode != model.MatchModeVoice {
		writeError(w, http.StatusBadRequest, "mode must be video or voice")
		return
	}
	if strings.TrimSpace(req.Region) == "" {
		req.Region = "global"
	}

	queueKey := "match:queue:" + string(req.Mode) + ":" + req.Region
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	partnerID, err := s.cache.Client.LPop(ctx, queueKey).Result()
	if errors.Is(err, redis.Nil) {
		ticket := model.MatchTicket{UserID: userID, Mode: req.Mode, Region: req.Region, CreatedAt: time.Now().UTC()}
		payload, _ := json.Marshal(ticket)
		if err := s.cache.Client.RPush(ctx, queueKey, payload).Err(); err != nil {
			writeError(w, http.StatusInternalServerError, "join queue failed")
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{"status": "waiting"})
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "match failed")
		return
	}

	var ticket model.MatchTicket
	if err := json.Unmarshal([]byte(partnerID), &ticket); err != nil || ticket.UserID == userID {
		writeJSON(w, http.StatusAccepted, map[string]any{"status": "waiting"})
		return
	}

	roomID := newID()
	s.hub.Notify(ticket.UserID, SignalMessage{Type: "matched", RoomID: roomID, PeerID: userID, Mode: string(req.Mode), Initiator: false})
	s.hub.Notify(userID, SignalMessage{Type: "matched", RoomID: roomID, PeerID: ticket.UserID, Mode: string(req.Mode), Initiator: true})
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "matched",
		"roomId":    roomID,
		"peerId":    ticket.UserID,
		"initiator": true,
	})
}

func (s *Server) leaveMatch(w http.ResponseWriter, r *http.Request, userID string) {
	// For production, store each user's active queue key separately and remove exactly that ticket.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	_ = s.cache.Client.Set(ctx, "match:left:"+userID, "1", 2*time.Minute).Err()
	writeJSON(w, http.StatusOK, map[string]string{"status": "left"})
}

func (s *Server) ws(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	userID, err := s.verifyToken(token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid token")
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: s.cfg.CORSOrigins,
	})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	client := s.hub.Register(userID, conn)
	defer s.hub.Unregister(userID, client)
	client.ReadLoop(r.Context(), s.hub)
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

func (s *Server) requireAuth(next func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		userID, err := s.verifyToken(raw)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := s.db.DB.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Err(); err != nil {
			writeError(w, http.StatusUnauthorized, "user not found")
			return
		}
		next(w, r, userID)
	}
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, allowed := range s.cfg.CORSOrigins {
			if origin == allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				break
			}
		}
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(b[:])
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
