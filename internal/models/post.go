package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PostModel post
type PostModel struct {
	ID              primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	PostID          string             `json:"postid" bson:"postid,omitempty"`
	Slug            string             `json:"slug" bson:"slug,omitempty"`
	Post            string             `json:"post" bson:"post,omitempty"`
	PostTeaser      string             `json:"post_teaser" bson:"post_teaser,omitempty"`
	TeaserImage     string             `json:"teaser_image" bson:"teaser_image,omitempty"`
	TeaserImageSlug string             `json:"teaser_image_slug" bson:"teaser_image_slug,omitempty"`
	Title           string             `json:"title" bson:"title,omitempty"`
	Keywords        string             `json:"keywords" bson:"keywords,omitempty"`
	Tags            []Tag              `json:"tags" bson:"tags,omitempty"`
	Status          string             `json:"status" bson:"status,omitempty"`
	Featured        string             `json:"featured" bson:"featured,omitempty"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at"`
}
