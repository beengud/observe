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
		`[{"col":"1","row":"1","text":"bad_verb"},{"col":"10","row":"1","text":"bad_arg"}]`,
		"[]",
		"null",
	)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "check", "bad stuff"}, fix.hc)
	})
	fix.Assert()
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "bad_verb") {
		t.Errorf("expected first error text in output, got: %q", out)
	}
	if !strings.Contains(out, "bad_arg") {
		t.Errorf("expected second error text in output, got: %q", out)
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

// TestCmdOpalValidateIngestSuccessEdge verifies successful ingest filter validation.
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

// TestCmdOpalValidateIngestWarnings verifies that messages from validate-ingest are treated as errors.
// (The API only returns errors, not warnings, for ingest filter validation.)
func TestCmdOpalValidateIngestWarnings(t *testing.T) {
	resp := validateIngestResponseHelper(
		`[{"message":"some validation message"}]`,
	)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	flagOpalDataset = ""
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "validate-ingest", "--dataset", "42918275", "filter true"}, fix.hc)
	})
	flagOpalDataset = ""
	fix.Assert()
	out := fix.op.OutputBuf.String()
	// All messages from validateIngestFilterExpression are treated as errors
	if !strings.Contains(out, "ERROR") {
		t.Errorf("expected ERROR for any validateIngest message, got: %q", out)
	}
}
