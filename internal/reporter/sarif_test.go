package reporter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/envguard/envguard/internal/audit"
	"github.com/envguard/envguard/internal/secrets"
	"github.com/envguard/envguard/internal/sync"
	"github.com/envguard/envguard/internal/validator"
)

func TestSARIF_ValidResult(t *testing.T) {
	result := validator.NewResult()
	var buf bytes.Buffer
	if err := SARIF(&buf, result, []string{".env"}, "2.0.0"); err != nil {
		t.Fatalf("SARIF failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	if log.Version != SARIFVersion {
		t.Errorf("expected version %q, got %q", SARIFVersion, log.Version)
	}
	if log.Schema != SARIFSchemaURL {
		t.Errorf("expected schema %q, got %q", SARIFSchemaURL, log.Schema)
	}
	if len(log.Runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(log.Runs))
	}
	run := log.Runs[0]
	if run.Tool.Driver.Name != "EnvGuard" {
		t.Errorf("expected tool name EnvGuard, got %q", run.Tool.Driver.Name)
	}
	if run.Tool.Driver.Version != "2.0.0" {
		t.Errorf("expected version 2.0.0, got %q", run.Tool.Driver.Version)
	}
	if len(run.Results) != 0 {
		t.Errorf("expected 0 results for valid run, got %d", len(run.Results))
	}
}

func TestSARIF_WithErrors(t *testing.T) {
	result := validator.NewResult()
	result.AddErrorWithSeverity("API_KEY", "required", "API_KEY is required", validator.SeverityError)
	result.AddErrorWithSeverity("DEBUG", "type", "DEBUG must be a boolean", validator.SeverityWarn)

	var buf bytes.Buffer
	if err := SARIF(&buf, result, []string{".env"}, "2.0.0"); err != nil {
		t.Fatalf("SARIF failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	run := log.Runs[0]
	if len(run.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(run.Results))
	}

	// Check rules are collected
	if len(run.Tool.Driver.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(run.Tool.Driver.Rules))
	}

	// Check first result
	r0 := run.Results[0]
	if r0.RuleID != "envguard/required" {
		t.Errorf("expected ruleId envguard/required, got %q", r0.RuleID)
	}
	if r0.Level != "error" {
		t.Errorf("expected level error, got %q", r0.Level)
	}
	if !strings.Contains(r0.Message.Text, "API_KEY") {
		t.Errorf("expected message to contain API_KEY, got %q", r0.Message.Text)
	}

	// Check second result
	r1 := run.Results[1]
	if r1.RuleID != "envguard/type" {
		t.Errorf("expected ruleId envguard/type, got %q", r1.RuleID)
	}
	if r1.Level != "warning" {
		t.Errorf("expected level warning, got %q", r1.Level)
	}
}

