package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestMain 係成個 package 測試嘅入口 hook（類似 Jest 嘅 globalSetup）。
// 呢度用嚟熄咗 Gin 嘅 debug log，等 test output 乾淨啲。
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// 每個 test 用呢個 helper 攞一個全新嘅 router + store。
// t.Helper() 話俾 testing framework 知：呢個 function 係 helper，
// 報錯時行號指向 caller，唔係指向呢度。
func newTestRouter(t *testing.T) (*gin.Engine, *BookStore) {
	t.Helper()
	store := NewBookStore()
	return setupRouter(store), store
}

// doRequest 模擬一次 HTTP request：
//
//	httptest.NewRecorder() = 假嘅 ResponseWriter，錄低 handler 寫咗啲乜
//	router.ServeHTTP(...)  = 直接餵 request 入 router，全程 in-process
func doRequest(t *testing.T, router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestCreateBookHandler(t *testing.T) {
	router, _ := newTestRouter(t)

	rec := doRequest(t, router, http.MethodPost, "/books",
		`{"title":"The Go Programming Language","author":"Donovan","year":2015}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	// 將 response JSON parse 返做 Book 嚟驗證內容
	var got Book
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if got.ID != 1 || got.Title != "The Go Programming Language" {
		t.Errorf("unexpected book: %+v", got)
	}
}

func TestCreateBookValidation(t *testing.T) {
	router, _ := newTestRouter(t)

	rec := doRequest(t, router, http.MethodPost, "/books", `{"title":"No Author"}`)

	// Laravel-style validation contract：422 + message + errors map
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}

	var resp struct {
		Message string              `json:"message"`
		Errors  map[string][]string `json:"errors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v; body: %s", err, rec.Body.String())
	}
	if resp.Message == "" {
		t.Errorf("message should not be empty")
	}
	// 只 assert 契約核心（author 有錯、有講原因），唔 assert 具體措辭 —
	// 措辭係 implementation，邊個 field 錯先係 contract
	if len(resp.Errors["author"]) == 0 {
		t.Errorf("errors.author should contain at least one message; body: %s", rec.Body.String())
	}
}

