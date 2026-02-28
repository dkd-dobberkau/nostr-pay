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
