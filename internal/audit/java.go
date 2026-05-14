package audit

import "regexp"

// javaExtractor finds Java env var references.
type javaExtractor struct{}

func (e *javaExtractor) Extensions() []string { return []string{".java", ".kt"} }

var javaPatterns = []*regexp.Regexp{
	// System.getenv("VAR")
	regexp.MustCompile(`System\.getenv\s*\(\s*["']([^"']+)["']\s*\)`),
	// System.getProperty("VAR")
	regexp.MustCompile(`System\.getProperty\s*\(\s*["']([^"']+)["']\s*\)`),
}

func (e *javaExtractor) Extract(filePath string, content string) []EnvRef {
	var refs []EnvRef
	for _, re := range javaPatterns {
		refs = append(refs, extractRegex(filePath, content, re)...)
	}
	return deduplicateRefs(refs)
}
