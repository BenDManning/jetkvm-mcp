---
type: Runtime Guide
title: Sessions, Admission, and Device Protocol
description: Managed per-device JetKVM session lifecycle, capacity controls, scheduling, and appliance protocol boundary.
tags: [runtime, webrtc, jetkvm]
openwiki:
  roles: [runtime, domain]
  change_kinds: [lifecycle, concurrency, failure-semantics]
  source_paths: [internal/jetkvm/manager.go, internal/jetkvm/device_owner.go, internal/jetkvm/scheduling.go, internal/jetkvm/provider.go, internal/jetkvm/rpc_session.go]
  symbols: [Manager, deviceOwner, generationScheduler, WebRTCConnector, rpcSession]
  test_paths: [internal/jetkvm/device_owner_test.go, internal/jetkvm/scheduling_test.go, internal/jetkvm/session_test.go, internal/jetkvm/admission_test.go]
  invariants: [A device owner shares one healthy resident generation across demand and rejects new work while cleanup is incomplete.]
  validation_commands: [go test ./internal/jetkvm]
---

# Sessions, Admission, and Device Protocol

`jetkvm.Manager` is the sole MCP device implementation. It looks up a configured device, admits work, and maps public actions to private appliance calls. On construction it creates one `deviceOwner` per device. The owner serializes mutation/HID work, lazily creates a resident `ConnectedSession` generation, shares it across queued demand, and closes it after idle expiry, terminal loss, explicit release, or manager shutdown. `WebRTCConnector.Connect` owns the mechanics of creating one generation.

```mermaid
sequenceDiagram
  participant Manager
  participant Owner as device owner
  participant Provider
  participant Auth as appliance auth
  participant Signal as signaling WebSocket
  participant Peer as WebRTC peer
  participant RPC as rpc data channel
  Manager->>Owner: register operation
  Owner->>Provider: Connect when no healthy generation
  Provider->>Auth: authenticate
  Provider->>Signal: dial and send offer
  Signal-->>Peer: answer and ICE candidates
  Peer-->>RPC: rpc channel opens
  Provider-->>Owner: resident generation ready
  Owner->>RPC: schedule appliance call
  Owner-->>Manager: operation result
  Owner->>Peer: close after idle, loss, or release
```

This sequence shows lazy generation setup, reuse through the owner, and eventual close.

The provider uses a new cookie jar, a cloned HTTP transport with `Proxy=nil`, TLS 1.2 minimum, and `InsecureSkipVerify` only when configured. It authenticates, converts appliance HTTP(S) base URL to WS(S) signaling, creates a WebRTC data channel named `rpc`, sends an offer, pumps signaling, and waits for the data channel to open. It always creates the video-capable profile, which accepts exactly an H264 video track with 90 kHz clock and packetization mode 1. Connection and RPC request defaults are 15 seconds and 10 seconds. Close cancels the internal context, closes RPC/signaling/peer/idle HTTP connections, closes video, and waits for pumps until its caller context ends.

## Ownership, admission, and ordering

`deviceOwner` is a command-loop state machine. Demand can share an in-flight setup attempt and, after readiness ping, a healthy resident generation. It starts the idle timer only when no workers, queue entries, or demand remain. Idle expiry cancels and closes that generation; an ended generation also triggers cleanup. A cleanup failure is latched as degraded and ordinary work returns `busy` rather than reconnecting over an incomplete close. `Manager.Close` stops every owner and waits through its shutdown budget. `TestDeviceOwnerPingValidatesAndReusesResidentGeneration`, `TestDeviceOwnerIdleLeaseStartsAfterWorkAndReleasesGeneration`, and `TestDeviceOwnerDoesNotReconnectAfterIncompleteIdleCleanup` are the closest lifecycle evidence.

### Explicit session release

`Manager.ReleaseSession` is the implementation behind [`jetkvm_release_session`](../mcp/public-surface.md). It first takes a per-device lifecycle admission latch, so it fails immediately with `busy/not_sent` if ordinary device work is admitted; it also requires a global operation permit. The owner then rejects queued, active, connecting, or cleanly-closing work. From an idle, taken-over, or uncertain owner it records `released` without reconnecting; from an active generation it closes and joins the generation before reporting `status: released`. Repeating a completed release is idempotent.

