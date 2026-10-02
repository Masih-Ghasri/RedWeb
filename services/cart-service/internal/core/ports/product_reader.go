package ports

import "context"

// ProductInfoReader is the read side of the shared product cache.
//
// The cart uses it for two things: pricing a line when it is added, and
// re-pricing every line on read so displayed totals track the catalogue.
//
// It reads the same Redis entry that order-service writes from catalog
// events ("product:{id}" → {"stock":n,"price":x}). On a cache miss it
// returns domain.ErrProductNotFound, which the HTTP layer turns into 404;
// any other error is an infrastructure fault and becomes 500.
type ProductInfoReader interface {
	GetProductInfo(ctx context.Context, productID string) (stock int, price float64, err error)
}
