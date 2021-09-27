package shortner

import (
	"context"
	"errors"
	"fmt"
	uuid "github.com/nu7hatch/gouuid"
	"shortlink-service/encoder"
	"shortlink-service/shortlink"
	"strconv"
	"time"
)

type DbClient interface {
	Get(ctx context.Context, key string) (*shortlink.Item, error)
	Set(ctx context.Context, key string, data *shortlink.Item) error
	CreateGetID(ctx context.Context, data *shortlink.Item) (uint64, error)
	Delete(ctx context.Context, key string) error
	IncVisits(ctx context.Context, key string) error
	GetVisits(ctx context.Context, key string) (int, error)
	AsArray(ctx context.Context) ([]*shortlink.Item, error)
}

type Client struct {
	baseUrl  string
	dbClient DbClient
}

func New(ctx context.Context, shortlinkBaseURL string, dbClient DbClient) (*Client, error) {
	c := Client{
		baseUrl:  shortlinkBaseURL,
		dbClient: dbClient,
	}
	return &c, nil
}

func (c *Client) GenerateShortLink(ctx context.Context, data *shortlink.Input) (string, error) {
	item := &shortlink.Item{
		Redirects: data.Redirects,
		Visits:    0,
	}

	if data.KeyType == shortlink.KeyTypeUuid {
		u, err := uuid.NewV4()
		if err != nil {
			return "", err
		}
		key := u.String()
		err = c.dbClient.Set(ctx, key, item)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s/u/%s", c.baseUrl, key), nil
	}

	id, err := c.dbClient.CreateGetID(ctx, item)
	if err != nil {
		return "", err
	}
	key := encoder.Encode(id)

	return fmt.Sprintf("%s/%s", c.baseUrl, key), nil
}

func (c *Client) GetLongURL(ctx context.Context, originKey string, t time.Time, kt shortlink.KeyType, incVisits bool) (string, error) {
	var key = originKey

	if kt == shortlink.KeyTypeStandard {
		decoded, err := encoder.Decode(key)
		if err != nil {
			return "", err
		}
		key = strconv.FormatUint(decoded, 10)
	}

	if incVisits {
		err := c.dbClient.IncVisits(ctx, key)
		if err != nil {
			return "", err
		}
	}

	data, err := c.dbClient.Get(ctx, key)
	if err != nil {
		return "", err
	}
	if data == nil {
		return "", fmt.Errorf("shortlink data is not exist for key %s", key)
	}

	return getUrlByTime(data, t)
}


func (c *Client) GelAllShortLinks(ctx context.Context) ([]*shortlink.Item, error) {
	return c.dbClient.AsArray(ctx)
}

func (c *Client) DeleteShortLink(ctx context.Context, key string) error {
	return c.dbClient.Delete(ctx, key)
}

func getUrlByTime(item *shortlink.Item, t time.Time) (string, error) {
	if len(item.Redirects) == 0 || len(item.Redirects) > 24 {
		return "", errors.New("invalid number of redirects")
	}

	h := t.Hour()
	for _, r := range item.Redirects {
		if h >= r.From && h < r.To {
			return r.URL, nil
		}
	}

	return item.Redirects[0].URL, nil
}
