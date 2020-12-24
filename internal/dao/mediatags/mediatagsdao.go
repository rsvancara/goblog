//Package mediatagsdao media data access object
package mediatagsdao

import (
	"context"
	"time"

	"github.com/rsvancara/goblog/internal/models"

	"github.com/rs/zerolog/log"
	"github.com/rsvancara/goblog/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//MediaTagsDAO stores media data access object information
type MediaTagsDAO struct {
	DBClient *mongo.Client
	Config   *config.AppConfig
}

// Initialize creates the connection and populates the suppression struct
func (m *MediaTagsDAO) Initialize(mclient *mongo.Client, config *config.AppConfig) error {

	err := mclient.Ping(context.TODO(), nil)

	if err != nil {
		log.Error().Err(err).Msg("Error connecting to mongodb")
	}

	log.Info().Msg("MediaTagsDAO connected successfully to mongodb")

	m.DBClient = mclient
	m.Config = config

	return nil
}

//InsertMediaTags insert media
func (m *MediaTagsDAO) InsertMediaTags(mediatag *models.MediaTagsModel) error {

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("mediatags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	insertResult, err := c.InsertOne(ctx, mediatag)
	if err != nil {
		return err
	}

	// Convert to object ID
	mediatag.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil
}

//UpdateMediaTags Update the title, keywords and description for media
func (m *MediaTagsDAO) UpdateMediaTags(mediatag *models.MediaTagsModel) error {

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("mediatags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	filter := bson.M{
		"tags_id": bson.M{
			"$eq": mediatag.TagsID, // check if bool field has value of 'false'
		},
	}

	//m.Documents = TagExtractor(m.Keywords)
	//m.Slug = slug.Make(m.Title)

	update := bson.M{
		"$set": bson.M{
			"name":      mediatag.Name,
			"documents": mediatag.Documents,
		},
	}

	_, err := c.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

// GetMediaTag populate the media object based on ID
func (m *MediaTagsDAO) GetMediaTag(id string) error {

	var mediatag models.MediaTagsModel

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("mediatags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err := c.FindOne(ctx, bson.M{"tags_id": id}).Decode(&mediatag)
	if err != nil {
		return err
	}

	return nil
}

// Exists Check to see if record exists and if it does return it
func (m *MediaTagsDAO) Exists(name string) (int64, error) {

	var count int64

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("mediatags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	count, err := c.CountDocuments(ctx, bson.M{"name": name})
	if err != nil {
		return 0, err
	}

	return count, nil

}

// GetMediaTagsCount get the total number of media tags
func (m *MediaTagsDAO) GetMediaTagsCount() (int64, error) {

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("mediatags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	count, err := c.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetMediaTagByName populate the media object based on ID
func (m *MediaTagsDAO) GetMediaTagByName(name string) error {

	var mediatag models.MediaTagsModel

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("mediatags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err := c.FindOne(ctx, bson.M{"name": name}).Decode(&mediatag)
	if err != nil {
		return err
	}

	return nil
}

// SearchMediaTagsByName Text search for MediaTags
func (m *MediaTagsDAO) SearchMediaTagsByName(name string) ([]models.MediaTagsModel, error) {

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("mediatags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var mediaTagsModels []models.MediaTagsModel

	filter := bson.M{
		"$text": bson.M{
			"$search": name,
		},
	}

	options := options.Find()

	// Sort by `_id` field descending
	options.SetSort(map[string]int{"name": 1})

	cur, err := c.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var mediatag models.MediaTagsModel
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&mediatag)
		if err != nil {
			return nil, err
		}

		mediaTagsModels = append(mediaTagsModels, mediatag)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return mediaTagsModels, nil

}

//DeleteAllTags Delete all tags, used when index needs to be rebuilt
func (m *MediaTagsDAO) DeleteAllTags() error {

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("mediatags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	_, err := c.DeleteMany(ctx, bson.M{})
	if err != nil {
		return err
	}

	return nil
}

//AllMediaTags retrieve all media tags
func (m *MediaTagsDAO) AllMediaTags() ([]models.MediaTagsModel, error) {

	var mediaTagsModels []models.MediaTagsModel

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("mediatags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	filter := bson.M{}

	options := options.Find()

	// Sort by `_id` field descending
	options.SetSort(map[string]int{"name": 1})

	cur, err := c.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var mediatag models.MediaTagsModel
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&m)
		if err != nil {
			return nil, err
		}

		mediaTagsModels = append(mediaTagsModels, mediatag)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return mediaTagsModels, nil
}
