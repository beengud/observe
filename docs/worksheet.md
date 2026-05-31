# worksheet

    observe worksheet list [--name <filter>]
    observe worksheet get <id>
    observe worksheet create <file.json>
    observe worksheet delete <id>

The worksheet command allows you to list, get, create, and delete worksheets
in your Observe workspace. Worksheets are exploratory data analysis documents
that contain named stages with OPAL pipeline expressions.

## Subcommands

### list

    observe worksheet list
    observe worksheet list --name "My Analysis"

Lists worksheets in the current workspace. Use --name to filter results by
name (case-insensitive substring match). The --workspace global flag sets
the workspace ID to search in.

### get

    observe worksheet get <id>

Fetches a single worksheet by ID and outputs it as JSON, including all stages
with their stageID and pipeline fields.

### create

    observe worksheet create worksheet.json

Creates a new worksheet from a JSON file containing a WorksheetInput object.
The JSON file must include at minimum: name, workspaceId, and stages.
The stages array contains objects with stageID and pipeline fields.

Example worksheet.json:

    {
      "name": "My Analysis",
      "workspaceId": "42379913",
      "stages": [
        {
          "stageID": "stage-1",
          "pipeline": "filter env = \"production\" | limit 100"
        }
      ]
    }

### delete

    observe worksheet delete <id>

Deletes the worksheet with the given ID. Outputs a confirmation message on
success, or an error message if the deletion fails.

## Object type

Worksheets can also be managed via the standard object commands:

    observe list worksheet
    observe get worksheet <id>
    observe delete worksheet <id>
