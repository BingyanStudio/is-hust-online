package model

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

const (
	CLIENT_STATUS_ONLINE   = 1 // matches proto ClientStatus_CLIENT_STATUS_ONLINE
	CLIENT_STATUS_OFFLINE  = 4 // matches proto ClientStatus_CLIENT_STATUS_OFFLINE
	CLIENT_STATUS_DISABLED = 5 // no direct proto equivalent, placed after OFFLINE
)

type Client struct {
	ID           bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Token        string        `json:"token,omitzero" bson:"token"`
	Name         string        `json:"name" bson:"name"`
	Location     string        `json:"location" bson:"location"`
	Capabilities int32         `json:"capabilities" bson:"capabilities"`
	Status       int           `json:"status" bson:"status"`
	IP           string        `json:"ip,omitzero" bson:"ip"`
	LastOnline   int64         `json:"last_online" bson:"last_online"`
	CreatedAt    int64         `json:"created_at" bson:"created_at"`
	Labels       []string      `json:"labels" bson:"labels"`
}
