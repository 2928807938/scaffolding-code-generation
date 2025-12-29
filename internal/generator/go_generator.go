package generator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tuza/scaffolding-code-generation/internal/config"
	"github.com/tuza/scaffolding-code-generation/internal/template"
)

// GoGenerator Go 项目生成器
type GoGenerator struct {
	config     *config.ProjectConfig
	tmplEngine *template.Engine
	tmplCtx    *template.Context
	outputDir  string
}

// NewGoGenerator 创建 Go 生成器
func NewGoGenerator(cfg *config.ProjectConfig) *GoGenerator {
	// 输出目录 = 输出路径 + 项目名称
	outputDir := filepath.Join(cfg.OutputPath, cfg.ProjectName)
	return &GoGenerator{
		config:     cfg,
		tmplEngine: template.NewEngine(),
		tmplCtx:    template.NewContext(cfg),
		outputDir:  outputDir,
	}
}

// Generate 生成项目
func (g *GoGenerator) Generate() error {
	steps := []struct {
		name string
		fn   func() error
	}{
		{"创建项目目录", g.createProjectDir},
		{"生成 go.work", g.generateWorkspace},
		{"生成 .gitignore", g.generateGitignore},
		{"生成 Makefile", g.generateMakefile},
		{"生成 BOM 模块", g.generateBOM},
		{"生成 share 模块", g.generateShare},
		{"生成 user/domain 模块", g.generateUserDomain},
		{"生成 user/infrastructure 模块", g.generateUserInfra},
		{"生成 user 聚合模块", g.generateUserModule},
		{"生成 api/user-api 模块", g.generateUserAPI},
		{"生成 api 聚合模块", g.generateAPIModule},
		{"生成 cmd/api 入口", g.generateCmd},
		{"生成 Dockerfile", g.generateDockerfile},
		{"生成 docker-compose.yml", g.generateDockerCompose},
		{"生成 .dockerignore", g.generateDockerignore},
		{"生成 README.md", g.generateReadme},
	}

	for _, step := range steps {
		fmt.Printf("   ✔ %s\n", step.name)
		if err := step.fn(); err != nil {
			return fmt.Errorf("%s 失败: %w", step.name, err)
		}
	}

	return nil
}

// createProjectDir 创建项目目录结构
func (g *GoGenerator) createProjectDir() error {
	dirs := []string{
		"bom",
		"share/errors",
		"share/utils",
		"share/types",
		"share/middleware",
		"user/domain/entity",
		"user/domain/repository",
		"user/domain/service",
		"user/domain/valueobject",
		"user/domain/event",
		"user/infrastructure/entity",
		"user/infrastructure/converter",
		"user/infrastructure/repository",
		"api/user-api/dto",
		"api/user-api/service",
		"api/user-api/http",
		"cmd/api",
	}

	for _, dir := range dirs {
		fullPath := filepath.Join(g.outputDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			return err
		}
	}

	return nil
}

// writeFile 写入文件
func (g *GoGenerator) writeFile(relativePath, content string) error {
	fullPath := filepath.Join(g.outputDir, relativePath)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(fullPath, []byte(content), 0644)
}

// renderAndWrite 渲染模板并写入文件
func (g *GoGenerator) renderAndWrite(tmplContent, relativePath string) error {
	rendered, err := g.tmplEngine.Render(tmplContent, g.tmplCtx)
	if err != nil {
		return err
	}
	return g.writeFile(relativePath, rendered)
}
