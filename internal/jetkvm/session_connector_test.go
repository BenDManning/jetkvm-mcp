package jetkvm

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"testing"
	"time"

	"github.com/BenDManning/jetkvm-mcp/internal/mcpserver"
)

type deterministicConnector struct {
	session *deterministicConnectedSession
	err     error
	calls   int
}

func (connector *deterministicConnector) Connect(context.Context, DeviceConfig) (ConnectedSession, error) {
	connector.calls++
	if connector.err != nil {
		return nil, connector.err
	}
	return connector.session, nil
}

type deterministicConnectedSession struct {
	*fakeSession
	lost       chan struct{}
	takenOver  bool
	closeCalls int
	closeStart chan struct{}
	closeDone  chan struct{}
}

func (session *deterministicConnectedSession) Done() <-chan struct{}    { return session.lost }
func (session *deterministicConnectedSession) RecognizedTakeover() bool { return session.takenOver }

func (session *deterministicConnectedSession) Close(ctx context.Context) error {
	session.closeCalls++
	if session.closeStart != nil {
		select {
		case <-session.closeStart:
		default:
			close(session.closeStart)
		}
	}
	if session.closeDone != nil {
		select {
		case <-session.closeDone:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func TestRecognizedTakeoverLatchesAndRejectsOrdinaryReconnect(t *testing.T) {
	base, _ := url.Parse("https://jetkvm.invalid")
	session := &deterministicConnectedSession{
		fakeSession: &fakeSession{results: map[string]any{methodPing: "pong"}},
		lost:        make(chan struct{}),
		takenOver:   true,
	}
	connector := &deterministicConnector{session: session}
	manager, err := NewManager([]DeviceConfig{{Name: "lab", BaseURL: *base}}, connector)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = manager.Close(context.Background()) })
	if _, err := manager.DebugRPC(context.Background(), "lab", methodPing, nil, false); err != nil {
		t.Fatal(err)
	}
	close(session.lost)
	deadline := time.After(time.Second)
	for manager.owners["lab"].Snapshot().Ownership != ownerOwnershipTakenOver {
		select {
		case <-deadline:
			t.Fatal("recognized takeover did not latch taken-over ownership")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	_, err = manager.DebugRPC(context.Background(), "lab", methodPing, nil, false)
	var classified *OperationError
	if !errors.As(err, &classified) || classified.Code != "session_taken_over" || classified.Outcome != ToolOutcomeNotSent || classified.Retryable {
		t.Fatalf("ordinary work after takeover = %#v", err)
	}
	if connector.calls != 1 {
		t.Fatalf("ordinary work after takeover opened %d connections", connector.calls)
	}
}

type terminalOperationSession struct {
	started   chan struct{}
	done      chan struct{}
	takenOver bool
}

func (session *terminalOperationSession) Call(ctx context.Context, method string, _ any, result any) error {
	if method == methodPing {
		if pong, ok := result.(*string); ok {
			*pong = "pong"
		}
		return nil
	}
	close(session.started)
	<-ctx.Done()
	return ErrSessionClosed
}
func (*terminalOperationSession) Upload(context.Context, string, io.Reader, int64) error { return nil }
func (*terminalOperationSession) CaptureH264(context.Context) ([]byte, time.Time, error) {
	return nil, time.Time{}, nil
}
func (session *terminalOperationSession) Done() <-chan struct{}    { return session.done }
func (*terminalOperationSession) Close(context.Context) error      { return nil }
func (session *terminalOperationSession) RecognizedTakeover() bool { return session.takenOver }
func (session *terminalOperationSession) SuppressHIDCleanup() bool { return session.takenOver }

func TestTerminalGenerationClassifiesInconclusiveDispatchByOwnershipAndOperation(t *testing.T) {
	for _, test := range []struct {
		name        string
		takenOver   bool
		mutation    bool
		wantCode    string
		wantOutcome string
	}{
		{name: "takeover read", takenOver: true, wantCode: "session_taken_over", wantOutcome: ToolOutcomeFailed},
		{name: "loss read", wantCode: "ownership_uncertain", wantOutcome: ToolOutcomeFailed},
		{name: "takeover mutation", takenOver: true, mutation: true, wantCode: "session_taken_over", wantOutcome: ToolOutcomeUnknown},
		{name: "loss mutation", mutation: true, wantCode: "ownership_uncertain", wantOutcome: ToolOutcomeUnknown},
	} {
		t.Run(test.name, func(t *testing.T) {
			base, _ := url.Parse("https://jetkvm.invalid")
			session := &terminalOperationSession{started: make(chan struct{}), done: make(chan struct{}), takenOver: test.takenOver}
			manager, err := NewManager([]DeviceConfig{{Name: "lab", BaseURL: *base}}, &releaseConnector{session: session})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = manager.Close(context.Background()) })
			result := make(chan error, 1)
			go func() {
				if test.mutation {
					_, callErr := manager.Power(context.Background(), "lab", mcpserver.PowerActionPressHostPowerButton, "")
					result <- callErr
					return
				}
				_, callErr := manager.DebugRPC(context.Background(), "lab", methodLocalVersion, json.RawMessage(`{}`), false)
				result <- callErr
			}()
			<-session.started
			close(session.done)
			var classified *OperationError
			if err := <-result; !errors.As(err, &classified) || classified.Code != test.wantCode || classified.Outcome != test.wantOutcome || classified.Retryable {
				t.Fatalf("terminal operation = %#v, want %s/%s nonretryable", err, test.wantCode, test.wantOutcome)
			}
		})
	}
}

func TestDeterministicConnectorControlsFailureLossAndCleanupCompletion(t *testing.T) {
	base, _ := url.Parse("https://jetkvm.invalid")
	newManager := func(connector SessionConnector) *Manager {
		manager, err := NewManager([]DeviceConfig{{Name: "lab", BaseURL: *base}}, connector)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = manager.Close(context.Background()) })
		return manager
	}
	connectErr := errors.New("controlled connect failure")
	manager := newManager(&deterministicConnector{err: connectErr})
	if _, err := manager.DebugRPC(context.Background(), "lab", methodPing, nil, false); !errors.Is(err, connectErr) {
		t.Fatalf("connect error = %v, want controlled failure", err)
	}
	pingErr := errors.New("controlled ping failure")
	manager = newManager(&deterministicConnector{session: &deterministicConnectedSession{
		fakeSession: &fakeSession{err: map[string]error{methodPing: pingErr}},
		lost:        make(chan struct{}),
	}})
	if _, err := manager.DebugRPC(context.Background(), "lab", methodPing, nil, false); !errors.Is(err, pingErr) {
		t.Fatalf("ping error = %v, want controlled RPC failure", err)
	}

	session := &deterministicConnectedSession{
		fakeSession: &fakeSession{results: map[string]any{methodPing: "pong"}},
		lost:        make(chan struct{}),
		closeStart:  make(chan struct{}),
		closeDone:   make(chan struct{}),
	}
	manager = newManager(&deterministicConnector{session: session})
	result := make(chan error, 1)
	go func() {
		_, callErr := manager.DebugRPC(context.Background(), "lab", methodPing, nil, false)
		result <- callErr
	}()
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	close(session.lost)
	<-session.closeStart
	select {
	case <-session.Done():
	default:
		t.Fatal("controlled session loss was not observable")
	}
	close(session.closeDone)
	deadline := time.After(time.Second)
	for manager.owners["lab"].Snapshot().Ownership != ownerOwnershipUncertain {
		select {
		case <-deadline:
			t.Fatal("controlled loss cleanup did not latch ownership uncertainty")
		default:
			time.Sleep(time.Millisecond)
		}
	}
	_, err := manager.DebugRPC(context.Background(), "lab", methodPing, nil, false)
	var classified *OperationError
	if !errors.As(err, &classified) || classified.Code != "ownership_uncertain" || classified.Outcome != ToolOutcomeNotSent || classified.Retryable {
		t.Fatalf("ordinary work after loss = %#v", err)
	}
	if connectorCalls := manager.owners["lab"].connector.(*deterministicConnector).calls; connectorCalls != 1 {
		t.Fatalf("ordinary work after loss opened %d connections", connectorCalls)
	}
}

