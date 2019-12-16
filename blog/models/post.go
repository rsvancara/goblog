package models

import (
	"context"
	"fmt"
	"time"

	"bf.go/blog/db"
	"github.com/segmentio/ksuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
)

// PostModel post
type PostModel struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	PostID    string             `json:"postid" bson:"postid,omitempty"`
	Post      string             `json:"post" bson:"post,omitempty"`
	Title     string             `json:"title" bson:"title,omitempty"`
	Tags      []string           `json:"tags" bson:"tags,omitempty"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// GetPost populate the post object based on ID
func (p *PostModel) GetPost(id string) error {

	var config db.Config
	var db db.Session

	config.DBUri = ""
	err := db.NewSession(&config)

	c := db.Client.Database("blog").Collection("posts")

	err = c.FindOne(context.TODO(), bson.M{"ID": id}).Decode(p)
	if err != nil {
		return err
	}
	defer db.Close()

	return nil
}

//InsertPost insert post
func (p *PostModel) InsertPost() error {

	var config db.Config
	config.DBUri = "mongodb://host.docker.internal:27017"

	var db db.Session

	err := db.NewSession(&config)
	if err != nil {
		return err
	}

	defer db.Close()

	// Manage the create and update time
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	p.PostID = genUUID()

	c := db.Client.Database("blog").Collection("posts")

	insertResult, err := c.InsertOne(context.TODO(), p)
	if err != nil {
		return err
	}

	// Convert to object ID
	p.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil
}

//DeletePost delete the post object base on ID
func (p *PostModel) DeletePost(id string) error {

	var config db.Config
	var db db.Session

	config.DBUri = ""
	err := db.NewSession(&config)
	if err != nil {
		return err
	}
	defer db.Close()

	return nil
}

//FindPostsByKeyWords find posts that match keywords. Variadic function that can take multiple values
func FindPostsByKeyWords(keywords ...string) ([]PostModel, error) {

	var config db.Config
	var db db.Session

	config.DBUri = "mongodb://host.docker.internal:27017"
	err := db.NewSession(&config)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return nil, nil
}

//AllPostsSortedByDate retrieve all posts sorted by creation date
func AllPostsSortedByDate() ([]PostModel, error) {

	var config db.Config
	var db db.Session

	var postModels []PostModel
	config.DBUri = "mongodb://host.docker.internal:27017"

	err := db.NewSession(&config)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	filter := bson.M{}

	cur, err := db.Client.Database("blog").Collection("posts").Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var p PostModel
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&p)
		if err != nil {
			return nil, err
		}

		postModels = append(postModels, p)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	fmt.Println(postModels)

	return postModels, nil
}

func genUUID() string {
	id := ksuid.New()
	return id.String()
}
