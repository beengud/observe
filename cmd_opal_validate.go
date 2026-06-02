package main

import (
	"fmt"
)

// gqlValidateIngestFilter validates an OPAL ingest filter expression against a dataset.
// The --dataset flag is registered on the parent opal FlagSet as flagOpalDataset (cmd_opal.go).
//
// Actual API schema: validateIngestFilterExpression returns [TaskResultError!] where
// TaskResultError only has { message } (no severity or symbol fields).
// A null result means the pipeline is valid; a non-empty array means errors.
var gqlValidateIngestFilter = compileGqlQuery(
	`query ValidateIngestFilter($pipeline: String!, $sourceDatasetID: ObjectId!) {
		validateIngestFilterExpression(pipeline: $pipeline, sourceDatasetID: $sourceDatasetID) {
			message
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

	// A null result means valid. A non-empty array means errors.
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
			fmt.Fprintf(fa.op, "ERROR: %s\n", msg)
			hasErrors = true
		}
	}

	if hasErrors {
		return ObserveError{Msg: "opal validate-ingest: pipeline has errors"}
	}
	fmt.Fprintf(fa.op, "OK\n")
	return nil
}
