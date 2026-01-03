package generator

// generateShare 生成 share 公共模块
func (g *GoGenerator) generateShare() error {
	// go.mod
	goModTmpl := `module {{.ModulePath}}/share

go 1.24.11

require (
	{{.ModulePath}}/bom v0.0.0

	// Hertz HTTP 框架
	github.com/cloudwego/hertz v0.9.3

	// 通用工具
	github.com/google/uuid v1.6.0
	github.com/bytedance/sonic v1.12.6
)

replace {{.ModulePath}}/bom => ../bom
`
	if err := g.renderAndWrite(goModTmpl, "share/go.mod"); err != nil {
		return err
	}

	// errors/app_error.go
	appErrorTmpl := `package errors

import "fmt"

// AppError 应用错误基类
type AppError struct {
	Code    int    ` + "`json:\"code\"`" + `    // 错误码
	Message string ` + "`json:\"message\"`" + ` // 错误信息
	Err     error  ` + "`json:\"-\"`" + `       // 原始错误
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 实现 errors.Unwrap 接口
func (e *AppError) Unwrap() error {
	return e.Err
}

// New 创建新的应用错误
func New(code int, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap 包装原始错误
func Wrap(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ==================== 通用错误 ====================
// 错误码分段: 1xxxx
// 示例: 10001, 10002...

const (
	// 通用错误码 10000-10999
	Success       = 200   // 成功
	BadRequest    = 10001 // 请求参数错误
	Unauthorized  = 10002 // 未授权
	Forbidden     = 10003 // 禁止访问
	NotFound      = 10004 // 资源不存在
	Conflict      = 10005 // 资源冲突
	InternalError = 10006 // 内部错误
)

// ErrBadRequest 请求参数错误
func ErrBadRequest(message string) *AppError {
	return New(BadRequest, message)
}

// ErrNotFound 资源不存在
func ErrNotFound(message string) *AppError {
	return New(NotFound, message)
}

// ErrUnauthorized 未授权
func ErrUnauthorized(message string) *AppError {
	return New(Unauthorized, message)
}

// ErrForbidden 禁止访问
func ErrForbidden(message string) *AppError {
	return New(Forbidden, message)
}

// ErrConflict 资源冲突
func ErrConflict(message string) *AppError {
	return New(Conflict, message)
}

// ErrInternal 内部错误
func ErrInternal(message string, err error) *AppError {
	return Wrap(InternalError, message, err)
}
`
	if err := g.writeFile("share/errors/app_error.go", appErrorTmpl); err != nil {
		return err
	}

	// errors/error_handler.go
	errorHandlerTmpl := `package errors

import (
	"context"
	"errors"
	"net/http"
	"{{.ModulePath}}/share/types"

	"github.com/cloudwego/hertz/pkg/app"
)

// HandleError 统一错误处理
// 支持处理 AppError 及其继承类型（如 UserError）
func HandleError(ctx context.Context, c *app.RequestContext, err error) {
	// 使用 errors.As 支持嵌入类型的解包
	var appErr *AppError
	if errors.As(err, &appErr) {
		status := getHTTPStatus(appErr.Code)
		c.JSON(status, types.Error(appErr.Code, appErr.Message))
		return
	}

	c.JSON(http.StatusInternalServerError, types.Error(InternalError, "内部服务错误"))
}

// getHTTPStatus 根据业务错误码获取对应的 HTTP 状态码
// 错误码分段规则:
//   10000-10999: 通用错误
//   11000-11999: User 模块
//   12000-12999: Order 模块
//   ...以此类推
func getHTTPStatus(code int) int {
	// 根据错误码末尾判断类型
	switch code % 100 {
	case 1: // xxx01: bad_request
		return http.StatusBadRequest
	case 2: // xxx02: unauthorized
		return http.StatusUnauthorized
	case 3: // xxx03: forbidden
		return http.StatusForbidden
	case 4: // xxx04: not_found
		return http.StatusNotFound
	case 5: // xxx05: conflict
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// IsAppError 判断是否为 AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError 将 error 转换为 AppError
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
`
	if err := g.renderAndWrite(errorHandlerTmpl, "share/errors/error_handler.go"); err != nil {
		return err
	}

	// types/response.go
	responseTmpl := `package types

// Response 统一响应结构
type Response struct {
	Code    int         ` + "`json:\"code\"`" + `
	Message string      ` + "`json:\"message\"`" + `
	Data    interface{} ` + "`json:\"data,omitempty\"`" + `
	TraceID string      ` + "`json:\"trace_id,omitempty\"`" + `
}

// Success 成功响应
func Success(data interface{}) *Response {
	return &Response{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// SuccessWithMessage 带消息的成功响应
func SuccessWithMessage(message string, data interface{}) *Response {
	return &Response{
		Code:    0,
		Message: message,
		Data:    data,
	}
}

// Error 错误响应
func Error(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

// PageResult 分页结果
type PageResult struct {
	List     interface{} ` + "`json:\"list\"`" + `
	Total    int64       ` + "`json:\"total\"`" + `
	Page     int         ` + "`json:\"page\"`" + `
	PageSize int         ` + "`json:\"page_size\"`" + `
}
`
	if err := g.writeFile("share/types/response.go", responseTmpl); err != nil {
		return err
	}

	return nil
}
