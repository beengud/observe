package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

var (
	flagsWorksheet      *pflag.FlagSet
	flagWorksheetName   string
)

var ErrWorksheetUsage = ObserveError{Msg: "usage: observe worksheet <list|get|create|delete> [args...]"}

func init() {
	flagsWorksheet = pflag.NewFlagSet("worksheet", pflag.ContinueOnError)
	flagsWorksheet.StringVar(&flagWorksheetName, "name", "", "Filter worksheets by name when listing")
	RegisterCommand(&Command{
		Name:  "worksheet",
		Help:  "List, get, create, or delete worksheets.",
		Flags: flagsWorksheet,
		Func:  cmdWorksheet,
	})
}

func cmdWorksheet(fa FuncArgs) error {
	if len(fa.args) < 2 {
		return ErrWorksheetUsage
	}
	switch fa.args[1] {
	case "list":
		return cmdWorksheetList(fa)
	case "get":
		return cmdWorksheetGet(fa)
	case "create":
		return cmdWorksheetCreate(fa)
	case "delete":
		return cmdWorksheetDelete(fa)
	default:
		return ObserveError{Msg: fmt.Sprintf("unknown worksheet subcommand %q; expected list, get, create, or delete", fa.args[1])}
	}
}

func cmdWorksheetList(fa FuncArgs) error {
	workspaceId := fa.cfg.WorkspaceIdOrName
	if workspaceId == "" {
		workspaceId = "42379913"
	}
	termMap := object{"workspaceId": workspaceId}
	if flagWorksheetName != "" {
		termMap["name"] = flagWorksheetName
	}
	obj, err := gqlWorksheetSearch.query(fa.cfg, fa.op, fa.hc, object{"terms": termMap})
	if err != nil {
		return err
	}
	if obj == nil {
		return nil
	}
	items, ok := obj.(array)
	if !ok {
		return fmt.Errorf("worksheet list: unexpected response type")
	}
	// Print header.
	fmt.Fprintf(fa.op, "%-20s %s\n", "id", "name")
	for _, item := range items {
		m, ok := item.(object)
		if !ok {
			continue
		}
		ws, ok := m["worksheet"]
		if !ok {
			continue
		}
		wsObj, ok := ws.(object)
		if !ok {
			continue
		}
		id, _ := wsObj["id"].(string)
		name, _ := wsObj["name"].(string)
		// Apply name filter if set (server-side search may return partial matches).
		if flagWorksheetName != "" && !strings.Contains(strings.ToLower(name), strings.ToLower(flagWorksheetName)) {
			continue
		}
		fmt.Fprintf(fa.op, "%-20s %s\n", id, name)
	}
	return nil
}

func cmdWorksheetGet(fa FuncArgs) error {
	if len(fa.args) != 3 {
		return ObserveError{Msg: "usage: observe worksheet get <id>"}
	}
	id := fa.args[2]
	obj, err := gqlGetWorksheet.query(fa.cfg, fa.op, fa.hc, object{"id": id})
	if err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("worksheet get: not found")
	}
	enc := json.NewEncoder(fa.op)
	enc.SetIndent("", "  ")
	return enc.Encode(obj)
}

func cmdWorksheetCreate(fa FuncArgs) error {
	if len(fa.args) != 3 {
		return ObserveError{Msg: "usage: observe worksheet create <file.json>"}
	}
	data, err := fa.fs.ReadFile(fa.args[2])
	if err != nil {
		return fmt.Errorf("worksheet: could not read file %q: %w", fa.args[2], err)
	}
	var input map[string]any
	if err := json.Unmarshal(data, &input); err != nil {
		return fmt.Errorf("worksheet: could not parse JSON from %q: %w", fa.args[2], err)
	}
	for _, f := range readOnlyWorksheetFields {
		delete(input, f)
	}
	obj, err := gqlSaveWorksheet.query(fa.cfg, fa.op, fa.hc, object{"wks": input})
	if err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("worksheet create: no result returned")
	}
	result, ok := obj.(object)
	if !ok {
		return fmt.Errorf("worksheet create: unexpected response type")
	}
	name, _ := result["name"].(string)
	id, _ := result["id"].(string)
	fmt.Fprintf(fa.op, "Created: %s (id: %s)\n", name, id)
	return nil
}

func cmdWorksheetDelete(fa FuncArgs) error {
	if len(fa.args) != 3 {
		return ObserveError{Msg: "usage: observe worksheet delete <id>"}
	}
	id := fa.args[2]
	obj, err := gqlDeleteWorksheet.query(fa.cfg, fa.op, fa.hc, object{"id": id})
	if err != nil {
		return err
	}
	if obj == nil {
		return nil
	}
	result, ok := obj.(object)
	if !ok {
		return fmt.Errorf("worksheet delete: unexpected response type")
	}
	if errMsg, ok := result["errorMessage"]; ok && errMsg != nil {
		if s, ok := errMsg.(string); ok && s != "" {
			return fmt.Errorf("worksheet delete: %s", s)
		}
	}
	fmt.Fprintf(fa.op, "Deleted worksheet %s\n", id)
	return nil
}
