---
type: Operations Runbook
title: Operations Runbook
description: Safe launch, liveness, telemetry, diagnostics, shutdown, and incident response for JetKVM MCP.
tags: [operations, telemetry, deployment]
openwiki:
  roles: [operations, runtime]
  change_kinds: [shutdown, telemetry]
  source_paths: [cmd/jetkvm-mcp/main.go, internal/jetkvm/device_owner.go, internal/telemetry/recorder.go]
  symbols: [run, shutdownBudget, Manager.Close, deviceOwner.Close, telemetry.New]
  test_paths: [cmd/jetkvm-mcp/main_test.go, internal/jetkvm/device_owner_test.go, internal/telemetry/recorder_test.go]
  validation_commands: [go test ./cmd/jetkvm-mcp ./internal/jetkvm ./internal/telemetry]
---

# Operations Runbook

## Bring-up

1. Copy `config.example.yaml`, reference secrets only through environment variable names, and protect config/environment access.
2. Validate before serving: `jetkvm-mcp config validate --config config.yaml`.
3. For a local MCP client, run `jetkvm-mcp --config config.yaml`; stdout is protocol-only and logs/telemetry go to stderr.
4. For local HTTP, run `jetkvm-mcp --config config.yaml --http 127.0.0.1:8080`; clients use `/mcp`, while `GET /healthz` returns `ok\n`.

A healthy endpoint proves only process admission and prior config/FFmpeg initialization. It does not prove an appliance is reachable, a host is powered, video exists, or a media action completed. Use `jetkvm_list_devices` then `jetkvm_get_status` for device observation.

## Failure response

Tool errors are JSON with `version`, `code`, `message`, `outcome`, and `retryable`. Codes are `operation_failed`, `canceled`, `timeout`, `invalid_input`, `busy`, `session_released`, `session_taken_over`, `ownership_uncertain`, `authentication_failed`, `device_unavailable`, `video_unavailable`, `no_signal`, and `protocol_error`; outcomes are `not_sent`, `failed`, and `unknown`. `not_sent` identifies validation/admission/pre-session rejection. Read failures with `failed` can be retried only when `retryable: true`; cancellation, timeout, device unavailability, and no signal are the read classes that can receive that flag after dispatch. For a mutation, `unknown` means the appliance or host may already have received an effect: do not replay it. Check status, host state, and virtual-media status first. Capacity rejection is immediate `busy/not_sent`; reduce concurrency instead of waiting for a server queue.

`jetkvm_release_session` returns only after local resources are released and remains sticky. To resume this server's ordinary device work after release or an externally recognized takeover, use `jetkvm_take_over_session` deliberately: it can displace another authenticated operator, is non-idempotent, and must not be used to replay prior work. `session_released` and `session_taken_over` are non-retryable ownership states; `ownership_uncertain` is not proof of release or authority. See [session lifecycle](../runtime/sessions-and-device-protocol.md).

`debug rpc` is intentionally outside MCP. `ping`, `getLocalVersion`, and `getActiveExtension` are safe-by-default; every other method needs `--unsafe-acknowledge-risk` and may mutate or reveal raw firmware data. It writes raw result only to its explicit stdout, so avoid secrets in parameters and protect command history.

## Shutdown and deployment

SIGINT/SIGTERM cancels stdio and HTTP work. HTTP stops accepting new work and drains for up to five seconds, then force-closes. That same shutdown budget is passed to `Manager.Close`, which stops every managed device owner and closes resident session generations. Give a supervisor longer than five seconds plus overhead; a clean listener drain is not proof that appliance sessions have already closed. Run as a dedicated unprivileged identity; restrict config and environment, disable core dumps where practical, mount media read-only when possible, and impose CPU/memory/PID/file/network limits outside the program.

For remote access, prefer loopback plus a TLS reverse proxy. Preserve external Host; do not rely on forwarding headers; protect Authorization from logs; stop routing new requests before SIGTERM; allow proxy connections to outlive the backend drain. Container specifics are in [release and container delivery](../delivery/release-and-container.md).

## Telemetry

`telemetry.New` emits bounded NDJSON to stderr using schema `jetkvm.operation.v2`. Events have random process/correlation IDs and closed enums for transport, operation, stage, code, outcome, and duration. Stages include admission, connect, auth, signaling, RPC, capture, cleanup, tool, and shutdown. The recorder uses bounded queues; on close it writes a telemetry summary with dropped-event and writer-failure status. Missing telemetry does not prove an operation did not execute. Treat stderr as access-controlled private operational data; repository guidance retains routine data at most 14 days and preserves only a relevant sanitized incident window. `internal/telemetry/recorder_test.go` is the focused proof; `docs/telemetry.md` gives the full schema/privacy policy.
