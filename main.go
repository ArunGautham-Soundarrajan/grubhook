package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ArunGautham-Soundarrajan/grubhook/internal/config"
	"github.com/ArunGautham-Soundarrajan/grubhook/internal/deliveroo"
	"github.com/ArunGautham-Soundarrajan/grubhook/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

// scrapeTimeout bounds the whole browser session: login plus fetching orders.
const scrapeTimeout = 3 * time.Minute

const failureScreenshot = "failure.png"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run holds the real work so deferred cleanup runs before main exits.
func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("opening db pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	log.Println("config loaded, database connected")

	orders, err := scrapeOrders(ctx, cfg)
	if err != nil {
		return err
	}
	if len(orders) == 0 {
		log.Println("no orders on account")
	}

	inserted, err := store.SaveOrders(ctx, pool, orders)
	if err != nil {
		return fmt.Errorf("saving orders: %w", err)
	}
	log.Printf("fetched %d orders, %d new", len(orders), inserted)
	return nil
}

func scrapeOrders(ctx context.Context, cfg config.Config) (orders []deliveroo.Order, err error) {
	ctx, cancel := context.WithTimeout(ctx, scrapeTimeout)
	defer cancel()

	client, err := deliveroo.New(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	defer func() {
		if err == nil {
			return
		}
		if shotErr := client.SaveScreenshot(failureScreenshot); shotErr != nil {
			log.Println("saving failure screenshot: ", shotErr)
		} else {
			log.Println("saved failure screenshot to ", failureScreenshot)
		}
	}()

	if err := client.Login(cfg.DeliverooEmail, cfg.DeliverooPassword); err != nil {
		return nil, err
	}
	return client.FetchOrders()
}
