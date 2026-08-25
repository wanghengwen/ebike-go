package templateutil

import (
	"regexp"
	"strings"
)

// paramRegex matches ${key} patterns and captures the key name.
var paramRegex = regexp.MustCompile(`\$\{([^}]+)\}`)

// MatchParam extracts all parameter names from ${key} placeholders in the
// given template string. It matches the behavior of Java's TemplateUtils.java.
func MatchParam(template string) []string {
	matches := paramRegex.FindAllStringSubmatch(template, -1)
	if len(matches) == 0 {
		return nil
	}
	params := make([]string, 0, len(matches))
	for _, m := range matches {
		params = append(params, m[1])
	}
	return params
}

// RenderTemplate replaces all ${key} placeholders in the template with the
// corresponding values from the params map. Unmatched placeholders are left
// as-is.
func RenderTemplate(template string, params map[string]string) string {
	return paramRegex.ReplaceAllStringFunc(template, func(match string) string {
		key := strings.TrimSuffix(strings.TrimPrefix(match, "${"), "}")
		if val, ok := params[key]; ok {
			return val
		}
		return match
	})
}
