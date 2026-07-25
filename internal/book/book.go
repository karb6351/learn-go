package book

import (
	"sort"
	"sync"

	"playground/book/internal/apperr"
)

type BaseBook struct {
	Title  string `json:"title" binding:"required"`
	Author string `json:"author" binding:"required"`
	Year   int    `json:"year" binding:"gte=0,lte=2100"`
}

// Book 係我哋嘅 domain model。
// 留意兩組 tags：
//   - json:"..."  控制 JSON serialize 出嚟嘅 field 名（類似 Jackson 嘅 @JsonProperty）
//   - binding:"..." 係 Gin 嘅 validation（類似 class-validator / @Valid）
type Book struct {
	ID int `json:"id"`
	BaseBook
}

// MemoryStore 係 in-memory repository。
// Go 嘅 HTTP server 每個 request 一條 goroutine，
// 所以共享嘅 map 一定要用 mutex 保護（map 唔係 thread-safe）。
type MemoryStore struct {
	mu     sync.RWMutex
	books  map[int]Book
	nextID int
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		books:  make(map[int]Book),
		nextID: 1,
	}
}

// Create 接收一本冇 ID 嘅書，assign ID 之後存起佢。
func (s *MemoryStore) Create(b Book) (Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b.ID = s.nextID
	s.nextID++
	s.books[b.ID] = b
	return b, nil
}

// List 回傳所有書。用 RLock 因為只係讀，多個 reader 可以並行。
func (s *MemoryStore) List() ([]Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Book, 0, len(s.books))
	for _, b := range s.books {
		result = append(result, b)
	}
	// sort the result slice by id
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result, nil
}

// Get 回傳單一本書。Go 嘅慣例係回傳 (value, error) 一對。
func (s *MemoryStore) Get(id int) (Book, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	b, ok := s.books[id]
	if !ok {
		return Book{}, &apperr.ResourceNotFoundError{Resource: "book", ID: id}
	}
	return b, nil
}

// Update 整本替換（PUT 語義）。
func (s *MemoryStore) Update(id int, b Book) (Book, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.books[id]; !ok {
		return Book{}, &apperr.ResourceNotFoundError{Resource: "book", ID: id}
	}
	b.ID = id
	s.books[id] = b
	return b, nil
}

// Delete 刪除一本書。
func (s *MemoryStore) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.books[id]; !ok {
		return &apperr.ResourceNotFoundError{Resource: "book", ID: id}
	}
	delete(s.books, id)
	return nil
}
