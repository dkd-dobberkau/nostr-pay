package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nostr-pay/nostr-pay/internal/store"
)

func setupTestDB(t *testing.T) store.Store {
	t.Helper()
	db, err := store.NewSQLite(":memory:")
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCreateAndGetUser(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	user := &store.User{
		Pubkey:         "npub_test_123",
		IsMerchant:     false,
		LNbitsWalletID: "wallet_abc",
	}

	if err := db.CreateUser(ctx, user); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	got, err := db.GetUser(ctx, "npub_test_123")
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}

	if got.Pubkey != "npub_test_123" {
		t.Errorf("Pubkey = %q, want %q", got.Pubkey, "npub_test_123")
	}
	if got.LNbitsWalletID != "wallet_abc" {
		t.Errorf("LNbitsWalletID = %q, want %q", got.LNbitsWalletID, "wallet_abc")
	}
}

func TestCreateAndGetPayment(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	payment := &store.Payment{
		ID:             "pay_001",
		Bolt11:         "lnbc100n1p...",
		AmountSats:     100,
		Memo:           "Test payment",
		ReceiverPubkey: "npub_receiver",
		PaymentHash:    "hash_abc",
		Status:         "pending",
	}

	if err := db.CreatePayment(ctx, payment); err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}

	got, err := db.GetPayment(ctx, "pay_001")
	if err != nil {
		t.Fatalf("GetPayment: %v", err)
	}

	if got.AmountSats != 100 {
		t.Errorf("AmountSats = %d, want 100", got.AmountSats)
	}
	if got.Status != "pending" {
		t.Errorf("Status = %q, want %q", got.Status, "pending")
	}
}

func TestUpdatePaymentStatus(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	payment := &store.Payment{
		ID:             "pay_002",
		Bolt11:         "lnbc200n1p...",
		AmountSats:     200,
		ReceiverPubkey: "npub_receiver",
		PaymentHash:    "hash_def",
		Status:         "pending",
	}
	db.CreatePayment(ctx, payment)

	now := time.Now()
	err := db.UpdatePaymentStatus(ctx, "pay_002", "paid", &now)
	if err != nil {
		t.Fatalf("UpdatePaymentStatus: %v", err)
	}

	got, _ := db.GetPayment(ctx, "pay_002")
	if got.Status != "paid" {
		t.Errorf("Status = %q, want %q", got.Status, "paid")
	}
	if got.SettledAt == nil {
		t.Error("SettledAt should not be nil")
	}
}

func TestGetPaymentByHash(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	payment := &store.Payment{
		ID:             "pay_003",
		Bolt11:         "lnbc300n1p...",
		AmountSats:     300,
		ReceiverPubkey: "npub_receiver",
		PaymentHash:    "unique_hash",
		Status:         "pending",
	}
	db.CreatePayment(ctx, payment)

	got, err := db.GetPaymentByHash(ctx, "unique_hash")
	if err != nil {
		t.Fatalf("GetPaymentByHash: %v", err)
	}
	if got.ID != "pay_003" {
		t.Errorf("ID = %q, want %q", got.ID, "pay_003")
	}
}

func TestListPaymentsByUser(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		db.CreatePayment(ctx, &store.Payment{
			ID:             fmt.Sprintf("pay_%d", i),
			Bolt11:         "lnbc...",
			AmountSats:     int64(100 * (i + 1)),
			ReceiverPubkey: "npub_user",
			PaymentHash:    fmt.Sprintf("hash_%d", i),
			Status:         "paid",
		})
	}

	payments, err := db.ListPaymentsByUser(ctx, "npub_user", 10, 0)
	if err != nil {
		t.Fatalf("ListPaymentsByUser: %v", err)
	}
	if len(payments) != 3 {
		t.Errorf("got %d payments, want 3", len(payments))
	}
}
