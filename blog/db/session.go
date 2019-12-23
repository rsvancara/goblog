package db

import (
	"context"

	"blog/blog/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Session mongodb session
type Session struct {
	Client *mongo.Client
}

// NewSession create new session
func (s *Session) NewSession() error {

	var err error

	cfg, err := config.GetConfig()

	clientOptions := options.Client().ApplyURI(cfg.Dburi)
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
	}

	return nil
}
