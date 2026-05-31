package main

import (
	"fmt"
)

// gqlValidateIngestFilter validates an OPAL ingest filter expression against a dataset.
// The --dataset flag is registered on the parent opal FlagSet as flagOpalDataset (cmd_opal.go).
var gqlValidateIngestFilter = compileGqlQuery(
	`query ValidateIngestFilter($pipeline: String!, $sourceDatasetID: ObjectId!) {
		validateIngestFilterExpression(pipeline: $pipeline, sourceDatasetID: $sourceDatasetID) {
			message
			severity
			symbol { offset line column length }
		}
	}`,
	"data", "validateIngestFilterExpression",
)

func cmdOpalValidateIngest(fa FuncArgs) error {
	if flagOpalDataset == "" {
		return ObserveError{Msg: "usage: observe opal validate-ingest --dataset <dataset-id> <pipeline>"}
	}
	if len(fa.args) < 3 {
		return ObserveError{Msg: "usage: observe opal validate-ingest --dataset <dataset-id> <pipeline>"}
	}
	pipeline := fa.args[2]

	result, err := gqlValidateIngestFilter.query(fa.cfg, fa.op, fa.hc, object{
		"pipeline":        pipeline,
		"sourceDatasetID": flagOpalDataset,
	})
	if err != nil {
		return err
	}

	// Result is an array of diagnostic messages (may be null/empty for success)
	var hasErrors bool
	if result != nil {
		diags, ok := result.(array)
		if !ok {
			// Single object case
			if m, ok := result.(object); ok {
				diags = array{m}
			}
		}
		for _, d := range diags {
			m, ok := d.(object)
			if !ok {
				continue
			}
			msg, _ := m["message"].(string)
			severity, _ := m["severity"].(string)
			sym := extractSymbol(m["symbol"])
			if severity == "error" || severity == "ERROR" {
				fmt.Fprintf(fa.op, "ERROR %d:%d: %s\n", sym.line, sym.column, msg)
				hasErrors = true
			} else {
				fmt.Fprintf(fa.op, "WARN %s: %s\n", severity, msg)
			}
		}
	}

	if hasErrors {
		return ObserveError{Msg: "opal validate-ingest: pipeline has errors"}
	}
	fmt.Fprintf(fa.op, "OK\n")
	return nil
}
