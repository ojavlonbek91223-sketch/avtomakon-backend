package websocket

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Event — server'dan klientga yuboriladigan event.
type Event struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

// Client — bitta WebSocket ulanish.
type Client struct {
	UserID uuid.UUID
	Send   chan []byte
}

// Hub — barcha aktiv klientlar.
type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[*Client]struct{} // user_id → {client1, client2 (turli qurilmalar)}
	logger  *zap.Logger
}

func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		clients: make(map[uuid.UUID]map[*Client]struct{}),
		logger:  logger,
	}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c.UserID]; !ok {
		h.clients[c.UserID] = make(map[*Client]struct{})
	}
	h.clients[c.UserID][c] = struct{}{}
	h.logger.Debug("ws ulandi", zap.String("user_id", c.UserID.String()))
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[c.UserID]; ok {
		delete(conns, c)
		close(c.Send)
		if len(conns) == 0 {
			delete(h.clients, c.UserID)
		}
	}
}

// SendToUser — bitta foydalanuvchining barcha qurilmalariga eventni yuboradi.
func (h *Hub) SendToUser(userID uuid.UUID, event Event) {
	h.mu.RLock()
	conns := h.clients[userID]
	if len(conns) == 0 {
		h.mu.RUnlock()
		return
	}
	clients := make([]*Client, 0, len(conns))
	for c := range conns {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	data, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("ws marshal", zap.Error(err))
		return
	}

	for _, c := range clients {
		select {
		case c.Send <- data:
		default:
			// Klient buffer to'la — yo'q qilamiz
			h.Unregister(c)
		}
	}
}

// SendToUsers — bir nechta foydalanuvchiga yuborish.
func (h *Hub) SendToUsers(userIDs []uuid.UUID, event Event) {
	for _, uid := range userIDs {
		h.SendToUser(uid, event)
	}
}

// IsOnline — foydalanuvchi WebSocket bilan ulanganmi.
func (h *Hub) IsOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns, ok := h.clients[userID]
	return ok && len(conns) > 0
}
