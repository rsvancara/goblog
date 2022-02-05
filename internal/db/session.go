package db

import (
	"context"
	"time"

	"goblog/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Session mongodb session
type Session struct {
	Client *mongo.Client
}

// NewSession create new session
func (s *Session) NewSession(cfg config.AppConfig) error {

	var err error

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

// GetMongoClient connect to the mongodb database
func GetMongoClient(config *config.AppConfig) (*mongo.Client, error) {

	//mongodb://master:master1234@docdb-2020-11-18-17-41-21-hss.cluster-cy44hjryoetp.us-west-2.docdb.amazonaws.com:27017/?ssl=true&ssl_ca_certs=rds-combined-ca-bundle.pem&replicaSet=rs0&readPreference=secondaryPreferred&retryWrites=false

	//mongoURI := fmt.Sprintf("mongodb://%s:%s@%s/?ssl=false&replicaSet=rs0&readPreference=secondaryPreferred&retryWrites=false", config.MongoUser, config.MongoPass, config.MongoHost)
	mongoURI := config.Dburi

	//tlsConfig, err := GetTLSConfig(config.MongoCert)
	//if err != nil {
	//	log.Error().Err(err).Str("service", "utility").Msgf("Error getting tls configuration from path: %s", config.MongoCert)
	//}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	//options := options.Client().ApplyURI(mongoURI).SetTLSConfig(tlsConfig).SetMaxPoolSize(50)
	options := options.Client().ApplyURI(mongoURI).SetMaxPoolSize(10)
	client, err := mongo.Connect(ctx, options)
	if err != nil {
		return nil, err
	}

	return client, nil
}
