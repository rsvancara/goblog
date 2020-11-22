package models

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rsvancara/goblog/internal/config"
	"github.com/rsvancara/goblog/internal/db"

	"github.com/dsoprea/go-exif"
	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
)

// MediaModel post
type MediaModel struct {
	ID                        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	MediaID                   string             `json:"media_id" bson:"media_id,omitempty"`
	Slug                      string             `json:"slug" bson:"slug,omitempty"`
	Keywords                  string             `json:"keywords" bson:"keywords,omitempty"`
	Category                  string             `json:"category" bson:"category,omitempty"`
	Title                     string             `json:"title" bson:"title,omitempty"`
	FileName                  string             `json:"file_name" bson:"file_name,omitempty"`
	S3Location                string             `json:"s3_location" bson:"s3_location,omitempty"`
	S3Thumbnail               string             `json:"s3_thumbnail" bson:"s3_thumbnail,omitempty"`
	S3LargeView               string             `json:"s3_largeview" bson:"s3_largeview,omitempty"`
	S3VeryLarge               string             `json:"s3_verylarge" bson:"s3_verylarge,omitempty"`
	S3Uploaded                string             `json:"s3_uploaded" bson:"s3_uploaded,omitempty"`
	Description               string             `json:"description" bson:"description,omitempty"`
	Checksum                  string             `json:"checksum" bson:"checksum,omitempty"`
	CreatedAt                 time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt                 time.Time          `json:"updated_at" bson:"updated_at"`
	Make                      string             `json:"make" bson:"make,omitempty"`                           //SONY
	Model                     string             `json:"model" bson:"model,omitempty"`                         //ILCE-7RM3
	Software                  string             `json:"software" bson:"software,omitempty"`                   //ILCE-7RM3 v2.10
	DateTime                  time.Time          `json:"datetime_taken" bson:"datetime_taken"`                 //2019:12:23 18:46:27
	Artist                    string             `json:"artist" bson:"artist,omitempty"`                       //randall svancara
	Copyright                 string             `json:"copyright" bson:"copyright,omitempty"`                 //vi
	ExposureTime              string             `json:"exposuretime" bson:"exposuretime,omitempty"`           //1/30
	FNumber                   string             `json:"fnumber" bson:"fnumber,omitempty"`                     //14/5
	ISOSpeedRatings           string             `json:"iso_speed_rating" bson:"iso_speed_rating,omitempty"`   //1600
	LightSource               string             `json:"light_source" bson:"light_source,omitempty"`           //0
	FocalLength               string             `json:"focal_length" bson:"focal_length,omitempty"`           //23/1
	PixelXDimension           string             `json:"pixel_x_dimension" bson:"pixel_x_dimension,omitempty"` //7968
	PixelYDimension           string             `json:"pixel_y_dimension" bson:"pixel_y_dimension,omitempty"` //5320
	FocalLengthIn35mmFilm     string             `json:"focal_length35" bson:"focal_length35,omitempty"`       //23
	LensModel                 string             `json:"lens_model" bson:"lens_model,omitempty"`               //FE 16-35mm F2.8 GM
	ExposureProgram           string             `json:"exposure_program" bson:"exposure_program,omitempty"`
	ExposureProgramTranslated string             `json:"exposure_program_translated" bson:"exposure_program_translated,omitempty"`
	FStop                     string             `json:"fstop" bson:"fstop,omitempty"`
	FStopTranslated           string             `json:"fstop_translated" bson:"fstop_translated,omitempty"`
	Tags                      []Tag              `json:"tags" bson:"tags"`
	Location                  string             `json:"location" bson:"location,omitempty"`
}

// Tag stores tag objects
type Tag struct {
	Keyword string `json:"tag" bson:"tag,omitempty"`
}

//MediaSearchQuery Media Search Object
type MediaSearchQuery struct {
	Tags     string `json:"tags"`
	Title    string `json:"title"`
	Category string `json:"category"`
}

//InsertMedia insert media
func (m *MediaModel) InsertMedia() error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	// Manage the create and update time
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	m.MediaID = GenUUID()
	m.Slug = slug.Make(m.Title)

	m.Tags = TagExtractor(m.Keywords)

	c := db.Client.Database(getMediaDB()).Collection("media")

	insertResult, err := c.InsertOne(context.TODO(), m)
	if err != nil {
		return err
	}

	// Convert to object ID
	m.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil
}

// TagExtractor Extract tags from keywords
func TagExtractor(keywords string) []Tag {

	var tagArray []Tag
	tokens := strings.Split(keywords, ",")

	for t := range tokens {

		var tg Tag
		tg.Keyword = strings.ToLower(strings.Trim(tokens[t], " "))
		tagArray = append(tagArray, tg)
	}

	return tagArray
}

