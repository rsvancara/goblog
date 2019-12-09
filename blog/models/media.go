package models

import "gopkg.in/mgo.v2/bson"

// Media represents media
type Media struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	Title       string
	Description string
}
