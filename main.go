package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

// setupRouter 將「砌 router」同「起 server」分開，
// 咁樣 test 先可以攞住個 router 嚟打，唔使真係 listen port。
// 相當於 NestJS 測試入面 Test.createTestingModule() 嘅角色。
func setupRouter(store BookRepository) *gin.Engine {
	r := gin.Default()
	// Middleware 要喺 routes 之前註冊；佢喺 c.Next() 之後先做嘢，
	// 所以雖然「排頭位」，實際係最後一個掂 response 嘅人（onion 結構）
	r.Use(ErrorHandler())
	handler := NewBookHandler(store)

	books := r.Group("/books")
	{
		books.POST("", handler.CreateBook)
		books.GET("", handler.ListBooks)
		books.GET("/:id", handler.GetBook)
		books.PUT("/:id", handler.UpdateBook)
		books.DELETE("/:id", handler.DeleteBook)
	}

	return r
}

func main() {
	store, err := NewGormBookStore("books.db")
	if err != nil {
		log.Fatalf("Failed to create book store: %v", err)
	}
	r := setupRouter(store)
	r.Run(":8089")
}
