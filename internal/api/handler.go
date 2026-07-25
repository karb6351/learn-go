package api

import (
	"playground/book/internal/book"

	"github.com/gin-gonic/gin"
)

// setupRouter 將「砌 router」同「起 server」分開，
// 咁樣 test 先可以攞住個 router 嚟打，唔使真係 listen port。
// 相當於 NestJS 測試入面 Test.createTestingModule() 嘅角色。
func SetupRouter(store book.Repository) *gin.Engine {
	r := gin.Default()
	// Middleware 要喺 routes 之前註冊；佢喺 c.Next() 之後先做嘢，
	// 所以雖然「排頭位」，實際係最後一個掂 response 嘅人（onion 結構）
	r.Use(ErrorHandler())
	bookHandler := NewBookHandler(store)

	books := r.Group("/books")
	{
		books.POST("", bookHandler.Create)
		books.GET("", bookHandler.List)
		books.GET("/:id", bookHandler.Get)
		books.PUT("/:id", bookHandler.Update)
		books.DELETE("/:id", bookHandler.Delete)
	}

	return r
}
