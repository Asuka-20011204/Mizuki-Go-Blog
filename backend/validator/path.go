package validator

import "strings"

// SafePath 检查路径组件是否包含危险字符，用于防止目录穿越攻击
// component 必须是单一文件名或目录名，不能包含路径分隔符或 ../
func SafePath(component string) bool {
	if component == "" {
		return false
	}
	// 阻止 . 和 .. 以及以点开头的隐藏文件/目录
	if component == "." || component == ".." {
		return false
	}
	// 阻止路径分隔符
	if strings.Contains(component, "/") || strings.Contains(component, "\\") {
		return false
	}
	// 阻止空字节注入
	if strings.ContainsRune(component, 0) {
		return false
	}
	return true
}

// SafeSlug 专门校验文章 slug：只允许字母、数字、短横线、下划线
func SafeSlug(slug string) bool {
	if !SafePath(slug) {
		return false
	}
	for _, r := range slug {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}
