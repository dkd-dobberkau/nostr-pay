package api

import (
	"encoding/json"
	"net/http"

	nostrauth "github.com/nostr-pay/nostr-pay/internal/nostr"
	"github.com/nostr-pay/nostr-pay/internal/payment"
)

type createInvoiceRequest struct {
	AmountSats int64  `json:"amount_sats"`
	Memo       string `json:"memo"`
}

type createInvoiceResponse struct {
	PaymentID   string `json:"payment_id"`
	Bolt11      string `json:"bolt11"`
	PaymentHash string `json:"payment_hash"`
}

func (s *Server) handleCreateInvoice(w http.ResponseWriter, r *http.Request) {
	pubkey := nostrauth.PubkeyFromContext(r.Context())

	var req createInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.AmountSats <= 0 {
		http.Error(w, "amount must be positive", http.StatusBadRequest)
		return
	}

	result, err := s.paymentSvc.CreateInvoice(r.Context(), &payment.CreateInvoiceInput{
		ReceiverPubkey: pubkey,
		AmountSats:     req.AmountSats,
		Memo:           req.Memo,
	})
	if err != nil {
		http.Error(w, "failed to create invoice", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createInvoiceResponse{
		PaymentID:   result.PaymentID,
		Bolt11:      result.Bolt11,
		PaymentHash: result.PaymentHash,
	})
}

func (s *Server) handleGetPayment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing payment id", http.StatusBadRequest)
		return
	}

	p, err := s.paymentSvc.GetPayment(r.Context(), id)
	if err != nil {
		http.Error(w, "payment not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (s *Server) handlePaymentHistory(w http.ResponseWriter, r *http.Request) {
	pubkey := nostrauth.PubkeyFromContext(r.Context())

	payments, err := s.paymentSvc.ListPayments(r.Context(), pubkey, 50, 0)
	if err != nil {
		http.Error(w, "failed to list payments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payments)
}

type webhookPayload struct {
	PaymentHash string `json:"payment_hash"`
}

func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	var payload webhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	if err := s.paymentSvc.HandleWebhook(r.Context(), payload.PaymentHash); err != nil {
		http.Error(w, "webhook processing failed", http.StatusInternalServerError)
		return
	}

	s.wsHub.NotifyPayment(payload.PaymentHash, "paid")

	w.WriteHeader(http.StatusOK)
}
