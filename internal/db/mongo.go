package db

import (
	"context"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	MongoClient *mongo.Client
	MongoDB     *mongo.Database
)

func InitMongoDB(conf config.MongoConfig) error {
	var err error

	MongoClient, err = mongo.Connect(nil, options.Client().ApplyURI(conf.URI))
	if err != nil {
		return err
	}
	MongoDB = MongoClient.Database(conf.Database)
	return nil
}

func CloseMongoDB() error {
	if MongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return MongoClient.Disconnect(ctx)
	}
	return nil
}
