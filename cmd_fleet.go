package main

import (
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/spf13/pflag"
)

var (
	flagsFleet         *pflag.FlagSet
	flagFleetWindow    time.Duration
)

const fleetDataset = "Default.Observe Agent/Events"

const opalFleetStatus = `filter kind = "AgentLifecycleEvent" | make_col host:string(identifiers["host.name"]), env:string(identifiers["observe.agent.environment"]), version:string(facets["observe.agent.version"]), instance_id:string(identifiers["observe.agent.instance.id"]), data_obj:parse_json(data) | make_col auth_ok:bool(data_obj.authCheck.passed) | pick_col valid_from, host, env, version, auth_ok, instance_id | sort desc(valid_from)`

const opalFleetVersions = `filter kind = "AgentLifecycleEvent" | make_col host:string(identifiers["host.name"]), env:string(identifiers["observe.agent.environment"]), version:string(facets["observe.agent.version"]) | pick_col valid_from, host, env, version | sort asc(version), asc(host)`

const opalFleetAuth = `filter kind = "AgentLifecycleEvent" | make_col host:string(identifiers["host.name"]), env:string(identifiers["observe.agent.environment"]), version:string(facets["observe.agent.version"]), data_obj:parse_json(data) | make_col auth_ok:bool(data_obj.authCheck.passed), auth_code:int64(data_obj.authCheck.responseCode), auth_url:string(data_obj.authCheck.url) | pick_col valid_from, host, env, version, auth_ok, auth_code, auth_url | sort asc(auth_ok), desc(valid_from)`

func opalFleetHost(hostname string) string {
	return fmt.Sprintf(
		`filter kind = "AgentLifecycleEvent" | filter string(identifiers["host.name"]) = %q | make_col host:string(identifiers["host.name"]), env:string(identifiers["observe.agent.environment"]), version:string(facets["observe.agent.version"]), data_obj:parse_json(data) | make_col auth_ok:bool(data_obj.authCheck.passed), start_time:from_nanoseconds(int64(data_obj.agentStartTime)*1000000000) | pick_col valid_from, host, env, version, auth_ok, start_time | sort desc(valid_from)`,
		hostname,
	)
}

func init() {
	flagsFleet = pflag.NewFlagSet("fleet", pflag.ContinueOnError)
	flagsFleet.DurationVar(&flagFleetWindow, "window", 20*time.Minute, "time window for the query (e.g. 20m, 24h, 168h)")
	RegisterCommand(&Command{
		Name:  "fleet",
		Help:  "Query fleet status of observe-agent instances from Default.Observe Agent/Events.",
		Flags: flagsFleet,
		Func:  cmdFleet,
	})
}

var (
	ErrFleetUsage = ObserveError{Msg: "usage: observe fleet <status|host <hostname>|versions|auth> [--window <duration>]"}
)

func cmdFleet(fa FuncArgs) error {
	// fa.args[0] is "fleet", fa.args[1] (if present) is the subcommand
	if len(fa.args) < 2 {
		return ErrFleetUsage
	}

	subcommand := fa.args[1]

	var opalText string
	var window time.Duration

	if flagFleetWindow > 0 {
		window = flagFleetWindow
	} else {
		window = 20 * time.Minute
	}

	switch subcommand {
	case "status":
		opalText = opalFleetStatus
	case "host":
		if len(fa.args) < 3 {
			return ObserveError{Msg: "usage: observe fleet host <hostname> [--window <duration>]"}
		}
		hostname := fa.args[2]
		opalText = opalFleetHost(hostname)
	case "versions":
		opalText = opalFleetVersions
	case "auth":
		opalText = opalFleetAuth
	default:
		return NewObserveError(nil, "unknown fleet subcommand %q; use status, host, versions, or auth", subcommand)
	}

	return runFleetQuery(fa, opalText, window)
}

func runFleetQuery(fa FuncArgs, opalText string, window time.Duration) error {
	nowTime := time.Now().Truncate(time.Second)
	toTime := nowTime.Add(-15 * time.Second).Truncate(time.Minute)
	fromTime := toTime.Add(-window)

	datasetPath := fleetDataset
	noLinkify := false
	req := V1ExportQueryRequest{
		Query: OpalQuery{
			OutputStage: "query",
			Stages: []StageQuery{
				{
					Inputs: []StageQueryInput{
						{
							InputName:   "_",
							DatasetPath: &datasetPath,
						},
					},
					StageID:  "query",
					Pipeline: opalText,
				},
			},
		},
		Presentation: &Presentation{
			Linkify: &noLinkify,
		},
	}

	tfmt := &CSVParsingColumnFormatter{
		ColumnFormatter: ColumnFormatter{
			Output:   fa.op,
			ColWidth: 64,
		},
	}
	defer tfmt.Close()
	var output io.Writer = tfmt

	uri := fmt.Sprintf("/v1/meta/export/query?startTime=%s&endTime=%s",
		url.QueryEscape(fromTime.Format(time.RFC3339)),
		url.QueryEscape(toTime.Format(time.RFC3339)))

	err, _ := RequestPOSTWithBodyOutput(fa.cfg, fa.op, fa.hc, uri, &req,
		headers("Accept", "text/csv", "Authorization", fa.cfg.AuthHeader()),
		output)
	return err
}
