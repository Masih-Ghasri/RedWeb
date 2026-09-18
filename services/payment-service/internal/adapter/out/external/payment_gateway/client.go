package payment_gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type MockBankClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewMockBankClient(baseURL string) *MockBankClient {
	return &MockBankClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *MockBankClient) Charge(ctx context.Context, orderID string, amount float64) (bool, error) {
	payload := map[string]interface{}{
		"order_id": orderID,
		"amount":   amount,
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/pay", bytes.NewBuffer(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to call bank API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	return false, nil
}
