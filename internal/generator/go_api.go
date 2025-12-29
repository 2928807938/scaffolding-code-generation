package generator

// generateUserAPI 生成 api/user-api 模块
func (g *GoGenerator) generateUserAPI() error {
	// go.mod
	goModTmpl := `module {{.ModulePath}}/api/user-api

go 1.24.11

require (
	{{.ModulePath}}/bom v0.0.0
	{{.ModulePath}}/share v0.0.0
	{{.ModulePath}}/user/domain v0.0.0
	{{.ModulePath}}/user/infrastructure v0.0.0

	// Hertz HTTP 框架
	github.com/cloudwego/hertz v0.9.3

	// 通用工具
	github.com/google/uuid v1.6.0
	github.com/bytedance/sonic v1.12.6

	// 验证器
	github.com/go-playground/validator/v10 v10.23.0
)

replace (
	{{.ModulePath}}/bom => ../../bom
	{{.ModulePath}}/share => ../../share
	{{.ModulePath}}/user/domain => ../../user/domain
	{{.ModulePath}}/user/infrastructure => ../../user/infrastructure
)
`
	if err := g.renderAndWrite(goModTmpl, "api/user-api/go.mod"); err != nil {
		return err
	}

	// dto/user_dto.go
	userDTOTmpl := `package dto

import (
	"time"

	"github.com/google/uuid"
	"{{.ModulePath}}/user/domain/entity"
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string ` + "`json:\"username\" vd:\"len($)>2 && len($)<51\"`" + `
	Email    string ` + "`json:\"email\" vd:\"email($)\"`" + `
	Password string ` + "`json:\"password\" vd:\"len($)>5 && len($)<51\"`" + `
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username *string ` + "`json:\"username,omitempty\" vd:\"len($)>2 && len($)<51\"`" + `
	Status   *int    ` + "`json:\"status,omitempty\" vd:\"$>=0 && $<=2\"`" + `
}

// UserResponse 用户响应
type UserResponse struct {
	ID        uuid.UUID ` + "`json:\"id\"`" + `
	Username  string    ` + "`json:\"username\"`" + `
	Email     string    ` + "`json:\"email\"`" + `
	Status    int       ` + "`json:\"status\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}

// ToUserResponse 将领域实体转换为响应 DTO
func ToUserResponse(user *entity.User) *UserResponse {
	if user == nil {
		return nil
	}
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email.String(),
		Status:    int(user.Status),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// ToUserResponseList 将领域实体列表转换为响应 DTO 列表
func ToUserResponseList(users []*entity.User) []*UserResponse {
	if users == nil {
		return nil
	}
	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = ToUserResponse(user)
	}
	return responses
}

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	Page     int ` + "`query:\"page\"`" + `
	PageSize int ` + "`query:\"page_size\"`" + `
}

// SetDefaults 设置默认值
func (r *ListUsersRequest) SetDefaults() {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.PageSize <= 0 {
		r.PageSize = 10
	}
}
`
	if err := g.renderAndWrite(userDTOTmpl, "api/user-api/dto/user_dto.go"); err != nil {
		return err
	}

	// service/user_app_service.go
	userAppServiceTmpl := `package service

import (
	"context"

	"github.com/google/uuid"
	"{{.ModulePath}}/api/user-api/dto"
	"{{.ModulePath}}/share/errors"
	"{{.ModulePath}}/share/utils"
	"{{.ModulePath}}/user/domain/entity"
	"{{.ModulePath}}/user/domain/repository"
	domainService "{{.ModulePath}}/user/domain/service"
)

// UserAppService 用户应用服务
type UserAppService struct {
	userRepo          repository.UserRepository
	userDomainService *domainService.UserDomainService
}

// NewUserAppService 创建用户应用服务
func NewUserAppService(userRepo repository.UserRepository) *UserAppService {
	return &UserAppService{
		userRepo:          userRepo,
		userDomainService: domainService.NewUserDomainService(userRepo),
	}
}

// CreateUser 创建用户
func (s *UserAppService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	// 密码加密
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.ErrInternal("密码加密失败", err)
	}

	// 调用领域服务创建用户
	user, err := s.userDomainService.CreateUser(ctx, req.Username, req.Email, passwordHash)
	if err != nil {
		switch err {
		case domainService.ErrEmailAlreadyExists:
			return nil, errors.ErrConflict("邮箱已被使用")
		case domainService.ErrUsernameAlreadyExists:
			return nil, errors.ErrConflict("用户名已被使用")
		case domainService.ErrInvalidEmail:
			return nil, errors.ErrBadRequest("邮箱格式不正确")
		default:
			return nil, errors.ErrInternal("创建用户失败", err)
		}
	}

	// 保存用户
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, errors.ErrInternal("保存用户失败", err)
	}

	return dto.ToUserResponse(user), nil
}

// GetUser 获取用户
func (s *UserAppService) GetUser(ctx context.Context, id uuid.UUID) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.ErrInternal("查询用户失败", err)
	}
	if user == nil {
		return nil, errors.ErrNotFound("用户不存在")
	}
	return dto.ToUserResponse(user), nil
}

// UpdateUser 更新用户
func (s *UserAppService) UpdateUser(ctx context.Context, id uuid.UUID, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.ErrInternal("查询用户失败", err)
	}
	if user == nil {
		return nil, errors.ErrNotFound("用户不存在")
	}

	// 更新字段
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Status != nil {
		user.Status = entity.UserStatus(*req.Status)
	}

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, errors.ErrInternal("更新用户失败", err)
	}

	return dto.ToUserResponse(user), nil
}

// DeleteUser 删除用户
func (s *UserAppService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return errors.ErrInternal("查询用户失败", err)
	}
	if user == nil {
		return errors.ErrNotFound("用户不存在")
	}

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return errors.ErrInternal("删除用户失败", err)
	}

	return nil
}

// ListUsers 查询用户列表
func (s *UserAppService) ListUsers(ctx context.Context, req *dto.ListUsersRequest) ([]*dto.UserResponse, int64, error) {
	req.SetDefaults()

	users, total, err := s.userRepo.List(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, errors.ErrInternal("查询用户列表失败", err)
	}

	return dto.ToUserResponseList(users), total, nil
}
`
	if err := g.renderAndWrite(userAppServiceTmpl, "api/user-api/service/user_app_service.go"); err != nil {
		return err
	}

	// http/user_handler.go
	userHandlerTmpl := `package http

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
	"{{.ModulePath}}/api/user-api/dto"
	"{{.ModulePath}}/api/user-api/service"
	"{{.ModulePath}}/share/types"
	"{{.ModulePath}}/user/domain/repository"
)

// UserHandler 用户 HTTP 处理器
type UserHandler struct {
	userAppService *service.UserAppService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{
		userAppService: service.NewUserAppService(userRepo),
	}
}

// CreateUser 创建用户
// @Summary 创建用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param request body dto.CreateUserRequest true "创建用户请求"
// @Success 200 {object} types.Response{data=dto.UserResponse}
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(ctx context.Context, c *app.RequestContext) {
	var req dto.CreateUserRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(consts.StatusBadRequest, types.Error(400, err.Error()))
		return
	}

	resp, err := h.userAppService.CreateUser(ctx, &req)
	if err != nil {
		types.HandleError(ctx, c, err)
		return
	}

	c.JSON(consts.StatusOK, types.Success(resp))
}

// GetUser 获取用户
// @Summary 获取用户详情
// @Tags 用户管理
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} types.Response{data=dto.UserResponse}
// @Router /api/v1/users/{id} [get]
func (h *UserHandler) GetUser(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(consts.StatusBadRequest, types.Error(400, "无效的用户ID"))
		return
	}

	resp, err := h.userAppService.GetUser(ctx, id)
	if err != nil {
		types.HandleError(ctx, c, err)
		return
	}

	c.JSON(consts.StatusOK, types.Success(resp))
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path string true "用户ID"
// @Param request body dto.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} types.Response{data=dto.UserResponse}
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(consts.StatusBadRequest, types.Error(400, "无效的用户ID"))
		return
	}

	var req dto.UpdateUserRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(consts.StatusBadRequest, types.Error(400, err.Error()))
		return
	}

	resp, err := h.userAppService.UpdateUser(ctx, id, &req)
	if err != nil {
		types.HandleError(ctx, c, err)
		return
	}

	c.JSON(consts.StatusOK, types.Success(resp))
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Tags 用户管理
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} types.Response
// @Router /api/v1/users/{id} [delete]
func (h *UserHandler) DeleteUser(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(consts.StatusBadRequest, types.Error(400, "无效的用户ID"))
		return
	}

	if err := h.userAppService.DeleteUser(ctx, id); err != nil {
		types.HandleError(ctx, c, err)
		return
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("删除成功", nil))
}

// ListUsers 查询用户列表
// @Summary 查询用户列表
// @Tags 用户管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} types.Response{data=types.PageResult}
// @Router /api/v1/users [get]
func (h *UserHandler) ListUsers(ctx context.Context, c *app.RequestContext) {
	var req dto.ListUsersRequest
	if err := c.BindQuery(&req); err != nil {
		c.JSON(consts.StatusBadRequest, types.Error(400, err.Error()))
		return
	}

	users, total, err := h.userAppService.ListUsers(ctx, &req)
	if err != nil {
		types.HandleError(ctx, c, err)
		return
	}

	req.SetDefaults()
	c.JSON(consts.StatusOK, types.Success(types.PageResult{
		List:     users,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}))
}
`
	if err := g.renderAndWrite(userHandlerTmpl, "api/user-api/http/user_handler.go"); err != nil {
		return err
	}

	return nil
}

// generateAPIModule 生成 api 聚合模块
func (g *GoGenerator) generateAPIModule() error {
	// go.mod
	goModTmpl := `module {{.ModulePath}}/api

go 1.24.11

require {{.ModulePath}}/api/user-api v0.0.0

replace {{.ModulePath}}/api/user-api => ./user-api
`
	return g.renderAndWrite(goModTmpl, "api/go.mod")
}
