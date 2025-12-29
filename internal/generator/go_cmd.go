package generator

// generateCmd 生成 cmd/api 入口模块
func (g *GoGenerator) generateCmd() error {
	// go.mod
	goModTmpl := `module {{.ModulePath}}/cmd/api

go 1.24.11

require (
	{{.ModulePath}}/bom v0.0.0
	{{.ModulePath}}/share v0.0.0
	{{.ModulePath}}/user/domain v0.0.0
	{{.ModulePath}}/user/infrastructure v0.0.0
	{{.ModulePath}}/api/user-api v0.0.0

	// Hertz HTTP 框架
	github.com/cloudwego/hertz v0.9.3

	// 数据库
	gorm.io/driver/postgres v1.5.11
	gorm.io/gorm v1.25.12

	// 配置管理
	github.com/spf13/viper v1.19.0
)

replace (
	{{.ModulePath}}/bom => ../../bom
	{{.ModulePath}}/share => ../../share
	{{.ModulePath}}/user/domain => ../../user/domain
	{{.ModulePath}}/user/infrastructure => ../../user/infrastructure
	{{.ModulePath}}/api/user-api => ../../api/user-api
)
`
	if err := g.renderAndWrite(goModTmpl, "cmd/api/go.mod"); err != nil {
		return err
	}

	// main.go
	mainGoTmpl := `package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	userHTTP "{{.ModulePath}}/api/user-api/http"
	"{{.ModulePath}}/share/middleware"
	infraEntity "{{.ModulePath}}/user/infrastructure/entity"
	infraRepo "{{.ModulePath}}/user/infrastructure/repository"
)

func main() {
	// 初始化数据库
	db, err := initDB()
	if err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(&infraEntity.UserPO{}); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化仓储
	userRepo := infraRepo.NewUserRepositoryImpl(db)

	// 初始化 Hertz
	port := getEnv("PORT", "8080")
	h := server.Default(server.WithHostPorts(":" + port))

	// 全局中间件
	h.Use(middleware.Trace())

	// 健康检查
	h.GET("/health", func(ctx context.Context, c *app.RequestContext) {
		c.JSON(consts.StatusOK, map[string]interface{}{
			"status": "ok",
		})
	})

	// 用户 API
	userHandler := userHTTP.NewUserHandler(userRepo)
	v1 := h.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("", userHandler.CreateUser)
			users.GET("", userHandler.ListUsers)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
	}

	// 启动服务
	log.Printf("服务启动在 :%s", port)
	h.Spin()
}

func initDB() (*gorm.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "{{.ProjectName}}")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
		host, user, password, dbname, port)

	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
`
	if err := g.renderAndWrite(mainGoTmpl, "cmd/api/main.go"); err != nil {
		return err
	}

	return nil
}
