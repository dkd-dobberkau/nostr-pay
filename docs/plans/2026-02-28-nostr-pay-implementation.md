# nostr-pay Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a Nostr-based instant payment platform with Lightning Network, supporting both merchant POS and P2P payments via QR codes.

**Architecture:** Split architecture — PWA (React/TS) handles Nostr identity and UI, Go backend handles LNbits integration and payment tracking. NIP-98 for API auth, NIP-46 for login, NIP-44 for encrypted receipts.

**Tech Stack:** Go 1.22+ (stdlib router + chi middleware), SQLite (ncruces/go-sqlite3), LNbits API, React 19/TypeScript/Vite, nostr-tools, TailwindCSS, qrcode.react, html5-qrcode

**Design Doc:** `docs/plans/2026-02-28-nostr-pay-design.md`

---

## Phase 1: Go Backend Foundation

### Task 1: Initialize Go Module & Project Structure

**Files:**
- Create: `go.mod`
- Create: `cmd/server/main.go`
- Create: `.gitignore`
- Create: `.env.example`

**Step 1: Initialize Go module**

Run:
```bash
cd /Users/olivier/Versioncontrol/local/nostr-pay
go mod init github.com/nostr-pay/nostr-pay
```

**Step 2: Create .gitignore**

```gitignore
# Binary
nostr-pay
/cmd/server/server

# Environment
.env
*.db
*.db-wal
*.db-shm

# Data
/data/

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store

# Node
web/node_modules/
web/dist/
```

**Step 3: Create .env.example**

```env
# LNbits
LNBITS_URL=https://lnbits.example.com
LNBITS_ADMIN_KEY=your-admin-key-here
LNBITS_INVOICE_KEY=your-invoice-key-here

# Server
SERVER_ADDR=:8080
DB_PATH=./data/nostr-pay.db

# Nostr Relays (comma-separated)
NOSTR_RELAYS=wss://relay.damus.io,wss://nos.lol

# CORS
CORS_ORIGINS=http://localhost:3000
```

**Step 4: Create minimal main.go**

```go
// cmd/server/main.go
package main

import (
	"fmt"
	"log/slog"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("nostr-pay server starting")
	fmt.Println("nostr-pay server")
}
```

**Step 5: Create directory structure**

Run:
```bash
mkdir -p internal/{api,lnbits,nostr,payment,merchant,store/migrations}
mkdir -p data
touch internal/api/.gitkeep internal/lnbits/.gitkeep internal/nostr/.gitkeep
touch internal/payment/.gitkeep internal/merchant/.gitkeep internal/store/.gitkeep
```

**Step 6: Verify build**

Run: `go build ./cmd/server/`
Expected: Compiles without errors

**Step 7: Commit**

```bash
git add -A
git commit -m "feat: initialize Go project structure"
```

---

### Task 2: Configuration Loading

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write the failing test**

```go
// internal/config/config_test.go
package config_test

import (
	"testing"

	"github.com/nostr-pay/nostr-pay/internal/config"
)

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("LNBITS_URL", "https://lnbits.test.com")
	t.Setenv("LNBITS_ADMIN_KEY", "admin-key-123")
	t.Setenv("LNBITS_INVOICE_KEY", "invoice-key-456")
	t.Setenv("SERVER_ADDR", ":9090")
	t.Setenv("DB_PATH", "/tmp/test.db")
	t.Setenv("NOSTR_RELAYS", "wss://relay1.com,wss://relay2.com")
	t.Setenv("CORS_ORIGINS", "http://localhost:3000,http://localhost:5173")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.LNbitsURL != "https://lnbits.test.com" {
		t.Errorf("LNbitsURL = %q, want %q", cfg.LNbitsURL, "https://lnbits.test.com")
	}
	if cfg.LNbitsAdminKey != "admin-key-123" {
		t.Errorf("LNbitsAdminKey = %q, want %q", cfg.LNbitsAdminKey, "admin-key-123")
	}
	if cfg.ServerAddr != ":9090" {
		t.Errorf("ServerAddr = %q, want %q", cfg.ServerAddr, ":9090")
	}
	if len(cfg.NostrRelays) != 2 {
		t.Errorf("NostrRelays len = %d, want 2", len(cfg.NostrRelays))
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("LNBITS_URL", "https://lnbits.test.com")
	t.Setenv("LNBITS_ADMIN_KEY", "key")
	t.Setenv("LNBITS_INVOICE_KEY", "key")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ServerAddr != ":8080" {
		t.Errorf("default ServerAddr = %q, want %q", cfg.ServerAddr, ":8080")
	}
	if cfg.DBPath != "./data/nostr-pay.db" {
		t.Errorf("default DBPath = %q, want %q", cfg.DBPath, "./data/nostr-pay.db")
	}
}

func TestLoadMissingRequired(t *testing.T) {
	// Clear all env vars
	t.Setenv("LNBITS_URL", "")
	t.Setenv("LNBITS_ADMIN_KEY", "")
	t.Setenv("LNBITS_INVOICE_KEY", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing required config")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -v`
Expected: FAIL — package doesn't exist

**Step 3: Write minimal implementation**

```go
// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	LNbitsURL        string
	LNbitsAdminKey   string
	LNbitsInvoiceKey string
	ServerAddr       string
	DBPath           string
	NostrRelays      []string
	CORSOrigins      []string
}

func Load() (*Config, error) {
	cfg := &Config{
		LNbitsURL:        os.Getenv("LNBITS_URL"),
		LNbitsAdminKey:   os.Getenv("LNBITS_ADMIN_KEY"),
		LNbitsInvoiceKey: os.Getenv("LNBITS_INVOICE_KEY"),
		ServerAddr:       getEnvDefault("SERVER_ADDR", ":8080"),
		DBPath:           getEnvDefault("DB_PATH", "./data/nostr-pay.db"),
	}

	if relays := os.Getenv("NOSTR_RELAYS"); relays != "" {
		cfg.NostrRelays = strings.Split(relays, ",")
	}

	if origins := os.Getenv("CORS_ORIGINS"); origins != "" {
		cfg.CORSOrigins = strings.Split(origins, ",")
	}

	if cfg.LNbitsURL == "" || cfg.LNbitsAdminKey == "" || cfg.LNbitsInvoiceKey == "" {
		return nil, fmt.Errorf("required config missing: LNBITS_URL, LNBITS_ADMIN_KEY, and LNBITS_INVOICE_KEY must be set")
	}

	return cfg, nil
}

func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/config/ -v`
Expected: PASS (3 tests)

**Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add configuration loading from environment"
```

---

### Task 3: SQLite Database Layer

**Files:**
- Create: `internal/store/store.go`
- Create: `internal/store/sqlite.go`
- Create: `internal/store/sqlite_test.go`

**Step 1: Install SQLite driver**

Run:
```bash
go get github.com/ncruces/go-sqlite3
go get github.com/ncruces/go-sqlite3/driver
go get github.com/ncruces/go-sqlite3/embed
```

**Step 2: Write the store interface**

```go
// internal/store/store.go
package store

