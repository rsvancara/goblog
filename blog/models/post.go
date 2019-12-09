package models

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Post post
type Post struct {
	ID       bson.ObjectId `bson:"_id,omitempty"`
	Post     string
	Title    string
	Keywords string
}

func postModelIndex() mgo.Index {
	return mgo.Index{
		Key:        []string{"Keywords"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
}
