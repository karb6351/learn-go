package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var errorTagMessages = map[string]string{
	"required": "The {field} is required.",
	"min":      "The {field} must be at least {value}.",
	"max":      "The {field} must be at most {value}.",
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

		// 呢個 switch 識得處理三種 error 型式：
		// 1. validator.ValidationErrors（用 errors.As）：代表 validate struct input（eg binding/validation）出錯
		//    ==> translate 做 HTTP 422 Unprocessable Entity。用 As 因為 validator.ValidationErrors 係一個 error slice type，要檢查「有冇包含呢個型態」。
		// 2. ErrBookNotFound（用 errors.Is）：業務邏輯自定義 error，代表資源搵唔到
		//    ==> translate 做 HTTP 404 Not Found。用 Is 因為 ErrBookNotFound 係一個 sentinel error，直接 match 就得。
		// 3. *strconv.NumError（用 errors.As）：通常出現於 query param / 路徑等 parse int 失敗
		//    ==> translate 做 HTTP 400 Bad Request。用 As 因為 *NumError 係 struct，可能用 wrap 攞出嚟。
		// 最後 fallback（兜底）交畀 HTTP 500 Internal Server Error
		//    ==> 設計意圖：如果有未 translate 嘅崩潰／預料外錯誤響度，都會用 500 alert，方便追蹤程式設計唔完善/有新 error case 未覆蓋。

		var validationErrors validator.ValidationErrors
		var numError *strconv.NumError
		if errors.As(err, &validationErrors) {
			// 處理 validation error，HTTP 422
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
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"message": "The given data was invalid.",
				"errors":  errs,
			})
		} else if errors.Is(err, ErrBookNotFound) {
			// 處理書本資源唔存在，HTTP 404
			c.JSON(http.StatusNotFound, gin.H{"message": "book not found"})
		} else if errors.As(err, &numError) {
			// 處理 integer parse 錯，HTTP 400
			c.JSON(http.StatusBadRequest, gin.H{"message": numError.Num + " is not a valid integer"})
		} else {
			// 任何未捕捉/未 translate 嘅error都出 HTTP 500
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal error"})
		}
	}
}
