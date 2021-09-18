package dbmemory

import (
	"context"
	"shortlink-service/shortlink"
	"strconv"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	ctx := context.Background()
	c, err := New(ctx)
	if err != nil {
		t.Fatalf("Error creating dbClient: %v", err)
	}

	shortlinkItem := shortlink.Item{
		Redirects: []shortlink.Redirect{
			{From: 0, To: 24, URL: "https://google.com"},
		},
		Visits: 0,
	}

	t.Log("Testing key concurrent increments...")
	count := 100000
	for i := 0; i < count; i++ {
		go func(item *shortlink.Item) {
			_, err := c.CreateGetID(ctx, item)
			if err != nil {
				t.Errorf("Error: %v", err)
				return
			}
		}(&shortlinkItem)
	}
	time.Sleep(time.Second)
	if c.lastKeyID != uint64(count) {
		t.Error("mismatch key count")
	}

	t.Log("Testing get...")
	key := strconv.Itoa(count)
	item, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("error getting shortlink item: %v", err)
	}
	if item == nil {
		t.Fatalf("failed getting item with key %s", key)
	}
	firstRedirect := item.Redirects[0]
	if firstRedirect.URL != "https://google.com" {
		t.Error("got wrong item url")
	}
	t.Logf("last item: from (%d) to (%d) redirectURL (%s)", firstRedirect.From, firstRedirect.To, firstRedirect.URL)

	t.Log("Testing visits...")
	visCount := 100007
	visKeyId, err := c.CreateGetID(ctx, &shortlinkItem)
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
		t.Fatalf("error get key vists: %v", err)
	}
	if v != visCount {
		t.Errorf("visits mismatch. expected: %d, got: %d", visCount, v)
	}
	t.Logf("visits: %d", v)
}
