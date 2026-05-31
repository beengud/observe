//go:build integration

package main

// Integration tests for opal commands against the live Observe tenant.
// Run with: go test -tags integration ./...
//
// These tests require the following environment variables (with fallback to
// hardcoded CI values):
//   OBSERVE_CUSTOMERID  — defaults to 109601619518
//   OBSERVE_AUTHTOKEN   — defaults to fNJn-aQOmgOUeIvosyQBLjRiNBVBBSZz
//   OBSERVE_SITE        — defaults to 109601619518.observeinc.com
//
// The following test cases are documented here and fleshed out in issue #9:
//
// TestOpalCheckGoodPipeline
//   Calls checkQueries with "filter true" — expects OK output, no errors, non-nil
//   resultSchema. Verifies the API round-trip and schema field output format.
//
// TestOpalCheckBadPipeline
//   Calls checkQueries with "not_a_verb 123" — expects ERROR output with line/column
//   position info from the API, and exit code 1.
//
// TestOpalCheckEmptyPipeline
//   Calls checkQueries with an empty string — API may return an error or
//   succeed with a schema; either way, the CLI should not panic.
//
// TestOpalVerbsReturnsResults
//   Calls verbsAndFunctions and verifies the verbs list is non-empty.
//   Spot-checks that "filter" and "limit" appear in the output.
//
// TestOpalFunctionsReturnsResults
//   Calls verbsAndFunctions and verifies the functions list is non-empty.
//   Spot-checks that common functions like "count" appear in the output.
//
// TestOpalValidateIngestGoodPipeline
//   Calls validateIngestFilterExpression against dataset 42918275 with
//   "filter true" — expects OK output (no errors).
//
// TestOpalValidateIngestBadPipeline
//   Calls validateIngestFilterExpression against dataset 42918275 with an
//   invalid expression — expects ERROR output and exit code 1.
