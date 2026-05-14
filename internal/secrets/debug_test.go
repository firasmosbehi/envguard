package secrets

import (
	"fmt"
	"testing"
)

func TestDebugScan(t *testing.T) {
	scanner := DefaultScanner()
	envVars := map[string]string{
		"API_KEY": "api_key=abcdefghijklmnopqrstuvwxyz1234",
		"DATA":    "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890",
	}
	matches := scanner.Scan(envVars)
	for _, m := range matches {
		fmt.Printf("Key=%s Rule=%s Severity=%s\n", m.Key, m.Rule, m.Severity)
	}
	fmt.Printf("Total matches: %d\n", len(matches))
}
