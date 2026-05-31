# dataset

    observe dataset dry-run <file.json>
    observe dataset impact <file.json>

The dataset command supports pipeline dry-run validation and downstream impact
analysis for datasets in your Observe workspace.

Both subcommands accept the same JSON input file with the following shape:

```json
{
  "workspaceId": "42379913",
  "dataset": { "name": "MyDataset" },
  "query": {
    "stageQueries": [{ "stageID": "main", "pipeline": "filter true" }]
  }
}
```

## Subcommands

### dry-run

    observe dataset dry-run <file.json>

Performs a dry-run of saving a dataset using the given pipeline definition.
No dataset is actually created or modified. The output reports:

- The dataset that would be saved (name and ID).
- Any datasets that would be dematerialized (rematerialized) as a result.
- Any compilation or validation errors.

Exits with status 1 if any error datasets are reported.

### impact

    observe dataset impact <file.json>

Analyzes which downstream datasets would be affected if the given dataset
definition were saved. Output is a table showing each affected dataset's
name, ID, and dependency type. Any error datasets are printed to stderr.

## Example

    observe dataset dry-run my-dataset.json
    observe dataset impact my-dataset.json
