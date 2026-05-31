package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestCmdListBoardNoFlags tests that `observe list board` (no search flags) uses the
// workspace-scoped plain list query.
func TestCmdListBoardNoFlags(t *testing.T) {
	// Reset search flags to make sure no flags are set.
	flagBoardScaffold = ""
	flagBoardSearchFolder = ""

	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"dashboardSearch":{"dashboards":[{"score":1,"dashboard":{"id":"100","name":"Ops Board","workspaceId":"42379913","updatedDate":"2026-01-01"}}]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"list", "board"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	if diff := cmp.Diff(fix.op.OutputBuf.String(), "id  name     \n100 Ops Board\n"); diff != "" {
		t.Error("unexpected data output:", diff)
	}
}

// TestCmdListBoardWithNameFlag tests that the --name flag triggers dashboardSearch.
func TestCmdListBoardWithNameFlag(t *testing.T) {
	flagBoardScaffold = "Ops"
	flagBoardSearchFolder = ""
	defer func() {
		flagBoardScaffold = ""
	}()

	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"dashboardSearch":{"results":[{"score":1,"dashboard":{"id":"100","name":"Ops Board","workspaceId":"42379913","updatedDate":"2026-01-01"}}]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"list", "board"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	if diff := cmp.Diff(fix.op.OutputBuf.String(), "id  name     \n100 Ops Board\n"); diff != "" {
		t.Error("unexpected data output:", diff)
	}
}

// TestCmdListBoardWithWorkspaceViaConfig tests that the global --workspace flag
// (which populates cfg.WorkspaceIdOrName) is used as the workspace filter in
// dashboardSearch when search is triggered by --name or --folder.
func TestCmdListBoardWithWorkspaceViaConfig(t *testing.T) {
	flagBoardScaffold = "Fleet"
	flagBoardSearchFolder = ""
	defer func() {
		flagBoardScaffold = ""
	}()

	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"dashboardSearch":{"results":[{"score":1,"dashboard":{"id":"200","name":"Fleet Board","workspaceId":"99999","updatedDate":"2026-01-02"}}]}}}`},
	)
	// Simulate --workspace 99999 global flag setting cfg.WorkspaceIdOrName.
	fix.cfg.WorkspaceIdOrName = "99999"
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"list", "board"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	if diff := cmp.Diff(fix.op.OutputBuf.String(), "id  name       \n200 Fleet Board\n"); diff != "" {
		t.Error("unexpected data output:", diff)
	}
}

// TestCmdListBoardWithFolderFlag tests that --folder triggers dashboardSearch.
func TestCmdListBoardWithFolderFlag(t *testing.T) {
	flagBoardScaffold = ""
	flagBoardSearchFolder = "folder-1"
	defer func() {
		flagBoardSearchFolder = ""
	}()

	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"dashboardSearch":{"results":[{"score":0.8,"dashboard":{"id":"300","name":"Folder Board","workspaceId":"42379913","updatedDate":"2026-01-03"}}]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"list", "board"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	if diff := cmp.Diff(fix.op.OutputBuf.String(), "id  name        \n300 Folder Board\n"); diff != "" {
		t.Error("unexpected data output:", diff)
	}
}

// TestCmdListBoardSearchNoResults tests that dashboardSearch returning empty results
// outputs just the header row (no data rows).
func TestCmdListBoardSearchNoResults(t *testing.T) {
	flagBoardScaffold = "NonExistent"
	flagBoardSearchFolder = ""
	defer func() {
		flagBoardScaffold = ""
	}()

	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"dashboardSearch":{"results":[]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"list", "board"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	// Only the header should be printed with no data rows.
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "id") || !strings.Contains(out, "name") {
		t.Error("expected header row in output:", out)
	}
	// Should not have any data lines beyond the header.
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line (header only), got %d: %q", len(lines), out)
	}
}

// TestBoardSearchUnpackItems tests the unpackBoardItems helper directly.
func TestBoardSearchUnpackItems(t *testing.T) {
	items := array{
		object{
			"score": float64(1),
			"dashboard": object{
				"id":          "abc",
				"name":        "Test Board",
				"workspaceId": "42379913",
				"updatedDate": "2026-01-01",
			},
		},
		// Item without "dashboard" key should be skipped.
		object{"score": float64(0)},
		// Non-object item should be skipped.
		"not-an-object",
	}

	infos := unpackBoardItems(items)
	if len(infos) != 1 {
		t.Fatalf("expected 1 result, got %d", len(infos))
	}
	if infos[0].Id != "abc" {
		t.Errorf("expected id=abc, got %s", infos[0].Id)
	}
	if infos[0].Name != "Test Board" {
		t.Errorf("expected name=Test Board, got %s", infos[0].Name)
	}
}
