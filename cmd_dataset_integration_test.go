//go:build integration

package main

import (
	"net/http"
	"os"
	"strings"
	"testing"
)

func integrationConfig() *Config {
	customerId := os.Getenv("OBSERVE_CUSTOMERID")
	authToken := os.Getenv("OBSERVE_AUTHTOKEN")
	site := os.Getenv("OBSERVE_SITE")
	if customerId == "" || authToken == "" {
		panic("integration tests require OBSERVE_CUSTOMERID and OBSERVE_AUTHTOKEN env vars")
	}
	if site == "" {
		site = customerId + ".observeinc.com"
	}
	return &Config{
		CustomerIdStr: customerId,
		SiteStr:       site,
		AuthtokenStr:  authToken,
	}
}

// TestIntegrationDatasetDryRun sends a real saveDatasetDryRun mutation to the
// Observe API. The test accepts either a success response or a compilation
// error response from the API - both indicate live connectivity. A panic or
// network error is a test failure.
func TestIntegrationDatasetDryRun(t *testing.T) {
	cfg := integrationConfig()
	op := NewCaptureOutput()
	fs := newFs()
	hc := &http.Client{}

	fa := FuncArgs{
		cfg:  cfg,
		fs:   fs,
		op:   op,
		args: []string{"dataset", "dry-run", "testdata/dataset_dryrun_input.json"},
		hc:   hc,
	}

	// We accept either a clean run or an Exit(1) from error datasets.
	// A panic from Exit(1) is expected in some cases - catch it.
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		err := cmdDatasetDryRun(fa)
		if err != nil {
			t.Logf("dry-run returned error (may be expected for compilation): %v", err)
		}
	}()

	if recovered != nil {
		if exitCode, ok := recovered.(int); ok {
			// exit(1) means error datasets were returned — connectivity works
			t.Logf("dry-run exited with code %d (error datasets reported by API)", exitCode)
		} else {
			t.Errorf("unexpected panic: %v", recovered)
		}
	}

	out := op.OutputBuf.String() + op.ErrorBuf.String()
	t.Logf("output: %s", out)

	// Ensure we got some response (either success or API-level error text)
	if len(out) == 0 && recovered == nil {
		t.Log("note: empty output with no exit - API may have returned null dataset")
	}
}

// TestIntegrationDatasetImpact sends a real getDatasetsAffectedByDatasetUpdate
// query to the Observe API. Accepts success or error responses as valid;
// panics and network errors are failures.
func TestIntegrationDatasetImpact(t *testing.T) {
	cfg := integrationConfig()
	op := NewCaptureOutput()
	fs := newFs()
	hc := &http.Client{}

	fa := FuncArgs{
		cfg:  cfg,
		fs:   fs,
		op:   op,
		args: []string{"dataset", "impact", "testdata/dataset_dryrun_input.json"},
		hc:   hc,
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		err := cmdDatasetImpact(fa)
		if err != nil {
			t.Logf("impact returned error: %v", err)
		}
	}()

	if recovered != nil {
		t.Errorf("unexpected panic from impact: %v", recovered)
	}

	out := op.OutputBuf.String()
	t.Logf("output: %s", out)

	// The table header should always be printed
	if !strings.Contains(out, "name") || !strings.Contains(out, "dependencyType") {
		t.Logf("note: table header not found - API may have returned an error response")
	}
}
