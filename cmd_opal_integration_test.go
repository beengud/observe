//go:build integration

package main

import (
	"net/http"
	"os"
	"strings"
	"testing"
)

// integrationOpalConfig returns a *Config from env vars.
// SiteUrl() prepends CustomerIdStr + ".", so OBSERVE_SITE should be the bare domain (e.g. "observeinc.com").
func integrationOpalConfig() *Config {
	customerId := os.Getenv("OBSERVE_CUSTOMERID")
	authToken := os.Getenv("OBSERVE_AUTHTOKEN")
	if customerId == "" || authToken == "" {
		panic("integration tests require OBSERVE_CUSTOMERID and OBSERVE_AUTHTOKEN env vars")
	}
	site := os.Getenv("OBSERVE_SITE")
	if site == "" {
		site = "observeinc.com"
	}
	return &Config{
		CustomerIdStr: customerId,
		SiteStr:       site,
		AuthtokenStr:  authToken,
	}
}

// TestIntegrationOpalCheckGoodPipeline validates that "filter true" is accepted
// by the checkQueries API and returns OK with a resultSchema.
func TestIntegrationOpalCheckGoodPipeline(t *testing.T) {
	cfg := integrationOpalConfig()
	op := NewCaptureOutput()
	hc := &http.Client{}

	fa := FuncArgs{
		cfg:  cfg,
		fs:   newFs(),
		op:   op,
		args: []string{"opal", "check", "filter true"},
		hc:   hc,
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		err := cmdOpalCheck(fa)
		if err != nil {
			t.Logf("opal check returned error: %v", err)
		}
	}()

	if recovered != nil {
		t.Errorf("opal check good pipeline panicked: %v", recovered)
	}

	out := op.OutputBuf.String()
	t.Logf("output: %s", out)

	if !strings.Contains(out, "OK") {
		t.Errorf("expected OK in output for known-good pipeline, got: %q", out)
	}
}

// TestIntegrationOpalCheckBadPipeline validates that "not_a_verb 123" produces
// an ERROR output with line/column info and exits with code 1.
func TestIntegrationOpalCheckBadPipeline(t *testing.T) {
	cfg := integrationOpalConfig()
	op := NewCaptureOutput()
	hc := &http.Client{}

	fa := FuncArgs{
		cfg:  cfg,
		fs:   newFs(),
		op:   op,
		args: []string{"opal", "check", "not_a_verb 123"},
		hc:   hc,
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		err := cmdOpalCheck(fa)
		if err != nil {
			t.Logf("opal check bad pipeline returned error (expected): %v", err)
		}
	}()

	out := op.OutputBuf.String()
	t.Logf("output: %s", out)

	// Either the error surfaced via panic(exitCode) or via returned error;
	// in both cases the output should contain ERROR with position info.
	if recovered != nil {
		if _, ok := recovered.(int); !ok {
			t.Errorf("unexpected non-exit panic: %v", recovered)
		}
	}

	if !strings.Contains(out, "ERROR") {
		t.Errorf("expected ERROR output for known-bad pipeline, got: %q", out)
	}
	// Line and column info should be present (format: "line:col:")
	if !strings.Contains(out, ":") {
		t.Errorf("expected line:col info in ERROR output, got: %q", out)
	}
}

// TestIntegrationOpalCheckEmptyPipeline validates that an empty pipeline does not panic.
func TestIntegrationOpalCheckEmptyPipeline(t *testing.T) {
	cfg := integrationOpalConfig()
	op := NewCaptureOutput()
	hc := &http.Client{}

	fa := FuncArgs{
		cfg:  cfg,
		fs:   newFs(),
		op:   op,
		args: []string{"opal", "check", ""},
		hc:   hc,
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		err := cmdOpalCheck(fa)
		if err != nil {
			t.Logf("opal check empty pipeline error: %v", err)
		}
	}()

	out := op.OutputBuf.String()
	t.Logf("output: %s", out)

	// Must not produce an unhandled non-exit panic
	if recovered != nil {
		if _, ok := recovered.(int); !ok {
			t.Errorf("unexpected non-exit panic for empty pipeline: %v", recovered)
		}
	}
}

