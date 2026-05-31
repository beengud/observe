package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/pflag"
)

var (
	flagsBoard            *pflag.FlagSet
	flagBoardScaffold     string
	flagBoardSearchFolder string
)

var ErrBoardUsage = ObserveError{Msg: "usage: observe board <create|update|scaffold|set-default|clear-default> [args...]"}

func init() {
	flagsBoard = pflag.NewFlagSet("board", pflag.ContinueOnError)
	flagBoardScaffold = ""
	flagsBoard.StringVar(&flagBoardScaffold, "name", "", "Filter boards by name when listing, or set name when scaffolding")
	flagsBoard.StringVar(&flagBoardSearchFolder, "folder", "", "Filter boards by folder ID when listing")
	RegisterCommand(&Command{
		Name:  "board",
		Help:  "Create, update, scaffold, or set/clear defaults for a board (dashboard).",
		Flags: flagsBoard,
		Func:  cmdBoard,
	})
}

func cmdBoard(fa FuncArgs) error {
	if len(fa.args) < 2 {
		return ErrBoardUsage
	}
	switch fa.args[1] {
	case "create":
		return cmdBoardCreate(fa)
	case "update":
		return cmdBoardUpdate(fa)
	case "scaffold":
		return cmdBoardScaffold(fa)
	case "set-default":
		return cmdBoardSetDefault(fa)
	case "clear-default":
		return cmdBoardClearDefault(fa)
	default:
		return ObserveError{Msg: fmt.Sprintf("unknown board subcommand %q; expected create, update, scaffold, set-default, or clear-default", fa.args[1])}
	}
}

var gqlSaveBoard = compileGqlQuery(
`mutation Board_Save($input: DashboardInput!) {
		saveDashboard(dash: $input) {
			id
			name
			workspaceId
		}
	}`,
	"data", "saveDashboard",
)

// readOnlyBoardFields are fields returned by the Observe API that are not
// accepted as input by the saveDashboard mutation.
var readOnlyBoardFields = []string{"updatedDate"}

func readBoardInput(fa FuncArgs, filePath string) (map[string]any, error) {
	data, err := fa.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("board: could not read file %q: %w", filePath, err)
	}
	var input map[string]any
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("board: could not parse JSON from %q: %w", filePath, err)
	}
	for _, f := range readOnlyBoardFields {
		delete(input, f)
	}
	// Normalize stages: the GraphQL DashboardStageInput type requires "input"
	// to be defined. Stages with no dataset input must send [] not omit the field.
	if stages, ok := input["stages"].([]any); ok {
		for _, s := range stages {
			if stage, ok := s.(map[string]any); ok {
				if _, hasInput := stage["input"]; !hasInput {
					stage["input"] = []any{}
				}
			}
		}
	}
	return input, nil
}

func cmdBoardCreate(fa FuncArgs) error {
	if len(fa.args) != 3 {
		return ObserveError{Msg: "usage: observe board create <file.json>"}
	}
	input, err := readBoardInput(fa, fa.args[2])
	if err != nil {
		return err
	}
	obj, err := gqlSaveBoard.query(fa.cfg, fa.op, fa.hc, object{"input": input})
	if err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("board create: no result returned")
	}
	result, ok := obj.(object)
	if !ok {
		return fmt.Errorf("board create: unexpected response type")
	}
	name, _ := result["name"].(string)
	id, _ := result["id"].(string)
	fmt.Fprintf(fa.op, "Created: %s (id: %s)\n", name, id)
	return nil
}

func cmdBoardUpdate(fa FuncArgs) error {
	if len(fa.args) != 4 {
		return ObserveError{Msg: "usage: observe board update <id> <file.json>"}
	}
	boardId := fa.args[2]
	input, err := readBoardInput(fa, fa.args[3])
	if err != nil {
		return err
	}
	input["id"] = boardId
	obj, err := gqlSaveBoard.query(fa.cfg, fa.op, fa.hc, object{"input": input})
	if err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("board update: no result returned")
	}
	result, ok := obj.(object)
	if !ok {
		return fmt.Errorf("board update: unexpected response type")
	}
	name, _ := result["name"].(string)
	id, _ := result["id"].(string)
	fmt.Fprintf(fa.op, "Updated: %s (id: %s)\n", name, id)
	return nil
}

