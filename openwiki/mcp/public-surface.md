---
type: API Reference
title: Public MCP Tool Surface
description: The fixed JetKVM MCP tools, schemas, result guarantees, annotations, and compatibility contract.
tags: [mcp, api, tools]
openwiki:
  roles: [domain, integration]
  change_kinds: [public-api, schema]
  source_paths: [internal/mcpserver/server.go, internal/mcpserver/controls.go, internal/mcpserver/testdata/tool-manifest.json, internal/jetkvm/manager.go]
  symbols: [NewWithTelemetry, Device, addInputTool, addReadTool, addMutationTool]
  test_paths: [internal/mcpserver/manifest_contract_test.go, internal/mcpserver/input_contract_test.go, internal/mcpserver/typed_results_test.go]
  invariants: [The static manifest and emitted structured output must match declared schemas across in-memory, stdio, and HTTP paths.]
  validation_commands: [go test ./internal/mcpserver]
---

# Public MCP Tool Surface

`mcpserver.NewWithTelemetry` registers a static 17-tool manifest. It advertises only `tools` capability and no tool-list change notifications. `internal/mcpserver/server.go` owns registration; `controls.go` owns control-family input/output schemas and validation. Each handler calls the `mcpserver.Device` boundary, implemented by `jetkvm.Manager`, which dispatches device work through the managed session lifecycle documented in [sessions and device protocol](../runtime/sessions-and-device-protocol.md).

## Tool families

| Family | Tools | Owner / effect |
| --- | --- | --- |
| Discovery and observation | `jetkvm_list_devices`, `jetkvm_get_status`, `jetkvm_capture_screen`, `jetkvm_get_virtual_media_status` | List is config-only; the others open a session. Capture returns a fresh PNG plus metadata. |
| Keyboard and pointer | `jetkvm_keyboard`, `jetkvm_mouse` | HID mutations. Keyboard supports bounded US-ASCII `type_text` or a named/printable `press_key`; mouse supports absolute/relative movement, click, and scroll. |
| Host power | `jetkvm_press_host_power_button`, `jetkvm_force_host_power_off`, `jetkvm_press_host_reset_button`, `jetkvm_turn_host_dc_power_on`, `jetkvm_turn_host_dc_power_off`, `jetkvm_wake_host_usb`, `jetkvm_wake_host_lan` | One physical consequence per named tool. LAN wake accepts only a configured target alias. |
| Virtual media | `jetkvm_mount_virtual_media_url`, `jetkvm_mount_virtual_media_file`, `jetkvm_unmount_virtual_media`, `jetkvm_upload_virtual_media_file` | URL fetch or confined local upload/mount. See [virtual media](../runtime/virtual-media.md). |

## Contract rules

Inputs are decoded against generated JSON Schema before handler invocation. Device and target identifiers are normalized and must have 1–128 Unicode code points. Invalid arguments become a value-free `invalid_input` tool error and are `not_sent`. Successful output is marshalled and validated again against its declared output schema before it is returned as `structuredContent` and a text JSON content block. Capture is the exception: it places its PNG `ImageContent` first, then safe metadata, avoiding a textual copy of image bytes.

`ResultStatusCompleted` means the appliance RPC returned; `observed` identifies an observation result. Neither proves a later physical state. Virtual-media public state deliberately reveals only mounted state, source class (`http`/`storage`), and mode, never URL, path, filename, or raw firmware fields.

Annotations communicate client-facing consequences: reads are read-only/idempotent; mutation tools are non-idempotent and commonly destructive. URL mounting is explicitly open-world because the appliance fetches the configured URL. They do not authorize a caller.

## Registration and schema pipeline

`addReadTool` and `addMutationTool` are thin registrations that call `addInputTool` with a fixed mutation classifier: read tools classify false and mutation tools classify true. `addInputTool` generates a JSON Schema from the Go input type when a tool did not supply one, bounds `device` and `target` properties to 1–128 code points, resolves defaults, and validates the schema before registering the SDK handler. It independently generates/resolves an output schema when absent. At call time it normalizes string device/target values, applies defaults, validates input, calls the handler, then marshals and validates output before populating `StructuredContent` and text content. This prevents a handler or provider from silently emitting a result outside the advertised schema.

The specialized control schemas encode operation-specific shapes: capture bounds are 1–3840 by 1–2160; keyboard has exclusive `type_text` or `press_key` branches and at most four unique modifiers; mouse branches require only the fields for their operation; media source values are bounded to 1–4096 characters. Output schemas constrain fixed action/operation/status values and capture MIME/dimensions. A capability change therefore needs an implementation behind `mcpserver.Device`, a named tool/registration and schemas, the manager dispatch, and a reviewed consumer-facing manifest update.

## Manifest compatibility seam

`internal/mcpserver/testdata/tool-manifest.json` is the compatibility snapshot. `TestToolManifestContract` connects through in-memory, stdio subprocess, and stateless HTTP paths; it demands one exact manifest, sorted tools, schemas and annotations, schema-valid representative results, and tool errors as versioned content rather than JSON-RPC failures. It also checks no MCP session ID is used over HTTP. A deliberate update uses:

```sh
make update-tool-manifest
```

Do not use that command as routine formatting: classify the public compatibility impact first and update implementation, exports/registration, fixture, consumer behavior, and the narrow test together. [Protocol compatibility](transports-and-compatibility.md) describes the wire-version rule.
