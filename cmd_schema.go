package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/pflag"
)

var (
	flagsSchema    *pflag.FlagSet
	flagSchemaType string
)

var ErrSchemaUsage = ObserveError{Msg: "usage: observe schema <introspect> [args...]"}

func init() {
	flagsSchema = pflag.NewFlagSet("schema", pflag.ContinueOnError)
	flagsSchema.StringVar(&flagSchemaType, "type", "", "Filter output to a specific type name")
	RegisterCommand(&Command{
		Name:  "schema",
		Help:  "Inspect the Observe GraphQL API schema.",
		Flags: flagsSchema,
		Func:  cmdSchema,
	})
}

func cmdSchema(fa FuncArgs) error {
	if len(fa.args) < 2 {
		return ErrSchemaUsage
	}
	switch fa.args[1] {
	case "introspect":
		return cmdSchemaIntrospect(fa)
	default:
		return ObserveError{Msg: fmt.Sprintf("unknown schema subcommand %q; expected introspect", fa.args[1])}
	}
}

var gqlSchemaIntrospect = compileGqlQuery(
	`query Schema_Introspect {
		__schema {
			queryType { name }
			mutationType { name }
			subscriptionType { name }
			types {
				kind
				name
				description
				fields(includeDeprecated: true) {
					name
					description
					isDeprecated
					type {
						kind
						name
						ofType { kind name ofType { kind name } }
					}
					args {
						name
						description
						type { kind name ofType { kind name } }
					}
				}
				inputFields {
					name
					description
					type {
						kind
						name
						ofType { kind name ofType { kind name } }
					}
				}
				enumValues(includeDeprecated: true) {
					name
					description
					isDeprecated
				}
				interfaces { name kind }
				possibleTypes { name kind }
			}
		}
	}`,
	"data", "__schema",
)

func cmdSchemaIntrospect(fa FuncArgs) error {
	result, err := gqlSchemaIntrospect.query(fa.cfg, fa.op, fa.hc, object{})
	if err != nil {
		return err
	}
	if result == nil {
		return fmt.Errorf("schema introspect: no result returned")
	}
	schema, ok := result.(object)
	if !ok {
		return fmt.Errorf("schema introspect: unexpected response type")
	}

	enc := json.NewEncoder(fa.op)
	enc.SetIndent("", "  ")

	if flagSchemaType != "" {
		types, _ := schema["types"].(array)
		for _, t := range types {
			typeObj, ok := t.(object)
			if !ok {
				continue
			}
			if name, _ := typeObj["name"].(string); name == flagSchemaType {
				return enc.Encode(typeObj)
			}
		}
		return fmt.Errorf("schema introspect: type %q not found", flagSchemaType)
	}

	return enc.Encode(schema)
}
