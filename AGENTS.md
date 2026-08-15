JetKVM MCP is a conventional Go program that operates JetKVM devices through MCP over stdio or Streamable HTTP.

## Commands

- Focused test: `go test ./path/to/package -run '^TestName$'`
- Build: `make build`
- Full validation: `go mod verify && make race && make verify`
- Format changed Go files: `gofmt -w <changed .go files>`

## Working rules

- Keep logs and diagnostics on stderr; reserve stdio-server stdout for MCP protocol traffic.
- Automated tests do not qualify physical compatibility; do not access physical hardware unless the task explicitly authorizes it.
- Keep credentials, device data, and generated artifacts out of source control.
- Use [README.md](README.md) for setup and operation, [docs/product-contract.md](docs/product-contract.md) before behavior changes, [docs/adr/README.md](docs/adr/README.md) before architecture changes, and [docs/protocol-sources.md](docs/protocol-sources.md) before protocol work or selective reuse.
- Do not prescribe or perform blanket rebases, stash clearing, branch pruning, merges, releases, deployments, or issue closure without checking live state and applicable authority.

## Agent skills

### Issue tracker

Issues live in GitHub Issues for `BenDManning/jetkvm-mcp`. See `docs/agents/issue-tracker.md`.

### Triage labels

Use the repository's existing label vocabulary; unconfigured roles require maintainer direction. See `docs/agents/triage-labels.md`.

### Domain docs

This is a single-context repository. See `docs/agents/domain.md`.
