package book

// Go testing 三條基本規則：
//   1. file 名以 _test.go 結尾（build 時會被剔走，唔會入正式 binary）
//   2. function 叫 TestXxx(t *testing.T)
//   3. 跑：go test ./...     詳細模式：go test -v ./...
//
// 冇 assert library！Go 嘅立場係：用普通 if + t.Errorf 就夠。
//   t.Errorf = 記低 fail，繼續跑落去（似 soft assertion）
//   t.Fatalf = 記低 fail，即刻停呢個 test（後面步驟冇意義時用）

import (
	"errors"
	"path/filepath"
	"testing"

	"playground/book/internal/apperr"
)

type storeFactory func(t *testing.T) Repository

func testBookRepository(t *testing.T, factory storeFactory) {

	t.Run("list returns empty slice", func(t *testing.T) {
		store := factory(t)
		pageResult, err := store.List(ListParams{Page: 1, Limit: 10})
		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}
		if pageResult.Total != 0 {
			t.Errorf("List() = %v, want []", pageResult.Total)
		}
		if len(pageResult.Items) != 0 {
			t.Errorf("List() = %v, want []", len(pageResult.Items))
		}
	})

	t.Run("list returns all books", func(t *testing.T) {
		store := factory(t)
		_, err := store.Create(Book{BaseBook: BaseBook{Title: "A", Author: "X", Year: 2020}})
		if err != nil {
			t.Fatalf("Create(Book{BaseBook: BaseBook{Title: A, Author: X, Year: 2020}}) returned unexpected error: %v", err)
		}
		_, err = store.Create(Book{BaseBook: BaseBook{Title: "B", Author: "Y", Year: 2021}})
		if err != nil {
			t.Fatalf("Create(Book{BaseBook: BaseBook{Title: B, Author: Y, Year: 2021}}) returned unexpected error: %v", err)
		}
		_, err = store.Create(Book{BaseBook: BaseBook{Title: "C", Author: "Z", Year: 2022}})
		if err != nil {
			t.Fatalf("Create(Book{BaseBook: BaseBook{Title: C, Author: Z, Year: 2022}}) returned unexpected error: %v", err)
		}
		pageResult, err := store.List(ListParams{Page: 1, Limit: 10})
		if err != nil {
			t.Fatalf("List() returned unexpected error: %v", err)
		}
		if len(pageResult.Items) != 3 {
			t.Fatalf("List() = %v, want 3", len(pageResult.Items))
		}
		if pageResult.Items[0].Title != "A" {
			t.Errorf("List() = %v, want [A, B, C]", pageResult.Items[0].Title)
		}
		if pageResult.Items[1].Title != "B" {
			t.Errorf("List() = %v, want [A, B, C]", pageResult.Items[1].Title)
		}
		if pageResult.Items[2].Title != "C" {
			t.Errorf("List() = %v, want [A, B, C]", pageResult.Items[2].Title)
		}
		if pageResult.Total != 3 {
			t.Errorf("List() = %v, want 3", pageResult.Total)
		}
		pageResult, err = store.List(ListParams{Page: 2, Limit: 2})
		if err != nil {
			t.Fatalf("List(Page: 2, Limit: 2) returned unexpected error: %v", err)
		}
		if len(pageResult.Items) != 1 {
			t.Errorf("List(Page: 2, Limit: 2) = %v, want 1", len(pageResult.Items))
		}
		if pageResult.Total != 3 {
			t.Errorf("List(Page: 2, Limit: 2) = %v, want 3", pageResult.Total)
		}
		pageResult, err = store.List(ListParams{Page: 99, Limit: 10})
		if err != nil {
			t.Fatalf("List(Page: 99, Limit: 10) returned unexpected error: %v", err)
		}
		if len(pageResult.Items) != 0 {
			t.Errorf("List(Page: 99, Limit: 10) = %v, want 0", len(pageResult.Items))
		}
		if pageResult.Total != 3 {
			t.Errorf("List(Page: 99, Limit: 10) = %v, want 3", pageResult.Total)
		}
	})

	t.Run("create assigns IDs", func(t *testing.T) {
		store := factory(t)

		b1, err := store.Create(Book{BaseBook: BaseBook{Title: "A", Author: "X", Year: 2020}})
		if err != nil {
			t.Fatalf("Create(Book{BaseBook: BaseBook{Title: A, Author: X, Year: 2020}}) returned unexpected error: %v", err)
		}
		b2, err := store.Create(Book{BaseBook: BaseBook{Title: "B", Author: "Y", Year: 2021}})
		if err != nil {
			t.Fatalf("Create(Book{BaseBook: BaseBook{Title: B, Author: Y, Year: 2021}}) returned unexpected error: %v", err)
		}

		if b1.ID < 1 || b2.ID < 1 {
			t.Errorf("book IDs = %d, %d, should be greater than 0", b1.ID, b2.ID)
		}

		if b1.ID == b2.ID {
			t.Errorf("book IDs = %d, %d, should be different", b1.ID, b2.ID)
		}
	})

	t.Run("get returns stored book", func(t *testing.T) {
		store := factory(t)

		created, err := store.Create(Book{BaseBook: BaseBook{Title: "A", Author: "X", Year: 2020}})
		if err != nil {
			t.Fatalf("Create(Book{BaseBook: BaseBook{Title: A, Author: X, Year: 2020}}) returned unexpected error: %v", err)
		}

		got, err := store.Get(created.ID)
		if err != nil {
			t.Fatalf("Get(%d) returned unexpected error: %v", created.ID, err)
		}
		// struct 冇 pointer field 嘅話可以直接 == 比較
		if got != created {
			t.Errorf("Get(%d) = %+v, want %+v", created.ID, got, created)
		}
	})

	t.Run("get missing book", func(t *testing.T) {
		store := factory(t)

		var resourceNotFoundError *apperr.ResourceNotFoundError

		_, err := store.Get(999)
		// 唔好淨係 check err != nil — 要 check 係「啱嗰種」error
		if !errors.Is(err, apperr.ErrNotFound) {
			t.Fatalf("Get(999) error = %v, want ErrNotFound", err)
		}
		if !errors.As(err, &resourceNotFoundError) {
			t.Fatalf("Get(999) error = %v, want ResourceNotFoundError", err)
		}

		if resourceNotFoundError.Resource != "book" {
			t.Errorf("ResourceNotFoundError.Resource = %s, want book", resourceNotFoundError.Resource)
		}
		if resourceNotFoundError.ID != 999 {
			t.Errorf("ResourceNotFoundError.ID = %d, want 999", resourceNotFoundError.ID)
		}
	})

	t.Run("update returns stored book", func(t *testing.T) {
		store := factory(t)
		created, err := store.Create(Book{BaseBook: BaseBook{Title: "A", Author: "X", Year: 2020}})
		if err != nil {
			t.Fatalf("Create(Book{BaseBook: BaseBook{Title: A, Author: X, Year: 2020}}) returned unexpected error: %v", err)
		}

		updated, err := store.Update(created.ID, Book{BaseBook: BaseBook{Title: "B", Author: "Y", Year: 2021}})
		if err != nil {
			t.Fatalf("Update(%d, Book{BaseBook: BaseBook{Title: B, Author: Y, Year: 2021}}) returned unexpected error: %v", created.ID, err)
		}
		expected := Book{ID: created.ID, BaseBook: BaseBook{Title: "B", Author: "Y", Year: 2021}}

		got, err := store.Get(created.ID)
		if err != nil {
			t.Fatalf("Get(%d) returned unexpected error: %v", created.ID, err)
		}
		if updated != expected {
			t.Errorf("Update(%d) = %+v, want %+v", created.ID, updated, expected)
		}
		if got != expected {
			t.Errorf("Update(%d) = %+v, want %+v", created.ID, got, expected)
		}
	})

	t.Run("update book with id hijack", func(t *testing.T) {
		store := factory(t)
		created, err := store.Create(Book{BaseBook: BaseBook{Title: "A", Author: "X", Year: 2020}})
		if err != nil {
			t.Fatalf("Create(Book{BaseBook: BaseBook{Title: A, Author: X, Year: 2020}}) returned unexpected error: %v", err)
		}
		updated, err := store.Update(created.ID, Book{ID: 9999, BaseBook: BaseBook{Title: "B", Author: "Y", Year: 2021}})
		if err != nil {
			t.Fatalf("Update(%d, Book{ID: 9999, BaseBook: BaseBook{Title: B, Author: Y, Year: 2021}}) returned unexpected error: %v", created.ID, err)
		}
		if updated.ID == 9999 {
			t.Errorf("Update (id hijack) = %d, want %d", updated.ID, created.ID)
		}
	})

	t.Run("update missing book", func(t *testing.T) {
		store := factory(t)
		var resourceNotFoundError *apperr.ResourceNotFoundError
		_, err := store.Update(999, Book{BaseBook: BaseBook{Title: "B", Author: "Y", Year: 2021}})
		if !errors.Is(err, apperr.ErrNotFound) {
			t.Fatalf("Update(999, Book{BaseBook: BaseBook{Title: B, Author: Y, Year: 2021}}) error = %v, want ErrNotFound", err)
		}
		if !errors.As(err, &resourceNotFoundError) {
			t.Fatalf("Update(999, Book{BaseBook: BaseBook{Title: B, Author: Y, Year: 2021}}) error = %v, want ResourceNotFoundError", err)
		}
		if resourceNotFoundError.Resource != "book" {
			t.Errorf("ResourceNotFoundError.Resource = %s, want book", resourceNotFoundError.Resource)
		}
		if resourceNotFoundError.ID != 999 {
			t.Errorf("ResourceNotFoundError.ID = %d, want 999", resourceNotFoundError.ID)
		}
	})

	t.Run("delete", func(t *testing.T) {
		tests := []struct {
			name    string // subtest 名，會顯示喺 go test -v 度
			setup   func(t *testing.T, s Repository) int
			wantErr error
		}{
			{
				name: "existing book is deleted",
				setup: func(t *testing.T, s Repository) int {
					book, err := s.Create(Book{BaseBook: BaseBook{Title: "A", Author: "X"}})
					if err != nil {
						t.Fatalf("Create(Book{BaseBook: BaseBook{Title: A, Author: X}}) returned unexpected error: %v", err)
					}
					return book.ID
				},
				wantErr: nil,
			}, {
				name: "missing book returns ErrNotFound",
				setup: func(t *testing.T, s Repository) int {
					return 42
				},
				wantErr: apperr.ErrNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) { // t.Run = subtest，每個 case 獨立 pass/fail
				store := factory(t)
				id := tt.setup(t, store)
				err := store.Delete(id)
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Delete(%d) error = %v, want %v", id, err, tt.wantErr)
				}
				if tt.wantErr == nil {
					_, err := store.Get(id)
					if !errors.Is(err, apperr.ErrNotFound) {
						t.Errorf("Get(%d) error = %v, want ErrNotFound", id, err)
					}
				}
			})
		}
	})
}

func TestMemoryBookRepository(t *testing.T) {
	testBookRepository(t, func(t *testing.T) Repository {
		return NewMemoryStore()
	})
}

func TestGormBookRepository(t *testing.T) {
	testBookRepository(t, func(t *testing.T) Repository {
		dbPath := filepath.Join(t.TempDir(), "books.db")
		store, err := NewGormStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create book store: %v", err)
		}
		return store
	})
}
