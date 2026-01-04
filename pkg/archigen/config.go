package archigen

// Config 项目生成配置
type Config struct {
	// Language 编程语言: "go", "java" 等，默认 "go"
	Language string

	// ProjectName 项目名称（必填）
	ProjectName string

	// ModulePath 模块路径（必填）
	// Go: github.com/org/project
	// Java: com.org.project
	ModulePath string

	// OutputPath 输出路径（必填）
	OutputPath string

	// UseRedis 是否使用 Redis
	UseRedis bool

	// GenerateDocker 是否生成 Docker 文件（默认 true）
	GenerateDocker bool

	// GenerateUserModule 是否生成示例模块（默认 true）
	GenerateUserModule bool
}

// NewConfig 创建默认配置
func NewConfig() *Config {
	return &Config{
		Language:           "go",
		UseRedis:           false,
		GenerateDocker:     true,
		GenerateUserModule: true,
	}
}
