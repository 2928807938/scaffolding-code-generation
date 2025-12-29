package template

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/tuza/scaffolding-code-generation/internal/config"
)

// Context 模板上下文
type Context struct {
	ProjectName  string // 项目名称
	ModulePath   string // 模块路径
	UseRedis     bool   // 是否使用 Redis
	Database     string // 数据库类型 (固定为 postgres)
	DBDriver     string // 数据库驱动
	DBDSNExample string // DSN 示例
}

// NewContext 从项目配置创建模板上下文
func NewContext(cfg *config.ProjectConfig) *Context {
	return &Context{
		ProjectName:  cfg.ProjectName,
		ModulePath:   cfg.ModulePath,
		UseRedis:     cfg.UseRedis,
		Database:     cfg.Database,
		DBDriver:     "gorm.io/driver/postgres",
		DBDSNExample: "host=localhost user=postgres password=postgres dbname=" + cfg.ProjectName + " port=5432 sslmode=disable",
	}
}

// Engine 模板引擎
type Engine struct {
	funcMap template.FuncMap
}

// NewEngine 创建模板引擎
func NewEngine() *Engine {
	return &Engine{
		funcMap: template.FuncMap{
			"toUpper":      strings.ToUpper,
			"toLower":      strings.ToLower,
			"toPascalCase": ToPascalCase,
			"toCamelCase":  ToCamelCase,
			"toSnakeCase":  ToSnakeCase,
			"toKebabCase":  ToKebabCase,
		},
	}
}

// Render 渲染模板
func (e *Engine) Render(tmplContent string, ctx *Context) (string, error) {
	tmpl, err := template.New("template").Funcs(e.funcMap).Parse(tmplContent)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
}
