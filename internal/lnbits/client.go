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
