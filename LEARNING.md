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
- [x] 第五章：GORM + SQLite（repository interface / persistence / shared contract suite，全綠 ✅）
- [x] 第六章：Project layout — `cmd/api` + `internal/{book,apperr,api}`，依賴一條直線 ✅
- [x] 第七章：Pagination — `ListParams`/`Page` contract、offset 分頁、query binding + defaults、Laravel envelope ✅
- [x] 第八章：Auth middleware — Bearer key、route group 局部上鎖、request phase、`errUnauthenticated` → 401 ✅
- [x] 第九章：Graceful shutdown — goroutine/channel/context 初體驗、`http.Server` + `Shutdown(ctx)`、`/slow` 實驗實證 ✅
- [ ] **第十章（未揀）：logging / rate limit（channel 再深造）/ context 貫穿 request（`c.Request.Context()` 落到 GORM）** ⬅ 下一步

## ✅ 第五章任務（任務 1–4 完成）

已完成：`BookRepository` interface（consumer 側，五 method 全帶 error 位）、`GormBookStore` 五個 CRUD（layering 乾淨、not-found 語義同 mem store 一致）、`main.go` 接 GORM store、persistence 用 `sqlite3` 直讀 `books.db` 驗證咗（AutoMigrate 自動起 table，embedded `BaseBook` 攤平做 columns），以及 memory/GORM 共用嘅 repository contract suite。

### 任務 4（bonus，完成 ✅）：一套 test 行兩款 store

採用 interface compliance suite 形態：`testBookRepository(t, factory)` 定義共用 contract，`TestMemoryBookRepository` / `TestGormBookRepository` 各自提供 `storeFactory`。Factory 收 `*testing.T`，令每個 GORM subtest 都可以用 `t.TempDir()` 建立獨立 SQLite DB；每個 case 重新 call factory，test 之間零共享 state。

Create contract 唔綁死 ID 必須由 1、2 開始，只要求 ID 大過 0 而且唔重複。Delete setup 回傳實際 `created.ID`，成功後再 `Get` 驗證資料真係消失；setup 明文接收最內層 subtest 嘅 `t`，避免 `Fatalf` 錯用 parent test。

共用 suite 五個 method 齊腳 ✅：Create、Get（含 not-found + As 攞 field）、List（空 + 多本 + 順序 contract）、Update（expected 兩炮 + postcondition Get + id hijack）、Delete（含 postcondition）。List 順序升做 contract（按 ID 升序）：mem store `sort.Slice`、GORM `Order("id ASC")` 明文承諾。

驗收照舊：`go vet ./... && go test ./...` 全綠。

## File 結構（第六章拆完）

| 位置 | Package | 角色 | 對應舊世界 |
|---|---|---|---|
| `cmd/api/main.go` | `main` | 貧血 wiring：開 store → `api.SetupRouter` → `Run` | NestJS bootstrap |
| `internal/book/` | `book` | `Book` model、`Repository` interface、`MemoryStore`、`GormStore`、contract suite | domain module（entity + repository） |
| `internal/apperr/` | `apperr` | `ErrNotFound` + `ResourceNotFoundError` error 家族（跨 domain 基建） | HttpException 家族嘅 domain 版 |
| `internal/api/` | `api` | handlers、`BookInput` DTO、`ErrorHandler` middleware、`SetupRouter`、httptest tests | Controller + exception filter |

依賴方向：`main → api → book → apperr`（一條直線，冇圈）

跑 server：`go run ./cmd/api`（port 8089）；跑 test：`go test -v ./...`

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

