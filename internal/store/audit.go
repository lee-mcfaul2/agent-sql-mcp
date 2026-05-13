package store

import (
	"context"
)

type AuditEntry struct {
	UserSub    string
	Tool       string
	Outcome    string
	DurationMs int
	Reason     string
}

func WriteAudit(ctx context.Context, p Pool, e AuditEntry) error {
	c, err := p.Acquire(ctx)
	if err != nil {
		return err
	}
	defer c.Release()
	_, err = c.Exec(ctx, SQLInsertAudit, e.UserSub, e.Tool, e.Outcome, e.DurationMs, e.Reason)
	return err
}
