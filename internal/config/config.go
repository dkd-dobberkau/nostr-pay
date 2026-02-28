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