//UpdateMedia Update the title, keywords and description for media
func (m *MediaModel) UpdateMedia() error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	c := db.Client.Database(getMediaDB()).Collection("media")

	filter := bson.M{
		"media_id": bson.M{
			"$eq": m.MediaID, // check if bool field has value of 'false'
		},
	}

	m.Tags = TagExtractor(m.Keywords)
	m.Slug = slug.Make(m.Title)

	update := bson.M{
		"$set": bson.M{
			"keywords":    m.Keywords,
			"title":       m.Title,
			"description": m.Description,
			"tags":        m.Tags,
			"category":    m.Category,
			"slug":        m.Slug,
			"location":    m.Location,
		},
	}

	result, err := c.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	fmt.Printf("updated %v record for media id  %s \n", result.ModifiedCount, m.MediaID)

	return nil
}

//SetS3Uploaded sets the status of the s3upload
func (m *MediaModel) SetS3Uploaded() error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	c := db.Client.Database(getMediaDB()).Collection("media")

	filter := bson.M{
		"media_id": bson.M{
			"$eq": m.MediaID, // check if bool field has value of 'false'
		},
	}

	update := bson.M{
		"$set": bson.M{
			"s3_uploaded":  m.S3Uploaded,
			"s3_location":  m.S3Location,
			"s3_thumbnail": m.S3Thumbnail,
			"s3_largeview": m.S3LargeView,
			"s3_verylarge": m.S3VeryLarge,
		},
	}

	result, err := c.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	fmt.Printf("updated s3 status for %v record for media id  %s \n", result.ModifiedCount, m.MediaID)

	return nil
}

// GetMedia populate the media object based on ID
func (m *MediaModel) GetMedia(id string) error {

	var db db.Session

	//config.DBUri = "mongodb://host.docker.internal:27017"
	err := db.NewSession()

	c := db.Client.Database(getMediaDB()).Collection("media")

	err = c.FindOne(context.TODO(), bson.M{"media_id": id}).Decode(m)

	// Translate special variables
	m.ExposureProgramTranslated = m.GetExposureProgramTranslated()
	m.FStopTranslated = m.CalculateFSTOP()

	if err != nil {
		return err
	}
	defer db.Close()

	return nil
}

// GetMediaBySlug populate the media object based on ID
func (m *MediaModel) GetMediaBySlug(slug string) error {

	var db db.Session

	//config.DBUri = "mongodb://host.docker.internal:27017"
	err := db.NewSession()

	c := db.Client.Database(getMediaDB()).Collection("media")

	err = c.FindOne(context.TODO(), bson.M{"slug": slug}).Decode(m)
	// Translate special variables
	m.ExposureProgramTranslated = m.GetExposureProgramTranslated()
	m.FStopTranslated = m.CalculateFSTOP()

	if err != nil {
		return err
	}
	defer db.Close()

	return nil
}

// GetExposureProgramTranslated translates numeric value to Exposure Mode
func (m *MediaModel) GetExposureProgramTranslated() string {
	exposureProgramMap := map[string]string{
		"0": "Not Defined",
		"1": "Manual",
		"2": "Program AE",
		"3": "Aperture-priority AE",
		"4": "Shutter speed priority AE",
		"5": "Creative (Slow speed)",
		"6": "Action (High speed)",
		"7": "Portrait",
		"8": "Landscape",
		"9": "Bulb",
	}

	if val, ok := exposureProgramMap[m.ExposureProgram]; ok {
		return val
	}

	return "Unknown"
}

// CalculateFSTOP Calculates the FSTOP Value for display purposes
func (m *MediaModel) CalculateFSTOP() string {

	vals := strings.Split(m.FNumber, "/")

	if len(vals) == 2 {

		num, err := strconv.ParseFloat(vals[0], 64)
		if err != nil {
			return "Unknown"
		}

		den, err := strconv.ParseFloat(vals[1], 64)
		if err != nil {
			return "Unknown"
		}

		fstop := fmt.Sprintf("%.1f", num/den)

		return fstop
	}

	return "Unknown"
}