import (
	"context"
	"time"
)

type User struct {
	Pubkey         string
	IsMerchant     bool
	LNbitsWalletID string
	CreatedAt      time.Time
}

type Payment struct {
	ID             string
	Bolt11         string
	AmountSats     int64
	Memo           string
	SenderPubkey   string
	ReceiverPubkey string
	PaymentHash    string
	Status         string // pending, paid, expired
	CreatedAt      time.Time
	SettledAt      *time.Time
}

type MerchantDailyStats struct {
	Pubkey           string
	Date             string
	TotalSats        int64
	TransactionCount int
}

type Store interface {
	// Users
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, pubkey string) (*User, error)
	UpdateUserMerchant(ctx context.Context, pubkey string, isMerchant bool) error

	// Payments
	CreatePayment(ctx context.Context, payment *Payment) error
	GetPayment(ctx context.Context, id string) (*Payment, error)
	GetPaymentByHash(ctx context.Context, paymentHash string) (*Payment, error)
	UpdatePaymentStatus(ctx context.Context, id string, status string, settledAt *time.Time) error
	ListPaymentsByUser(ctx context.Context, pubkey string, limit, offset int) ([]*Payment, error)

	// Merchant
	GetMerchantDailyStats(ctx context.Context, pubkey string, date string) (*MerchantDailyStats, error)
	ListMerchantTransactions(ctx context.Context, pubkey string, limit, offset int) ([]*Payment, error)

	Close() error
}
```

**Step 3: Write the failing tests**

```go
// internal/store/sqlite_test.go
package store_test

import (
	"context"
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
```

Note: Add `"fmt"` to imports for the last test.

**Step 4: Run tests to verify they fail**

Run: `go test ./internal/store/ -v`
Expected: FAIL — NewSQLite not defined

**Step 5: Write SQLite implementation**

```go
// internal/store/sqlite.go
package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

type sqliteStore struct {
	db *sql.DB
}

func NewSQLite(dsn string) (Store, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Enable WAL mode and set busy timeout
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	s := &sqliteStore{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return s, nil
}

func (s *sqliteStore) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		pubkey TEXT PRIMARY KEY,
		is_merchant BOOLEAN DEFAULT FALSE,
		lnbits_wallet_id TEXT DEFAULT '',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS payments (
		id TEXT PRIMARY KEY,
		bolt11 TEXT NOT NULL,
		amount_sats INTEGER NOT NULL,
		memo TEXT DEFAULT '',
		sender_pubkey TEXT DEFAULT '',
		receiver_pubkey TEXT NOT NULL,
		payment_hash TEXT UNIQUE,
		status TEXT DEFAULT 'pending',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		settled_at TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS merchant_daily_stats (
		pubkey TEXT,
		date TEXT,
		total_sats INTEGER DEFAULT 0,
		transaction_count INTEGER DEFAULT 0,
		PRIMARY KEY (pubkey, date)
	);

	CREATE INDEX IF NOT EXISTS idx_payments_receiver ON payments(receiver_pubkey);
	CREATE INDEX IF NOT EXISTS idx_payments_sender ON payments(sender_pubkey);
	CREATE INDEX IF NOT EXISTS idx_payments_hash ON payments(payment_hash);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *sqliteStore) Close() error {
	return s.db.Close()
}

// Users

func (s *sqliteStore) CreateUser(ctx context.Context, user *User) error {
	_, err := s.db.ExecContext(ctx,
		"INSERT INTO users (pubkey, is_merchant, lnbits_wallet_id) VALUES (?, ?, ?)",
		user.Pubkey, user.IsMerchant, user.LNbitsWalletID,
	)
	return err
}

func (s *sqliteStore) GetUser(ctx context.Context, pubkey string) (*User, error) {
	u := &User{}
	err := s.db.QueryRowContext(ctx,
		"SELECT pubkey, is_merchant, lnbits_wallet_id, created_at FROM users WHERE pubkey = ?",
		pubkey,
	).Scan(&u.Pubkey, &u.IsMerchant, &u.LNbitsWalletID, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *sqliteStore) UpdateUserMerchant(ctx context.Context, pubkey string, isMerchant bool) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE users SET is_merchant = ? WHERE pubkey = ?",
		isMerchant, pubkey,
	)
	return err
}

// Payments

func (s *sqliteStore) CreatePayment(ctx context.Context, payment *Payment) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO payments (id, bolt11, amount_sats, memo, sender_pubkey, receiver_pubkey, payment_hash, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		payment.ID, payment.Bolt11, payment.AmountSats, payment.Memo,
		payment.SenderPubkey, payment.ReceiverPubkey, payment.PaymentHash, payment.Status,
	)
	return err
}

func (s *sqliteStore) GetPayment(ctx context.Context, id string) (*Payment, error) {
	p := &Payment{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, bolt11, amount_sats, memo, sender_pubkey, receiver_pubkey,
		        payment_hash, status, created_at, settled_at
		 FROM payments WHERE id = ?`,
		id,
	).Scan(&p.ID, &p.Bolt11, &p.AmountSats, &p.Memo, &p.SenderPubkey,
		&p.ReceiverPubkey, &p.PaymentHash, &p.Status, &p.CreatedAt, &p.SettledAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *sqliteStore) GetPaymentByHash(ctx context.Context, paymentHash string) (*Payment, error) {
	p := &Payment{}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, bolt11, amount_sats, memo, sender_pubkey, receiver_pubkey,
		        payment_hash, status, created_at, settled_at
		 FROM payments WHERE payment_hash = ?`,
		paymentHash,
	).Scan(&p.ID, &p.Bolt11, &p.AmountSats, &p.Memo, &p.SenderPubkey,
		&p.ReceiverPubkey, &p.PaymentHash, &p.Status, &p.CreatedAt, &p.SettledAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *sqliteStore) UpdatePaymentStatus(ctx context.Context, id string, status string, settledAt *time.Time) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE payments SET status = ?, settled_at = ? WHERE id = ?",
		status, settledAt, id,
	)
	return err
}

