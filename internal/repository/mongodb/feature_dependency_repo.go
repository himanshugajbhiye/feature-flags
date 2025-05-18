package mongodb

import (
	"context"
	"feature-flags/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type FeatureDependencyRepository struct {
	collection *mongo.Collection
}

func NewFeatureDependencyRepository(db *mongo.Database) *FeatureDependencyRepository {
	return &FeatureDependencyRepository{
		collection: db.Collection("feature_dependencies"),
	}
}

func (r *FeatureDependencyRepository) Create(ctx context.Context, dependency *models.FeatureDependency) error {
	dependency.CreatedAt = time.Now()
	dependency.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, dependency)
	if err != nil {
		return err
	}

	dependency.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *FeatureDependencyRepository) GetChildren(ctx context.Context, parentID primitive.ObjectID) ([]primitive.ObjectID, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"parent_id": parentID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var dependencies []models.FeatureDependency
	if err := cursor.All(ctx, &dependencies); err != nil {
		return nil, err
	}

	children := make([]primitive.ObjectID, len(dependencies))
	for i, dep := range dependencies {
		children[i] = dep.ChildID
	}
	return children, nil
}

func (r *FeatureDependencyRepository) GetParents(ctx context.Context, childID primitive.ObjectID) ([]primitive.ObjectID, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"child_id": childID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var dependencies []models.FeatureDependency
	if err := cursor.All(ctx, &dependencies); err != nil {
		return nil, err
	}

	parents := make([]primitive.ObjectID, len(dependencies))
	for i, dep := range dependencies {
		parents[i] = dep.ParentID
	}
	return parents, nil
}

func (r *FeatureDependencyRepository) Delete(ctx context.Context, parentID, childID primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{
		"parent_id": parentID,
		"child_id":  childID,
	})
	return err
}

func (r *FeatureDependencyRepository) Exists(ctx context.Context, parentID, childID primitive.ObjectID) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"parent_id": parentID,
		"child_id":  childID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
