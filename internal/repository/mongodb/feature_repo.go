package mongodb

import (
	"context"
	"feature-flags/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FeatureRepository struct {
	collection *mongo.Collection
}

func NewFeatureRepository(db *mongo.Database) *FeatureRepository {
	return &FeatureRepository{
		collection: db.Collection("features"),
	}
}

func (r *FeatureRepository) Create(ctx context.Context, feature *models.Feature) error {
	feature.CreatedAt = time.Now()
	feature.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, feature)
	if err != nil {
		return err
	}

	feature.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *FeatureRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Feature, error) {
	var feature models.Feature
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&feature)
	if err != nil {
		return nil, err
	}
	return &feature, nil
}

func (r *FeatureRepository) Update(ctx context.Context, feature *models.Feature) error {
	feature.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": feature.ID},
		bson.M{"$set": feature},
	)
	return err
}

func (r *FeatureRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *FeatureRepository) List(ctx context.Context) ([]*models.Feature, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var features []*models.Feature
	if err = cursor.All(ctx, &features); err != nil {
		return nil, err
	}
	return features, nil
}

func (r *FeatureRepository) BulkUpdate(ctx context.Context, ids []primitive.ObjectID, update bson.M) error {
	_, err := r.collection.UpdateMany(
		ctx,
		bson.M{"_id": bson.M{"$in": ids}},
		bson.M{"$set": update},
	)
	return err
}
