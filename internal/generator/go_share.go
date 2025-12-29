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

// AppError 应用错误
type AppError struct {
	Code    int    ` + "`json:\"code\"`" + `
	Message string ` + "`json:\"message\"`" + `
	Err     error  ` + "`json:\"-\"`" + `
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

// 预定义错误码
const (
	CodeSuccess       = 0
	CodeBadRequest    = 400
	CodeUnauthorized  = 401
	CodeForbidden     = 403
	CodeNotFound      = 404
	CodeConflict      = 409
	CodeInternalError = 500
)

// NewAppError 创建应用错误
func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// ErrBadRequest 请求参数错误
func ErrBadRequest(message string) *AppError {
	return NewAppError(CodeBadRequest, message, nil)
}

// ErrNotFound 资源不存在
func ErrNotFound(message string) *AppError {
	return NewAppError(CodeNotFound, message, nil)
}

// ErrUnauthorized 未授权
func ErrUnauthorized(message string) *AppError {
	return NewAppError(CodeUnauthorized, message, nil)
}

// ErrForbidden 禁止访问
func ErrForbidden(message string) *AppError {
	return NewAppError(CodeForbidden, message, nil)
}

// ErrConflict 资源冲突
func ErrConflict(message string) *AppError {
	return NewAppError(CodeConflict, message, nil)
}

// ErrInternal 内部错误
func ErrInternal(message string, err error) *AppError {
	return NewAppError(CodeInternalError, message, err)
}
`
	if err := g.writeFile("share/errors/app_error.go", appErrorTmpl); err != nil {
		return err
	}

	// utils/crypto.go
	cryptoTmpl := `package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 哈希密码
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash 验证密码
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// SHA256Hash 计算 SHA256 哈希
func SHA256Hash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
`
	if err := g.writeFile("share/utils/crypto.go", cryptoTmpl); err != nil {
		return err
	}

	// utils/trace.go
	traceTmpl := `package utils

import (
	"github.com/google/uuid"
)

// GenerateTraceID 生成追踪 ID
func GenerateTraceID() string {
	return uuid.New().String()
}
`
	if err := g.writeFile("share/utils/trace.go", traceTmpl); err != nil {
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

	// types/error_handler.go
	errorHandlerTmpl := `package types

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"{{.ModulePath}}/share/errors"
)

// HandleError 统一错误处理
func HandleError(ctx context.Context, c *app.RequestContext, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		status := http.StatusInternalServerError
		switch appErr.Code {
		case errors.CodeBadRequest:
			status = http.StatusBadRequest
		case errors.CodeUnauthorized:
			status = http.StatusUnauthorized
		case errors.CodeForbidden:
			status = http.StatusForbidden
		case errors.CodeNotFound:
			status = http.StatusNotFound
		case errors.CodeConflict:
			status = http.StatusConflict
		}
		c.JSON(status, Error(appErr.Code, appErr.Message))
		return
	}

	c.JSON(http.StatusInternalServerError, Error(errors.CodeInternalError, "内部服务错误"))
}
`
	if err := g.renderAndWrite(errorHandlerTmpl, "share/types/error_handler.go"); err != nil {
		return err
	}

	// middleware/trace.go
	traceMiddlewareTmpl := `package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"{{.ModulePath}}/share/utils"
)

const TraceIDKey = "X-Trace-ID"

// Trace 追踪中间件
func Trace() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		traceID := string(c.GetHeader(TraceIDKey))
		if traceID == "" {
			traceID = utils.GenerateTraceID()
		}
		c.Set(TraceIDKey, traceID)
		c.Header(TraceIDKey, traceID)
		c.Next(ctx)
	}
}

// GetTraceID 获取追踪 ID
func GetTraceID(c *app.RequestContext) string {
	if traceID, exists := c.Get(TraceIDKey); exists {
		return traceID.(string)
	}
	return ""
}
`
	if err := g.renderAndWrite(traceMiddlewareTmpl, "share/middleware/trace.go"); err != nil {
		return err
	}

	return nil
}