func (s *sqliteStore) ListPaymentsByUser(ctx context.Context, pubkey string, limit, offset int) ([]*Payment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, bolt11, amount_sats, memo, sender_pubkey, receiver_pubkey,
		        payment_hash, status, created_at, settled_at
		 FROM payments
		 WHERE receiver_pubkey = ? OR sender_pubkey = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		pubkey, pubkey, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*Payment
	for rows.Next() {
		p := &Payment{}
		if err := rows.Scan(&p.ID, &p.Bolt11, &p.AmountSats, &p.Memo, &p.SenderPubkey,
			&p.ReceiverPubkey, &p.PaymentHash, &p.Status, &p.CreatedAt, &p.SettledAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}

// Merchant

func (s *sqliteStore) GetMerchantDailyStats(ctx context.Context, pubkey string, date string) (*MerchantDailyStats, error) {
	stats := &MerchantDailyStats{}
	err := s.db.QueryRowContext(ctx,
		"SELECT pubkey, date, total_sats, transaction_count FROM merchant_daily_stats WHERE pubkey = ? AND date = ?",
		pubkey, date,
	).Scan(&stats.Pubkey, &stats.Date, &stats.TotalSats, &stats.TransactionCount)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (s *sqliteStore) ListMerchantTransactions(ctx context.Context, pubkey string, limit, offset int) ([]*Payment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, bolt11, amount_sats, memo, sender_pubkey, receiver_pubkey,
		        payment_hash, status, created_at, settled_at
		 FROM payments
		 WHERE receiver_pubkey = ? AND status = 'paid'
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		pubkey, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*Payment
	for rows.Next() {
		p := &Payment{}
		if err := rows.Scan(&p.ID, &p.Bolt11, &p.AmountSats, &p.Memo, &p.SenderPubkey,
			&p.ReceiverPubkey, &p.PaymentHash, &p.Status, &p.CreatedAt, &p.SettledAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}
```

**Step 6: Run tests to verify they pass**

Run: `go test ./internal/store/ -v`
Expected: PASS (all 5 tests)

**Step 7: Commit**

```bash
git add internal/store/ go.mod go.sum
git commit -m "feat: add SQLite store with users, payments, and merchant stats"
```

---

### Task 4: LNbits API Client

**Files:**
- Create: `internal/lnbits/client.go`
- Create: `internal/lnbits/client_test.go`

**Step 1: Write the failing tests**

```go
// internal/lnbits/client_test.go
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
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/lnbits/ -v`
Expected: FAIL — package doesn't exist

**Step 3: Write implementation**

```go
// internal/lnbits/client.go
package lnbits

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	baseURL    string
	adminKey   string
	invoiceKey string
	httpClient *http.Client
}

func NewClient(baseURL, adminKey, invoiceKey string) *Client {
	return &Client{
		baseURL:    baseURL,
		adminKey:   adminKey,
		invoiceKey: invoiceKey,
		httpClient: &http.Client{},
	}
}

// Request/Response types

type CreateInvoiceRequest struct {
	Amount  int64  `json:"amount"`
	Memo    string `json:"memo,omitempty"`
	Webhook string `json:"webhook,omitempty"`
}

type CreateInvoiceResponse struct {
	PaymentHash    string `json:"payment_hash"`
	PaymentRequest string `json:"payment_request"`
	CheckingID     string `json:"checking_id"`
}

type PaymentStatus struct {
	Paid        bool   `json:"paid"`
	Preimage    string `json:"preimage"`
	PaymentHash string `json:"payment_hash"`
	Amount      int64  `json:"amount"`
}

type Wallet struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Balance int64  `json:"balance"`
}

// Methods

func (c *Client) CreateInvoice(ctx context.Context, req *CreateInvoiceRequest) (*CreateInvoiceResponse, error) {
	body := map[string]any{
		"out":    false,
		"amount": req.Amount,
		"memo":   req.Memo,
		"unit":   "sat",
	}
	if req.Webhook != "" {
		body["webhook"] = req.Webhook
	}

	resp, err := c.doRequest(ctx, http.MethodPost, "/api/v1/payments", c.invoiceKey, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("lnbits: create invoice returned status %d", resp.StatusCode)
	}

	var result CreateInvoiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("lnbits: decode response: %w", err)
	}
	return &result, nil
}

func (c *Client) CheckPayment(ctx context.Context, paymentHash string) (*PaymentStatus, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/payments/"+paymentHash, c.invoiceKey, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lnbits: check payment returned status %d", resp.StatusCode)
	}

	var result PaymentStatus
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("lnbits: decode response: %w", err)
	}
	return &result, nil
}

func (c *Client) GetWallet(ctx context.Context) (*Wallet, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/wallet", c.invoiceKey, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lnbits: get wallet returned status %d", resp.StatusCode)
	}

	var result Wallet
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("lnbits: decode response: %w", err)
	}
	return &result, nil
}

func (c *Client) doRequest(ctx context.Context, method, path, apiKey string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("lnbits: marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("lnbits: create request: %w", err)
	}

	req.Header.Set("X-Api-Key", apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/lnbits/ -v`
Expected: PASS (3 tests)

**Step 5: Commit**

```bash
git add internal/lnbits/
git commit -m "feat: add LNbits API client with invoice, payment, and wallet support"
```

---

### Task 5: NIP-98 Auth Middleware

**Files:**
- Create: `internal/nostr/verify.go`
- Create: `internal/nostr/verify_test.go`

**Step 1: Install go-nostr**

Run:
```bash
go get github.com/nbd-wtf/go-nostr
```

**Step 2: Write the failing tests**

```go
// internal/nostr/verify_test.go
package nostr_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	gonostr "github.com/nbd-wtf/go-nostr"
	nostrauth "github.com/nostr-pay/nostr-pay/internal/nostr"
)

func createSignedAuthEvent(t *testing.T, url, method string) (string, string) {
	t.Helper()

	sk := gonostr.GeneratePrivateKey()
	pk, _ := gonostr.GetPublicKey(sk)

	event := gonostr.Event{
		Kind:      27235, // NIP-98
		CreatedAt: gonostr.Timestamp(time.Now().Unix()),
		Tags: gonostr.Tags{
			{"u", url},
			{"method", method},
		},
		Content: "",
	}
	event.Sign(sk)

	eventJSON, _ := json.Marshal(event)
	token := base64.StdEncoding.EncodeToString(eventJSON)
	return "Nostr " + token, pk
}

func TestAuthMiddleware_ValidEvent(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pubkey := nostrauth.PubkeyFromContext(r.Context())
		if pubkey == "" {
			t.Error("pubkey should be set in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	mw := nostrauth.AuthMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/test", nil)
	token, _ := createSignedAuthEvent(t, "http://example.com/api/test", "GET")
	req.Header.Set("Authorization", token)

	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	mw := nostrauth.AuthMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/test", nil)
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}

func TestAuthMiddleware_InvalidSignature(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})

	mw := nostrauth.AuthMiddleware(handler)

	// Tamper with a valid event
	event := gonostr.Event{
		Kind:      27235,
		CreatedAt: gonostr.Timestamp(time.Now().Unix()),
		Tags:      gonostr.Tags{{"u", "http://example.com/api/test"}, {"method", "GET"}},
		Content:   "",
	}
	sk := gonostr.GeneratePrivateKey()
	event.Sign(sk)
	event.Sig = "0000000000000000000000000000000000000000000000000000000000000000" + event.Sig[64:] // corrupt sig

	eventJSON, _ := json.Marshal(event)
	token := "Nostr " + base64.StdEncoding.EncodeToString(eventJSON)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/test", nil)
	req.Header.Set("Authorization", token)
	rr := httptest.NewRecorder()
	mw.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rr.Code)
	}
}
```

**Step 3: Run tests to verify they fail**

Run: `go test ./internal/nostr/ -v`
Expected: FAIL — package doesn't exist

**Step 4: Write implementation**

```go
// internal/nostr/verify.go
package nostr

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	gonostr "github.com/nbd-wtf/go-nostr"
)

