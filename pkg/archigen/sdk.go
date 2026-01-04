package archigen

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tuza/scaffolding-code-generation/internal/config"
	"github.com/tuza/scaffolding-code-generation/internal/generator"
)

// Generator 项目生成器
type Generator struct {
	config *Config
}

// New 创建生成器
func New(cfg *Config, opts ...Option) (*Generator, error) {
	// 验证必填参数
	if cfg.ProjectName == "" {
		return nil, ErrProjectNameEmpty
	}
	if cfg.ModulePath == "" {
		return nil, ErrModulePathEmpty
	}
	if cfg.OutputPath == "" {
		return nil, ErrOutputPathEmpty
	}

	// 设置默认值
	if cfg.Language == "" {
		cfg.Language = "go"
	}

	// 检查目录是否已存在
	projectPath := filepath.Join(cfg.OutputPath, cfg.ProjectName)
	if _, err := os.Stat(projectPath); err == nil {
		return nil, ErrDirAlreadyExists
	}

	g := &Generator{
		config: cfg,
	}

	// 应用可选配置
	for _, opt := range opts {
		opt(g)
	}

	return g, nil
}

// Generate 执行生成
func (g *Generator) Generate(ctx context.Context) (*Result, error) {
	// 转换为内部配置
	internalCfg := &config.ProjectConfig{
		ProjectName: g.config.ProjectName,
		Language:    config.Language(g.config.Language),
		ModulePath:  g.config.ModulePath,
		OutputPath:  g.config.OutputPath,
		UseRedis:    g.config.UseRedis,
		Database:    "postgres",
		Deployment:  "docker",
	}

	// 创建内部生成器
	gen := generator.NewGenerator(internalCfg)
	if gen == nil {
		return nil, ErrUnsupportedLanguage
	}

	// 执行生成
	if err := gen.Generate(); err != nil {
		return nil, fmt.Errorf("生成失败: %w", err)
	}

	// 构建结果
	projectPath := filepath.Join(g.config.OutputPath, g.config.ProjectName)
	modules := []string{
		"bom",
		"share",
		"user",
		"api",
		"cmd",
	}

	return NewResult(projectPath, modules), nil
}

// GetProjectPath 获取生成的项目路径
func (g *Generator) GetProjectPath() string {
	return filepath.Join(g.config.OutputPath, g.config.ProjectName)
}
