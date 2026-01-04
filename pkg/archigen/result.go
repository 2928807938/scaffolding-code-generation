package archigen

// Result 生成结果
type Result struct {
	// ProjectPath 项目路径
	ProjectPath string

	// Modules 生成的模块列表
	Modules []string
}

// NewResult 创建生成结果
func NewResult(projectPath string, modules []string) *Result {
	return &Result{
		ProjectPath: projectPath,
		Modules:     modules,
	}
}
