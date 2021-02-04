package models

import (
	"context"
	"fmt"
	"time"

	"goblog/internal/config"
	"goblog/internal/db"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// GetAffiliate get affiliate by ID
func (a *Affiliate) GetAffiliate(id string) error {
	var db db.Session

	//config.DBUri = "mongodb://host.docker.internal:27017"
	err := db.NewSession()

	c := db.Client.Database(getAffiliateDB()).Collection("affiliate")

	err = c.FindOne(context.TODO(), bson.M{"affiliate_id": id}).Decode(a)
	if err != nil {
		return err
	}
	defer db.Close()

	return nil
}

// InsertAffiliate create affiliate database entry
func (a *Affiliate) InsertAffiliate() error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	// Manage the create and update time
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
	a.AffiliateID = GenUUID()
	a.Slug = slug.Make(a.AffiliateTitle)

	c := db.Client.Database(getAffiliateDB()).Collection("affiliate")

	insertResult, err := c.InsertOne(context.TODO(), a)
	if err != nil {
		return err
	}

	// Convert to object ID
	a.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil
}

// EditAffiliate edit affiliate database entry
func (a *Affiliate) EditAffiliate() error {
	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	c := db.Client.Database(getAffiliateDB()).Collection("affiliate")

	filter := bson.M{
		"affiliate_id": bson.M{
			"$eq": a.AffiliateID, // check if bool field has value of 'false'
		},
	}

	a.UpdatedAt = time.Now()
	a.Slug = slug.Make(a.AffiliateTitle)

	update := bson.M{
		"$set": bson.M{
			"affiliate_link":  a.AffiliateLink,
			"affiliate_title": a.AffiliateTitle,
			"description":     a.Description,
			"Updated_at":      a.UpdatedAt,
			"category":        a.Category,
			"slug":            a.Slug,
		},
	}

	result, err := c.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	fmt.Printf("updated %v record for affiliate id  %s \n", result.ModifiedCount, a.AffiliateID)

	return nil
}

// GetAllAffiliateOrderByDate edit affiliate database entry
func GetAllAffiliateOrderByDate() ([]Affiliate, error) {
	//var config db.Config
	var db db.Session

	var affiliates []Affiliate
	//config.DBUri = "mongodb://host.docker.internal:27017"

	err := db.NewSession()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	filter := bson.M{}

	options := options.Find()

	// Sort by `_id` field descending
	options.SetSort(map[string]int{"created_at": -1})

	cur, err := db.Client.Database(getAffiliateDB()).Collection("affiliate").Find(context.TODO(), filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var a Affiliate
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&a)
		if err != nil {
			return nil, err
		}

		affiliates = append(affiliates, a)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return affiliates, nil
}

// DeleteAffiliate delete the affiliate
func (a *Affiliate) DeleteAffiliate() error {

	//var config db.Config
	var db db.Session

	//config.DBUri = "mongodb://host.docker.internal:27017"
	err := db.NewSession()
	if err != nil {
		return err
	}
	defer db.Close()

	c := db.Client.Database(getAffiliateDB()).Collection("affiliate")
	_, err = c.DeleteOne(context.TODO(), bson.M{"affiliate_id": a.AffiliateID})

	if err != nil {
		return err
	}

	return nil
}

func getAffiliateDB() string {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("error getting configuration: %s", err)
		return ""
	}

	return cfg.MongoDatabase

}
