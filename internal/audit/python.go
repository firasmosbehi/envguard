package audit

import "regexp"

// pythonExtractor finds Python env var references.
type pythonExtractor struct{}

func (e *pythonExtractor) Extensions() []string { return []string{".py"} }

var pythonPatterns = []*regexp.Regexp{
	// os.environ['VAR'] or os.environ["VAR"]
	regexp.MustCompile(`os\.environ\[["']([^"']+)["']\]`),
	// os.environ.get('VAR') or os.environ.get("VAR")
	regexp.MustCompile(`os\.environ\.get\s*\(\s*["']([^"']+)["']\s*\)`),
	// os.getenv('VAR') or os.getenv("VAR")
	regexp.MustCompile(`os\.getenv\s*\(\s*["']([^"']+)["']\s*\)`),
	// os.putenv('VAR', ...) or os.putenv("VAR", ...)
	regexp.MustCompile(`os\.putenv\s*\(\s*["']([^"']+)["']\s*,`),
	// environ.get('VAR') (common shortcut)
	regexp.MustCompile(`environ\.get\s*\(\s*["']([^"']+)["']\s*\)`),
}

func (e *pythonExtractor) Extract(filePath string, content string) []EnvRef {
	var refs []EnvRef
	for _, re := range pythonPatterns {
		refs = append(refs, extractRegex(filePath, content, re)...)
	}
	return deduplicateRefs(refs)
}
