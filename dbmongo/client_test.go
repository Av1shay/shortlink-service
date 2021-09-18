package dbmongo

import (
	"context"
	"shortlink-service/shortlink"
	"strconv"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	ctx := context.Background()

	conf := Config{
		URI:           "mongodb://username:password@localhost:26000",
		DbName:        "db",
		ItemsCollName: "shortlinks",
	}
	c, err := New(ctx, conf)
	if err != nil {
		t.Fatalf("error creating mongo client: %v", err)
	}

	item := shortlink.Item{
		Redirects: []shortlink.Redirect{
			{From: 0, To: 24, URL: "https://google.com"},
		},
		Visits: 0,
	}
	key, err := c.CreateGetID(ctx, &item)
	if err != nil {
		t.Fatalf("Error create data: %v", err)
	}

	slItem, err := c.Get(ctx, strconv.FormatUint(key, 10))
	if err != nil {
		t.Fatalf("Error getting data: %v", err)
	}
	t.Logf("got shortlink: %v", slItem)

	t.Log("Testing visits...")
	visCount := 1005
	visKeyId, err := c.CreateGetID(ctx, &item)
	if err != nil {
		t.Fatalf("error creating key: %v", err)
	}
	visKey := strconv.FormatUint(visKeyId, 10)
	for i := 0; i < visCount; i++ {
		go func(k string) {
			err = c.IncVisits(ctx, k)
			if err != nil {
				t.Errorf("Error: %v", err)
				return
			}
		}(visKey)
	}
	time.Sleep(time.Second)
	v, err := c.GetVisits(ctx, visKey)
	if err != nil {
		t.Fatalf("error getting key: %v", err)
	}
	if v != visCount {
		t.Errorf("visits mismatch. expected: %d, got: %d", visCount, v)
	}
	t.Logf("visits: %d", v)

	items, err := c.AsArray(ctx)
	if err != nil {
		t.Fatalf("error getting results: %v", err)
	}
	t.Logf("total items: %d", len(items))

	err = c.Disconnect(ctx)
	if err != nil {
		t.Errorf("failed to disconnect mongo client: %v", err)
	}
}
