package validator

import (
	"fmt"
	"strconv"
	"strings"
)

// coerceString returns the trimmed string value.
func coerceString(value string) (string, error) {
	return strings.TrimSpace(value), nil
}

// coerceInteger parses a string as a base-10 integer.
func coerceInteger(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("expected integer, got empty string")
	}
	v, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("expected integer, got %q", value)
	}
	return v, nil
}

// coerceFloat parses a string as a floating-point number.
func coerceFloat(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("expected float, got empty string")
	}
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("expected float, got %q", value)
	}
	return v, nil
}

// coerceBoolean parses a string as a boolean.
// Accepted true values: true, 1, yes, on (case-insensitive)
// Accepted false values: false, 0, no, off (case-insensitive)
func coerceBoolean(value string) (bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return false, fmt.Errorf("expected boolean, got empty string")
	}

	switch strings.ToLower(value) {
	case "true", "1", "yes", "on":
		return true, nil
	case "false", "0", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("expected boolean, got %q", value)
	}
}
