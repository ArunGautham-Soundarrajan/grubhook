package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ArunGautham-Soundarrajan/grubhook/internal/gmailclient"
	"github.com/ArunGautham-Soundarrajan/grubhook/internal/parser"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
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

	srv, err := gmailclient.New(ctx, tok, config)
	if err != nil {
		log.Fatal("error initialising gmail service", err)
	}

	messages, err := parser.GetLast30dMessageIDs(ctx, srv, "noreply@t.deliveroo.com")
	if err != nil {
		log.Fatalf("error fetching messages: %v", err)
	}
	for _, m := range messages {
		orderDetails, err := parser.GetOrderTotalFromMsg(ctx, srv, m)
		if err != nil {
			log.Println("error parsing details for message id:", m.Id, err)
		}
		fmt.Printf("%+v\n", orderDetails)
	}
}
