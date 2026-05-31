# monitor

    observe monitor preview-query <file.json>
    observe monitor preview [--workspace <id>] <file.json>
    observe monitor alarms [--workspace <id>] [--monitor-id <id>] [--since <duration>] [--level <level>]

The monitor command provides access to Monitor V2 resources in your Observe
workspace. Use `observe list monitor` and `observe get monitor` to browse
existing monitors, and the subcommands below for advanced operations.

Monitor V2 is Observe's rule-based alerting engine. Each monitor watches an
OPAL pipeline and fires alarms when conditions are met (e.g. a count exceeds a
threshold). Alarms have levels: critical, error, warning, or informational.

## Listing and Getting Monitors

    observe list monitor [<name-substring>]
    observe get monitor <id>

`observe list monitor` queries `searchMonitorV2` and displays a table with
columns: id, name, disabled, updatedDate. Pass an optional substring to filter
by name. Use `--workspace <id>` to target a non-default workspace.

`observe get monitor <id>` retrieves a single monitor by ID using `monitorV2`
and prints it in YAML format, including the definition block.

## Subcommands

### preview-query

    observe monitor preview-query <file.json>

Reads a `MonitorV2Input` JSON file and calls `evaluateMonitorV2Source` to
compile the monitor definition into its OPAL pipeline representation. Prints
the generated pipeline and the result schema (field names and types). Useful
for validating a monitor definition before creating it.

The JSON file must be a valid `MonitorV2Input` object with at minimum:
- `name` (string)
- `ruleKind` (MonitorV2RuleKind enum)
- `definition` (MonitorV2Definition union type)

### preview

    observe monitor preview [--workspace <id>] <file.json>

Reads a `MonitorV2Input` JSON file and calls `previewMonitorV2` against the
workspace to evaluate whether the monitor would currently fire. Prints:
- "Would fire: true/false"
- Any sample alarm groupings with their level and timestamp

Use `--workspace <id>` or the global `--workspace` flag to specify the target
workspace. Defaults to workspace `42379913`.

### alarms

    observe monitor alarms [--workspace <id>] [--monitor-id <id>] [--since <duration>] [--level <level>]

Searches for Monitor V2 alarms using `searchMonitorV2Alarms` and displays
them in a table with columns: id, monitorId, level, status, startTime, endTime.

Options:
- `--monitor-id <id>`: filter alarms to a specific monitor
- `--since <duration>`: look back this far (default: 24h); accepts Go duration
  strings such as `1h`, `24h`, `7d` (interpreted as hours/minutes/seconds)
- `--level <level>`: filter by alarm level (critical, error, warning, informational)

Use the global `--workspace` flag to target a non-default workspace.

## Examples

List all monitors in the default workspace:

    observe list monitor

List monitors whose name contains "CPU":

    observe list monitor CPU

Get a specific monitor:

    observe get monitor 41234567

Evaluate a monitor definition:

    observe monitor preview-query my-monitor.json

Preview whether a monitor would fire now:

    observe monitor preview my-monitor.json

Preview in a specific workspace:

    observe --workspace 42379913 monitor preview my-monitor.json

Show alarms from the last 7 days:

    observe monitor alarms --since 168h

Show only critical alarms for a specific monitor:

    observe monitor alarms --monitor-id 41234567 --level critical

## MonitorV2Input JSON Format

    {
      "name": "My CPU Monitor",
      "ruleKind": "COUNT",
      "definition": {
        "compareFunction": "GREATER",
        "countAggFunction": "COUNT",
        "threshold": 90
      }
    }
