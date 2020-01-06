package models

import (
	"blog/blog/db"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2/bson"
)

// MediaModel post
type MediaModel struct {
	ID                    primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	MediaID               string             `json:"media_id" bson:"media_id,omitempty"`
	Keywords              string             `json:"keywords" bson:"keywords,omitempty"`
	FileName              string             `json:"file_name" bson:"file_name,omitempty"`
	S3Location            string             `json:"s3_location" bson:"s3_location,omitempty"`
	S3Uploaded            string             `json:"s3_uploaded" bson:"s3_uploaded,omitempty"`
	Description           string             `json:"description" bson:"description,omitempty"`
	Checksum              string             `json:"checksum" bson:"checksum,omitempty"`
	CreatedAt             time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt             time.Time          `json:"updated_at" bson:"updated_at"`
	Make                  string             `json:"make" bson:"make,omitempty"`                           //SONY
	Model                 string             `json:"model" bson:"model,omitempty"`                         //ILCE-7RM3
	Software              string             `json:"software" bson:"software,omitempty"`                   //ILCE-7RM3 v2.10
	DateTime              time.Time          `json:"datetime_taken" bson:"datetime_taken"`                 //2019:12:23 18:46:27
	Artist                string             `json:"artist" bson:"artist,omitempty"`                       //randall svancara
	Copyright             string             `json:"copyright" bson:"copyright,omitempty"`                 //vi
	ExposureTime          string             `json:"exposuretime" bson:"exposuretime,omitempty"`           //1/30
	FNumber               string             `json:"fnumber" bson:"fnumber,omitempty"`                     //14/5
	ISOSpeedRatings       string             `json:"iso_speed_rating" bson:"iso_speed_rating,omitempty"`   //1600
	LightSource           string             `json:"light_source" bson:"light_source,omitempty"`           //0
	FocalLength           string             `json:"focal_length" bson:"focal_length,omitempty"`           //23/1
	PixelXDimension       string             `json:"pixel_x_dimension" bson:"pixel_x_dimension,omitempty"` //7968
	PixelYDimension       string             `json:"pixel_y_dimension" bson:"pixel_y_dimension,omitempty"` //5320
	FocalLengthIn35mmFilm string             `json:"focal_length35" bson:"focal_length35,omitempty"`       //23
	LensModel             string             `json:"lens_model" bson:"lens_model,omitempty"`               //FE 16-35mm F2.8 GM
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
	m.MediaID = genUUID()

	c := db.Client.Database("blog").Collection("media")

	insertResult, err := c.InsertOne(context.TODO(), m)
	if err != nil {
		return err
	}

	// Convert to object ID
	m.ID = insertResult.InsertedID.(primitive.ObjectID)

	return nil
}

//SetS3Uploaded sets the status of the s3upload
func (m *MediaModel) SetS3Uploaded(status string, s3_location string) error {

	var db db.Session

	err := db.NewSession()
	if err != nil {
		return err
	}

	defer db.Close()

	c := db.Client.Database("blog").Collection("media")

	filter := bson.M{
		"media_id": bson.M{
			"$eq": m.MediaID, // check if bool field has value of 'false'
		},
	}

	update := bson.M{
		"$set": bson.M{
			"s3_uploaded": status,
			"s3_location": s3_location,
		},
	}

	result, err := c.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	fmt.Println(result)

	return nil
}
