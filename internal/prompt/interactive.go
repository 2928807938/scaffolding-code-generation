package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/tuza/scaffolding-code-generation/internal/config"
)

// Interactive 交互式问答
type Interactive struct{}

// NewInteractive 创建交互式问答实例
func NewInteractive() *Interactive {
	return &Interactive{}
}

// AskProjectConfig 询问项目配置
func (i *Interactive) AskProjectConfig() (*config.ProjectConfig, error) {
	cfg := config.NewProjectConfig()

	// 1. 询问项目名称
	projectName, err := i.askProjectName()
	if err != nil {
		return nil, err
	}
	cfg.ProjectName = projectName

	// 2. 询问开发语言
	language, err := i.askLanguage()
	if err != nil {
		return nil, err
	}
	cfg.Language = language

	// 3. 询问是否托管到远程仓库
	isHosted, err := i.askIsHosted()
	if err != nil {
		return nil, err
	}

	// 4. 根据是否托管决定模块路径
	var modulePath string
	if isHosted {
		// 托管到远程仓库，需要输入完整模块路径
		modulePath, err = i.askModulePath(projectName)
		if err != nil {
			return nil, err
		}
	} else {
		// 不托管，使用项目名称作为模块路径
		modulePath = projectName
	}
	cfg.ModulePath = modulePath

	// 5. 询问是否使用 Redis
	useRedis, err := i.askUseRedis()
	if err != nil {
		return nil, err
	}
	cfg.UseRedis = useRedis

	// 6. 询问是否自定义输出路径
	customOutputPath, err := i.askCustomOutputPath()
	if err != nil {
		return nil, err
	}

	// 7. 根据是否自定义决定输出路径
	var outputPath string
	if customOutputPath {
		// 用户自定义路径
		outputPath, err = i.askOutputPath()
		if err != nil {
			return nil, err
		}
	} else {
		// 使用默认路径（当前目录）
		outputPath = i.getDefaultOutputPath()
	}
	cfg.OutputPath = outputPath

	return cfg, nil
}

// askProjectName 询问项目名称
func (i *Interactive) askProjectName() (string, error) {
	var projectName string
	prompt := &survey.Input{
		Message: "请输入项目名称:",
		Help:    "项目名称将作为目录名和 Go 模块名的一部分",
	}

	validator := survey.ComposeValidators(
		survey.Required,
		func(val interface{}) error {
			str := val.(string)
			// 验证项目名称格式（只允许字母、数字、下划线、中划线）
			matched, _ := regexp.MatchString(`^[a-zA-Z][a-zA-Z0-9_-]*$`, str)
			if !matched {
				return fmt.Errorf("项目名称必须以字母开头，只能包含字母、数字、下划线和中划线")
			}
			return nil
		},
	)

	err := survey.AskOne(prompt, &projectName, survey.WithValidator(validator))
	if err != nil {
		return "", err
	}

	return projectName, nil
}

// askLanguage 询问开发语言
func (i *Interactive) askLanguage() (config.Language, error) {
	var language string
	prompt := &survey.Select{
		Message: "请选择开发语言:",
		Options: []string{"Go", "Java (即将支持)"},
		Default: "Go",
	}

	err := survey.AskOne(prompt, &language)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(language, "Java") {
		fmt.Println("⚠️  Java 语言支持即将推出，当前默认使用 Go")
		return config.LanguageGo, nil
	}

	return config.LanguageGo, nil
}

// askIsHosted 询问是否托管到远程仓库
func (i *Interactive) askIsHosted() (bool, error) {
	var isHosted bool
	prompt := &survey.Confirm{
		Message: "是否托管到远程仓库（如 GitHub、GitLab 等）?",
		Default: true,
		Help:    "如果选择是，需要输入完整的模块路径；选择否则使用项目名称作为模块路径",
	}

	err := survey.AskOne(prompt, &isHosted)
	if err != nil {
		return false, err
	}

	return isHosted, nil
}

// askModulePath 询问模块路径
func (i *Interactive) askModulePath(projectName string) (string, error) {
	var modulePath string
	defaultPath := fmt.Sprintf("github.com/yourname/%s", projectName)

	prompt := &survey.Input{
		Message: "请输入 Go 模块路径:",
		Default: defaultPath,
		Help:    "Go 模块路径，例如: github.com/username/project",
	}

	validator := survey.ComposeValidators(
		survey.Required,
		func(val interface{}) error {
			str := val.(string)
			// 简单验证模块路径格式
			if !strings.Contains(str, "/") {
				return fmt.Errorf("模块路径格式不正确，应类似: github.com/username/project")
			}
			return nil
		},
	)

	err := survey.AskOne(prompt, &modulePath, survey.WithValidator(validator))
	if err != nil {
		return "", err
	}

	return modulePath, nil
}

// askUseRedis 询问是否使用 Redis
func (i *Interactive) askUseRedis() (bool, error) {
	var useRedis bool
	prompt := &survey.Confirm{
		Message: "是否使用 Redis?",
		Default: true,
	}

	err := survey.AskOne(prompt, &useRedis)
	if err != nil {
		return false, err
	}

	return useRedis, nil
}

// askCustomOutputPath 询问是否自定义输出路径
func (i *Interactive) askCustomOutputPath() (bool, error) {
	var customPath bool
	prompt := &survey.Confirm{
		Message: "是否自定义项目生成路径?",
		Default: false,
		Help:    "如果选择否，项目将生成在当前目录下；选择是则需要指定生成路径",
	}

	err := survey.AskOne(prompt, &customPath)
	if err != nil {
		return false, err
	}

	return customPath, nil
}

// askOutputPath 询问输出路径
func (i *Interactive) askOutputPath() (string, error) {
	var outputPath string

	prompt := &survey.Input{
		Message: "请输入项目生成路径:",
		Help:    "项目将生成在此路径下，例如: D:\\Projects 或 /home/user/projects",
	}

	validator := survey.ComposeValidators(
		survey.Required,
		func(val interface{}) error {
			str := val.(string)
			// 检查路径是否为绝对路径
			if !filepath.IsAbs(str) {
				return fmt.Errorf("请输入绝对路径，例如: D:\\Projects 或 /home/user/projects")
			}
			return nil
		},
	)

	err := survey.AskOne(prompt, &outputPath, survey.WithValidator(validator))
	if err != nil {
		return "", err
	}

	return outputPath, nil
}

// getDefaultOutputPath 获取默认输出路径（当前工作目录）
func (i *Interactive) getDefaultOutputPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}
