package main

import (
	"strings"
	"testing"
)

// checkQueriesResponse builds a mock GraphQL response for the checkQueries operation.
// The API returns an array at checkQueries level; errors use {col, row, text} not {message, severity, symbol}.
func checkQueriesResponse(errorsJSON, warningsJSON, fieldListJSON string) string {
	return `{"data":{"checkQueries":[{"parsedPipeline":{"errors":` + errorsJSON + `,"warnings":` + warningsJSON + `},"resultSchema":{"fieldList":` + fieldListJSON + `}}]}}`
}

func TestCmdOpalCheckSuccess(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, checkQueriesResponse("[]", "[]", `[{"name":"timestamp"},{"name":"log"}]`)},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "check", "filter true"}, fix.hc)
	fix.Assert()
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "OK") {
		t.Errorf("expected OK in output, got: %q", out)
	}
	if !strings.Contains(out, "timestamp") {
		t.Errorf("expected schema field 'timestamp' in output, got: %q", out)
	}
}

func TestCmdOpalCheckErrors(t *testing.T) {
	resp := checkQueriesResponse(
		`[{"col":"1","row":"1","text":"not_a_verb"}]`,
		"[]",
		"null",
	)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "check", "not_a_verb 123"}, fix.hc)
	})
	fix.Assert()
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "ERROR") {
		t.Errorf("expected ERROR in output, got: %q", out)
	}
	if !strings.Contains(out, "not_a_verb") {
		t.Errorf("expected error text in output, got: %q", out)
	}
}

func TestCmdOpalCheckWarnings(t *testing.T) {
	resp := checkQueriesResponse(
		"[]",
		`[{"kind":"deprecation","symbol":{"col":"1","row":"1"}}]`,
		"[]",
	)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "check", "filter true"}, fix.hc)
	fix.Assert()
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "WARN") {
		t.Errorf("expected WARN in output, got: %q", out)
	}
	if !strings.Contains(out, "deprecation") {
		t.Errorf("expected warning kind in output, got: %q", out)
	}
	if !strings.Contains(out, "OK") {
		t.Errorf("expected OK in output after warnings, got: %q", out)
	}
}

func TestCmdOpalCheckNoArgs(t *testing.T) {
	fix := startFixture(t)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "check"}, fix.hc)
	})
	fix.Assert()
}

func TestCmdOpalCheckUnknownSubcommand(t *testing.T) {
	fix := startFixture(t)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "bogus"}, fix.hc)
	})
	fix.Assert()
}

func TestCmdOpalCheckFile(t *testing.T) {
	// This test just verifies the flag path fails gracefully with a bad file (before HTTP call).
	fix := startFixture(t)
	flagOpalFile = ""
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "check", "--file", "/nonexistent/path/opal.txt"}, fix.hc)
	})
	flagOpalFile = "" // reset after parse so subsequent tests are not affected
	// No HTTP requests should have been made (error before network call)
	if fix.rix != 0 {
		t.Errorf("expected 0 HTTP requests, got %d", fix.rix)
	}
}

func TestCmdOpalNoSubcommand(t *testing.T) {
	fix := startFixture(t)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal"}, fix.hc)
	})
	fix.Assert()
}

// TestCmdOpalCheckEmptyTextError verifies that errors with empty text (from compilation
// without an input dataset) are treated as OK rather than as real errors.
func TestCmdOpalCheckEmptyTextError(t *testing.T) {
	flagOpalFile = "" // ensure --file flag is not set from a previous test
	// This simulates what the live API returns for "filter true" without an input dataset.
	resp := checkQueriesResponse(
		`[{"col":"0","row":"0","text":""}]`,
		"[]",
		"null",
	)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "check", "filter true"}, fix.hc)
	fix.Assert()
	out := fix.op.OutputBuf.String()
	// Empty-text errors are skipped; should still print OK
	if !strings.Contains(out, "OK") {
		t.Errorf("expected OK for empty-text error (compilation without input), got: %q", out)
	}
}
