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
	t.Setenv("LNBITS_URL", "")
	t.Setenv("LNBITS_ADMIN_KEY", "")
	t.Setenv("LNBITS_INVOICE_KEY", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing required config")
	}
}
