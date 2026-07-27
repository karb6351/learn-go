package book

// ── 第五章骨架：GORM + SQLite ──────────────────────────────────
// GORM 對應你舊世界：Laravel 嘅 Eloquent / NestJS 嘅 TypeORM。
// 但冇 decorator、冇 magic — 全部係普通 function call，
// error 照舊行 (value, error) 雙回傳嗰套。

import (
	"context"
	"errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"playground/book/internal/apperr"
)

// GormStore 揸住一個 *gorm.DB（底層係 connection pool）。
// 留意：同 in-memory 版唔同，呢度冇 mutex —
// concurrency 交返俾 database 同 connection pool 處理。
type GormStore struct {
	db *gorm.DB
}

// NewGormStore 開 DB + AutoMigrate。
// AutoMigrate = TypeORM synchronize / Laravel migration 嘅懶人版：
// 照住 Book struct 開／執 table。Embedded 嘅 BaseBook fields 會攤平做
// columns（title / author / year）— 同 JSON marshal 攤平係同一個道理。
func NewGormStore(path string) (*GormStore, error) {
	db, err := gorm.Open(sqlite.Open(path+"?_busy_timeout=5000&_txlock=immediate"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Book{}); err != nil {
		return nil, err
	}
	return &GormStore{db: db}, nil
}

func (s *GormStore) Create(ctx context.Context, b Book) (Book, error) {
	if err := s.db.WithContext(ctx).Create(&b).Error; err != nil {
		return Book{}, err
	}
	return b, nil
}

func (s *GormStore) List(ctx context.Context, p ListParams) (Page, error) {
	var books []Book = make([]Book, 0)
	var total int64
	if err := s.db.Model(&Book{}).Count(&total).Error; err != nil {
		return Page{}, err
	}
	if total == 0 {
		return Page{Items: books, Total: 0}, nil
	}

	if err := s.db.Order("id ASC").Offset((p.Page - 1) * p.Limit).Limit(p.Limit).Find(&books).Error; err != nil {
		return Page{}, err
	}

	return Page{Items: books, Total: int(total)}, nil
}

func (s *GormStore) Get(ctx context.Context, id int) (Book, error) {
	var b Book
	if err := s.db.First(&b, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Book{}, &apperr.ResourceNotFoundError{Resource: "book", ID: id}
		}
		return Book{}, err
	}
	return b, nil
}

func (s *GormStore) Update(ctx context.Context, id int, b Book) (Book, error) {
	var existingBook Book

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 喺呢度做嘢...
		if err := tx.First(&existingBook, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &apperr.ResourceNotFoundError{Resource: "book", ID: id}
			}
			return err
		}
		existingBook.BaseBook = b.BaseBook
		if err := tx.Save(&existingBook).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return Book{}, err
	}

	return existingBook, nil
}

func (s *GormStore) Delete(ctx context.Context, id int) error {
	result := s.db.Delete(&Book{}, id)
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return &apperr.ResourceNotFoundError{Resource: "book", ID: id}
	}
	return nil
}
