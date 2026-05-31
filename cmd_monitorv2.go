package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

var (
	flagsMonitor     *pflag.FlagSet
	flagMonitorId    string
	flagMonitorSince time.Duration
	flagMonitorLevel string
)

var ErrMonitorUsage = ObserveError{Msg: "usage: observe monitor <preview-query|preview|alarms> [args...]"}

func init() {
	flagsMonitor = pflag.NewFlagSet("monitor", pflag.ContinueOnError)
	flagsMonitor.StringVar(&flagMonitorId, "monitor-id", "", "Monitor ID to filter alarms")
	flagsMonitor.DurationVar(&flagMonitorSince, "since", 24*time.Hour, "Duration to look back for alarms (e.g. 24h, 7d)")
	flagsMonitor.StringVar(&flagMonitorLevel, "level", "", "Alarm level filter: critical|error|warning|informational")
	RegisterCommand(&Command{
		Name:  "monitor",
		Help:  "Query Monitor V2 resources: preview-query, preview, and alarms.",
		Flags: flagsMonitor,
		Func:  cmdMonitor,
	})
}

func cmdMonitor(fa FuncArgs) error {
	if len(fa.args) < 2 {
		return ErrMonitorUsage
	}
	switch fa.args[1] {
	case "preview-query":
		return cmdMonitorPreviewQuery(fa)
	case "preview":
		return cmdMonitorPreview(fa)
	case "alarms":
		return cmdMonitorAlarms(fa)
	default:
		return ObserveError{Msg: fmt.Sprintf("unknown monitor subcommand %q; expected preview-query, preview, or alarms", fa.args[1])}
	}
}

// resolveMonitorWorkspace returns the workspace ID from config or default.
// The global --workspace flag sets cfg.WorkspaceIdOrName.
func resolveMonitorWorkspace(cfg *Config) string {
	if cfg.WorkspaceIdOrName != "" {
		return cfg.WorkspaceIdOrName
	}
	return "42379913"
}

// readMonitorV2Input reads a MonitorV2Input JSON file from disk.
func readMonitorV2Input(fa FuncArgs, filePath string) (map[string]any, error) {
	data, err := fa.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("monitor: could not read file %q: %w", filePath, err)
	}
	var input map[string]any
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("monitor: could not parse JSON from %q: %w", filePath, err)
	}
	return input, nil
}

// ---- preview-query ----

var gqlEvaluateMonitorV2Source = compileGqlQuery(
	`query EvaluateMonitorV2Source($input: MonitorV2Input!) {
		evaluateMonitorV2Source(input: $input) {
			pipeline
			resultSchema { fields { name type } }
		}
	}`,
	"data", "evaluateMonitorV2Source",
)

func cmdMonitorPreviewQuery(fa FuncArgs) error {
	// args: ["monitor", "preview-query", "<file.json>"]
	if len(fa.args) != 3 {
		return ObserveError{Msg: "usage: observe monitor preview-query <file.json>"}
	}
	input, err := readMonitorV2Input(fa, fa.args[2])
	if err != nil {
		return err
	}
	obj, err := gqlEvaluateMonitorV2Source.query(fa.cfg, fa.op, fa.hc, object{"input": input})
	if err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("monitor preview-query: no result returned")
	}
	result, ok := obj.(object)
	if !ok {
		return fmt.Errorf("monitor preview-query: unexpected response type")
	}
	pipeline, _ := result["pipeline"].(string)
	fmt.Fprintf(fa.op, "Pipeline:\n%s\n", pipeline)

	if schemaObj, ok := result["resultSchema"]; ok && schemaObj != nil {
		if schema, ok := schemaObj.(object); ok {
			if fieldsAny, ok := schema["fields"]; ok && fieldsAny != nil {
				if fields, ok := fieldsAny.(array); ok {
					fmt.Fprintf(fa.op, "Result schema fields:\n")
					for _, f := range fields {
						if fObj, ok := f.(object); ok {
							name, _ := fObj["name"].(string)
							typ, _ := fObj["type"].(string)
							fmt.Fprintf(fa.op, "  %s: %s\n", name, typ)
						}
					}
				}
			}
		}
	}
	return nil
}

// ---- preview ----

