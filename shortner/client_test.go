package shortner

import (
	"context"
	"path/filepath"
	"shortlink-service/dbmemory"
	"shortlink-service/shortlink"
	"testing"
	"time"
)

func TestClient_GenerateShortLink(t *testing.T) {
	ctx := context.Background()

	dbClient, err := dbmemory.New(ctx)
	if err != nil {
		t.Fatalf("error creating db client: %v", err)
	}

	c, err := New(ctx, "http://localhost", dbClient)
	if err != nil {
		t.Fatalf("error creating client: %v", err)
	}

	input := shortlink.Input{
		KeyType: shortlink.KeyTypeStandard,
		Redirects: []shortlink.Redirect{
			{From: 0, To: 8, URL: "https://google.com"},
			{From: 8, To: 21, URL: "https://youtube.com"},
			{From: 21, To: 24, URL: "https://github.com"},
		},
	}

	t.Log("Generating standard shortlink....")
	sl, err := c.GenerateShortLink(ctx, &input)
	if err != nil {
		t.Errorf("error generating shortlink: %v", err)
	} else {
		t.Logf("generated shortlink: %s", sl)
	}

	t.Log("Generating uuid shortlink....")
	input.KeyType = shortlink.KeyTypeUuid
	slUid, err := c.GenerateShortLink(ctx, &input)
	if err != nil {
		t.Fatalf("error generating uuid shortlink: %v", err)
	} else {
		t.Logf("generated uuid shortlink: %s", slUid)
	}

	t.Log("Getting url from key....")
	slKey := filepath.Base(sl)
	n := time.Now()
	urlTime := time.Date(n.Year(), n.Month(), n.Day(), 14, 0, 0, 0, n.Location())
	url, err := c.GetLongURL(ctx, slKey, urlTime, shortlink.KeyTypeStandard, false)
	if err != nil {
		t.Fatalf("failed to get shortlink data by key: %v", err)
	}
	if url != "https://youtube.com" {
		t.Fatalf("url is not correct. expected: %s, got: %s", "https://youtube.com", url)
	}
	t.Logf("found url %s", url)

	t.Log("Getting url from uuid key....")
	uuidKey := filepath.Base(slUid)
	urlTime = time.Date(n.Year(), n.Month(), n.Day(), 21, 0, 0, 0, n.Location())
	url, err = c.GetLongURL(ctx, uuidKey, urlTime, shortlink.KeyTypeUuid, false)
	if err != nil {
		t.Fatalf("failed to get shortlink data by key: %v", err)
	}
	if url != "https://github.com" {
		t.Fatalf("url is not correct. expected: %s, got: %s", "https://youtube.com", url)
	}
	t.Logf("found url %s", url)
}
