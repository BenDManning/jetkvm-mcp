---
type: Source Map
title: Source Map
description: Intent-oriented map of JetKVM MCP implementation, contracts, operations, and verification assets.
tags: [navigation, architecture, testing]
openwiki:
  roles: [repository, architecture]
  source_paths: [cmd/jetkvm-mcp/main.go, internal/jetkvm/manager.go, internal/jetkvm/device_owner.go]
---

# Source Map

This is a navigation map, not a directory inventory. Follow the linked canonical page before editing a subsystem.

| Area | Primary source | Contract or dependency | Focused tests / command |
| --- | --- | --- | --- |
| CLI and process lifecycle | `cmd/jetkvm-mcp/main.go` | config, telemetry, MCP SDK | `cmd/jetkvm-mcp/main_test.go`, `http_integration_test.go`, `stdio_integration_test.go` |
| MCP registration and outputs | `internal/mcpserver/server.go`, `controls.go` | official Go SDK, JSON Schema; `Device.ReleaseSession` and `Device.TakeOverSession` bridge local authority lifecycle tools | `manifest_contract_test.go`, `release_session_test.go`, `takeover_session_test.go`, `typed_results_test.go`, `input_contract_test.go` |
| HTTP policy / protocol revision | `internal/mcpserver/transport.go`, `protocol_version.go` | Streamable HTTP | `origin_test.go`, `protocol_version_test.go`, `transport_test.go` |
| Appliance manager and ownership | `internal/jetkvm/manager.go`, `device_owner.go`, `provider.go`, `scheduling.go` | `mcpserver.Device`; managed resident generation plus release/takeover authority per device | `release_session_test.go`, `takeover_session_test.go`, `session_connector_test.go`, `device_owner_test.go`, `scheduling_test.go`, `manager_test.go`, `admission_test.go` |
| WebRTC and RPC | `internal/jetkvm/provider.go`, `auth.go`, `signaling.go`, `rpc_session.go` | managed owner connects to JetKVM appliance | `session_protocol_test.go`, `rpc_codec_fuzz_test.go`, `signaling_fuzz_test.go` |
| HID, power, status | `internal/jetkvm/hid.go`, `manager.go` | appliance RPC | `controls_manager_test.go`, `typed_results_test.go` |
| Screenshot | `internal/jetkvm/capture.go`, `video*.go`, `decoder_ffmpeg.go` | Pion WebRTC, FFmpeg | `capture_test.go`, `video_fuzz_test.go`, `decoder_test.go` |
| Virtual media | `internal/jetkvm/virtual_media.go` | appliance storage and HTTP fetch | `virtual_media_test.go`, `virtual_media_fuzz_test.go` |
| Configuration / identifiers / origins | `internal/config`, `internal/identifier`, `internal/httporigin` | YAML and environment | `config_test.go`, `privacy_test.go`, `origin_test.go` |
| Telemetry | `internal/telemetry/recorder.go` | stderr NDJSON | `recorder_test.go`, `mcpserver/telemetry_test.go` |
| Protocol gate command | `cmd/jetkvm-mcp-protocol-gates`, `internal/protocolgate` | checked-in pins and npm lock | `make protocol-gates` |
| Appliance evidence and qualification | `cmd/jetkvm-mcp-upstream-drift`, `cmd/jetkvm-mcp-validate`, `internal/compatibility` | reviewed upstream and ledger | [compatibility](../integrations/jetkvm-compatibility.md) |
| CI/release policy tests | `internal/cipolicy`, `internal/fuzzpolicy`, `internal/releasepolicy` | workflows, Makefile, release manifests | `make ci-quality`, release policy tests |

Checked-in product and operational specifications live under `docs/`; validate claims against the implementation above. `docs/threat-model.md`, `docs/telemetry.md`, `docs/protocol-gates.md`, and `docs/ci-quality.md` are useful detailed supporting evidence.
