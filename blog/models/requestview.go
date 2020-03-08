package models

import (
	"blog/blog/config"
	"blog/blog/db"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

// RequestView represents a request view
type RequestView struct {
	ID                primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	RequestViewID     string             `json:"requestviewid" bson:"requestviewid"`
	FunctionalBrowser string             `json:"functionalbrowser" bson:"functionalbrowser,omitempty"`
	SessionID         string             `json:"sessionid" bson:"sessionid"`
	OSVersion         string             `json:"osversion" bson:"osversion,omitempty"`
	OS                string             `json:"os" bson:"os,omitempty"`
	UserAgent         string             `json:"useragent" bson:"useragent,omitempty"`
	NavAppVersion     string             `json:"navappversion" bson:"navappversion,omitempty"`
	NavPlatform       string             `json:"navplatform" bson:"navplatform,omitempty"`
	NavBrowser        string             `json:"navbrowser" bson:"navbrowser,omitempty"`
	BrowserVersion    string             `json:"browserversion" bson:"browserversion,omitempty"`
	PTag              string             `json:"ptag" bson:"ptag"`
	CreatedAt         time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at" bson:"updated_at"`
	HeaderUserAgent   string             `json:"header_user_agent" bson:"header_user_agent"`
	IPAddress         string             `json:"ipaddress" bson:"ipaddress"`
	RequestURL        string             `json:"request_url" bson:"request_url"`
	City              string             `json:"city" bson:"city"`
	Country           string             `json:"country" bson:"country"`
	RawRequest        string             `json:"raw_request" bson:"raw_request"`
}

//CreateRequestView create a new requestview
func (r *RequestView) CreateRequestView() error {
	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	// Manage the create and update time
	r.CreatedAt = time.Now()
	r.RequestViewID = GenUUID()

	c := db.Client.Database(getRequestViewDB()).Collection("requestview")

	insertResult, err := c.InsertOne(context.TODO(), r)
	if err != nil {
		return err
	}

	// Convert to object ID
	r.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil

}

// UpdateRequestView update requestView Record by PTag
func (r *RequestView) UpdateRequestView() error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	// Manage update time
	r.UpdatedAt = time.Now()

	c := db.Client.Database(getRequestViewDB()).Collection("requestview")

	filter := bson.M{
		"ptag": bson.M{
			"$eq": r.PTag, // check if bool field has value of 'false'
		},
	}

	update := bson.M{
		"$set": bson.M{
			"functionalbrowser": r.FunctionalBrowser,
			"osversion":         r.OSVersion,
			"os":                r.OS,
			"useragent":         r.UserAgent,
			"navappversion":     r.NavAppVersion,
			"navplatform":       r.NavPlatform,
			"navbrowser":        r.NavBrowser,
			"browserversion":    r.BrowserVersion,
			"updated_at":        r.UpdatedAt,
		},
	}

	updateResult, err := c.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))

	if err != nil {
		return err
	}

	if updateResult != nil && updateResult.MatchedCount == 0 {
		fmt.Printf("could not find any requestview records to update for ptag valute %s", r.PTag)
	}

	return nil
}

// GetRequestViewByPTAG get a requestview by pageid
func (r *RequestView) GetRequestViewByPTAG(id string) error {
	var db db.Session

	err := db.NewSession()

	c := db.Client.Database(getRequestViewDB()).Collection("requestview")

	err = c.FindOne(context.TODO(), bson.M{"ptag": id}).Decode(r)
	if err != nil {
		return err
	}
	defer db.Close()

	return nil
}

// GetRequestViewsBySessionID get a list of requestviews by sessionid
func (r *RequestView) GetRequestViewsBySessionID(id string) ([]RequestView, error) {

	var db db.Session

	var rv []RequestView

	err := db.NewSession()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	//filter := bson.M{}
	filter := bson.M{
		"sessionid": bson.M{
			"$eq": id,
		},
	}

	options := options.Find()

	// Sort by `_id` field descending
	options.SetSort(map[string]int{"created_at": -1})

	cur, err := db.Client.Database(getRequestViewDB()).Collection("requestview").Find(context.TODO(), filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var r RequestView
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&r)
		if err != nil {
			return nil, err
		}

		rv = append(rv, r)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return rv, nil
}

func getRequestViewDB() string {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("error getting configuration: %s", err)
		return ""
	}
	return cfg.MongoDatabase
}
