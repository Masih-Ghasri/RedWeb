package ports

import "context"

// ProductCache is the write side of the shared product cache.
//
// The read side is InventoryCache (inventory_cache_port.go). Both address the
// same Redis key so checkout prices an order from exactly the data the cache
// consumer wrote:
//
//	key     "product:{productID}"
//	value   {"stock":<int>,"price":<float64>}
//
// Keep that JSON shape in sync with RedisProductCache.SetProduct and
// RedisInventoryCache.GetProductStockAndPrice — changing one without the
// other breaks order creation with "product not found in cache".
type ProductCache interface {
	SetProduct(ctx context.Context, productID string, stock int, price float64) error
	DeleteProduct(ctx context.Context, productID string) error
}
