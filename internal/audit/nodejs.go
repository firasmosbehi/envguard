package audit

import "regexp"

// nodeExtractor finds Node.js/TypeScript env var references.
type nodeExtractor struct{}

func (e *nodeExtractor) Extensions() []string {
	return []string{".js", ".ts", ".jsx", ".tsx", ".mjs", ".cjs"}
}

var nodePatterns = []*regexp.Regexp{
	// process.env.VAR (dot notation)
	regexp.MustCompile(`process\.env\.(\w+)`),
	// process.env["VAR"] or process.env['VAR']
	regexp.MustCompile(`process\.env\[["']([^"']+)["']\]`),
	// import.meta.env.VAR (Vite)
	regexp.MustCompile(`import\.meta\.env\.(\w+)`),
	// Deno.env.get("VAR")
	regexp.MustCompile(`Deno\.env\.get\s*\(\s*["']([^"']+)["']\s*\)`),
	// Bun.env.VAR
	regexp.MustCompile(`Bun\.env\.(\w+)`),
}

func (e *nodeExtractor) Extract(filePath string, content string) []EnvRef {
	var refs []EnvRef
	for _, re := range nodePatterns {
		refs = append(refs, extractRegex(filePath, content, re)...)
	}
	return deduplicateRefs(refs)
}
