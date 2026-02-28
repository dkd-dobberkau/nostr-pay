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
