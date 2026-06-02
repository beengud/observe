package main

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
)

var ErrDatasetUsage = ObserveError{Msg: "usage: observe dataset <dry-run|impact> [args...]"}

func init() {
	RegisterCommand(&Command{
		Name: "dataset",
		Help: "Perform dataset pipeline dry-run and impact analysis.",
		Func: cmdDataset,
	})
}

func cmdDataset(fa FuncArgs) error {
	if len(fa.args) < 2 {
		return ErrDatasetUsage
	}
	switch fa.args[1] {
	case "dry-run":
		return cmdDatasetDryRun(fa)
	case "impact":
		return cmdDatasetImpact(fa)
	default:
		return ObserveError{Msg: fmt.Sprintf("unknown dataset subcommand %q; expected dry-run or impact", fa.args[1])}
	}
}

// datasetCmdInput is the shape of the JSON file accepted by `observe dataset dry-run` and
// `observe dataset impact`.
type datasetCmdInput struct {
	WorkspaceId string         `json:"workspaceId"`
	Dataset     map[string]any `json:"dataset"`
	Query       map[string]any `json:"query"`
}

func readDatasetInput(fa FuncArgs, filePath string) (*datasetCmdInput, error) {
	data, err := fa.fs.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("dataset: could not read file %q: %w", filePath, err)
	}
	var input datasetCmdInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("dataset: could not parse JSON from %q: %w", filePath, err)
	}
	return &input, nil
}

var gqlSaveDatasetDryRun = compileGqlQuery(
	`mutation SaveDatasetDryRun($workspaceId: ObjectId!, $dataset: DatasetInput!, $query: MultiStageQueryInput!) {
		saveDatasetDryRun(workspaceId: $workspaceId, dataset: $dataset, query: $query) {
			dataset { id name }
			dematerializedDatasets { id name }
			errorDatasets { dataset { id name } errorText }
		}
	}`,
	"data", "saveDatasetDryRun",
)

func cmdDatasetDryRun(fa FuncArgs) error {
	if len(fa.args) != 3 {
		return ObserveError{Msg: "usage: observe dataset dry-run <file.json>"}
	}
	input, err := readDatasetInput(fa, fa.args[2])
	if err != nil {
		return err
	}

	obj, err := gqlSaveDatasetDryRun.query(fa.cfg, fa.op, fa.hc, object{
		"workspaceId": input.WorkspaceId,
		"dataset":     input.Dataset,
		"query":       input.Query,
	})
	if err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("dataset dry-run: no result returned")
	}
	result, ok := obj.(object)
	if !ok {
		return fmt.Errorf("dataset dry-run: unexpected response type")
	}

	// Print main dataset result
	if ds, ok := result["dataset"].(object); ok {
		name, _ := ds["name"].(string)
		id, _ := ds["id"].(string)
		fmt.Fprintf(fa.op, "Dataset: %s (%s)\n", name, id)
	}

	// Print dematerialized datasets
	if demats, ok := result["dematerializedDatasets"].(array); ok {
		for _, item := range demats {
			if ds, ok := item.(object); ok {
				name, _ := ds["name"].(string)
				id, _ := ds["id"].(string)
				fmt.Fprintf(fa.op, "Would rematerialize: %s (%s)\n", name, id)
			}
		}
	}

	// Print error datasets and track if any errors exist
	var hasErrors bool
	if errs, ok := result["errorDatasets"].(array); ok {
		for _, item := range errs {
			if errDs, ok := item.(object); ok {
				errorText, _ := errDs["errorText"].(string)
				var dsName string
				if ds, ok := errDs["dataset"].(object); ok {
					dsName, _ = ds["name"].(string)
				}
				fmt.Fprintf(fa.op, "Error in %s: %s\n", dsName, errorText)
				hasErrors = true
			}
		}
	}

	if hasErrors {
		fa.op.Exit(1)
	}
	return nil
}

var gqlGetDatasetsAffectedByUpdate = compileGqlQuery(
	`query DatasetsAffectedByUpdate($workspaceId: ObjectId!, $dataset: DatasetInput!, $query: MultiStageQueryInput) {
		getDatasetsAffectedByDatasetUpdate(workspaceId: $workspaceId, dataset: $dataset, query: $query) {
			affectedDatasets { dataset { id name } dependencyType }
			errorDatasets { dataset { id name } errorText }
		}
	}`,
	"data", "getDatasetsAffectedByDatasetUpdate",
)

func cmdDatasetImpact(fa FuncArgs) error {
	if len(fa.args) != 3 {
		return ObserveError{Msg: "usage: observe dataset impact <file.json>"}
	}
	input, err := readDatasetInput(fa, fa.args[2])
	if err != nil {
		return err
	}

	obj, err := gqlGetDatasetsAffectedByUpdate.query(fa.cfg, fa.op, fa.hc, object{
		"workspaceId": input.WorkspaceId,
		"dataset":     input.Dataset,
		"query":       input.Query,
	})
	if err != nil {
		return err
	}
	if obj == nil {
		return fmt.Errorf("dataset impact: no result returned")
	}
	result, ok := obj.(object)
	if !ok {
		return fmt.Errorf("dataset impact: unexpected response type")
	}

	// Print affected datasets as a table
	tw := tabwriter.NewWriter(fa.op, 0, 0, 1, ' ', 0)
	fmt.Fprintln(tw, "name\tid\tdependencyType")
	if affected, ok := result["affectedDatasets"].(array); ok {
		for _, item := range affected {
			if aff, ok := item.(object); ok {
				depType, _ := aff["dependencyType"].(string)
				var dsName, dsId string
				if ds, ok := aff["dataset"].(object); ok {
					dsName, _ = ds["name"].(string)
					dsId, _ = ds["id"].(string)
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\n", dsName, dsId, depType)
			}
		}
	}
	tw.Flush()

	// Print error datasets via Error() so they appear in test captures
	if errs, ok := result["errorDatasets"].(array); ok {
		for _, item := range errs {
			if errDs, ok := item.(object); ok {
				errorText, _ := errDs["errorText"].(string)
				var dsName string
				if ds, ok := errDs["dataset"].(object); ok {
					dsName, _ = ds["name"].(string)
				}
				fa.op.Error("Error in %s: %s\n", dsName, errorText)
			}
		}
	}

	return nil
}
