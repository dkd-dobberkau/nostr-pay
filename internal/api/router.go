package api

import (
	"net/http"

	nostrauth "github.com/nostr-pay/nostr-pay/internal/nostr"
)

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// Public endpoints
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("POST /api/payments/webhook", s.handleWebhook)

	// Authenticated endpoints
	mux.Handle("POST /api/payments/invoice", nostrauth.AuthMiddleware(
		http.HandlerFunc(s.handleCreateInvoice),
	))
	mux.Handle("GET /api/payments/{id}", nostrauth.AuthMiddleware(
		http.HandlerFunc(s.handleGetPayment),
	))
	mux.Handle("GET /api/payments/history", nostrauth.AuthMiddleware(
		http.HandlerFunc(s.handlePaymentHistory),
	))

	// Apply global middleware
	var handler http.Handler = mux
	handler = corsMiddleware(handler)
	handler = loggingMiddleware(handler)

	return handler
}
