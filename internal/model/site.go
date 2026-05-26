package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	SITE_STATUS_ENABLED = iota
	SITE_STATUS_DISABLED
)

type Site struct {
	ID          bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string        `json:"name" bson:"name"`
	Description string        `json:"description" bson:"description"`
	URL         string        `json:"url" bson:"url"`
	Status      int           `json:"status" bson:"status"`
	Type        string        `json:"type" bson:"type"`
	CreatedAt   int64         `json:"created_at" bson:"created_at"`
}
