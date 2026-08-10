# JetKVM MCP repository contract

This repository is the clean implementation for Forgejo issue #1. It has no shared Git history or remotes with `jetkvm-mcp-legacy`.

## Product

Build one conventional Go program that operates JetKVM through MCP over stdio or Streamable HTTP. Use the official MCP Go SDK and MCP 2026-07-28 behavior. Keep logs on stderr.

The required surface is status, screenshots, keyboard, mouse, virtual media, and seven separately named power/wake tools. Raw RPC is a local `debug rpc` CLI subcommand only and must never be registered as an MCP tool.

HTTP binds to loopback without authentication by default. Bearer authentication is optional and required for non-loopback binding. Do not add OAuth or deprecated HTTP+SSE initially.

Distribution is exactly: Go install (recommended), downloadable Linux/macOS binaries for amd64/arm64, and one Linux container image for amd64/arm64. Windows is unsupported and not planned.

## Exclusions

Do not add policy/grant engines, qualification databases, firmware approval matrices, decoder sidecars or IPC services, multi-container runtime architecture, fake-device authority, speculative attacker framing, or custom MCP transports.

## Legacy boundary

Legacy checkouts are read-only evidence until the clean build is verified. Never add them as remotes, merge/cherry-pick their commits, copy `.git`, or bulk-copy their trees. Reuse only reviewed source/protocol details needed by an active test-first slice. Preserve attribution and source-pinned protocol evidence where appropriate.

## Development

Use vertical RED-GREEN-REFACTOR slices. Run the focused failing test before implementation, then focused and full tests after. Prefer standard Go layout and the smallest dependency set justified by implemented behavior.

Do not commit, push, publish, merge, release, archive the legacy remote, or delete legacy local checkouts without Anchor’s explicit action at the applicable approval gate.