func TestGetBookNotFound(t *testing.T) {
	router, _ := newTestRouter(t)

	rec := doRequest(t, router, http.MethodGet, "/books/999", "")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestGetBookHandlerValidation(t *testing.T) {
	router, _ := newTestRouter(t)

	rec := doRequest(t, router, http.MethodGet, "/books/0", "")

	// Laravel-style validation contract：422 + message + s map
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}

	var resp struct {
		Message string              `json:"message"`
		Errors  map[string][]string `json:"errors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v; body: %s", err, rec.Body.String())
	}
	if resp.Message == "" {
		t.Errorf("message should not be empty")
	}
	// 只 assert 契約核心（id 有錯、有講原因），唔 assert 具體措辭 —
	// 措辭係 implementation，邊個 field 錯先係 contract
	if len(resp.Errors["id"]) == 0 {
		t.Errorf("errors.id should contain at least one message; body: %s", rec.Body.String())
	}
}

func TestGetBookHandler(t *testing.T) {

	tests := []struct {
		name       string // subtest 名，會顯示喺 go test -v 度
		seed       func(s *BookStore)
		path       string
		body       string
		wantStatus int
	}{
		{
			name: "normal get",
			seed: func(s *BookStore) {
				s.Create(Book{BaseBook: BaseBook{Title: "Book 1", Author: "Author 1", Year: 2015}})
			},
			path:       "/books/1",
			body:       "",
			wantStatus: http.StatusOK,
		},
		{
			name: "uri id is not a valid integer",
			seed: func(s *BookStore) {
			},
			path:       "/books/not-a-number",
			body:       "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "book not found",
			seed: func(s *BookStore) {
			},
			path:       "/books/999",
			body:       "",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { // t.Run = subtest，每個 case 獨立 pass/fail
			router, store := newTestRouter(t)
			tt.seed(store)

			rec := doRequest(t, router, http.MethodGet, tt.path, tt.body)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var got Book
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("response is not valid JSON: %v", err)
				}
				if got.Title != "Book 1" || got.Author != "Author 1" || got.Year != 2015 {
					t.Errorf("unexpected book: %+v", got)
				}
			}
		})
	}
}

func TestUpdateBookHandler(t *testing.T) {
	tests := []struct {
		name       string // subtest 名，會顯示喺 go test -v 度
		seed       func(s *BookStore)
		path       string
		body       string
		wantStatus int
	}{
		{
			name: "normal update",
			seed: func(s *BookStore) {
				s.Create(Book{BaseBook: BaseBook{Title: "Book 1", Author: "Author 1", Year: 2015}})
			},
			path:       "/books/1",
			body:       `{"title":"B","author":"Y","year":2016}`,
			wantStatus: http.StatusOK,
		},
		{
			name: "missing book",
			seed: func(s *BookStore) {
			},
			path:       "/books/42",
			body:       `{"title":"B","author":"Y","year":2016}`,
			wantStatus: http.StatusNotFound,
		},
		{
			name: "uri id is not a valid integer",
			seed: func(s *BookStore) {
			},
			path:       "/books/not-a-number",
			body:       "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing author",
			seed: func(s *BookStore) {
				s.Create(Book{BaseBook: BaseBook{Title: "Book 1", Author: "Author 1", Year: 2015}})
			},
			path:       "/books/1",
			body:       `{"title":"B","year":2016}`,
			wantStatus: http.StatusUnprocessableEntity, // Laravel style：validation 一律 422

		},
		{
			name: "year too old",
			seed: func(s *BookStore) {
				s.Create(Book{BaseBook: BaseBook{Title: "Book 1", Author: "Author 1", Year: 2015}})
			},
			path:       "/books/1",
			body:       `{"title":"B","author":"Y","year":3000}`,
			wantStatus: http.StatusUnprocessableEntity, // Laravel style：validation 一律 422

		},
		{
			name: "id hijack",
			seed: func(s *BookStore) {
				s.Create(Book{BaseBook: BaseBook{Title: "Book 1", Author: "Author 1", Year: 2015}})
			},
			path:       "/books/1",
			body:       `{"id":999,"title":"B","author":"Y","year":2016}`,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { // t.Run = subtest，每個 case 獨立 pass/fail
			router, store := newTestRouter(t)
			tt.seed(store)

			rec := doRequest(t, router, http.MethodPut, tt.path, tt.body)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusOK {
				var got Book
				if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
					t.Fatalf("response is not valid JSON: %v", err)
				}
				if got.Title != "B" || got.Author != "Y" || got.Year != 2016 {
					t.Errorf("unexpected book: %+v", got)
				}
				if got.ID != 1 {
					t.Errorf("unexpected id: %d", got.ID)
				}
			}
		})
	}
}

func TestDeleteBookHandlerValidation(t *testing.T) {
	router, _ := newTestRouter(t)

	rec := doRequest(t, router, http.MethodDelete, "/books/0", "")

	// Laravel-style validation contract：422 + message + errors map
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}

	var resp struct {
		Message string              `json:"message"`
		Errors  map[string][]string `json:"errors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not valid JSON: %v; body: %s", err, rec.Body.String())
	}
	if resp.Message == "" {
		t.Errorf("message should not be empty")
	}
	// 只 assert 契約核心（id 有錯、有講原因），唔 assert 具體措辭 —
	// 措辭係 implementation，邊個 field 錯先係 contract
	if len(resp.Errors["id"]) == 0 {
		t.Errorf("errors.id should contain at least one message; body: %s", rec.Body.String())
	}
}

func TestDeleteBookHandler(t *testing.T) {

	tests := []struct {
		name       string // subtest 名，會顯示喺 go test -v 度
		seed       func(s *BookStore)
		path       string
		body       string
		wantStatus int
	}{
		{
			name: "normal delete",
			seed: func(s *BookStore) {
				s.Create(Book{BaseBook: BaseBook{Title: "Book 1", Author: "Author 1", Year: 2015}})
			},
			path:       "/books/1",
			body:       "",
			wantStatus: http.StatusNoContent,
		},
		{
			name: "uri id is not a valid integer",
			seed: func(s *BookStore) {
			},
			path:       "/books/abc",
			body:       "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { // t.Run = subtest，每個 case 獨立 pass/fail
			router, store := newTestRouter(t)
			tt.seed(store)

			rec := doRequest(t, router, http.MethodDelete, tt.path, tt.body)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			if tt.wantStatus == http.StatusNoContent {
				if rec.Body.String() != "" {
					t.Errorf("body should be empty; body: %s", rec.Body.String())
				}
			}
		})
	}
}
