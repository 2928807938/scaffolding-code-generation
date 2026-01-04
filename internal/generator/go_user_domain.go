package generator

// generateUserDomain 生成 user/domain 模块
func (g *GoGenerator) generateUserDomain() error {
	// go.mod
	goModTmpl := `module {{.ModulePath}}/user/domain

go 1.24.11

require (
	{{.ModulePath}}/bom v0.0.0
	{{.ModulePath}}/share v0.0.0

	// 通用工具
	github.com/google/uuid v1.6.0
)

replace (
	{{.ModulePath}}/bom => ../../bom
	{{.ModulePath}}/share => ../../share
)
`
	if err := g.renderAndWrite(goModTmpl, "user/domain/go.mod"); err != nil {
		return err
	}

	// enum/user_status.go
	userStatusEnumTmpl := `package enum

// UserStatus 用户状态枚举
type UserStatus int

const (
	UserStatusInactive UserStatus = 0 // 未激活
	UserStatusActive   UserStatus = 1 // 已激活
	UserStatusDisabled UserStatus = 2 // 已禁用
)

// String 返回状态描述
func (s UserStatus) String() string {
	switch s {
	case UserStatusInactive:
		return "inactive"
	case UserStatusActive:
		return "active"
	case UserStatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

// IsValid 验证状态是否有效
func (s UserStatus) IsValid() bool {
	return s >= UserStatusInactive && s <= UserStatusDisabled
}

// IsActive 是否激活状态
func (s UserStatus) IsActive() bool {
	return s == UserStatusActive
}

// IsInactive 是否未激活状态
func (s UserStatus) IsInactive() bool {
	return s == UserStatusInactive
}

// IsDisabled 是否禁用状态
func (s UserStatus) IsDisabled() bool {
	return s == UserStatusDisabled
}
`
	if err := g.writeFile("user/domain/enum/user_status.go", userStatusEnumTmpl); err != nil {
		return err
	}

	// errors/user_error.go
	userErrorTmpl := `package errors

import (
	"fmt"

	"{{.ModulePath}}/share/errors"
)

// ==================== User 模块错误 ====================
// 错误码分段: 11xxx (11000-11999)

const (
	// User 模块错误码
	UserNotFound           = 11001 // 用户不存在
	UserAlreadyExists      = 11002 // 用户已存在
	UserInvalidPassword    = 11003 // 密码错误
	UserDisabled           = 11004 // 用户已被禁用
	UserExpired            = 11005 // 用户已过期
	UserInvalidEmail       = 11006 // 邮箱格式错误
	UserEmailAlreadyExists = 11007 // 邮箱已被使用
	UserUsernameExists     = 11008 // 用户名已被使用
)

// UserError User 模块错误，继承自 AppError
type UserError struct {
	*errors.AppError
}

// NewUserError 创建 User 错误
func NewUserError(code int, message string) *UserError {
	return &UserError{
		AppError: errors.New(code, message),
	}
}

// WrapUserError 包装原始错误
func WrapUserError(code int, message string, err error) *UserError {
	return &UserError{
		AppError: errors.Wrap(code, message, err),
	}
}

// Error 实现 error 接口
func (e *UserError) Error() string {
	if e.AppError.Err != nil {
		return fmt.Sprintf("[User:%d] %s: %v", e.AppError.Code, e.AppError.Message, e.AppError.Err)
	}
	return fmt.Sprintf("[User:%d] %s", e.AppError.Code, e.AppError.Message)
}

// ==================== User 预定义错误（message 已定义） ====================

var (
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = &UserError{
		AppError: errors.New(UserNotFound, "用户不存在"),
	}

	// ErrUserAlreadyExists 用户已存在
	ErrUserAlreadyExists = &UserError{
		AppError: errors.New(UserAlreadyExists, "用户已存在"),
	}

	// ErrUserInvalidPassword 密码错误
	ErrUserInvalidPassword = &UserError{
		AppError: errors.New(UserInvalidPassword, "密码错误"),
	}

	// ErrUserDisabled 用户已被禁用
	ErrUserDisabled = &UserError{
		AppError: errors.New(UserDisabled, "用户已被禁用"),
	}

	// ErrUserExpired 用户已过期
	ErrUserExpired = &UserError{
		AppError: errors.New(UserExpired, "用户已过期"),
	}

	// ErrUserInvalidEmail 邮箱格式不正确
	ErrUserInvalidEmail = &UserError{
		AppError: errors.New(UserInvalidEmail, "邮箱格式不正确"),
	}

	// ErrUserEmailAlreadyExists 邮箱已被使用
	ErrUserEmailAlreadyExists = &UserError{
		AppError: errors.New(UserEmailAlreadyExists, "邮箱已被使用"),
	}

	// ErrUserUsernameExists 用户名已被使用
	ErrUserUsernameExists = &UserError{
		AppError: errors.New(UserUsernameExists, "用户名已被使用"),
	}
)
`
	if err := g.renderAndWrite(userErrorTmpl, "user/domain/errors/user_error.go"); err != nil {
		return err
	}

	// entity/user.go
	userEntityTmpl := `package entity

import (
	"{{.ModulePath}}/user/domain/enum"
	"{{.ModulePath}}/user/domain/valueobject"

	"github.com/google/uuid"
	basegorm "{{.ModulePath}}/share/repository/gorm"
)

// User 用户实体 - 聚合根
type User struct {
	ID           uuid.UUID
	Username     string
	Email        valueobject.Email
	PasswordHash string
	Status       enum.UserStatus
	basegorm.AuditFields
}

// Activate 激活用户
func (u *User) Activate() {
	u.Status = enum.UserStatusActive
	u.Touch()
}

// Disable 禁用用户
func (u *User) Disable() {
	u.Status = enum.UserStatusDisabled
	u.Touch()
}

// ChangePassword 修改密码
func (u *User) ChangePassword(newPasswordHash string) {
	u.PasswordHash = newPasswordHash
	u.Touch()
}

// UpdateEmail 更新邮箱
func (u *User) UpdateEmail(email valueobject.Email) {
	u.Email = email
	u.Touch()
}

// IsActive 判断用户是否激活
func (u *User) IsActive() bool {
	return u.Status == enum.UserStatusActive
}
`
	if err := g.renderAndWrite(userEntityTmpl, "user/domain/entity/user.go"); err != nil {
		return err
	}

	// repository/user_repository.go
	userRepoTmpl := `package repository

import (
	"context"

	baseRepo "{{.ModulePath}}/share/repository"
	"{{.ModulePath}}/user/domain/entity"

	"github.com/google/uuid"
)

// UserRepository 用户仓储接口，继承可查询仓储
type UserRepository interface {
	// 继承可查询仓储（包含 CRUD、分页、条件查询等）
	baseRepo.QueryableRepository[entity.User, uuid.UUID]

	// FindByEmail 根据邮箱查找用户
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	// FindByUsername 根据用户名查找用户
	FindByUsername(ctx context.Context, username string) (*entity.User, error)

	// ExistsByEmail 检查邮箱是否存在
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByUsername 检查用户名是否存在
	ExistsByUsername(ctx context.Context, username string) (bool, error)
}
`
	if err := g.renderAndWrite(userRepoTmpl, "user/domain/repository/user_repository.go"); err != nil {
		return err
	}

	// service/user_domain_service.go
	userDomainServiceTmpl := `package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"{{.ModulePath}}/user/domain/enum"
	"{{.ModulePath}}/user/domain/errors"
	"{{.ModulePath}}/user/domain/entity"
	"{{.ModulePath}}/user/domain/repository"
	"{{.ModulePath}}/user/domain/valueobject"
)

// UserDomainService 用户领域服务
type UserDomainService struct {
	userRepo repository.UserRepository
}

// NewUserDomainService 创建用户领域服务
func NewUserDomainService(userRepo repository.UserRepository) *UserDomainService {
	return &UserDomainService{
		userRepo: userRepo,
	}
}

// CreateUser 创建用户（包含业务规则校验）
func (s *UserDomainService) CreateUser(ctx context.Context, username string, emailStr string, passwordHash string) (*entity.User, error) {
	// 验证邮箱格式
	email, err := valueobject.NewEmail(emailStr)
	if err != nil {
		return nil, errors.ErrUserInvalidEmail
	}

	// 检查邮箱是否已存在
	exists, err := s.userRepo.ExistsByEmail(ctx, emailStr)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrUserEmailAlreadyExists
	}

	// 检查用户名是否已存在
	exists, err = s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrUserUsernameExists
	}

	// 创建用户实体
	now := time.Now()
	user := &entity.User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Status:       enum.UserStatusInactive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return user, nil
}

// GetUser 获取用户（包含业务规则校验）
func (s *UserDomainService) GetUser(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}
	return user, nil
}

// UpdateUser 更新用户（包含业务规则校验）
func (s *UserDomainService) UpdateUser(ctx context.Context, id uuid.UUID, username string, status *enum.UserStatus) (*entity.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.ErrUserNotFound
	}

	// 更新字段（如果提供了新值）
	if username != "" {
		user.Username = username
	}
	if status != nil {
		user.Status = *status
	}
	user.Touch()

	return user, nil
}

// DeleteUser 删除用户（包含业务规则校验）
func (s *UserDomainService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.ErrUserNotFound
	}

	return s.userRepo.Delete(ctx, id)
}
`
	if err := g.renderAndWrite(userDomainServiceTmpl, "user/domain/service/user_domain_service.go"); err != nil {
		return err
	}

	// valueobject/email.go
	emailVOTmpl := `package valueobject

import (
	"errors"
	"regexp"
	"strings"
)

var (
	ErrInvalidEmailFormat = errors.New("无效的邮箱格式")
	emailRegex            = regexp.MustCompile(` + "`" + `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$` + "`" + `)
)

// Email 邮箱值对象
type Email string

// NewEmail 创建邮箱值对象
func NewEmail(email string) (Email, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if !emailRegex.MatchString(email) {
		return "", ErrInvalidEmailFormat
	}
	return Email(email), nil
}

// String 转换为字符串
func (e Email) String() string {
	return string(e)
}

// Domain 获取邮箱域名
func (e Email) Domain() string {
	parts := strings.Split(string(e), "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

// LocalPart 获取邮箱本地部分
func (e Email) LocalPart() string {
	parts := strings.Split(string(e), "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}
`
	if err := g.writeFile("user/domain/valueobject/email.go", emailVOTmpl); err != nil {
		return err
	}

	// valueobject/password.go
	passwordVOTmpl := `package valueobject

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword = errors.New("无效的密码")
	ErrPasswordMismatch = errors.New("密码不匹配")
)

// Password 密码值对象
type Password struct {
	hash string
}

// NewPassword 从明文创建密码值对象（自动加密）
// 使用 cost=12 提高安全性（默认是10，每增加1计算时间翻倍）
func NewPassword(plainPassword string) (*Password, error) {
	if plainPassword == "" {
		return nil, ErrInvalidPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), 12)
	if err != nil {
		return nil, err
	}
	return &Password{hash: string(hash)}, nil
}

// NewPasswordFromHash 从哈希值创建密码值对象
func NewPasswordFromHash(hash string) (*Password, error) {
	if hash == "" {
		return nil, ErrInvalidPassword
	}
	return &Password{hash: hash}, nil
}

// String 返回哈希值（注意：不返回明文密码）
func (p *Password) String() string {
	return p.hash
}

// Hash 返回密码哈希
func (p *Password) Hash() string {
	return p.hash
}

// Verify 验证密码是否匹配
func (p *Password) Verify(plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(p.hash), []byte(plainPassword))
	return err == nil
}
`
	if err := g.writeFile("user/domain/valueobject/password.go", passwordVOTmpl); err != nil {
		return err
	}

	// event/user_events.go
	userEventsTmpl := `package event

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent 领域事件接口
type DomainEvent interface {
	EventName() string
	OccurredAt() time.Time
}

// UserCreatedEvent 用户创建事件
type UserCreatedEvent struct {
	UserID     uuid.UUID
	Username   string
	Email      string
	occurredAt time.Time
}

func NewUserCreatedEvent(userID uuid.UUID, username, email string) *UserCreatedEvent {
	return &UserCreatedEvent{
		UserID:     userID,
		Username:   username,
		Email:      email,
		occurredAt: time.Now(),
	}
}

func (e *UserCreatedEvent) EventName() string {
	return "user.created"
}

func (e *UserCreatedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// UserActivatedEvent 用户激活事件
type UserActivatedEvent struct {
	UserID     uuid.UUID
	occurredAt time.Time
}

func NewUserActivatedEvent(userID uuid.UUID) *UserActivatedEvent {
	return &UserActivatedEvent{
		UserID:     userID,
		occurredAt: time.Now(),
	}
}

func (e *UserActivatedEvent) EventName() string {
	return "user.activated"
}

func (e *UserActivatedEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// UserPasswordChangedEvent 用户密码修改事件
type UserPasswordChangedEvent struct {
	UserID     uuid.UUID
	occurredAt time.Time
}

func NewUserPasswordChangedEvent(userID uuid.UUID) *UserPasswordChangedEvent {
	return &UserPasswordChangedEvent{
		UserID:     userID,
		occurredAt: time.Now(),
	}
}

func (e *UserPasswordChangedEvent) EventName() string {
	return "user.password_changed"
}

func (e *UserPasswordChangedEvent) OccurredAt() time.Time {
	return e.occurredAt
}
`
	if err := g.writeFile("user/domain/event/user_events.go", userEventsTmpl); err != nil {
		return err
	}

	return nil
}
