package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config mongo config
type Config struct {
	DBUri  string `json:"dburi"`
	Secret string `json:"secret"`
}

// Session mongodb session
type Session struct {
	Client *mongo.Client
}

// NewSession create new session
func (s *Session) NewSession(config *Config) error {

	var err error

	clientOptions := options.Client().ApplyURI(config.DBUri)
	s.Client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return err
	}

	return nil
}

// Close close session
func (s *Session) Close() error {

	// Close the connection once no longer needed
	err := s.Client.Disconnect(context.TODO())
	if err != nil {
		return err
	} else {
		fmt.Println("Connection to MongoDB closed.")
	}
	return nil
}
