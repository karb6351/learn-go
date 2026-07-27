package book

import (
	"context"
	"errors"
	"path/filepath"
	"sync"
	"testing"

	"playground/book/internal/apperr"
)

// 呢個 test 專門針對 GormStore.Update 嘅 TOCTOU 罅（First 同 Save 之間冇嘢㩒住）。
// MemoryStore 冇呢個病（成段 check+write 喺一把 exclusive Lock 罩住），
// 所以佢唔屬於 Repository contract suite — 係一個 implementation-specific 嘅 bug test。
//
// 策略：概率轟炸。條罅得幾微秒闊，冇辦法喺 test 度「暫停」GORM，
// 唯有 loop N round，每 round 兩條 goroutine 同時開火，博佢喺某一 round 撞入條罅。
//
// Invariant：「Delete return nil（成功）⇒ 之後 Get 必須 not found」
// 破咗 = 殭屍書現形。
func TestGormStoreUpdateDeleteRace(t *testing.T) {
	const rounds = 100

	dbPath := filepath.Join(t.TempDir(), "race.db")
	store, err := NewGormStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()

	for i := 0; i < rounds; i++ {
		created, err := store.Create(ctx, Book{BaseBook: BaseBook{Title: "victim", Author: "A", Year: 2020}})
		if err != nil {
			t.Fatalf("round %d: Create failed: %v", i, err)
		}

		var wg sync.WaitGroup
		wg.Add(2)

		// 每條 goroutine 寫自己嗰個變數，主線等 wg.Wait() 完先讀 —
		// WaitGroup 提供咗 happens-before 保證，所以咁樣冇 data race
		var deleteErr error

		go func() {
			defer wg.Done()
			// Update 嘅 return 唔使理 — 佢輸咗場 race 回 not found 係正常結局，
			// 我哋唔係測佢，係利用佢做攻擊者。
			store.Update(ctx, created.ID, Book{BaseBook: BaseBook{Title: "attacker", Author: "B", Year: 2021}})
		}()

		go func() {
			defer wg.Done()
			deleteErr = store.Delete(ctx, created.ID)
		}()

		wg.Wait()

		if deleteErr == nil {
			_, err := store.Get(ctx, created.ID)
			if err == nil {
				t.Fatalf("round %d: zombie — Get succeeded after successful Delete", i)
			}
			if !errors.Is(err, apperr.ErrNotFound) {
				t.Fatalf("round %d: Get returned not found: %v", i, err.Error())
			}
		}
	}
}
