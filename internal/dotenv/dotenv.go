// Package dotenv provides parsing of .env files into key-value maps.
package dotenv

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Parse reads a .env file and returns a map of variable names to their raw string values.
// It handles comments, quoted values, and empty lines.
func Parse(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open env file %s: %w", path, err)
	}
	defer file.Close()

	vars := make(map[string]string)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024), 1024*1024) // 1MB max line length
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		if key != "" {
			vars[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read env file %s: %w", path, err)
	}

	return vars, nil
}

// parseLine parses a single line of a .env file.
// Returns the key, value, and any error.
func parseLine(line string) (string, string, error) {
	// Find the first '=' character
	eqIdx := strings.Index(line, "=")
	if eqIdx == -1 {
		// Lines without '=' are treated as comments/empty
		return "", "", nil
	}

	key := strings.TrimSpace(line[:eqIdx])
	value := line[eqIdx+1:]

	if key == "" {
		return "", "", fmt.Errorf("empty variable name")
	}

	value = strings.TrimSpace(value)

	// Handle quoted values
	value = unquote(value)

	return key, value, nil
}

// unquote removes surrounding quotes from a value and handles escape sequences.
func unquote(value string) string {
	if len(value) < 2 {
		return value
	}

	// Double-quoted strings
	if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		value = value[1 : len(value)-1]
		value = unescapeDoubleQuotes(value)
		return value
	}

	// Single-quoted strings
	if strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`) {
		return value[1 : len(value)-1]
	}

	return value
}

// unescapeDoubleQuotes processes escape sequences inside double-quoted strings.
func unescapeDoubleQuotes(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				result.WriteByte('\n')
				i++
			case 't':
				result.WriteByte('\t')
				i++
			case 'r':
				result.WriteByte('\r')
				i++
			case '\\':
				result.WriteByte('\\')
				i++
			case '"':
				result.WriteByte('"')
				i++
			default:
				result.WriteByte(s[i])
			}
		} else {
			result.WriteByte(s[i])
		}
	}
	return result.String()
}
