package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FeatureDependency struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ParentID  primitive.ObjectID `bson:"parent_id" json:"parent_id"`
	ChildID   primitive.ObjectID `bson:"child_id" json:"child_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
