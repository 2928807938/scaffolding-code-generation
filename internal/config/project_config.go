package config

// Language 开发语言类型
type Language string

const (
	LanguageGo   Language = "go"
	LanguageJava Language = "java" // 预留
)

// ProjectConfig 项目配置
type ProjectConfig struct {
	// 用户输入的配置
	ProjectName string   // 项目名称
	Language    Language // 开发语言
	UseRedis    bool     // 是否使用 Redis
	ModulePath  string   // Go 模块路径 (例如: github.com/username/project)
	OutputPath  string   // 输出路径 (项目生成的目标目录)

	// 固定配置
	Database   string // 固定为 "postgres"
	Deployment string // 固定为 "docker"
}

// NewProjectConfig 创建项目配置，设置默认值
func NewProjectConfig() *ProjectConfig {
	return &ProjectConfig{
		Database:   "postgres",
		Deployment: "docker",
		Language:   LanguageGo,
	}
}

// Validate 验证配置
func (c *ProjectConfig) Validate() error {
	if c.ProjectName == "" {
		return ErrProjectNameEmpty
	}
	if c.ModulePath == "" {
		return ErrModulePathEmpty
	}
	return nil
}
