package mongo

import (
	"context"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"

	"github.com/valyala/bytebufferpool"
	"go.mongodb.org/mongo-driver/mongo"
)

var all = []byte("*")

// New returns a new configured redis provider
func New(cfg Config) (*Provider, error) {
	if cfg.Addr == "" {
		return nil, errConfigAddrEmpty
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(cfg.Addr))
	if err != nil {
		return nil, errMongoConnection(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return nil, errMongoConnection(err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, errMongoConnection(err)
	}

	db := client.Database(cfg.DB)
	collection := db.Collection(cfg.Collection)

	p := &Provider{
		config:     cfg,
		client:     client,
		db:         db,
		collection: collection,
	}

	return p, nil
}

func (p *Provider) getMongoSessionKey(sessionID []byte) string {
	key := bytebufferpool.Get()
	key.SetString(p.config.KeyPrefix)
	key.WriteString(":")
	key.Write(sessionID)

	keyStr := key.String()

	bytebufferpool.Put(key)

	return keyStr
}

// Get returns the data of the given session id
func (p *Provider) Get(id []byte) ([]byte, error) {
	key := p.getMongoSessionKey(id)
	ctx, _ := context.WithTimeout(context.Background(), 30 * time.Second)


	reply, err := p.collection.FindOne(ctx, bson.M{
		"_id": key,
	}).DecodeBytes()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	return reply, nil

}

// Save saves the session data and expiration from the given session id
func (p *Provider) Save(id, data []byte, expiration time.Duration) error {
	key := p.getMongoSessionKey(id)
	ctx, _ := context.WithTimeout(context.Background(), 30 * time.Second)

	_, err :=  p.collection.InsertOne(ctx, bson.M{
		"_id": key,
		"data": data,
		"expiration": expiration,
	})
	return err
}

// Regenerate updates the session id and expiration with the new session id
// of the the given current session id
func (p *Provider) Regenerate(id, newID []byte, expiration time.Duration) error {
	key := p.getMongoSessionKey(id)
	newKey := p.getMongoSessionKey(newID)
	ctx, _ := context.WithTimeout(context.Background(), 30 * time.Second)

	exists, err := p.collection.FindOne(ctx, bson.M{
		"_id": key,
	}).DecodeBytes()
	if err != nil {
		return err
	}

	if len(exists) > 0 { // Exist
		if err = p.collection.FindOneAndUpdate(ctx, bson.M{
			"_id": key,
		}, bson.M{
			"$set": bson.M{
				"_id": newKey,
			},
		}).Err(); err != nil {
			return err
		}

		//if err = p.db.Expire(context.Background(), newKey, expiration).Err(); err != nil {
		//	return err
		//}
	}

	return nil
}

// Destroy destroys the session from the given id
func (p *Provider) Destroy(id []byte) error {
	key := p.getMongoSessionKey(id)
	ctx, _ := context.WithTimeout(context.Background(), 30 * time.Second)

	_, err := p.collection.DeleteOne(ctx, bson.M{
		"_id": key,
	})
	return err
}

// Count returns the total of stored sessions
func (p *Provider) Count() int {
	ctx, _ := context.WithTimeout(context.Background(), 30 * time.Second)
	reply, err := p.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0
	}

	return int(reply)
}

// NeedGC indicates if the GC needs to be run
func (p *Provider) NeedGC() bool {
	return false
}

// GC destroys the expired sessions
func (p *Provider) GC() {}
