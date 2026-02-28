package api

import (
	"github.com/nostr-pay/nostr-pay/internal/payment"
	"github.com/nostr-pay/nostr-pay/internal/store"
)

type Server struct {
	store      store.Store
	paymentSvc *payment.Service
}

func NewServer(store store.Store, paymentSvc *payment.Service) *Server {
	return &Server{
		store:      store,
		paymentSvc: paymentSvc,
	}
}
