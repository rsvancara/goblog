//Package postsdao media data access object
package postsdao

import (
	"context"
	"time"

	"github.com/gosimple/slug"
	"goblog/internal/models"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"goblog/internal/config"
)

//PostsDAO stores media data access object information
type PostsDAO struct {
	DBClient *mongo.Client
	Config   *config.AppConfig
}

// Initialize creates the connection and populates the suppression struct
func (p *PostsDAO) Initialize(mclient *mongo.Client, config *config.AppConfig) error {

	err := mclient.Ping(context.TODO(), nil)

	if err != nil {
		log.Error().Err(err).Msg("Error connecting to mongodb")
	}

	log.Info().Msg("PostsDAO connected successfully to mongodb")

	p.DBClient = mclient
	p.Config = config

	return nil
}

// GetPost populate the post object based on ID
func (p *PostsDAO) GetPost(id string) (models.PostModel, error) {

	var post models.PostModel

	c := p.DBClient.Database(p.Config.MongoDatabase).Collection("posts")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err := c.FindOne(ctx, bson.M{"postid": id}).Decode(&post)
	if err != nil {
		return post, err
	}

	return post, nil
}

// GetPostBySlug populate the post object based on slug
func (p *PostsDAO) GetPostBySlug(id string) (models.PostModel, error) {

	var post models.PostModel

	c := p.DBClient.Database(p.Config.MongoDatabase).Collection("posts")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err := c.FindOne(ctx, bson.M{"slug": id}).Decode(&post)
	if err != nil {
		return post, err
	}

	return post, nil
}

//InsertPost insert post
func (p *PostsDAO) InsertPost(post *models.PostModel) error {

	c := p.DBClient.Database(p.Config.MongoDatabase).Collection("posts")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	// Manage the create and update time
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()
	post.PostID = models.GenUUID()
	post.Slug = slug.Make(post.Title)
	post.Tags = models.TagExtractor(post.Keywords)

	insertResult, err := c.InsertOne(ctx, post)
	if err != nil {
		return err
	}

	// Convert to object ID
	post.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil
}

//UpdatePost update existing post
func (p *PostsDAO) UpdatePost(post *models.PostModel) error {

	c := p.DBClient.Database(p.Config.MongoDatabase).Collection("posts")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	// Manage update time
	post.UpdatedAt = time.Now()

	filter := bson.M{
		"postid": bson.M{
			"$eq": post.PostID, // check if bool field has value of 'false'
		},
	}

	post.Slug = slug.Make(post.Title)
	post.Tags = models.TagExtractor(post.Keywords)

	update := bson.M{
		"$set": bson.M{
			"title":             post.Title,
			"post":              post.Post,
			"updated_at":        post.UpdatedAt,
			"status":            post.Status,
			"post_teaser":       post.PostTeaser,
			"featured":          post.Featured,
			"slug":              post.Slug,
			"keywords":          post.Keywords,
			"tags":              post.Tags,
			"teaser_image":      post.TeaserImage,
			"teaser_image_slug": post.TeaserImageSlug,
		},
	}

	updateResult, err := c.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))

	if err != nil {
		return err
	}

	log.Info().Msgf("Found and updated %d post record(s)", updateResult.MatchedCount)

	return nil
}

//DeletePost delete the post object base on ID
func (p *PostsDAO) DeletePost(post *models.PostModel) error {

	c := p.DBClient.Database(p.Config.MongoDatabase).Collection("posts")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	_, err := c.DeleteOne(ctx, bson.M{"postid": post.PostID})
	if err != nil {
		return err
	}

	return nil
}

//FindPostsByKeyWords find posts that match keywords. Variadic function that can take multiple values
func (p *PostsDAO) FindPostsByKeyWords(keywords ...string) ([]models.PostModel, error) {

	//c := p.DBClient.Database(p.Config.MongoDatabase).Collection("posts")

	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	//defer cancel()

	return nil, nil
}

//AllPostsSortedByDate retrieve all posts sorted by creation date
func (p *PostsDAO) AllPostsSortedByDate() ([]models.PostModel, error) {

	c := p.DBClient.Database(p.Config.MongoDatabase).Collection("posts")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var postModels []models.PostModel
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
		var post models.PostModel
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&post)
		if err != nil {
			return nil, err
		}

		postModels = append(postModels, post)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return postModels, nil
}