Release is local resource ownership, not a device operation: it does not alter appliance or attached-host state, but it is intentionally sticky. Ordinary work does not reconnect a released owner; another authenticated operator must take over before normal work resumes. If cleanup cannot be completed, ownership is `uncertain` and the caller must not represent it as released; a later release retries cleanup. Preserve the order **latch -> drain/close -> released result**, and do not relax fail-fast admission. `release_session_test.go` provides the focused matrix: idle/no-connect, active-generation closure, uncertain-cleanup retry, work rejection, latch timing, cancellation, global capacity, and repeat release.

Default `Limits` are 16 total operations, 4 per device, 8 **connection attempts**, 2 captures, 2 decoders, and a 60-second session idle lease. Global, per-device, connection-attempt, capture, and decoder permits are non-blocking: exhausted capacity returns `busy` with `not_sent`, not a queue. A healthy resident generation does not consume a connection-attempt permit. Mutation/HID sequencing is different: the owner uses a cancellable FIFO gate that waits while the caller context is live. Thus unrelated devices and ordinary reads can progress, while HID, power, and virtual-media mutations for one device are serialized.

Within a generation, ordinary RPCs use a fair RPC gate. The current capability profile permits a read RPC to interleave between mutation RPC calls and permits a read RPC while capture is acquiring a frame; a session-wide fallback blocks those overlaps. Capture serializes frame acquisition, copies a completed frame before decoding, and may release the generation lease while decoding proceeds. Keep these deliberately narrow concurrency guarantees when changing scheduling: `TestInitialCapabilityProfileInterleavesReadAtMutationRPCBoundary`, `TestInitialCapabilityProfileOverlapsOneRPCWithFrameAcquisition`, `TestSessionWideFallbackDoesNotOverlapRPCAndFrameAcquisition`, and `TestCaptureReleasesGenerationLeaseAfterImmutableFrameCopy` cover them.

`Status` first requires `ping == "pong"`, then performs best-effort probes; failed optional probes become warnings unless context cancellation/deadline requires termination. Each power action maps to one appliance RPC in `powerRequest`. Unknown device/target and invalid local preparation fail before session creation.

## Failure taxonomy

`internal/jetkvm/errors.go` classifies transport/protocol/auth/video failures into `OperationError` codes and outcomes. `mcpserver.toolFailure` emits the stable JSON envelope `{version, code, message, outcome, retryable}`. The accepted public codes are `operation_failed`, `canceled`, `timeout`, `invalid_input`, `busy`, `authentication_failed`, `device_unavailable`, `video_unavailable`, `no_signal`, and `protocol_error`; outcomes are only `not_sent`, `failed`, or `unknown`.

Validation, unknown device/target, capacity, and cancellation before session callback are `not_sent`. A failure before RPC dispatch is likewise `not_sent`; write failure, timeout, cancellation, or an unconfirmed response after a mutation can be `unknown`, while a confirmed appliance RPC error can be `failed`. `classifyReadFailure` uses `failed`, and read `canceled`, `timeout`, `device_unavailable`, or `no_signal` is retryable unless already `not_sent`. Mutation failures are always non-retryable; absent reliable classification also defaults them to `unknown`. This distinction is the operational invariant, not a suggestion to re-run mutations.

## Change navigation

When changing ownership or scheduling, start in `internal/jetkvm/device_owner.go` and `scheduling.go`, then trace `Manager.withOperation` and `Manager.ReleaseSession` in `manager.go`. Preserve: no connection attempt after all setup waiters leave; no reconnect after incomplete cleanup; release remains fail-fast against admitted work and does not report completion before active resources close; cancellation before dispatch is `not_sent`; and stale completions cannot alter a replacement generation. `device_owner_test.go` names the ordinary lifecycle cases, `release_session_test.go` owns explicit release, and `scheduling_test.go` covers interleaving, cancellation, and capture frame ownership. Run `go test ./internal/jetkvm`; add `make race` when changing owner, scheduler, decoder, or video concurrency.

Test ownership also includes `manager_test.go`, `admission_test.go`, `provider_test.go`, `session_test.go`, `session_protocol_test.go`, `auth_test.go`, and signaling/RPC fuzz tests. The appliance compatibility evidence and drift process are documented in [JetKVM compatibility](../integrations/jetkvm-compatibility.md); the [control and capture guide](control-and-capture.md) consumes these scheduling rules for HID and screenshots.
