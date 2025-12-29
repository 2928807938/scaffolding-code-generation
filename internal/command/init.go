package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tuza/scaffolding-code-generation/internal/config"
	"github.com/tuza/scaffolding-code-generation/internal/generator"
	"github.com/tuza/scaffolding-code-generation/internal/prompt"
)

// NewInitCommand 创建 init 命令
func NewInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "初始化一个新的 DDD 项目",
		Long: `初始化一个基于领域驱动设计（DDD）的 Go 项目。

该命令会引导你完成项目配置，并生成完整的项目骨架，包括：
  - BOM 依赖管理模块
  - 公共组件模块 (share)
  - 用户模块示例 (user/domain + user/infrastructure)
  - API 模块 (api/user-api)
  - 主程序入口 (cmd/api)
  - Docker 配置文件`,
		Example: `  archi-gen init`,
		RunE:    runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println()
	fmt.Println("🚀 欢迎使用 Archi-Gen 项目脚手架!")
	fmt.Println()
	fmt.Println("   该工具将帮助你创建一个基于 DDD 的 Go 项目")
	fmt.Println("   技术栈: Go + Hertz + Kitex + GORM + PostgreSQL + Docker")
	fmt.Println()

	// 创建交互式问答
	interactive := prompt.NewInteractive()

	// 获取用户配置
	cfg, err := interactive.AskProjectConfig()
	if err != nil {
		return fmt.Errorf("获取配置失败: %w", err)
	}

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 检查目录是否已存在
	projectFullPath := filepath.Join(cfg.OutputPath, cfg.ProjectName)
	if _, err := os.Stat(projectFullPath); !os.IsNotExist(err) {
		return fmt.Errorf("目录 '%s' 已存在", projectFullPath)
	}

	// 打印配置摘要
	printConfigSummary(cfg)

	fmt.Println()
	fmt.Println("✨ 正在生成项目骨架...")
	fmt.Println()

	// 创建生成器并生成项目
	gen := generator.NewGenerator(cfg)
	if gen == nil {
		return fmt.Errorf("不支持的语言: %s", cfg.Language)
	}

	if err := gen.Generate(); err != nil {
		return fmt.Errorf("生成项目失败: %w", err)
	}

	// 打印完成信息
	printSuccessMessage(cfg)

	return nil
}

// printConfigSummary 打印配置摘要
func printConfigSummary(cfg *config.ProjectConfig) {
	fmt.Println()
	fmt.Println("📋 项目配置:")
	fmt.Printf("   项目名称: %s\n", cfg.ProjectName)
	fmt.Printf("   模块路径: %s\n", cfg.ModulePath)
	fmt.Printf("   生成路径: %s\n", filepath.Join(cfg.OutputPath, cfg.ProjectName))
	fmt.Printf("   开发语言: %s\n", cfg.Language)
	fmt.Printf("   数据库:   PostgreSQL\n")
	fmt.Printf("   缓存:     %s\n", boolToYesNo(cfg.UseRedis))
	fmt.Printf("   部署方式: Docker\n")
}

// printSuccessMessage 打印成功信息
func printSuccessMessage(cfg *config.ProjectConfig) {
	projectFullPath := filepath.Join(cfg.OutputPath, cfg.ProjectName)
	fmt.Println()
	fmt.Println("🎉 项目骨架生成成功!")
	fmt.Println()
	fmt.Printf("📦 项目路径: %s\n", projectFullPath)
	fmt.Println()
	fmt.Println("🚀 快速开始:")
	fmt.Printf("   cd %s\n", projectFullPath)
	fmt.Println("   go work sync")
	if cfg.UseRedis {
		fmt.Println("   docker-compose up -d postgres redis")
	} else {
		fmt.Println("   docker-compose up -d postgres")
	}
	fmt.Println("   go run ./cmd/api/main.go")
	fmt.Println()
	fmt.Println("📖 访问 http://localhost:8080/health 检查服务状态")
	fmt.Println()
}

func boolToYesNo(b bool) string {
	if b {
		return "Redis (是)"
	}
	return "无"
}
