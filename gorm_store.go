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

// TODO(你嚟寫): 五個 CRUD method，簽名同 in-memory BookStore 嗰五個一模一樣。
//
// GORM v2 API 提示：
//   - Create:  s.db.Create(&b) — GORM 會將新 ID 寫返入 b.ID（諗下點解要 &）
//   - List:    var books []Book → s.db.Find(&books)
//   - Get:     var b Book → s.db.First(&b, id) — 搵唔到嗰陣 .Error 係 gorm.ErrRecordNotFound
//   - Update:  先 First 確認存在，再 s.db.Save(&b)（記得保持返 in-memory 版嘅語義：id 以 URL 為準）
//   - Delete:  result := s.db.Delete(&Book{}, id) — Delete 唔存在嘅 id 唔會出 error！
//              要靠 result.RowsAffected 自己判斷（諗下：呢個似唔似 map 嘅 zero value 哲學？）
//   - 每個 db 操作都要檢查 .Error — GORM 唔會 throw，唔檢查就係 silent failure
//
// ⚠️ 關鍵設計位（layering rule）：
//   gorm.ErrRecordNotFound 係 GORM 嘅詞彙，唔准漏出呢個 file。
//   Caller（handler / middleware / test）只識我哋自己嘅 error 家族
//   （ErrNotFound / ResourceNotFoundError）。檢查用邊件工具、translate 做乜，
//   你啱啱先操練完 — GORM 嗰粒係 sentinel 定係 type，你上堂答過 😉
//   其他 error（DB 死咗等）照原樣回傳 — middleware 兜底 500 會接住。

func (s *GormBookStore) Create(b Book) (Book, error) {
	if err := s.db.Create(&b).Error; err != nil {
		return Book{}, err
	}
	return b, nil
}

func (s *GormBookStore) List() ([]Book, error) {
	var books []Book
	if err := s.db.Find(&books).Error; err != nil {
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
