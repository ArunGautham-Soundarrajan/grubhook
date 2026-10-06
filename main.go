package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ArunGautham-Soundarrajan/grubhook/internal/config"
	"github.com/ArunGautham-Soundarrajan/grubhook/internal/deliveroo"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("opening db pool: ", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("connecting to db: ", err)
	}
	log.Println("config loaded, database connected")

	client := deliveroo.New()
	defer client.Close()

	client.Login(cfg.DeliverooEmail, cfg.DeliverooPassword)

	orders, err := client.FetchOrders()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(orders, "orders")
}
