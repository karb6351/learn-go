# Go + Gin 學習筆記：Book CRUD API

背景：有 NestJS / Spring MVC / Laravel 經驗，用呢個 project 學 Go + Gin。
學習方式：導師起骨架，核心設計決定同關鍵 code 自己寫，寫完 review。

## 進度總覽

- [x] 第一章：基本 CRUD（model / in-memory store / 5 個 endpoints）
- [x] 第二章：DTO 分離 + struct embedding（`BaseBook`）
- [x] 第三章：Testing（table-driven / httptest / red-green）
- [x] 第四章：Error handling middleware（translation switch + handler refactor，全綠 ✅）
- [x] 第 4.5 章：Generic error handling（id validation 統一 + contract tests + 完工註解 ✅）
- [ ] **第五章（未揀）：GORM + SQLite / Project layout 拆 package / 其他 middleware（auth / logging / rate limit）** ⬅ 下次由呢度開始

## 下次開波

冇未完成任務 — 淨係要揀第五章方向。4.5 章最終形態：

- 三個 endpoint 統一用 `BookParam`（`uri:"id" binding:"min=1"`）+ `ShouldBindUri`
- Handler 完全唔識 HTTP status — 一律 `c.Error(err)` + `c.Abort()` 交貨（架構 A 徹底版）
- Middleware translation switch 四款：`ValidationErrors`→422 Laravel style / `ErrBookNotFound`→404 / `*strconv.NumError`→400 / 兜底→500（警報器）
- id contract 已鎖：`/books/abc`→400、`/books/0`→422、`/books/999`→404、DELETE 成功→204 冇 body
- `APIError` 實驗完剷咗（YAGNI）— 曾經行過「handler 譯 domain error 做 APIError」嘅架構 B，最後揀返 A；兩個架構嘅 trade-off 見速查表

## File 結構

| File | 角色 | 對應舊世界 |
|---|---|---|
| `main.go` | wiring + `setupRouter()`（抽出嚟先測到） | NestJS module / TestingModule |
| `book.go` | `BaseBook`/`Book` model + in-memory store (mutex) | Entity + Repository |
| `handlers.go` | HTTP handlers + `BookInput` DTO | Controller + DTO |
| `middleware.go` | `ErrorHandler()` — 全 API error 出口（translation switch 完工，有完工版教學註解） | Laravel Handler::render / NestJS exception filter |
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
- Is/As 分界線（4.5 章考核已過）：睇返個 error 當初**設計成一粒定一款** — 全 program 共用一粒值（sentinel，唔孭 data）→ `Is` 認人；一個 type 隨時 new、孭住 fields 要提取（`ValidationErrors`、`*strconv.NumError`、GORM 嘅 `*PgError`）→ `As` 認款。GORM 嘅 `ErrRecordNotFound` 就係官方 sentinel
- `errors.As` 個 target 要 `&`：As 做嘅係將 chain 入面嗰粒指針**抄落你個變數**（純指針賦值，冇 new 冇 copy struct）；pass by value 之下想 function 改你個變數就要交 address — 同 `json.Unmarshal` 要 `&` 同一個道理。變數 type 跟 chain 嗰粒原樣（`*strconv.NumError` 就宣告 `*strconv.NumError`），傳入時先加 `&`
- Gin URI binding 塞唔入 `int` 出嘅係 stdlib `*strconv.NumError`（`.Num` 孭住原始輸入）— middleware 加 `As` 分支譯 400，唔好俾佢跌落 500 兜底
- 400 vs 422 嘅分工（我哋揀嘅 contract）：parse 唔到（malformed，`abc`）→ 400；parse 到但犯 rule（`0` 撞 `min=1`）→ 422 Laravel style
- Handler 徹底唔識 status code（架構 A）先消滅到 copy-paste；架構 B（handler 譯 domain→`APIError`）都合法但要收 per-handler 重複嘅代價 — 試過，倒返轉頭

**Testing（4.5 章新增）**
- 「碰巧 pass」再現：copy-paste test 冇改 method，DELETE test 全程打緊 GET 照樣全綠 — pass 唔代表 test 緊你以為嗰樣嘢；新 test case 要親眼見過佢紅（red-green 嘅 red 係證據）
- 204 No Content 嘅 contract 唔只係個 status — 「冇 body」都要 assert
- Error message 係 public contract：自己寫、自己控制，唔好 pipe library 嘅 `err.Error()` 出街（`strconv.ParseInt: parsing...` 漏內部實現）；但 test 只 assert「邊個 field 有錯」唔 assert 措辭

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

- NestJS `throw new NotFoundException()` 可以喺深層直接「飛」上 exception filter — 背後係咩語言機制？Go 明知有（近親：panic/recover）特登唔用嚟做日常 error handling — 嫌佢邊樣嘢？（提示：睇 function signature 知唔知佢會唔會中途走佬）

- `json.Unmarshal(data, v)` 用 `any` 所以冇 compile-time 保護 — generics 版會點設計？
- `c.Errors` 係 slice（一個 request 可掛多個 error），我哋淨處理 `.Last()` — 幾時會想處理晒？
