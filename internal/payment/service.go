package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/nostr-pay/nostr-pay/internal/lnbits"
	"github.com/nostr-pay/nostr-pay/internal/store"
)

type LNbitsClient interface {
	CreateInvoice(ctx context.Context, req *lnbits.CreateInvoiceRequest) (*lnbits.CreateInvoiceResponse, error)
	CheckPayment(ctx context.Context, hash string) (*lnbits.PaymentStatus, error)
}

type Service struct {
	store   store.Store
	lnbits  LNbitsClient
	baseURL string
}

func NewService(store store.Store, lnbits LNbitsClient, baseURL string) *Service {
	return &Service{
		store:   store,
		lnbits:  lnbits,
		baseURL: baseURL,
	}
}

type CreateInvoiceInput struct {
	ReceiverPubkey string
	SenderPubkey   string
	AmountSats     int64
	Memo           string
}

type CreateInvoiceResult struct {
	PaymentID   string
	Bolt11      string
	PaymentHash string
}

func (s *Service) CreateInvoice(ctx context.Context, input *CreateInvoiceInput) (*CreateInvoiceResult, error) {
	webhookURL := s.baseURL + "/api/payments/webhook"

	resp, err := s.lnbits.CreateInvoice(ctx, &lnbits.CreateInvoiceRequest{
		Amount:  input.AmountSats,
		Memo:    input.Memo,
		Webhook: webhookURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create lnbits invoice: %w", err)
	}

	paymentID := fmt.Sprintf("pay_%d", time.Now().UnixNano())

	payment := &store.Payment{
		ID:             paymentID,
		Bolt11:         resp.PaymentRequest,
		AmountSats:     input.AmountSats,
		Memo:           input.Memo,
		SenderPubkey:   input.SenderPubkey,
		ReceiverPubkey: input.ReceiverPubkey,
		PaymentHash:    resp.PaymentHash,
		Status:         "pending",
	}

	if err := s.store.CreatePayment(ctx, payment); err != nil {
		return nil, fmt.Errorf("store payment: %w", err)
	}

	return &CreateInvoiceResult{
		PaymentID:   paymentID,
		Bolt11:      resp.PaymentRequest,
		PaymentHash: resp.PaymentHash,
	}, nil
}

func (s *Service) HandleWebhook(ctx context.Context, paymentHash string) error {
	// Verify payment with LNbits
	status, err := s.lnbits.CheckPayment(ctx, paymentHash)
	if err != nil {
		return fmt.Errorf("check payment: %w", err)
	}

	if !status.Paid {
		return nil // Not paid yet, ignore
	}

	p, err := s.store.GetPaymentByHash(ctx, paymentHash)
	if err != nil {
		return fmt.Errorf("get payment by hash: %w", err)
	}

	now := time.Now()
	if err := s.store.UpdatePaymentStatus(ctx, p.ID, "paid", &now); err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}

	return nil
}

func (s *Service) GetPayment(ctx context.Context, id string) (*store.Payment, error) {
	return s.store.GetPayment(ctx, id)
}

func (s *Service) ListPayments(ctx context.Context, pubkey string, limit, offset int) ([]*store.Payment, error) {
	return s.store.ListPaymentsByUser(ctx, pubkey, limit, offset)
}
