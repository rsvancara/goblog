package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// FakeRequest post
type FakeRequest struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	FakeRequestID string             `json:"fakerequest_id" bson:"fakerequest_id,omitempty"`
	URL           string             `json:"url" bson:"url,omitempty"`
	CreatedAt     time.Time          `json:"created_at" bson:"created_at,omitempty"`
	IPAddress     string             `json:"ipaddress" bson:"ip_address,omitempty"`
	UserAgent     string             `json:"useragent" bson:"useragent,omitempty"`
	City          string             `json:"city" bson:"city,omitempty"`
	TimeZone      string             `json:"timezone" bson:"timezone,omitempty" `
	Country       string             `json:"country" bson:"country,omitempty"`
	Username      string             `json:"username" bson:"username,omitempty"`
	Password      string             `json:"password" bson:"password,omitempty"`
	SessionID     string             `json:"session_id" bson:"session_id,omitempty"`
}
