package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

var (
	flagsOpal       *pflag.FlagSet
	flagOpalFile    string
	flagOpalDataset string
)

var ErrOpalUsage = ObserveError{Msg: "usage: observe opal <check|verbs|functions|validate-ingest> [args...]"}

func init() {
	flagsOpal = pflag.NewFlagSet("opal", pflag.ContinueOnError)
	flagsOpal.StringVarP(&flagOpalFile, "file", "f", "", "Read OPAL pipeline from file instead of argument")
	flagsOpal.StringVar(&flagOpalDataset, "dataset", "", "Source dataset ID for validate-ingest subcommand")
	RegisterCommand(&Command{
		Name:  "opal",
		Help:  "Validate and inspect OPAL pipelines and functions.",
		Flags: flagsOpal,
		Func:  cmdOpal,
	})
}

func cmdOpal(fa FuncArgs) error {
	if len(fa.args) < 2 {
		return ErrOpalUsage
	}
	switch fa.args[1] {
	case "check":
		return cmdOpalCheck(fa)
	case "verbs":
		return cmdOpalVerbs(fa)
	case "functions":
		return cmdOpalFunctions(fa)
	case "validate-ingest":
		return cmdOpalValidateIngest(fa)
	default:
		return ObserveError{Msg: fmt.Sprintf("unknown opal subcommand %q; expected check, verbs, functions, or validate-ingest", fa.args[1])}
	}
}

// gqlCheckQueries validates an OPAL pipeline via the checkQueries GraphQL operation.
// Errors with empty text mean the pipeline requires an input dataset; they are suppressed.
var gqlCheckQueries = compileGqlQuery(
	`query CheckQueries($queries: MultiStageQueryInput!) {
		checkQueries(queries: $queries) {
			parsedPipeline {
				errors { col row text }
				warnings { kind symbol { col row } }
			}
			resultSchema { fieldList { name } }
		}
	}`,
	"data", "checkQueries", "0",
)

// opalPos holds row:col position info from the API response.
type opalPos struct {
	row string
	col string
}

func extractPos(sym any) opalPos {
	p := opalPos{}
	if m, ok := sym.(object); ok {
		if v, ok := m["row"].(string); ok {
			p.row = v
		}
		if v, ok := m["col"].(string); ok {
			p.col = v
		}
	}
	return p
}

func cmdOpalCheck(fa FuncArgs) error {
	var pipeline string

	if flagOpalFile != "" {
		data, err := os.ReadFile(flagOpalFile)
		if err != nil {
			return fmt.Errorf("opal check: could not read file %q: %w", flagOpalFile, err)
		}
		pipeline = string(data)
	} else if len(fa.args) >= 3 {
		pipeline = fa.args[2]
	} else {
		return ObserveError{Msg: "usage: observe opal check <pipeline> | observe opal check --file <path>"}
	}

	stage := object{
		"stageID":  "stage-1",
		"pipeline": pipeline,
		"input":    array{},
	}
	queries := object{
		"outputStage": "stage-1",
		"stages":      array{stage},
	}

	result, err := gqlCheckQueries.query(fa.cfg, fa.op, fa.hc, object{"queries": queries})
	if err != nil {
		return err
	}

	res, ok := result.(object)
	if !ok {
		return fmt.Errorf("opal check: unexpected response type")
	}

	// Extract parsedPipeline
	pp, _ := res["parsedPipeline"].(object)
	var errors []object
	var warnings []object

	if pp != nil {
		if errList, ok := pp["errors"].(array); ok {
			for _, e := range errList {
				if m, ok := e.(object); ok {
					// Skip errors with empty text — these indicate "compilation requires an input
					// dataset" and are not real syntax errors in the pipeline itself.
					if text, _ := m["text"].(string); text != "" {
						errors = append(errors, m)
					}
				}
			}
		}
		if warnList, ok := pp["warnings"].(array); ok {
			for _, w := range warnList {
				if m, ok := w.(object); ok {
					warnings = append(warnings, m)
				}
			}
		}
	}

	if len(errors) > 0 {
		for _, e := range errors {
			text, _ := e["text"].(string)
			row, _ := e["row"].(string)
			col, _ := e["col"].(string)
			fmt.Fprintf(fa.op, "ERROR %s:%s: %s\n", row, col, text)
		}
		return ObserveError{Msg: "opal check: pipeline has errors"}
	}

	for _, w := range warnings {
		kind, _ := w["kind"].(string)
		sym, _ := w["symbol"].(object)
		pos := extractPos(sym)
		fmt.Fprintf(fa.op, "WARN %s %s:%s\n", kind, pos.row, pos.col)
	}

	// Print OK and optionally the result schema fields
	fmt.Fprintf(fa.op, "OK\n")
	if schema, ok := res["resultSchema"].(object); ok {
		if fields, ok := schema["fieldList"].(array); ok {
			for _, f := range fields {
				if fm, ok := f.(object); ok {
					name, _ := fm["name"].(string)
					fmt.Fprintf(fa.op, "  %s\n", name)
				}
			}
		}
	}

	return nil
}
