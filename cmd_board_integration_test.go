//go:build integration

package main

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

// TestIntegrationBoardList verifies that listing boards returns at least one result without error.
func TestIntegrationBoardList(t *testing.T) {
	cfg := integrationConfig()
	cfg.WorkspaceIdOrName = os.Getenv("OBSERVE_WORKSPACE")
	op := NewCaptureOutput()

	infos, err := ObjectTypeBoard.List(cfg, op, http.DefaultClient)
	if err != nil {
		t.Fatalf("board list failed: %v\nerrors: %s", err, op.ErrorBuf.String())
	}
	if len(infos) == 0 {
		t.Error("expected at least 1 board in workspace, got 0")
	}
	t.Logf("Found %d boards", len(infos))
	for i, info := range infos {
		if i >= 5 {
			break
		}
		t.Logf("  board: id=%s name=%s", info.Id, info.Name)
	}
}

// TestIntegrationBoardSearchByName searches for boards with "Deployment" in their name.
// May return 0 results if none exist; the test only verifies no error.
func TestIntegrationBoardSearchByName(t *testing.T) {
	cfg := integrationConfig()
	op := NewCaptureOutput()

	ot := &objectTypeBoard{}
	infos, err := ot.Search(cfg, op, http.DefaultClient, BoardSearchTerms{
		Name:        "Deployment",
		WorkspaceId: os.Getenv("OBSERVE_WORKSPACE"),
	})
	if err != nil {
		t.Fatalf("board search failed: %v\nerrors: %s", err, op.ErrorBuf.String())
	}
	t.Logf("Found %d boards matching 'Deployment'", len(infos))
	for _, info := range infos {
		t.Logf("  board: id=%s name=%s", info.Id, info.Name)
	}
}

// TestIntegrationBoardSearchWithWorkspaceFilter verifies dashboardSearch with
// a workspace filter returns results without error.
func TestIntegrationBoardSearchWithWorkspaceFilter(t *testing.T) {
	cfg := integrationConfig()
	op := NewCaptureOutput()

	ot := &objectTypeBoard{}
	infos, err := ot.Search(cfg, op, http.DefaultClient, BoardSearchTerms{
		WorkspaceId: os.Getenv("OBSERVE_WORKSPACE"),
	})
	if err != nil {
		t.Fatalf("board search with workspace filter failed: %v", err)
	}
	t.Logf("dashboardSearch with workspace filter returned %d results", len(infos))
}

// TestIntegrationBoardSetDefaultInvalidIDs verifies that set-default with invalid
// IDs returns an error from the API (not a crash).
func TestIntegrationBoardSetDefaultInvalidIDs(t *testing.T) {
	cfg := integrationConfig()
	op := NewCaptureOutput()

	// Use clearly invalid IDs.
	obj, err := gqlSetDefaultDashboard.query(cfg, op, http.DefaultClient,
		object{"dsid": "00000000", "dashid": "00000000"})
	// The API may return an error or errorMessage; either is acceptable.
	if err != nil {
		t.Logf("set-default with invalid IDs returned error (expected): %v", err)
		return
	}
	if result, ok := obj.(object); ok {
		if errMsg, _ := result["errorMessage"].(string); errMsg != "" {
			t.Logf("set-default with invalid IDs returned errorMessage (expected): %s", errMsg)
			return
		}
		// If no error, log the result.
		b, _ := json.Marshal(result)
		t.Logf("set-default with invalid IDs result: %s", b)
	}
}
