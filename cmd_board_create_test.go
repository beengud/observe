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

// TestReadBoardInputStageIDNormalization verifies that "stageID" is renamed to
// "id" so the GraphQL StageQueryInput receives the correct field name. Without
// this normalization the API silently ignores the stage label and generates
// random IDs, breaking all card.stageId layout references.
func TestReadBoardInputStageIDNormalization(t *testing.T) {
	fa := FuncArgs{fs: NewFakeFs()}
	boardJSON := []byte(`{
		"name": "Test",
		"workspaceId": "42379913",
		"layout": {},
		"stages": [
			{"stageID": "stage-abc", "pipeline": "limit 10", "input": [{"inputName":"main","datasetId":"1"}]},
			{"stageID": "stage-xyz", "pipeline": "limit 5", "input": []}
		]
	}`)
	fa.fs.WriteFile("b.json", boardJSON, 0664)
	input, err := readBoardInput(fa, "b.json")
	if err != nil {
		t.Fatal(err)
	}
	stages := input["stages"].([]any)
	for _, s := range stages {
		stage := s.(map[string]any)
		if _, hasStageID := stage["stageID"]; hasStageID {
			t.Error("stageID should have been removed from stage")
		}
		if _, hasID := stage["id"]; !hasID {
			t.Error("id should have been added to stage")
		}
	}
	if stages[0].(map[string]any)["id"] != "stage-abc" {
		t.Errorf("expected id=stage-abc, got %v", stages[0].(map[string]any)["id"])
	}
	if stages[1].(map[string]any)["id"] != "stage-xyz" {
		t.Errorf("expected id=stage-xyz, got %v", stages[1].(map[string]any)["id"])
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
