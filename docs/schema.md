# schema

    observe schema introspect
    observe schema introspect --type Dashboard

The schema command inspects the live Observe GraphQL API schema. Use it to
discover available types, fields, queries, and mutations — including any new
fields that may have been added to existing types like Dashboard.

## Subcommands

### introspect

    observe schema introspect [--type <TypeName>]

Runs a GraphQL introspection query against the configured Observe tenant and
outputs the result as formatted JSON. By default the full schema is output,
including all types, their fields, input fields, and enum values.

Use --type to filter the output to a single named type. This is useful for
checking the current set of fields on a specific type without parsing the
entire schema.

## Flags

    --type <TypeName>   Filter output to just the named type (e.g. Dashboard,
                        DashboardInput, Mutation, Query).

## Examples

Dump the full schema to a file for offline inspection:

    observe schema introspect > schema.json

Check what fields are available on the Dashboard type:

    observe schema introspect --type Dashboard

Find all mutations related to dashboards using jq:

    observe schema introspect --type Mutation | jq '.fields[] | select(.name | test("(?i)dashboard"))'

List all input fields accepted by saveDashboard:

    observe schema introspect --type DashboardInput
