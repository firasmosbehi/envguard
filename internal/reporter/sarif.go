package reporter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/envguard/envguard/internal/audit"
	"github.com/envguard/envguard/internal/secrets"
	"github.com/envguard/envguard/internal/sync"
	"github.com/envguard/envguard/internal/validator"
)

// SARIFVersion is the SARIF spec version we emit.
const SARIFVersion = "2.1.0"

// SARIFSchemaURL is the canonical SARIF 2.1.0 JSON schema URI.
const SARIFSchemaURL = "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json"

// sarifLog is the top-level SARIF document.
type sarifLog struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

// sarifRun represents a single analysis run.
type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

// sarifTool describes the analysis tool.
type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

// sarifDriver describes the tool driver.
type sarifDriver struct {
	Name    string      `json:"name"`
	Version string      `json:"version,omitempty"`
	Rules   []sarifRule `json:"rules,omitempty"`
}

// sarifRule describes a rule that was evaluated.
type sarifRule struct {
	ID               string        `json:"id"`
	Name             string        `json:"name,omitempty"`
	ShortDescription *sarifMessage `json:"shortDescription,omitempty"`
	HelpURI          string        `json:"helpUri,omitempty"`
	DefaultConfig    *sarifConfig  `json:"defaultConfiguration,omitempty"`
}

// sarifConfig holds default rule configuration.
type sarifConfig struct {
	Level string `json:"level,omitempty"`
}

// sarifResult is a single finding.
type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations,omitempty"`
}

// sarifMessage holds a text or markdown message.
type sarifMessage struct {
	Text string `json:"text"`
}

// sarifLocation points to where the issue was found.
type sarifLocation struct {
	PhysicalLocation *sarifPhysicalLocation `json:"physicalLocation,omitempty"`
	LogicalLocations []sarifLogicalLocation `json:"logicalLocations,omitempty"`
}

// sarifPhysicalLocation points to a file.
type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
}

// sarifArtifactLocation identifies a file.
type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

// sarifLogicalLocation identifies a logical location (e.g., env var).
type sarifLogicalLocation struct {
	Name               string `json:"name"`
	Kind               string `json:"kind"`
	FullyQualifiedName string `json:"fullyQualifiedName,omitempty"`
}

// severityToSARIFLevel maps EnvGuard severity to SARIF level.
func severityToSARIFLevel(sev validator.Severity) string {
	switch sev {
	case validator.SeverityError:
		return "error"
	case validator.SeverityWarn:
		return "warning"
	case validator.SeverityInfo:
		return "note"
	default:
		return "error"
	}
}

// ruleIDForError creates a stable rule ID from a validation error.
func ruleIDForError(err validator.ValidationError) string {
	if err.Rule != "" {
		return fmt.Sprintf("envguard/%s", err.Rule)
	}
	return "envguard/validation"
}

// helpURIForRule returns a help URI for a given rule.
func helpURIForRule(rule string) string {
	return fmt.Sprintf("https://github.com/envguard/envguard/blob/main/docs/rules.md#%s", rule)
}

// SARIF writes a SARIF 2.1.0 report for validation results to w.
func SARIF(w io.Writer, result *validator.Result, envFilePaths []string, version string) error {
	// Collect unique rules
	ruleMap := make(map[string]sarifRule)
	addRule := func(err validator.ValidationError) {
		id := ruleIDForError(err)
		if _, exists := ruleMap[id]; exists {
			return
		}
		level := severityToSARIFLevel(err.Severity)
		ruleMap[id] = sarifRule{
			ID:   id,
			Name: err.Rule,
			ShortDescription: &sarifMessage{
				Text: err.Rule,
			},
			HelpURI: helpURIForRule(err.Rule),
			DefaultConfig: &sarifConfig{
				Level: level,
			},
		}
	}

	for _, err := range result.Errors {
		addRule(err)
	}
	for _, warn := range result.Warnings {
		addRule(warn)
	}

	rules := make([]sarifRule, 0, len(ruleMap))
	for _, r := range ruleMap {
		rules = append(rules, r)
	}

	// Build results
	var results []sarifResult

	artifactURI := ".env"
	if len(envFilePaths) > 0 {
		artifactURI = envFilePaths[0]
	}

	for _, err := range result.Errors {
		results = append(results, sarifResult{
			RuleID: ruleIDForError(err),
			Level:  severityToSARIFLevel(err.Severity),
			Message: sarifMessage{
				Text: fmt.Sprintf("%s: %s", err.Key, err.Message),
			},
			Locations: []sarifLocation{
				{
					PhysicalLocation: &sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{
							URI: artifactURI,
						},
					},
					LogicalLocations: []sarifLogicalLocation{
						{
							Name:               err.Key,
							Kind:               "environment-variable",
							FullyQualifiedName: err.Key,
						},
					},
				},
			},
		})
	}

	for _, warn := range result.Warnings {
		results = append(results, sarifResult{
			RuleID: ruleIDForError(warn),
			Level:  "warning",
			Message: sarifMessage{
				Text: fmt.Sprintf("%s: %s", warn.Key, warn.Message),
			},
			Locations: []sarifLocation{
				{
					PhysicalLocation: &sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{
							URI: artifactURI,
						},
					},
					LogicalLocations: []sarifLogicalLocation{
						{
							Name:               warn.Key,
							Kind:               "environment-variable",
							FullyQualifiedName: warn.Key,
						},
					},
				},
			},
		})
	}

	log := sarifLog{
		Version: SARIFVersion,
		Schema:  SARIFSchemaURL,
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:    "EnvGuard",
						Version: version,
						Rules:   rules,
					},
				},
				Results: results,
			},
		},
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(log)
}

