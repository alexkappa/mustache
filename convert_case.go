package mustache

import (
	"regexp"
	"strings"
)

var (
	matchSnakeCase = regexp.MustCompile("(^[A-Za-z])|_([A-Za-z])")
)

func toCamelCase(camelCase string) string {
	return matchSnakeCase.ReplaceAllStringFunc(camelCase, func(s string) string {
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})
}
