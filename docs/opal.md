# opal

    observe opal check <pipeline>
    observe opal check --file <path>
    observe opal verbs
    observe opal functions
    observe opal validate-ingest --dataset <dataset-id> <pipeline>

The opal command provides tools for validating and inspecting OPAL pipelines
and the functions and verbs available in the OPAL query language.

## Subcommands

### check

    observe opal check <pipeline>
    observe opal check --file <path>

Validates an OPAL pipeline string using the Observe checkQueries API. On error,
prints each error as `ERROR line:col: message` and exits with code 1. On
warnings only, prints each warning as `WARN kind: message` and exits 0. On
success, prints `OK` and the result schema fields.

### verbs

    observe opal verbs

Lists all OPAL verbs available in the tenant, sorted alphabetically, with their
category and description, in tab-separated columns.

### functions

    observe opal functions

Lists all OPAL functions available in the tenant, sorted alphabetically, with
their category, return type, and description, in tab-separated columns.

### validate-ingest

    observe opal validate-ingest --dataset <dataset-id> <pipeline>

Validates an OPAL ingest filter expression against a source dataset using the
Observe validateIngestFilterExpression API. Uses the same error output format
as `opal check`.

## Examples

    observe opal check "filter true"
    observe opal check --file my_pipeline.opal
    observe opal verbs
    observe opal functions
    observe opal validate-ingest --dataset 42918275 "filter FIELDS.severity = 'error'"
