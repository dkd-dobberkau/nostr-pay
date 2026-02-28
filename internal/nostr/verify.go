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
