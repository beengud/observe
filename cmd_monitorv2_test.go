package main

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// --- Issue #15: list monitor / get monitor ---

func TestCmdListMonitor(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"searchMonitorV2":{"monitors":[
			{"id":"mon-001","name":"CPU Alert","description":"CPU usage monitor","disabled":false,"updatedDate":"2024-01-01T00:00:00Z"},
			{"id":"mon-002","name":"Memory Alert","description":"Memory usage monitor","disabled":true,"updatedDate":"2024-01-02T00:00:00Z"}
		]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"list", "monitor"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "mon-001") {
		t.Error("expected mon-001 in output:", out)
	}
	if !strings.Contains(out, "CPU Alert") {
		t.Error("expected CPU Alert in output:", out)
	}
	if !strings.Contains(out, "mon-002") {
		t.Error("expected mon-002 in output:", out)
	}
	if !strings.Contains(out, "Memory Alert") {
		t.Error("expected Memory Alert in output:", out)
	}
}

func TestCmdListMonitorEmpty(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"searchMonitorV2":{"monitors":[]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"list", "monitor"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	// With empty monitors, the table should only have the header row (no data rows).
	out := fix.op.OutputBuf.String()
	// header should be present
	if !strings.Contains(out, "id") {
		t.Error("expected header row in output:", out)
	}
	// no data rows expected
	if strings.Contains(out, "mon-") {
		t.Error("unexpected monitor data in empty list output:", out)
	}
}

func TestCmdListMonitorDefaultWorkspace(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"searchMonitorV2":{"monitors":[]}}}`},
	)
	// No workspace set in cfg, should use default 42379913
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"list", "monitor"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	// Verify the request was made (rix == 1 means the request was consumed)
	fix.Assert()
}

func TestCmdGetMonitor(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"monitorV2":{
			"id":"mon-001",
			"name":"CPU Alert",
			"description":"CPU usage monitor",
			"disabled":false,
			"updatedDate":"2024-01-01T00:00:00Z",
			"definition":{"compareFunction":"GREATER","countAggFunction":"COUNT","threshold":90}
		}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"get", "monitor", "mon-001"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "mon-001") {
		t.Error("expected id in output:", out)
	}
	if !strings.Contains(out, "CPU Alert") {
		t.Error("expected name in output:", out)
	}
	if !strings.Contains(out, "monitor") {
		t.Error("expected type in output:", out)
	}
}

func TestCmdListMonitorTableColumns(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"searchMonitorV2":{"monitors":[
			{"id":"mon-abc","name":"Test Monitor","description":"","disabled":false,"updatedDate":"2024-06-01T12:00:00Z"}
		]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"list", "monitor"}, fix.hc)
	out := fix.op.OutputBuf.String()
	// Table should contain all 4 presentation columns
	if !strings.Contains(out, "id") {
		t.Error("missing 'id' column header:", out)
	}
	if !strings.Contains(out, "name") {
		t.Error("missing 'name' column header:", out)
	}
	if !strings.Contains(out, "disabled") {
		t.Error("missing 'disabled' column header:", out)
	}
	if !strings.Contains(out, "updatedDate") {
		t.Error("missing 'updatedDate' column header:", out)
	}
	if !strings.Contains(out, "mon-abc") {
		t.Error("missing monitor id in data:", out)
	}
}

func TestCmdMonitorUnknownSubcommand(t *testing.T) {
	fix := startFixture(t)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from Exit on error")
		}
	}()
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "unknown-subcommand"}, fix.hc)
}

func TestCmdMonitorNoSubcommand(t *testing.T) {
	fix := startFixture(t)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from Exit on error")
		}
	}()
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor"}, fix.hc)
}

// --- Issue #16: preview-query ---

func TestCmdMonitorPreviewQuery(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"evaluateMonitorV2Source":{
			"pipeline":"filter value > 90",
			"resultSchema":{"fields":[
				{"name":"timestamp","type":"time"},
				{"name":"value","type":"float64"}
			]}
		}}}`},
	)
	input := `{"name":"CPU Alert","ruleKind":"COUNT","definition":{"compareFunction":"GREATER","threshold":90}}`
	fix.fs.WriteFile("monitor_input.json", []byte(input), 0)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "preview-query", "monitor_input.json"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "filter value > 90") {
		t.Error("expected pipeline in output:", out)
	}
	if !strings.Contains(out, "timestamp") {
		t.Error("expected field name in output:", out)
	}
	if !strings.Contains(out, "time") {
		t.Error("expected field type in output:", out)
	}
}

