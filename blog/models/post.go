package models

import (
	"context"
	"fmt"
	"time"

	"blog/blog/db"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

// PostModel post
type PostModel struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	PostID     string             `json:"postid" bson:"postid,omitempty"`
	Slug       string             `json:"slug" bson:"slug,omitempty"`
	Post       string             `json:"post" bson:"post,omitempty"`
	PostTeaser string             `json:"post_teaser" bson:"post_teaser,omitempty"`
	Title      string             `json:"title" bson:"title,omitempty"`
	Keywords   string             `json:"keywords" bson:"keywords,omitempty"`
	Tags       []Tag              `json:"tags" bson:"tags,omitempty"`
	Status     string             `json:"status" bson:"status,omitempty"`
	Featured   string             `json:"featured" bson:"featured,omitempty"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

// GetPost populate the post object based on ID
func (p *PostModel) GetPost(id string) error {

	var db db.Session

	err := db.NewSession()

	c := db.Client.Database("blog").Collection("posts")

	err = c.FindOne(context.TODO(), bson.M{"postid": id}).Decode(p)
	if err != nil {
		return err
	}
	defer db.Close()

	return nil
}

// GetPostBySlug populate the post object based on slug
func (p *PostModel) GetPostBySlug(id string) error {

	var db db.Session

	err := db.NewSession()

	c := db.Client.Database("blog").Collection("posts")

	err = c.FindOne(context.TODO(), bson.M{"slug": id}).Decode(p)
	if err != nil {
		return err
	}
	defer db.Close()

	return nil
}

//InsertPost insert post
func (p *PostModel) InsertPost() error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	// Manage the create and update time
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	p.PostID = genUUID()
	p.Slug = slug.Make(p.Title)
	p.Tags = TagExtractor(p.Keywords)

	c := db.Client.Database("blog").Collection("posts")

	insertResult, err := c.InsertOne(context.TODO(), p)
	if err != nil {
		return err
	}

	// Convert to object ID
	p.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil
}

//UpdatePost update existing post
func (p *PostModel) UpdatePost() error {

	//var config db.Config
	//config.DBUri = "mongodb://host.docker.internal:27017"

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	// Manage update time
	p.UpdatedAt = time.Now()

	c := db.Client.Database("blog").Collection("posts")

	filter := bson.M{
		"postid": bson.M{
			"$eq": p.PostID, // check if bool field has value of 'false'
		},
	}

	p.Slug = slug.Make(p.Title)
	p.Tags = TagExtractor(p.Keywords)

	update := bson.M{
		"$set": bson.M{
			"title":       p.Title,
			"post":        p.Post,
			"updated_at":  p.UpdatedAt,
			"status":      p.Status,
			"post_teaser": p.PostTeaser,
			"featured":    p.Featured,
			"slug":        p.Slug,
			"keywords":    p.Keywords,
			"tags":        p.Tags,
		},
	}

	updateResult, err := c.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))

	if err != nil {
		return err
	}

	fmt.Printf("Found and updated %d post record(s)", updateResult.MatchedCount)

	return nil
}

//DeletePost delete the post object base on ID
func (p *PostModel) DeletePost() error {

	//var config db.Config
	var db db.Session

	//config.DBUri = "mongodb://host.docker.internal:27017"
	err := db.NewSession()
	if err != nil {
		return err
	}
	defer db.Close()

	c := db.Client.Database("blog").Collection("posts")
	_, err = c.DeleteOne(context.TODO(), bson.M{"postid": p.PostID})

	if err != nil {
		return err
	}

	return nil
}

//FindPostsByKeyWords find posts that match keywords. Variadic function that can take multiple values
func FindPostsByKeyWords(keywords ...string) ([]PostModel, error) {

	//var config db.Config
	var db db.Session

	//config.DBUri = "mongodb://host.docker.internal:27017"
	err := db.NewSession()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return nil, nil
}

//AllPostsSortedByDate retrieve all posts sorted by creation date
func AllPostsSortedByDate() ([]PostModel, error) {

	//var config db.Config
	var db db.Session

	var postModels []PostModel
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

	cur, err := db.Client.Database("blog").Collection("posts").Find(context.TODO(), filter, options)
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

	return postModels, nil
}
