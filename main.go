package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"shortlink-service/dbmongo"
	"shortlink-service/server"
	"shortlink-service/shortner"
	"time"
)

const defaultPort = "8080"

func main() {
	ctx := context.Background()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	dbClient, err := dbmongo.New(ctx, dbmongo.Config{
		Uri:            os.Getenv("MONGO_URI"),
		DbName:         os.Getenv("MONGO_DB"),
		CollectionName: os.Getenv("MONGO_COLLECTION"),
	})

	shortnerClient, err := shortner.New(ctx, os.Getenv("SHORTLINK_BASE_URL"), dbClient)
	if err != nil {
		log.Fatalf("Error create shortner client: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))

	_, err = server.New(ctx, shortnerClient, r)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Fatal(http.ListenAndServe(":"+port, r))
}
