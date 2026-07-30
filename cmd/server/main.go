// Command server starts the QRForge HTTP service.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dojah/qrforge/internal/config"
	"github.com/dojah/qrforge/internal/server"
	"github.com/dojah/qrforge/internal/shortener"
)

func main() {
	cfg := config.Load()

	store, closeStore := buildStore(cfg)
	defer closeStore()

	srv, err := server.New(cfg, store)
	if err != nil {
		log.Fatalf("failed to build server: %v", err)
	}

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           srv.Handler(),
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
	}

	go func() {
		log.Printf("QRForge listening on :%s", cfg.Port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("stopped")
}

// buildStore selects the link backend: MongoDB when MONGO_URI is configured,
// otherwise a non-persistent in-memory store. It returns a cleanup function to
// run on shutdown.
func buildStore(cfg config.Config) (shortener.Store, func()) {
	if cfg.MongoURI == "" {
		log.Println("shortener: using in-memory store (set MONGO_URI for persistence)")
		return shortener.NewMemoryStore(), func() {}
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.MongoTimeout)
	defer cancel()

	mongoStore, err := shortener.NewMongoStore(ctx, cfg.MongoURI, cfg.MongoDB, cfg.MongoTimeout)
	if err != nil {
		log.Fatalf("shortener: could not connect to MongoDB: %v", err)
	}
	log.Printf("shortener: using MongoDB store (db=%s)", cfg.MongoDB)

	return mongoStore, func() {
		closeCtx, closeCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer closeCancel()
		if err := mongoStore.Close(closeCtx); err != nil {
			log.Printf("shortener: mongo disconnect error: %v", err)
		}
	}
}
