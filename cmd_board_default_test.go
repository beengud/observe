package main

import (
	"strings"
	"testing"
)

// TestCmdBoardSetDefaultSuccess tests the set-default subcommand success path.
func TestCmdBoardSetDefaultSuccess(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"setDefaultDashboard":{"success":true,"errorMessage":null}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"board", "set-default", "ds-100", "board-200"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	if !strings.Contains(fix.op.OutputBuf.String(), "Default dashboard set successfully") {
		t.Error("expected success message, got:", fix.op.OutputBuf.String())
	}
}

// TestCmdBoardSetDefaultError tests the set-default subcommand with an API error response.
func TestCmdBoardSetDefaultError(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"setDefaultDashboard":{"success":false,"errorMessage":"dataset not found"}}}`},
	)
	func() {
		defer func() { recover() }()
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"board", "set-default", "invalid-ds", "board-200"}, fix.hc)
	}()
	if !strings.Contains(fix.op.ErrorBuf.String(), "dataset not found") {
		t.Error("expected error message, got:", fix.op.ErrorBuf.String())
	}
}

// TestCmdBoardSetDefaultUsage tests that wrong arg count returns usage error.
func TestCmdBoardSetDefaultUsage(t *testing.T) {
	fix := startFixture(t)
	func() {
		defer func() { recover() }()
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"board", "set-default"}, fix.hc)
	}()
	if !strings.Contains(fix.op.ErrorBuf.String(), "usage: observe board set-default") {
		t.Error("expected usage error, got:", fix.op.ErrorBuf.String())
	}
}

// TestCmdBoardClearDefaultSuccess tests the clear-default subcommand success path.
func TestCmdBoardClearDefaultSuccess(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"clearDefaultDashboard":{"success":true,"errorMessage":null}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"board", "clear-default", "ds-100"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	if !strings.Contains(fix.op.OutputBuf.String(), "Default dashboard cleared successfully") {
		t.Error("expected success message, got:", fix.op.OutputBuf.String())
	}
}

// TestCmdBoardClearDefaultError tests the clear-default subcommand with an API error.
func TestCmdBoardClearDefaultError(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"clearDefaultDashboard":{"success":false,"errorMessage":"permission denied"}}}`},
	)
	func() {
		defer func() { recover() }()
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"board", "clear-default", "invalid-ds"}, fix.hc)
	}()
	if !strings.Contains(fix.op.ErrorBuf.String(), "permission denied") {
		t.Error("expected error message, got:", fix.op.ErrorBuf.String())
	}
}

// TestCmdBoardClearDefaultUsage tests wrong arg count for clear-default.
func TestCmdBoardClearDefaultUsage(t *testing.T) {
	fix := startFixture(t)
	func() {
		defer func() { recover() }()
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"board", "clear-default"}, fix.hc)
	}()
	if !strings.Contains(fix.op.ErrorBuf.String(), "usage: observe board clear-default") {
		t.Error("expected usage error, got:", fix.op.ErrorBuf.String())
	}
}

// TestCmdBoardSearchNoResultsWithFolder tests that dashboardSearch with folder
// and no results produces only a header line.
func TestCmdBoardSearchNoResultsWithFolder(t *testing.T) {
	flagBoardScaffold = ""
	flagBoardSearchFolder = "nonexistent-folder"
	defer func() { flagBoardSearchFolder = "" }()

	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"dashboardSearch":{"results":[]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"list", "board"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "id") {
		t.Error("expected header in output:", out)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line (header only), got %d: %q", len(lines), out)
	}
}

// TestBoardSearchTermsAllFields tests that Search with all three term fields
// passes them correctly into the query.
func TestBoardSearchTermsAllFields(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"dashboardSearch":{"results":[{"score":1,"dashboard":{"id":"d1","name":"D","workspaceId":"ws1","updatedDate":"2026-01-01"}}]}}}`},
	)
	ot := &objectTypeBoard{}
	infos, err := ot.Search(fix.cfg, fix.op, fix.hc, BoardSearchTerms{
		Name:        "D",
		WorkspaceId: "ws1",
		FolderId:    "f1",
	})
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	if len(infos) != 1 {
		t.Fatalf("expected 1 result, got %d", len(infos))
	}
	if infos[0].Id != "d1" {
		t.Errorf("expected id=d1, got %s", infos[0].Id)
	}
}
