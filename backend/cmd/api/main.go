package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "random-match/backend/docs"
	"random-match/backend/internal/config"
	"random-match/backend/internal/server"
	"random-match/backend/internal/store"
)

// @title						Random Match API
// @version					1.0
// @description				Backend API for anonymous random voice/video matching and WebSocket signaling.
// @BasePath					/
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer " followed by a JWT token.
func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, err := connectMongo(ctx, cfg)
	if err != nil {
		log.Fatalf("connect mongo: %v", err)
	}
	defer db.Close(context.Background())

	cache, err := connectRedis(ctx, cfg)
	if err != nil {
		log.Fatalf("connect redis: %v", err)
	}
	defer cache.Close()

	app := server.New(cfg, db, cache)
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           app.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("api listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func connectMongo(ctx context.Context, cfg config.Config) (*store.Mongo, error) {
	return retry(ctx, 60*time.Second, 2*time.Second, "mongo", func() (*store.Mongo, error) {
		return store.NewMongo(ctx, cfg.MongoURI, cfg.MongoDB)
	})
}

func connectRedis(ctx context.Context, cfg config.Config) (*store.Redis, error) {
	return retry(ctx, 60*time.Second, 2*time.Second, "redis", func() (*store.Redis, error) {
		return store.NewRedis(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	})
}

func retry[T any](ctx context.Context, timeout, interval time.Duration, name string, connect func() (T, error)) (T, error) {
	deadline := time.Now().Add(timeout)
	var zero T
	var lastErr error

	for {
		value, err := connect()
		if err == nil {
			return value, nil
		}
		lastErr = err
		if time.Now().After(deadline) {
			return zero, lastErr
		}
		log.Printf("connect %s failed, retrying in %s: %v", name, interval, err)

		select {
		case <-ctx.Done():
			return zero, ctx.Err()
		case <-time.After(interval):
		}
	}
}