func TestCmdMonitorPreviewQueryMissingFile(t *testing.T) {
	fix := startFixture(t)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from Exit on error")
		}
	}()
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "preview-query", "nonexistent.json"}, fix.hc)
}

func TestCmdMonitorPreviewQueryMissingArg(t *testing.T) {
	fix := startFixture(t)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from Exit on error")
		}
	}()
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "preview-query"}, fix.hc)
}

func TestCmdMonitorPreviewQueryNoSchema(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"evaluateMonitorV2Source":{
			"pipeline":"filter count > 5",
			"resultSchema":null
		}}}`},
	)
	input := `{"name":"Count Alert","ruleKind":"COUNT","definition":{"compareFunction":"GREATER","threshold":5}}`
	fix.fs.WriteFile("monitor_input2.json", []byte(input), 0)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "preview-query", "monitor_input2.json"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "filter count > 5") {
		t.Error("expected pipeline in output:", out)
	}
}

// --- Issue #17: preview ---

func TestCmdMonitorPreviewWouldFire(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"previewMonitorV2":{
			"wouldFire":true,
			"samples":[
				{"groupings":[{"name":"host","value":"web-01"}],"level":"critical","timestamp":"2024-01-01T00:00:00Z"}
			]
		}}}`},
	)
	input := `{"name":"CPU Alert","ruleKind":"COUNT","definition":{"compareFunction":"GREATER","threshold":90}}`
	fix.fs.WriteFile("preview_input.json", []byte(input), 0)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "preview", "preview_input.json"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "Would fire: true") {
		t.Error("expected 'Would fire: true' in output:", out)
	}
	if !strings.Contains(out, "critical") {
		t.Error("expected level in output:", out)
	}
	if !strings.Contains(out, "web-01") {
		t.Error("expected grouping value in output:", out)
	}
}

func TestCmdMonitorPreviewWouldNotFire(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"previewMonitorV2":{
			"wouldFire":false,
			"samples":[]
		}}}`},
	)
	input := `{"name":"CPU Alert","ruleKind":"COUNT","definition":{"compareFunction":"GREATER","threshold":90}}`
	fix.fs.WriteFile("preview_input2.json", []byte(input), 0)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "preview", "preview_input2.json"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "Would fire: false") {
		t.Error("expected 'Would fire: false' in output:", out)
	}
}

func TestCmdMonitorPreviewMultipleGroupings(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"previewMonitorV2":{
			"wouldFire":true,
			"samples":[
				{"groupings":[{"name":"host","value":"db-01"},{"name":"env","value":"prod"}],"level":"warning","timestamp":"2024-01-01T06:00:00Z"}
			]
		}}}`},
	)
	input := `{"name":"DB Alert","ruleKind":"COUNT","definition":{"compareFunction":"GREATER","threshold":50}}`
	fix.fs.WriteFile("preview_input3.json", []byte(input), 0)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "preview", "preview_input3.json"}, fix.hc)
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "db-01") {
		t.Error("expected first grouping value in output:", out)
	}
	if !strings.Contains(out, "prod") {
		t.Error("expected second grouping value in output:", out)
	}
	if !strings.Contains(out, "warning") {
		t.Error("expected level in output:", out)
	}
}

func TestCmdMonitorPreviewMissingArg(t *testing.T) {
	fix := startFixture(t)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from Exit on error")
		}
	}()
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "preview"}, fix.hc)
}

// --- Issue #18: alarms ---

func TestCmdMonitorAlarms(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"searchMonitorV2Alarms":{"alarms":[
			{"id":"alarm-001","monitorId":"mon-001","level":"critical","status":"active","startTime":"2024-01-01T00:00:00Z","endTime":"","groupings":[{"name":"host","value":"web-01"}]},
			{"id":"alarm-002","monitorId":"mon-002","level":"warning","status":"resolved","startTime":"2024-01-01T01:00:00Z","endTime":"2024-01-01T02:00:00Z","groupings":[]}
		]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "alarms"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "alarm-001") {
		t.Error("expected alarm-001 in output:", out)
	}
	if !strings.Contains(out, "critical") {
		t.Error("expected level in output:", out)
	}
	if !strings.Contains(out, "alarm-002") {
		t.Error("expected alarm-002 in output:", out)
	}
	if !strings.Contains(out, "warning") {
		t.Error("expected warning level in output:", out)
	}
}

