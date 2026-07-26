package main

import (
	"log"
	"os"

	"playground/book/internal/api"
	"playground/book/internal/book"
)

func main() {
	store, err := book.NewGormStore("books.db")
	if err != nil {
		log.Fatalf("Failed to create book store: %v", err)
	}
	key := os.Getenv("API_KEY")
	if key == "" {
		log.Fatalf("API_KEY is not set")
	}
	r := api.SetupRouter(store, key)
	r.Run(":8089")
}
