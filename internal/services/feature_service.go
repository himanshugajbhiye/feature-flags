package services

import (
	"context"
	"errors"
	"feature-flags/internal/models"
	"feature-flags/internal/repository/mongodb"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FeatureService struct {
	featureRepo    *mongodb.FeatureRepository
	dependencyRepo *mongodb.FeatureDependencyRepository
}

func NewFeatureService(featureRepo *mongodb.FeatureRepository, dependencyRepo *mongodb.FeatureDependencyRepository) *FeatureService {
	return &FeatureService{
		featureRepo:    featureRepo,
		dependencyRepo: dependencyRepo,
	}
}

func (s *FeatureService) CreateFeature(ctx context.Context, feature *models.Feature) error {
	return s.featureRepo.Create(ctx, feature)
}

func (s *FeatureService) AddChild(ctx context.Context, parentID, childID primitive.ObjectID) error {
	// Check for cyclic dependency
	if err := s.checkCyclicDependency(ctx, parentID, childID); err != nil {
		return err
	}

	// Check if dependency already exists
	exists, err := s.dependencyRepo.Exists(ctx, parentID, childID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("dependency already exists")
	}

	// Create the dependency
	dependency := &models.FeatureDependency{
		ParentID: parentID,
		ChildID:  childID,
	}
	return s.dependencyRepo.Create(ctx, dependency)
}

func (s *FeatureService) DisableFeature(ctx context.Context, id primitive.ObjectID) error {
	// Queue for BFS
	queue := []primitive.ObjectID{id}
	// Track all features to disable
	featuresToDisable := make([]primitive.ObjectID, 0)
	// Track disabled features for logging
	disabledFeatures := make([]string, 0)

	// First collect all features to disable
	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		feature, err := s.featureRepo.GetByID(ctx, currentID)
		if err != nil {
			return fmt.Errorf("failed to get feature: %w", err)
		}

		// Skip if already disabled
		if !feature.IsEnabled {
			continue
		}

		// Add to features to disable
		featuresToDisable = append(featuresToDisable, currentID)
		disabledFeatures = append(disabledFeatures, feature.Name)

		// Get children and add to queue
		children, err := s.dependencyRepo.GetChildren(ctx, currentID)
		if err != nil {
			return fmt.Errorf("failed to get children: %w", err)
		}
		queue = append(queue, children...)
	}

	// Disable all features in one bulk update
	if len(featuresToDisable) > 0 {
		if err := s.featureRepo.BulkUpdate(ctx, featuresToDisable, bson.M{
			"is_enabled": false,
			"updated_at": time.Now(),
		}); err != nil {
			return fmt.Errorf("failed to bulk disable features: %w", err)
		}
	}

	log.Printf("Successfully disabled %d features: %v", len(disabledFeatures), disabledFeatures)
	return nil
}

func (s *FeatureService) EnableFeature(ctx context.Context, id primitive.ObjectID) error {
	feature, err := s.featureRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Get all parents
	parents, err := s.dependencyRepo.GetParents(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get parents: %w", err)
	}

	// Check if any parent is disabled
	for _, parentID := range parents {
		parent, err := s.featureRepo.GetByID(ctx, parentID)
		if err != nil {
			return fmt.Errorf("failed to get parent feature: %w", err)
		}
		if !parent.IsEnabled {
			return errors.New("cannot enable feature: parent feature is disabled")
		}
	}

	feature.IsEnabled = true
	feature.UpdatedAt = time.Now()
	return s.featureRepo.Update(ctx, feature)
}

func (s *FeatureService) checkCyclicDependency(ctx context.Context, parentID, childID primitive.ObjectID) error {
	if parentID == childID {
		return errors.New("cannot add self as child")
	}

	visited := make(map[primitive.ObjectID]bool)
	return s.dfs(ctx, childID, parentID, visited)
}

func (s *FeatureService) dfs(ctx context.Context, current, target primitive.ObjectID, visited map[primitive.ObjectID]bool) error {
	if current == target {
		return errors.New("cyclic dependency detected")
	}

	if visited[current] {
		return nil
	}

	visited[current] = true

	// Get children of current feature
	children, err := s.dependencyRepo.GetChildren(ctx, current)
	if err != nil {
		return err
	}

	for _, childID := range children {
		if err := s.dfs(ctx, childID, target, visited); err != nil {
			return err
		}
	}

	return nil
}

func (s *FeatureService) GetFeatureStatus(ctx context.Context, id primitive.ObjectID) (*models.Feature, error) {
	feature, err := s.featureRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get feature: %w", err)
	}
	return feature, nil
}
