package ports

import "context"

type InventoryCache interface {
	GetProductStockAndPrice(ctx context.Context, productID string) (int, float64, error)
}
