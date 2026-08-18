---
type: delivery guide
title: Release and Container Delivery
description: Native and container artifact construction, release admission, verification evidence, and deployment constraints.
tags: [release, container, ci]
---

# Release and Container Delivery

`.goreleaser.yaml` produces static (`CGO_ENABLED=0`) `jetkvm-mcp` archives for Linux amd64 and arm64 only, injects `main.version`, includes `LICENSE` and third-party notices, emits SHA-256 checksums, and produces archive SPDX JSON SBOMs. `Dockerfile` builds with Go 1.26.6, then runs the binary in Debian trixie-slim with `ca-certificates` and FFmpeg as UID/GID 10001. OCI labels carry source/revision/version/created build arguments.

## Local artifact checks

| Intent | Command | What it proves |
| --- | --- | --- |
| Normal binary | `make build` | Trimpath local binary in `bin/jetkvm-mcp` |
| Full native snapshot | `make release-snapshot` | Notices, GoReleaser snapshot, and native artifact verification |
| Container build | `make container` | Local image with dev metadata |
| Platform container verification | `make container-verify` | amd64 and arm64 images, SPDX output, smoke/metadata validation |
| OCI release rehearsal | `make container-release-snapshot` | Multi-platform OCI archive, per-platform SBOM, manifest verification |

Release tooling is pinned in `tools`; Makefile release targets force the release Go version and invoke checked-in verification scripts. Do not infer that packaging definitions mean a release exists.

## Workflow admission

`release.yml` has verify (pull request), rehearse (`workflow_dispatch` from `main`), and publish (protected annotated semver tag) modes. The publish path additionally requires owner actor, first run attempt, tag resolving to `GITHUB_SHA`, ancestry from `main`, and successful required CI job names. Native and container subjects are staged independently and passed to the final composite action. Rehearsal signs/attests retained evidence but publishes nothing. Publish has credentials for GitHub Release/package publication and emits immutable release records, checksums, signatures/provenance/SBOM evidence; `latest` moves only after complete stable publication.

Production tags also require structured JSON release notes and a reviewed physical-qualification ledger entry for the exact tag commit. Consult `docs/physical-qualification.md` and release workflow policy tests before modifying this gate.

## Runtime constraints

The image contains no config or credentials. Supply config and media as read-only mounts; inject secret environment values at process start; use a read-only root filesystem with writable `/tmp` where needed; retain non-root user, drop capabilities, disallow privilege escalation, and set CPU/memory/PID limits. FFmpeg is present in the image, unlike a native binary installation where it must be on `PATH`. See [operations](../operations/runbook.md) for service/proxy shutdown behavior.
