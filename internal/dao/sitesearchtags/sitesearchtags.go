//Package mediatagsdao media data access object
package sitesearchtags

import (
	"context"
	"fmt"
	"time"

	"goblog/internal/models"

	"goblog/internal/config"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//MediaTagsDAO stores media data access object information
type SiteSearchTagsDAO struct {
	DBClient *mongo.Client
	Config   *config.AppConfig
}

// Initialize creates the connection and populates the suppression struct
func (m *SiteSearchTagsDAO) Initialize(mclient *mongo.Client, config *config.AppConfig) error {

	err := mclient.Ping(context.TODO(), nil)

	if err != nil {
		log.Error().Err(err).Msg("Error connecting to mongodb")
	}

	//log.Info().Msg("MediaTagsDAO connected successfully to mongodb")

	m.DBClient = mclient
	m.Config = config

	return nil
}

//Insert SiteSearchTags insert sitesearchtags
func (m *SiteSearchTagsDAO) InsertSiteSearchTags(sitesearchtag *models.SiteSearchTagsModel) error {

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("sitesearchtags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	insertResult, err := c.InsertOne(ctx, sitesearchtag)
	if err != nil {
		return err
	}

	// Convert to object ID
	sitesearchtag.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil
}

//UpdateSiteSearchTags Update the title, keywords and description for site searches
func (m *SiteSearchTagsDAO) UpdateSiteSearchTags(sitesearchtag *models.SiteSearchTagsModel) error {

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("sitesearchtags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	filter := bson.M{
		"tags_id": bson.M{
			"$eq": sitesearchtag.TagsID, // check if bool field has value of 'false'
		},
	}

	//m.Documents = TagExtractor(m.Keywords)
	//m.Slug = slug.Make(m.Title)

	update := bson.M{
		"$set": bson.M{
			"name":      sitesearchtag.Name,
			"documents": sitesearchtag.Documents,
		},
	}

	_, err := c.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

// GetSiteSearchTag populate the media object based on ID
func (s *SiteSearchTagsDAO) GetSiteSearchTag(id string) error {

	var sitesearchtag models.SiteSearchTagsModel

	c := s.DBClient.Database(s.Config.MongoDatabase).Collection("sitesearchtags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err := c.FindOne(ctx, bson.M{"tags_id": id}).Decode(&sitesearchtag)
	if err != nil {
		return err
	}

	return nil
}

// Exists Check to see if record exists and if it does return it
func (s *SiteSearchTagsDAO) Exists(name string) (int64, error) {

	var count int64

	c := s.DBClient.Database(s.Config.MongoDatabase).Collection("sitesearchtags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	count, err := c.CountDocuments(ctx, bson.M{"name": name})
	if err != nil {
		return 0, err
	}

	return count, nil

}

// GetSiteSearchTagsCount get the total number of sitesearch tags
func (s *SiteSearchTagsDAO) GetSiteSearchTagsCount() (int64, error) {

	c := s.DBClient.Database(s.Config.MongoDatabase).Collection("sitesearchtags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	count, err := c.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}
	return count, nil
}

// GetSiteSearchTagByName populate the sitesearch object based on ID
func (m *SiteSearchTagsDAO) GetSiteSearchTagByName(name string) (models.SiteSearchTagsModel, error) {

	var sitesearchmodel models.SiteSearchTagsModel

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("sitesearchtags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	err := c.FindOne(ctx, bson.M{"name": name}).Decode(&sitesearchmodel)
	if err != nil {
		return sitesearchmodel, err
	}

	return sitesearchmodel, err
}

// SearchMediaTagsByName Text search for SiteSearchTags
func (m *SiteSearchTagsDAO) SearchMediaTagsByName(name string) ([]models.SiteSearchTagsModel, error) {

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("sitesearchtags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var siteSearchTagsModels []models.SiteSearchTagsModel

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
		var sitesearchtag models.SiteSearchTagsModel
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&sitesearchtag)
		if err != nil {
			return nil, err
		}

		siteSearchTagsModels = append(siteSearchTagsModels, sitesearchtag)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return siteSearchTagsModels, nil

}

//DeleteAllTags Delete all tags, used when index needs to be rebuilt
func (m *SiteSearchTagsDAO) DeleteAllTags() error {

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("sitesearchtags")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	_, err := c.DeleteMany(ctx, bson.M{})
	if err != nil {
		return err
	}

	return nil
}

//AllMediaTags retrieve all media tags
func (m *SiteSearchTagsDAO) AllSiteSearchTags() ([]models.SiteSearchTagsModel, error) {

	var siteSearchTagsModels []models.SiteSearchTagsModel

	c := m.DBClient.Database(m.Config.MongoDatabase).Collection("sitesearchtags")

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
		var sitesearchtag models.SiteSearchTagsModel
		// To decode into a struct, use cursor.Decode()
		err := cur.Decode(&m)
		if err != nil {
			return nil, err
		}

		siteSearchTagsModels = append(siteSearchTagsModels, sitesearchtag)

	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return siteSearchTagsModels, nil
}

//AddTagsSearchIndex update the search index with new as items are added, called when anything is added or updated
func (s *SiteSearchTagsDAO) AddTagsSearchIndex(docID string, doctype string, tags []string) error {

	for _, v := range tags {

		count, err := s.Exists(v)
		if err != nil {
			return fmt.Errorf("error attempting to get record count for keyword %s with error %s", v, err)
		}

		log.Info().Msgf("Found %d site search tag records for keyworkd %s", count, v)

		// Determine if the document exists already
		if count == 0 {
			log.Info().Msgf("]Tag does not exist for %s in the database", v)
			var newSTM models.SiteSearchTagsModel
			newSTM.Name = v
			newSTM.TagsID = models.GenUUID()
			var docs []string
			docs = append(docs, docID)
			newSTM.Documents = docs
			log.Info().Msgf("Inserting new tag %s into database", v)
			err = s.InsertSiteSearchTags(&newSTM)
			if err != nil {
				return fmt.Errorf("error inserting new site search tag for keyword %s with error %s", v, err)
			}
			log.Info().Msgf("Tag %s inserted successfully", v)
			// If not, then we add to existing documents
		} else {

			stm, err := s.GetSiteSearchTagByName(v)
			if err != nil {
				return fmt.Errorf("error getting current instance of searchtag for keyword %s with error %s", v, err)
			}
			log.Info().Msgf("Found existing searchtagid record for %s", stm.Name)
			//fmt.Println(mtm.Documents)

			// Get the list of documents
			docs := stm.Documents

			// For the list of documents, find the document ID we are looking for
			// If not found, then we update the document list with the document ID
			found := false
			for _, d := range docs {
				if d == v {
					found = true
				}
			}

			if found {
				log.Info().Msgf("Updating tag, %s with document id %s", v, docID)
				docs = append(docs, docID)
				stm.Documents = docs
				//fmt.Println(mtm)
				err = s.UpdateSiteSearchTags(&stm)
				if err != nil {
					return fmt.Errorf("error updating searchtag for keyword %s with error %s", v, err)
				}
			}
		}
	}
	return nil
}
