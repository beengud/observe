# board

    observe board create my-dashboard.json
    observe board update 43092236 my-dashboard.json
    observe board scaffold --name "My Dashboard" > my-dashboard.json

The board command allows you to create, update, and scaffold boards (dashboards)
in your Observe workspace. Boards organize visualizations of data using stages
and layouts.

To list, get, or delete boards, use the standard object commands:

    observe list board
    observe get board <id>
    observe delete board <id>

## Subcommands

### create

    observe board create <file.json>

Creates a new board from a JSON file containing a DashboardInput object. The
required fields are: name, workspaceId, and layout. The stages field is
optional and describes the OPAL pipelines backing the visualizations.

### update

    observe board update <id> <file.json>

Updates an existing board identified by its ID. Reads the same JSON format as
create, and sets the id field automatically from the command-line argument.

### scaffold

    observe board scaffold [--name NAME]

Prints a minimal board template JSON to stdout. Redirect it to a file to
customize and then use with create or update. Use --name to set the board name
in the template.

## Example

    observe board scaffold --name "My Dashboard" > my-dashboard.json
    observe board create my-dashboard.json
