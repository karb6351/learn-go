package apperr

import (
	"errors"
	"fmt"
)

// ErrNotFound 係一個 sentinel error。
// Go 冇 exception，錯誤係用 return value 傳返出去，
// caller 用 errors.Is() 嚟判斷係邊種錯。
var ErrNotFound = errors.New("resource not found")

// ResourceNotFoundError 係一個 custom error。
// caller 用 errors.As() 嚟 unpack。
type ResourceNotFoundError struct {
	Resource string
	ID       int
}

func (e *ResourceNotFoundError) Error() string {
	return fmt.Sprintf("%s %d not found", e.Resource, e.ID)
}

func (e *ResourceNotFoundError) Unwrap() error {
	return ErrNotFound
}
