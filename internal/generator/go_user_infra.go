package generator

// generateUserInfra 生成 user/infrastructure 模块
func (g *GoGenerator) generateUserInfra() error {
	// go.mod
	goModTmpl := `module {{.ModulePath}}/user/infrastructure

go 1.24.11

require (
	{{.ModulePath}}/bom v0.0.0
	{{.ModulePath}}/share v0.0.0
	{{.ModulePath}}/user/domain v0.0.0

	// 通用工具
	github.com/google/uuid v1.6.0

	// 数据库
	gorm.io/gorm v1.25.12
)

replace (
	{{.ModulePath}}/bom => ../../bom
	{{.ModulePath}}/share => ../../share
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

// UserPO 用户持久化对象，与数据库表字段对应
type UserPO struct {
	ID           uuid.UUID ` + "`gorm:\"type:uuid;primaryKey\"`" + `
	Username     string    ` + "`gorm:\"type:varchar(50);uniqueIndex;not null\"`" + `
	Email        string    ` + "`gorm:\"type:varchar(100);uniqueIndex;not null\"`" + `
	PasswordHash string    ` + "`gorm:\"type:varchar(255);not null\"`" + `
	Status       int       ` + "`gorm:\"type:int;default:0\"`" + `

	// 审计字段 - 与数据库表字段对应
	CreatedAt time.Time ` + "`gorm:\"autoCreateTime\" json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`gorm:\"autoUpdateTime\" json:\"updated_at\"`" + `
	DeletedAt time.Time ` + "`gorm:\"index\" json:\"deleted_at,omitempty\"`" + `
	Version   int       ` + "`gorm:\"default:1\" json:\"version\"`" + `
}

// TableName 指定表名
func (UserPO) TableName() string {
	return "users"
}

// GetID 获取实体主键
func (u *UserPO) GetID() uuid.UUID {
	return u.ID
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

	basegorm "{{.ModulePath}}/share/repository/gorm"
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
		AuditFields: basegorm.AuditFields{
			CreatedAt: po.CreatedAt,
			UpdatedAt: po.UpdatedAt,
			Version:   po.Version,
		},
	}
}

// ToPO 将领域实体转换为 PO
func (c *UserConverter) ToPO(user *entity.User) *infraEntity.UserPO {
	if user == nil {
		return nil
	}

	po := &infraEntity.UserPO{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email.String(),
		PasswordHash: user.PasswordHash,
		Status:       int(user.Status),
	}
	po.CreatedAt = user.CreatedAt
	po.UpdatedAt = user.UpdatedAt
	po.Version = user.Version
	return po
}
`
	if err := g.renderAndWrite(converterTmpl, "user/infrastructure/converter/user_converter.go"); err != nil {
		return err
	}

	// repository/user_repository_impl.go
	repoImplTmpl := `package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"{{.ModulePath}}/share/repository"
	basegorm "{{.ModulePath}}/share/repository/gorm"
	"{{.ModulePath}}/user/domain/entity"
	domainRepo "{{.ModulePath}}/user/domain/repository"
	"{{.ModulePath}}/user/infrastructure/converter"
	infraEntity "{{.ModulePath}}/user/infrastructure/entity"
)

// UserRepositoryImpl 用户仓储实现
type UserRepositoryImpl struct {
	repo      *basegorm.QueryableGormRepository[infraEntity.UserPO, uuid.UUID]
	converter *converter.UserConverter
}

// NewUserRepositoryImpl 创建用户仓储实现
func NewUserRepositoryImpl(db *gorm.DB) domainRepo.UserRepository {
	return &UserRepositoryImpl{
		repo:      basegorm.NewQueryableGormRepository[infraEntity.UserPO, uuid.UUID](db),
		converter: converter.NewUserConverter(),
	}
}

// Create 创建用户（实现 BaseRepository）
func (r *UserRepositoryImpl) Create(ctx context.Context, user *entity.User) error {
	po := r.converter.ToPO(user)
	return r.repo.Create(ctx, po)
}

// CreateBatch 批量创建用户（实现 BaseRepository）
func (r *UserRepositoryImpl) CreateBatch(ctx context.Context, users []*entity.User) error {
	if len(users) == 0 {
		return nil
	}
	pos := make([]*infraEntity.UserPO, len(users))
	for i, user := range users {
		pos[i] = r.converter.ToPO(user)
	}
	return r.repo.CreateBatch(ctx, pos)
}

// GetByID 根据 ID 查找用户（实现 BaseRepository）
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	po, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if po == nil {
		return nil, nil
	}
	return r.converter.ToEntity(po), nil
}

// FindByID 根据 ID 查找用户（别名方法，保持兼容）
func (r *UserRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	return r.GetByID(ctx, id)
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	poList, err := r.repo.Where(ctx, repository.Eq("email", email))
	if err != nil {
		return nil, err
	}
	if len(poList) == 0 {
		return nil, nil
	}
	return r.converter.ToEntity(poList[0]), nil
}

// FindByUsername 根据用户名查找用户
func (r *UserRepositoryImpl) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	poList, err := r.repo.Where(ctx, repository.Eq("username", username))
	if err != nil {
		return nil, err
	}
	if len(poList) == 0 {
		return nil, nil
	}
	return r.converter.ToEntity(poList[0]), nil
}

// Update 更新用户（实现 BaseRepository）
func (r *UserRepositoryImpl) Update(ctx context.Context, user *entity.User) error {
	po := r.converter.ToPO(user)
	return r.repo.Update(ctx, po)
}

// Delete 删除用户（实现 BaseRepository）
func (r *UserRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.repo.Delete(ctx, id)
}

// List 查询全部用户列表（实现 BaseRepository）
func (r *UserRepositoryImpl) List(ctx context.Context) ([]*entity.User, error) {
	pos, err := r.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]*entity.User, len(pos))
	for i, po := range pos {
		users[i] = r.converter.ToEntity(po)
	}
	return users, nil
}

// Page 分页查询（实现 BaseRepository）
func (r *UserRepositoryImpl) Page(ctx context.Context, request *repository.PageRequest) (*repository.PageResult[*entity.User], error) {
	poResult, err := r.repo.Page(ctx, request)
	if err != nil {
		return nil, err
	}
	// 转换为领域实体
	items := make([]*entity.User, len(poResult.Items))
	for i, po := range poResult.Items {
		items[i] = r.converter.ToEntity(po)
	}
	return repository.NewPageResult(items, poResult.Total, poResult.Page, poResult.Size), nil
}

// Where 条件查询（实现 QueryableRepository）
func (r *UserRepositoryImpl) Where(ctx context.Context, conditions ...*repository.Condition) ([]*entity.User, error) {
	poList, err := r.repo.Where(ctx, conditions...)
	if err != nil {
		return nil, err
	}
	users := make([]*entity.User, len(poList))
	for i, po := range poList {
		users[i] = r.converter.ToEntity(po)
	}
	return users, nil
}

// Count 统计数量（实现 QueryableRepository）
func (r *UserRepositoryImpl) Count(ctx context.Context, conditions ...*repository.Condition) (int64, error) {
	return r.repo.Count(ctx, conditions...)
}

// Exists 存在性检查（实现 QueryableRepository）
func (r *UserRepositoryImpl) Exists(ctx context.Context, conditions ...*repository.Condition) (bool, error) {
	return r.repo.Exists(ctx, conditions...)
}

// Query 获取查询构建器（实现 QueryableRepository）
func (r *UserRepositoryImpl) Query() repository.QueryBuilder[entity.User] {
	return NewUserQueryBuilder(r.repo.Query(), r.converter)
}

// ExistsByEmail 检查邮箱是否存在
func (r *UserRepositoryImpl) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	return r.repo.Exists(ctx, repository.Eq("email", email))
}

// ExistsByUsername 检查用户名是否存在
func (r *UserRepositoryImpl) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	return r.repo.Exists(ctx, repository.Eq("username", username))
}

// UserQueryBuilder 用户查询构建器（包装 PO 构建器，自动转换）
type UserQueryBuilder struct {
	poBuilder repository.QueryBuilder[infraEntity.UserPO]
	converter *converter.UserConverter
}

// NewUserQueryBuilder 创建用户查询构建器
func NewUserQueryBuilder(poBuilder repository.QueryBuilder[infraEntity.UserPO], converter *converter.UserConverter) *UserQueryBuilder {
	return &UserQueryBuilder{
		poBuilder: poBuilder,
		converter: converter,
	}
}

func (b *UserQueryBuilder) Where(condition *repository.Condition) repository.QueryBuilder[entity.User] {
	b.poBuilder.Where(condition)
	return b
}

func (b *UserQueryBuilder) And(conditions ...*repository.Condition) repository.QueryBuilder[entity.User] {
	b.poBuilder.And(conditions...)
	return b
}

func (b *UserQueryBuilder) OrderBy(field string) repository.QueryBuilder[entity.User] {
	b.poBuilder.OrderBy(field)
	return b
}

func (b *UserQueryBuilder) OrderByDesc(field string) repository.QueryBuilder[entity.User] {
	b.poBuilder.OrderByDesc(field)
	return b
}

func (b *UserQueryBuilder) Limit(limit int) repository.QueryBuilder[entity.User] {
	b.poBuilder.Limit(limit)
	return b
}

func (b *UserQueryBuilder) Offset(offset int) repository.QueryBuilder[entity.User] {
	b.poBuilder.Offset(offset)
	return b
}

func (b *UserQueryBuilder) Select(fields ...string) repository.QueryBuilder[entity.User] {
	b.poBuilder.Select(fields...)
	return b
}

func (b *UserQueryBuilder) Find(ctx context.Context) ([]*entity.User, error) {
	pos, err := b.poBuilder.Find(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]*entity.User, len(pos))
	for i, po := range pos {
		users[i] = b.converter.ToEntity(po)
	}
	return users, nil
}

func (b *UserQueryBuilder) First(ctx context.Context) (*entity.User, error) {
	po, err := b.poBuilder.First(ctx)
	if err != nil || po == nil {
		return nil, err
	}
	return b.converter.ToEntity(po), nil
}

func (b *UserQueryBuilder) Count(ctx context.Context) (int64, error) {
	return b.poBuilder.Count(ctx)
}

func (b *UserQueryBuilder) Exists(ctx context.Context) (bool, error) {
	return b.poBuilder.Exists(ctx)
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
