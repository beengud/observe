package main

import (
	"net/http"
	"strings"
	"testing"
)

// validateIngestResponseHelper builds a mock GraphQL response for validateIngestFilterExpression.
func validateIngestResponseHelper(diagsJSON string) string {
	return `{"data":{"validateIngestFilterExpression":` + diagsJSON + `}}`
}

// TestCmdOpalCheckEmptyPipeline verifies that an empty pipeline string is sent to
// the API rather than triggering a local usage error.
func TestCmdOpalCheckEmptyPipeline(t *testing.T) {
	resp := checkQueriesResponse("[]", "[]", `[]`)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "check", ""}, fix.hc)
	fix.Assert()
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "OK") {
		t.Errorf("expected OK for empty pipeline (accepted by API), got: %q", out)
	}
}

// TestCmdOpalCheckMultipleErrors verifies that all errors are printed, not just the first.
func TestCmdOpalCheckMultipleErrors(t *testing.T) {
	resp := checkQueriesResponse(
		`[{"message":"first error","severity":"error","symbol":{"offset":0,"line":1,"column":1,"length":3}},{"message":"second error","severity":"error","symbol":{"offset":10,"line":1,"column":10,"length":5}}]`,
		"[]",
		"[]",
	)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "check", "bad stuff"}, fix.hc)
	})
	fix.Assert()
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "first error") {
		t.Errorf("expected first error in output, got: %q", out)
	}
	if !strings.Contains(out, "second error") {
		t.Errorf("expected second error in output, got: %q", out)
	}
}

// TestCmdOpalCheckNetworkError verifies that a network-level failure is surfaced as an error.
func TestCmdOpalCheckNetworkError(t *testing.T) {
	// Use a server that immediately returns 500.
	fix := startFixture(t,
		testRequest{"/v1/meta", 500, `{"errors":[{"message":"internal server error"}]}`},
	)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "check", "filter true"}, fix.hc)
	})
	fix.Assert()
}

// errorHttpClient always returns an error (simulates connection failure).
type errorHttpClient struct{}

func (e *errorHttpClient) Do(req *http.Request) (*http.Response, error) {
	return nil, &httpDialError{"connection refused"}
}

type httpDialError struct{ msg string }

func (e *httpDialError) Error() string { return e.msg }

// TestCmdOpalCheckConnectionError verifies behavior when HTTP connection fails.
func TestCmdOpalCheckConnectionError(t *testing.T) {
	fix := startFixture(t) // no requests expected
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "check", "filter true"}, &errorHttpClient{})
	})
	// The error should surface, not panic internally
}

// TestCmdOpalVerbsNetworkError verifies that a network failure from opal verbs is surfaced.
func TestCmdOpalVerbsNetworkError(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 500, `{"errors":[{"message":"internal server error"}]}`},
	)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "verbs"}, fix.hc)
	})
	fix.Assert()
}

// TestCmdOpalFunctionsNetworkError verifies that a network failure from opal functions is surfaced.
func TestCmdOpalFunctionsNetworkError(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 500, `{"errors":[{"message":"internal server error"}]}`},
	)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "functions"}, fix.hc)
	})
	fix.Assert()
}

// TestCmdOpalValidateIngestSuccess verifies successful ingest filter validation.
func TestCmdOpalValidateIngestSuccessEdge(t *testing.T) {
	resp := validateIngestResponseHelper(`[]`)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	flagOpalDataset = ""
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "validate-ingest", "--dataset", "42918275", "filter true"}, fix.hc)
	flagOpalDataset = ""
	fix.Assert()
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "OK") {
		t.Errorf("expected OK in output, got: %q", out)
	}
}

// TestCmdOpalValidateIngestWarnings verifies that severity other than "error" is treated as WARN.
func TestCmdOpalValidateIngestWarnings(t *testing.T) {
	resp := validateIngestResponseHelper(
		`[{"message":"deprecated function","severity":"warning","symbol":{"offset":0,"line":1,"column":1,"length":5}}]`,
	)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	flagOpalDataset = ""
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "validate-ingest", "--dataset", "42918275", "filter true"}, fix.hc)
	flagOpalDataset = ""
	fix.Assert()
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "WARN") {
		t.Errorf("expected WARN for severity=warning, got: %q", out)
	}
	if !strings.Contains(out, "OK") {
		t.Errorf("expected OK after warnings (no errors), got: %q", out)
	}
}
