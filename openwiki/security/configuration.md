---
type: Security Guide
title: Configuration and Security Boundaries
description: Strict runtime configuration, secret resolution, HTTP admission, appliance TLS policy, and trust boundaries.
tags: [security, configuration, http]
openwiki:
  roles: [security, operations]
  change_kinds: [configuration, admission-control]
  source_paths: [internal/config/config.go, internal/jetkvm/manager.go, internal/mcpserver/transport.go, internal/httporigin]
  symbols: [config.Load, ValidateLimits, trustedHostAndOrigin, requireBearer]
  test_paths: [internal/config/config_test.go, internal/config/privacy_test.go, internal/mcpserver/origin_test.go]
  validation_commands: [go test ./internal/config ./internal/mcpserver]
---

# Configuration and Security Boundaries

`internal/config/config.go` loads exactly one strict YAML document up to 1 MiB. Unknown YAML fields, no devices, invalid limits, malformed aliases, unsupported URL forms, invalid origin policy, and unresolved required environment variables fail without echoing private values. `config.example.yaml` is the safe shape reference.

## Configuration model

Each `devices.<alias>` supplies an HTTP(S) appliance `url`, optional `password_env`, `insecure_skip_verify`, absolute `media_directory`, exact `media_url_allowed_origins`, and named `wake_on_lan` targets. URLs may retain a path prefix but may not contain credentials, query, or fragment. Alias and target names normalize by trimming and must be unique after normalization. `password_env` and `http.bearer_token_env` name valid non-empty environment values; credential bytes never belong in YAML.

`limits` defaults to 16 operations, 4 per device, 8 connection attempts, 2 captures, 2 decoders, and a 60-second `session_idle_timeout`. Operations, per-device operations, captures, and decoders must be 1–1024; connection attempts must be 1–64; idle timeout must be 10 seconds through one hour. Per-device operations, connection attempts, and captures cannot exceed operations. Connection attempts bound only new resident-generation setup; healthy resident sessions do not consume them. See [sessions](../runtime/sessions-and-device-protocol.md) for scheduling semantics.

Validate offline, without FFmpeg, listener, or device session:

```sh
jetkvm-mcp config validate --config config.yaml
```

`main_test.go` proves that this mode is offline and that invalid config errors do not expose config paths or private URL values. `config_test.go` and `privacy_test.go` own parsing/privacy detail.

## HTTP exposure policy

A non-loopback `--http HOST:PORT` requires a non-empty resolved bearer token. `http.allowed_origins` is a public Host/origin admission list, not a CORS list: each entry is an exact HTTP(S) origin with no credentials/path/query/fragment/wildcard. Loopback and `localhost` Hosts need no list. For public requests, Host must correspond to an allowed origin; browser `Origin` must be a single exact same-origin value. Invalid origins are forbidden before bearer authentication. Native clients can omit `Origin`, but public Host is still checked. The backend ignores `Forwarded` and `X-Forwarded-*`.

The bearer parser accepts one `Authorization: Bearer` token68 credential and uses constant-time digest comparison. No OAuth and no CORS response headers are implemented. Put remote deployment behind a TLS reverse proxy that preserves public Host, forwards Authorization without logging it, limits bodies to 1 MiB, and keeps backend access protected.

## Appliance trust and data privacy

Device HTTP clients explicitly disable proxy use, use TLS 1.2 or newer, retain cookies in a per-session jar, and do not follow redirects. `insecure_skip_verify` is an explicit per-device opt-in for local appliances, not a general recommendation. Screenshots, typed text, key intent, paths, media URLs, status, and raw diagnostic output are private operational data. Public MCP results redact virtual-media sources; stderr telemetry uses a closed vocabulary rather than payload fields. See [operations](../operations/runbook.md) and [virtual media](../runtime/virtual-media.md).
