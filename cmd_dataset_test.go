package main

import (
	"strings"
	"testing"
)

const testDryRunInput = `{
  "workspaceId": "42379913",
  "dataset": { "name": "MyDataset" },
  "query": { "stageQueries": [{ "stageID": "main", "pipeline": "filter true" }] }
}`

func TestCmdDatasetNoArgs(t *testing.T) {
	fix := startFixture(t)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset"}, fix.hc)
	})
	fix.Assert()
}

func TestCmdDatasetUnknownSubcommand(t *testing.T) {
	fix := startFixture(t)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "frobulate"}, fix.hc)
	})
	fix.Assert()
}

func TestCmdDatasetDryRunMissingArgs(t *testing.T) {
	fix := startFixture(t)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "dry-run"}, fix.hc)
	})
	fix.Assert()
}

func TestCmdDatasetDryRunMalformedJSON(t *testing.T) {
	fix := startFixture(t)
	fix.fs.WriteFile("bad.json", []byte(`not valid json`), 0)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "dry-run", "bad.json"}, fix.hc)
	})
	if !strings.Contains(fix.op.ErrorBuf.String(), "could not parse JSON") {
		t.Error("expected JSON parse error in output:", fix.op.ErrorBuf.String())
	}
	fix.Assert()
}

func TestCmdDatasetDryRunMissingFile(t *testing.T) {
	fix := startFixture(t)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "dry-run", "nonexistent.json"}, fix.hc)
	})
	if !strings.Contains(fix.op.ErrorBuf.String(), "could not read file") {
		t.Error("expected file-not-found error in output:", fix.op.ErrorBuf.String())
	}
	fix.Assert()
}

func TestCmdDatasetDryRunSuccess(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"saveDatasetDryRun":{
			"dataset":{"id":"99001","name":"MyDataset"},
			"dematerializedDatasets":[{"id":"88001","name":"DownstreamA"}],
			"errorDatasets":[]
		}}}`},
	)
	fix.fs.WriteFile("input.json", []byte(testDryRunInput), 0)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "dry-run", "input.json"}, fix.hc)
	fix.Assert()

	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "Dataset: MyDataset (99001)") {
		t.Errorf("expected dataset line in output, got: %q", out)
	}
	if !strings.Contains(out, "Would rematerialize: DownstreamA (88001)") {
		t.Errorf("expected dematerialized line in output, got: %q", out)
	}
	if fix.op.ErrorBuf.Len() > 0 {
		t.Errorf("unexpected error output: %q", fix.op.ErrorBuf.String())
	}
}

func TestCmdDatasetDryRunWithErrors(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"saveDatasetDryRun":{
			"dataset":null,
			"dematerializedDatasets":[],
			"errorDatasets":[{"dataset":{"id":"77001","name":"BadDataset"},"errorText":"syntax error near token"}]
		}}}`},
	)
	fix.fs.WriteFile("input.json", []byte(testDryRunInput), 0)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "dry-run", "input.json"}, fix.hc)
	})
	fix.Assert()

	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "Error in BadDataset: syntax error near token") {
		t.Errorf("expected error dataset line in output, got: %q", out)
	}
}

func TestCmdDatasetDryRunNoErrorDatasets(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"saveDatasetDryRun":{
			"dataset":{"id":"99002","name":"CleanDataset"},
			"dematerializedDatasets":[],
			"errorDatasets":[]
		}}}`},
	)
	fix.fs.WriteFile("input.json", []byte(testDryRunInput), 0)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "dry-run", "input.json"}, fix.hc)
	fix.Assert()

	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "Dataset: CleanDataset (99002)") {
		t.Errorf("expected dataset line, got: %q", out)
	}
	if strings.Contains(out, "Would rematerialize") {
		t.Errorf("unexpected rematerialization line: %q", out)
	}
}

func TestCmdDatasetImpactMissingArgs(t *testing.T) {
	fix := startFixture(t)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "impact"}, fix.hc)
	})
	fix.Assert()
}

func TestCmdDatasetImpactMalformedJSON(t *testing.T) {
	fix := startFixture(t)
	fix.fs.WriteFile("bad.json", []byte(`{invalid`), 0)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "impact", "bad.json"}, fix.hc)
	})
	if !strings.Contains(fix.op.ErrorBuf.String(), "could not parse JSON") {
		t.Error("expected JSON parse error:", fix.op.ErrorBuf.String())
	}
	fix.Assert()
}

func TestCmdDatasetImpactSuccess(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"getDatasetsAffectedByDatasetUpdate":{
			"affectedDatasets":[
				{"dataset":{"id":"55001","name":"Alpha"},"dependencyType":"DIRECT"},
				{"dataset":{"id":"55002","name":"Beta"},"dependencyType":"INDIRECT"}
			],
			"errorDatasets":[]
		}}}`},
	)
	fix.fs.WriteFile("input.json", []byte(testDryRunInput), 0)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "impact", "input.json"}, fix.hc)
	fix.Assert()

	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "Alpha") {
		t.Errorf("expected Alpha in output, got: %q", out)
	}
	if !strings.Contains(out, "55001") {
		t.Errorf("expected id 55001 in output, got: %q", out)
	}
	if !strings.Contains(out, "DIRECT") {
		t.Errorf("expected DIRECT dependencyType in output, got: %q", out)
	}
	if !strings.Contains(out, "Beta") {
		t.Errorf("expected Beta in output, got: %q", out)
	}
	if !strings.Contains(out, "INDIRECT") {
		t.Errorf("expected INDIRECT dependencyType in output, got: %q", out)
	}
}