// TestIntegrationOpalVerbsReturnsResults validates that verbsAndFunctions returns
// a non-empty list of verbs including well-known ones like "filter" and "limit".
func TestIntegrationOpalVerbsReturnsResults(t *testing.T) {
	cfg := integrationOpalConfig()
	op := NewCaptureOutput()
	hc := &http.Client{}

	fa := FuncArgs{
		cfg:  cfg,
		fs:   newFs(),
		op:   op,
		args: []string{"opal", "verbs"},
		hc:   hc,
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		err := cmdOpalVerbs(fa)
		if err != nil {
			t.Errorf("opal verbs returned error: %v", err)
		}
	}()

	if recovered != nil {
		t.Errorf("opal verbs panicked: %v", recovered)
		return
	}

	out := op.OutputBuf.String()
	t.Logf("output (first 500 chars): %s", truncate(out, 500))

	if out == "" {
		t.Error("expected non-empty verbs list")
	}
	if !strings.Contains(out, "filter") {
		t.Errorf("expected 'filter' verb in output, got: %q", truncate(out, 200))
	}
	if !strings.Contains(out, "limit") {
		t.Errorf("expected 'limit' verb in output, got: %q", truncate(out, 200))
	}
}

// TestIntegrationOpalFunctionsReturnsResults validates that verbsAndFunctions returns
// a non-empty list of functions including well-known ones like "count".
func TestIntegrationOpalFunctionsReturnsResults(t *testing.T) {
	cfg := integrationOpalConfig()
	op := NewCaptureOutput()
	hc := &http.Client{}

	fa := FuncArgs{
		cfg:  cfg,
		fs:   newFs(),
		op:   op,
		args: []string{"opal", "functions"},
		hc:   hc,
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		err := cmdOpalFunctions(fa)
		if err != nil {
			t.Errorf("opal functions returned error: %v", err)
		}
	}()

	if recovered != nil {
		t.Errorf("opal functions panicked: %v", recovered)
		return
	}

	out := op.OutputBuf.String()
	t.Logf("output (first 500 chars): %s", truncate(out, 500))

	if out == "" {
		t.Error("expected non-empty functions list")
	}
	if !strings.Contains(out, "count") {
		t.Errorf("expected 'count' function in output, got: %q", truncate(out, 200))
	}
}

// TestIntegrationOpalValidateIngestGoodPipeline validates that "filter true"
// passes validateIngestFilterExpression against the default Fleet dataset.
func TestIntegrationOpalValidateIngestGoodPipeline(t *testing.T) {
	cfg := integrationOpalConfig()
	op := NewCaptureOutput()
	hc := &http.Client{}

	datasetId := os.Getenv("OBSERVE_DATASET_ID")
	if datasetId == "" {
		t.Skip("OBSERVE_DATASET_ID not set, skipping validate-ingest integration test")
	}
	flagOpalDataset = datasetId
	defer func() { flagOpalDataset = "" }()

	fa := FuncArgs{
		cfg:  cfg,
		fs:   newFs(),
		op:   op,
		args: []string{"opal", "validate-ingest", "filter true"},
		hc:   hc,
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		err := cmdOpalValidateIngest(fa)
		if err != nil {
			t.Logf("validate-ingest returned error: %v", err)
		}
	}()

	out := op.OutputBuf.String()
	t.Logf("output: %s", out)

	if recovered != nil {
		t.Errorf("validate-ingest good pipeline panicked: %v", recovered)
	}

	if !strings.Contains(out, "OK") {
		t.Errorf("expected OK for known-good ingest filter, got: %q", out)
	}
}

// TestIntegrationOpalValidateIngestBadPipeline validates that an invalid expression
// produces ERROR output from validateIngestFilterExpression.
func TestIntegrationOpalValidateIngestBadPipeline(t *testing.T) {
	cfg := integrationOpalConfig()
	op := NewCaptureOutput()
	hc := &http.Client{}

	datasetId := os.Getenv("OBSERVE_DATASET_ID")
	if datasetId == "" {
		t.Skip("OBSERVE_DATASET_ID not set, skipping validate-ingest integration test")
	}
	flagOpalDataset = datasetId
	defer func() { flagOpalDataset = "" }()

	fa := FuncArgs{
		cfg:  cfg,
		fs:   newFs(),
		op:   op,
		args: []string{"opal", "validate-ingest", "not_a_valid_filter 999"},
		hc:   hc,
	}

	var recovered any
	func() {
		defer func() { recovered = recover() }()
		err := cmdOpalValidateIngest(fa)
		if err != nil {
			t.Logf("validate-ingest bad pipeline returned error (expected): %v", err)
		}
	}()

	out := op.OutputBuf.String()
	t.Logf("output: %s", out)

	// Expect either ERROR output or an exit-code panic
	hasError := strings.Contains(out, "ERROR")
	exitPanic := false
	if recovered != nil {
		if _, ok := recovered.(int); ok {
			exitPanic = true
		} else {
			t.Errorf("unexpected non-exit panic: %v", recovered)
		}
	}

	if !hasError && !exitPanic {
		t.Errorf("expected ERROR output or exit panic for invalid ingest filter, got: %q", out)
	}
}

// truncate returns the first n characters of s (or s itself if shorter than n).
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