var boardScaffoldTemplate = map[string]any{
	"name":        "My Dashboard",
	"workspaceId": "42379913",
	"layout": map[string]any{
		"autoPack": true,
		"gridLayout": map[string]any{
			"sections": []any{
				map[string]any{
					"card": map[string]any{
						"title":    "Section",
						"closed":   false,
						"cardType": "section",
					},
					"items": []any{
						map[string]any{
							"card": map[string]any{
								"stageId":  "stage-abc123",
								"cardType": "stage",
							},
							"layout": map[string]any{
								"h": 12,
								"w": 12,
								"x": 0,
								"y": 0,
							},
						},
					},
				},
			},
		},
		"stageListLayout": map[string]any{
			"timeRange": map[string]any{
				"display":               "Past 24 hours",
				"millisFromCurrentTime": 86400000,
				"timeRangeInfo": map[string]any{
					"key":        "PRESETS",
					"name":       "Presets",
					"presetType": "PAST_24_HOURS",
				},
			},
			"isModified": false,
			"parameters": []any{},
		},
	},
	"stages": []any{
		map[string]any{
			"stageID":  "stage-abc123",
			"pipeline": "limit 100",
			"input": []any{
				map[string]any{
					"inputName": "main",
					"datasetId": "42450596",
				},
			},
		},
	},
}

func cmdBoardScaffold(fa FuncArgs) error {
	tmpl := copyMap(boardScaffoldTemplate)
	if flagBoardScaffold != "" {
		tmpl["name"] = flagBoardScaffold
	}
	enc := json.NewEncoder(fa.op)
	enc.SetIndent("", "  ")
	return enc.Encode(tmpl)
}

var gqlSetDefaultDashboard = compileGqlQuery(
`mutation Board_SetDefault($dsid: ObjectId!, $dashid: ObjectId!) {
		setDefaultDashboard(dsid: $dsid, dashid: $dashid) {
			success
			errorMessage
		}
	}`,
	"data", "setDefaultDashboard",
)

var gqlClearDefaultDashboard = compileGqlQuery(
`mutation Board_ClearDefault($dsid: ObjectId!) {
		clearDefaultDashboard(dsid: $dsid) {
			success
			errorMessage
		}
	}`,
	"data", "clearDefaultDashboard",
)

func cmdBoardSetDefault(fa FuncArgs) error {
	if len(fa.args) != 4 {
		return ObserveError{Msg: "usage: observe board set-default <dataset-id> <board-id>"}
	}
	datasetId := fa.args[2]
	boardId := fa.args[3]
	obj, err := gqlSetDefaultDashboard.query(fa.cfg, fa.op, fa.hc, object{"dsid": datasetId, "dashid": boardId})
	if err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("board set-default: no result returned")
	}
	result, ok := obj.(object)
	if !ok {
		return fmt.Errorf("board set-default: unexpected response type")
	}
	if errMsg, ok := result["errorMessage"]; ok && errMsg != nil {
		if s, ok := errMsg.(string); ok && s != "" {
			return fmt.Errorf("board set-default: %s", s)
		}
	}
	fmt.Fprintf(fa.op, "Default dashboard set successfully\n")
	return nil
}

func cmdBoardClearDefault(fa FuncArgs) error {
	if len(fa.args) != 3 {
		return ObserveError{Msg: "usage: observe board clear-default <dataset-id>"}
	}
	datasetId := fa.args[2]
	obj, err := gqlClearDefaultDashboard.query(fa.cfg, fa.op, fa.hc, object{"dsid": datasetId})
	if err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("board clear-default: no result returned")
	}
	result, ok := obj.(object)
	if !ok {
		return fmt.Errorf("board clear-default: unexpected response type")
	}
	if errMsg, ok := result["errorMessage"]; ok && errMsg != nil {
		if s, ok := errMsg.(string); ok && s != "" {
			return fmt.Errorf("board clear-default: %s", s)
		}
	}
	fmt.Fprintf(fa.op, "Default dashboard cleared successfully\n")
	return nil
}

func copyMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
