package mongodb

import (
	"context"
	"errors"

	"github.com/Masih-Ghasri/RedWeb/services/catalog-service/internal/core/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoProductRepository struct {
	collection *mongo.Collection
}

func NewMongoProductRepository(db *mongo.Database) *MongoProductRepository {
	return &MongoProductRepository{
		collection: db.Collection("products"),
	}
}

func (r *MongoProductRepository) Save(ctx context.Context, product *domain.Product) error {
	doc := toDocument(product)
	_, err := r.collection.InsertOne(ctx, doc)
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
