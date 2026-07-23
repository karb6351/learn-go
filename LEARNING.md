# Go + Gin 學習筆記：Book CRUD API

背景：有 NestJS / Spring MVC / Laravel 經驗，用呢個 project 學 Go + Gin。
學習方式：導師起骨架，核心設計決定同關鍵 code 自己寫，寫完 review。

## 進度總覽

- [x] 第一章：基本 CRUD（model / in-memory store / 5 個 endpoints）
- [x] 第二章：DTO 分離 + struct embedding（`BaseBook`）
- [x] 第三章：Testing（table-driven / httptest / red-green）
- [x] 第四章：Error handling middleware（translation switch + handler refactor，全綠 ✅）
- [x] 第 4.5 章：Generic error handling（id validation 統一 + contract tests + 完工註解 ✅）
- [x] 第 4.9 章（bonus）：Error 家族 — generic `ErrNotFound` + `%w` + `ResourceNotFoundError`/`Unwrap` 橋樑 ✅
- [ ] **第五章：GORM + SQLite（進行中 — 任務睇下面）** ⬅ 而家喺呢度
- [ ] 第六章（未揀）：Project layout 拆 package / 其他 middleware（auth / logging / rate limit）

## ⚠️ 第五章任務（任務 1–3 完成 ✅，剩任務 4）

已完成：`BookRepository` interface（consumer 側，五 method 全帶 error 位）、`GormBookStore` 五個 CRUD（layering 乾淨、not-found 語義同 mem store 一致）、`main.go` 接 GORM store、persistence 用 `sqlite3` 直讀 `books.db` 驗證咗（AutoMigrate 自動起 table，embedded `BaseBook` 攤平做 columns）。

### 任務 4（bonus，下次做）：test suite 對兩個 implementation 通用

而家 store tests（`book_test.go`）寫死 `NewBookStore()`，GORM store 零 unit test — Delete 嘅 `RowsAffected` 陷阱嗰類蟲，handler test（mem store 陪練）係唔會出聲嘅。諗下點樣用 table（名 + constructor function 嘅 list）令同一套 test 兩款 store 都行一次。提示：GORM 版 test 要諗 DB file 點樣每個 test 隔離（`t.TempDir()` 係你朋友）。

驗收照舊：`go vet ./... && go test ./...` 全綠。

## File 結構

| File | 角色 | 對應舊世界 |
|---|---|---|
| `main.go` | wiring + `setupRouter()`（抽出嚟先測到） | NestJS module / TestingModule |
| `book.go` | `BaseBook`/`Book` model + in-memory store (mutex) | Entity + Repository |
| `handlers.go` | HTTP handlers + `BookInput` DTO | Controller + DTO |
| `middleware.go` | `ErrorHandler()` — 全 API error 出口（translation switch 完工，有完工版教學註解） | Laravel Handler::render / NestJS exception filter |
| `gorm_store.go` | `GormBookStore` — GORM + SQLite store（骨架已起，CRUD 施工中） | Eloquent model / TypeORM repository |
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
- `%w` wrapping：`fmt.Errorf("book %d: %w", id, ErrNotFound)` 起一個 `Unwrap() error` 節點 → error chain 係字面意義嘅 linked list；`Is`/`As` 個 loop 係「比對 → 剝一層 → 再比對」行到底（睇過 stdlib `errors/wrap.go` 原文）
- Error 家族靠 **Unwrap chain** 唔係 embedding：`ResourceNotFoundError`（type，孭 `Resource`/`ID`）+ `Unwrap() → ErrNotFound`（sentinel）= `As` 攞料、`Is` 認族兩樣都得；舊 `Is` test 一行不改照綠就係橋樑生效嘅證據。GORM 都係咁玩
- 命名 convention：sentinel **值**用 `Err` 頭（`ErrNotFound`）、error **type** 用 `Error` 尾（`ResourceNotFoundError`、`*strconv.NumError`）— 個名已經話你知用 `Is` 定 `As`
- `As` 完就用抽出嚟嗰粒變數，唔好再掂 `err` — `err.Error()` 係成條 chain 嘅串連版，第日多包一層就漏晒出街
- Middleware if/else chain 唔使 strategy pattern：分支數目跟 error **類別**增長（個位數封頂），唔係跟 resource／endpoint；「爆」嘅憂慮應該用 generic sentinel + wrap 喺源頭解決。Registry 要等到拆 package、各 package 自行註冊嗰陣先有真需求（rule of three / 痛咗先抽象）
- Rename 紀律：grep 出完整名單（**包埋 comment**）→ 逐個銷 → 再 grep 驗屍；唔好憑記憶改

**Interface + GORM（第五章新增）**
- Consumer-defined interface：食客開單 — method 名單以 handler 實際用到為準，定義擺 consumer file（同 Java/NestJS provider-side 調轉）；grep `h.store\.` 就係張單
- **Accept interfaces, return structs**：parameter 收 interface（`setupRouter(store BookRepository)`），constructor 交 concrete（`NewBookStore() *BookStore`）— return interface 會迫 caller 用 type assertion（`x.(T)`）剝殼攞返自己嘅嘢，兼且令 implementation 側反向依賴 consumer 嘅 interface（拆 package 時會變 import cycle）
- Structural typing 接駁實感：`NewBookHandler(store)` 一句，`*BookStore` 冇宣佈過 implements 就塞得入 `BookRepository` — method set 齊就自動收貨
- Interface signature 要遷就「識失敗嘅 implementation」：mem store 唔會 fail 但 GORM 會 → 五個 method 全帶 error 位；改 signature 之後 **compiler error list = todo list**，逐個清（TS 都係咁，PHP 就要等 runtime 炸）
- GORM 對應：`AutoMigrate` ≈ TypeORM synchronize（dev 玩具）；`First` 搵唔到先出 `gorm.ErrRecordNotFound`（sentinel → `errors.Is`）；**`Delete` 唔存在嘅 id 唔會 error** — `result.Error == nil` + `RowsAffected == 0`，兩個獨立兄弟 check 順序行，唔好 nested / `||` 縮埋
- 每個 db 操作都要摸 `.Error`（`Create`/`Find`/`Save` 全部識失敗）— GORM 唔 throw，唔檢查就係 silent failure
- 多條件 error handling 寫完用 **truth table 過一次**（每個情況 × 應該回乜 × 實際回乜）— 靠倒模句子執 code 係會將啲括號執錯位
- `main()` 起場失敗用 `log.Fatal(err)` — server 未起，冇 client 要靚 response，全 project 唯一「直接死」嘅位
- SQLite = 一個 file 就係成個 DB：`sqlite3 books.db "select..."` 直讀，繞過 server 驗 persistence 係最狠嘅證據

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