func TestCmdMonitorAlarmsEmpty(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"searchMonitorV2Alarms":{"alarms":[]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "alarms"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	// With no alarms, output should have header only
	out := fix.op.OutputBuf.String()
	if strings.Contains(out, "alarm-") {
		t.Error("unexpected alarm data in empty output:", out)
	}
}

func TestCmdMonitorAlarmsTableColumns(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"searchMonitorV2Alarms":{"alarms":[
			{"id":"alarm-xyz","monitorId":"mon-abc","level":"informational","status":"active","startTime":"2024-06-01T00:00:00Z","endTime":"","groupings":[]}
		]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "alarms"}, fix.hc)
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "id") {
		t.Error("missing 'id' column:", out)
	}
	if !strings.Contains(out, "monitorId") {
		t.Error("missing 'monitorId' column:", out)
	}
	if !strings.Contains(out, "level") {
		t.Error("missing 'level' column:", out)
	}
	if !strings.Contains(out, "status") {
		t.Error("missing 'status' column:", out)
	}
	if !strings.Contains(out, "startTime") {
		t.Error("missing 'startTime' column:", out)
	}
	if !strings.Contains(out, "endTime") {
		t.Error("missing 'endTime' column:", out)
	}
}

func TestCmdMonitorAlarmsMultipleGroupings(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"searchMonitorV2Alarms":{"alarms":[
			{"id":"alarm-mg","monitorId":"mon-001","level":"error","status":"active","startTime":"2024-06-01T00:00:00Z","endTime":"","groupings":[{"name":"host","value":"db-01"},{"name":"region","value":"us-east-1"}]}
		]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "alarms"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "alarm-mg") {
		t.Error("expected alarm id in output:", out)
	}
	if !strings.Contains(out, "error") {
		t.Error("expected level in output:", out)
	}
}

func TestMonitorV2ObjectTypeRegistered(t *testing.T) {
	ot := GetObjectType("monitor")
	if ot == nil {
		t.Fatal("expected 'monitor' object type to be registered")
	}
	if ot.TypeName() != "monitor" {
		t.Errorf("expected TypeName 'monitor', got %q", ot.TypeName())
	}
	if !ot.CanList() {
		t.Error("expected CanList() == true")
	}
	if !ot.CanGet() {
		t.Error("expected CanGet() == true")
	}
	if ot.CanCreate() {
		t.Error("expected CanCreate() == false")
	}
	if ot.CanDelete() {
		t.Error("expected CanDelete() == false")
	}
}

func TestMonitorV2PresentationLabels(t *testing.T) {
	ot := GetObjectType("monitor")
	if ot == nil {
		t.Fatal("expected 'monitor' object type to be registered")
	}
	labels := ot.GetPresentationLabels()
	expected := []string{"id", "name", "disabled", "updatedDate"}
	if diff := cmp.Diff(labels, expected); diff != "" {
		t.Errorf("unexpected presentation labels:\n%s", diff)
	}
}

// --- Issue #19: additional unit test coverage ---

func TestCmdListMonitorDisabledTrue(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"searchMonitorV2":{"monitors":[
			{"id":"mon-disabled","name":"Disabled Monitor","description":"","disabled":true,"updatedDate":"2024-01-01T00:00:00Z"}
		]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"list", "monitor"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "mon-disabled") {
		t.Error("expected disabled monitor id in output:", out)
	}
	if !strings.Contains(out, "true") {
		t.Error("expected 'true' disabled value in output:", out)
	}
}

func TestCmdListMonitorDisabledFalse(t *testing.T) {
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"searchMonitorV2":{"monitors":[
			{"id":"mon-enabled","name":"Active Monitor","description":"","disabled":false,"updatedDate":"2024-02-01T00:00:00Z"}
		]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"list", "monitor"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "false") {
		t.Error("expected 'false' disabled value in output:", out)
	}
}

func TestCmdMonitorPreviewQueryInvalidJSON(t *testing.T) {
	fix := startFixture(t)
	fix.fs.WriteFile("bad.json", []byte("{not valid json"), 0)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from Exit on error with invalid JSON")
		}
	}()
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "preview-query", "bad.json"}, fix.hc)
}

