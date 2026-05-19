package server

import (
	"context"
	"encoding/json"
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
}

type Client struct {
	userID string
	conn   *websocket.Conn
	send   chan SignalMessage
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string]*Client)}
}

func (h *Hub) Register(userID string, conn *websocket.Conn) *Client {
	client := &Client{
		userID: userID,
		conn:   conn,
		send:   make(chan SignalMessage, 16),
	}
	h.mu.Lock()
	h.clients[userID] = client
	h.mu.Unlock()
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
}

func (h *Hub) Notify(userID string, msg SignalMessage) {
	h.mu.RLock()
	client := h.clients[userID]
	h.mu.RUnlock()
	if client == nil {
		return
	}
	select {
	case client.send <- msg:
	default:
	}
}

func (c *Client) ReadLoop(ctx context.Context, hub *Hub) {
	for {
		_, payload, err := c.conn.Read(ctx)
		if err != nil {
			return
		}
		var msg SignalMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			continue
		}
		if msg.PeerID != "" {
			hub.Notify(msg.PeerID, msg)
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
			return
		}
	}
}

func mustJSON(v any) []byte {
	payload, _ := json.Marshal(v)
	return payload
}
