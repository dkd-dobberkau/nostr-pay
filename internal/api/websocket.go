package api

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHub struct {
	mu      sync.RWMutex
	clients map[string]map[*websocket.Conn]bool // paymentHash -> connections
}

func NewWSHub() *WSHub {
	return &WSHub{
		clients: make(map[string]map[*websocket.Conn]bool),
	}
}

func (h *WSHub) Subscribe(paymentHash string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[paymentHash] == nil {
		h.clients[paymentHash] = make(map[*websocket.Conn]bool)
	}
	h.clients[paymentHash][conn] = true
}

func (h *WSHub) Unsubscribe(paymentHash string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if conns, ok := h.clients[paymentHash]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(h.clients, paymentHash)
		}
	}
}

func (h *WSHub) NotifyPayment(paymentHash string, status string) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conns := h.clients[paymentHash]
	for conn := range conns {
		if err := conn.WriteJSON(map[string]string{
			"payment_hash": paymentHash,
			"status":       status,
		}); err != nil {
			slog.Error("ws write error", "error", err)
		}
	}
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	paymentHash := r.URL.Query().Get("payment_hash")
	if paymentHash == "" {
		http.Error(w, "missing payment_hash", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	s.wsHub.Subscribe(paymentHash, conn)
	defer s.wsHub.Unsubscribe(paymentHash, conn)

	// Keep connection alive until client disconnects
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
