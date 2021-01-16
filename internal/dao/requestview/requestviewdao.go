//requestviewdao provides data access object for requestviews
package requestviewdao

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rsvancara/goblog/internal/config"
	"github.com/rsvancara/goblog/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

//MediaDAO stores media data access object information
type RequestViewDAO struct {
	DBClient *mongo.Client
	Config   *config.AppConfig
}

// Initialize creates the connection and populates the suppression struct
func (r *RequestViewDAO) Initialize(mclient *mongo.Client, config *config.AppConfig) error {

	err := mclient.Ping(context.TODO(), nil)

	if err != nil {
		log.Error().Err(err).Msg("Error connecting to mongodb")
	}

	log.Info().Msg("MediaDAO connected successfully to mongodb")

	r.DBClient = mclient
	r.Config = config

	return nil
}

//CreateRequestView create a new requestview
func (r *RequestViewDAO) CreateRequestView(rv *models.RequestView) error {

	// Manage the create and update time
	rv.CreatedAt = time.Now()
	rv.RequestViewID = models.GenUUID()

	c := r.DBClient.Database(r.Config.MongoDatabase).Collection("requestview")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	insertResult, err := c.InsertOne(ctx, rv)
	if err != nil {
		return err
	}

	// Convert to object ID
	rv.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil

}

// UpdateRequestView update requestView Record by PTag
func (r *RequestViewDAO) UpdateRequestView(rv *models.RequestView) error {

	// Manage update time
	rv.UpdatedAt = time.Now()

	c := r.DBClient.Database(r.Config.MongoDatabase).Collection("requestview")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	filter := bson.M{
		"ptag": bson.M{
			"$eq": rv.PTag, // check if bool field has value of 'false'
		},
	}

	update := bson.M{
		"$set": bson.M{
			"functionalbrowser": rv.FunctionalBrowser,
			"osversion":         rv.OSVersion,
			"os":                rv.OS,
			"useragent":         rv.UserAgent,
			"navappversion":     rv.NavAppVersion,
			"navplatform":       rv.NavPlatform,
			"navbrowser":        rv.NavBrowser,
			"browserversion":    rv.BrowserVersion,
			"updated_at":        rv.UpdatedAt,
		},
	}

	updateResult, err := c.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))

	if err != nil {
		return err
	}

	if updateResult != nil && updateResult.MatchedCount == 0 {
		fmt.Printf("could not find any requestview records to update for ptag valute %s", rv.PTag)
	}

	return nil
}

// GetRequestViewByPTAG get a requestview by pageid
func (r *RequestViewDAO) GetRequestViewByPTAG(id string) (models.RequestView, error) {

	var rv models.RequestView

	c := r.DBClient.Database(r.Config.MongoDatabase).Collection("requestview")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err := c.FindOne(ctx, bson.M{"ptag": id}).Decode(r)
	if err != nil {
		return rv, err
	}

	return rv, nil
}

// GetRequestViewsBySessionID get a list of requestviews by sessionid
func (r *RequestViewDAO) GetRequestViewsBySessionID(id string) ([]models.RequestView, error) {

	var rvlist []models.RequestView

	c := r.DBClient.Database(r.Config.MongoDatabase).Collection("requestview")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	//filter := bson.M{}
	filter := bson.M{
		"sessionid": bson.M{
			"$eq": id,
		},
	}

	options := options.Find()

	// Sort by `_id` field descending
	options.SetSort(map[string]int{"created_at": -1})

	cur, err := c.Find(ctx, filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var rv models.RequestView
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&rv)
		if err != nil {
			return nil, err
		}

		rvlist = append(rvlist, rv)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return rvlist, nil
}
