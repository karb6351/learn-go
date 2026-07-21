# Go + Gin 學習筆記：Book CRUD API

背景：有 NestJS / Spring MVC / Laravel 經驗，用呢個 project 學 Go + Gin。
學習方式：導師起骨架，核心設計決定同關鍵 code 自己寫，寫完 review。

## 進度總覽

- [x] 第一章：基本 CRUD（model / in-memory store / 5 個 endpoints）
- [x] 第二章：DTO 分離 + struct embedding（`BaseBook`）
- [x] 第三章：Testing（table-driven / httptest / red-green）
- [x] 第四章：Error handling middleware（translation switch + handler refactor，全綠 ✅）
- [ ] **第 4.5 章：Generic error handling（進行中 — 收尾任務睇下面）** ⬅ 而家喺呢度
- [ ] 第五章（未開始，任揀）：GORM + SQLite / Project layout / 其他 middleware

## ⚠️ 未完成任務（返嚟由呢度開始）

第四章已收貨。4.5 章已做咗：`errorTagMessages` map（`{field}`/`{value}` placeholder + comma-ok 兜底）、`APIError` type（`api_error.go`）+ middleware `errors.As` 分支、`BookParam` + `ShouldBindUri`（PUT/DELETE）。中途試過 `NotFoundError` embed `APIError` + store 回 HTTP-aware error，review 後倒返轉頭（layering + embedding≠inheritance，睇速查表）。

收尾任務（test 而家全綠，但有已知行為分歧未鎖 contract）：

### 任務 1：統一 id validation

而家 `GET /books/0` → 404（`GetBook` 用 `strconv.Atoi`）但 `DELETE /books/0` → 400（`BookParam` 嘅 `required` 當 0 = 冇提供，zero value 誤殺）。二揀一統一晒三個 endpoint；如果揀 `BookParam`，`required` 應改做 `min=1`，並改返句 error message（而家 0 嗰句 "id must be an integer" 係大話）。

### 任務 2：加 test 鎖住 id contract

`/books/abc`（400）同 `/books/0`（跟任務 1 嘅決定）而家零 coverage — 行為分歧就係咁靜靜雞溜入嚟。落 table-driven cases。

### 任務 3（可選）：改寫 `middleware.go` 嘅 TODO 註解做完工版教學註解

驗收：`go vet ./... && go test ./...` 全綠。

## File 結構

| File | 角色 | 對應舊世界 |
|---|---|---|
| `main.go` | wiring + `setupRouter()`（抽出嚟先測到） | NestJS module / TestingModule |
| `book.go` | `BaseBook`/`Book` model + in-memory store (mutex) | Entity + Repository |
| `handlers.go` | HTTP handlers + `BookInput` DTO | Controller + DTO |
| `middleware.go` | `ErrorHandler()` — 全 API error 出口（translation switch 完工） | Laravel Handler::render / NestJS exception filter |
| `api_error.go` | `APIError`（status + message，HTTP 層先出世） | NestJS HttpException |
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
- `errors.Is`（係咪呢粒 = 身份 `==`）vs `errors.As`（係咪呢類 = concrete type 比對，再塞落你個變數所以要 `&`）
- `errors.Is` 一定要同「共用嘅 sentinel 變數」比 — 即場 new 一粒內容一樣嘅嚟比永遠 false（教訓：`errors.Is(err, &NotFoundError{...})`）
- 任何 named type 有 `Error() string` 就係 error（structural typing）— `ValidationErrors` 本身係 `[]FieldError` slice；對比 NestJS/Java 要明文 `extends`（nominal）
- Validator 對每個 field fail-fast（第一個 fail 嘅 tag 就停），但 field 之間唔互相截停；`map[string][]string` 個 shape 係跟 contract（Laravel）唔係跟 validator 實現
- 自定義 error type（`APIError` 揸 status+message）= Go 版 HttpException；但 embedding ≠ inheritance — `*NotFoundError` embed `APIError` 都唔會被 `errors.As(&apiErr)` 捉到，Go 砌 error 家族用 Unwrap chain 唔係 embedding（未實作，第日玩）
- Layering：repository 唔應該識 HTTP — store 回 domain error（sentinel），HTTP status 翻譯係 middleware 嘅事（= Eloquent throw ModelNotFoundException，Handler::render 譯 404）
- `binding:"required"` 對 int 嘅「冇提供」= zero value → `0` 被誤殺（zero value 哲學再現）；binding 世界嘅 comma-ok = pointer field（`*int`，nil = absent）
- Map 讀 missing key 回 zero value（`""`）唔會出錯 — `errorTagMessages[tag]` 冇兜底就會回空 message，comma-ok 救返
- Test 全綠 ≠ 冇問題：只證明「有承諾嘅嘢冇爛」，coverage gap 入面嘅行為可以靜靜雞變（`APIError` 一度全部跌落 500 都冇 test 出聲）

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

## 開波前復習（學生自己 flag 咗）

- **`errors.Is` vs `errors.As` 未係 100% 明** — 復習方向：Is = 身份（`==` 比對同一粒 sentinel 變數）；As = 類型（沿 chain 搵 concrete type，搵到塞落你個變數所以要 `&`）。可以攞 project 入面三個真實用例行一次：`errors.Is(err, ErrBookNotFound)`、`errors.As(err, &ve)`、`errors.As(err, &apiError)`，逐個問「呢度點解用呢個唔用另一個」

## 未答完嘅思考題

- NestJS `throw new NotFoundException()` 可以喺深層直接「飛」上 exception filter — 背後係咩語言機制？Go 明知有（近親：panic/recover）特登唔用嚟做日常 error handling — 嫌佢邊樣嘢？（提示：睇 function signature 知唔知佢會唔會中途走佬）

- `json.Unmarshal(data, v)` 用 `any` 所以冇 compile-time 保護 — generics 版會點設計？
- `c.Errors` 係 slice（一個 request 可掛多個 error），我哋淨處理 `.Last()` — 幾時會想處理晒？