type contextKey string

const pubkeyKey contextKey = "pubkey"

func PubkeyFromContext(ctx context.Context) string {
	v, _ := ctx.Value(pubkeyKey).(string)
	return v
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Nostr ") {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Nostr ")
		eventJSON, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			http.Error(w, "invalid authorization token", http.StatusUnauthorized)
			return
		}

		var event gonostr.Event
		if err := json.Unmarshal(eventJSON, &event); err != nil {
			http.Error(w, "invalid event format", http.StatusUnauthorized)
			return
		}

		// Verify event kind is 27235 (NIP-98)
		if event.Kind != 27235 {
			http.Error(w, "invalid event kind", http.StatusUnauthorized)
			return
		}

		// Verify signature
		ok, err := event.CheckSignature()
		if err != nil || !ok {
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		// Verify event is recent (within 60 seconds)
		eventTime := time.Unix(int64(event.CreatedAt), 0)
		if time.Since(eventTime).Abs() > 60*time.Second {
			http.Error(w, "event too old", http.StatusUnauthorized)
			return
		}

		// Verify URL tag matches request
		urlTag := event.Tags.GetFirst([]string{"u"})
		if urlTag == nil || len(*urlTag) < 2 {
			http.Error(w, "missing url tag", http.StatusUnauthorized)
			return
		}

		// Verify method tag matches request
		methodTag := event.Tags.GetFirst([]string{"method"})
		if methodTag == nil || len(*methodTag) < 2 {
			http.Error(w, "missing method tag", http.StatusUnauthorized)
			return
		}
		if strings.ToUpper((*methodTag)[1]) != r.Method {
			http.Error(w, "method mismatch", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), pubkeyKey, event.PubKey)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
```

**Step 5: Run tests to verify they pass**

Run: `go test ./internal/nostr/ -v`
Expected: PASS (3 tests)

**Step 6: Commit**

```bash
git add internal/nostr/ go.mod go.sum
git commit -m "feat: add NIP-98 authentication middleware"
```

---

### Task 6: Payment Service & Handlers

**Files:**
- Create: `internal/payment/service.go`
- Create: `internal/payment/service_test.go`
- Create: `internal/api/handlers_payment.go`
- Create: `internal/api/handlers_health.go`

**Step 1: Write the failing tests for payment service**

```go
// internal/payment/service_test.go
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
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/payment/ -v`
Expected: FAIL — package doesn't exist

**Step 3: Write payment service**

```go
// internal/payment/service.go
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

	now := time.Now()
	if err := s.store.UpdatePaymentStatus(ctx, "", "paid", &now); err != nil {
		// Try by hash instead
		p, err2 := s.store.GetPaymentByHash(ctx, paymentHash)
		if err2 != nil {
			return fmt.Errorf("get payment by hash: %w", err2)
		}
		if err := s.store.UpdatePaymentStatus(ctx, p.ID, "paid", &now); err != nil {
			return fmt.Errorf("update payment status: %w", err)
		}
	}

	return nil
}

func (s *Service) GetPayment(ctx context.Context, id string) (*store.Payment, error) {
	return s.store.GetPayment(ctx, id)
}

func (s *Service) ListPayments(ctx context.Context, pubkey string, limit, offset int) ([]*store.Payment, error) {
	return s.store.ListPaymentsByUser(ctx, pubkey, limit, offset)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/payment/ -v`
Expected: PASS (2 tests)

**Step 5: Write API handlers**

```go
// internal/api/handlers_health.go
package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

```go
// internal/api/handlers_payment.go
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

	w.WriteHeader(http.StatusOK)
}
```

**Step 6: Commit**

```bash
git add internal/payment/ internal/api/
git commit -m "feat: add payment service and API handlers"
```

---

### Task 7: API Router & Server Setup

**Files:**
- Create: `internal/api/server.go`
- Create: `internal/api/router.go`
- Create: `internal/api/middleware.go`
- Modify: `cmd/server/main.go`

**Step 1: Write the server struct and router**

```go
// internal/api/server.go
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
```

```go
// internal/api/router.go
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
```

```go
// internal/api/middleware.go
package api

import (
	"log/slog"
	"net/http"
	"time"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start).String(),
		)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

**Step 2: Update main.go**

```go
// cmd/server/main.go
package main

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/nostr-pay/nostr-pay/internal/api"
	"github.com/nostr-pay/nostr-pay/internal/config"
	"github.com/nostr-pay/nostr-pay/internal/lnbits"
	"github.com/nostr-pay/nostr-pay/internal/payment"
	"github.com/nostr-pay/nostr-pay/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0o755); err != nil {
		slog.Error("failed to create data dir", "error", err)
		os.Exit(1)
	}

	db, err := store.NewSQLite(cfg.DBPath)
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	lnbitsClient := lnbits.NewClient(cfg.LNbitsURL, cfg.LNbitsAdminKey, cfg.LNbitsInvoiceKey)
	paymentSvc := payment.NewService(db, lnbitsClient, "http://localhost"+cfg.ServerAddr)

	srv := api.NewServer(db, paymentSvc)

	slog.Info("starting server", "addr", cfg.ServerAddr)
	if err := http.ListenAndServe(cfg.ServerAddr, srv.Routes()); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
```

**Step 3: Verify build**

Run: `go build ./cmd/server/`
Expected: Compiles without errors

**Step 4: Run all tests**

Run: `go test ./... -v`
Expected: All tests PASS

**Step 5: Commit**

```bash
git add cmd/server/ internal/api/ go.mod go.sum
git commit -m "feat: add API router, middleware, and server entrypoint"
```

---

## Phase 2: PWA Frontend

### Task 8: Initialize React PWA

**Files:**
- Create: `web/` directory with Vite + React + TypeScript + TailwindCSS

**Step 1: Create Vite project**

Run:
```bash
cd /Users/olivier/Versioncontrol/local/nostr-pay
npm create vite@latest web -- --template react-ts
cd web
npm install
```

**Step 2: Install dependencies**

Run:
```bash
cd /Users/olivier/Versioncontrol/local/nostr-pay/web
npm install nostr-tools qrcode.react html5-qrcode zustand react-router-dom
npm install -D tailwindcss @tailwindcss/vite
```

**Step 3: Configure TailwindCSS**

Replace `web/src/index.css` with:
```css
@import "tailwindcss";
```

Update `web/vite.config.ts`:
```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
  ],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})
```

**Step 4: Set up basic App with router**

Replace `web/src/App.tsx`:
```tsx
import { BrowserRouter, Routes, Route } from 'react-router-dom'

