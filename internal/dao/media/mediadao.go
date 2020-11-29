//Package mediadao media data access object
package mediadao

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/rsvancara/goblog/internal/config"
	"github.com/rsvancara/goblog/internal/db"
	"github.com/rsvancara/goblog/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//MediaDao stores media data access object information
type MediaDao struct {
	DBClient *mongo.Client
	Config   *config.AppConfig
}

// Initialize creates the connection and populates the suppression struct
func (m *MediaDao) Initialize(mclient *mongo.Client, config *config.AppConfig) error {
	m.DBClient = mclient
	m.Config = config

	return nil
}

//InsertMedia insert media
func (m *MediaDao) InsertMedia(media models.MediaModel) error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	// Manage the create and update time
	media.CreatedAt = time.Now()
	media.UpdatedAt = time.Now()
	media.MediaID = models.GenUUID()
	media.Slug = slug.Make(media.Title)

	media.Tags = models.TagExtractor(media.Keywords)

	c := db.Client.Database(m.Config.MongoDatabase).Collection("media")

	insertResult, err := c.InsertOne(context.TODO(), m)
	if err != nil {
		return err
	}

	// Convert to object ID
	media.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil
}

//UpdateMedia Update the title, keywords and description for media
func (m *MediaDao) UpdateMedia(media models.MediaModel) error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	c := db.Client.Database(m.Config.MongoDatabase).Collection("media")

	filter := bson.M{
		"media_id": bson.M{
			"$eq": media.MediaID, // check if bool field has value of 'false'
		},
	}

	media.Tags = models.TagExtractor(media.Keywords)
	media.Slug = slug.Make(media.Title)

	update := bson.M{
		"$set": bson.M{
			"keywords":    media.Keywords,
			"title":       media.Title,
			"description": media.Description,
			"tags":        media.Tags,
			"category":    media.Category,
			"slug":        media.Slug,
			"location":    media.Location,
		},
	}

	result, err := c.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	fmt.Printf("updated %v record for media id  %s \n", result.ModifiedCount, media.MediaID)

	return nil
}

// GetMedia populate the media object based on ID
func (m *MediaDao) GetMedia(id string) (models.MediaModel, error) {

	var db db.Session

	var media models.MediaModel

	//config.DBUri = "mongodb://host.docker.internal:27017"
	err := db.NewSession()

	c := db.Client.Database(m.Config.MongoDatabase).Collection("media")

	err = c.FindOne(context.TODO(), bson.M{"media_id": id}).Decode(m)

	// Translate special variables
	media.ExposureProgramTranslated = media.GetExposureProgramTranslated()
	media.FStopTranslated = media.CalculateFSTOP()

	if err != nil {
		return media, err
	}
	defer db.Close()

	return media, nil
}

// GetMediaBySlug populate the media object based on ID
func (m *MediaDao) GetMediaBySlug(slug string) (models.MediaModel, error) {

	var db db.Session

	var media models.MediaModel

	//config.DBUri = "mongodb://host.docker.internal:27017"
	err := db.NewSession()

	c := db.Client.Database(m.Config.MongoDatabase).Collection("media")

	err = c.FindOne(context.TODO(), bson.M{"slug": slug}).Decode(media)
	// Translate special variables
	media.ExposureProgramTranslated = media.GetExposureProgramTranslated()
	media.FStopTranslated = media.CalculateFSTOP()

	if err != nil {
		return media, err
	}
	defer db.Close()

	return media, nil
}

//DeleteMedia delete the media object base on ID
func (m *MediaDao) DeleteMedia(media models.MediaModel) error {

	//var config db.Config
	var db db.Session

	//config.DBUri = "mongodb://host.docker.internal:27017"
	err := db.NewSession()
	if err != nil {
		return err
	}
	defer db.Close()

	c := db.Client.Database(m.Config.MongoDatabase).Collection("media")
	_, err = c.DeleteOne(context.TODO(), bson.M{"media_id": media.MediaID})

	if err != nil {
		return err
	}

	return nil
}

