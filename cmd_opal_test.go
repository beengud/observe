package main

import (
	"strings"
	"testing"
)

// checkQueriesResponse builds a mock GraphQL response for the checkQueries operation.
func checkQueriesResponse(errorsJSON, warningsJSON, fieldsJSON string) string {
	return `{"data":{"checkQueries":{"parsedPipeline":{"errors":` + errorsJSON + `,"warnings":` + warningsJSON + `},"resultSchema":{"fields":` + fieldsJSON + `}}}}`
}

func TestCmdOpalCheckSuccess(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, checkQueriesResponse("[]", "[]", `[{"name":"timestamp","type":"datetime"},{"name":"log","type":"string"}]`)},
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
		`[{"message":"unknown verb","severity":"error","symbol":{"offset":0,"line":1,"column":1,"length":3}}]`,
		"[]",
		"[]",
	)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "check", "bad_verb 123"}, fix.hc)
	})
	fix.Assert()
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "ERROR") {
		t.Errorf("expected ERROR in output, got: %q", out)
	}
	if !strings.Contains(out, "unknown verb") {
		t.Errorf("expected error message in output, got: %q", out)
	}
}

func TestCmdOpalCheckWarnings(t *testing.T) {
	resp := checkQueriesResponse(
		"[]",
		`[{"kind":"deprecation","message":"this is deprecated","symbol":{"offset":0,"line":1,"column":1,"length":5}}]`,
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
	resp := checkQueriesResponse("[]", "[]", `[]`)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	// Write pipeline to fake FS — note: cmdOpalCheck uses os.ReadFile for --file,
	// so we use a real temp file here via the standard library.
	// The fake FS is not used for --file reads in the current implementation.
	// This test just verifies the flag path fails gracefully with a bad file.
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "check", "--file", "/nonexistent/path/opal.txt"}, fix.hc)
	})
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
