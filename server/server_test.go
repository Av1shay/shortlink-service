package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	db "shortlink-service/dbmemory"
	"shortlink-service/shortlink"
	"shortlink-service/shortner"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	ctx := context.Background()

	dbClient, err := db.New(ctx)
	if err != nil {
		log.Fatalf("Error create db client: %v", err)
	}

	shortnerClient, err := shortner.New(ctx, "http://localhost:8080", dbClient)
	if err != nil {
		log.Fatalf("Error create shortner client: %v", err)
	}

	inputs := []shortlink.Input{
		{
			KeyType: shortlink.KeyTypeStandard,
			Redirects: []shortlink.Redirect{
				{From: 0, To: 8, URL: "https://google.com"},
				{From: 8, To: 21, URL: "https://youtube.com"},
				{From: 21, To: 24, URL: "https://github.com"},
			},
		},
		{
			KeyType: shortlink.KeyTypeUuid,
			Redirects: []shortlink.Redirect{
				{From: 0, To: 8, URL: "https://www.yahoo.com"},
				{From: 8, To: 21, URL: "https://pkg.go.dev"},
				{From: 21, To: 24, URL: "https://github.com"},
			},
		},
		{
			KeyType: shortlink.KeyTypeUuid,
			Redirects: []shortlink.Redirect{
				{From: 0, To: 8, URL: "https://www.yahoo.com"},
				{From: 8, To: 21, URL: "https://pkg.go.dev/blablabla"},
			},
		},
	}

	for _, input := range inputs {
		_, err = shortnerClient.GenerateShortLink(ctx, &input)
		if err != nil {
			t.Fatalf("failed to create shortlink: %v", err)
		}
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))

	s, err := New(ctx, shortnerClient, r)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	err = s.CheckRedirects(ctx)
	if err != nil {
		log.Fatalf("error checking redirect: %v", err)
	}

}
