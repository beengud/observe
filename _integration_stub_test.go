// Package main contains integration test stubs that document the planned
// integration test coverage for dashboard and worksheet commands.
// These tests are NOT gated by the integration build tag and serve as
// documentation for what should be validated against a live Observe tenant.
//
// Actual integration tests (with //go:build integration) are in:
//   - cmd_board_integration_test.go
//   - cmd_worksheet_integration_test.go
package main

// IntegrationTestPlan documents the intended integration test coverage:
//
//   Board (dashboard) integration tests (cmd_board_integration_test.go):
//   - List boards in workspace 42379913: expect >= 1 result
//   - Search boards by name "Deployment": expect >= 0 results, no error
//   - dashboardSearch with workspace filter only: verify response shape
//   - set-default: expects a real dataset-id and board-id (skipped if not set)
//   - clear-default: reverses set-default (skipped if not set)
//
//   Worksheet integration tests (cmd_worksheet_integration_test.go):
//   - List worksheets in workspace 42379913: may be empty, verify no error
//   - Create worksheet: save a test worksheet
//   - Get worksheet: verify it exists with the created ID
//   - Delete worksheet: cleanup (always attempted via defer)
//
// Environment variables:
//   OBSERVE_CUSTOMERID  - defaults to 109601619518
//   OBSERVE_AUTHTOKEN   - defaults to hardcoded dev token
//
// Run with: go test -tags integration ./...
const IntegrationTestPlan = "documented above"
