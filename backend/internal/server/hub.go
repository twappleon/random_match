package server

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type SignalMessage struct {
	Type      string          `json:"type"`
	RoomID    string          `json:"roomId,omitempty"`
	PeerID    string          `json:"peerId,omitempty"`
	Mode      string          `json:"mode,omitempty"`
	Initiator bool            `json:"initiator,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]*Client
	peers   map[string]string
}

type Client struct {
	userID string
	conn   *websocket.Conn
	send   chan SignalMessage
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]*Client),
		peers:   make(map[string]string),
	}
}

func (h *Hub) Register(userID string, conn *websocket.Conn) *Client {
	client := &Client{
		userID: userID,
		conn:   conn,
		send:   make(chan SignalMessage, 16),
	}
	h.mu.Lock()
	h.clients[userID] = client
	active := len(h.clients)
	h.mu.Unlock()
	log.Printf("hub register user_id=%s active_clients=%d", userID, active)
	go client.WriteLoop()
	return client
}

func (h *Hub) Unregister(userID string, client *Client) {
	h.mu.Lock()
	if h.clients[userID] == client {
		delete(h.clients, userID)
		close(client.send)
	}
	h.mu.Unlock()
	peerID := h.Unpair(userID)
	h.mu.RLock()
	active := len(h.clients)
	h.mu.RUnlock()
	log.Printf("hub unregister user_id=%s active_clients=%d", userID, active)
	if peerID != "" {
		h.Notify(peerID, SignalMessage{Type: "peer-left", PeerID: userID})
	}
}

func (h *Hub) Notify(userID string, msg SignalMessage) {
	h.mu.RLock()
	client := h.clients[userID]
	h.mu.RUnlock()
	if client == nil {
		log.Printf("signal drop target_user_id=%s type=%s reason=no_client", userID, msg.Type)
		return
	}
	select {
	case client.send <- msg:
		log.Printf("signal enqueue target_user_id=%s peer_id=%s type=%s room_id=%s", userID, msg.PeerID, msg.Type, msg.RoomID)
	default:
		log.Printf("signal drop target_user_id=%s peer_id=%s type=%s reason=send_queue_full", userID, msg.PeerID, msg.Type)
	}
}

func (h *Hub) IsOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[userID] != nil
}

func (h *Hub) Pair(userID, peerID string) {
	h.mu.Lock()
	h.peers[userID] = peerID
	h.peers[peerID] = userID
	h.mu.Unlock()
}

func (h *Hub) Unpair(userID string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	peerID := h.peers[userID]
	delete(h.peers, userID)
	if peerID != "" && h.peers[peerID] == userID {
		delete(h.peers, peerID)
	}
	return peerID
}

func (c *Client) ReadLoop(ctx context.Context, hub *Hub) {
	for {
		_, payload, err := c.conn.Read(ctx)
		if err != nil {
			log.Printf("signal read end user_id=%s err=%v", c.userID, err)
			return
		}
		var msg SignalMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			log.Printf("signal invalid_json user_id=%s err=%v", c.userID, err)
			continue
		}
		if msg.PeerID != "" {
			targetUserID := msg.PeerID
			msg.PeerID = c.userID
			log.Printf("signal forward from_user_id=%s to_user_id=%s type=%s room_id=%s", c.userID, targetUserID, msg.Type, msg.RoomID)
			hub.Notify(targetUserID, msg)
		}
	}
}

func (c *Client) WriteLoop() {
	ctx := context.Background()
	for msg := range c.send {
		writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := c.conn.Write(writeCtx, websocket.MessageText, mustJSON(msg))
		cancel()
		if err != nil {
			log.Printf("signal write end user_id=%s err=%v", c.userID, err)
			return
		}
		log.Printf("signal sent user_id=%s peer_id=%s type=%s room_id=%s", c.userID, msg.PeerID, msg.Type, msg.RoomID)
	}
}

func mustJSON(v any) []byte {
	payload, _ := json.Marshal(v)
	return payload
}
