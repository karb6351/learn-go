package api

// ── 第八章骨架：Auth middleware（request phase）──────────────
// ErrorHandler 係 onion 嘅 response phase（c.Next() 之後執手尾）；
// 呢個係佢嘅鏡像：request 未入 handler 之前查身份證。

import (
	"crypto/subtle"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

var errUnauthenticated = errors.New("unauthenticated")

// Auth 係一個 middleware factory：收條 key，回一個「記住咗佢」嘅 middleware。
// 形態同 ErrorHandler() 一樣，多咗個參數 — return 出去嗰個 function 係
// closure：佢困住咗 key 呢個變數，第時每個 request 行到都攞返嚟用。
// 對應舊世界：NestJS Guard 靠 constructor injection 收依賴 — 呢度冇 DI，
// closure 就係個袋。
func Auth(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok {
			c.Error(errUnauthenticated)
			c.Abort()
			return
		}

		if subtle.ConstantTimeCompare([]byte(token), []byte(key)) == 0 {
			c.Error(errUnauthenticated)
			c.Abort()
			return
		}
	}
}