function Home() {
  return (
    <div className="min-h-screen bg-gray-950 text-white flex items-center justify-center">
      <div className="text-center">
        <h1 className="text-4xl font-bold mb-4">nostr-pay</h1>
        <p className="text-gray-400">Instant Lightning Payments via Nostr</p>
      </div>
    </div>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
      </Routes>
    </BrowserRouter>
  )
}
```

**Step 5: Verify dev server starts**

Run:
```bash
cd /Users/olivier/Versioncontrol/local/nostr-pay/web
npm run dev
```
Expected: Vite dev server starts at http://localhost:5173

**Step 6: Commit**

```bash
cd /Users/olivier/Versioncontrol/local/nostr-pay
git add web/ .gitignore
git commit -m "feat: initialize React PWA with Vite, TailwindCSS, and router"
```

---

### Task 9: API Client & Auth Store

**Files:**
- Create: `web/src/lib/api.ts`
- Create: `web/src/stores/auth.ts`

**Step 1: Create API client**

```typescript
// web/src/lib/api.ts
const API_BASE = '/api'

async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
  authToken?: string
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...((options.headers as Record<string, string>) || {}),
  }

  if (authToken) {
    headers['Authorization'] = authToken
  }

  const response = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  })

  if (!response.ok) {
    const text = await response.text()
    throw new Error(`API error ${response.status}: ${text}`)
  }

  return response.json()
}

export interface CreateInvoiceResponse {
  payment_id: string
  bolt11: string
  payment_hash: string
}

export interface Payment {
  ID: string
  Bolt11: string
  AmountSats: number
  Memo: string
  SenderPubkey: string
  ReceiverPubkey: string
  PaymentHash: string
  Status: string
  CreatedAt: string
  SettledAt: string | null
}

export const api = {
  health: () => apiFetch<{ status: string }>('/health'),

  createInvoice: (amountSats: number, memo: string, token: string) =>
    apiFetch<CreateInvoiceResponse>(
      '/payments/invoice',
      {
        method: 'POST',
        body: JSON.stringify({ amount_sats: amountSats, memo }),
      },
      token
    ),

  getPayment: (id: string, token: string) =>
    apiFetch<Payment>(`/payments/${id}`, {}, token),

  getPaymentHistory: (token: string) =>
    apiFetch<Payment[]>('/payments/history', {}, token),
}
```

**Step 2: Create auth store with NIP-46 + NIP-98**

```typescript
// web/src/stores/auth.ts
import { create } from 'zustand'
import { finalizeEvent, type Event } from 'nostr-tools/pure'
import { hexToBytes } from '@noble/hashes/utils'

interface AuthState {
  pubkey: string | null
  secretKey: Uint8Array | null
  isLoggedIn: boolean
  login: (nsec: string) => void
  logout: () => void
  createAuthToken: (url: string, method: string) => string
}

export const useAuth = create<AuthState>((set, get) => ({
  pubkey: null,
  secretKey: null,
  isLoggedIn: false,

  login: (secretKeyHex: string) => {
    const secretKey = hexToBytes(secretKeyHex)
    // Derive pubkey from secret key
    const { getPublicKey } = require('nostr-tools/pure')
    const pubkey = getPublicKey(secretKey)

    set({ pubkey, secretKey, isLoggedIn: true })
  },

  logout: () => {
    set({ pubkey: null, secretKey: null, isLoggedIn: false })
  },

  createAuthToken: (url: string, method: string): string => {
    const { secretKey } = get()
    if (!secretKey) throw new Error('Not logged in')

    const event: Event = finalizeEvent({
      kind: 27235,
      created_at: Math.floor(Date.now() / 1000),
      tags: [
        ['u', url],
        ['method', method],
      ],
      content: '',
    }, secretKey)

    const eventJson = JSON.stringify(event)
    const token = btoa(eventJson)
    return `Nostr ${token}`
  },
}))
```

Note: This is a simplified auth for MVP. In production, NIP-46 remote signing replaces direct key handling. We'll add NIP-46 in Task 11.

**Step 3: Commit**

```bash
git add web/src/lib/ web/src/stores/
git commit -m "feat: add API client and auth store with NIP-98 token generation"
```

---

### Task 10: QR Code Components

**Files:**
- Create: `web/src/components/QRGenerator.tsx`
- Create: `web/src/components/QRScanner.tsx`

**Step 1: Create QR code generator**

```tsx
// web/src/components/QRGenerator.tsx
import { QRCodeSVG } from 'qrcode.react'

interface QRGeneratorProps {
  value: string
  size?: number
  label?: string
}

export function QRGenerator({ value, size = 256, label }: QRGeneratorProps) {
  return (
    <div className="flex flex-col items-center gap-4">
      <div className="bg-white p-4 rounded-2xl">
        <QRCodeSVG
          value={value}
          size={size}
          level="M"
          bgColor="#ffffff"
          fgColor="#000000"
        />
      </div>
      {label && (
        <p className="text-sm text-gray-400 text-center">{label}</p>
      )}
    </div>
  )
}
```

**Step 2: Create QR code scanner**

```tsx
// web/src/components/QRScanner.tsx
import { useEffect, useRef, useState } from 'react'
import { Html5Qrcode } from 'html5-qrcode'

interface QRScannerProps {
  onScan: (result: string) => void
  onError?: (error: string) => void
}

export function QRScanner({ onScan, onError }: QRScannerProps) {
  const scannerRef = useRef<Html5Qrcode | null>(null)
  const [isScanning, setIsScanning] = useState(false)
  const containerRef = useRef<string>(`qr-reader-${Date.now()}`)

  useEffect(() => {
    const scanner = new Html5Qrcode(containerRef.current)
    scannerRef.current = scanner

    scanner.start(
      { facingMode: 'environment' },
      { fps: 10, qrbox: { width: 250, height: 250 } },
      (decodedText) => {
        scanner.stop().then(() => {
          setIsScanning(false)
          onScan(decodedText)
        })
      },
      () => {} // Ignore scan errors (no QR in frame)
    ).then(() => {
      setIsScanning(true)
    }).catch((err) => {
      onError?.(err.toString())
    })

    return () => {
      if (scannerRef.current?.isScanning) {
        scannerRef.current.stop().catch(() => {})
      }
    }
  }, [onScan, onError])

  return (
    <div className="flex flex-col items-center gap-4">
      <div
        id={containerRef.current}
        className="w-full max-w-sm rounded-2xl overflow-hidden"
      />
      {isScanning && (
        <p className="text-sm text-gray-400">Scanning for QR code...</p>
      )}
    </div>
  )
}
```

**Step 3: Commit**

```bash
git add web/src/components/
git commit -m "feat: add QR code generator and scanner components"
```

---

### Task 11: Payment Pages (Receive & Pay)

**Files:**
- Create: `web/src/pages/ReceivePage.tsx`
- Create: `web/src/pages/PayPage.tsx`
- Create: `web/src/pages/HistoryPage.tsx`
- Create: `web/src/components/Layout.tsx`
- Modify: `web/src/App.tsx`

**Step 1: Create layout component**

```tsx
// web/src/components/Layout.tsx
import { Link, Outlet, useLocation } from 'react-router-dom'