func TestConnectedSessionCleanupCanBeRetriedAfterCallerContextExpires(t *testing.T) {
	connectedCtx, cancelConnected := context.WithCancel(context.Background())
	connected := &connectedSession{ctx: connectedCtx, cancel: cancelConnected}
	releasePump := make(chan struct{})
	connected.pumps.Add(1)
	go func() {
		<-releasePump
		connected.pumps.Done()
	}()

	expired, cancelExpired := context.WithCancel(context.Background())
	cancelExpired()
	if err := connected.Close(expired); !errors.Is(err, context.Canceled) {
		t.Fatalf("expired cleanup error = %v, want canceled", err)
	}
	select {
	case <-connected.Done():
	default:
		t.Fatal("cleanup did not close the session after caller cancellation")
	}
	close(releasePump)
	if err := connected.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestManagerUsesConnectorAndClosesResidentSessionAtShutdown(t *testing.T) {
	base, _ := url.Parse("https://jetkvm.invalid")
	session := &deterministicConnectedSession{
		fakeSession: &fakeSession{results: map[string]any{methodPing: "pong"}},
		lost:        make(chan struct{}),
	}
	connector := &deterministicConnector{session: session}
	manager, err := NewManager([]DeviceConfig{{Name: "lab", BaseURL: *base}}, connector)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := manager.DebugRPC(context.Background(), "lab", methodPing, nil, false); err != nil {
		t.Fatal(err)
	}
	if connector.calls != 1 || session.closeCalls != 0 {
		t.Fatalf("connect calls = %d, close calls = %d; want one resident session", connector.calls, session.closeCalls)
	}
	if err := manager.Close(context.Background()); err != nil {
		t.Fatal(err)
	}
	if session.closeCalls != 1 {
		t.Fatalf("shutdown close calls = %d, want 1", session.closeCalls)
	}
}
