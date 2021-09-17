package dbmemory

import (
	"context"
	"errors"
	"fmt"
	"shortlink-service/shortlink"
	"strconv"
	"sync"
)

type Client struct {
	lastKeyID uint64
	mu        sync.Mutex
	storage   map[string]*shortlink.Item
}

func New(ctx context.Context) (*Client, error) {
	var lastKeyID uint64 = 0
	storage := make(map[string]*shortlink.Item)
	c := Client{lastKeyID: lastKeyID, storage: storage}
	return &c, nil
}

func (c *Client) Get(ctx context.Context, key string) (*shortlink.Item, error) {
	if item, ok := c.storage[key]; ok {
		return item, nil
	}
	return nil, fmt.Errorf("item with key %s is not exist", key)
}

func (c *Client) Set(ctx context.Context, key string, data *shortlink.Item) error {
	if _, ok := c.storage[key]; ok {
		return errors.New("key already exist")
	}
	c.storage[key] = data
	return nil
}

func (c *Client) CreateGetID(ctx context.Context, data *shortlink.Item) (uint64, error) {
	c.mu.Lock()
	c.lastKeyID++
	idKey := strconv.FormatUint(c.lastKeyID, 10)
	c.storage[idKey] = data
	c.mu.Unlock()
	return c.lastKeyID, nil
}

func (c *Client) Delete(ctx context.Context, key string) error {
	_, ok := c.storage[key]
	if ok {
		delete(c.storage, key)
	}
	return nil
}

func (c *Client) IncVisits(ctx context.Context, key string) error {
	c.mu.Lock()
	if item, ok := c.storage[key]; ok {
		item.Visits++
	}
	c.mu.Unlock()
	return nil
}

func (c *Client) GetVisits(ctx context.Context, key string) (int, error) {
	if item, ok := c.storage[key]; ok {
		return item.Visits, nil
	}
	return 0, errors.New("key is not exist")
}

func (c *Client) AsArray(ctx context.Context) ([]*shortlink.Item, error) {
	var items []*shortlink.Item

	for key, item := range c.storage {
		items = append(items, &shortlink.Item{
			Key:       key,
			Redirects: item.Redirects,
			Visits:    item.Visits,
		})
	}

	return items, nil
}
