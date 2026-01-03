package generator

// generateUserInfra 生成 user/infrastructure 模块
func (g *GoGenerator) generateUserInfra() error {
	// go.mod
	goModTmpl := `module {{.ModulePath}}/user/infrastructure

go 1.24.11

require (
	{{.ModulePath}}/bom v0.0.0
	{{.ModulePath}}/user/domain v0.0.0

	// 通用工具
	github.com/google/uuid v1.6.0

	// 数据库
	gorm.io/gorm v1.25.12
)

replace (
	{{.ModulePath}}/bom => ../../bom
	{{.ModulePath}}/user/domain => ../domain
)
`
	if err := g.renderAndWrite(goModTmpl, "user/infrastructure/go.mod"); err != nil {
		return err
	}

	// entity/user_po.go
	userPOTmpl := `package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserPO 用户持久化对象
type UserPO struct {
	ID           uuid.UUID ` + "`gorm:\"type:uuid;primaryKey\"`" + `
	Username     string    ` + "`gorm:\"type:varchar(50);uniqueIndex;not null\"`" + `
	Email        string    ` + "`gorm:\"type:varchar(100);uniqueIndex;not null\"`" + `
	PasswordHash string    ` + "`gorm:\"type:varchar(255);not null\"`" + `
	Status       int       ` + "`gorm:\"type:int;default:0\"`" + `
	CreatedAt    time.Time ` + "`gorm:\"autoCreateTime\"`" + `
	UpdatedAt    time.Time ` + "`gorm:\"autoUpdateTime\"`" + `
}

// TableName 指定表名
func (UserPO) TableName() string {
	return "users"
}
`
	if err := g.writeFile("user/infrastructure/entity/user_po.go", userPOTmpl); err != nil {
		return err
	}

	// converter/user_converter.go
	converterTmpl := `package converter

import (
	infraEntity "{{.ModulePath}}/user/infrastructure/entity"
	"{{.ModulePath}}/user/domain/entity"
	"{{.ModulePath}}/user/domain/enum"
	"{{.ModulePath}}/user/domain/valueobject"
)

// UserConverter 用户转换器
type UserConverter struct{}

// NewUserConverter 创建用户转换器
func NewUserConverter() *UserConverter {
	return &UserConverter{}
}

// ToEntity 将 PO 转换为领域实体
func (c *UserConverter) ToEntity(po *infraEntity.UserPO) *entity.User {
	if po == nil {
		return nil
	}

	email, _ := valueobject.NewEmail(po.Email)

	return &entity.User{
		ID:           po.ID,
		Username:     po.Username,
		Email:        email,
		PasswordHash: po.PasswordHash,
		Status:       enum.UserStatus(po.Status),
		CreatedAt:    po.CreatedAt,
		UpdatedAt:    po.UpdatedAt,
	}
}

// ToPO 将领域实体转换为 PO
func (c *UserConverter) ToPO(user *entity.User) *infraEntity.UserPO {
	if user == nil {
		return nil
	}

	return &infraEntity.UserPO{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email.String(),
		PasswordHash: user.PasswordHash,
		Status:       int(user.Status),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}
`
	if err := g.renderAndWrite(converterTmpl, "user/infrastructure/converter/user_converter.go"); err != nil {
		return err
	}

	// repository/user_repository_impl.go
	repoImplTmpl := `package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"{{.ModulePath}}/user/domain/entity"
	"{{.ModulePath}}/user/domain/repository"
	"{{.ModulePath}}/user/infrastructure/converter"
	infraEntity "{{.ModulePath}}/user/infrastructure/entity"
)

// UserRepositoryImpl 用户仓储实现
type UserRepositoryImpl struct {
	db        *gorm.DB
	converter *converter.UserConverter
}

// NewUserRepositoryImpl 创建用户仓储实现
func NewUserRepositoryImpl(db *gorm.DB) repository.UserRepository {
	return &UserRepositoryImpl{
		db:        db,
		converter: converter.NewUserConverter(),
	}
}

// Save 保存用户
func (r *UserRepositoryImpl) Save(ctx context.Context, user *entity.User) error {
	po := r.converter.ToPO(user)
	return r.db.WithContext(ctx).Create(po).Error
}

// FindByID 根据 ID 查找用户
func (r *UserRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	var po infraEntity.UserPO
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.converter.ToEntity(&po), nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var po infraEntity.UserPO
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&po).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.converter.ToEntity(&po), nil
}

// FindByUsername 根据用户名查找用户
func (r *UserRepositoryImpl) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	var po infraEntity.UserPO
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&po).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.converter.ToEntity(&po), nil
}

// Update 更新用户
func (r *UserRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	po := r.converter.ToPO(user)
	return r.db.WithContext(ctx).Save(po).Error
}

// Delete 删除用户
func (r *UserRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&infraEntity.UserPO{}).Error
}

// List 分页查询用户列表
func (r *UserRepositoryImpl) List(ctx context.Context, page, pageSize int) ([]*entity.User, int64, error) {
	var total int64
	var poList []*infraEntity.UserPO

	// 查询总数
	if err := r.db.WithContext(ctx).Model(&infraEntity.UserPO{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Find(&poList).Error; err != nil {
		return nil, 0, err
	}

	// 转换为领域实体
	users := make([]*entity.User, len(poList))
	for i, po := range poList {
		users[i] = r.converter.ToEntity(po)
	}
	return users, total, nil
}

// ExistsByEmail 检查邮箱是否存在
func (r *UserRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&infraEntity.UserPO{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// ExistsByUsername 检查用户名是否存在
func (r *UserRepositoryImpl) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&infraEntity.UserPO{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}
`
	if err := g.renderAndWrite(repoImplTmpl, "user/infrastructure/repository/user_repository_impl.go"); err != nil {
		return err
	}

	return nil
}

// generateUserModule 生成 user 聚合模块
func (g *GoGenerator) generateUserModule() error {
	// go.mod
	goModTmpl := `module {{.ModulePath}}/user

go 1.24.11

replace (
	{{.ModulePath}}/user/domain => ./domain
	{{.ModulePath}}/user/infrastructure => ./infrastructure
)
`
	return g.renderAndWrite(goModTmpl, "user/go.mod")
}
