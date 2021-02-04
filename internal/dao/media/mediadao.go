//Package mediadao media data access object
package mediadao

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"goblog/internal/config"
	"goblog/internal/models"
)

//MediaDAO stores media data access object information
type MediaDAO struct {
	DBClient *mongo.Client
	Config   *config.AppConfig
}

// MediaSearch Search Parameters
type MediaSearch struct {
	SearchString string
}

// Initialize creates the connection and populates the suppression struct
func (m *MediaDAO) Initialize(mclient *mongo.Client, config *config.AppConfig) error {

	err := mclient.Ping(context.TODO(), nil)

	if err != nil {
		log.Error().Err(err).Msg("Error connecting to mongodb")
	}

	log.Info().Msg("MediaDAO connected successfully to mongodb")

	m.DBClient = mclient
	m.Config = config

	return nil
}

//InsertMedia insert media
func (m *MediaDAO) InsertMedia(media *models.MediaModel) error {

	// Manage the create and update time
	media.CreatedAt = time.Now()
	media.UpdatedAt = time.Now()
	media.MediaID = models.GenUUID()
	media.Slug = slug.Make(media.Title)

	media.Tags = models.TagExtractor(media.Keywords)

	// Connect to our database
	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("media")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	insertResult, err := c.InsertOne(ctx, media)
	if err != nil {
		return err
	}

	// Convert to object ID
	media.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil
}

//UpdateMedia Update the title, keywords and description for media
func (m *MediaDAO) UpdateMedia(media models.MediaModel) error {

	// Connect to our database
	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("media")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

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

	result, err := c.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	fmt.Printf("updated %v record for media id  %s \n", result.ModifiedCount, media.MediaID)

	return nil
}

// GetMedia populate the media object based on ID
func (m *MediaDAO) GetMedia(id string) (models.MediaModel, error) {

	var media models.MediaModel

	// Connect to our database
	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("media")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err := c.FindOne(ctx, bson.M{"media_id": id}).Decode(&media)

	// Translate special variables
	media.ExposureProgramTranslated = media.GetExposureProgramTranslated()
	media.FStopTranslated = media.CalculateFSTOP()

	if err != nil {
		return media, err
	}

	return media, nil
}

// GetMediaBySlug populate the media object based on ID
func (m *MediaDAO) GetMediaBySlug(slug string) (models.MediaModel, error) {

	var media models.MediaModel

	// Connect to our database
	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("media")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err := c.FindOne(ctx, bson.M{"slug": slug}).Decode(&media)
	if err != nil {
		return media, err
	}

	// Translate special variables
	media.ExposureProgramTranslated = media.GetExposureProgramTranslated()
	media.FStopTranslated = media.CalculateFSTOP()

	return media, nil
}

//DeleteMedia delete the media object base on ID
func (m *MediaDAO) DeleteMedia(media models.MediaModel) error {

	// Connect to our database
	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("media")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	_, err := c.DeleteOne(ctx, bson.M{"media_id": media.MediaID})
	if err != nil {
		return err
	}

	return nil
}

//GetMediaListByCategory Obtains the list of media by category sorted by date
func (m *MediaDAO) GetMediaListByCategory(category string) ([]models.MediaModel, error) {

	var mediaModels []models.MediaModel

	// Connect to our database
	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("media")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	filter := bson.M{"category": category}

	options := options.Find()

	// Sort by `date` field descending
	options.SetSort(map[string]int{"created_at": -1})

	cur, err := c.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var media models.MediaModel
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&media)
		if err != nil {
			return nil, err
		}

		// Translate special variables
		media.ExposureProgramTranslated = media.GetExposureProgramTranslated()
		media.FStopTranslated = media.CalculateFSTOP()

		fmt.Println(media)

		mediaModels = append(mediaModels, media)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return mediaModels, nil
}

//AllMediaSortedByDate retrieve all posts sorted by creation date
func (m *MediaDAO) AllMediaSortedByDate() ([]models.MediaModel, error) {

	var mediaModels []models.MediaModel
	//config.DBUri = "mongodb://host.docker.internal:27017"

	// Connect to our database
	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("media")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	filter := bson.M{}

	options := options.Find()

	// Sort by `_id` field descending
	options.SetSort(map[string]int{"created_at": -1})

	cur, err := c.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(ctx)

	for cur.Next(ctx) {
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
func (m *MediaDAO) AllCategories() ([]string, error) {

	var categories []string

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("media")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	filter := bson.M{}

	//options := options.Distinct()

	cur, err := c.Distinct(ctx, "category", filter)
	if err != nil {
		return categories, err
	}

	for _, v := range cur {
		categories = append(categories, v.(string))
	}

	return categories, nil

}

//MediaSearch retrieve all posts sorted by creation date
func (m *MediaDAO) MediaSearch(mediasearch MediaSearch) ([]models.MediaModel, error) {

	//r := strings.NewReader(searchJSON)

	//var ms models.MediaSearchQuery
	//err := json.NewDecoder(r).Decode(&ms)
	//if err != nil {

	//	return nil, fmt.Errorf("Error converting search string to JSON with error %s", err)
	//}

	//filter :=

	//fmt.Println(ms)

	IsFilter := false

	if len(strings.TrimSpace(strings.TrimSuffix(mediasearch.SearchString, "\n"))) > 3 {
		IsFilter = true
	}

	var filter bson.M

	if IsFilter == true {
		filter = bson.M{
			"$text": bson.M{
				"$search": mediasearch.SearchString,
			},
		}
	} else {
		filter = bson.M{}
	}

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("media")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	options := options.Find()

	// Sort by `_id` field descending
	options.SetSort(map[string]int{"created_at": -1})

	//cur, err := c.Find(context.TODO(), filter, options)
	//if err != nil {
	//	return nil, err
	//}

	var cur *mongo.Cursor
	var err error

	log.Info().Msgf("Looking for [%s] and filter value is %t", mediasearch.SearchString, IsFilter)

	if IsFilter == true {
		cur, err = c.Find(context.TODO(), filter, options)
		if err != nil {
			return nil, err
		}
	} else {
		cur, err = c.Find(context.TODO(), bson.M{}, options)
		if err != nil {
			return nil, err
		}
	}

	var mediaModels []models.MediaModel

	defer cur.Close(ctx)

	for cur.Next(ctx) {
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

//SetS3Uploaded sets the status of the s3upload
func (m *MediaDAO) SetS3Uploaded(media *models.MediaModel) error {

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("media")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	filter := bson.M{
		"media_id": bson.M{
			"$eq": media.MediaID, // check if bool field has value of 'false'
		},
	}

	update := bson.M{
		"$set": bson.M{
			"s3_uploaded":  media.S3Uploaded,
			"s3_location":  media.S3Location,
			"s3_thumbnail": media.S3Thumbnail,
			"s3_largeview": media.S3LargeView,
			"s3_verylarge": media.S3VeryLarge,
		},
	}

	result, err := c.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	log.Info().Msgf("updated s3 status for %d record for media id  %s", result.ModifiedCount, media.MediaID)

	return nil
}
