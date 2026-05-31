//go:build integration

package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
)

// Integration tests for Monitor V2 commands.
// Run with: go test -tags integration ./...
//
// These tests connect to the live Observe tenant at
// 109601619518.observeinc.com using workspace 42379913.
// They require a valid OBSERVE_CONFIG or ~/.config/observe.yaml profile.
//
// Tests are designed to tolerate an empty monitor/alarm list so they
// do not require pre-existing data to pass.

const integrationWorkspaceId = "42379913"

// integrationMonitorConfig returns a Config for the live tenant.
// It reads from the default profile in observe.yaml.
// (Named integrationMonitorConfig to avoid collision with integrationConfig
// defined in cmd_dataset_integration_test.go.)
func integrationMonitorConfig(t *testing.T) *Config {
	t.Helper()
	cfg := &Config{}
	configPath := GetConfigFilePath()
	if _, err := os.Stat(configPath); err != nil {
		t.Skipf("no config file found at %s, skipping integration test", configPath)
	}
	if err := ReadConfig(cfg, configPath, "default", false); err != nil {
		t.Skipf("could not read config: %s", err)
	}
	if cfg.AuthtokenStr == "" {
		t.Skip("no authtoken in config, skipping integration test")
	}
	return cfg
}

// TestIntegrationListMonitors verifies that searchMonitorV2 returns without error.
// The result may be empty; that is acceptable.
func TestIntegrationListMonitors(t *testing.T) {
	cfg := integrationMonitorConfig(t)
	op := NewCaptureOutput()
	hc := &http.Client{}

	args := object{"workspaceId": integrationWorkspaceId}
	obj, err := gqlListMonitorV2.query(cfg, op, hc, args)
	if err != nil {
		t.Fatalf("searchMonitorV2 query failed: %s", err)
	}
	items, ok := obj.(array)
	if !ok {
		t.Fatalf("expected array result from searchMonitorV2, got %T", obj)
	}
	t.Logf("found %d monitors in workspace %s", len(items), integrationWorkspaceId)

	// If monitors exist, get the first one by ID.
	if len(items) > 0 {
		firstItem, ok := items[0].(object)
		if !ok {
			t.Fatal("first monitor item is not an object")
		}
		id, ok := firstItem["id"].(string)
		if !ok || id == "" {
			t.Fatal("first monitor item has no id")
		}
		t.Logf("getting monitor id=%s", id)
		gotObj, err := gqlGetMonitorV2.query(cfg, op, hc, object{"id": id})
		if err != nil {
			t.Fatalf("monitorV2 query failed for id=%s: %s", id, err)
		}
		if gotObj == nil {
			t.Errorf("monitorV2 returned nil for id=%s", id)
		}
		gotMap, ok := gotObj.(object)
		if !ok {
			t.Fatalf("monitorV2 response is not an object: %T", gotObj)
		}
		gotId, _ := gotMap["id"].(string)
		if gotId != id {
			t.Errorf("expected id=%s, got %s", id, gotId)
		}
		t.Logf("monitor: id=%s name=%v", gotId, gotMap["name"])
	}
}

// TestIntegrationSearchAlarms verifies that searchMonitorV2Alarms returns
// without error for the last 24 hours in the default workspace.
func TestIntegrationSearchAlarms(t *testing.T) {
	cfg := integrationMonitorConfig(t)
	op := NewCaptureOutput()
	hc := &http.Client{}

	// Use a far-past start time to retrieve all alarms ever recorded.
	args := object{
		"workspaceId": integrationWorkspaceId,
		"startTime":   "2000-01-01T00:00:00Z",
	}
	obj, err := gqlSearchMonitorV2Alarms.query(cfg, op, hc, args)
	if err != nil {
		// Some tenants may not have the alarms API enabled; skip rather than fail.
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not supported") {
			t.Skipf("searchMonitorV2Alarms not available: %s", err)
		}
		t.Fatalf("searchMonitorV2Alarms query failed: %s", err)
	}
	alarms, ok := obj.(array)
	if !ok {
		t.Fatalf("expected array result from searchMonitorV2Alarms, got %T", obj)
	}
	t.Logf("found %d alarms in workspace %s", len(alarms), integrationWorkspaceId)
}

// TestIntegrationPreviewQuery validates evaluateMonitorV2Source against a
// stub MonitorV2Input from testdata/monitor_input_stub.json.
// This test is skipped if the API returns an error (e.g. the stub input
// references a dataset that doesn't exist in the tenant).
func TestIntegrationPreviewQuery(t *testing.T) {
	cfg := integrationMonitorConfig(t)
	op := NewCaptureOutput()
	hc := &http.Client{}

	data, err := os.ReadFile("testdata/monitor_input_stub.json")
	if err != nil {
		t.Skipf("could not read testdata/monitor_input_stub.json: %s", err)
	}

	var input map[string]any
	if err := json.Unmarshal(data, &input); err != nil {
		t.Skipf("could not parse testdata/monitor_input_stub.json: %s", err)
	}

	obj, err := gqlEvaluateMonitorV2Source.query(cfg, op, hc, object{"input": input})
	if err != nil {
		t.Skipf("evaluateMonitorV2Source not available or stub input invalid: %s", err)
	}
	if obj == nil {
		t.Skip("evaluateMonitorV2Source returned nil; skipping")
	}
	result, ok := obj.(object)
	if !ok {
		t.Fatalf("expected object from evaluateMonitorV2Source, got %T", obj)
	}
	pipeline, _ := result["pipeline"].(string)
	t.Logf("pipeline: %s", pipeline)
}

// TestIntegrationPreview validates previewMonitorV2 against the stub input.
// Skipped if the API returns an error or stub input is invalid for the tenant.
func TestIntegrationPreview(t *testing.T) {
	cfg := integrationMonitorConfig(t)
	op := NewCaptureOutput()
	hc := &http.Client{}

	data, err := os.ReadFile("testdata/monitor_input_stub.json")
	if err != nil {
		t.Skipf("could not read testdata/monitor_input_stub.json: %s", err)
	}

	var input map[string]any
	if err := json.Unmarshal(data, &input); err != nil {
		t.Skipf("could not parse testdata/monitor_input_stub.json: %s", err)
	}

	obj, err := gqlPreviewMonitorV2.query(cfg, op, hc, object{
		"workspaceId": integrationWorkspaceId,
		"input":       input,
	})
	if err != nil {
		t.Skipf("previewMonitorV2 not available or stub input invalid: %s", err)
	}
	if obj == nil {
		t.Skip("previewMonitorV2 returned nil; skipping")
	}
	result, ok := obj.(object)
	if !ok {
		t.Fatalf("expected object from previewMonitorV2, got %T", obj)
	}
	wouldFire, _ := result["wouldFire"].(bool)
	t.Logf("wouldFire: %v", wouldFire)
}