// ExifExtractor Extract EXIF Information from image
func (m *MediaModel) ExifExtractor(f *os.File) error {

	m.Make = "Unknown"
	m.Model = "Unknown"
	m.DateTime = time.Now()
	m.Artist = "Unknown"
	m.LensModel = "Uknown"
	m.FocalLength = "Unknown"
	m.LightSource = "Unknown"
	m.ExposureProgram = "Uknown"

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	exifData, err := exif.SearchAndExtractExif(data)
	if err != nil {
		if err == exif.ErrNoExif {
			return err
		}
		return err
	}

	// parse exif information
	im := exif.NewIfdMappingWithStandard()
	ti := exif.NewTagIndex()

	visitor := func(fqIfdPath string, ifdIndex int, tagId uint16, tagType exif.TagType, valueContext exif.ValueContext) (err error) {
		ifdPath, err := im.StripPathPhraseIndices(fqIfdPath)
		if err != nil {
			fmt.Printf("Error stripping phrase indices: %s\n", err)
			return nil
		}

		it, err := ti.Get(ifdPath, tagId)
		if err != nil {
			fmt.Printf("Warning: getting information about non-IFD tags: %s\n", err)
			return nil
		}

		valueString := ""

		if tagType.Type() == exif.TypeUndefined {
			value, err := exif.UndefinedValue(ifdPath, tagId, valueContext, tagType.ByteOrder())
			if err != nil {
				valueString = "!UNDEFINED!"
			}
			valueString = fmt.Sprintf("%v", value)
		} else {
			valueString, err = tagType.ResolveAsString(valueContext, true)
			if err != nil {
				fmt.Printf("error resolving tag: %s\n", err)
			}
		}

		// Obtain the various components and add exif information
		if it.Name == "Make" {
			m.Make = valueString
		}

		if it.Name == "Model" {
			m.Model = valueString
		}

		if it.Name == "Software" {
			m.Software = valueString
		}

		if it.Name == "DateTime" {
			layOut := "2006:01:02 15:04:05 MST"
			//"2019:12:23 18:46:27"
			timeStamp, _ := time.Parse(layOut, fmt.Sprintf("%s PST", valueString))
			m.DateTime = timeStamp
		}

		if it.Name == "Artist" {
			m.Artist = valueString
		}

		if it.Name == "ExposureTime" {
			m.ExposureTime = valueString
		}

		if it.Name == "FNumber" {
			m.FNumber = valueString
			m.FStop = m.CalculateFSTOP()
		}

		if it.Name == "ISOSpeedRatings" {
			m.ISOSpeedRatings = valueString
		}

		if it.Name == "LightSource" {
			m.LightSource = valueString
		}

		if it.Name == "FocalLength" {
			m.FocalLength = valueString
		}

		if it.Name == "PixelXDimension" {
			m.PixelXDimension = valueString
		}

		if it.Name == "PixelYDimension" {
			m.PixelYDimension = valueString
		}

		if it.Name == "FocalLengthIn35mmFilm" {
			m.FocalLengthIn35mmFilm = valueString
		}

		if it.Name == "LensModel" {
			m.LensModel = valueString
		}

		if it.Name == "ExposureProgram" {
			m.ExposureProgram = valueString
		}

		//fmt.Printf("FQ-IFD-PATH=[%s] ID=(0x%04x) NAME=[%s] COUNT=(%d) TYPE=[%s] VALUE=[%s]\n", fqIfdPath, tagId, it.Name, valueContext.UnitCount, tagType.Name(), valueString)
		return nil
	}

	_, err = exif.Visit(exif.IfdStandard, im, ti, exifData, visitor)

	if err != nil {
		return err
	}
	return nil
}

//DeleteMedia delete the media object base on ID
func (m *MediaModel) DeleteMedia() error {

	//var config db.Config
	var db db.Session

	//config.DBUri = "mongodb://host.docker.internal:27017"
	err := db.NewSession()
	if err != nil {
		return err
	}
	defer db.Close()

	c := db.Client.Database(getMediaDB()).Collection("media")
	_, err = c.DeleteOne(context.TODO(), bson.M{"media_id": m.MediaID})

	if err != nil {
		return err
	}

	return nil
}

//GetMediaListByCategory Obtains the list of media by category sorted by date
func GetMediaListByCategory(category string) ([]MediaModel, error) {
	//var config db.Config
	var db db.Session

	var mediaModels []MediaModel
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

	cur, err := db.Client.Database(getMediaDB()).Collection("media").Find(context.TODO(), filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var m MediaModel
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

//AllMediaSortedByDate retrieve all posts sorted by creation date
func AllMediaSortedByDate() ([]MediaModel, error) {

	//var config db.Config
	var db db.Session

	var mediaModels []MediaModel
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

	cur, err := db.Client.Database(getMediaDB()).Collection("media").Find(context.TODO(), filter, options)
	if err != nil {
		return nil, err
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var m MediaModel
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

//AllCategories return a list of categories
func AllCategories() ([]string, error) {
	//var config db.Config
	var db db.Session

	err := db.NewSession()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	filter := bson.M{}

	options := options.Distinct()

	cur, err := db.Client.Database(getMediaDB()).Collection("media").Distinct(context.TODO(), "category", filter, options)
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
func MediaSearch(searchJSON string) ([]MediaModel, error) {

	r := strings.NewReader(searchJSON)

	var ms MediaSearchQuery
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
		cur, err = db.Client.Database(getMediaDB()).Collection("media").Find(context.TODO(), filter, options)
		if err != nil {
			return nil, err
		}
	} else {
		cur, err = db.Client.Database(getMediaDB()).Collection("media").Find(context.TODO(), bson.M{}, options)
		if err != nil {
			return nil, err
		}
	}

	var mediaModels []MediaModel

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {
		var m MediaModel
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

func getMediaDB() string {
	cfg, err := config.GetConfig()
	if err != nil {
		fmt.Printf("error getting configuration: %s", err)
		return ""
	}

	return cfg.MongoDatabase
}
