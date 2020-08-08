package models

import (
	"blog/blog/config"
	"blog/blog/db"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

// MediaTags tags associated with media
type MediaTagsModel struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	TagsID    string             `json:"tags_id" bson:"tags_id,omitempty"`     // Unique identifier
	Name      string             `json:"name" bson:"name,omitempty"`           // Tag Key word
	Documents []string           `json:"documents" bson:"documents,omitempty"` // List of Document IDs
}

//InsertMediaTags insert media
func (m *MediaTagsModel) InsertMediaTags() error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	// Manage the create and update time

	c := db.Client.Database(getMediaTagsDB()).Collection("mediatags")

	insertResult, err := c.InsertOne(context.TODO(), m)
	if err != nil {
		return err
	}

	// Convert to object ID
	m.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil
}

//UpdateMediaTags Update the title, keywords and description for media
func (m *MediaTagsModel) UpdateMediaTags() error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	c := db.Client.Database(getMediaTagsDB()).Collection("mediatags")

	filter := bson.M{
		"tags_id": bson.M{
			"$eq": m.TagsID, // check if bool field has value of 'false'
		},
	}

	//m.Documents = TagExtractor(m.Keywords)
	//m.Slug = slug.Make(m.Title)

	update := bson.M{
		"$set": bson.M{
			"name":      m.Name,
			"documents": m.Documents,
		},
	}

	result, err := c.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	fmt.Printf("updated %v record for media tag id  %s \n", result.ModifiedCount, m.TagsID)

	return nil
}

// GetMediaTag populate the media object based on ID
func (m *MediaTagsModel) GetMediaTag(id string) error {

	var db db.Session

	err := db.NewSession()

	c := db.Client.Database(getMediaTagsDB()).Collection("mediatags")

	err = c.FindOne(context.TODO(), bson.M{"tags_id": id}).Decode(m)
	if err != nil {
		return err
	}
	defer db.Close()

	return nil
}

// Exists Check to see if record exists and if it does return it
func (m *MediaTagsModel) Exists(name string) (int64, error) {
	var db db.Session

	err := db.NewSession()

	var count int64

	c := db.Client.Database(getMediaTagsDB()).Collection("mediatags")

	count, err = c.CountDocuments(context.TODO(), bson.M{"name": name})
	if err != nil {
		return 0, err
	}
	defer db.Close()

	return count, nil

}

// GetMediaTagsCount get the total number of media tags
func (m *MediaTagsModel) GetMediaTagsCount() (int64, error) {
	var db db.Session

	err := db.NewSession()
	if err != nil {
		return 0, err
	}
	c := db.Client.Database(getMediaTagsDB()).Collection("mediatags")

	count, err := c.CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetMediaTagByName populate the media object based on ID
func (m *MediaTagsModel) GetMediaTagByName(name string) error {

	var db db.Session

	err := db.NewSession()

	c := db.Client.Database(getMediaTagsDB()).Collection("mediatags")

	err = c.FindOne(context.TODO(), bson.M{"name": name}).Decode(m)
	if err != nil {
		return err
	}
	defer db.Close()

	return nil
}

// SearchMediaTagsByName Text search for MediaTags
func SearchMediaTagsByName(name string) ([]MediaTagsModel, error) {

	//var config db.Config
	var db db.Session

	var mediaTagsModels []MediaTagsModel

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

	cur, err := db.Client.Database(getMediaTagsDB()).Collection("mediatags").Find(context.TODO(), filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var m MediaTagsModel
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&m)
		if err != nil {
			return nil, err
		}

		mediaTagsModels = append(mediaTagsModels, m)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return mediaTagsModels, nil

}

//DeleteAllTags Delete all tags, used when index needs to be rebuilt
func DeleteAllTags() error {
	//var config db.Config
	var db db.Session
	err := db.NewSession()
	if err != nil {
		return err
	}
	defer db.Close()

	result, err := db.Client.Database(getMediaTagsDB()).Collection("mediatags").DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		return err
	}

	fmt.Printf("Deleted %v documents", result.DeletedCount)

	return nil
}

//AllMediaTags retrieve all media tags
func AllMediaTags() ([]MediaTagsModel, error) {

	//var config db.Config
	var db db.Session

	var mediaTagsModels []MediaTagsModel

	err := db.NewSession()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	filter := bson.M{}

	options := options.Find()

	// Sort by `_id` field descending
	options.SetSort(map[string]int{"name": 1})

	cur, err := db.Client.Database(getMediaTagsDB()).Collection("mediatags").Find(context.TODO(), filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var m MediaTagsModel
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&m)
		if err != nil {
			return nil, err
		}

		mediaTagsModels = append(mediaTagsModels, m)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return mediaTagsModels, nil
}

func getMediaTagsDB() string {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("error getting configuration: %s", err)
		return ""
	}

	return cfg.MongoDatabase

}
