package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"random-match/backend/internal/config"
	"random-match/backend/internal/server"
	"random-match/backend/internal/store"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	db, err := store.NewMongo(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("connect mongo: %v", err)
	}
	defer db.Close(context.Background())

	cache, err := store.NewRedis(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
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
