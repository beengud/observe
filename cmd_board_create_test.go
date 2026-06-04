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

// TestReadBoardInputStageInputNormalization verifies that each stage input gets
// stageId: "" when it is not already present. StageInput.stageId is String!
// (non-null) in the GraphQL schema; omitting it causes the API to store null,
// which any subsequent query with stageId: String! will reject with "requested
// element is null which schema does not allow".
func TestReadBoardInputStageInputNormalization(t *testing.T) {
	fa := FuncArgs{fs: NewFakeFs()}
	boardJSON := []byte(`{
		"name": "Test",
		"workspaceId": "42379913",
		"layout": {},
		"stages": [
			{
				"stageID": "stage-abc",
				"pipeline": "limit 10",
				"input": [
					{"inputName": "main", "datasetId": "42450595"},
					{"inputName": "other", "datasetId": "42450596", "stageId": "stage-xyz"}
				]
			}
		]
	}`)
	fa.fs.WriteFile("b.json", boardJSON, 0664)
	input, err := readBoardInput(fa, "b.json")
	if err != nil {
		t.Fatal(err)
	}
	stages := input["stages"].([]any)
	inputs := stages[0].(map[string]any)["input"].([]any)
	// First entry had no stageId or inputRole — should be normalized
	entry0 := inputs[0].(map[string]any)
	if entry0["stageId"] != "" {
		t.Errorf("expected stageId=\"\" for dataset input, got %v", entry0["stageId"])
	}
	if entry0["inputRole"] != "Data" {
		t.Errorf("expected inputRole=\"Data\" for dataset input, got %v", entry0["inputRole"])
	}
	// Second entry already had stageId — should be preserved; inputRole normalized
	entry1 := inputs[1].(map[string]any)
	if entry1["stageId"] != "stage-xyz" {
		t.Errorf("expected stageId=\"stage-xyz\" preserved, got %v", entry1["stageId"])
	}
	if entry1["inputRole"] != "Data" {
		t.Errorf("expected inputRole=\"Data\" normalized, got %v", entry1["inputRole"])
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
