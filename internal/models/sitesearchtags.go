package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SiteSearchTags tags associated with sitesearch
type SiteSearchTagsModel struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	TagsID    string             `json:"tags_id" bson:"tags_id,omitempty"`     // Unique identifier
	Name      string             `json:"name" bson:"name,omitempty"`           // Tag Key word
	Documents []string           `json:"documents" bson:"documents,omitempty"` // List of Document IDs
	DocType   string             `json:"doctype" bson:"doctype,omitempty"`     // Document Type
}