// SARIFAudit writes a SARIF 2.1.0 report for audit findings to w.
func SARIFAudit(w io.Writer, result *audit.Result, version string) error {
	// Collect unique rules
	ruleMap := make(map[string]sarifRule)
	for _, f := range result.Findings {
		id := fmt.Sprintf("envguard.audit/%s", f.Type)
		if _, exists := ruleMap[id]; exists {
			continue
		}
		level := "warning"
		switch f.Type {
		case audit.MissingVar, audit.MissingRequired:
			level = "error"
		case audit.UndocumentedVar:
			level = "warning"
		case audit.UnusedVar:
			level = "note"
		}
		ruleMap[id] = sarifRule{
			ID:   id,
			Name: string(f.Type),
			ShortDescription: &sarifMessage{
				Text: string(f.Type),
			},
			HelpURI: helpURIForRule(string(f.Type)),
			DefaultConfig: &sarifConfig{
				Level: level,
			},
		}
	}

	rules := make([]sarifRule, 0, len(ruleMap))
	for _, r := range ruleMap {
		rules = append(rules, r)
	}

	var results []sarifResult
	for _, f := range result.Findings {
		level := "warning"
		switch f.Type {
		case audit.MissingVar, audit.MissingRequired:
			level = "error"
		case audit.UndocumentedVar:
			level = "warning"
		case audit.UnusedVar:
			level = "note"
		}
		var locations []sarifLocation
		if f.File != "" {
			locations = append(locations, sarifLocation{
				PhysicalLocation: &sarifPhysicalLocation{
					ArtifactLocation: sarifArtifactLocation{
						URI: f.File,
					},
				},
			})
		}
		results = append(results, sarifResult{
			RuleID:    fmt.Sprintf("envguard.audit/%s", f.Type),
			Level:     level,
			Message:   sarifMessage{Text: f.Message},
			Locations: locations,
		})
	}

	log := sarifLog{
		Version: SARIFVersion,
		Schema:  SARIFSchemaURL,
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:    "EnvGuard",
						Version: version,
						Rules:   rules,
					},
				},
				Results: results,
			},
		},
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(log)
}

