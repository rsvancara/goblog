package models

import (
	"context"
	"fmt"

	"goblog/internal/config"
	"goblog/internal/db"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SiteSearchTags tags associated with sitesearch
type SiteSearchTagsModel struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	TagsID    string             `json:"tags_id" bson:"tags_id,omitempty"`     // Unique identifier
	Name      string             `json:"name" bson:"name,omitempty"`           // Tag Key word
	Documents []string           `json:"documents" bson:"documents,omitempty"` // List of Document IDs
	DocType   string             `json:"doctype" bson:"doctype,omitempty"`     // Document Type
}

//InsertSiteSearchtags insert sitesearch
func (s *SiteSearchTagsModel) InsertSiteSearchTags() error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	// Manage the create and update time

	c := db.Client.Database(getSiteSearchTagsDB()).Collection("sitesearchtags")

	insertResult, err := c.InsertOne(context.TODO(), s)
	if err != nil {
		return err
	}

	// Convert to object ID
	s.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil
}

//UpdateSearchTags Update the title, keywords and description for sitesearch
func (s *SiteSearchTagsModel) UpdateSiteSearchTags() error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	c := db.Client.Database(getSiteSearchTagsDB()).Collection("sitesearchtags")

	filter := bson.M{
		"tags_id": bson.M{
			"$eq": s.TagsID, // check if bool field has value of 'false'
		},
	}

	//m.Documents = TagExtractor(m.Keywords)
	//m.Slug = slug.Make(m.Title)

	update := bson.M{
		"$set": bson.M{
			"name":      s.Name,
			"documents": s.Documents,
		},
	}

	result, err := c.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	log.Info().Msgf("updated %v record for sitesearch tag id  %s \n", result.ModifiedCount, s.TagsID)

	return nil
}

// GetSiteSearchTag populate the sitesearch object based on ID
func (s *SiteSearchTagsModel) GetSiteSearchTag(id string) error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	c := db.Client.Database(getSiteSearchTagsDB()).Collection("sitesearchtags")

	err = c.FindOne(context.TODO(), bson.M{"tags_id": id}).Decode(s)
	if err != nil {
		return err
	}
	defer db.Close()

	return nil
}

// Exists Check to see if record exists and if it does return it
func (s *SiteSearchTagsModel) Exists(name string) (int64, error) {
	var db db.Session

	err := db.NewSession()
	if err != nil {
		return 0.0, err
	}

	var count int64

	c := db.Client.Database(getSiteSearchTagsDB()).Collection("sitesearchtags")

	count, err = c.CountDocuments(context.TODO(), bson.M{"name": name})
	if err != nil {
		return 0, err
	}
	defer db.Close()

	return count, nil

}

// GetSiteSearchTagsCount get the total number of Sitesearch tags
func (s *SiteSearchTagsModel) GetSiteSearchTagsCount() (int64, error) {
	var db db.Session

	err := db.NewSession()
	if err != nil {
		return 0, err
	}
	c := db.Client.Database(getSiteSearchTagsDB()).Collection("sitesearchtags")

	count, err := c.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetSitesearchTagByName populate the Sitesearch object based on ID
func (s *SiteSearchTagsModel) GetSiteSearchTagsByName(name string) error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	c := db.Client.Database(getSiteSearchTagsDB()).Collection("sitesearchtags")

	err = c.FindOne(context.TODO(), bson.M{"name": name}).Decode(s)
	if err != nil {
		return err
	}
	defer db.Close()

	return nil
}

// Search SiteSearchTagsByName Text search for Site Search Tags
func SearchSiteSearchTagsByName(name string) ([]SiteSearchTagsModel, error) {

	//var config db.Config
	var db db.Session

	var siteSearchTagsModel []SiteSearchTagsModel

	err := db.NewSession()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	filter := bson.M{
		"$text": bson.M{
			"$search": name,
		},
	}

	options := options.Find()

	// Sort by `_id` field descending
	options.SetSort(map[string]int{"name": 1})

	cur, err := db.Client.Database(getSiteSearchTagsDB()).Collection("sitesearchtags").Find(context.TODO(), filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var s SiteSearchTagsModel
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&s)
		if err != nil {
			return nil, err
		}

		siteSearchTagsModel = append(siteSearchTagsModel, s)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return siteSearchTagsModel, nil

}

//DeleteAllTags Delete all tags, used when index needs to be rebuilt
func DeleteAllSiteSearchTags() error {
	//var config db.Config
	var db db.Session
	err := db.NewSession()
	if err != nil {
		return err
	}
	defer db.Close()

	result, err := db.Client.Database(getSiteSearchTagsDB()).Collection("sitesearchtags").DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		return err
	}

	fmt.Printf("Deleted %v documents", result.DeletedCount)

	return nil
}

//All SiteSearchTags retrieve all sitesearch tags
func AllSiteSearchTags() ([]SiteSearchTagsModel, error) {

	//var config db.Config
	var db db.Session

	var siteSearchTagsModel []SiteSearchTagsModel

	err := db.NewSession()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	filter := bson.M{}

	options := options.Find()

	// Sort by `_id` field descending
	options.SetSort(map[string]int{"name": 1})

	cur, err := db.Client.Database(getSiteSearchTagsDB()).Collection("sitesearchtags").Find(context.TODO(), filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var s SiteSearchTagsModel
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&s)
		if err != nil {
			return nil, err
		}

		siteSearchTagsModel = append(siteSearchTagsModel, s)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return siteSearchTagsModel, nil
}

func getSiteSearchTagsDB() string {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("error getting configuration: %s", err)
		return ""
	}

	return cfg.MongoDatabase

}
