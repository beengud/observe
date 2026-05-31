//go:build integration

package main

import (
	"net/http"
	"testing"
)

// TestIntegrationWorksheetList verifies that listing worksheets in workspace 42379913
// returns no error (results may be empty).
func TestIntegrationWorksheetList(t *testing.T) {
	cfg := integrationConfig()
	cfg.WorkspaceIdOrName = "42379913"
	op := NewCaptureOutput()

	infos, err := ObjectTypeWorksheet.List(cfg, op, http.DefaultClient)
	if err != nil {
		t.Fatalf("worksheet list failed: %v\nerrors: %s", err, op.ErrorBuf.String())
	}
	t.Logf("Found %d worksheets", len(infos))
	for i, info := range infos {
		if i >= 5 {
			break
		}
		t.Logf("  worksheet: id=%s name=%s", info.Id, info.Name)
	}
}

// TestIntegrationWorksheetCreateGetDelete creates a worksheet, verifies it exists
// via get, then deletes it as cleanup. The delete is always attempted.
func TestIntegrationWorksheetCreateGetDelete(t *testing.T) {
	cfg := integrationConfig()
	cfg.WorkspaceIdOrName = "42379913"
	op := NewCaptureOutput()

	// Create worksheet input.
	input := object{
		"name":        "Claude Integration Test Worksheet",
		"workspaceId": "42379913",
		"stages": []any{
			object{
				"stageID":  "s1",
				"pipeline": "limit 1",
			},
		},
	}

	// Create
	ot := &objectTypeWorksheet{}
	created, err := ot.Create(cfg, op, http.DefaultClient, input)
	if err != nil {
		t.Fatalf("worksheet create failed: %v\nerrors: %s", err, op.ErrorBuf.String())
	}
	if created == nil {
		t.Fatal("worksheet create: nil result")
	}
	info := created.GetInfo()
	t.Logf("Created worksheet: id=%s name=%s", info.Id, info.Name)

	// Always attempt cleanup.
	defer func() {
		deleteOp := NewCaptureOutput()
		if err := ot.Delete(cfg, deleteOp, http.DefaultClient, info.Id); err != nil {
			t.Logf("cleanup: worksheet delete failed (id=%s): %v", info.Id, err)
		} else {
			t.Logf("cleanup: deleted worksheet id=%s", info.Id)
		}
	}()

	if info.Id == "" {
		t.Fatal("created worksheet has empty ID")
	}

	// Get to verify it exists.
	getOp := NewCaptureOutput()
	got, err := ot.Get(cfg, getOp, http.DefaultClient, info.Id)
	if err != nil {
		t.Fatalf("worksheet get failed: %v", err)
	}
	if got == nil {
		t.Fatal("worksheet get returned nil - worksheet not found after creation")
	}
	gotInfo := got.GetInfo()
	if gotInfo.Name != "Claude Integration Test Worksheet" {
		t.Errorf("expected name %q, got %q", "Claude Integration Test Worksheet", gotInfo.Name)
	}
	t.Logf("Verified worksheet exists: id=%s name=%s", gotInfo.Id, gotInfo.Name)
}

// TestIntegrationWorksheetSearchByName searches worksheets by name and verifies
// no error is returned.
func TestIntegrationWorksheetSearchByName(t *testing.T) {
	cfg := integrationConfig()
	op := NewCaptureOutput()

	termMap := object{
		"workspaceId": "42379913",
		"name":        "Claude",
	}
	obj, err := gqlWorksheetSearch.query(cfg, op, http.DefaultClient, object{"terms": termMap})
	if err != nil {
		t.Fatalf("worksheet search failed: %v", err)
	}
	if obj == nil {
		t.Log("worksheet search returned nil (no results)")
		return
	}
	items, ok := obj.(array)
	if !ok {
		t.Fatalf("unexpected response type: %T", obj)
	}
	t.Logf("worksheetSearch for 'Claude' returned %d results", len(items))
}
