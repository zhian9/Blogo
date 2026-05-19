// SPDX-License-Identifier: MIT
//
// Copyright (c) 2026-present 李星云 (lxy911)
//
// Project: Blogo
// Repository: https://github.com/zhian9/Blogo

package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/pkg/errors"
)

// 兼容性封装 pkg/errors
var (
	WithStack = errors.WithStack // 为错误添加调用栈
	Wrap      = errors.Wrap      // 包装错误（带上下文）
	Wrapf     = errors.Wrapf     // 包装错误（带格式化）
	Is        = errors.Is        // 判断错误是否匹配
	Errorf    = errors.Errorf    // 创建带格式化的错误
)

// 预定义常用错误的 ID（用于前端国际化或错误分类）
const (
	DefaultBadRequestID            = "bad_request"
	DefaultUnauthorizedID          = "unauthorized"
	DefaultForbiddenID             = "forbidden"
	DefaultNotFoundID              = "not_found"
	DefaultMethodNotAllowedID      = "method_not_allowed"
	DefaultTooManyRequestsID       = "too_many_requests"
	DefaultRequestEntityTooLargeID = "request_entity_too_large"
	DefaultInternalServerErrorID   = "internal_server_error"
	DefaultConflictID              = "conflict"
	DefaultRequestTimeoutID        = "request_timeout"
)

// Error 自定义错误结构体 实现 error 接口，并支持 JSON 序列化。
type Error struct {
	ID     string `json:"id,omitempty"`     // 错误标识（前端友好）
	Code   int32  `json:"code,omitempty"`   // HTTP 状态码
	Detail string `json:"detail,omitempty"` // 详细描述
	Status string `json:"status,omitempty"` // 状态文本（如 "Not Found"）
}

// Error 实现 error 接口，返回 JSON 字符串（便于日志记录）
func (e *Error) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// New 创建自定义错误
func New(id, detail string, code int32) error {
	return &Error{
		ID:     id,
		Code:   code,
		Detail: detail,
		Status: http.StatusText(int(code)),
	}
}

// Parse 尝试将字符串解析为 *Error。
func Parse(err string) *Error {
	e := new(Error)
	errr := json.Unmarshal([]byte(err), e)
	if errr != nil {
		// 解析失败 → 整个字符串作为错误详情
		e.Detail = err
	}
	return e
}

// BadRequest 生成 400 错误（客户端请求错误）
func BadRequest(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultBadRequestID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusBadRequest,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusBadRequest),
	}
}

// Unauthorized 生成 401 错误（未认证）
func Unauthorized(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultUnauthorizedID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusUnauthorized,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusUnauthorized),
	}
}

// Forbidden 生成 403 错误（无权限）
func Forbidden(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultForbiddenID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusForbidden,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusForbidden),
	}
}

// NotFound 生成 404 错误（资源不存在）
func NotFound(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultNotFoundID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusNotFound,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusNotFound),
	}
}

// MethodNotAllowed 生成 405 错误（方法不允许）
func MethodNotAllowed(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultMethodNotAllowedID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusMethodNotAllowed,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusMethodNotAllowed),
	}
}

// TooManyRequests 生成 429 错误（请求过多）
func TooManyRequests(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultTooManyRequestsID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusTooManyRequests,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusTooManyRequests),
	}
}

// Timeout 生成 408 错误（请求超时）
func Timeout(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultRequestTimeoutID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusRequestTimeout,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusRequestTimeout),
	}
}

// Conflict 生成 409 错误（资源冲突）
func Conflict(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultConflictID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusConflict,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusConflict),
	}
}

// RequestEntityTooLarge 生成 413 错误（请求体过大）
func RequestEntityTooLarge(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultRequestEntityTooLargeID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusRequestEntityTooLarge,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusRequestEntityTooLarge),
	}
}

// InternalServerError 生成 500 错误（服务器内部错误）
func InternalServerError(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultInternalServerErrorID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusInternalServerError,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

// Equal 比较两个错误是否相等（基于 Code）
func Equal(err1 error, err2 error) bool {
	verr1, ok1 := err1.(*Error)
	verr2, ok2 := err2.(*Error)

	if ok1 != ok2 {
		return false
	}

	if !ok1 {
		// 都不是 *Error → 直接比较指针
		return err1 == err2
	}

	// 只比较 Code（HTTP 语义）
	return verr1.Code == verr2.Code
}

// FromError 尝试将任意 error 转换为 *Error
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	if verr, ok := err.(*Error); ok && verr != nil {
		return verr
	}
	return Parse(err.Error())
}

// As 在错误链中查找第一个 *Error 类型的错误
// 与 errors.As 标准库兼容，支持 Wrap/Wrapf 包装的错误
func As(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	var merr *Error
	if errors.As(err, &merr) {
		return merr, true
	}
	return nil, false
}

// MultiError 用于聚合多个错误（如表单验证失败）
type MultiError struct {
	lock   *sync.Mutex
	Errors []error
}

// NewMultiError 创建多错误容器
func NewMultiError() *MultiError {
	return &MultiError{
		lock:   &sync.Mutex{},
		Errors: make([]error, 0),
	}
}

// Append 添加错误（非并发安全）
func (e *MultiError) Append(err error) {
	e.Errors = append(e.Errors, err)
}

// AppendWithLock 添加错误（并发安全）
func (e *MultiError) AppendWithLock(err error) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.Append(err)
}

// HasErrors 判断是否包含错误
func (e *MultiError) HasErrors() bool {
	return len(e.Errors) > 0
}

// Error 实现 error 接口，返回 JSON 字符串
func (e *MultiError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}
func NewBadRequest(format string, a ...interface{}) error {
	return BadRequest("", format, a...) // id 传空，自动使用 DefaultBadRequestID
}