//GetMediaListByCategory Obtains the list of media by category sorted by date
func (m *MediaDao) GetMediaListByCategory(category string) ([]models.MediaModel, error) {
	//var config db.Config
	var db db.Session

	var mediaModels []models.MediaModel
	//config.DBUri = "mongodb://host.docker.internal:27017"

	err := db.NewSession()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	filter := bson.M{"category": category}

	options := options.Find()

	// Sort by `_id` field descending
	options.SetSort(map[string]int{"created_at": -1})

	cur, err := db.Client.Database(m.Config.MongoDatabase).Collection("media").Find(context.TODO(), filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var media models.MediaModel
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&m)

		// Translate special variables
		media.ExposureProgramTranslated = media.GetExposureProgramTranslated()
		media.FStopTranslated = media.CalculateFSTOP()

		if err != nil {
			return nil, err
		}

		mediaModels = append(mediaModels, media)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return mediaModels, nil
}

//AllMediaSortedByDate retrieve all posts sorted by creation date
func (m *MediaDao) AllMediaSortedByDate() ([]models.MediaModel, error) {

	//var config db.Config
	var db db.Session

	var mediaModels []models.MediaModel
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

	cur, err := db.Client.Database(m.Config.MongoDatabase).Collection("media").Find(context.TODO(), filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var media models.MediaModel
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&media)

		// Translate special variables
		media.ExposureProgramTranslated = media.GetExposureProgramTranslated()
		media.FStopTranslated = media.CalculateFSTOP()

		if err != nil {
			return nil, err
		}

		mediaModels = append(mediaModels, media)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return mediaModels, nil
}

//AllCategories return a list of categories
func (m *MediaDao) AllCategories() ([]string, error) {
	//var config db.Config
	var db db.Session

	err := db.NewSession()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	filter := bson.M{}

	options := options.Distinct()

	cur, err := db.Client.Database(m.Config.MongoDatabase).Collection("media").Distinct(context.TODO(), "category", filter, options)
	if err != nil {
		return nil, err
	}

	var categories []string

	for _, v := range cur {
		categories = append(categories, v.(string))
	}

	return categories, nil

}

//MediaSearch retrieve all posts sorted by creation date
func (m *MediaDao) MediaSearch(searchJSON string) ([]models.MediaModel, error) {

	r := strings.NewReader(searchJSON)

	var ms models.MediaSearchQuery
	err := json.NewDecoder(r).Decode(&ms)
	if err != nil {

		return nil, fmt.Errorf("Error converting search string to JSON with error %s", err)
	}

	fmt.Println(ms)

	var filter bson.D
	isSearch := false

	if ms.Title != "" {
		//filter = append(filter, bson.E{"title", ms.Title})
		filter = append(filter, bson.E{Key: "title", Value: bson.D{
			{"$regex", primitive.Regex{Pattern: fmt.Sprintf("^%s", ms.Title), Options: "i"}},
		}},
		)
		isSearch = true
	}

	if ms.Category != "" {
		filter = append(filter, bson.E{"category", ms.Category})
		isSearch = true
	}

	if ms.Tags != "" {

		filter = append(filter, bson.E{Key: "keywords", Value: bson.D{
			{"$regex", primitive.Regex{Pattern: fmt.Sprintf("%s", ms.Tags), Options: "i"}},
		}},
		)
		isSearch = true

		//filter = append(filter, bson.E{"category", ms.Category})
	}

	var cur *mongo.Cursor

	//var config db.Config
	var db db.Session

	err = db.NewSession()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	options := options.Find()

	// Sort by `_id` field descending
	options.SetSort(map[string]int{"created_at": -1})

	if isSearch == true {
		cur, err = db.Client.Database(m.Config.MongoDatabase).Collection("media").Find(context.TODO(), filter, options)
		if err != nil {
			return nil, err
		}
	} else {
		cur, err = db.Client.Database(m.Config.MongoDatabase).Collection("media").Find(context.TODO(), bson.M{}, options)
		if err != nil {
			return nil, err
		}
	}

	var mediaModels []models.MediaModel

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var m models.MediaModel
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&m)

		// Translate special variables
		m.ExposureProgramTranslated = m.GetExposureProgramTranslated()
		m.FStopTranslated = m.CalculateFSTOP()

		if err != nil {
			return nil, err
		}

		mediaModels = append(mediaModels, m)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return mediaModels, nil
}
