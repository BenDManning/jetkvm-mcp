---
type: integration guide
title: JetKVM Compatibility Evidence
description: Evidence-backed appliance integration boundary, upstream drift checks, compatibility ledger claims, and read-only physical validation.
tags: [integration, compatibility, jetkvm]
---

# JetKVM Compatibility Evidence

The appliance protocol is an integration boundary, not an unbounded compatibility claim. Runtime implementation is in [sessions and device protocol](../runtime/sessions-and-device-protocol.md); this page records how its assumptions are reviewed, monitored, and physically checked.

## Reviewed upstream boundary

`docs/compatibility/jetkvm-upstream-surfaces.json` pins a JetKVM upstream commit and six monitored surfaces: **auth**, **signaling**, **RPC**, **video**, **HID**, and **virtual media**. `cmd/jetkvm-mcp-upstream-drift` requires a local checkout of exactly `https://github.com/jetkvm/kvm`, verifies the pin exists, compares that pin to `--target` (default `HEAD`) only across the reviewed paths, and emits a sanitized JSON report. It exits 0 for `no_drift`, 1 for `drift`, and 2 for invalid inputs, unavailable refs/pin, or comparison failure.

```sh
jetkvm-mcp-upstream-drift --upstream /path/to/jetkvm-kvm --target HEAD
```

A drift result is a review trigger, not proof of breakage. Re-inspect the changed surface against provider/auth/signaling/RPC/video/HID/media code, refresh source evidence as appropriate, add tests or implementation changes, and update the reviewed pin/ledger only with a bounded claim. This repository's active ignore policy prevents OpenWiki from inspecting git history, so this documentation does not make history-derived assertions.

## Ledger claim limits

`docs/compatibility/jetkvm-ledger.json` is enforced by `internal/compatibility/ledger_test.go`. It supports exactly these evidence classes:

| Evidence class | What it can claim | Required boundary |
| --- | --- | --- |
| `source_review` | Reviewed upstream surface coverage | exact JetKVM source SHA; firmware not observed; offline upstream environment; pass |
| `source_drift` | Changed monitored upstream paths need review | exact JetKVM SHA; firmware not observed; offline upstream environment; review required |
| `read_only_hardware` | Narrow physical MCP observation | source not attributed; firmware not retained; non-production physical environment; pass |

Every entry carries an exact server source SHA, allowed check names, date, and non-empty limitations. Do not turn a source review into a firmware qualification, retain private device facts in the ledger, or treat a read-only run as proof of mutating behavior.

## Read-only physical validation

`cmd/jetkvm-mcp-validate` launches the server over stdio with a real configured device and a two-minute overall context. It requires four listed tools to advertise read-only/non-destructive/idempotent/non-open-world annotations, then runs: configured-device list, status, virtual-media status, and one capture under a 30-second sub-timeout. It validates schema shape, redaction, PNG content/dimensions, and only reports sanitized check names and capture dimensions/size. It deliberately discards child stderr and excludes device names, paths, status values, image bytes, and error text from output.

It does **not** mutate hardware, qualify HID/power/mount/upload paths, retain firmware identity, or prove broad appliance interoperability. Run it only against authorized non-production equipment and update the ledger with its limitations:

```sh
jetkvm-mcp-validate --binary ./bin/jetkvm-mcp --config config.yaml --device lab
```

Renew review/qualification when any monitored upstream surface, provider/session protocol, public tool contract, firmware target, capture/FFmpeg path, or physical test fixture changes. The CI-level protocol evidence is separate: see [testing and CI](../quality/testing-and-ci.md).
