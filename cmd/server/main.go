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
