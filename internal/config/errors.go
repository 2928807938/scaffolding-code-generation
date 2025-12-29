package config

import "errors"

var (
	ErrProjectNameEmpty = errors.New("项目名称不能为空")
	ErrModulePathEmpty  = errors.New("模块路径不能为空")
)
