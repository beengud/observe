package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestCmdWorksheetList tests that `observe worksheet list` returns results.
func TestCmdWorksheetList(t *testing.T) {
	flagWorksheetName = ""

	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"worksheetSearch":{"results":[{"score":1,"worksheet":{"id":"ws-100","name":"My Analysis","workspaceId":"42379913","updatedDate":"2026-01-01"}}]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"worksheet", "list"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "ws-100") {
		t.Error("expected worksheet id in output:", out)
	}
	if !strings.Contains(out, "My Analysis") {
		t.Error("expected worksheet name in output:", out)
	}
}

// TestCmdWorksheetListWithNameFilter tests filtering by name.
func TestCmdWorksheetListWithNameFilter(t *testing.T) {
	flagWorksheetName = "Analysis"
	defer func() { flagWorksheetName = "" }()

	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"worksheetSearch":{"results":[{"score":1,"worksheet":{"id":"ws-100","name":"My Analysis","workspaceId":"42379913","updatedDate":"2026-01-01"}}]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"worksheet", "list"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "ws-100") {
		t.Error("expected worksheet id in output:", out)
	}
}

// TestCmdWorksheetListEmpty tests that empty results produce just a header.
func TestCmdWorksheetListEmpty(t *testing.T) {
	flagWorksheetName = ""

	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"worksheetSearch":{"results":[]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"worksheet", "list"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	// Should only have a header line.
	if !strings.Contains(out, "id") {
		t.Error("expected header in output:", out)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line (header only), got %d: %q", len(lines), out)
	}
}

// TestCmdWorksheetGet tests that `observe worksheet get <id>` outputs JSON.
func TestCmdWorksheetGet(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"worksheet":{"id":"ws-100","name":"My Analysis","workspaceId":"42379913","updatedDate":"2026-01-01","stages":[{"stageID":"s1","pipeline":"limit 10"}]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"worksheet", "get", "ws-100"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "ws-100") {
		t.Error("expected id in output:", out)
	}
	if !strings.Contains(out, "My Analysis") {
		t.Error("expected name in output:", out)
	}
	if !strings.Contains(out, "stages") {
		t.Error("expected stages in output:", out)
	}
}

// TestCmdWorksheetCreate tests that `observe worksheet create <file>` creates a worksheet.
func TestCmdWorksheetCreate(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"saveWorksheet":{"id":"ws-new","name":"New Sheet","workspaceId":"42379913"}}}`},
	)
	inputJSON := `{"name":"New Sheet","workspaceId":"42379913","stages":[{"stageID":"s1","pipeline":"limit 10"}]}`
	fix.fs.WriteFile("worksheet.json", []byte(inputJSON), 0644)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"worksheet", "create", "worksheet.json"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	if diff := cmp.Diff(fix.op.OutputBuf.String(), "Created: New Sheet (id: ws-new)\n"); diff != "" {
		t.Error("unexpected output:", diff)
	}
}

// TestCmdWorksheetCreateRoundTrip tests create followed by get to verify the worksheet.
func TestCmdWorksheetCreateRoundTrip(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"saveWorksheet":{"id":"ws-rt","name":"RoundTrip","workspaceId":"42379913"}}}`},
		testRequest{"/v1/meta", 200, `{"data":{"worksheet":{"id":"ws-rt","name":"RoundTrip","workspaceId":"42379913","updatedDate":"2026-01-01","stages":[{"stageID":"s1","pipeline":"limit 5"}]}}}`},
	)
	inputJSON := `{"name":"RoundTrip","workspaceId":"42379913","stages":[{"stageID":"s1","pipeline":"limit 5"}]}`
	fix.fs.WriteFile("ws.json", []byte(inputJSON), 0644)

	// Create
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"worksheet", "create", "ws.json"}, fix.hc)
	if !strings.Contains(fix.op.OutputBuf.String(), "ws-rt") {
		t.Error("create: expected id in output:", fix.op.OutputBuf.String())
	}

	// Get
	fix.op.OutputBuf.Reset()
	fix.op.ErrorBuf.Reset()
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"worksheet", "get", "ws-rt"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("get: unexpected error:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "RoundTrip") {
		t.Error("get: expected name in output:", out)
	}
	fix.Assert()
}

// TestCmdWorksheetDelete tests that `observe worksheet delete <id>` deletes a worksheet.
func TestCmdWorksheetDelete(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"deleteWorksheet":{"success":true,"errorMessage":null}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"worksheet", "delete", "ws-100"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	if diff := cmp.Diff(fix.op.OutputBuf.String(), "Deleted worksheet ws-100\n"); diff != "" {
		t.Error("unexpected output:", diff)
	}
}

// TestCmdWorksheetDeleteError tests that error response is surfaced.
func TestCmdWorksheetDeleteError(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"deleteWorksheet":{"success":false,"errorMessage":"worksheet not found"}}}`},
	)
	// RunCommandWithConfig calls op.Exit(1) on error, which panics in CaptureOutput.
	// Recover the panic so we can check the error output.
	func() {
		defer func() { recover() }()
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"worksheet", "delete", "ws-missing"}, fix.hc)
	}()
	if !strings.Contains(fix.op.ErrorBuf.String(), "worksheet not found") {
		t.Error("expected error message in error output:", fix.op.ErrorBuf.String())
	}
}

// TestCmdWorksheetUnknownSubcommand tests error on unknown subcommand.
func TestCmdWorksheetUnknownSubcommand(t *testing.T) {
	fix := startFixture(t)
	func() {
		defer func() { recover() }()
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"worksheet", "frobnicate"}, fix.hc)
	}()
	if !strings.Contains(fix.op.ErrorBuf.String(), "unknown worksheet subcommand") {
		t.Error("expected unknown subcommand error:", fix.op.ErrorBuf.String())
	}
}
