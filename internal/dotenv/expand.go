// Package dotenv provides parsing of .env files into key-value maps.
package dotenv

import (
	"fmt"
	"strings"
)

// Expand performs variable interpolation on all values in the given map.
// Supported syntax:
//
//	${VAR}          - substitute value of VAR
//	${VAR:-default} - substitute value of VAR, or "default" if VAR is unset/empty
//	${VAR:?error}   - substitute value of VAR, or return error if VAR is unset/empty
//	\${VAR}         - literal ${VAR}, no expansion
//
// Circular references are detected and return an error.
func Expand(vars map[string]string) error {
	for key, value := range vars {
		expanded, err := expandValue(value, vars, make(map[string]bool))
		if err != nil {
			return fmt.Errorf("variable %q: %w", key, err)
		}
		vars[key] = expanded
	}

	return nil
}

func expandValue(value string, vars map[string]string, stack map[string]bool) (string, error) {
	var result strings.Builder
	i := 0

	for i < len(value) {
		// Handle escaped \${VAR} or \$VAR
		if i < len(value)-1 && value[i] == '\\' && (value[i+1] == '$' || value[i+1] == '{') {
			if value[i+1] == '$' {
				result.WriteByte('$')
				i += 2
				continue
			}
			// \{ is just a literal backslash-brace, not special
			result.WriteByte(value[i])
			i++
			continue
		}

		// Look for ${...}
		if i < len(value)-1 && value[i] == '$' && i+1 < len(value) && value[i+1] == '{' {
			end := findClosingBrace(value, i+2)
			if end == -1 {
				result.WriteByte(value[i])
				i++
				continue
			}

			expr := value[i+2 : end]
			expanded, err := expandExpr(expr, vars, stack)
			if err != nil {
				return "", err
			}
			result.WriteString(expanded)
			i = end + 1
			continue
		}

		result.WriteByte(value[i])
		i++
	}

	return result.String(), nil
}

func findClosingBrace(s string, start int) int {
	depth := 1
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func expandExpr(expr string, vars map[string]string, stack map[string]bool) (string, error) {
	var name, modifier, arg string

	// Check for :- modifier
	if idx := strings.Index(expr, ":-"); idx != -1 {
		name = expr[:idx]
		modifier = ":-"
		arg = expr[idx+2:]
	} else if idx := strings.Index(expr, ":?"); idx != -1 {
		name = expr[:idx]
		modifier = ":?"
		arg = expr[idx+2:]
	} else {
		name = expr
	}

	name = strings.TrimSpace(name)

	// Circular reference detection
	if stack[name] {
		return "", fmt.Errorf("circular reference detected: %s", name)
	}

	val, exists := vars[name]

	switch modifier {
	case ":-":
		if !exists || val == "" {
			// Expand the default value too (it may contain references)
			return expandValue(arg, vars, stack)
		}
	case ":?":
		if !exists || val == "" {
			if arg == "" {
				arg = fmt.Sprintf("variable %q is required", name)
			}
			return "", fmt.Errorf("%s", arg)
		}
	}

	if !exists {
		return "", nil
	}

	// Expand the value itself (it may contain nested references)
	if strings.Contains(val, "${") {
		newStack := make(map[string]bool)
		for k, v := range stack {
			newStack[k] = v
		}
		newStack[name] = true
		return expandValue(val, vars, newStack)
	}

	return val, nil
}
