package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FeatureType string

const (
	FeatureTypeBasic      FeatureType = "basic"
	FeatureTypePremium    FeatureType = "premium"
	FeatureTypeEnterprise FeatureType = "enterprise"
)

type Feature struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Type      FeatureType        `bson:"type" json:"type"`
	IsEnabled bool               `bson:"is_enabled" json:"is_enabled"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
