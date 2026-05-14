package audit

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Extractor finds environment variable references in source files.
type Extractor interface {
	// Extensions returns the file extensions this extractor handles (e.g., [".go"]).
	Extensions() []string
	// Extract finds env var references in the given file content.
	Extract(filePath string, content string) []EnvRef
}

// registry holds all registered extractors.
var registry = []Extractor{
	&goExtractor{},
	&nodeExtractor{},
	&pythonExtractor{},
	&rustExtractor{},
	&rubyExtractor{},
	&javaExtractor{},
}

// walkFiles walks the source directory and extracts env refs from supported files.
func walkFiles(dir string, exclude []string) ([]EnvRef, error) {
	var refs []EnvRef

	excludeMatchers := make([]*regexp.Regexp, 0, len(exclude))
	for _, pattern := range exclude {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid exclude pattern %q: %w", pattern, err)
		}
		excludeMatchers = append(excludeMatchers, re)
	}

	isExcluded := func(path string) bool {
		for _, re := range excludeMatchers {
			if re.MatchString(path) {
				return true
			}
		}
		return false
	}

	extMap := make(map[string]Extractor)
	for _, ex := range registry {
		for _, ext := range ex.Extensions() {
			extMap[ext] = ex
		}
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip unreadable files
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(dir, path)
		if isExcluded(rel) || isExcluded(path) {
			return nil
		}
		ext := filepath.Ext(path)
		extractor, ok := extMap[ext]
		if !ok {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}
		content := string(data)
		fileRefs := extractor.Extract(rel, content)
		refs = append(refs, fileRefs...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return refs, nil
}

// extractRegex finds matches in content using a regex.
// The regex should have at least one capture group for the variable name.
// Only the first capture group per match is extracted.
func extractRegex(filePath string, content string, re *regexp.Regexp) []EnvRef {
	var refs []EnvRef
	scanner := bufio.NewScanner(strings.NewReader(content))
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		matches := re.FindAllStringSubmatchIndex(line, -1)
		for _, m := range matches {
			// m[0], m[1] = full match; m[2], m[3] = first capture group
			if len(m) >= 4 && m[2] >= 0 && m[3] >= 0 {
				name := line[m[2]:m[3]]
				if name != "" {
					refs = append(refs, EnvRef{
						Var:     name,
						File:    filePath,
						Line:    lineNum,
						Context: strings.TrimSpace(line),
					})
				}
			}
		}
	}
	return refs
}

// deduplicateRefs removes duplicate references (same var+file+line).
func deduplicateRefs(refs []EnvRef) []EnvRef {
	seen := make(map[string]bool)
	var result []EnvRef
	for _, r := range refs {
		key := fmt.Sprintf("%s:%s:%d", r.Var, r.File, r.Line)
		if !seen[key] {
			seen[key] = true
			result = append(result, r)
		}
	}
	return result
}
