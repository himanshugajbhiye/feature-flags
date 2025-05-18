package services

import (
	"context"
	"feature-flags/internal/models"
	"feature-flags/internal/repository/mongodb"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	testMongoURI = "mongodb://admin:password123@localhost:27017"
	testDBName   = "feature_flags_test"
)

func setupTestDB(t *testing.T) (*mongo.Database, func()) {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(testMongoURI))
	require.NoError(t, err)

	db := client.Database(testDBName)

	// Clean up function
	cleanup := func() {
		err := db.Drop(ctx)
		require.NoError(t, err)
		err = client.Disconnect(ctx)
		require.NoError(t, err)
	}

	return db, cleanup
}

func setupFeatureService(t *testing.T) (*FeatureService, func()) {
	db, cleanup := setupTestDB(t)

	featureRepo := mongodb.NewFeatureRepository(db)
	dependencyRepo := mongodb.NewFeatureDependencyRepository(db)
	service := NewFeatureService(featureRepo, dependencyRepo)

	return service, cleanup
}

func TestFeatureService_CreateFeature(t *testing.T) {
	service, cleanup := setupFeatureService(t)
	defer cleanup()

	ctx := context.Background()
	feature := &models.Feature{
		Name:      "test-feature",
		Type:      models.FeatureTypeBasic,
		IsEnabled: true,
	}

	err := service.CreateFeature(ctx, feature)
	require.NoError(t, err)
	assert.NotEmpty(t, feature.ID)

	// Verify feature was created
	retrieved, err := service.GetFeatureStatus(ctx, feature.ID)
	require.NoError(t, err)
	assert.Equal(t, feature.Name, retrieved.Name)
	assert.Equal(t, feature.Type, retrieved.Type)
	assert.Equal(t, feature.IsEnabled, retrieved.IsEnabled)
}

func TestFeatureService_AddChild(t *testing.T) {
	service, cleanup := setupFeatureService(t)
	defer cleanup()

	ctx := context.Background()

	// Create parent feature
	parent := &models.Feature{
		Name:      "parent-feature",
		Type:      models.FeatureTypeBasic,
		IsEnabled: true,
	}
	err := service.CreateFeature(ctx, parent)
	require.NoError(t, err)

	// Create child feature
	child := &models.Feature{
		Name:      "child-feature",
		Type:      models.FeatureTypeBasic,
		IsEnabled: true,
	}
	err = service.CreateFeature(ctx, child)
	require.NoError(t, err)

	// Test adding child
	err = service.AddChild(ctx, parent.ID, child.ID)
	require.NoError(t, err)

	// Test adding same child again (should fail)
	err = service.AddChild(ctx, parent.ID, child.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dependency already exists")

	// Test adding self as child (should fail)
	err = service.AddChild(ctx, parent.ID, parent.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot add self as child")
}

func TestFeatureService_DisableFeature(t *testing.T) {
	service, cleanup := setupFeatureService(t)
	defer cleanup()

	ctx := context.Background()

	// Create parent feature
	parent := &models.Feature{
		Name:      "parent-feature",
		Type:      models.FeatureTypeBasic,
		IsEnabled: true,
	}
	err := service.CreateFeature(ctx, parent)
	require.NoError(t, err)

	// Create child feature
	child := &models.Feature{
		Name:      "child-feature",
		Type:      models.FeatureTypeBasic,
		IsEnabled: true,
	}
	err = service.CreateFeature(ctx, child)
	require.NoError(t, err)

	// Add child to parent
	err = service.AddChild(ctx, parent.ID, child.ID)
	require.NoError(t, err)

	// Disable parent feature
	err = service.DisableFeature(ctx, parent.ID)
	require.NoError(t, err)

	// Verify both features are disabled
	parentStatus, err := service.GetFeatureStatus(ctx, parent.ID)
	require.NoError(t, err)
	assert.False(t, parentStatus.IsEnabled)

	childStatus, err := service.GetFeatureStatus(ctx, child.ID)
	require.NoError(t, err)
	assert.False(t, childStatus.IsEnabled)
}

func TestFeatureService_EnableFeature(t *testing.T) {
	service, cleanup := setupFeatureService(t)
	defer cleanup()

	ctx := context.Background()

	// Create parent feature
	parent := &models.Feature{
		Name:      "parent-feature",
		Type:      models.FeatureTypeBasic,
		IsEnabled: false,
	}
	err := service.CreateFeature(ctx, parent)
	require.NoError(t, err)

	// Create child feature
	child := &models.Feature{
		Name:      "child-feature",
		Type:      models.FeatureTypeBasic,
		IsEnabled: false,
	}
	err = service.CreateFeature(ctx, child)
	require.NoError(t, err)

	// Add child to parent
	err = service.AddChild(ctx, parent.ID, child.ID)
	require.NoError(t, err)

	// Try to enable child while parent is disabled (should fail)
	err = service.EnableFeature(ctx, child.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parent feature is disabled")

	// Enable parent
	err = service.EnableFeature(ctx, parent.ID)
	require.NoError(t, err)

	// Now enable child
	err = service.EnableFeature(ctx, child.ID)
	require.NoError(t, err)

	// Verify both features are enabled
	parentStatus, err := service.GetFeatureStatus(ctx, parent.ID)
	require.NoError(t, err)
	assert.True(t, parentStatus.IsEnabled)

	childStatus, err := service.GetFeatureStatus(ctx, child.ID)
	require.NoError(t, err)
	assert.True(t, childStatus.IsEnabled)
}

func TestFeatureService_CyclicDependency(t *testing.T) {
	service, cleanup := setupFeatureService(t)
	defer cleanup()

	ctx := context.Background()

	// Create three features
	feature1 := &models.Feature{
		Name:      "feature-1",
		Type:      models.FeatureTypeBasic,
		IsEnabled: true,
	}
	err := service.CreateFeature(ctx, feature1)
	require.NoError(t, err)

	feature2 := &models.Feature{
		Name:      "feature-2",
		Type:      models.FeatureTypeBasic,
		IsEnabled: true,
	}
	err = service.CreateFeature(ctx, feature2)
	require.NoError(t, err)

	feature3 := &models.Feature{
		Name:      "feature-3",
		Type:      models.FeatureTypeBasic,
		IsEnabled: true,
	}
	err = service.CreateFeature(ctx, feature3)
	require.NoError(t, err)

	// Create a cycle: feature1 -> feature2 -> feature3 -> feature1
	err = service.AddChild(ctx, feature1.ID, feature2.ID)
	require.NoError(t, err)

	err = service.AddChild(ctx, feature2.ID, feature3.ID)
	require.NoError(t, err)

	// This should fail due to cyclic dependency
	err = service.AddChild(ctx, feature3.ID, feature1.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cyclic dependency detected")
}
