package dbmongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"shortlink-service/shortlink"
	"strconv"
	"sync"
)

type DocState int

const (
	docStateSoftDelete DocState = 0
	docStateActive     DocState = 1
)

type Config struct {
	Uri            string
	DbName         string
	CollectionName string
}

type Client struct {
	mu         sync.Mutex
	mClient    *mongo.Client
	collection *mongo.Collection
}

func New(ctx context.Context, config Config) (*Client, error) {
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(config.Uri))
	if err != nil {
		return nil, err
	}
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	c := Client{
		mClient:    mongoClient,
		collection: mongoClient.Database(config.DbName).Collection(config.CollectionName),
	}
	return &c, nil
}

func (c *Client) Get(ctx context.Context, key string) (*shortlink.Item, error) {
	var res shortlink.Item
	filter := bson.D{{"key", key}}
	err := c.collection.FindOne(ctx, filter).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) Set(ctx context.Context, key string, data *shortlink.Item) error {
	_, err := c.collection.InsertOne(ctx, bson.D{
		{"key", key},
		{"redirects", data.Redirects},
		{"visits", data.Visits},
		{"state", docStateActive},
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CreateGetID(ctx context.Context, data *shortlink.Item) (uint64, error) {
	c.mu.Lock()
	count, err := c.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		c.mu.Unlock()
		return 0, err
	}
	key := count + 1
	c.mu.Unlock()
	_, err = c.collection.InsertOne(ctx, bson.D{
		{"key", strconv.FormatInt(key, 10)},
		{"redirects", data.Redirects},
		{"visits", data.Visits},
		{"state", docStateActive},
	})
	if err != nil {
		return 0, err
	}
	return uint64(key), nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	filter := bson.D{{"key", key}}
	update := bson.D{{"$set", bson.D{{"state", docStateSoftDelete}}}}
	_, err := c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) IncVisits(ctx context.Context, key string) error {
	filter := bson.D{{"key", key}}
	update := bson.D{{"$inc", bson.D{{"visits", 1}}}}
	_, err := c.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetVisits(ctx context.Context, key string) (int, error) {
	item, err := c.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	return item.Visits, nil
}

func (c *Client) AsArray(ctx context.Context) ([]*shortlink.Item, error) {
	var items []*shortlink.Item
	cur, err := c.collection.Find(ctx, bson.D{{"state", docStateActive}})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result shortlink.Item
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}
		items = append(items, &result)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (c *Client) Disconnect(ctx context.Context) error {
	if err := c.mClient.Disconnect(ctx); err != nil {
		return err
	}
	return nil
}
