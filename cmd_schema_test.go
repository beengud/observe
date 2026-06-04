package main

import (
	"strings"
	"testing"
)

const schemaIntrospectResponse = `{"data":{"__schema":{
	"queryType":{"name":"Query"},
	"mutationType":{"name":"Mutation"},
	"subscriptionType":null,
	"types":[
		{"kind":"OBJECT","name":"Dashboard","description":"A dashboard","fields":[{"name":"id","description":null,"isDeprecated":false,"type":{"kind":"SCALAR","name":"String","ofType":null},"args":[]}],"inputFields":null,"enumValues":null,"interfaces":[],"possibleTypes":null},
		{"kind":"OBJECT","name":"Query","description":null,"fields":[],"inputFields":null,"enumValues":null,"interfaces":[],"possibleTypes":null}
	]
}}}`

// TestCmdSchemaIntrospect verifies full schema output.
func TestCmdSchemaIntrospect(t *testing.T) {
	flagSchemaType = ""
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, schemaIntrospectResponse},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"schema", "introspect"}, fix.hc)
	if fix.op.ErrorBuf.String() != "" {
		t.Error("unexpected error output:", fix.op.ErrorBuf.String())
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "queryType") {
		t.Error("expected queryType in output:", out)
	}
	if !strings.Contains(out, "Dashboard") {
		t.Error("expected Dashboard type in output:", out)
	}
}

// TestCmdSchemaIntrospectFilterByType verifies --type filtering narrows to one type.
func TestCmdSchemaIntrospectFilterByType(t *testing.T) {
	flagSchemaType = "Dashboard"
	defer func() { flagSchemaType = "" }()
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, schemaIntrospectResponse},
	)
	RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"schema", "introspect"}, fix.hc)
	if fix.op.ErrorBuf.String() != "" {
		t.Error("unexpected error output:", fix.op.ErrorBuf.String())
	}
	out := fix.op.OutputBuf.String()
	if !strings.Contains(out, "Dashboard") {
		t.Error("expected Dashboard in output:", out)
	}
	if strings.Contains(out, "queryType") {
		t.Error("filtered output should not contain top-level schema keys:", out)
	}
}

// TestCmdSchemaIntrospectFilterTypeNotFound verifies error when --type is not in schema.
func TestCmdSchemaIntrospectFilterTypeNotFound(t *testing.T) {
	flagSchemaType = "NonExistentType"
	defer func() { flagSchemaType = "" }()
	fix := startFixture(t,
		testRequest{"/v1/meta", 200, schemaIntrospectResponse},
	)
	func() {
		defer func() { recover() }()
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"schema", "introspect"}, fix.hc)
	}()
	if !strings.Contains(fix.op.ErrorBuf.String(), "not found") {
		t.Error("expected 'not found' error, got:", fix.op.ErrorBuf.String())
	}
}

// TestCmdSchemaUnknownSubcommand verifies error on unknown subcommand.
func TestCmdSchemaUnknownSubcommand(t *testing.T) {
	flagSchemaType = ""
	fix := startFixture(t)
	func() {
		defer func() { recover() }()
		RunCommandWithConfig(fix.cfg, fix.fs, fix.op, []string{"schema", "bogus"}, fix.hc)
	}()
	if !strings.Contains(fix.op.ErrorBuf.String(), "unknown schema subcommand") {
		t.Error("expected unknown subcommand error, got:", fix.op.ErrorBuf.String())
	}
}
