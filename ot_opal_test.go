package main

import (
	"strings"
	"testing"
)

// verbsAndFunctionsResponse builds a mock GraphQL response for verbsAndFunctions.
func verbsAndFunctionsResponse(verbsJSON, functionsJSON string) string {
	return `{"data":{"verbsAndFunctions":{"verbs":` + verbsJSON + `,"functions":` + functionsJSON + `}}}`
}

func TestCmdOpalVerbs(t *testing.T) {
	resp := verbsAndFunctionsResponse(
		`[{"name":"filter","description":"Filter rows","category":"row"},{"name":"aggregate","description":"Aggregate rows","category":"aggregation"},{"name":"limit","description":"Limit rows","category":"row"}]`,
		`[]`,
	)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "verbs"}, fix.hc)
	fix.Assert()
	out := fix.op.OutputBuf.String()
	// Verify sorted alphabetically
	aggIdx := strings.Index(out, "aggregate")
	filterIdx := strings.Index(out, "filter")
	limitIdx := strings.Index(out, "limit")
	if aggIdx < 0 || filterIdx < 0 || limitIdx < 0 {
		t.Errorf("expected all verbs in output, got: %q", out)
	}
	if !(aggIdx < filterIdx && filterIdx < limitIdx) {
		t.Errorf("expected alphabetical order (aggregate < filter < limit), got: %q", out)
	}
	// Verify tab-separated columns
	if !strings.Contains(out, "\t") {
		t.Errorf("expected tab-separated output, got: %q", out)
	}
}

func TestCmdOpalFunctions(t *testing.T) {
	resp := verbsAndFunctionsResponse(
		`[]`,
		`[{"name":"sum","description":"Sum values","category":"aggregation","returnType":"float64"},{"name":"count","description":"Count rows","category":"aggregation","returnType":"int64"},{"name":"avg","description":"Average values","category":"aggregation","returnType":"float64"}]`,
	)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "functions"}, fix.hc)
	fix.Assert()
	out := fix.op.OutputBuf.String()
	// Verify sorted alphabetically
	avgIdx := strings.Index(out, "avg")
	countIdx := strings.Index(out, "count")
	sumIdx := strings.Index(out, "sum")
	if avgIdx < 0 || countIdx < 0 || sumIdx < 0 {
		t.Errorf("expected all functions in output, got: %q", out)
	}
	if !(avgIdx < countIdx && countIdx < sumIdx) {
		t.Errorf("expected alphabetical order (avg < count < sum), got: %q", out)
	}
	// Verify tab-separated columns
	if !strings.Contains(out, "\t") {
		t.Errorf("expected tab-separated output, got: %q", out)
	}
	// Verify returnType is included
	if !strings.Contains(out, "float64") {
		t.Errorf("expected returnType in functions output, got: %q", out)
	}
}

func TestCmdOpalVerbsEmpty(t *testing.T) {
	resp := verbsAndFunctionsResponse(`[]`, `[]`)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "verbs"}, fix.hc)
	fix.Assert()
	out := fix.op.OutputBuf.String()
	if out != "" {
		t.Errorf("expected empty output for empty verbs, got: %q", out)
	}
}

func TestCmdOpalFunctionsEmpty(t *testing.T) {
	resp := verbsAndFunctionsResponse(`[]`, `[]`)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "functions"}, fix.hc)
	fix.Assert()
	out := fix.op.OutputBuf.String()
	if out != "" {
		t.Errorf("expected empty output for empty functions, got: %q", out)
	}
}
