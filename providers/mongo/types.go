package mongo

import (
	"go.mongodb.org/mongo-driver/mongo"
)

// Config provider settings
type Config struct {
	// Key prefix
	KeyPrefix string

	// host:port address.
	Addr string

	// Database to be selected after connecting to the server.
	DB string

	// Collection
	Collection string
}

// Provider backend manager
type Provider struct {
	config     Config
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
}
