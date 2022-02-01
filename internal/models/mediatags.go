package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaTags tags associated with media
type MediaTagsModel struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	TagsID    string             `json:"tags_id" bson:"tags_id,omitempty"`     // Unique identifier
	Name      string             `json:"name" bson:"name,omitempty"`           // Tag Key word
	Documents []string           `json:"documents" bson:"documents,omitempty"` // List of Document IDs
}
