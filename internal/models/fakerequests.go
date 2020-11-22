package models

import (
	"context"
	"time"

	"github.com/rsvancara/goblog/internal/db"

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

//InsertFakeRequest insert FakeRequest
func (f *FakeRequest) InsertFakeRequest() error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	// Manage the create and update time
	f.CreatedAt = time.Now()
	f.FakeRequestID = GenUUID()

	c := db.Client.Database(getPostDB()).Collection("fakerequests")

	insertResult, err := c.InsertOne(context.TODO(), f)
	if err != nil {
		return err
	}

	// Convert to object ID
	f.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil
}
