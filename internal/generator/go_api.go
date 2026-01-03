package generator

// generateUserAPI 生成 api/user-api 模块
func (g *GoGenerator) generateUserAPI() error {
	// go.mod
	goModTmpl := `module {{.ModulePath}}/api/user-api

go 1.24.11

require (

	// Hertz HTTP 框架
	github.com/cloudwego/hertz v0.9.3

	// 通用工具
	github.com/google/uuid v1.6.0
	{{.ModulePath}}/share v0.0.0
	{{.ModulePath}}/user/domain v0.0.0
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

	// dto/vo/user_vo.go
	userVoTmpl := `package vo

import (
	"time"

	"github.com/google/uuid"
)

// UserVo 用户响应视图对象
type UserVo struct {
	ID        uuid.UUID ` + "`json:\"id\"`" + `
	Username  string    ` + "`json:\"username\"`" + `
	Email     string    ` + "`json:\"email\"`" + `
	Status    int       ` + "`json:\"status\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
	UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}
`
	if err := g.writeFile("api/user-api/dto/vo/user_vo.go", userVoTmpl); err != nil {
		return err
	}

	// dto/request/user_request.go
	userRequestTmpl := `package request

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
	if err := g.writeFile("api/user-api/dto/request/user_request.go", userRequestTmpl); err != nil {
		return err
	}

	// converter/user_converter.go
	userConverterTmpl := `package converter

import (
	"{{.ModulePath}}/api/user-api/dto/vo"
	"{{.ModulePath}}/user/domain/entity"
)

// UserConverter 用户转换器
type UserConverter struct{}

// NewUserConverter 创建用户转换器
func NewUserConverter() *UserConverter {
	return &UserConverter{}
}

// ToVo 将领域实体转换为视图对象
func (c *UserConverter) ToVo(user *entity.User) *vo.UserVo {
	if user == nil {
		return nil
	}
	return &vo.UserVo{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email.String(),
		Status:    int(user.Status),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
`
	if err := g.renderAndWrite(userConverterTmpl, "api/user-api/converter/user_converter.go"); err != nil {
		return err
	}

	// service/user_app_service.go
	userAppServiceTmpl := `package service

import (
	"context"

	"github.com/google/uuid"
	"{{.ModulePath}}/api/user-api/converter"
	"{{.ModulePath}}/api/user-api/dto/request"
	"{{.ModulePath}}/api/user-api/dto/vo"
	"{{.ModulePath}}/user/domain/enum"
	"{{.ModulePath}}/user/domain/repository"
	domainService "{{.ModulePath}}/user/domain/service"
	"{{.ModulePath}}/user/domain/valueobject"
)

// UserAppService 用户应用服务
type UserAppService struct {
	userRepo          repository.UserRepository
	userDomainService *domainService.UserDomainService
	converter         *converter.UserConverter
}

// NewUserAppService 创建用户应用服务
func NewUserAppService(userRepo repository.UserRepository) *UserAppService {
	return &UserAppService{
		userRepo:          userRepo,
		userDomainService: domainService.NewUserDomainService(userRepo),
		converter:         converter.NewUserConverter(),
	}
}

// CreateUser 创建用户
func (s *UserAppService) CreateUser(ctx context.Context, req *request.CreateUserRequest) (*vo.UserVo, error) {
	// 密码加密
	password, err := valueobject.NewPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 调用领域服务创建用户
	user, err := s.userDomainService.CreateUser(ctx, req.Username, req.Email, password.Hash())
	if err != nil {
		return nil, err
	}

	// 保存用户
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	return s.converter.ToVo(user), nil
}

// GetUser 获取用户
func (s *UserAppService) GetUser(ctx context.Context, id uuid.UUID) (*vo.UserVo, error) {
	user, err := s.userDomainService.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.converter.ToVo(user), nil
}

// UpdateUser 更新用户
func (s *UserAppService) UpdateUser(ctx context.Context, id uuid.UUID, req *request.UpdateUserRequest) (*vo.UserVo, error) {
	username := ""
	if req.Username != nil {
		username = *req.Username
	}
	var status *enum.UserStatus
	if req.Status != nil {
		s := enum.UserStatus(*req.Status)
		status = &s
	}

	user, err := s.userDomainService.UpdateUser(ctx, id, username, status)
	if err != nil {
		return nil, err
	}

	// 保存更新
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return s.converter.ToVo(user), nil
}

// DeleteUser 删除用户
func (s *UserAppService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.userDomainService.DeleteUser(ctx, id)
}

// ListUsers 查询用户列表
func (s *UserAppService) ListUsers(ctx context.Context, req *request.ListUsersRequest) ([]*vo.UserVo, int64, error) {
	req.SetDefaults()

	users, total, err := s.userRepo.List(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	// 转换为响应 DTO
	responses := make([]*vo.UserVo, len(users))
	for i, user := range users {
		responses[i] = s.converter.ToVo(user)
	}
	return responses, total, nil
}
`
	if err := g.renderAndWrite(userAppServiceTmpl, "api/user-api/service/user_app_service.go"); err != nil {
		return err
	}

	// http/user_handler.go
	userHandlerTmpl := `package http

import (
	"context"
	"{{.ModulePath}}/share/errors"

	"{{.ModulePath}}/api/user-api/dto/request"
	"{{.ModulePath}}/api/user-api/service"
	"{{.ModulePath}}/share/types"
	"{{.ModulePath}}/user/domain/repository"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/google/uuid"
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
// @Param request body request.CreateUserRequest true "创建用户请求"
// @Success 200 {object} types.Response{data=vo.UserVo}
// @Router /api/v1/users [post]
func (h *UserHandler) CreateUser(ctx context.Context, c *app.RequestContext) {
	var req request.CreateUserRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(consts.StatusBadRequest, types.Error(400, err.Error()))
		return
	}

	resp, err := h.userAppService.CreateUser(ctx, &req)
	if err != nil {
		errors.HandleError(ctx, c, err)
		return
	}

	c.JSON(consts.StatusOK, types.Success(resp))
}

// GetUser 获取用户
// @Summary 获取用户详情
// @Tags 用户管理
// @Produce json
// @Param id path string true "用户ID"
// @Success 200 {object} types.Response{data=vo.UserVo}
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
		errors.HandleError(ctx, c, err)
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
// @Param request body request.UpdateUserRequest true "更新用户请求"
// @Success 200 {object} types.Response{data=vo.UserVo}
// @Router /api/v1/users/{id} [put]
func (h *UserHandler) UpdateUser(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(consts.StatusBadRequest, types.Error(400, "无效的用户ID"))
		return
	}

	var req request.UpdateUserRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(consts.StatusBadRequest, types.Error(400, err.Error()))
		return
	}

	resp, err := h.userAppService.UpdateUser(ctx, id, &req)
	if err != nil {
		errors.HandleError(ctx, c, err)
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
		errors.HandleError(ctx, c, err)
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
	var req request.ListUsersRequest
	if err := c.BindQuery(&req); err != nil {
		c.JSON(consts.StatusBadRequest, types.Error(400, err.Error()))
		return
	}

	users, total, err := h.userAppService.ListUsers(ctx, &req)
	if err != nil {
		errors.HandleError(ctx, c, err)
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
