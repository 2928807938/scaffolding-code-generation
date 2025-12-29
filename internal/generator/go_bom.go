package generator

// generateBOM 生成 BOM 模块
func (g *GoGenerator) generateBOM() error {
	// go.mod
	goModTmpl := `module {{.ModulePath}}/bom

go 1.24.11

require (
	// Hertz HTTP 框架
	github.com/cloudwego/hertz v0.9.3

	// Kitex RPC 框架
	github.com/cloudwego/kitex v0.11.3

	// 通用工具
	github.com/google/uuid v1.6.0
	github.com/bytedance/sonic v1.12.6

	// 数据库
	gorm.io/driver/postgres v1.5.11
	gorm.io/gorm v1.25.12

	// 配置管理
	github.com/spf13/viper v1.19.0

	// 日志
	github.com/cloudwego/hertz/pkg/common/hlog v0.9.3

	// 验证器
	github.com/go-playground/validator/v10 v10.23.0
{{if .UseRedis}}
	// 缓存
	github.com/redis/go-redis/v9 v9.7.0
{{end}})
`
	if err := g.renderAndWrite(goModTmpl, "bom/go.mod"); err != nil {
		return err
	}

	// bom.go
	bomGoTmpl := `// Package bom 是 Bill of Materials 模块，用于统一管理所有依赖版本
// 其他模块通过 replace 指令引用此模块，自动继承依赖版本
package bom

import (
	// Hertz HTTP 框架
	_ "github.com/cloudwego/hertz/pkg/app"
	_ "github.com/cloudwego/hertz/pkg/app/server"
	_ "github.com/cloudwego/hertz/pkg/protocol/consts"
	_ "github.com/cloudwego/hertz/pkg/common/hlog"

	// Kitex RPC 框架
	_ "github.com/cloudwego/kitex/client"
	_ "github.com/cloudwego/kitex/server"
	_ "github.com/cloudwego/kitex/pkg/klog"

	// 通用工具
	_ "github.com/google/uuid"
	_ "github.com/bytedance/sonic"

	// 数据库
	_ "gorm.io/driver/postgres"
	_ "gorm.io/gorm"

	// 配置管理
	_ "github.com/spf13/viper"

	// 验证器
	_ "github.com/go-playground/validator/v10"
{{if .UseRedis}}
	// 缓存
	_ "github.com/redis/go-redis/v9"
{{end}})
`
	return g.renderAndWrite(bomGoTmpl, "bom/bom.go")
}
