---
type: Testing Guide
title: Testing, Protocol Gates, and CI
description: Test layers, focused validation, fuzzing, protocol interoperability gates, and CI evidence.
tags: [testing, ci, protocol]
openwiki:
  roles: [testing, delivery]
  change_kinds: [concurrency, protocol, release]
  source_paths: [Makefile, .github/workflows/ci.yml, internal/jetkvm/device_owner.go, internal/jetkvm/scheduling.go]
  symbols: [deviceOwner, generationScheduler]
  test_paths: [internal/jetkvm/device_owner_test.go, internal/jetkvm/scheduling_test.go, internal/jetkvm/admission_test.go]
  validation_commands: [go test ./internal/jetkvm]
---

# Testing, Protocol Gates, and CI

The repository treats generated MCP output, transport parity, security admission, and release subjects as contracts. Start with the narrowest affected package, then use the project targets below.

| Change | Focused tests | Minimal validation |
| --- | --- | --- |
| Tool/schema/result change | `internal/mcpserver/manifest_contract_test.go`, `input_contract_test.go`, `typed_results_test.go` | `go test ./internal/mcpserver` |
| Local release/takeover authority | `internal/mcpserver/release_session_test.go`, `takeover_session_test.go`, `internal/jetkvm/release_session_test.go`, `takeover_session_test.go`, `session_connector_test.go` | `go test ./internal/mcpserver ./internal/jetkvm -run 'Test(ReleaseSession|TakeOverSession|ManagerReleaseSession|ManagerTakeOverSession|DeviceOwnerReleaseSession|RecognizedTakeover)'` |
| HTTP/version/origin change | `origin_test.go`, `protocol_version_test.go`, `transport_test.go`, `cancellation_test.go` | `go test ./internal/mcpserver ./cmd/jetkvm-mcp` |
| Managed session, scheduling, or device protocol | `device_owner_test.go`, `scheduling_test.go`, `manager_test.go`, `provider_test.go`, `session_protocol_test.go` | `go test ./internal/jetkvm` |
| Admission limits | `admission_test.go` | `go test ./internal/jetkvm -run '^TestManagerAdmission'` |
| Media | `virtual_media_test.go`, `virtual_media_fuzz_test.go` | `go test ./internal/jetkvm` |
| Config/privacy | `internal/config/config_test.go`, `privacy_test.go` | `go test ./internal/config` |
| Build/release policy | `internal/cipolicy`, `internal/releasepolicy` | `go test ./internal/cipolicy ./internal/releasepolicy` |

`make test` runs all Go tests. `make protocol-gates` builds the server then uses `cmd/jetkvm-mcp-protocol-gates` to load reviewed pins, lock-install official tooling, run selected official conformance scenarios, run Inspector against stdio and HTTP, and compare canonical results across transports. It writes sanitized gate summaries/artifacts. Update pins only with a reviewed compatibility decision; `internal/protocolgate` validates pins, npm lock, scenario inventory, and result parsing. For a local release or takeover seam, first run the focused lifecycle suites above, then run `go test ./internal/mcpserver` before updating the manifest fixture: the public tool must resolve and behave consistently over all consumer transports, not merely compile at `Manager.ReleaseSession` or `Manager.TakeOverSession`. Takeover changes also require `session_connector_test.go` coverage for takeover latching, no ordinary reconnect, and pre-terminal dispatch fencing.

Fuzz targets are inventory-controlled by `testdata/fuzz-targets.json` and `internal/fuzzpolicy`. `make fuzz-smoke` runs policy validation and one-second targets; `make fuzz` uses 30 seconds. Relevant fuzzers cover RPC codec, signaling, video, virtual media, and config. Managed-session changes should first run `device_owner_test.go` and `scheduling_test.go`: reuse after readiness ping, shared setup, idle expiry, no reconnect after failed cleanup, stale-generation isolation, cancellable waits, RPC/frame interleaving, and immutable frame handoff are behavior-level invariants. Use a real FFmpeg race lane with `make race` or `make race-coverage` when changing owner, scheduler, decoder, or video concurrency.

## CI matrix

`.github/workflows/ci.yml` runs four required jobs: **Go quality** on Go 1.26.6 (`make ci-quality`, including format, module checks, race coverage, vet, staticcheck, govulncheck, fuzz smoke, and release snapshot), **Minimum Go** on 1.25.13 (`make ci-minimum`), **MCP protocol** (protocol gates with Node 22.22.0), and **Container** (multi-platform smoke/metadata verification). Coverage and protocol artifacts are retained for 30 days. Pull request workflow runs use read-only permissions and PR runs may be canceled by newer runs.

The quality target is broad and potentially expensive; do not substitute it for a focused package test during iteration. Conversely, use it before merging substantial runtime, public contract, security, or release changes. See [release delivery](../delivery/release-and-container.md) for the separate tag policy and [compatibility evidence](../integrations/jetkvm-compatibility.md) for real-appliance validation limits.
