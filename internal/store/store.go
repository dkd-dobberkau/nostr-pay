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
