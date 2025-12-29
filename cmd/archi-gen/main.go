package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tuza/scaffolding-code-generation/internal/command"
)

var version = "1.0.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "archi-gen",
		Short: "DDD 项目脚手架生成器",
		Long: `Archi-Gen 是一个基于领域驱动设计（DDD）的 Go 项目脚手架生成器。

它可以帮助你快速创建一个结构清晰、分层明确的 Go 项目，包括：
  - 领域层 (Domain Layer)
  - 应用层 (Application Layer)
  - 基础设施层 (Infrastructure Layer)
  - 接口层 (Interface Layer)

技术栈：
  - Go 1.24+
  - Gin (HTTP 框架)
  - GORM (ORM)
  - PostgreSQL (数据库)
  - Redis (缓存，可选)
  - Docker (容器化部署)`,
		Version: version,
	}

	// 添加子命令
	rootCmd.AddCommand(command.NewInitCommand())

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
