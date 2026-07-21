package main

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
	"testing"
)

func TestCreateAssignsSequentialIDs(t *testing.T) {
	store := NewBookStore() // 每個 test 開個新 store — test 之間零共享

	b1 := store.Create(Book{BaseBook: BaseBook{Title: "A", Author: "X", Year: 2020}})
	b2 := store.Create(Book{BaseBook: BaseBook{Title: "B", Author: "Y", Year: 2021}})

	if b1.ID != 1 {
		t.Errorf("first book: got ID %d, want 1", b1.ID)
	}
	if b2.ID != 2 {
		t.Errorf("second book: got ID %d, want 2", b2.ID)
	}
}

func TestGetReturnsStoredBook(t *testing.T) {
	store := NewBookStore()
	created := store.Create(Book{BaseBook: BaseBook{Title: "A", Author: "X", Year: 2020}})

	got, err := store.Get(created.ID)
	if err != nil {
		t.Fatalf("Get(%d) returned unexpected error: %v", created.ID, err)
	}
	// struct 冇 pointer field 嘅話可以直接 == 比較
	if got != created {
		t.Errorf("Get(%d) = %+v, want %+v", created.ID, got, created)
	}
}

func TestGetNotFound(t *testing.T) {
	store := NewBookStore()

	_, err := store.Get(999)
	// 唔好淨係 check err != nil — 要 check 係「啱嗰種」error
	if !errors.Is(err, ErrBookNotFound) {
		t.Errorf("Get(999) error = %v, want ErrBookNotFound", err)
	}
}

// Table-driven test — Go 社群最核心嘅 test pattern。
// 同一段測試邏輯，用一個 slice 餵唔同 case，
// 相當於 Jest 嘅 it.each / JUnit 嘅 @ParameterizedTest。
func TestDelete(t *testing.T) {
	tests := []struct {
		name    string // subtest 名，會顯示喺 go test -v 度
		setup   func(s *BookStore)
		id      int
		wantErr error
	}{
		{
			name:    "existing book is deleted",
			setup:   func(s *BookStore) { s.Create(Book{BaseBook: BaseBook{Title: "A", Author: "X"}}) },
			id:      1,
			wantErr: nil,
		},
		{
			name:    "missing book returns ErrBookNotFound",
			setup:   func(s *BookStore) {},
			id:      42,
			wantErr: ErrBookNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { // t.Run = subtest，每個 case 獨立 pass/fail
			store := NewBookStore()
			tt.setup(store)

			err := store.Delete(tt.id)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Delete(%d) error = %v, want %v", tt.id, err, tt.wantErr)
			}
		})
	}
}
