package mongodb

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// productsCollection is the single place the collection name is spelled out.
const productsCollection = "products"

type MongoProductRepository struct {
	collection *mongo.Collection
}

func NewMongoProductRepository(db *mongo.Database) *MongoProductRepository {
	return &MongoProductRepository{
		collection: db.Collection(productsCollection),
	}
}

// EnsureIndexes creates the indexes the listing queries rely on. It is
// idempotent, so calling it on every startup is cheap and keeps a fresh
// database from falling back to collection scans.
//
// Note that the name search is a case-insensitive substring match expressed
// as an anchored-free regex, which cannot use an index for anything but an
// exact value. That is a deliberate trade for search-box semantics; a
// `$text` index would be faster but matches whole tokens instead.
func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	_, err := db.Collection(productsCollection).Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}},
		{Keys: bson.D{{Key: "price", Value: 1}}},
		{Keys: bson.D{{Key: "stock", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
	})
	return err
}

// Save inserts a new product. InsertOne is intentional: re-saving an existing
// id would either duplicate or silently overwrite, and neither is what
// "create" means.
func (r *MongoProductRepository) Save(ctx context.Context, product *domain.Product) error {
	_, err := r.collection.InsertOne(ctx, toDocument(product))
	return err
}

func (r *MongoProductRepository) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	var doc ProductDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, domain.ErrProductNotFound
		}
		return nil, err
	}
	return toDomain(&doc), nil
}

func (r *MongoProductRepository) FindAll(ctx context.Context, filter domain.ProductFilter) ([]*domain.Product, int, error) {
	query := buildQuery(filter)

	total, err := r.collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	page, pageSize := normalizePagination(filter)
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(int64((page - 1) * pageSize)).
		SetLimit(int64(pageSize))

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var docs []ProductDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, 0, err
	}

	products := make([]*domain.Product, 0, len(docs))
	for i := range docs {
		products = append(products, toDomain(&docs[i]))
	}

	return products, int(total), nil
}

// Update replaces the whole document, keeping created_at (carried over from
// the copy read before the change) and letting domain.Update refresh
// updated_at.
func (r *MongoProductRepository) Update(ctx context.Context, product *domain.Product) error {
	res, err := r.collection.ReplaceOne(ctx, bson.M{"_id": product.ID}, toDocument(product))
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

func (r *MongoProductRepository) Delete(ctx context.Context, id string) error {
	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

// buildQuery turns a ProductFilter into a Mongo filter document.
//
// The user-supplied search string is passed through regexp.QuoteMeta first:
// an unescaped string could otherwise inject regex metacharacters and turn a
// cheap query into a pathological one.
func buildQuery(filter domain.ProductFilter) bson.M {
	query := bson.M{}

	if q := strings.TrimSpace(filter.Q); q != "" {
		query["name"] = bson.M{
			"$regex":   regexp.QuoteMeta(q),
			"$options": "i",
		}
	}

	price := bson.M{}
	if filter.MinPrice != nil {
		price["$gte"] = *filter.MinPrice
	}
	if filter.MaxPrice != nil {
		price["$lte"] = *filter.MaxPrice
	}
	if len(price) > 0 {
		query["price"] = price
	}

	if filter.InStockOnly {
		query["stock"] = bson.M{"$gt": 0}
	}

	return query
}

// normalizePagination guarantees a usable page/pageSize pair even if the
// filter arrives unclamped — a negative page would otherwise produce a
// negative skip.
func normalizePagination(filter domain.ProductFilter) (int, int) {
	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 {
		pageSize = domain.DefaultPageSize
	}
	return page, pageSize
}
