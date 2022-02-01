package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Affiliate affiliate link
type Affiliate struct {
	ID             primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	AffiliateID    string             `json:"affiliate_id" bson:"affiliate_id,omitempty"`
	AffiliateLink  string             `json:"affiliate_link" bson:"affiliate_link,omitempty"`
	Description    string             `json:"description" bson:"description,omitempty"`
	AffiliateTitle string             `json:"title" bson:"title,omitempty"`
	Slug           string             `json:"slug" bson:"slug,omitempty"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"` //CreatedAt date record was created
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"` //UpdatedAt date record was updated
	Category       string             `json:"category" bson:"category,omitempty"`
}