var gqlPreviewMonitorV2 = compileGqlQuery(
	`query PreviewMonitorV2($workspaceId: ObjectId!, $input: MonitorV2Input!) {
		previewMonitorV2(workspaceId: $workspaceId, input: $input) {
			wouldFire
			samples { groupings { name value } level timestamp }
		}
	}`,
	"data", "previewMonitorV2",
)

func cmdMonitorPreview(fa FuncArgs) error {
	// args: ["monitor", "preview", "<file.json>"]
	if len(fa.args) != 3 {
		return ObserveError{Msg: "usage: observe monitor preview <file.json>"}
	}
	input, err := readMonitorV2Input(fa, fa.args[2])
	if err != nil {
		return err
	}
	workspaceId := resolveMonitorWorkspace(fa.cfg)
	obj, err := gqlPreviewMonitorV2.query(fa.cfg, fa.op, fa.hc, object{
		"workspaceId": workspaceId,
		"input":       input,
	})
	if err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("monitor preview: no result returned")
	}
	result, ok := obj.(object)
	if !ok {
		return fmt.Errorf("monitor preview: unexpected response type")
	}
	wouldFire, _ := result["wouldFire"].(bool)
	fmt.Fprintf(fa.op, "Would fire: %v\n", wouldFire)

	if samplesAny, ok := result["samples"]; ok && samplesAny != nil {
		if samples, ok := samplesAny.(array); ok && len(samples) > 0 {
			fmt.Fprintf(fa.op, "Samples:\n")
			for _, s := range samples {
				sObj, ok := s.(object)
				if !ok {
					continue
				}
				level, _ := sObj["level"].(string)
				timestamp, _ := sObj["timestamp"].(string)
				fmt.Fprintf(fa.op, "  level=%s timestamp=%s", level, timestamp)
				if groupingsAny, ok := sObj["groupings"]; ok && groupingsAny != nil {
					if groupings, ok := groupingsAny.(array); ok {
						for _, g := range groupings {
							gObj, ok := g.(object)
							if !ok {
								continue
							}
							gName, _ := gObj["name"].(string)
							gValue, _ := gObj["value"].(string)
							fmt.Fprintf(fa.op, " %s=%s", gName, gValue)
						}
					}
				}
				fmt.Fprintf(fa.op, "\n")
			}
		}
	}
	return nil
}

// ---- alarms ----

var gqlSearchMonitorV2Alarms = compileGqlQuery(
	`query SearchMonitorV2Alarms($workspaceId: ObjectId!, $monitorId: ObjectId, $startTime: Time, $endTime: Time, $levels: [MonitorV2AlarmLevel!]) {
		searchMonitorV2Alarms(workspaceId: $workspaceId, monitorId: $monitorId, startTime: $startTime, endTime: $endTime, levels: $levels) {
			alarms { id monitorId level status startTime endTime groupings { name value } }
		}
	}`,
	"data", "searchMonitorV2Alarms", "alarms",
)

func cmdMonitorAlarms(fa FuncArgs) error {
	workspaceId := resolveMonitorWorkspace(fa.cfg)

	args := object{"workspaceId": workspaceId}

	if flagMonitorId != "" {
		args["monitorId"] = flagMonitorId
	}

	if flagMonitorSince > 0 {
		startTime := time.Now().Add(-flagMonitorSince).UTC().Format(time.RFC3339)
		args["startTime"] = startTime
	}

	if flagMonitorLevel != "" {
		args["levels"] = []string{flagMonitorLevel}
	}

	obj, err := gqlSearchMonitorV2Alarms.query(fa.cfg, fa.op, fa.hc, args)
	if err != nil {
		return err
	}
	if obj == nil {
		return nil
	}
	alarms, ok := obj.(array)
	if !ok {
		return fmt.Errorf("monitor alarms: unexpected response type")
	}

	out := &ColumnFormatter{
		Output:          fa.op,
		OmitLineDrawing: true,
	}
	out.SetColumnNames([]string{"id", "monitorId", "level", "status", "startTime", "endTime"})
	for _, a := range alarms {
		aObj, ok := a.(object)
		if !ok {
			continue
		}
		id, _ := aObj["id"].(string)
		monitorId, _ := aObj["monitorId"].(string)
		level, _ := aObj["level"].(string)
		status, _ := aObj["status"].(string)
		startTime, _ := aObj["startTime"].(string)
		endTime, _ := aObj["endTime"].(string)
		out.AddRow([]string{id, monitorId, level, status, startTime, endTime})
	}
	out.Close()
	return nil
}