**Goroutine / Channel / Shutdown（第九章新增）**
- Goroutine = 平工人：`go f()` 即開即走；幾 KB 起跳開百萬條都得，M:N 織落真 thread；**main 一返全員陪葬**；Gin 每個 request 一條（∴ 第一章個 map 要 mutex）
- 對 Java：`go` 後面個 closure ≈ Runnable（份工），goroutine ≈ Thread（工人）但輕百倍，runtime scheduler ≈ ExecutorService（語言送嘅）；Java 21 virtual threads（Loom）= Java 版 goroutine
- Concurrency（結構：同時處理緊）≠ Parallelism（執行：呢一刻真係並行，睇 core 數 scheduler 話事）— **寫 code 必須當 parallel**，interleave 係 bonus 唔係依靠（Node 承諾單線程所以唔使 lock，Go 冇承諾）
- Channel：有型別管道，兩邊 block = 通訊兼同步一體；`<-ch` = Go 版 `await`（真瞓，零 CPU）；buffer `1` = 一格信箱 — `signal.Notify` 派信「放低就走」，冇格兼未有人接 = 封信跌落坑渠
- `signal.Notify` 係 **subscription**（"Notify *me*"）= `process.on('SIGTERM')`；Go stdlib 異步事件一律 channel 交貨（`time.After`、`ctx.Done()`）— Node push callback，Go pull channel
- **訂閱要喺瞓覺之前**：Notify 收埋喺 `ErrServerClosed` 分支入面 = 循環等待，機器從未通電 — SIGTERM 到門口冇人訂閱，默認即殺，graceful 全套變裝飾
- `r.Run` 係 `http.ListenAndServe(addr, engine)` 語法糖；`gin.Engine` 本身就係 `http.Handler`（httptest 嘅 `router.ServeHTTP` 日日用緊呢個事實）；自己揸 `&http.Server{}` 先有 `Shutdown` 呢個軚盤
- `Shutdown(ctx)` 係成個清場流程本人：**即刻**閂門 → 等 in-flight 食完 → ctx deadline 到冇情講斬線；X 秒係天花板唔係梯級（客走得快佢收得快）
- `http.ErrServerClosed` 係收工紙唔係死因 — sentinel + `errors.Is`，「有 error 而且唔係收工紙先好死」
- `.env`（godotenv）係 dev 便利品：load 唔到唔應該 `Fatal`（production 冇 `.env` file 係常態）— 真判官係 `API_KEY == ""` 嗰句
- Graceful shutdown 係 test 測唔到嘅行為 — 靠實驗儀式攞證據：`/slow` endpoint + mid-flight SIGTERM，對比 curl 有冇食完成碗嘢

**Auth middleware（第八章新增）**
- Middleware factory：`Auth(key string) gin.HandlerFunc` — 收 config 回 middleware，closure 係個袋（= NestJS Guard 嘅 constructor injection，冇 DI 版）；config 旅程：env → `main`（冇 key = `log.Fatal`，唔准裸奔開張）→ `SetupRouter` → closure
- Gin chain 係外面 for loop 推：純放行乜都唔使做，**要停必須 `c.Abort()`**、要留後着先 call `c.Next()`。同 Express 啱啱相反（Express 要自己 `next()` 交棒，唔交 = 吊死）— 危險方向唔同：Express 漏 next 蟲好快現形，**Gin 漏 Abort = 著住 401 衫入屋做嘢**（security hole）
- Route group 局部上鎖：`books.Group("")` + `.Use(Auth(key))` 開子 group，「讀公開、寫上鎖」；小心分房 — 上鎖對象係 POST/PUT/DELETE 唔係 GET（試過裝反，變咗「睇書要證件、加書自由」）
- 401 = unauthenticated（我唔識你），403 = forbidden（識你但你唔准）— 401 個官方名 "Unauthorized" 係 HTTP 史上出名嘅改錯名
- Sentinel 嘅 scope 跟佢嘅「旅程」：`errUnauthenticated` 產同消都喺 `api` package，唔過邊界 → 細楷唔 export、唔使入 `apperr` — 同 `ErrNotFound`（跨 package）對比就明
- `subtle.ConstantTimeCompare` 防 timing attack（`==` 逐 byte 比會漏時間差）；但佢**相等回 1** — 條件寫 `!= 0` 就係倒轉裝閘：啱 key 拒、錯 key 放，兩隻蟲（呢隻 + middleware 未加 401 分支）疊埋變「啱 key → 500」
- `strings.CutPrefix` 拆 "Bearer " 前綴 — 又係 comma-ok 家族
- Test helper 加 optional 參數：`doRequest(..., headers map[string]string)` 傳 `nil` 代表冇 — nil map 讀係安全嘅（第一章知識翻叮）

