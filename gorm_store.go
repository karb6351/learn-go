package main

// ── 第五章骨架：GORM + SQLite ──────────────────────────────────
// GORM 對應你舊世界：Laravel 嘅 Eloquent / NestJS 嘅 TypeORM。
// 但冇 decorator、冇 magic — 全部係普通 function call，
// error 照舊行 (value, error) 雙回傳嗰套。

import (
	"errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// GormBookStore 揸住一個 *gorm.DB（底層係 connection pool）。
// 留意：同 in-memory 版唔同，呢度冇 mutex —
// concurrency 交返俾 database 同 connection pool 處理。
type GormBookStore struct {
	db *gorm.DB
}

// NewGormBookStore 開 DB + AutoMigrate。
// AutoMigrate = TypeORM synchronize / Laravel migration 嘅懶人版：
// 照住 Book struct 開／執 table。Embedded 嘅 BaseBook fields 會攤平做
// columns（title / author / year）— 同 JSON marshal 攤平係同一個道理。
func NewGormBookStore(path string) (*GormBookStore, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Book{}); err != nil {
		return nil, err
	}
	return &GormBookStore{db: db}, nil
}

func (s *GormBookStore) Create(b Book) (Book, error) {
	if err := s.db.Create(&b).Error; err != nil {
		return Book{}, err
	}
	return b, nil
}

func (s *GormBookStore) List() ([]Book, error) {
	var books []Book
	if err := s.db.Order("id ASC").Find(&books).Error; err != nil {
		return nil, err
	}
	return books, nil
}

func (s *GormBookStore) Get(id int) (Book, error) {
	var b Book
	if err := s.db.First(&b, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Book{}, &ResourceNotFoundError{Resource: "book", ID: id}
		}
		return Book{}, err
	}
	return b, nil
}

func (s *GormBookStore) Update(id int, b Book) (Book, error) {
	var existingBook Book
	if err := s.db.First(&existingBook, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Book{}, &ResourceNotFoundError{Resource: "book", ID: id}
		}
		return Book{}, err
	}
	existingBook.BaseBook = b.BaseBook
	if err := s.db.Save(&existingBook).Error; err != nil {
		return Book{}, err
	}
	return existingBook, nil
}

func (s *GormBookStore) Delete(id int) error {
	result := s.db.Delete(&Book{}, id)
	if err := result.Error; err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return &ResourceNotFoundError{Resource: "book", ID: id}
	}
	return nil
}
