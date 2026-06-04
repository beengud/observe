package main

import (
	"strings"
	"testing"
)

var minimalBoardJSON = []byte(`{"name":"My Board","workspaceId":"42379913","layout":{"autoPack":true,"gridLayout":{"sections":[]},"stageListLayout":{"isModified":false,"parameters":[]}},"stages":[]}`)

// TestCmdBoardCreate verifies that board create prints the name, id, visibility, and a view URL.
func TestCmdBoardCreate(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"saveDashboard":{"id":"42000001","name":"My Board","workspaceId":"42379913","folderId":"42379919","visibility":"Listed"}}}`},
	)
	fix.fs.WriteFile("board.json", minimalBoardJSON, 0664)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"board", "create", "board.json"}, fix.hc)
	if fix.op.ErrorBuf.String() != "" {
		t.Error("unexpected error output:", fix.op.ErrorBuf.String())
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "Created: My Board (id: 42000001)") {
		t.Error("expected created message, got:", out)
	}
	if !strings.Contains(out, "Visibility: Listed") {
		t.Error("expected visibility line, got:", out)
	}
	if !strings.Contains(out, "View:") {
		t.Error("expected view URL line, got:", out)
	}
	if !strings.Contains(out, "/workspace/42379913/dashboard/42000001") {
		t.Error("expected workspace/board path in URL, got:", out)
	}
}

// TestCmdBoardUpdate verifies that board update prints the name, id, visibility, and a view URL.
func TestCmdBoardUpdate(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"saveDashboard":{"id":"42000001","name":"My Board","workspaceId":"42379913","folderId":"42379919","visibility":"Listed"}}}`},
	)
	fix.fs.WriteFile("board.json", minimalBoardJSON, 0664)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"board", "update", "42000001", "board.json"}, fix.hc)
	if fix.op.ErrorBuf.String() != "" {
		t.Error("unexpected error output:", fix.op.ErrorBuf.String())
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "Updated: My Board (id: 42000001)") {
		t.Error("expected updated message, got:", out)
	}
	if !strings.Contains(out, "Visibility: Listed") {
		t.Error("expected visibility line, got:", out)
	}
	if !strings.Contains(out, "View:") {
		t.Error("expected view URL line, got:", out)
	}
	if !strings.Contains(out, "/workspace/42379913/dashboard/42000001") {
		t.Error("expected workspace/board path in URL, got:", out)
	}
}

// TestBoardViewURL verifies URL construction for a production-style config.
func TestBoardViewURL(t *testing.T) {
	cfg := &Config{
		CustomerIdStr: "109601619518",
		SiteStr:       "observeinc.com",
	}
	got := boardViewURL(cfg, "42379913", "43102612")
	want := "https://109601619518.observeinc.com/workspace/42379913/dashboard/43102612"
	if got != want {
		t.Errorf("boardViewURL: got %q, want %q", got, want)
	}
}
