# fleet

    observe fleet status --window 20m
    observe fleet host my-server.example.com --window 24h
    observe fleet versions --window 168h
    observe fleet auth --window 20m

The fleet command queries the `Default.Observe Agent/Events` resource dataset to
give you visibility into your deployed observe-agent instances. This dataset
receives heartbeat events approximately every 10 minutes, so a `--window` of
`20m` is sufficient to see every currently active agent.

## Subcommands

### status

    observe fleet status [--window <duration>]

Shows the current agent inventory: for each recent heartbeat, the host name,
platform environment (windows, linux, macos, kubernetes, docker), agent version,
auth check result, and agent instance ID. Results are sorted newest first.

    observe fleet status --window 20m

### host

    observe fleet host <hostname> [--window <duration>]

Shows the event history for a single host over the given window. In addition to
the status columns, this shows the agent start time so you can track restarts.
Results are sorted newest first.

    observe fleet host my-server.example.com --window 24h

### versions

    observe fleet versions [--window <duration>]

Shows the version distribution across your fleet, sorted by version and then
host name. Use this to identify hosts that are running outdated agent versions.

    observe fleet versions --window 168h

### auth

    observe fleet auth [--window <duration>]

Shows auth check status from all agents. Failures are sorted first so they are
easy to spot. Each row includes the HTTP response code and the auth URL that was
checked, making it straightforward to diagnose authentication problems.

    observe fleet auth --window 20m

## Time Window

The `--window` flag accepts Go duration strings such as `20m`, `1h`, `24h`, and
`168h` (one week). The window is anchored at the current time and extends
backward by the specified duration.

For current agent inventory, `--window 20m` is recommended because agents send
heartbeats every 10 minutes. For historical analysis, use longer windows such
as `--window 24h` or `--window 168h`.

## Dataset

All subcommands query `Default.Observe Agent/Events`, which is a resource
dataset in your Observe workspace. This dataset stores `AgentLifecycleEvent`
records containing host name, environment, agent version, instance ID, agent
start time, and authentication check results.

## Examples

Show all currently active agents:

    observe fleet status --window 20m

Show the last 24 hours of events for a specific host:

    observe fleet host prod-web-01.example.com --window 24h

Show version distribution across the whole fleet for the past week:

    observe fleet versions --window 168h

Find agents with auth failures in the last 20 minutes:

    observe fleet auth --window 20m
