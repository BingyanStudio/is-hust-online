package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	CLIENT_STATUS_ONLINE = iota
	CLIENT_STATUS_OFFLINE
	CLIENT_STATUS_DISABLED
)

type Client struct {
	ID           bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Token        string        `json:"token" bson:"token"`
	Name         string        `json:"name" bson:"name"`
	Location     string        `json:"location" bson:"location"`
	Capabilities int32         `json:"capabilities" bson:"capabilities"`
	Status       int           `json:"status" bson:"status"`
	IP           string        `json:"ip" bson:"ip"`
	LastOnline   int64         `json:"last_online" bson:"last_online"`
	CreatedAt    int64         `json:"created_at" bson:"created_at"`
}
