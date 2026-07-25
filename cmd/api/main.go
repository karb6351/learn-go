package main

import (
	"log"

	"playground/book/internal/api"
	"playground/book/internal/book"
)

func main() {
	store, err := book.NewGormStore("books.db")
	if err != nil {
		log.Fatalf("Failed to create book store: %v", err)
	}
	r := api.SetupRouter(store)
	r.Run(":8089")
}
