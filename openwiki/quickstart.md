---
type: Quickstart
title: JetKVM MCP Wiki Quickstart
description: Practical navigation for safely operating and changing the JetKVM MCP server.
tags: [quickstart, architecture, mcp]
openwiki:
  roles: [repository, workflow]
  source_paths: [cmd/jetkvm-mcp/main.go, internal/mcpserver/server.go, internal/jetkvm/manager.go, internal/jetkvm/device_owner.go]
  symbols: [run, NewWithTelemetry, NewManager, deviceOwner]
  validation_commands: [go test ./internal/jetkvm]
---

# JetKVM MCP Wiki Quickstart

JetKVM MCP is a single-process Go MCP server that controls configured JetKVM appliances. It serves 19 tools over stdio or stateless Streamable HTTP. A `jetkvm.Manager` maintains one resident, authenticated WebRTC session generation per device while demand exists, then closes it after its idle lease. `jetkvm_release_session` yields local ownership; the destructive `jetkvm_take_over_session` validates or acquires this server's authoritative session and can displace another authenticated operator. FFmpeg is required for screenshot capture.

## Start here

1. Read [architecture overview](architecture/overview.md) for composition and ownership.
2. Read [public MCP surface](mcp/public-surface.md) before changing a tool, schema, annotation, or result.
3. Read [configuration and security](security/configuration.md) before deployment or HTTP exposure.
4. Use [operations runbook](operations/runbook.md) for safe startup, diagnosis, and shutdown.

```sh
jetkvm-mcp config validate --config config.yaml
jetkvm-mcp --config config.yaml
# or: jetkvm-mcp --config config.yaml --http 127.0.0.1:8080
```

## Wiki map

- **MCP contract:** [public tools and schemas](mcp/public-surface.md), [transports and protocol compatibility](mcp/transports-and-compatibility.md).
- **Runtime domains:** [session/device protocol](runtime/sessions-and-device-protocol.md), [controls and capture](runtime/control-and-capture.md), [virtual media](runtime/virtual-media.md).
- **Safety and operations:** [configuration](security/configuration.md), [runbook and telemetry](operations/runbook.md).
- **Integration and evidence:** [JetKVM compatibility evidence](integrations/jetkvm-compatibility.md).
- **Engineering evidence:** [testing and CI](quality/testing-and-ci.md), [release and container delivery](delivery/release-and-container.md), [source map](architecture/source-map.md).

## Task routing

| Change area or user intent | Relevant wiki page | Exact source entry points | Important symbols or types | Focused tests | Minimal validation command |
| --- | --- | --- | --- | --- | --- |
| Add/change tool, schema, result, or annotation | [public surface](mcp/public-surface.md) | `internal/mcpserver/server.go`, `internal/mcpserver/controls.go` | `NewWithTelemetry`, `Device`, `addInputTool`, `TakeOverSessionToolName` | `manifest_contract_test.go`, `input_contract_test.go`, `takeover_session_test.go` | `go test ./internal/mcpserver` |
| Change release/takeover authority, resident session, connection, RPC, or scheduling | [sessions](runtime/sessions-and-device-protocol.md) | `internal/jetkvm/manager.go`, `device_owner.go`, `provider.go`, `scheduling.go`, `rpc_session.go` | `Manager.ReleaseSession`, `Manager.TakeOverSession`, `deviceOwner`, `WebRTCConnector`, `generationScheduler`, `rpcSession` | `release_session_test.go`, `takeover_session_test.go`, `session_connector_test.go`, `scheduling_test.go`, `session_test.go` | `go test ./internal/jetkvm` |
| Change MCP revision, stdio, or HTTP behavior | [transports](mcp/transports-and-compatibility.md) | `internal/mcpserver/protocol_version.go`, `internal/mcpserver/transport.go`, `cmd/jetkvm-mcp/main.go` | `SupportedProtocolVersion`, `serveHTTP`, `run` | `protocol_version_test.go`, `origin_test.go`, `transport_test.go` | `go test ./internal/mcpserver ./cmd/jetkvm-mcp` |
| Change power, HID, capture | [control and capture](runtime/control-and-capture.md) | `internal/jetkvm/manager.go`, `hid.go`, `capture.go`, `video*.go` | `Manager`, `CaptureScreen`, `scheduledSession` | `controls_manager_test.go`, `capture_test.go`, `video_test.go` | `go test ./internal/jetkvm ./internal/mcpserver` |
| Change media upload/mount | [virtual media](runtime/virtual-media.md) | `internal/jetkvm/virtual_media.go` | `Manager.VirtualMedia`, `openMediaFile` | `virtual_media_test.go`, `virtual_media_fuzz_test.go` | `go test ./internal/jetkvm` |
| Change config, auth, proxy, or session lease policy | [configuration](security/configuration.md), [sessions](runtime/sessions-and-device-protocol.md) | `internal/config/config.go`, `internal/mcpserver/transport.go`, `internal/jetkvm/manager.go` | `config.Load`, `Limits`, `trustedHostAndOrigin` | `config_test.go`, `privacy_test.go`, `admission_test.go` | `go test ./internal/config ./internal/jetkvm ./internal/mcpserver` |
| Change release, container, or CI | [delivery](delivery/release-and-container.md), [testing](quality/testing-and-ci.md) | `Makefile`, `.github/workflows/ci.yml`, `.github/workflows/release.yml`, `Dockerfile` | `ci-quality`, `container-verify`, `finalize-release` | `internal/cipolicy`, `internal/releasepolicy` | relevant `make` target |
| Assess firmware/upstream effects | [compatibility](integrations/jetkvm-compatibility.md) | `cmd/jetkvm-mcp-upstream-drift`, `cmd/jetkvm-mcp-validate` | `jetkvm-mcp-upstream-drift`, `jetkvm-mcp-validate` | `internal/compatibility` tests | focused drift or authorized validator |

## Validation ladder

Use package tests while iterating. Run `make test` for repository-wide Go tests. Run `make protocol-gates` when the wire contract or SDK behavior changes; run `make ci-quality` before merging broad runtime, security, contract, or release changes. `make ci-quality` includes race coverage, analyzers, fuzz smoke, and snapshot artifacts, so it is not the routine first check.

## Operational rule

For a mutation error with `outcome: unknown`, do **not** blindly retry: it may have reached the appliance/host. Inspect read-only status first. A successful `/healthz` only establishes process liveness, not device readiness.

## Backlog

- **Git history:** `.openwikiignore` restricts this run from inspecting the Git range after the recorded `gitHead` in `/openwiki/.last-update.json`; no history-derived claim is recorded.
- **LangSmith observations:** `/openwiki/.langsmith.json` provides connection metadata only, not a usable runtime-observation corpus, so no runtime-observation claim is recorded.
