package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var errorTagMessages = map[string]string{
	"required": "The {field} is required.",
	"min":      "The {field} must be at least {value} characters.",
	"max":      "The {field} must be at most {value} characters.",
	"email":    "The {field} must be a valid email address.",
	"url":      "The {field} must be a valid URL.",
	"ip":       "The {field} must be a valid IP address.",
	"gte":      "The {field} must be greater than or equal to {value}.",
	"lte":      "The {field} must be less than or equal to {value}.",
	"eq":       "The {field} must be equal to {value}.",
	"ne":       "The {field} must be not equal to {value}.",
	"gt":       "The {field} must be greater than {value}.",
	"lt":       "The {field} must be less than {value}.",
}

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
		var validationErrors validator.ValidationErrors
		var apiError *APIError
		if errors.As(err, &validationErrors) {
			errs := make(map[string][]string)
			for _, fe := range validationErrors {
				fieldName := strings.ToLower(fe.Field())
				originalMessage, ok := errorTagMessages[fe.Tag()]
				if !ok {
					originalMessage = "The {field} is invalid."
				}
				fieldMessage := strings.Replace(originalMessage, "{value}", fe.Param(), 1)
				fieldMessage = strings.Replace(fieldMessage, "{field}", fieldName, 1)
				errs[fieldName] = append(errs[fieldName], fieldMessage)
			}
			c.JSON(http.StatusUnprocessableEntity, gin.H{"message": "The given data was invalid.", "errors": errs})
		} else if errors.Is(err, ErrBookNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "book not found"})
		} else if errors.As(err, &apiError) {
			c.JSON(apiError.Status, gin.H{"message": apiError.Message})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
		}
	}
}