// SARIFScan writes a SARIF 2.1.0 report for secret scan results to w.
func SARIFScan(w io.Writer, matches []secrets.SecretMatch, envFilePaths []string, version string) error {
	// Collect unique rules
	ruleMap := make(map[string]sarifRule)
	for _, m := range matches {
		id := fmt.Sprintf("envguard.secrets/%s", m.Rule)
		if _, exists := ruleMap[id]; exists {
			continue
		}
		ruleMap[id] = sarifRule{
			ID:   id,
			Name: m.Rule,
			ShortDescription: &sarifMessage{
				Text: m.Message,
			},
			HelpURI: helpURIForRule(m.Rule),
			DefaultConfig: &sarifConfig{
				Level: "error",
			},
		}
	}

	rules := make([]sarifRule, 0, len(ruleMap))
	for _, r := range ruleMap {
		rules = append(rules, r)
	}

	artifactURI := ".env"
	if len(envFilePaths) > 0 {
		artifactURI = envFilePaths[0]
	}

	var results []sarifResult
	for _, m := range matches {
		results = append(results, sarifResult{
			RuleID: fmt.Sprintf("envguard.secrets/%s", m.Rule),
			Level:  "error",
			Message: sarifMessage{
				Text: fmt.Sprintf("%s: %s (redacted: %s)", m.Key, m.Message, m.Redacted),
			},
			Locations: []sarifLocation{
				{
					PhysicalLocation: &sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{
							URI: artifactURI,
						},
					},
					LogicalLocations: []sarifLogicalLocation{
						{
							Name:               m.Key,
							Kind:               "environment-variable",
							FullyQualifiedName: m.Key,
						},
					},
				},
			},
		})
	}

	log := sarifLog{
		Version: SARIFVersion,
		Schema:  SARIFSchemaURL,
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:    "EnvGuard",
						Version: version,
						Rules:   rules,
					},
				},
				Results: results,
			},
		},
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(log)
}

// LintFinding represents a single lint finding.
type LintFinding struct {
	Level   string `json:"level"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// SARIFLint writes a SARIF 2.1.0 report for lint findings to w.
func SARIFLint(w io.Writer, findings []LintFinding, version string) error {
	// Collect unique rules
	ruleMap := make(map[string]sarifRule)
	for _, f := range findings {
		id := fmt.Sprintf("envguard.lint/%s", f.Rule)
		if _, exists := ruleMap[id]; exists {
			continue
		}
		level := f.Level
		if level == "" {
			level = "warning"
		}
		ruleMap[id] = sarifRule{
			ID:   id,
			Name: f.Rule,
			ShortDescription: &sarifMessage{
				Text: f.Rule,
			},
			HelpURI: helpURIForRule(f.Rule),
			DefaultConfig: &sarifConfig{
				Level: level,
			},
		}
	}

	rules := make([]sarifRule, 0, len(ruleMap))
	for _, r := range ruleMap {
		rules = append(rules, r)
	}

	var results []sarifResult
	for _, f := range findings {
		level := f.Level
		if level == "" {
			level = "warning"
		}
		results = append(results, sarifResult{
			RuleID: fmt.Sprintf("envguard.lint/%s", f.Rule),
			Level:  level,
			Message: sarifMessage{
				Text: f.Message,
			},
			Locations: []sarifLocation{
				{
					PhysicalLocation: &sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{
							URI: "envguard.yaml",
						},
					},
				},
			},
		})
	}

	log := sarifLog{
		Version: SARIFVersion,
		Schema:  SARIFSchemaURL,
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:    "EnvGuard",
						Version: version,
						Rules:   rules,
					},
				},
				Results: results,
			},
		},
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(log)
}

// SARIFSync writes a SARIF 2.1.0 report for sync drift to w.
func SARIFSync(w io.Writer, result *sync.Result, version string) error {
	ruleMap := make(map[string]sarifRule)
	for _, d := range result.Diffs {
		id := fmt.Sprintf("envguard.sync/%s", d.Type)
		if _, exists := ruleMap[id]; exists {
			continue
		}
		level := "warning"
		if d.Type == "missing-in-example" || d.Type == "missing-in-env" {
			level = "warning"
		}
		ruleMap[id] = sarifRule{
			ID:   id,
			Name: d.Type,
			ShortDescription: &sarifMessage{
				Text: d.Type,
			},
			HelpURI: helpURIForRule(d.Type),
			DefaultConfig: &sarifConfig{
				Level: level,
			},
		}
	}

	rules := make([]sarifRule, 0, len(ruleMap))
	for _, r := range ruleMap {
		rules = append(rules, r)
	}

	var results []sarifResult
	for _, d := range result.Diffs {
		msg := d.Key
		if d.EnvVal != "" {
			msg += "=" + d.EnvVal
		}
		if d.ExVal != "" {
			msg += " (example: " + d.ExVal + ")"
		}
		results = append(results, sarifResult{
			RuleID:  fmt.Sprintf("envguard.sync/%s", d.Type),
			Level:   "warning",
			Message: sarifMessage{Text: msg},
		})
	}

	log := sarifLog{
		Version: SARIFVersion,
		Schema:  SARIFSchemaURL,
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:    "EnvGuard",
						Version: version,
						Rules:   rules,
					},
				},
				Results: results,
			},
		},
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(log)
}
