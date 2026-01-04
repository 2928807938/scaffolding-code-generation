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

	// GORM ORM 框架
	gorm.io/gorm v1.25.12
	gorm.io/driver/mysql v1.5.7
	gorm.io/driver/postgres v1.5.11
	gorm.io/driver/sqlite v1.5.7
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

	// repository/base.go
	repoBaseTmpl := `package repository

import (
	"context"
)

// BaseRepository 基础仓储接口，定义通用的 CRUD 操作
// T 为实体类型，ID 为主键类型
type BaseRepository[T any, ID comparable] interface {
	// Create 创建单个实体
	Create(ctx context.Context, entity *T) error

	// CreateBatch 批量创建实体
	CreateBatch(ctx context.Context, entities []*T) error

	// GetByID 根据主键查询
	GetByID(ctx context.Context, id ID) (*T, error)

	// Update 更新实体
	Update(ctx context.Context, entity *T) error

	// Delete 删除实体（逻辑删除）
	Delete(ctx context.Context, id ID) error

	// List 查询全部列表
	List(ctx context.Context) ([]*T, error)

	// Page 分页查询
	Page(ctx context.Context, request *PageRequest) (*PageResult[*T], error)
}

// TransactionalRepository 支持事务的仓储接口
type TransactionalRepository interface {
	// BeginTx 开启事务
	BeginTx(ctx context.Context) (context.Context, error)

	// Commit 提交事务
	Commit(ctx context.Context) error

	// Rollback 回滚事务
	Rollback(ctx context.Context) error
}

// Entity 实体接口，所有实体必须实现此接口
type Entity[ID comparable] interface {
	// GetID 获取实体主键
	GetID() ID
}
`
	if err := g.writeFile("share/repository/base.go", repoBaseTmpl); err != nil {
		return err
	}

	// repository/builder.go
	repoBuilderTmpl := `package repository

import (
	"context"
)

// QueryBuilder 查询构建器接口，提供链式调用的查询构建能力
type QueryBuilder[T any] interface {
	// Where 添加查询条件
	Where(condition *Condition) QueryBuilder[T]

	// And 添加 AND 条件
	And(conditions ...*Condition) QueryBuilder[T]

	// OrderBy 添加排序（升序）
	OrderBy(field string) QueryBuilder[T]

	// OrderByDesc 添加排序（降序）
	OrderByDesc(field string) QueryBuilder[T]

	// Limit 限制返回数量
	Limit(limit int) QueryBuilder[T]

	// Offset 设置偏移量
	Offset(offset int) QueryBuilder[T]

	// Select 指定查询字段
	Select(fields ...string) QueryBuilder[T]

	// Find 执行查询，返回结果列表
	Find(ctx context.Context) ([]*T, error)

	// First 执行查询，返回第一条结果
	First(ctx context.Context) (*T, error)

	// Count 执行统计查询
	Count(ctx context.Context) (int64, error)

	// Exists 执行存在性检查
	Exists(ctx context.Context) (bool, error)
}

// QueryOptions 查询选项，用于存储构建器的状态
type QueryOptions struct {
	Conditions []*Condition // 查询条件列表
	OrderBys   []OrderBy    // 排序规则列表
	LimitVal   int          // 限制数量
	OffsetVal  int          // 偏移量
	Fields     []string     // 查询字段
}

// NewQueryOptions 创建查询选项
func NewQueryOptions() *QueryOptions {
	return &QueryOptions{
		Conditions: make([]*Condition, 0),
		OrderBys:   make([]OrderBy, 0),
		LimitVal:   0,
		OffsetVal:  0,
		Fields:     make([]string, 0),
	}
}

// AddCondition 添加条件
func (o *QueryOptions) AddCondition(condition *Condition) *QueryOptions {
	o.Conditions = append(o.Conditions, condition)
	return o
}

// AddConditions 批量添加条件
func (o *QueryOptions) AddConditions(conditions ...*Condition) *QueryOptions {
	o.Conditions = append(o.Conditions, conditions...)
	return o
}

// AddOrderBy 添加排序
func (o *QueryOptions) AddOrderBy(field string, desc bool) *QueryOptions {
	o.OrderBys = append(o.OrderBys, OrderBy{Field: field, Desc: desc})
	return o
}

// SetLimit 设置限制
func (o *QueryOptions) SetLimit(limit int) *QueryOptions {
	o.LimitVal = limit
	return o
}

// SetOffset 设置偏移
func (o *QueryOptions) SetOffset(offset int) *QueryOptions {
	o.OffsetVal = offset
	return o
}

// SetFields 设置查询字段
func (o *QueryOptions) SetFields(fields ...string) *QueryOptions {
	o.Fields = fields
	return o
}
`
	if err := g.writeFile("share/repository/builder.go", repoBuilderTmpl); err != nil {
		return err
	}

	// repository/page.go
	repoPageTmpl := `package repository

// PageRequest 分页请求
type PageRequest struct {
	Page       int          ` + "`json:\"page\"`" + `       // 页码（从 1 开始）
	Size       int          ` + "`json:\"size\"`" + `       // 每页数量
	Conditions []*Condition ` + "`json:\"conditions\"`" + ` // 查询条件列表
	OrderBy    []OrderBy    ` + "`json:\"order_by\"`" + `   // 排序规则
}

// OrderBy 排序规则
type OrderBy struct {
	Field string ` + "`json:\"field\"`" + ` // 排序字段
	Desc  bool   ` + "`json:\"desc\"`" + `  // 是否降序
}

// NewPageRequest 创建分页请求
func NewPageRequest(page, size int) *PageRequest {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}
	return &PageRequest{
		Page:       page,
		Size:       size,
		Conditions: make([]*Condition, 0),
		OrderBy:    make([]OrderBy, 0),
	}
}

// WithCondition 添加查询条件
func (p *PageRequest) WithCondition(condition *Condition) *PageRequest {
	p.Conditions = append(p.Conditions, condition)
	return p
}

// WithOrderBy 添加排序规则
func (p *PageRequest) WithOrderBy(field string, desc bool) *PageRequest {
	p.OrderBy = append(p.OrderBy, OrderBy{Field: field, Desc: desc})
	return p
}

// Offset 计算偏移量
func (p *PageRequest) Offset() int {
	return (p.Page - 1) * p.Size
}

// PageResult 分页结果
type PageResult[T any] struct {
	Items      []T   ` + "`json:\"items\"`" + `       // 当前页数据列表
	Total      int64 ` + "`json:\"total\"`" + `       // 总记录数
	Page       int   ` + "`json:\"page\"`" + `        // 当前页码
	Size       int   ` + "`json:\"size\"`" + `        // 每页数量
	TotalPages int   ` + "`json:\"total_pages\"`" + ` // 总页数
}

// NewPageResult 创建分页结果
func NewPageResult[T any](items []T, total int64, page, size int) *PageResult[T] {
	totalPages := int(total) / size
	if int(total)%size != 0 {
		totalPages++
	}
	return &PageResult[T]{
		Items:      items,
		Total:      total,
		Page:       page,
		Size:       size,
		TotalPages: totalPages,
	}
}

// HasNext 是否有下一页
func (p *PageResult[T]) HasNext() bool {
	return p.Page < p.TotalPages
}

// HasPrev 是否有上一页
func (p *PageResult[T]) HasPrev() bool {
	return p.Page > 1
}

// IsEmpty 是否为空
func (p *PageResult[T]) IsEmpty() bool {
	return len(p.Items) == 0
}
`
	if err := g.writeFile("share/repository/page.go", repoPageTmpl); err != nil {
		return err
	}

	// repository/queryable.go
	repoQueryableTmpl := `package repository

import "context"

// Operator 操作符类型
type Operator string

const (
	// 比较操作符
	OpEqual          Operator = "="
	OpNotEqual       Operator = "!="
	OpGreaterThan    Operator = ">"
	OpGreaterOrEqual Operator = ">="
	OpLessThan       Operator = "<"
	OpLessOrEqual    Operator = "<="

	// 模糊匹配
	OpLike Operator = "LIKE"

	// 集合操作
	OpIn    Operator = "IN"
	OpNotIn Operator = "NOT IN"

	// 区间操作
	OpBetween Operator = "BETWEEN"

	// 空值检查
	OpIsNull    Operator = "IS NULL"
	OpIsNotNull Operator = "IS NOT NULL"
)

// Condition 查询条件
type Condition struct {
	Field    string      // 要查询的字段名
	Operator Operator    // 操作符
	Value    interface{} // 比较的值
}

// NewCondition 创建查询条件
func NewCondition(field string, operator Operator, value interface{}) *Condition {
	return &Condition{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

// Eq 等于条件
func Eq(field string, value interface{}) *Condition {
	return NewCondition(field, OpEqual, value)
}

// NotEq 不等于条件
func NotEq(field string, value interface{}) *Condition {
	return NewCondition(field, OpNotEqual, value)
}

// Gt 大于条件
func Gt(field string, value interface{}) *Condition {
	return NewCondition(field, OpGreaterThan, value)
}

// Gte 大于等于条件
func Gte(field string, value interface{}) *Condition {
	return NewCondition(field, OpGreaterOrEqual, value)
}

// Lt 小于条件
func Lt(field string, value interface{}) *Condition {
	return NewCondition(field, OpLessThan, value)
}

// Lte 小于等于条件
func Lte(field string, value interface{}) *Condition {
	return NewCondition(field, OpLessOrEqual, value)
}

// Like 模糊匹配条件
func Like(field string, value string) *Condition {
	return NewCondition(field, OpLike, value)
}

// In 包含条件
func In(field string, values interface{}) *Condition {
	return NewCondition(field, OpIn, values)
}

// NotIn 不包含条件
func NotIn(field string, values interface{}) *Condition {
	return NewCondition(field, OpNotIn, values)
}

// Between 区间条件
func Between(field string, start, end interface{}) *Condition {
	return NewCondition(field, OpBetween, []interface{}{start, end})
}

// IsNull 为空条件
func IsNull(field string) *Condition {
	return NewCondition(field, OpIsNull, nil)
}

// IsNotNull 不为空条件
func IsNotNull(field string) *Condition {
	return NewCondition(field, OpIsNotNull, nil)
}

// QueryableRepository 可查询仓储接口，提供条件查询能力
type QueryableRepository[T any, ID comparable] interface {
	BaseRepository[T, ID]
	// Where 条件查询
	Where(ctx context.Context, conditions ...*Condition) ([]*T, error)

	// Count 统计数量
	Count(ctx context.Context, conditions ...*Condition) (int64, error)

	// Exists 存在性检查
	Exists(ctx context.Context, conditions ...*Condition) (bool, error)

	// Query 获取查询构建器
	Query() QueryBuilder[T]
}
`
	if err := g.writeFile("share/repository/queryable.go", repoQueryableTmpl); err != nil {
		return err
	}

	// repository/gorm/base_entity.go
	gormBaseEntityTmpl := `package gorm

import (
	"time"

	"gorm.io/gorm"
)

// BaseEntity 基础实体，包含通用的审计字段
// 业务实体通过组合方式继承这些字段
type BaseEntity struct {
	ID        int            ` + "`gorm:\"primaryKey;autoIncrement\" json:\"id\"`" + `
	CreatedAt time.Time      ` + "`gorm:\"autoCreateTime\" json:\"created_at\"`" + `
	UpdatedAt time.Time      ` + "`gorm:\"autoUpdateTime\" json:\"updated_at\"`" + `
	DeletedAt gorm.DeletedAt ` + "`gorm:\"index\" json:\"deleted_at,omitempty\"`" + `
	Version   int            ` + "`gorm:\"default:1\" json:\"version\"`" + `
}

// GetID 获取实体主键
func (e *BaseEntity) GetID() int {
	return e.ID
}

// IsDeleted 判断是否已删除
func (e *BaseEntity) IsDeleted() bool {
	return e.DeletedAt.Valid
}

// SetCreatedAt 设置创建时间
func (e *BaseEntity) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

// SetUpdatedAt 设置更新时间
func (e *BaseEntity) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = t
}

// IncrementVersion 版本号递增
func (e *BaseEntity) IncrementVersion() {
	e.Version++
}

// GetVersion 获取版本号
func (e *BaseEntity) GetVersion() int {
	return e.Version
}

// AuditFields 审计字段，不包含 ID，可供自定义主键类型的实体组合使用
type AuditFields struct {
	CreatedAt time.Time      ` + "`gorm:\"autoCreateTime\" json:\"created_at\"`" + `
	UpdatedAt time.Time      ` + "`gorm:\"autoUpdateTime\" json:\"updated_at\"`" + `
	DeletedAt gorm.DeletedAt ` + "`gorm:\"index\" json:\"deleted_at,omitempty\"`" + `
	Version   int            ` + "`gorm:\"default:1\" json:\"version\"`" + `
}

// IsDeleted 判断是否已删除
func (e *AuditFields) IsDeleted() bool {
	return e.DeletedAt.Valid
}

// SetCreatedAt 设置创建时间
func (e *AuditFields) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

// SetUpdatedAt 设置更新时间
func (e *AuditFields) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = t
}

// IncrementVersion 版本号递增
func (e *AuditFields) IncrementVersion() {
	e.Version++
}

// GetVersion 获取版本号
func (e *AuditFields) GetVersion() int {
	return e.Version
}

// Touch 更新修改时间
func (e *AuditFields) Touch() {
	e.UpdatedAt = time.Now()
}

// Auditable 可审计接口，实现此接口的实体将自动填充审计字段
type Auditable interface {
	SetCreatedAt(t time.Time)
	SetUpdatedAt(t time.Time)
	IncrementVersion()
	GetVersion() int
}
`
	if err := g.writeFile("share/repository/gorm/base_entity.go", gormBaseEntityTmpl); err != nil {
		return err
	}

	// repository/gorm/hooks.go
	gormHooksTmpl := `package gorm

import (
	"time"

	"gorm.io/gorm"
)

// BeforeCreate GORM 创建前钩子
// 自动设置创建时间、更新时间和版本号
func (e *BaseEntity) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	if e.UpdatedAt.IsZero() {
		e.UpdatedAt = now
	}
	if e.Version == 0 {
		e.Version = 1
	}
	return nil
}

// BeforeUpdate GORM 更新前钩子
// 自动更新更新时间和版本号
func (e *BaseEntity) BeforeUpdate(tx *gorm.DB) error {
	e.UpdatedAt = time.Now()
	e.Version++
	return nil
}

// RegisterAuditCallbacks 注册审计回调到 GORM
// 为所有实现 Auditable 接口的实体自动填充审计字段
func RegisterAuditCallbacks(db *gorm.DB) {
	// 创建前回调
	db.Callback().Create().Before("gorm:create").Register("audit:before_create", func(tx *gorm.DB) {
		if tx.Statement.Schema == nil {
			return
		}

		now := time.Now()

		// 设置创建时间
		if field := tx.Statement.Schema.LookUpField("CreatedAt"); field != nil {
			if _, isZero := field.ValueOf(tx.Statement.Context, tx.Statement.ReflectValue); isZero {
				_ = field.Set(tx.Statement.Context, tx.Statement.ReflectValue, now)
			}
		}

		// 设置更新时间
		if field := tx.Statement.Schema.LookUpField("UpdatedAt"); field != nil {
			if _, isZero := field.ValueOf(tx.Statement.Context, tx.Statement.ReflectValue); isZero {
				_ = field.Set(tx.Statement.Context, tx.Statement.ReflectValue, now)
			}
		}

		// 设置版本号
		if field := tx.Statement.Schema.LookUpField("Version"); field != nil {
			if val, isZero := field.ValueOf(tx.Statement.Context, tx.Statement.ReflectValue); isZero || val == 0 {
				_ = field.Set(tx.Statement.Context, tx.Statement.ReflectValue, 1)
			}
		}
	})

	// 更新前回调
	db.Callback().Update().Before("gorm:update").Register("audit:before_update", func(tx *gorm.DB) {
		if tx.Statement.Schema == nil {
			return
		}

		// 设置更新时间
		if field := tx.Statement.Schema.LookUpField("UpdatedAt"); field != nil {
			_ = field.Set(tx.Statement.Context, tx.Statement.ReflectValue, time.Now())
		}

		// 版本号递增（乐观锁）
		if field := tx.Statement.Schema.LookUpField("Version"); field != nil {
			if val, _ := field.ValueOf(tx.Statement.Context, tx.Statement.ReflectValue); val != nil {
				if version, ok := val.(int); ok {
					_ = field.Set(tx.Statement.Context, tx.Statement.ReflectValue, version+1)
				}
			}
		}
	})
}
`
	if err := g.writeFile("share/repository/gorm/hooks.go", gormHooksTmpl); err != nil {
		return err
	}

	// repository/gorm/factory.go
	gormFactoryTmpl := `package gorm

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseType 数据库类型
type DatabaseType string

const (
	MySQL      DatabaseType = "mysql"
	PostgreSQL DatabaseType = "postgres"
	SQLite     DatabaseType = "sqlite"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type            DatabaseType    // 数据库类型
	Host            string          // 主机地址
	Port            int             // 端口
	Database        string          // 数据库名
	Username        string          // 用户名
	Password        string          // 密码
	Charset         string          // 字符集（MySQL）
	SSLMode         string          // SSL 模式（PostgreSQL）
	MaxIdleConns    int             // 最大空闲连接数
	MaxOpenConns    int             // 最大打开连接数
	ConnMaxLifetime time.Duration   // 连接最大生命周期
	ConnMaxIdleTime time.Duration   // 连接最大空闲时间
	LogLevel        logger.LogLevel // 日志级别
	SlowThreshold   time.Duration   // 慢查询阈值
}

// DefaultConfig 默认配置
func DefaultConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Type:            PostgreSQL,
		Host:            "localhost",
		Port:            5432,
		Charset:         "utf8mb4",
		SSLMode:         "disable",
		MaxIdleConns:    10,
		MaxOpenConns:    100,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
		LogLevel:        logger.Info,
		SlowThreshold:   200 * time.Millisecond,
	}
}

// DatabaseFactory 数据库工厂
type DatabaseFactory struct {
	config *DatabaseConfig
}

// NewDatabaseFactory 创建数据库工厂
func NewDatabaseFactory(config *DatabaseConfig) *DatabaseFactory {
	if config == nil {
		config = DefaultConfig()
	}
	return &DatabaseFactory{config: config}
}

// Create 创建数据库连接
func (f *DatabaseFactory) Create() (*gorm.DB, error) {
	dialector, err := f.getDialector()
	if err != nil {
		return nil, err
	}

	// GORM 配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(f.config.LogLevel),
	}

	// 打开数据库连接
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层 SQL 连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(f.config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(f.config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(f.config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(f.config.ConnMaxIdleTime)

	// 注册审计回调
	RegisterAuditCallbacks(db)

	return db, nil
}

// getDialector 根据数据库类型获取 Dialector
func (f *DatabaseFactory) getDialector() (gorm.Dialector, error) {
	switch f.config.Type {
	case MySQL:
		return f.getMySQLDialector(), nil
	case PostgreSQL:
		return f.getPostgresDialector(), nil
	case SQLite:
		return f.getSQLiteDialector(), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", f.config.Type)
	}
}

// getMySQLDialector 获取 MySQL Dialector
func (f *DatabaseFactory) getMySQLDialector() gorm.Dialector {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		f.config.Username,
		f.config.Password,
		f.config.Host,
		f.config.Port,
		f.config.Database,
		f.config.Charset,
	)
	return mysql.Open(dsn)
}

// getPostgresDialector 获取 PostgreSQL Dialector
func (f *DatabaseFactory) getPostgresDialector() gorm.Dialector {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		f.config.Host,
		f.config.Port,
		f.config.Username,
		f.config.Password,
		f.config.Database,
		f.config.SSLMode,
	)
	return postgres.Open(dsn)
}

// getSQLiteDialector 获取 SQLite Dialector
func (f *DatabaseFactory) getSQLiteDialector() gorm.Dialector {
	return sqlite.Open(f.config.Database)
}

// CreateWithDSN 使用 DSN 创建数据库连接
func CreateWithDSN(dbType DatabaseType, dsn string, config *DatabaseConfig) (*gorm.DB, error) {
	if config == nil {
		config = DefaultConfig()
	}

	var dialector gorm.Dialector
	switch dbType {
	case MySQL:
		dialector = mysql.Open(dsn)
	case PostgreSQL:
		dialector = postgres.Open(dsn)
	case SQLite:
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	// GORM 配置
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(config.LogLevel),
	}

	// 打开数据库连接
	db, err := gorm.Open(dialector, gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 获取底层 SQL 连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// 配置连接池
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// 注册审计回调
	RegisterAuditCallbacks(db)

	return db, nil
}

// AutoMigrate 自动迁移表结构
func AutoMigrate(db *gorm.DB, models ...interface{}) error {
	return db.AutoMigrate(models...)
}
`
	if err := g.writeFile("share/repository/gorm/factory.go", gormFactoryTmpl); err != nil {
		return err
	}

	// repository/gorm/repository.go
	gormRepositoryTmpl := `package gorm

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"{{.ModulePath}}/share/repository"
)

// 事务上下文键
type txKey struct{}

// GormRepository 基于 GORM 的通用仓储实现
type GormRepository[T any, ID comparable] struct {
	db *gorm.DB
}

// NewGormRepository 创建 GORM 仓储实例
func NewGormRepository[T any, ID comparable](db *gorm.DB) *GormRepository[T, ID] {
	return &GormRepository[T, ID]{
		db: db,
	}
}

// DB 获取底层 GORM DB 实例
func (r *GormRepository[T, ID]) DB() *gorm.DB {
	return r.db
}

// getDB 获取数据库连接（支持事务）
func (r *GormRepository[T, ID]) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

// Create 创建单个实体
func (r *GormRepository[T, ID]) Create(ctx context.Context, entity *T) error {
	return r.getDB(ctx).Create(entity).Error
}

// CreateBatch 批量创建实体
func (r *GormRepository[T, ID]) CreateBatch(ctx context.Context, entities []*T) error {
	if len(entities) == 0 {
		return nil
	}
	return r.getDB(ctx).Create(entities).Error
}

// GetByID 根据主键查询
func (r *GormRepository[T, ID]) GetByID(ctx context.Context, id ID) (*T, error) {
	var entity T
	err := r.getDB(ctx).First(&entity, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// Update 更新实体
func (r *GormRepository[T, ID]) Update(ctx context.Context, entity *T) error {
	return r.getDB(ctx).Save(entity).Error
}

// Delete 删除实体（逻辑删除）
func (r *GormRepository[T, ID]) Delete(ctx context.Context, id ID) error {
	var entity T
	return r.getDB(ctx).Delete(&entity, id).Error
}

// List 查询全部列表
func (r *GormRepository[T, ID]) List(ctx context.Context) ([]*T, error) {
	var entities []*T
	err := r.getDB(ctx).Find(&entities).Error
	if err != nil {
		return nil, err
	}
	return entities, nil
}

// Page 分页查询
func (r *GormRepository[T, ID]) Page(ctx context.Context, request *repository.PageRequest) (*repository.PageResult[*T], error) {
	db := r.getDB(ctx)

	// 应用查询条件
	if len(request.Conditions) > 0 {
		db = ApplyConditions(db, request.Conditions...)
	}

	// 统计总数
	var total int64
	var entity T
	if err := db.Model(&entity).Count(&total).Error; err != nil {
		return nil, err
	}

	// 应用排序
	for _, order := range request.OrderBy {
		if order.Desc {
			db = db.Order(order.Field + " DESC")
		} else {
			db = db.Order(order.Field + " ASC")
		}
	}

	// 应用分页
	db = db.Offset(request.Offset()).Limit(request.Size)

	// 查询数据
	var entities []*T
	if err := db.Find(&entities).Error; err != nil {
		return nil, err
	}

	return repository.NewPageResult(entities, total, request.Page, request.Size), nil
}

// BeginTx 开启事务
func (r *GormRepository[T, ID]) BeginTx(ctx context.Context) (context.Context, error) {
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return ctx, tx.Error
	}
	return context.WithValue(ctx, txKey{}, tx), nil
}

// Commit 提交事务
func (r *GormRepository[T, ID]) Commit(ctx context.Context) error {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.Commit().Error
	}
	return errors.New("no transaction in context")
}

// Rollback 回滚事务
func (r *GormRepository[T, ID]) Rollback(ctx context.Context) error {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.Rollback().Error
	}
	return errors.New("no transaction in context")
}

// WithTx 在事务中执行操作
func (r *GormRepository[T, ID]) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	txCtx, err := r.BeginTx(ctx)
	if err != nil {
		return err
	}

	if err := fn(txCtx); err != nil {
		_ = r.Rollback(txCtx)
		return err
	}

	return r.Commit(txCtx)
}

// 确保实现了接口
var _ repository.BaseRepository[any, int] = (*GormRepository[any, int])(nil)
var _ repository.TransactionalRepository = (*GormRepository[any, int])(nil)
`
	if err := g.renderAndWrite(gormRepositoryTmpl, "share/repository/gorm/repository.go"); err != nil {
		return err
	}

	// repository/gorm/queryable.go
	gormQueryableTmpl := `package gorm

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"{{.ModulePath}}/share/repository"
)

// QueryableGormRepository 可查询的 GORM 仓储实现
type QueryableGormRepository[T any, ID comparable] struct {
	*GormRepository[T, ID]
}

// NewQueryableGormRepository 创建可查询的 GORM 仓储实例
func NewQueryableGormRepository[T any, ID comparable](db *gorm.DB) *QueryableGormRepository[T, ID] {
	return &QueryableGormRepository[T, ID]{
		GormRepository: NewGormRepository[T, ID](db),
	}
}

// Where 条件查询
func (r *QueryableGormRepository[T, ID]) Where(ctx context.Context, conditions ...*repository.Condition) ([]*T, error) {
	db := ApplyConditions(r.getDB(ctx), conditions...)
	var entities []*T
	if err := db.Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// Count 统计数量
func (r *QueryableGormRepository[T, ID]) Count(ctx context.Context, conditions ...*repository.Condition) (int64, error) {
	db := ApplyConditions(r.getDB(ctx), conditions...)
	var count int64
	var entity T
	if err := db.Model(&entity).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Exists 存在性检查
func (r *QueryableGormRepository[T, ID]) Exists(ctx context.Context, conditions ...*repository.Condition) (bool, error) {
	count, err := r.Count(ctx, conditions...)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Query 获取查询构建器
func (r *QueryableGormRepository[T, ID]) Query() repository.QueryBuilder[T] {
	return NewGormQueryBuilder[T](r.GormRepository.DB())
}

// ApplyConditions 将条件列表应用到 GORM 查询（包级函数）
func ApplyConditions(db *gorm.DB, conditions ...*repository.Condition) *gorm.DB {
	for _, cond := range conditions {
		db = ApplyCondition(db, cond)
	}
	return db
}

// ApplyCondition 应用单个条件到 GORM 查询（包级函数）
func ApplyCondition(db *gorm.DB, cond *repository.Condition) *gorm.DB {
	switch cond.Operator {
	case repository.OpEqual:
		return db.Where(fmt.Sprintf("%s = ?", cond.Field), cond.Value)
	case repository.OpNotEqual:
		return db.Where(fmt.Sprintf("%s != ?", cond.Field), cond.Value)
	case repository.OpGreaterThan:
		return db.Where(fmt.Sprintf("%s > ?", cond.Field), cond.Value)
	case repository.OpGreaterOrEqual:
		return db.Where(fmt.Sprintf("%s >= ?", cond.Field), cond.Value)
	case repository.OpLessThan:
		return db.Where(fmt.Sprintf("%s < ?", cond.Field), cond.Value)
	case repository.OpLessOrEqual:
		return db.Where(fmt.Sprintf("%s <= ?", cond.Field), cond.Value)
	case repository.OpLike:
		return db.Where(fmt.Sprintf("%s LIKE ?", cond.Field), cond.Value)
	case repository.OpIn:
		return db.Where(fmt.Sprintf("%s IN ?", cond.Field), cond.Value)
	case repository.OpNotIn:
		return db.Where(fmt.Sprintf("%s NOT IN ?", cond.Field), cond.Value)
	case repository.OpBetween:
		if values, ok := cond.Value.([]interface{}); ok && len(values) == 2 {
			return db.Where(fmt.Sprintf("%s BETWEEN ? AND ?", cond.Field), values[0], values[1])
		}
		return db
	case repository.OpIsNull:
		return db.Where(fmt.Sprintf("%s IS NULL", cond.Field))
	case repository.OpIsNotNull:
		return db.Where(fmt.Sprintf("%s IS NOT NULL", cond.Field))
	default:
		return db
	}
}

// GormQueryBuilder GORM 查询构建器实现
type GormQueryBuilder[T any] struct {
	db      *gorm.DB
	options *repository.QueryOptions
}

// NewGormQueryBuilder 创建 GORM 查询构建器
func NewGormQueryBuilder[T any](db *gorm.DB) *GormQueryBuilder[T] {
	return &GormQueryBuilder[T]{
		db:      db,
		options: repository.NewQueryOptions(),
	}
}

// Where 添加查询条件
func (b *GormQueryBuilder[T]) Where(condition *repository.Condition) repository.QueryBuilder[T] {
	b.options.AddCondition(condition)
	return b
}

// And 添加 AND 条件
func (b *GormQueryBuilder[T]) And(conditions ...*repository.Condition) repository.QueryBuilder[T] {
	b.options.AddConditions(conditions...)
	return b
}

// OrderBy 添加排序（升序）
func (b *GormQueryBuilder[T]) OrderBy(field string) repository.QueryBuilder[T] {
	b.options.AddOrderBy(field, false)
	return b
}

// OrderByDesc 添加排序（降序）
func (b *GormQueryBuilder[T]) OrderByDesc(field string) repository.QueryBuilder[T] {
	b.options.AddOrderBy(field, true)
	return b
}

// Limit 限制返回数量
func (b *GormQueryBuilder[T]) Limit(limit int) repository.QueryBuilder[T] {
	b.options.SetLimit(limit)
	return b
}

// Offset 设置偏移量
func (b *GormQueryBuilder[T]) Offset(offset int) repository.QueryBuilder[T] {
	b.options.SetOffset(offset)
	return b
}

// Select 指定查询字段
func (b *GormQueryBuilder[T]) Select(fields ...string) repository.QueryBuilder[T] {
	b.options.SetFields(fields...)
	return b
}

// build 构建 GORM 查询
func (b *GormQueryBuilder[T]) build(ctx context.Context) *gorm.DB {
	db := b.db.WithContext(ctx)

	// 应用查询条件
	for _, cond := range b.options.Conditions {
		db = ApplyCondition(db, cond)
	}

	// 应用字段选择
	if len(b.options.Fields) > 0 {
		db = db.Select(b.options.Fields)
	}

	// 应用排序
	for _, order := range b.options.OrderBys {
		if order.Desc {
			db = db.Order(order.Field + " DESC")
		} else {
			db = db.Order(order.Field + " ASC")
		}
	}

	// 应用分页
	if b.options.LimitVal > 0 {
		db = db.Limit(b.options.LimitVal)
	}
	if b.options.OffsetVal > 0 {
		db = db.Offset(b.options.OffsetVal)
	}

	return db
}

// Find 执行查询，返回结果列表
func (b *GormQueryBuilder[T]) Find(ctx context.Context) ([]*T, error) {
	var entities []*T
	if err := b.build(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// First 执行查询，返回第一条结果
func (b *GormQueryBuilder[T]) First(ctx context.Context) (*T, error) {
	var entity T
	if err := b.build(ctx).First(&entity).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &entity, nil
}

// Count 执行统计查询
func (b *GormQueryBuilder[T]) Count(ctx context.Context) (int64, error) {
	var count int64
	var entity T
	db := b.db.WithContext(ctx)

	// 只应用查询条件
	for _, cond := range b.options.Conditions {
		db = ApplyCondition(db, cond)
	}

	if err := db.Model(&entity).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Exists 执行存在性检查
func (b *GormQueryBuilder[T]) Exists(ctx context.Context) (bool, error) {
	count, err := b.Count(ctx)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Page 执行分页查询
func (b *GormQueryBuilder[T]) Page(ctx context.Context, page, size int) (*repository.PageResult[*T], error) {
	// 统计总数
	total, err := b.Count(ctx)
	if err != nil {
		return nil, err
	}

	// 设置分页参数
	b.options.SetOffset((page - 1) * size)
	b.options.SetLimit(size)

	// 查询数据
	entities, err := b.Find(ctx)
	if err != nil {
		return nil, err
	}

	return repository.NewPageResult(entities, total, page, size), nil
}

// 确保实现了接口
var _ repository.QueryableRepository[any, int] = (*QueryableGormRepository[any, int])(nil)
var _ repository.QueryBuilder[any] = (*GormQueryBuilder[any])(nil)
`
	if err := g.renderAndWrite(gormQueryableTmpl, "share/repository/gorm/queryable.go"); err != nil {
		return err
	}

	return nil
}