const navItems = [
  { path: '/receive', label: 'Receive', icon: '↓' },
  { path: '/pay', label: 'Pay', icon: '↑' },
  { path: '/history', label: 'History', icon: '☰' },
  { path: '/merchant/pos', label: 'POS', icon: '◻' },
]

export function Layout() {
  const location = useLocation()

  return (
    <div className="min-h-screen bg-gray-950 text-white flex flex-col">
      <header className="px-4 py-3 border-b border-gray-800 flex items-center justify-between">
        <Link to="/" className="text-xl font-bold">nostr-pay</Link>
      </header>

      <main className="flex-1 p-4">
        <Outlet />
      </main>

      <nav className="border-t border-gray-800 px-4 py-2">
        <div className="flex justify-around">
          {navItems.map((item) => (
            <Link
              key={item.path}
              to={item.path}
              className={`flex flex-col items-center py-2 px-3 rounded-lg text-sm ${
                location.pathname === item.path
                  ? 'text-amber-400'
                  : 'text-gray-500 hover:text-gray-300'
              }`}
            >
              <span className="text-lg">{item.icon}</span>
              <span>{item.label}</span>
            </Link>
          ))}
        </div>
      </nav>
    </div>
  )
}
```

**Step 2: Create receive page**

```tsx
// web/src/pages/ReceivePage.tsx
import { useState } from 'react'
import { QRGenerator } from '../components/QRGenerator'
import { api } from '../lib/api'
import { useAuth } from '../stores/auth'

export function ReceivePage() {
  const [amount, setAmount] = useState('')
  const [memo, setMemo] = useState('')
  const [invoice, setInvoice] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const { createAuthToken } = useAuth()

  const handleGenerate = async () => {
    const sats = parseInt(amount)
    if (isNaN(sats) || sats <= 0) {
      setError('Enter a valid amount')
      return
    }

    setLoading(true)
    setError(null)

    try {
      const url = `${window.location.origin}/api/payments/invoice`
      const token = createAuthToken(url, 'POST')
      const result = await api.createInvoice(sats, memo, token)
      setInvoice(result.bolt11)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create invoice')
    } finally {
      setLoading(false)
    }
  }

  if (invoice) {
    return (
      <div className="flex flex-col items-center gap-6 pt-8">
        <h2 className="text-2xl font-bold">{amount} sats</h2>
        <QRGenerator
          value={`lightning:${invoice}`}
          size={280}
          label="Scan to pay with Lightning"
        />
        <button
          onClick={() => { setInvoice(null); setAmount(''); setMemo('') }}
          className="text-gray-400 hover:text-white"
        >
          New Invoice
        </button>
      </div>
    )
  }

  return (
    <div className="max-w-sm mx-auto pt-8">
      <h2 className="text-2xl font-bold mb-6">Receive Payment</h2>

      <div className="space-y-4">
        <div>
          <label className="block text-sm text-gray-400 mb-1">Amount (sats)</label>
          <input
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            placeholder="1000"
            className="w-full bg-gray-900 border border-gray-700 rounded-lg px-4 py-3 text-white text-lg"
          />
        </div>

        <div>
          <label className="block text-sm text-gray-400 mb-1">Memo (optional)</label>
          <input
            type="text"
            value={memo}
            onChange={(e) => setMemo(e.target.value)}
            placeholder="What's this for?"
            className="w-full bg-gray-900 border border-gray-700 rounded-lg px-4 py-3 text-white"
          />
        </div>

        {error && <p className="text-red-400 text-sm">{error}</p>}

        <button
          onClick={handleGenerate}
          disabled={loading}
          className="w-full bg-amber-500 hover:bg-amber-600 text-black font-bold py-3 rounded-lg disabled:opacity-50"
        >
          {loading ? 'Generating...' : 'Generate QR Code'}
        </button>
      </div>
    </div>
  )
}
```

**Step 3: Create pay page**

```tsx
// web/src/pages/PayPage.tsx
import { useState } from 'react'
import { QRScanner } from '../components/QRScanner'

export function PayPage() {
  const [scannedInvoice, setScannedInvoice] = useState<string | null>(null)
  const [manualInvoice, setManualInvoice] = useState('')
  const [showScanner, setShowScanner] = useState(true)

  const handleScan = (result: string) => {
    // Strip lightning: prefix if present
    const invoice = result.replace(/^lightning:/i, '')
    setScannedInvoice(invoice)
    setShowScanner(false)
  }

  const handlePay = () => {
    const invoice = scannedInvoice || manualInvoice
    if (!invoice) return

    // For MVP: open in user's Lightning wallet
    window.location.href = `lightning:${invoice}`
  }

  return (
    <div className="max-w-sm mx-auto pt-8">
      <h2 className="text-2xl font-bold mb-6">Pay</h2>

      {showScanner && !scannedInvoice ? (
        <div className="space-y-4">
          <QRScanner onScan={handleScan} />
          <div className="text-center">
            <button
              onClick={() => setShowScanner(false)}
              className="text-gray-400 hover:text-white text-sm"
            >
              Enter invoice manually
            </button>
          </div>
        </div>
      ) : (
        <div className="space-y-4">
          {scannedInvoice ? (
            <div className="bg-gray-900 border border-gray-700 rounded-lg p-4">
              <p className="text-sm text-gray-400 mb-1">Scanned Invoice</p>
              <p className="text-white font-mono text-xs break-all">
                {scannedInvoice.substring(0, 60)}...
              </p>
            </div>
          ) : (
            <div>
              <label className="block text-sm text-gray-400 mb-1">Lightning Invoice</label>
              <textarea
                value={manualInvoice}
                onChange={(e) => setManualInvoice(e.target.value)}
                placeholder="lnbc..."
                rows={4}
                className="w-full bg-gray-900 border border-gray-700 rounded-lg px-4 py-3 text-white font-mono text-sm"
              />
            </div>
          )}

          <button
            onClick={handlePay}
            disabled={!scannedInvoice && !manualInvoice}
            className="w-full bg-amber-500 hover:bg-amber-600 text-black font-bold py-3 rounded-lg disabled:opacity-50"
          >
            Pay with Lightning Wallet
          </button>

          <button
            onClick={() => { setScannedInvoice(null); setShowScanner(true); setManualInvoice('') }}
            className="w-full text-gray-400 hover:text-white py-2"
          >
            Scan Again
          </button>
        </div>
      )}
    </div>
  )
}
```

**Step 4: Create history page**

```tsx
// web/src/pages/HistoryPage.tsx
import { useEffect, useState } from 'react'
import { api, type Payment } from '../lib/api'
import { useAuth } from '../stores/auth'

