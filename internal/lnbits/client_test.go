package lnbits_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nostr-pay/nostr-pay/internal/lnbits"
)

func TestCreateInvoice(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/payments" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("X-Api-Key") != "test-invoice-key" {
			t.Errorf("missing or wrong API key")
		}

		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)

		if body["out"] != false {
			t.Errorf("out = %v, want false", body["out"])
		}
		if body["amount"] != float64(1000) {
			t.Errorf("amount = %v, want 1000", body["amount"])
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"payment_hash":    "hash_123",
			"payment_request": "lnbc1000n1p...",
			"checking_id":     "check_123",
		})
	}))
	defer server.Close()

	client := lnbits.NewClient(server.URL, "test-admin-key", "test-invoice-key")

	invoice, err := client.CreateInvoice(context.Background(), &lnbits.CreateInvoiceRequest{
		Amount: 1000,
		Memo:   "Test payment",
	})
	if err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}

	if invoice.PaymentHash != "hash_123" {
		t.Errorf("PaymentHash = %q, want %q", invoice.PaymentHash, "hash_123")
	}
	if invoice.PaymentRequest != "lnbc1000n1p..." {
		t.Errorf("PaymentRequest = %q, want %q", invoice.PaymentRequest, "lnbc1000n1p...")
	}
}

func TestCheckPayment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/payments/hash_123" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		json.NewEncoder(w).Encode(map[string]any{
			"paid":         true,
			"preimage":     "preimage_abc",
			"payment_hash": "hash_123",
			"amount":       1000000, // millisats
		})
	}))
	defer server.Close()

	client := lnbits.NewClient(server.URL, "test-admin-key", "test-invoice-key")

	status, err := client.CheckPayment(context.Background(), "hash_123")
	if err != nil {
		t.Fatalf("CheckPayment: %v", err)
	}

	if !status.Paid {
		t.Error("expected Paid = true")
	}
	if status.Preimage != "preimage_abc" {
		t.Errorf("Preimage = %q, want %q", status.Preimage, "preimage_abc")
	}
}

func TestGetWallet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/wallet" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		json.NewEncoder(w).Encode(map[string]any{
			"id":      "wallet_001",
			"name":    "Test Wallet",
			"balance": 500000, // millisats
		})
	}))
	defer server.Close()

	client := lnbits.NewClient(server.URL, "test-admin-key", "test-invoice-key")

	wallet, err := client.GetWallet(context.Background())
	if err != nil {
		t.Fatalf("GetWallet: %v", err)
	}

	if wallet.ID != "wallet_001" {
		t.Errorf("ID = %q, want %q", wallet.ID, "wallet_001")
	}
	if wallet.Balance != 500000 {
		t.Errorf("Balance = %d, want 500000", wallet.Balance)
	}
}
