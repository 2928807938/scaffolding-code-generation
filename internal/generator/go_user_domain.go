package generator

// generateUserDomain 生成 user/domain 模块
func (g *GoGenerator) generateUserDomain() error {
	// go.mod
	goModTmpl := `module {{.ModulePath}}/user/domain

go 1.24.11

require (
	{{.ModulePath}}/bom v0.0.0

	// 通用工具
	github.com/google/uuid v1.6.0
)

replace {{.ModulePath}}/bom => ../../bom
`
	if err := g.renderAndWrite(goModTmpl, "user/domain/go.mod"); err != nil {
		return err
	}

	// entity/user.go
	userEntityTmpl := `package entity

import (
	"time"

	"github.com/google/uuid"
	"{{.ModulePath}}/user/domain/valueobject"
)

// UserStatus 用户状态
type UserStatus int

const (
	UserStatusInactive UserStatus = 0 // 未激活
	UserStatusActive   UserStatus = 1 // 已激活
	UserStatusDisabled UserStatus = 2 // 已禁用
)

// User 用户实体 - 聚合根
type User struct {
	ID           uuid.UUID
	Username     string
	Email        valueobject.Email
	PasswordHash string
	Status       UserStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser 创建新用户
func NewUser(username string, email valueobject.Email, passwordHash string) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New(),
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
		Status:       UserStatusInactive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Activate 激活用户
func (u *User) Activate() {
	u.Status = UserStatusActive
	u.UpdatedAt = time.Now()
}

// Disable 禁用用户
func (u *User) Disable() {
	u.Status = UserStatusDisabled
	u.UpdatedAt = time.Now()
}

// ChangePassword 修改密码
func (u *User) ChangePassword(newPasswordHash string) {
	u.PasswordHash = newPasswordHash
	u.UpdatedAt = time.Now()
}

// UpdateEmail 更新邮箱
func (u *User) UpdateEmail(email valueobject.Email) {
	u.Email = email
	u.UpdatedAt = time.Now()
}

// IsActive 判断用户是否激活
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}
`
	if err := g.renderAndWrite(userEntityTmpl, "user/domain/entity/user.go"); err != nil {
		return err
	}

	// repository/user_repository.go
	userRepoTmpl := `package repository

import (
	"context"

	"github.com/google/uuid"
	"{{.ModulePath}}/user/domain/entity"
)

// UserRepository 用户仓储接口
type UserRepository interface {
	// Save 保存用户
	Save(ctx context.Context, user *entity.User) error

	// FindByID 根据 ID 查找用户
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// FindByEmail 根据邮箱查找用户
	FindByEmail(ctx context.Context, email string) (*entity.User, error)

	// FindByUsername 根据用户名查找用户
	FindByUsername(ctx context.Context, username string) (*entity.User, error)

	// Update 更新用户
	Update(ctx context.Context, user *entity.User) error

	// Delete 删除用户
	Delete(ctx context.Context, id uuid.UUID) error

	// List 分页查询用户列表
	List(ctx context.Context, page, pageSize int) ([]*entity.User, int64, error)

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
	"errors"

	"{{.ModulePath}}/user/domain/entity"
	"{{.ModulePath}}/user/domain/repository"
	"{{.ModulePath}}/user/domain/valueobject"
)

var (
	ErrEmailAlreadyExists    = errors.New("邮箱已被使用")
	ErrUsernameAlreadyExists = errors.New("用户名已被使用")
	ErrInvalidEmail          = errors.New("邮箱格式不正确")
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
		return nil, ErrInvalidEmail
	}

	// 检查邮箱是否已存在
	exists, err := s.userRepo.ExistsByEmail(ctx, emailStr)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// 检查用户名是否已存在
	exists, err = s.userRepo.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameAlreadyExists
	}

	// 创建用户实体
	user := entity.NewUser(username, email, passwordHash)

	return user, nil
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
	emailRegex            = regexp.MustCompile(` + "`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$`" + `)
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
