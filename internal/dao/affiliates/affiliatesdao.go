package affiliatesdao

import (
	"context"
	"time"

	"goblog/internal/config"
	"goblog/internal/models"

	"github.com/gosimple/slug"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//AffiliatesDAO stores media data access object information
type AffiliatesDAO struct {
	DBClient *mongo.Client
	Config   *config.AppConfig
}

// Initialize creates the connection and populates the suppression struct
func (a *AffiliatesDAO) Initialize(mclient *mongo.Client, config *config.AppConfig) error {

	err := mclient.Ping(context.TODO(), nil)

	if err != nil {
		log.Error().Err(err).Msg("Error connecting to mongodb")
	}

	//log.Info().Msg("MediaDAO connected successfully to mongodb")

	a.DBClient = mclient
	a.Config = config

	return nil
}

// GetAffiliate get affiliate by ID
func (a *AffiliatesDAO) GetAffiliate(id string) (models.Affiliate, error) {

	var affiliate models.Affiliate

	// Connect to our database
	c := a.DBClient.Database(a.Config.MongoDatabase).Collection("affiliates")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err := c.FindOne(ctx, bson.M{"affiliate_id": id}).Decode(affiliate)
	if err != nil {
		return affiliate, err
	}

	return affiliate, nil
}

// InsertAffiliate create affiliate database entry
func (a *AffiliatesDAO) InsertAffiliate(affiliate *models.Affiliate) error {

	// Connect to our database
	c := a.DBClient.Database(a.Config.MongoDatabase).Collection("affiliates")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	// Manage the create and update time
	affiliate.CreatedAt = time.Now()
	affiliate.UpdatedAt = time.Now()
	affiliate.AffiliateID = models.GenUUID()
	affiliate.Slug = slug.Make(affiliate.AffiliateTitle)

	insertResult, err := c.InsertOne(ctx, affiliate)
	if err != nil {
		return err
	}

	// Convert to object ID
	affiliate.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil
}

// EditAffiliate edit affiliate database entry
func (a *AffiliatesDAO) EditAffiliate(affiliate *models.Affiliate) error {

	// Connect to our database
	c := a.DBClient.Database(a.Config.MongoDatabase).Collection("affiliates")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	filter := bson.M{
		"affiliate_id": bson.M{
			"$eq": affiliate.AffiliateID, // check if bool field has value of 'false'
		},
	}

	affiliate.UpdatedAt = time.Now()
	affiliate.Slug = slug.Make(affiliate.AffiliateTitle)

	update := bson.M{
		"$set": bson.M{
			"affiliate_link":  affiliate.AffiliateLink,
			"affiliate_title": affiliate.AffiliateTitle,
			"description":     affiliate.Description,
			"Updated_at":      affiliate.UpdatedAt,
			"category":        affiliate.Category,
			"slug":            affiliate.Slug,
		},
	}

	result, err := c.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	log.Info().Msgf("updated %v record for affiliate id  %s \n", result.ModifiedCount, affiliate.AffiliateID)

	return nil
}

// GetAllAffiliateOrderByDate edit affiliate database entry
func (a *AffiliatesDAO) GetAllAffiliateOrderByDate() ([]models.Affiliate, error) {

	// Connect to our database
	c := a.DBClient.Database(a.Config.MongoDatabase).Collection("affiliates")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var affiliatesList []models.Affiliate
	//config.DBUri = "mongodb://host.docker.internal:27017"

	filter := bson.M{}

	options := options.Find()

	// Sort by `_id` field descending
	options.SetSort(map[string]int{"created_at": -1})

	cur, err := c.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var affiliate models.Affiliate
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&a)
		if err != nil {
			return nil, err
		}

		affiliatesList = append(affiliatesList, affiliate)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return affiliatesList, nil
}

// DeleteAffiliate delete the affiliate
func (a *AffiliatesDAO) DeleteAffiliate(affiliate *models.Affiliate) error {

	// Connect to our database
	c := a.DBClient.Database(a.Config.MongoDatabase).Collection("affiliates")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	_, err := c.DeleteOne(ctx, bson.M{"affiliate_id": affiliate.AffiliateID})

	if err != nil {
		return err
	}

	return nil
}
