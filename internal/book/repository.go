package book

import "context"

// ListParams 係 List 嘅入參 struct。
// 用 struct 唔用 (page, limit int)：第日加 sort/filter 係加 field，
// 唔使改 signature（= 唔使全 project domino）。
//
// Contract：呢度假設 params 已經有效（Page ≥ 1、1 ≤ Limit ≤ 100）—
// 驗證係 HTTP 層（DTO binding）嘅責任，repository 唔重複做。
type ListParams struct {
	Page  int
	Limit int
}

// Page 係 List 嘅出參：一頁嘅嘢 + 全表總數。
// Total 永遠係全表數量，唔係本頁數量 — client 靠佢計總頁數。
type Page struct {
	Items []Book
	Total int
}

// Contract 補充（contract suite 會逼供呢啲行為）：
//   - page 超晒範圍 → Items 空、Total 照回全表數（zero value 哲學，唔係 error）
//   - 順序照舊承諾：按 ID 升序
type Repository interface {
	Create(ctx context.Context, b Book) (Book, error)
	List(ctx context.Context, p ListParams) (Page, error)
	Get(ctx context.Context, id int) (Book, error)
	Update(ctx context.Context, id int, b Book) (Book, error)
	Delete(ctx context.Context, id int) error
}
