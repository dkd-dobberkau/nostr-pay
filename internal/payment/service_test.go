package payment_test

import (
	"context"
	"testing"

	"github.com/nostr-pay/nostr-pay/internal/lnbits"
	"github.com/nostr-pay/nostr-pay/internal/payment"
	"github.com/nostr-pay/nostr-pay/internal/store"
)

// Mock LNbits client
type mockLNbits struct {
	invoiceResp *lnbits.CreateInvoiceResponse
	paymentResp *lnbits.PaymentStatus
}

func (m *mockLNbits) CreateInvoice(ctx context.Context, req *lnbits.CreateInvoiceRequest) (*lnbits.CreateInvoiceResponse, error) {
	return m.invoiceResp, nil
}

func (m *mockLNbits) CheckPayment(ctx context.Context, hash string) (*lnbits.PaymentStatus, error) {
	return m.paymentResp, nil
}

func TestCreateInvoice(t *testing.T) {
	db, _ := store.NewSQLite(":memory:")
	defer db.Close()

	mock := &mockLNbits{
		invoiceResp: &lnbits.CreateInvoiceResponse{
			PaymentHash:    "hash_001",
			PaymentRequest: "lnbc1000n1p...",
		},
	}

	svc := payment.NewService(db, mock, "http://localhost:8080")

	result, err := svc.CreateInvoice(context.Background(), &payment.CreateInvoiceInput{
		ReceiverPubkey: "npub_receiver",
		AmountSats:     1000,
		Memo:           "Test",
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}

	if result.Bolt11 != "lnbc1000n1p..." {
		t.Errorf("Bolt11 = %q, want %q", result.Bolt11, "lnbc1000n1p...")
	}
	if result.PaymentHash != "hash_001" {
		t.Errorf("PaymentHash = %q, want %q", result.PaymentHash, "hash_001")
	}

	// Verify payment was persisted
	p, err := db.GetPaymentByHash(context.Background(), "hash_001")
	if err != nil {
		t.Fatalf("GetPaymentByHash: %v", err)
	}
	if p.Status != "pending" {
		t.Errorf("Status = %q, want %q", p.Status, "pending")
	}
}

func TestHandleWebhook(t *testing.T) {
	db, _ := store.NewSQLite(":memory:")
	defer db.Close()

	mock := &mockLNbits{
		paymentResp: &lnbits.PaymentStatus{
			Paid:     true,
			Preimage: "preimage_abc",
		},
	}

	svc := payment.NewService(db, mock, "http://localhost:8080")

	// Create a pending payment first
	db.CreatePayment(context.Background(), &store.Payment{
		ID:             "pay_001",
		Bolt11:         "lnbc...",
		AmountSats:     500,
		ReceiverPubkey: "npub_receiver",
		PaymentHash:    "hash_webhook",
		Status:         "pending",
	})

	err := svc.HandleWebhook(context.Background(), "hash_webhook")
	if err != nil {
		t.Fatalf("HandleWebhook: %v", err)
	}

	p, _ := db.GetPayment(context.Background(), "pay_001")
	if p.Status != "paid" {
		t.Errorf("Status = %q, want %q", p.Status, "paid")
	}
}
