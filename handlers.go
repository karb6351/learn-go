package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type BookRepository interface {
	Create(b Book) (Book, error)
	List() ([]Book, error)
	Get(id int) (Book, error)
	Update(id int, b Book) (Book, error)
	Delete(id int) error
}

// BookHandler 揸住 store 嘅 reference。
// 呢個就係 Go 版嘅 "constructor injection" —
// 冇 DI container，直接喺 main() 度砌好傳入嚟。

type BookParam struct {
	ID int `uri:"id" binding:"min=1"`
}
type BookHandler struct {
	store BookRepository
}

type BookInput struct {
	BaseBook
}

func NewBookHandler(store BookRepository) *BookHandler {
	return &BookHandler{store: store}
}

// POST /books
func (h *BookHandler) CreateBook(c *gin.Context) {
	var input BookInput

	// Handler 唔再自己 translate error：c.Error 將 err 掛上 context，
	// c.Abort 停止 chain，剩低嘅交俾 ErrorHandler middleware
	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		c.Abort()
		return
	}

	created, err := h.store.Create(Book{
		BaseBook: input.BaseBook,
	})
	if err != nil {
		c.Error(err)
		c.Abort()
		return
	}
	c.JSON(http.StatusCreated, created)
}

// GET /books
func (h *BookHandler) ListBooks(c *gin.Context) {
	books, err := h.store.List()
	if err != nil {
		c.Error(err)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, books)
}

// GET /books/:id
func (h *BookHandler) GetBook(c *gin.Context) {
	var paramInput BookParam

	if err := c.ShouldBindUri(&paramInput); err != nil {
		c.Error(err)
		c.Abort()
		return
	}

	book, err := h.store.Get(paramInput.ID)
	if err != nil {
		c.Error(err)
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, book)
}

// PUT /books/:id
func (h *BookHandler) UpdateBook(c *gin.Context) {

	var input BookInput
	var paramInput BookParam

	if err := c.ShouldBindUri(&paramInput); err != nil {
		c.Error(err)
		c.Abort()
		return
	}

	id := paramInput.ID

	if err := c.ShouldBindJSON(&input); err != nil {
		c.Error(err)
		c.Abort()
		return // Gin 唔會自動停，一定要自己 return！
	}

	updated, err := h.store.Update(id, Book{
		BaseBook: input.BaseBook,
	})

	if err != nil {
		c.Error(err)
		c.Abort()
		return
	}
	c.JSON(http.StatusOK, updated)
}

// DELETE /books/:id
func (h *BookHandler) DeleteBook(c *gin.Context) {

	var paramInput BookParam

	if err := c.ShouldBindUri(&paramInput); err != nil {
		c.Error(err)
		c.Abort()
		return
	}

	id := paramInput.ID

	if err := h.store.Delete(id); err != nil {
		c.Error(err)
		c.Abort()
		return
	}

	c.Status(http.StatusNoContent) // 204，冇 body
}