export function HistoryPage() {
  const [payments, setPayments] = useState<Payment[]>([])
  const [loading, setLoading] = useState(true)
  const { createAuthToken, isLoggedIn } = useAuth()

  useEffect(() => {
    if (!isLoggedIn) return

    const url = `${window.location.origin}/api/payments/history`
    const token = createAuthToken(url, 'GET')

    api.getPaymentHistory(token)
      .then(setPayments)
      .catch(console.error)
      .finally(() => setLoading(false))
  }, [isLoggedIn, createAuthToken])

  if (!isLoggedIn) {
    return (
      <div className="text-center pt-8">
        <p className="text-gray-400">Login to view payment history</p>
      </div>
    )
  }

  if (loading) {
    return <div className="text-center pt-8 text-gray-400">Loading...</div>
  }

  return (
    <div className="max-w-sm mx-auto pt-4">
      <h2 className="text-2xl font-bold mb-6">Payment History</h2>

      {payments.length === 0 ? (
        <p className="text-gray-400 text-center">No payments yet</p>
      ) : (
        <div className="space-y-3">
          {payments.map((p) => (
            <div
              key={p.ID}
              className="bg-gray-900 border border-gray-800 rounded-lg p-4"
            >
              <div className="flex justify-between items-center">
                <div>
                  <p className="font-bold">
                    {p.Status === 'paid' ? '+' : ''}{p.AmountSats} sats
                  </p>
                  {p.Memo && <p className="text-sm text-gray-400">{p.Memo}</p>}
                </div>
                <span
                  className={`text-xs px-2 py-1 rounded ${
                    p.Status === 'paid'
                      ? 'bg-green-900 text-green-400'
                      : p.Status === 'expired'
                      ? 'bg-red-900 text-red-400'
                      : 'bg-yellow-900 text-yellow-400'
                  }`}
                >
                  {p.Status}
                </span>
              </div>
              <p className="text-xs text-gray-600 mt-2">
                {new Date(p.CreatedAt).toLocaleString()}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
```

**Step 5: Update App.tsx with routes**

```tsx
// web/src/App.tsx
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { Layout } from './components/Layout'
import { ReceivePage } from './pages/ReceivePage'
import { PayPage } from './pages/PayPage'
import { HistoryPage } from './pages/HistoryPage'

function Home() {
  return (
    <div className="flex flex-col items-center justify-center pt-20 gap-6">
      <h1 className="text-4xl font-bold">nostr-pay</h1>
      <p className="text-gray-400 text-center max-w-xs">
        Instant Lightning payments powered by Nostr
      </p>
    </div>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Home />} />
          <Route path="/receive" element={<ReceivePage />} />
          <Route path="/pay" element={<PayPage />} />
          <Route path="/history" element={<HistoryPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  )
}
```

**Step 6: Verify dev server starts**

Run:
```bash
cd /Users/olivier/Versioncontrol/local/nostr-pay/web
npm run dev
```
Expected: No compile errors, pages render

**Step 7: Commit**

```bash
cd /Users/olivier/Versioncontrol/local/nostr-pay
git add web/src/
git commit -m "feat: add receive, pay, and history pages with layout"
```

---

### Task 12: Merchant POS Page

**Files:**
- Create: `web/src/pages/MerchantPOS.tsx`
- Modify: `web/src/App.tsx`

**Step 1: Create POS page with numpad**

```tsx
// web/src/pages/MerchantPOS.tsx
import { useState, useCallback } from 'react'
import { QRGenerator } from '../components/QRGenerator'
import { api } from '../lib/api'
import { useAuth } from '../stores/auth'

type POSState = 'input' | 'waiting' | 'paid'

export function MerchantPOS() {
  const [amount, setAmount] = useState('0')
  const [state, setState] = useState<POSState>('input')
  const [invoice, setInvoice] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const { createAuthToken } = useAuth()

  const handleNumpad = useCallback((key: string) => {
    setAmount((prev) => {
      if (key === 'C') return '0'
      if (key === '←') return prev.length > 1 ? prev.slice(0, -1) : '0'
      if (prev === '0') return key
      return prev + key
    })
  }, [])

  const handleCharge = async () => {
    const sats = parseInt(amount)
    if (sats <= 0) return

    setError(null)
    setState('waiting')

    try {
      const url = `${window.location.origin}/api/payments/invoice`
      const token = createAuthToken(url, 'POST')
      const result = await api.createInvoice(sats, `POS Payment`, token)
      setInvoice(result.bolt11)

      // Poll for payment status
      const pollInterval = setInterval(async () => {
        try {
          const paymentUrl = `${window.location.origin}/api/payments/${result.payment_id}`
          const pollToken = createAuthToken(paymentUrl, 'GET')
          const payment = await api.getPayment(result.payment_id, pollToken)
          if (payment.Status === 'paid') {
            clearInterval(pollInterval)
            setState('paid')
          }
        } catch {
          // Ignore poll errors
        }
      }, 2000)

      // Timeout after 10 minutes
      setTimeout(() => {
        clearInterval(pollInterval)
        if (state === 'waiting') {
          setError('Invoice expired')
          setState('input')
          setInvoice(null)
        }
      }, 600_000)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed')
      setState('input')
    }
  }

  const handleReset = () => {
    setAmount('0')
    setState('input')
    setInvoice(null)
    setError(null)
  }

  // Paid state
  if (state === 'paid') {
    return (
      <div className="fixed inset-0 bg-green-950 flex flex-col items-center justify-center gap-6 z-50">
        <div className="text-6xl">&#10003;</div>
        <h2 className="text-4xl font-bold text-green-400">Paid!</h2>
        <p className="text-2xl text-green-300">{parseInt(amount).toLocaleString()} sats</p>
        <button
          onClick={handleReset}
          className="mt-8 bg-green-800 hover:bg-green-700 text-white font-bold py-4 px-12 rounded-xl text-xl"
        >
          Next Customer
        </button>
      </div>
    )
  }

  // QR display state
  if (state === 'waiting' && invoice) {
    return (
      <div className="fixed inset-0 bg-gray-950 flex flex-col items-center justify-center gap-6 z-50">
        <h2 className="text-3xl font-bold">{parseInt(amount).toLocaleString()} sats</h2>
        <QRGenerator value={`lightning:${invoice}`} size={300} />
        <p className="text-gray-400 animate-pulse">Waiting for payment...</p>
        <button
          onClick={handleReset}
          className="text-gray-600 hover:text-gray-400 text-sm mt-4"
        >
          Cancel
        </button>
      </div>
    )
  }

  // Numpad input state
  const numpadKeys = ['1', '2', '3', '4', '5', '6', '7', '8', '9', 'C', '0', '←']

  return (
    <div className="fixed inset-0 bg-gray-950 flex flex-col z-50">
      <div className="flex-1 flex flex-col items-center justify-center">
        <p className="text-gray-500 text-sm mb-2">Amount (sats)</p>
        <p className="text-5xl font-bold tabular-nums">
          {parseInt(amount).toLocaleString()}
        </p>
      </div>

      {error && <p className="text-red-400 text-center text-sm mb-2">{error}</p>}

      <div className="grid grid-cols-3 gap-2 p-4 max-w-sm mx-auto w-full">
        {numpadKeys.map((key) => (
          <button
            key={key}
            onClick={() => handleNumpad(key)}
            className="bg-gray-900 hover:bg-gray-800 text-white text-2xl font-bold py-5 rounded-xl active:bg-gray-700"
          >
            {key}
          </button>
        ))}
      </div>

      <div className="p-4 max-w-sm mx-auto w-full">
        <button
          onClick={handleCharge}
          disabled={amount === '0'}
          className="w-full bg-amber-500 hover:bg-amber-600 text-black font-bold py-4 rounded-xl text-xl disabled:opacity-30"
        >
          Charge
        </button>
      </div>
    </div>
  )
}
```

**Step 2: Add POS route to App.tsx**

Add the import and route to `web/src/App.tsx`:

```tsx
import { MerchantPOS } from './pages/MerchantPOS'

// Inside <Routes>:
<Route path="/merchant/pos" element={<MerchantPOS />} />
```

**Step 3: Verify POS renders**

Run: `cd web && npm run dev`
Navigate to `http://localhost:5173/merchant/pos`
Expected: Fullscreen numpad with charge button

**Step 4: Commit**

```bash
cd /Users/olivier/Versioncontrol/local/nostr-pay
git add web/src/
git commit -m "feat: add merchant POS page with numpad and QR display"
```

---

## Phase 3: Docker & Deployment

### Task 13: Dockerize the Application

**Files:**
- Create: `Dockerfile`
- Create: `web/Dockerfile`
- Create: `docker-compose.yml`
- Create: `web/nginx.conf`

**Step 1: Create Go backend Dockerfile**

```dockerfile
# Dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o nostr-pay ./cmd/server/

FROM alpine:3.20
RUN apk add --no-cache ca-certificates curl
RUN adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=builder /app/nostr-pay .
RUN mkdir -p /app/data && chown -R appuser:appuser /app
USER appuser
HEALTHCHECK --interval=30s --timeout=10s CMD curl -f http://localhost:8080/api/health || exit 1
EXPOSE 8080
CMD ["./nostr-pay"]
```

**Step 2: Create frontend nginx config**

```nginx
# web/nginx.conf
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://api:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

**Step 3: Create frontend Dockerfile**

```dockerfile
# web/Dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```

**Step 4: Create docker-compose.yml**

```yaml
services:
  api:
    build: .
    container_name: nostr-pay-api
    ports:
      - "8080:8080"
    environment:
      - SERVER_ADDR=:8080
    env_file:
      - .env
    volumes:
      - ./data:/app/data
    restart: unless-stopped
    networks:
      - nostr-pay
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/api/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  web:
    build: ./web
    container_name: nostr-pay-web
    ports:
      - "3000:80"
    depends_on:
      api:
        condition: service_healthy
    restart: unless-stopped
    networks:
      - nostr-pay

networks:
  nostr-pay:
    driver: bridge
```

**Step 5: Verify Docker build**

Run:
```bash
docker compose build
```
Expected: Both images build successfully

**Step 6: Commit**

```bash
git add Dockerfile web/Dockerfile web/nginx.conf docker-compose.yml .env.example
git commit -m "feat: add Docker setup with Go API and nginx frontend"
```

---

## Phase 4: Integration & Polish

### Task 14: WebSocket for Real-Time Payment Updates

**Files:**
- Create: `internal/api/websocket.go`
- Modify: `internal/api/router.go`
- Modify: `internal/api/handlers_payment.go`

This task adds real-time payment status updates via WebSocket, so the POS screen updates instantly when a payment is received instead of polling.

**Step 1: Install gorilla/websocket**

Run:
```bash
go get github.com/gorilla/websocket
```

**Step 2: Create WebSocket hub**

```go
// internal/api/websocket.go
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
```

**Step 3: Add wsHub to Server and router**

Update `internal/api/server.go`:
```go
type Server struct {
	store      store.Store
	paymentSvc *payment.Service
	wsHub      *WSHub
}

func NewServer(store store.Store, paymentSvc *payment.Service) *Server {
	return &Server{
		store:      store,
		paymentSvc: paymentSvc,
		wsHub:      NewWSHub(),
	}
}
```

Add to `internal/api/router.go`:
```go
mux.HandleFunc("GET /api/ws", s.handleWS)
```

**Step 4: Notify WebSocket clients on webhook**

Update `handleWebhook` in `internal/api/handlers_payment.go` to add after successful webhook processing:
```go
s.wsHub.NotifyPayment(payload.PaymentHash, "paid")
```

**Step 5: Verify build**

Run: `go build ./cmd/server/`
Expected: Compiles

**Step 6: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass

**Step 7: Commit**

```bash
git add internal/api/ go.mod go.sum
git commit -m "feat: add WebSocket hub for real-time payment notifications"
```

---

### Task 15: PWA Manifest & Service Worker

**Files:**
- Create: `web/public/manifest.json`
- Modify: `web/index.html`

**Step 1: Create PWA manifest**

```json
{
  "name": "nostr-pay",
  "short_name": "nostr-pay",
  "description": "Instant Lightning Payments via Nostr",
  "start_url": "/",
  "display": "standalone",
  "background_color": "#030712",
  "theme_color": "#f59e0b",
  "icons": [
    {
      "src": "/icon-192.png",
      "sizes": "192x192",
      "type": "image/png"
    },
    {
      "src": "/icon-512.png",
      "sizes": "512x512",
      "type": "image/png"
    }
  ]
}
```

**Step 2: Update index.html**

Add to `<head>` in `web/index.html`:
```html
<link rel="manifest" href="/manifest.json" />
<meta name="theme-color" content="#f59e0b" />
<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no" />
```

**Step 3: Commit**

```bash
git add web/public/ web/index.html
git commit -m "feat: add PWA manifest for installable app"
```

---

### Task 16: Final Integration Test & Cleanup

**Step 1: Run all Go tests**

Run: `go test ./... -v -count=1`
Expected: All tests pass

**Step 2: Build Docker images**

Run: `docker compose build`
Expected: Both images build

**Step 3: Verify frontend builds**

Run:
```bash
cd web && npm run build
```
Expected: Build completes without errors

**Step 4: Run the full stack locally**

Create a `.env` file with test values:
```bash
cp .env.example .env
# Edit .env with your LNbits credentials
```

Run:
```bash
docker compose up --build
```
Expected: Both services start, health check passes

**Step 5: Final commit**

```bash
git add -A
git commit -m "chore: final integration checks and cleanup"
```

---

## Summary

| Phase | Tasks | Description |
|-------|-------|-------------|
| 1 | Tasks 1-7 | Go backend: config, SQLite, LNbits client, NIP-98 auth, payment service, API router |
| 2 | Tasks 8-12 | PWA frontend: Vite setup, API client, QR components, payment pages, merchant POS |
| 3 | Task 13 | Docker: Dockerfile, docker-compose, nginx |
| 4 | Tasks 14-16 | Polish: WebSocket, PWA manifest, integration tests |

**Total: 16 tasks across 4 phases**