**Pagination（第七章新增）**
- Contract 先行：default（page=1/limit=10）、上限（`max=100` = 自助 DoS 掣）、超範圍 = 空頁唔係 error（zero value 哲學）、`Total` 永遠全表數 — 全部寫入 `repository.go` 註解做正式契約，suite 逼供
- 入參用 struct（`ListParams`）唔用散裝 int：第日加 sort/filter 係加 field 唔使 domino；出參 `Page{Items, Total}` 一件過
- `ShouldBindQuery` + `form:"page,default=1"`：query binding 世界嘅「absent vs zero」由 `default` tag 分居 — 冇佢 bare `GET /books` 會 422（zero value 又一案發現場）
- Off-by-one 係分頁頭號蟲：offset = `(page-1)*limit`，記住 `*` 先過 `-` — 冇括號嗰條式 page=1 時負 offset 被 GORM 好心當冇事，`Page: 1` 嘅 test 永遠發現唔到，必須有「第二頁」test case
- **nil slice vs 空 slice**：`len`/`range` 眼中孖生，`json.Marshal` 眼中一個 `null` 一個 `[]` — 出街嘅 slice 開波用 `[]T{}`，suite 加 `Items != nil` 鎖
- GORM `Count` 要 `Model()` 指定表、要 `*int64`（自己 `int()` 轉返）；total 同攞頁係兩條 query
- Domain type 唔着 JSON 衫：envelope（`data`/`meta`）喺 handler 砌 `gin.H`，`book.Page` 零 HTTP 知識 — 同「repository 唔識 status code」同一條家規；`meta` echo 生效值等 client 知 default 發生咗
- Debug 判案二定律（今章兩隻蟲各示範一次）：兩個 implementation 齊 fail 同一 test → 疑犯係 test；一過一唔過 → 疑犯係唔過嗰個 implementation

**Project layout（第六章新增）**
- Package 係按 **domain／依賴邊界**分，唔係按技術類型（Laravel `Controllers/`、`Models/` 抽屜 ✗；NestJS domain module ✓）；一個 folder = 一個 package
- 大細楷 = exported/unexported，拆 package 逼你逐個 identifier 決定「API 定內臟」— 特登唔升嘅先係好設計（`books` map、`mu`、`errorTagMessages` 全部收埋）
- `internal/` = compiler 執法嘅 module 私有領域；`cmd/<app>/main.go` = 貧血 wiring 慣例
- Package 命名：短、細楷、講「提供乜」；`global`/`common`/`utils` = 雜物房反 pattern。合法 shared package（`apperr` 出世記）係被逼搬出嚟：內聚單一 + 兩個 domain 真係爭緊先抽 — 唔係預留
- 口吃規則：package 名已提供 context — `book.BookRepository` → `book.Repository`、`book.GormBookStore` → `book.GormStore`
- Import cycle 靠依賴方向設計避免：`main → api → book → apperr` 一條直線；interface 擺 `book`（而唔係 consumer `api`）係為咗俾 contract suite + 兩個 store 指名用，trade-off 講得出就得
- Module path（`go.mod` 第一行）係所有 internal import 嘅前綴

**Testing（第五章新增）**
- Interface compliance suite：測試針對 `BookRepository` contract 寫一次，兩個 implementation 各自注入 factory；對應舊世界嘅 repository contract / driver conformance tests
- Factory function 係 test seam：`func(t *testing.T) BookRepository` 將「點樣開 implementation」同「應該有咩行為」分離；constructor error 可以即場 `t.Fatalf`
- Isolation 粒度要落到每個 subtest：每次 `factory(t)` 開全新 store；GORM 用 `filepath.Join(t.TempDir(), "books.db")`，test 完自動清理
- Shared contract 唔好承諾 implementation detail：ID 只驗正數兼唯一，唔假設一定由 1 開始；刪除用 Create 實際回傳嘅 ID
- Command 成功唔等於 state 正確：Delete 回 nil 後再 Get，驗證 postcondition 真係 `ErrNotFound`，先捉到「假成功／刪錯資料」
- Nested subtest helper 要收最內層嘅 `*testing.T`：`FailNow`/`Fatalf` 只可以由執行該 test 嘅 goroutine 呼叫，唔好 closure 捕捉 parent `t`
- **Flaky test**：Go map iteration order 係刻意隨機化（防止順序變隱形 contract）— test assert `books[0]` 就係將順序寫入 contract，冇 implementation 承諾過就會間歇死；抽打隨機蟲用 `go test -count=30`，一次綠對付唔到佢
- 順序要就正式升做 contract（兩個 implementation 都要明文承諾：mem `sort.Slice` / GORM `Order`），要就 test 讓步只驗 set membership — 唔可以「test 想要但冇人承諾」
- GORM chain 有分**配置**（`Where`/`Order`/`Limit`）同**收尾**（`Find`/`First`/`Save`）：收尾嗰下 SQL 已射出，`Find(...).Order(...)` 個 `Order` 係 no-op — 當時 test 照綠係 SQLite rowid 自然順序賞面（綠 ≠ 啱，again）
- 醫紅嘅鐵律：fix 唔可以係「令佢唔紅」，要係「令佢驗啱嘢」— 剷 assertion 醫紅 = 閹咗個 test（`update` 一度連「冇 update 過」都放行）；紅咗第一條問題係「邊個講大話 — code 定 test？」兩個獨立 implementation 一齊以同一方式 fail → 疑犯係 test

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
