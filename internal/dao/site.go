package dao

import (
	"context"
	"time"

	"github.com/BingyanStudio/is-hust-online/internal/db"
	"github.com/BingyanStudio/is-hust-online/internal/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const siteCollection = "sites"

func InsertSite(ctx context.Context, site *model.Site) error {
	site.CreatedAt = time.Now().Unix()
	_, err := db.MongoDB.Collection(siteCollection).InsertOne(ctx, site)
	return err
}

func FindSiteByID(ctx context.Context, id bson.ObjectID) (*model.Site, error) {
	var site model.Site
	err := db.MongoDB.Collection(siteCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&site)
	if err != nil {
		return nil, err
	}
	return &site, nil
}

func FindSites(ctx context.Context, filter bson.M, page, pageSize int64) ([]model.Site, int64, error) {
	col := db.MongoDB.Collection(siteCollection)

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

	var sites []model.Site
	if err := cursor.All(ctx, &sites); err != nil {
		return nil, 0, err
	}
	return sites, total, nil
}

func UpdateSite(ctx context.Context, id bson.ObjectID, update bson.M) error {
	_, err := db.MongoDB.Collection(siteCollection).UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	return err
}

func DeleteSite(ctx context.Context, id bson.ObjectID) error {
	_, err := db.MongoDB.Collection(siteCollection).DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func FindAllEnabledSites(ctx context.Context) ([]model.Site, error) {
	cursor, err := db.MongoDB.Collection(siteCollection).Find(ctx, bson.M{"status": model.SITE_STATUS_ENABLED})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sites []model.Site
	if err := cursor.All(ctx, &sites); err != nil {
		return nil, err
	}
	return sites, nil
}
