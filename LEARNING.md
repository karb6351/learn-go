# Go + Gin 學習筆記：Book CRUD API

背景：有 NestJS / Spring MVC / Laravel 經驗，用呢個 project 學 Go + Gin。
學習方式：導師起骨架，核心設計決定同關鍵 code 自己寫，寫完 review。

## 進度總覽

- [x] 第一章：基本 CRUD（model / in-memory store / 5 個 endpoints）
- [x] 第二章：DTO 分離 + struct embedding（`BaseBook`）
- [x] 第三章：Testing（table-driven / httptest / red-green）
- [ ] **第四章：Error handling middleware（進行中 — 有未完成任務，睇下面）** ⬅ 而家喺呢度
- [ ] 第五章（未開始，任揀）：GORM + SQLite / Project layout / 其他 middleware

## ⚠️ 未完成任務（返嚟由呢度開始）

而家 test suite 係 **刻意紅色** 狀態（red-green-refactor 嘅 red）：

```
$ go test ./...
FAIL: TestCreateBookValidation          status = 500, want 422
FAIL: TestUpdateBookHandler/missing_author   status = 400, want 422
```

目標：實現 Laravel-style validation error contract，令兩個 test 轉綠。

### 任務 1：寫 `middleware.go` 入面個 translation switch

位置：`middleware.go` 嘅 TODO。目標 contract（Laravel style，validation 用 422）：

```json
{
  "message": "The given data was invalid.",
  "errors": { "author": ["The author field is required."] }
}
```

- validation error 用 `errors.As` 提取 `validator.ValidationErrors`（import `github.com/go-playground/validator/v10`），逐條 `fe.Field()`（記得轉細楷）+ `fe.Tag()`
- `ErrBookNotFound` 用 `errors.Is` → 404 `{"message":"book not found"}`
- 兜底 → 500 `{"message":"internal error"}`
- 記得剷走 placeholder 嗰兩行

### 任務 2：Refactor `UpdateBook`

跟 `CreateBook`（已示範）嘅 pattern：error 唔再自己 `c.JSON`，改用 `c.Error(err)` + `c.Abort()` + `return`，交俾 middleware translate。之後可以順手 refactor 埋 `GetBook` / `DeleteBook`。

驗收：`go vet ./... && go test ./...` 全綠。

## File 結構

| File | 角色 | 對應舊世界 |
|---|---|---|
| `main.go` | wiring + `setupRouter()`（抽出嚟先測到） | NestJS module / TestingModule |
| `book.go` | `BaseBook`/`Book` model + in-memory store (mutex) | Entity + Repository |
| `handlers.go` | HTTP handlers + `BookInput` DTO | Controller + DTO |
| `middleware.go` | `ErrorHandler()` — 全 API error 出口（施工中） | Laravel Handler::render / NestJS exception filter |
| `*_test.go` | store unit tests + httptest handler tests | PHPUnit / Jest + supertest |

跑 server：`go run .`（port 8089）；跑 test：`go test -v ./...`

## 已學概念速查（每個都親手撞過／答過問題）

**語言基礎**
- `make`：map/slice/channel 專用初始化；nil map 讀得寫唔得（寫 = panic）
- Zero value 哲學:「唔存在」回 zero value 唔係 exception；comma-ok（`v, ok := m[k]`）分辨兩者
- `defer`：登記喺 function 離開嗰刻執行（= try/finally）；Lock 完下一行即 defer Unlock 係鐵律
- Value semantics：全部 pass by value；`Create(b Book)` 改 copy 再 `return b`（value in, value out）
- Pointer：`.` 自動 deref（冇 `->`）；`&` 只出現喺 caller / composite literal，parameter 宣告用 `*T`
- Return `&BookStore{...}` 冇 dangling 問題 — escape analysis 自動搬上 heap
- `sync.RWMutex`：goroutine 共享 state 必須上鎖；mutex 唔可以 copy（所以 store 用 pointer receiver）

**Error handling**
- 冇 exception：`(value, error)` 雙回傳 + sentinel error（`ErrBookNotFound`）
- `errors.Is`（係咪呢粒）vs `errors.As`（係咪呢類，提取出嚟用）
- Handler 寫完 error response 記得 `return`（Gin 唔會自動停）
- Panic 留俾 programmer bug；`gin.Default()` 有 Recovery middleware 兜底

**設計**
- DTO 分離：`BookInput` 冇 ID field → id 騎劫「make invalid states unrepresentable」
- Struct embedding ≠ inheritance：has-a 唔係 is-a，冇 polymorphism；JSON 自動攤平；validation 穿透
- `BaseBook` 對稱 embedding 好過 `Book` embed `BookInput`（依賴方向問題）
- Middleware onion model：`c.Next()` 前 = request phase，後 = response phase；error translation 放後段

**Testing**
- Table-driven + `t.Run` subtest；`t.Errorf`（繼續）vs `t.Fatalf`（即停）；`t.Helper()`
- httptest 全 in-process，唔使開 port
- 每個 subtest 開新 router + store（教訓：seed 錯咗另一個 store，「碰巧 pass」嘅 404）
- Test 承諾唔好 test 意外：assert contract（status、field 名），唔好 assert 措辭（change detector）
- `go vet`：捉 mutex copy、Unmarshal 冇 `&` 等「compile 到但九成錯」嘅嘢，CI 必跑

## 未答完嘅思考題

- `json.Unmarshal(data, v)` 用 `any` 所以冇 compile-time 保護 — generics 版會點設計？
- `c.Errors` 係 slice（一個 request 可掛多個 error），我哋淨處理 `.Last()` — 幾時會想處理晒？
