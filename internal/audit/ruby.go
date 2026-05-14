package audit

import "regexp"

// rubyExtractor finds Ruby env var references.
type rubyExtractor struct{}

func (e *rubyExtractor) Extensions() []string { return []string{".rb"} }

var rubyPatterns = []*regexp.Regexp{
	// ENV['VAR'] or ENV["VAR"]
	regexp.MustCompile(`ENV\[["']([^"']+)["']\]`),
	// ENV.fetch('VAR') or ENV.fetch("VAR")
	regexp.MustCompile(`ENV\.fetch\s*\(\s*["']([^"']+)["']\s*\)`),
	// ENV.fetch('VAR', default)
	regexp.MustCompile(`ENV\.fetch\s*\(\s*["']([^"']+)["']\s*,`),
}

func (e *rubyExtractor) Extract(filePath string, content string) []EnvRef {
	var refs []EnvRef
	for _, re := range rubyPatterns {
		refs = append(refs, extractRegex(filePath, content, re)...)
	}
	return deduplicateRefs(refs)
}
