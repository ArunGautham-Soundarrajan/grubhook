package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/ArunGautham-Soundarrajan/grubhook/internal/config"
	"github.com/ArunGautham-Soundarrajan/grubhook/internal/deliveroo"
	"github.com/ArunGautham-Soundarrajan/grubhook/internal/store"
	"github.com/go-rod/rod/lib/proto"
	"github.com/jackc/pgx/v5/pgxpool"
)

// scrapeTimeout bounds the whole browser session: login plus fetching orders.
const scrapeTimeout = 3 * time.Minute

func main() {
	// JSON logs so fields (url, err, ...) are queryable once shipped to Loki.
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := run(); err != nil {
		slog.Error("run failed", "err", err)
		os.Exit(1)
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
	slog.Info("config loaded, database connected")

	orders, err := scrapeOrders(ctx, cfg)
	if err != nil {
		return err
	}
	if len(orders) == 0 {
		slog.Info("no orders on account")
	}

	inserted, err := store.SaveOrders(ctx, pool, orders)
	if err != nil {
		return fmt.Errorf("saving orders: %w", err)
	}
	slog.Info("orders synced", "fetched", len(orders), "new", inserted)
	return nil
}

func scrapeOrders(ctx context.Context, cfg config.Config) (orders []deliveroo.Order, err error) {
	ctx, cancel := context.WithTimeout(ctx, scrapeTimeout)
	defer cancel()

	var cookies []*proto.NetworkCookieParam
	if cfg.CookiesFile != "" {
		if cookies, err = deliveroo.LoadCookies(cfg.CookiesFile); err != nil {
			return nil, err
		}
		slog.Info("loaded saved cookies", "count", len(cookies))
	}

	client, err := deliveroo.New(ctx, cookies)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	defer func() {
		if err != nil {
			saveFailureDiagnostics(client, cfg.ArtifactsDir)
		}
	}()

	if err := client.Login(cfg.DeliverooEmail, cfg.DeliverooPassword); err != nil {
		return nil, err
	}
	return client.FetchOrders()
}

// saveFailureDiagnostics logs what the page showed when the scrape failed and
// writes a screenshot and the page HTML to dir. Each step is best effort: a
// failure here is logged, never returned, so it can't mask the real error.
func saveFailureDiagnostics(client *deliveroo.Client, dir string) {
	if d, err := client.Diagnose(); err != nil {
		slog.Error("reading page at failure", "err", err)
	} else {
		slog.Error("page at failure",
			"url", d.URL,
			"title", d.Title,
			"messages", d.Messages,
			"body_text", d.BodyText,
		)
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		slog.Error("creating artifacts dir", "dir", dir, "err", err)
		return
	}
	prefix := filepath.Join(dir, time.Now().UTC().Format("20060102T150405Z")+"-failure")

	if err := client.SaveScreenshot(prefix + ".png"); err != nil {
		slog.Error("saving failure screenshot", "err", err)
	} else {
		slog.Info("saved failure screenshot", "path", prefix+".png")
	}
	if err := client.SaveHTML(prefix + ".html"); err != nil {
		slog.Error("saving failure page HTML", "err", err)
	} else {
		slog.Info("saved failure page HTML", "path", prefix+".html")
	}
}
