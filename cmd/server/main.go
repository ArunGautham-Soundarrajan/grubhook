package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/ArunGautham-Soundarrajan/grubhook/internal/api"
	"github.com/ArunGautham-Soundarrajan/grubhook/internal/db"
	"github.com/ArunGautham-Soundarrajan/grubhook/internal/gmailclient"
	"github.com/ArunGautham-Soundarrajan/grubhook/internal/handlers"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	tok := oauth2.Token{
		RefreshToken: os.Getenv("GMAIL_REFRESH_TOKEN"),
	}

	config := oauth2.Config{
		ClientID:     os.Getenv("GMAIL_CLIENT_ID"),
		ClientSecret: os.Getenv("GMAIL_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		Scopes:       []string{gmail.GmailReadonlyScope},
	}

	ctx := context.Background()

	// gmail service setup
	srv, err := gmailclient.New(ctx, tok, config)
	if err != nil {
		log.Fatal("error initialising gmail service", err)
	}

	// postgres setup
	pool, err := pgxpool.New(ctx, os.Getenv("GOOSE_DBSTRING"))
	if err != nil {
		log.Fatal("error opening db pool", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("error connecting to db", err)
	}

	queries := db.New(pool)
	statsHandler := &handlers.StatsHandler{
		GmailSrv: srv,
		Queries:  queries,
		Pool:     pool,
	}

	// router
	router := api.NewRouter(statsHandler)
	log.Fatal(http.ListenAndServe(":8080", router))
}