func TestSARIF_WithWarnings(t *testing.T) {
	result := validator.NewResult()
	result.AddWarning("OLD_VAR", "deprecated", "OLD_VAR is deprecated")

	var buf bytes.Buffer
	if err := SARIF(&buf, result, []string{".env"}, "2.0.0"); err != nil {
		t.Fatalf("SARIF failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	run := log.Runs[0]
	if len(run.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(run.Results))
	}
	if run.Results[0].Level != "warning" {
		t.Errorf("expected level warning, got %q", run.Results[0].Level)
	}
}

func TestSARIF_Locations(t *testing.T) {
	result := validator.NewResult()
	result.AddError("DB_HOST", "required", "DB_HOST is required")

	var buf bytes.Buffer
	if err := SARIF(&buf, result, []string{"config/.env"}, "2.0.0"); err != nil {
		t.Fatalf("SARIF failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	loc := log.Runs[0].Results[0].Locations[0]
	if loc.PhysicalLocation.ArtifactLocation.URI != "config/.env" {
		t.Errorf("expected artifact URI config/.env, got %q", loc.PhysicalLocation.ArtifactLocation.URI)
	}
	if loc.LogicalLocations[0].Name != "DB_HOST" {
		t.Errorf("expected logical location DB_HOST, got %q", loc.LogicalLocations[0].Name)
	}
	if loc.LogicalLocations[0].Kind != "environment-variable" {
		t.Errorf("expected kind environment-variable, got %q", loc.LogicalLocations[0].Kind)
	}
}

func TestSARIFScan(t *testing.T) {
	matches := []secrets.SecretMatch{
		{Key: "AWS_KEY", Rule: "aws-access-key", Message: "AWS key detected", Redacted: "AKIA..."},
		{Key: "GITHUB_TOKEN", Rule: "github-token", Message: "GitHub token detected", Redacted: "ghp_..."},
	}

	var buf bytes.Buffer
	if err := SARIFScan(&buf, matches, []string{".env"}, "2.0.0"); err != nil {
		t.Fatalf("SARIFScan failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	run := log.Runs[0]
	if len(run.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(run.Results))
	}
	if run.Results[0].RuleID != "envguard.secrets/aws-access-key" {
		t.Errorf("unexpected ruleId: %q", run.Results[0].RuleID)
	}
	if run.Results[0].Level != "error" {
		t.Errorf("expected level error, got %q", run.Results[0].Level)
	}
	if len(run.Tool.Driver.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(run.Tool.Driver.Rules))
	}
}

func TestSARIFLint(t *testing.T) {
	findings := []LintFinding{
		{Level: "error", Rule: "redundant", Message: "variable X: required and default are mutually exclusive"},
		{Level: "warning", Rule: "missing-description", Message: "variable Y: missing description"},
	}

	var buf bytes.Buffer
	if err := SARIFLint(&buf, findings, "2.0.0"); err != nil {
		t.Fatalf("SARIFLint failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	run := log.Runs[0]
	if len(run.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(run.Results))
	}
	if run.Results[0].Level != "error" {
		t.Errorf("expected level error, got %q", run.Results[0].Level)
	}
	if run.Results[1].Level != "warning" {
		t.Errorf("expected level warning, got %q", run.Results[1].Level)
	}
	if run.Results[0].Locations[0].PhysicalLocation.ArtifactLocation.URI != "envguard.yaml" {
		t.Errorf("expected artifact URI envguard.yaml, got %q", run.Results[0].Locations[0].PhysicalLocation.ArtifactLocation.URI)
	}
}

func TestSARIFScan_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := SARIFScan(&buf, []secrets.SecretMatch{}, []string{".env"}, "2.0.0"); err != nil {
		t.Fatalf("SARIFScan failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	if len(log.Runs[0].Results) != 0 {
		t.Errorf("expected 0 results for empty scan, got %d", len(log.Runs[0].Results))
	}
}

func TestSeverityToSARIFLevel(t *testing.T) {
	tests := []struct {
		sev      validator.Severity
		expected string
	}{
		{validator.SeverityError, "error"},
		{validator.SeverityWarn, "warning"},
		{validator.SeverityInfo, "note"},
		{validator.Severity("unknown"), "error"},
	}

	for _, tt := range tests {
		t.Run(string(tt.sev), func(t *testing.T) {
			got := severityToSARIFLevel(tt.sev)
			if got != tt.expected {
				t.Errorf("severityToSARIFLevel(%q) = %q, want %q", tt.sev, got, tt.expected)
			}
		})
	}
}

func TestRuleIDForError_EmptyRule(t *testing.T) {
	err := validator.ValidationError{Key: "FOO", Rule: "", Message: "something wrong"}
	got := ruleIDForError(err)
	if got != "envguard/validation" {
		t.Errorf("ruleIDForError with empty rule = %q, want %q", got, "envguard/validation")
	}
}

func TestRuleIDForError_WithRule(t *testing.T) {
	err := validator.ValidationError{Key: "FOO", Rule: "required", Message: "missing"}
	got := ruleIDForError(err)
	if got != "envguard/required" {
		t.Errorf("ruleIDForError with rule = %q, want %q", got, "envguard/required")
	}
}

func TestSARIF_EmptyEnvFilePaths(t *testing.T) {
	result := validator.NewResult()
	result.AddError("DB_HOST", "required", "DB_HOST is required")

	var buf bytes.Buffer
	if err := SARIF(&buf, result, []string{}, "2.0.0"); err != nil {
		t.Fatalf("SARIF failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	loc := log.Runs[0].Results[0].Locations[0]
	if loc.PhysicalLocation.ArtifactLocation.URI != ".env" {
		t.Errorf("expected default artifact URI .env, got %q", loc.PhysicalLocation.ArtifactLocation.URI)
	}
}

func TestSARIFAudit_Empty(t *testing.T) {
	result := audit.NewResult()

	var buf bytes.Buffer
	if err := SARIFAudit(&buf, result, "2.0.0"); err != nil {
		t.Fatalf("SARIFAudit failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	if len(log.Runs[0].Results) != 0 {
		t.Errorf("expected 0 results for empty audit, got %d", len(log.Runs[0].Results))
	}
	if len(log.Runs[0].Tool.Driver.Rules) != 0 {
		t.Errorf("expected 0 rules for empty audit, got %d", len(log.Runs[0].Tool.Driver.Rules))
	}
}

func TestSARIFAudit_AllFindingTypes(t *testing.T) {
	result := audit.NewResult()
	result.AddFinding(audit.Finding{Type: audit.MissingVar, Var: "MISSING", File: "app.go", Line: 10, Message: "MISSING is used in code but not defined in .env"})
	result.AddFinding(audit.Finding{Type: audit.MissingRequired, Var: "REQ", Message: "REQ is required by schema but missing from .env"})
	result.AddFinding(audit.Finding{Type: audit.UndocumentedVar, Var: "SECRET", File: "config.go", Line: 5, Message: "SECRET is undocumented"})
	result.AddFinding(audit.Finding{Type: audit.UnusedVar, Var: "OLD", Message: "OLD is unused"})

	var buf bytes.Buffer
	if err := SARIFAudit(&buf, result, "2.0.0"); err != nil {
		t.Fatalf("SARIFAudit failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	run := log.Runs[0]
	if len(run.Results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(run.Results))
	}
	if len(run.Tool.Driver.Rules) != 4 {
		t.Fatalf("expected 4 rules, got %d", len(run.Tool.Driver.Rules))
	}

	levelMap := make(map[string]string)
	for _, r := range run.Results {
		levelMap[r.RuleID] = r.Level
	}

	if levelMap["envguard.audit/missing"] != "error" {
		t.Errorf("expected missing level error, got %q", levelMap["envguard.audit/missing"])
	}
	if levelMap["envguard.audit/missing-required"] != "error" {
		t.Errorf("expected missing-required level error, got %q", levelMap["envguard.audit/missing-required"])
	}
	if levelMap["envguard.audit/undocumented"] != "warning" {
		t.Errorf("expected undocumented level warning, got %q", levelMap["envguard.audit/undocumented"])
	}
	if levelMap["envguard.audit/unused"] != "note" {
		t.Errorf("expected unused level note, got %q", levelMap["envguard.audit/unused"])
	}
}

func TestSARIFAudit_NoFile(t *testing.T) {
	result := audit.NewResult()
	result.AddFinding(audit.Finding{Type: audit.UnusedVar, Var: "OLD", Message: "OLD is unused"})

	var buf bytes.Buffer
	if err := SARIFAudit(&buf, result, "2.0.0"); err != nil {
		t.Fatalf("SARIFAudit failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	result0 := log.Runs[0].Results[0]
	if len(result0.Locations) != 0 {
		t.Errorf("expected 0 locations when File is empty, got %d", len(result0.Locations))
	}
}

func TestSARIFAudit_DuplicateRules(t *testing.T) {
	result := audit.NewResult()
	result.AddFinding(audit.Finding{Type: audit.MissingVar, Var: "A", Message: "A missing"})
	result.AddFinding(audit.Finding{Type: audit.MissingVar, Var: "B", Message: "B missing"})

	var buf bytes.Buffer
	if err := SARIFAudit(&buf, result, "2.0.0"); err != nil {
		t.Fatalf("SARIFAudit failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	// Both findings have same type, so only one rule should be created
	if len(log.Runs[0].Tool.Driver.Rules) != 1 {
		t.Errorf("expected 1 unique rule, got %d", len(log.Runs[0].Tool.Driver.Rules))
	}
	if len(log.Runs[0].Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(log.Runs[0].Results))
	}
}

func TestSARIFSync_Empty(t *testing.T) {
	result := &sync.Result{Diffs: []sync.Diff{}}

	var buf bytes.Buffer
	if err := SARIFSync(&buf, result, "2.0.0"); err != nil {
		t.Fatalf("SARIFSync failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	if len(log.Runs[0].Results) != 0 {
		t.Errorf("expected 0 results for empty sync, got %d", len(log.Runs[0].Results))
	}
}

func TestSARIFSync_WithDiffs(t *testing.T) {
	result := &sync.Result{
		Diffs: []sync.Diff{
			{Type: "missing-in-example", Key: "NEW_VAR", EnvVal: "secret"},
			{Type: "missing-in-env", Key: "OLD_VAR", ExVal: "example-value"},
			{Type: "value-mismatch", Key: "DIFF_VAR", EnvVal: "a", ExVal: "b"},
		},
	}

	var buf bytes.Buffer
	if err := SARIFSync(&buf, result, "2.0.0"); err != nil {
		t.Fatalf("SARIFSync failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	run := log.Runs[0]
	if len(run.Results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(run.Results))
	}
	if len(run.Tool.Driver.Rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(run.Tool.Driver.Rules))
	}

	msgMap := make(map[string]string)
	for _, r := range run.Results {
		msgMap[r.RuleID] = r.Message.Text
	}

	if !strings.Contains(msgMap["envguard.sync/missing-in-example"], "NEW_VAR") {
		t.Errorf("expected missing-in-example message to contain NEW_VAR, got %q", msgMap["envguard.sync/missing-in-example"])
	}
	if !strings.Contains(msgMap["envguard.sync/missing-in-env"], "OLD_VAR") {
		t.Errorf("expected missing-in-env message to contain OLD_VAR, got %q", msgMap["envguard.sync/missing-in-env"])
	}
	if !strings.Contains(msgMap["envguard.sync/value-mismatch"], "DIFF_VAR") {
		t.Errorf("expected value-mismatch message to contain DIFF_VAR, got %q", msgMap["envguard.sync/value-mismatch"])
	}
}

func TestSARIFSync_DuplicateRuleTypes(t *testing.T) {
	result := &sync.Result{
		Diffs: []sync.Diff{
			{Type: "missing-in-example", Key: "A", EnvVal: "1"},
			{Type: "missing-in-example", Key: "B", EnvVal: "2"},
		},
	}

	var buf bytes.Buffer
	if err := SARIFSync(&buf, result, "2.0.0"); err != nil {
		t.Fatalf("SARIFSync failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	if len(log.Runs[0].Tool.Driver.Rules) != 1 {
		t.Errorf("expected 1 unique rule, got %d", len(log.Runs[0].Tool.Driver.Rules))
	}
	if len(log.Runs[0].Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(log.Runs[0].Results))
	}
}

func TestSARIFLint_EmptyLevel(t *testing.T) {
	findings := []LintFinding{
		{Level: "", Rule: "no-level", Message: "level is empty"},
	}

	var buf bytes.Buffer
	if err := SARIFLint(&buf, findings, "2.0.0"); err != nil {
		t.Fatalf("SARIFLint failed: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("failed to unmarshal SARIF: %v", err)
	}

	if log.Runs[0].Results[0].Level != "warning" {
		t.Errorf("expected default level warning, got %q", log.Runs[0].Results[0].Level)
	}
	if log.Runs[0].Tool.Driver.Rules[0].DefaultConfig.Level != "warning" {
		t.Errorf("expected default rule config level warning, got %q", log.Runs[0].Tool.Driver.Rules[0].DefaultConfig.Level)
	}
}
