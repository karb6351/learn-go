package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"context"
	"errors"
	"net/http"
	"os/signal"
	"playground/book/internal/api"
	"playground/book/internal/book"
	"syscall"
	"time"
)

func main() {
	store, err := book.NewGormStore("books.db")
	if err != nil {
		log.Fatalf("Failed to create book store: %v", err)
	}
	_ = godotenv.Load()

	key := os.Getenv("API_KEY")
	if key == "" {
		log.Fatalf("API_KEY is not set")
	}
	r := api.SetupRouter(store, key)

	srv := &http.Server{
		Addr:    ":8089",
		Handler: r,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}
	fmt.Println("server closed")
}
