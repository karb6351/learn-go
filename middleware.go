package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandler 係全 API 嘅 error 出口（相當於 Laravel 嘅 Handler::render /
// NestJS 嘅 exception filter）。Handler 唔再自己寫 error response，
// 改為 c.Error(err) + c.Abort()，由呢度統一 translate 做 HTTP response。
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // ── 之前：request 未入 handler；之後：handler 行完 ──

		if len(c.Errors) == 0 {
			return // 冇 error 掛喺 context，唔關我事
		}
		err := c.Errors.Last().Err

		// TODO(你嚟寫): translation switch，將 err 轉做 Laravel-style response
		//
		// 目標 contract：
		//   validation 失敗 → 422 + {"message":"The given data was invalid.",
		//                             "errors":{"author":["The author field is required."]}}
		//   ErrBookNotFound → 404 + {"message":"book not found"}
		//   其他（兜底）     → 500 + {"message":"internal error"}
		//
		// 工具提示：
		//   1. validation error 要用 errors.As 提取 concrete type：
		//        var ve validator.ValidationErrors        // import "github.com/go-playground/validator/v10"
		//        if errors.As(err, &ve) {
		//            for _, fe := range ve {
		//                fe.Field()  // "Author"（Go field 名，記得轉細楷）
		//                fe.Tag()    // "required" / "gte" / ...
		//            }
		//        }
		//   2. errors map 嘅 type 係 map[string][]string（一個 field 可以有多條 message）
		//   3. message 措辭你話事 — 想似 Laravel 可以砌 "The <field> field is <tag>."
		//   4. ErrBookNotFound 用返 errors.Is（sentinel 比對）
		//
		// 記住剷埋下面兩行 placeholder
		_ = err
		c.JSON(http.StatusInternalServerError, gin.H{"message": "TODO: not translated yet"})
	}
}
