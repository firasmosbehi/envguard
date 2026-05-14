package audit

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/envguard/envguard/internal/dotenv"
	"github.com/envguard/envguard/internal/schema"
)

// Options configures the audit behavior.
type Options struct {
	SrcDir     string
	EnvPath    string
	SchemaPath string
	Exclude    []string
	IgnoreVars []string
	Strict     bool
}

// Run performs the source code audit.
func Run(opts Options) (*Result, error) {
	result := NewResult()

	// Load .env file if provided
	envVars := make(map[string]string)
	if opts.EnvPath != "" {
		vars, err := dotenv.Parse(opts.EnvPath)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to parse .env file: %w", err)
		}
		envVars = vars
	}

	// Load schema if provided
	var sch *schema.Schema
	if opts.SchemaPath != "" {
		s, err := schema.Parse(opts.SchemaPath)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to parse schema file: %w", err)
		}
		sch = s
	}

	// Extract env var references from source code
	refs, err := walkFiles(opts.SrcDir, opts.Exclude)
	if err != nil {
		return nil, fmt.Errorf("failed to scan source files: %w", err)
	}

	// Build sets for fast lookup
	ignoreSet := make(map[string]bool)
	for _, v := range opts.IgnoreVars {
		ignoreSet[strings.TrimSpace(v)] = true
	}

	envSet := make(map[string]bool)
	for k := range envVars {
		envSet[k] = true
	}

	schemaSet := make(map[string]bool)
	requiredSet := make(map[string]bool)
	if sch != nil {
		for name, v := range sch.Env {
			schemaSet[name] = true
			if v.Required {
				requiredSet[name] = true
			}
		}
	}

	codeSet := make(map[string]bool)
	codeRefs := make(map[string][]EnvRef)
	for _, ref := range refs {
		if ignoreSet[ref.Var] {
			continue
		}
		codeSet[ref.Var] = true
		codeRefs[ref.Var] = append(codeRefs[ref.Var], ref)
	}

	// Find missing: used in code but not in .env
	for refVar := range codeSet {
		if !envSet[refVar] {
			refs := codeRefs[refVar]
			var loc string
			if len(refs) > 0 {
				loc = fmt.Sprintf("%s:%d", refs[0].File, refs[0].Line)
			}
			result.AddFinding(Finding{
				Type:    MissingVar,
				Var:     refVar,
				File:    refs[0].File,
				Line:    refs[0].Line,
				Message: fmt.Sprintf("%s is used in code (%s) but not defined in .env", refVar, loc),
			})
		}
	}

	// Find undocumented: used in code but not in schema
	if sch != nil {
		for refVar := range codeSet {
			if !schemaSet[refVar] {
				refs := codeRefs[refVar]
				var loc string
				if len(refs) > 0 {
					loc = fmt.Sprintf("%s:%d", refs[0].File, refs[0].Line)
				}
				result.AddFinding(Finding{
					Type:    UndocumentedVar,
					Var:     refVar,
					File:    refs[0].File,
					Line:    refs[0].Line,
					Message: fmt.Sprintf("%s is used in code (%s) but not documented in schema", refVar, loc),
				})
			}
		}
	}

	// Find unused: in .env but not in code
	for envVar := range envSet {
		if !codeSet[envVar] && !ignoreSet[envVar] {
			result.AddFinding(Finding{
				Type:    UnusedVar,
				Var:     envVar,
				Message: fmt.Sprintf("%s is defined in .env but never referenced in code", envVar),
			})
		}
	}

	// Find missing required: required by schema but missing from .env
	if sch != nil {
		for reqVar := range requiredSet {
			if !envSet[reqVar] {
				result.AddFinding(Finding{
					Type:    MissingRequired,
					Var:     reqVar,
					Message: fmt.Sprintf("%s is required by schema but missing from .env", reqVar),
				})
			}
		}
	}

	// Sort findings for consistent output
	sort.Slice(result.Findings, func(i, j int) bool {
		if result.Findings[i].Type != result.Findings[j].Type {
			return result.Findings[i].Type < result.Findings[j].Type
		}
		return result.Findings[i].Var < result.Findings[j].Var
	})

	return result, nil
}

// KnownRuntimeVars returns a list of commonly ignored runtime environment variables.
func KnownRuntimeVars() []string {
	return []string{
		"CI", "HOME", "USER", "USERNAME", "PWD", "SHELL", "PATH",
		"NODE_ENV", "RAILS_ENV", "RACK_ENV", "APP_ENV", "ENV",
		"HOSTNAME", "TERM", "LANG", "LC_ALL", "EDITOR", "PS1",
		"SSH_CONNECTION", "SSH_CLIENT", "SSH_TTY", "DISPLAY",
		"XDG_SESSION_TYPE", "XDG_CURRENT_DESKTOP",
		"TMPDIR", "TMP", "TEMP",
		"GOPATH", "GOROOT", "GOVERSION",
		"PYTHONPATH", "VIRTUAL_ENV",
		"JAVA_HOME", "JDK_HOME",
	}
}
