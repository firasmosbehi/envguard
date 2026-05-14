package audit

import "regexp"

// rustExtractor finds Rust env var references.
type rustExtractor struct{}

func (e *rustExtractor) Extensions() []string { return []string{".rs"} }

var rustPatterns = []*regexp.Regexp{
	// std::env::var("VAR")
	regexp.MustCompile(`std::env::var\s*\(\s*["']([^"']+)["']\s*\)`),
	// std::env::var_os("VAR")
	regexp.MustCompile(`std::env::var_os\s*\(\s*["']([^"']+)["']\s*\)`),
	// env::var("VAR") (with use std::env)
	regexp.MustCompile(`env::var\s*\(\s*["']([^"']+)["']\s*\)`),
	// env::var_os("VAR")
	regexp.MustCompile(`env::var_os\s*\(\s*["']([^"']+)["']\s*\)`),
	// std::env::set_var("VAR", ...)
	regexp.MustCompile(`std::env::set_var\s*\(\s*["']([^"']+)["']\s*,`),
}

func (e *rustExtractor) Extract(filePath string, content string) []EnvRef {
	var refs []EnvRef
	for _, re := range rustPatterns {
		refs = append(refs, extractRegex(filePath, content, re)...)
	}
	return deduplicateRefs(refs)
}
