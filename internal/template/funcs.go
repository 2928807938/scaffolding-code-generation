package template

import (
	"strings"
	"unicode"
)

// ToPascalCase 转换为 PascalCase (大驼峰)
// 例如: user_name -> UserName, user-api -> UserApi
func ToPascalCase(s string) string {
	words := splitWords(s)
	for i, word := range words {
		words[i] = capitalize(word)
	}
	return strings.Join(words, "")
}

// ToCamelCase 转换为 camelCase (小驼峰)
// 例如: user_name -> userName, user-api -> userApi
func ToCamelCase(s string) string {
	words := splitWords(s)
	for i, word := range words {
		if i == 0 {
			words[i] = strings.ToLower(word)
		} else {
			words[i] = capitalize(word)
		}
	}
	return strings.Join(words, "")
}

// ToSnakeCase 转换为 snake_case
// 例如: UserName -> user_name, userApi -> user_api
func ToSnakeCase(s string) string {
	words := splitCamelCase(s)
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	return strings.Join(words, "_")
}

// ToKebabCase 转换为 kebab-case
// 例如: UserName -> user-name, userApi -> user-api
func ToKebabCase(s string) string {
	words := splitCamelCase(s)
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	return strings.Join(words, "-")
}

// splitWords 按照下划线和中划线分割字符串
func splitWords(s string) []string {
	// 先替换中划线为下划线
	s = strings.ReplaceAll(s, "-", "_")
	// 按下划线分割
	return strings.Split(s, "_")
}

// splitCamelCase 按照驼峰分割字符串
func splitCamelCase(s string) []string {
	var words []string
	var currentWord strings.Builder

	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			if currentWord.Len() > 0 {
				words = append(words, currentWord.String())
				currentWord.Reset()
			}
		}
		currentWord.WriteRune(r)
	}

	if currentWord.Len() > 0 {
		words = append(words, currentWord.String())
	}

	// 如果输入包含下划线或中划线，也要处理
	var result []string
	for _, word := range words {
		parts := splitWords(word)
		result = append(result, parts...)
	}

	return result
}

// capitalize 首字母大写
func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(string(s[0])) + strings.ToLower(s[1:])
}
