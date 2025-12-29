package generator

import "github.com/tuza/scaffolding-code-generation/internal/config"

// Generator 生成器接口
type Generator interface {
	Generate() error
}

// NewGenerator 根据语言创建对应的生成器
func NewGenerator(cfg *config.ProjectConfig) Generator {
	switch cfg.Language {
	case config.LanguageGo:
		return NewGoGenerator(cfg)
	case config.LanguageJava:
		// Java 生成器预留
		return nil
	default:
		return NewGoGenerator(cfg)
	}
}
