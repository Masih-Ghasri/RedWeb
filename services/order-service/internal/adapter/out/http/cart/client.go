package cart

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Masih-Ghasri/RedWeb/services/order-service/internal/core/ports"
)

// maxBodyBytes caps how much of cart-service's response we are willing to
// read. A cart cannot legitimately be this large, and without a cap a
// misbehaving upstream could hold a checkout open indefinitely.
const maxBodyBytes = 1 << 20 // 1 MiB

// CartHTTPClient implements ports.CartClient over HTTP.
type CartHTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewCartHTTPClient instantiates a new HTTP client for the cart service.
func NewCartHTTPClient(baseURL string) *CartHTTPClient {
	return &CartHTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

// GetCart fetches a checkout snapshot.
//
// Error mapping is the interesting part:
//   - 404                → ports.ErrCartNotFound  (caller's cart id is wrong)
//   - any other non-200  → ports.ErrCartUnavailable (→503)
//
// A decoding failure is also reported as unavailable: refusing to build an
// order from a cart we cannot read is better than guessing at its contents.
func (c *CartHTTPClient) GetCart(ctx context.Context, cartID string) (*ports.CartSnapshot, error) {
	url := fmt.Sprintf("%s/api/v1/carts/%s", c.baseURL, cartID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build cart request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ports.ErrCartUnavailable, err)
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return nil, ports.ErrCartNotFound
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%w: cart service answered %s", ports.ErrCartUnavailable, resp.Status)
	}

	var payload struct {
		CustomerID string `json:"customer_id"`
		Items      []struct {
			ProductID string `json:"product_id"`
			Quantity  int    `json:"quantity"`
		} `json:"items"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxBodyBytes)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: decode cart response: %v", ports.ErrCartUnavailable, err)
	}

	if payload.CustomerID == "" {
		return nil, fmt.Errorf("%w: cart %s carries no customer_id", ports.ErrCartUnavailable, cartID)
	}

	snapshot := &ports.CartSnapshot{
		CustomerID: payload.CustomerID,
		Items:      make([]ports.CartItemRef, 0, len(payload.Items)),
	}
	for _, item := range payload.Items {
		snapshot.Items = append(snapshot.Items, ports.CartItemRef{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return snapshot, nil
}

// ClearCart empties a cart after a successful checkout. Both 200 and 404 are
// success: clearing something already gone is the desired end state.
//
// The body is drained (up to maxBodyBytes) so the connection can be reused
// instead of torn down on every order.
func (c *CartHTTPClient) ClearCart(ctx context.Context, cartID string) error {
	url := fmt.Sprintf("%s/api/v1/carts/%s", c.baseURL, cartID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("build cart clear request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ports.ErrCartUnavailable, err)
	}
	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxBodyBytes))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("%w: cart clear answered %s", ports.ErrCartUnavailable, resp.Status)
	}

	return nil
}
