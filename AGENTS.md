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

## Work tracking and delivery

- GitHub issues and pull requests are the canonical task-status system. Keep roadmap issue #32 current when an item's status, dependency, or execution order changes materially; do not create parallel Markdown TODO lists or introduce Beads/bd.
- Before stopping, run checks for changed surfaces, commit only intended paths, update the issue and roadmap, and leave a durable handoff with the worktree, branch, exact commit, last verified command and result, blocker, and next action.
- Once branch publication is authorized, do not stop at "ready to push": push the exact branch, verify its remote head and CI, and update or close the issue. If publication is unauthorized or fails, preserve a checkpoint and do not claim delivery.
- Do not prescribe or perform blanket rebases, stash clearing, branch pruning, merges, releases, deployments, or issue closure without checking live state and applicable authority.

## Agent skills

### Issue tracker

Issues live in GitHub Issues for `BenDManning/jetkvm-mcp`. See `docs/agents/issue-tracker.md`.

### Triage labels

Use the repository's existing label vocabulary; unconfigured roles require maintainer direction. See `docs/agents/triage-labels.md`.

### Domain docs

This is a single-context repository. See `docs/agents/domain.md`.
