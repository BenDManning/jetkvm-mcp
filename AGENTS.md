JetKVM MCP is a conventional Go program that operates JetKVM devices through MCP over stdio or Streamable HTTP.

## Commands

- Focused test: `go test ./path/to/package -run '^TestName$'`
- Build: `make build`
- Full validation: `go mod verify && make race && make verify`
- Format changed Go files: `gofmt -w <changed .go files>`

## Working rules

- Implement behavior in vertical RED-GREEN-REFACTOR slices: run the focused failing test first, then the focused and full checks after implementation.
- Keep logs and diagnostics on stderr; reserve stdio-server stdout for MCP protocol traffic.
- Automated tests do not qualify physical compatibility; do not access physical hardware unless the task explicitly authorizes it.
- Keep credentials, device data, and generated artifacts out of source control.
- Use [README.md](README.md) for setup and operation, [docs/product-contract.md](docs/product-contract.md) before behavior changes, and [docs/protocol-sources.md](docs/protocol-sources.md) before protocol work or selective reuse.