func TestCmdMonitorPreviewInvalidJSON(t *testing.T) {
	fix := startFixture(t)
	fix.fs.WriteFile("bad2.json", []byte("{not valid json"), 0)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from Exit on error with invalid JSON")
		}
	}()
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "preview", "bad2.json"}, fix.hc)
}

func TestCmdMonitorAlarmsSingleGrouping(t *testing.T) {
	// Verify that alarms with a single grouping are properly included in the table output
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"searchMonitorV2Alarms":{"alarms":[
			{"id":"alarm-single","monitorId":"mon-001","level":"critical","status":"active","startTime":"2024-06-01T00:00:00Z","endTime":"","groupings":[{"name":"host","value":"prod-web-01"}]}
		]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "alarms"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "alarm-single") {
		t.Error("expected alarm id in output:", out)
	}
	if !strings.Contains(out, "critical") {
		t.Error("expected level in output:", out)
	}
}

func TestMonitorV2FromObjectHelper(t *testing.T) {
	m := object{
		"id":          "mon-test",
		"name":        "Test Monitor",
		"description": "A test",
		"disabled":    true,
		"updatedDate": "2024-01-01T00:00:00Z",
	}
	o := monitorV2FromObject(m)
	if o.Id != "mon-test" {
		t.Errorf("expected Id 'mon-test', got %q", o.Id)
	}
	if o.Name != "Test Monitor" {
		t.Errorf("expected Name 'Test Monitor', got %q", o.Name)
	}
	if o.Description != "A test" {
		t.Errorf("expected Description 'A test', got %q", o.Description)
	}
	if o.Disabled != "true" {
		t.Errorf("expected Disabled 'true', got %q", o.Disabled)
	}
	if o.UpdatedDate != "2024-01-01T00:00:00Z" {
		t.Errorf("expected UpdatedDate '2024-01-01T00:00:00Z', got %q", o.UpdatedDate)
	}
}

func TestMonitorV2FromObjectHelperNilFields(t *testing.T) {
	// Ensure monitorV2FromObject handles nil field values without panicking
	m := object{
		"id":   nil,
		"name": nil,
	}
	o := monitorV2FromObject(m)
	if o.Id != "" {
		t.Errorf("expected empty Id, got %q", o.Id)
	}
	if o.Name != "" {
		t.Errorf("expected empty Name, got %q", o.Name)
	}
}

func TestCmdGetMonitorDefinitionJSON(t *testing.T) {
	// Verify that the definition field is marshalled to JSON in get output
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"monitorV2":{
			"id":"mon-def",
			"name":"Def Monitor",
			"description":"",
			"disabled":false,
			"updatedDate":"2024-01-01T00:00:00Z",
			"definition":{"compareFunction":"GREATER","countAggFunction":"COUNT","threshold":100}
		}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"get", "monitor", "mon-def"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "mon-def") {
		t.Error("expected id in output:", out)
	}
	// The definition should appear as JSON in the state section
	if !strings.Contains(out, "GREATER") {
		t.Error("expected definition content in output:", out)
	}
}

func TestCmdMonitorPreviewQueryEmptyFields(t *testing.T) {
	// Verify empty result schema fields array is handled gracefully
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"evaluateMonitorV2Source":{
			"pipeline":"filter true",
			"resultSchema":{"fields":[]}
		}}}`},
	)
	input := `{"name":"Empty Schema","ruleKind":"COUNT","definition":{"compareFunction":"GREATER","threshold":0}}`
	fix.fs.WriteFile("empty_schema.json", []byte(input), 0)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "preview-query", "empty_schema.json"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "filter true") {
		t.Error("expected pipeline in output:", out)
	}
}

func TestCmdMonitorAlarmsNoSamples(t *testing.T) {
	// Alarm result set with nil groupings field
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, `{"data":{"searchMonitorV2Alarms":{"alarms":[
			{"id":"alarm-nogroupings","monitorId":"mon-001","level":"warning","status":"resolved","startTime":"2024-06-01T00:00:00Z","endTime":"2024-06-01T01:00:00Z","groupings":null}
		]}}}`},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"monitor", "alarms"}, fix.hc)
	if diff := fix.op.ErrorBuf.String(); diff != "" {
		t.Error("unexpected error output:", diff)
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "alarm-nogroupings") {
		t.Error("expected alarm id in output:", out)
	}
}
