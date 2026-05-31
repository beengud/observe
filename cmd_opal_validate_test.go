package main

import (
	"strings"
	"testing"
)

// validateIngestResponse builds a mock GraphQL response for validateIngestFilterExpression.
// The actual API returns only { message } on TaskResultError (no severity or symbol).
func validateIngestResponse(diagsJSON string) string {
	return `{"data":{"validateIngestFilterExpression":` + diagsJSON + `}}`
}

func TestCmdOpalValidateIngestSuccess(t *testing.T) {
	resp := validateIngestResponse(`[]`)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	flagOpalDataset = ""
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "validate-ingest", "--dataset", "42918275", "filter true"}, fix.hc)
	flagOpalDataset = "" // reset after parse
	fix.Assert()
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "OK") {
		t.Errorf("expected OK in output, got: %q", out)
	}
}

func TestCmdOpalValidateIngestErrors(t *testing.T) {
	resp := validateIngestResponse(
		`[{"message":"1,1: unknown verb \"bad_filter\""}]`,
	)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	flagOpalDataset = ""
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "validate-ingest", "--dataset", "42918275", "bad_filter"}, fix.hc)
	})
	flagOpalDataset = ""
	fix.Assert()
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "ERROR") {
		t.Errorf("expected ERROR in output, got: %q", out)
	}
	if !strings.Contains(out, "bad_filter") {
		t.Errorf("expected error message in output, got: %q", out)
	}
}

func TestCmdOpalValidateIngestNoDataset(t *testing.T) {
	fix := startFixture(t)
	flagOpalDataset = ""
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "validate-ingest", "filter true"}, fix.hc)
	})
	flagOpalDataset = ""
	fix.Assert()
}

func TestCmdOpalValidateIngestNoPipeline(t *testing.T) {
	fix := startFixture(t)
	flagOpalDataset = ""
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "validate-ingest", "--dataset", "42918275"}, fix.hc)
	})
	flagOpalDataset = ""
	fix.Assert()
}

func TestCmdOpalValidateIngestNullResponse(t *testing.T) {
	// Null response means no errors (valid pipeline)
	resp := validateIngestResponse(`null`)
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, resp},
	)
	flagOpalDataset = ""
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"opal", "validate-ingest", "--dataset", "42918275", "filter true"}, fix.hc)
	flagOpalDataset = ""
	fix.Assert()
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "OK") {
		t.Errorf("expected OK for null (no errors) response, got: %q", out)
	}
}
