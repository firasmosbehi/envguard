package audit

import "regexp"

// goExtractor finds Go env var references.
type goExtractor struct{}

func (e *goExtractor) Extensions() []string { return []string{".go"} }

var goPatterns = []*regexp.Regexp{
	// os.Getenv("VAR")
	regexp.MustCompile(`os\.Getenv\s*\(\s*["']([^"']+)["']\s*\)`),
	// os.LookupEnv("VAR")
	regexp.MustCompile(`os\.LookupEnv\s*\(\s*["']([^"']+)["']\s*\)`),
	// os.Setenv("VAR", ...)
	regexp.MustCompile(`os\.Setenv\s*\(\s*["']([^"']+)["']\s*,`),
	// os.Unsetenv("VAR")
	regexp.MustCompile(`os\.Unsetenv\s*\(\s*["']([^"']+)["']\s*\)`),
}

func (e *goExtractor) Extract(filePath string, content string) []EnvRef {
	var refs []EnvRef
	for _, re := range goPatterns {
		refs = append(refs, extractRegex(filePath, content, re)...)
	}
	return deduplicateRefs(refs)
}
