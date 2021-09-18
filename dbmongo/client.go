package dbmongo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"shortlink-service/shortlink"
	"strconv"
)

type DocState int

const (
	docStateSoftDelete DocState = 0
	docStateActive     DocState = 1
)

type Config struct {
	URI           string
	DbName        string
	ItemsCollName string
}

type Client struct {
	mongoClient *mongo.Client
	items       *mongo.Collection
	counters    *mongo.Collection
}

func New(ctx context.Context, config Config) (*Client, error) {
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(config.URI))
	if err != nil {
		return nil, err
	}
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}
	c := Client{
		mongoClient: mongoClient,
		items:       mongoClient.Database(config.DbName).Collection(config.ItemsCollName),
		counters:    mongoClient.Database(config.DbName).Collection("counters"),
	}
	return &c, nil
}

func (c *Client) Get(ctx context.Context, key string) (*shortlink.Item, error) {
	var res shortlink.Item
	filter := bson.D{{"key", key}}
	err := c.items.FindOne(ctx, filter).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) Set(ctx context.Context, key string, data *shortlink.Item) error {
	id, err := c.getNextSeq(ctx, "shortlinkId")
	if err != nil {
		return err
	}
	_, err = c.items.InsertOne(ctx, bson.D{
		{"_id", id},
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
	id, err := c.getNextSeq(ctx, "shortlinkId")
	if err != nil {
		return 0, err
	}
	_, err = c.items.InsertOne(ctx, bson.D{
		{"_id", id},
		{"key", strconv.FormatUint(id, 10)},
		{"redirects", data.Redirects},
		{"visits", data.Visits},
		{"state", docStateActive},
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	filter := bson.D{{"key", key}}
	update := bson.D{{"$set", bson.D{{"state", docStateSoftDelete}}}}
	_, err := c.items.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) IncVisits(ctx context.Context, key string) error {
	filter := bson.D{{"key", key}}
	update := bson.D{{"$inc", bson.D{{"visits", 1}}}}
	_, err := c.items.UpdateOne(ctx, filter, update)
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
	cur, err := c.items.Find(ctx, bson.D{{"state", docStateActive}})
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
	if err := c.mongoClient.Disconnect(ctx); err != nil {
		return err
	}
	return nil
}

func (c *Client) getNextSeq(ctx context.Context, name string) (uint64, error) {
	var res struct {
		Seq uint64
	}
	filter := bson.D{{"_id", name}}
	update := bson.D{{"$inc", bson.D{{"seq", 1}}}}
	err := c.counters.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().
		SetReturnDocument(options.After)).
		Decode(&res)
	if err != nil {
		return 0, err
	}
	return res.Seq, nil
}
