package archigen

// Option 可选配置函数
type Option func(*Generator)

// WithoutDocker 不生成 Docker 文件
func WithoutDocker() Option {
	return func(g *Generator) {
		g.config.GenerateDocker = false
	}
}

// WithoutUserModule 不生成示例模块
func WithoutUserModule() Option {
	return func(g *Generator) {
		g.config.GenerateUserModule = false
	}
}

// WithOutputPath 设置输出路径
func WithOutputPath(path string) Option {
	return func(g *Generator) {
		g.config.OutputPath = path
	}
}

// WithUseRedis 启用 Redis
func WithUseRedis() Option {
	return func(g *Generator) {
		g.config.UseRedis = true
	}
}
