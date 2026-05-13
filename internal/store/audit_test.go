package store

import (
	"context"
	"errors"
	"testing"
)

type fakePool struct {
	acquireErr error
	exec       func(sql string, args ...any) (any, error)
}

func (f *fakePool) Acquire(_ context.Context) (Conn, error) {
	if f.acquireErr != nil {
		return nil, f.acquireErr
	}
	return &fakeConn{exec: f.exec}, nil
}
func (f *fakePool) Close() {}

type fakeConn struct {
	exec func(sql string, args ...any) (any, error)
}

func (c *fakeConn) Exec(_ context.Context, sql string, args ...any) (any, error) {
	return c.exec(sql, args...)
}
func (c *fakeConn) Query(_ context.Context, _ string, _ ...any) (Rows, error) {
	return nil, errors.New("not implemented")
}
func (c *fakeConn) QueryRow(_ context.Context, _ string, _ ...any) Row { return nil }
func (c *fakeConn) Release()                                            {}

func TestWriteAudit_OK(t *testing.T) {
	var captured []any
	p := &fakePool{exec: func(sql string, args ...any) (any, error) {
		captured = args
		return nil, nil
	}}
	err := WriteAudit(context.Background(), p, AuditEntry{
		UserSub: "alice", Tool: "search_customer", Outcome: "ok", DurationMs: 12, Reason: "",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(captured) != 5 || captured[0] != "alice" || captured[1] != "search_customer" {
		t.Fatalf("unexpected args: %#v", captured)
	}
}

func TestWriteAudit_AcquireError(t *testing.T) {
	p := &fakePool{acquireErr: errors.New("pool drained")}
	if err := WriteAudit(context.Background(), p, AuditEntry{}); err == nil {
		t.Fatal("expected error")
	}
}
