// Package audit provides source code analysis for environment variable usage.
package audit

// EnvRef represents a reference to an environment variable in source code.
type EnvRef struct {
	Var     string `json:"var"`
	File    string `json:"file"`
	Line    int    `json:"line"`
	Context string `json:"context"`
}

// FindingType categorizes audit findings.
type FindingType string

const (
	// MissingVar means the variable is used in code but not defined in .env.
	MissingVar FindingType = "missing"
	// UnusedVar means the variable is in .env but never referenced in code.
	UnusedVar FindingType = "unused"
	// UndocumentedVar means the variable is used in code but not in the schema.
	UndocumentedVar FindingType = "undocumented"
	// MissingRequired means the variable is required by schema but missing from .env.
	MissingRequired FindingType = "missing-required"
)

// Finding represents a single audit issue.
type Finding struct {
	Type    FindingType `json:"type"`
	Var     string      `json:"var"`
	File    string      `json:"file,omitempty"`
	Line    int         `json:"line,omitempty"`
	Message string      `json:"message"`
}

// Result holds the complete audit output.
type Result struct {
	Findings []Finding `json:"findings"`
}

// NewResult creates an empty audit result.
func NewResult() *Result {
	return &Result{Findings: make([]Finding, 0)}
}

// AddFinding records an audit finding.
func (r *Result) AddFinding(f Finding) {
	r.Findings = append(r.Findings, f)
}

// HasFindings returns true if there are any findings.
func (r *Result) HasFindings() bool {
	return len(r.Findings) > 0
}

// CountByType returns the number of findings of a given type.
func (r *Result) CountByType(ft FindingType) int {
	count := 0
	for _, f := range r.Findings {
		if f.Type == ft {
			count++
		}
	}
	return count
}