func TestCmdDatasetImpactEmptyAffectedList(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"getDatasetsAffectedByDatasetUpdate":{
			"affectedDatasets":[],
			"errorDatasets":[]
		}}}`},
	)
	fix.fs.WriteFile("input.json", []byte(testDryRunInput), 0)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "impact", "input.json"}, fix.hc)
	fix.Assert()

	out := fix.op.OutputBuf.String()
	// Header should still be printed even with empty results
	if !strings.Contains(out, "name") || !strings.Contains(out, "id") || !strings.Contains(out, "dependencyType") {
		t.Errorf("expected table header in output, got: %q", out)
	}
}

func TestCmdDatasetImpactWithErrors(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"getDatasetsAffectedByDatasetUpdate":{
			"affectedDatasets":[],
			"errorDatasets":[{"dataset":{"id":"66001","name":"ErrorDs"},"errorText":"compilation failed"}]
		}}}`},
	)
	fix.fs.WriteFile("input.json", []byte(testDryRunInput), 0)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "impact", "input.json"}, fix.hc)
	fix.Assert()

	errOut := fix.op.ErrorBuf.String()
	if !strings.Contains(errOut, "Error in ErrorDs: compilation failed") {
		t.Errorf("expected error dataset in error output, got: %q", errOut)
	}
}

// TestCmdDatasetDryRunMultipleErrors verifies that multiple errorDatasets are all printed
// and that exit 1 is still triggered.
func TestCmdDatasetDryRunMultipleErrors(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"saveDatasetDryRun":{
			"dataset":null,
			"dematerializedDatasets":[],
			"errorDatasets":[
				{"dataset":{"id":"77001","name":"Err1"},"errorText":"first error"},
				{"dataset":{"id":"77002","name":"Err2"},"errorText":"second error"}
			]
		}}}`},
	)
	fix.fs.WriteFile("input.json", []byte(testDryRunInput), 0)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "dry-run", "input.json"}, fix.hc)
	})
	fix.Assert()

	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "Error in Err1: first error") {
		t.Errorf("expected first error in output, got: %q", out)
	}
	if !strings.Contains(out, "Error in Err2: second error") {
		t.Errorf("expected second error in output, got: %q", out)
	}
}

// TestCmdDatasetDryRunGqlError verifies that a GraphQL-level error (errors field in response)
// is surfaced as an error.
func TestCmdDatasetDryRunGqlError(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"errors":[{"message":"unauthorized"}]}`},
	)
	fix.fs.WriteFile("input.json", []byte(testDryRunInput), 0)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "dry-run", "input.json"}, fix.hc)
	})
	fix.Assert()

	if !strings.Contains(fix.op.ErrorBuf.String(), "unauthorized") {
		t.Errorf("expected GQL error in output, got: %q", fix.op.ErrorBuf.String())
	}
}

// TestCmdDatasetImpactGqlError verifies that a GraphQL-level error is surfaced.
func TestCmdDatasetImpactGqlError(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"errors":[{"message":"forbidden"}]}`},
	)
	fix.fs.WriteFile("input.json", []byte(testDryRunInput), 0)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "impact", "input.json"}, fix.hc)
	})
	fix.Assert()

	if !strings.Contains(fix.op.ErrorBuf.String(), "forbidden") {
		t.Errorf("expected GQL error in output, got: %q", fix.op.ErrorBuf.String())
	}
}

// TestCmdDatasetImpactMissingFile verifies that a missing input file produces an error.
func TestCmdDatasetImpactMissingFile(t *testing.T) {
	fix := startFixture(t)
	mustPanic(t, func() {
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"dataset", "impact", "nonexistent.json"}, fix.hc)
	})
	if !strings.Contains(fix.op.ErrorBuf.String(), "could not read file") {
		t.Error("expected file-not-found error in output:", fix.op.ErrorBuf.String())
	}
	fix.Assert()
}
