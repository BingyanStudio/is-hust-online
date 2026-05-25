package dao

import (
	"context"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/db"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const clientCollection = "clients"

func InsertClient(ctx context.Context, client *model.Client) error {
	client.CreatedAt = time.Now().Unix()
	_, err := db.MongoDB.Collection(clientCollection).InsertOne(ctx, client)
	return err
}

func FindClientByID(ctx context.Context, id bson.ObjectID) (*model.Client, error) {
	var client model.Client
	err := db.MongoDB.Collection(clientCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&client)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func FindClientByToken(ctx context.Context, token string) (*model.Client, error) {
	var client model.Client
	err := db.MongoDB.Collection(clientCollection).FindOne(ctx, bson.M{"token": token}).Decode(&client)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func FindClients(ctx context.Context, filter bson.M, page, pageSize int64) ([]model.Client, int64, error) {
	col := db.MongoDB.Collection(clientCollection)

	total, err := col.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().
		SetSkip((page - 1) * pageSize).
		SetLimit(pageSize).
		SetSort(bson.M{"created_at": -1})

	cursor, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var clients []model.Client
	if err := cursor.All(ctx, &clients); err != nil {
		return nil, 0, err
	}
	return clients, total, nil
}

func UpdateClient(ctx context.Context, id bson.ObjectID, update bson.M) error {
	_, err := db.MongoDB.Collection(clientCollection).UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func DeleteClient(ctx context.Context, id bson.ObjectID) error {
	_, err := db.MongoDB.Collection(clientCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func FindOnlineClients(ctx context.Context) ([]model.Client, error) {
	cursor, err := db.MongoDB.Collection(clientCollection).Find(ctx, bson.M{"status": model.CLIENT_STATUS_ONLINE})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var clients []model.Client
	if err := cursor.All(ctx, &clients); err != nil {
		return nil, err
	}
	return clients, nil
}
