package archigen

import "errors"

var (
	// ErrProjectNameEmpty 项目名称为空
	ErrProjectNameEmpty = errors.New("项目名称不能为空")

	// ErrModulePathEmpty 模块路径为空
	ErrModulePathEmpty = errors.New("模块路径不能为空")

	// ErrOutputPathEmpty 输出路径为空
	ErrOutputPathEmpty = errors.New("输出路径不能为空")

	// ErrDirAlreadyExists 目录已存在
	ErrDirAlreadyExists = errors.New("目录已存在")

	// ErrUnsupportedLanguage 不支持的语言
	ErrUnsupportedLanguage = errors.New("不支持的编程语言")
)
