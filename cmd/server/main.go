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
