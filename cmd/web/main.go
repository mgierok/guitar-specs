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

	"guitar-specs/internal/app"
)

func main() {
	cfg := app.LoadConfig()
	a := app.New(cfg)

	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           a.Router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Println("server starting", "addr", cfg.Addr())
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Println("server error:", err)
			os.Exit(1)
		}
	}()

	// SIGINT/SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("server stopped")
}
