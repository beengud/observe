package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

var (
	flagsOpal     *pflag.FlagSet
	flagOpalFile  string
)

var ErrOpalUsage = ObserveError{Msg: "usage: observe opal <check|verbs|functions|validate-ingest> [args...]"}

func init() {
	flagsOpal = pflag.NewFlagSet("opal", pflag.ContinueOnError)
	flagsOpal.StringVarP(&flagOpalFile, "file", "f", "", "Read OPAL pipeline from file instead of argument")
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

// cmdOpalVerbs and cmdOpalFunctions are implemented in ot_opal.go (issue #6).
// cmdOpalValidateIngest is implemented in cmd_opal_validate.go (issue #7).

// gqlCheckQueries validates an OPAL pipeline using the checkQueries GraphQL operation.
var gqlCheckQueries = compileGqlQuery(
	`query CheckQueries($queries: MultiStageQueryInput!, $params: QueryParams) {
		checkQueries(queries: $queries, params: $params) {
			parsedPipeline {
				errors { message severity symbol { offset line column length } }
				warnings { kind message symbol { offset line column length } }
			}
			resultSchema { fields { name type } }
		}
	}`,
	"data", "checkQueries",
)

// opalSymbol holds position info from the API response.
type opalSymbol struct {
	line   int
	column int
}

func extractSymbol(sym any) opalSymbol {
	s := opalSymbol{}
	if m, ok := sym.(object); ok {
		if v, ok := m["line"]; ok && v != nil {
			if f, ok := v.(float64); ok {
				s.line = int(f)
			}
		}
		if v, ok := m["column"]; ok && v != nil {
			if f, ok := v.(float64); ok {
				s.column = int(f)
			}
		}
	}
	return s
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

	stageQuery := object{
		"stageID":  "stage-1",
		"pipeline": pipeline,
	}
	queries := object{
		"stageQueries": array{stageQuery},
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
					errors = append(errors, m)
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
			msg, _ := e["message"].(string)
			sym := extractSymbol(e["symbol"])
			fmt.Fprintf(fa.op, "ERROR %d:%d: %s\n", sym.line, sym.column, msg)
		}
		return ObserveError{Msg: "opal check: pipeline has errors"}
	}

	for _, w := range warnings {
		msg, _ := w["message"].(string)
		kind, _ := w["kind"].(string)
		fmt.Fprintf(fa.op, "WARN %s: %s\n", kind, msg)
	}

	// Print OK and optionally the result schema fields
	fmt.Fprintf(fa.op, "OK\n")
	if schema, ok := res["resultSchema"].(object); ok {
		if fields, ok := schema["fields"].(array); ok {
			for _, f := range fields {
				if fm, ok := f.(object); ok {
					name, _ := fm["name"].(string)
					typ, _ := fm["type"].(string)
					fmt.Fprintf(fa.op, "  %s\t%s\n", name, typ)
				}
			}
		}
	}

	return nil
}
