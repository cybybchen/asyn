package blockword

import "regexp"

const (
	userNameStr = "^[\u4e00-\u9fa5_a-zA-Z0-9]{1,8}$" // 1-8个 中文|字母|数字|_
)

var (
	userNameReg = regexp.MustCompile(userNameStr)
)

func IsValidUserName(name string) bool {
	return isMatch(userNameReg, name)
}

// 判断val是否能正确匹配exp中的正则表达式。
// val可以是[]byte, []rune, string类型。
func isMatch(exp *regexp.Regexp, val interface{}) bool {
	switch v := val.(type) {
	case []rune:
		return exp.MatchString(string(v))
	case []byte:
		return exp.Match(v)
	case string:
		return exp.MatchString(v)
	default:
		return false
	}
}
