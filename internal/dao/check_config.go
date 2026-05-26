package dao

import (
	"context"

	"github.com/BingyanStudio/is-hust-online/internal/db"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const checkConfigCollection = "check_configs"

func InsertCheckConfig(ctx context.Context, cc *model.CheckConfig) error {
	_, err := db.MongoDB.Collection(checkConfigCollection).InsertOne(ctx, cc)
	return err
}

func FindCheckConfigByID(ctx context.Context, id bson.ObjectID) (*model.CheckConfig, error) {
	var cc model.CheckConfig
	err := db.MongoDB.Collection(checkConfigCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&cc)
	if err != nil {
		return nil, err
	}
	return &cc, nil
}

func FindCheckConfigs(ctx context.Context, filter bson.M, page, pageSize int64) ([]model.CheckConfig, int64, error) {
	col := db.MongoDB.Collection(checkConfigCollection)

	total, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.M{"_id": -1})

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var ccs []model.CheckConfig
	if err := cursor.All(ctx, &ccs); err != nil {
		return nil, 0, err
	}
	return ccs, total, nil
}

func UpdateCheckConfig(ctx context.Context, id bson.ObjectID, update bson.M) error {
	_, err := db.MongoDB.Collection(checkConfigCollection).UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func DeleteCheckConfig(ctx context.Context, id bson.ObjectID) error {
	_, err := db.MongoDB.Collection(checkConfigCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func FindCheckConfigsBySiteID(ctx context.Context, siteID bson.ObjectID) ([]model.CheckConfig, error) {
	cursor, err := db.MongoDB.Collection(checkConfigCollection).Find(ctx, bson.M{"site_id": siteID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ccs []model.CheckConfig
	if err := cursor.All(ctx, &ccs); err != nil {
		return nil, err
	}
	return ccs, nil
}

func FindCheckConfigsByClientID(ctx context.Context, clientID bson.ObjectID) ([]model.CheckConfig, error) {
	cursor, err := db.MongoDB.Collection(checkConfigCollection).Find(ctx, bson.M{"client_id": clientID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ccs []model.CheckConfig
	if err := cursor.All(ctx, &ccs); err != nil {
		return nil, err
	}
	return ccs, nil
}
