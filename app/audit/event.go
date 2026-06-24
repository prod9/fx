package audit

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"fx.prodigy9.co/data"
	"fx.prodigy9.co/data/page"
	"fx.prodigy9.co/fxlog"
)

// Actor is the operator behind a recorded action. The zero value is the system actor (no
// session) — it persists as NULL actor columns and renders as "system". A user id is
// never 0 (BIGSERIAL starts at 1), so 0 is an unambiguous "absent" sentinel.
type Actor struct {
	ID    int64
	Email string
}

// Event is a stored audit row. It is the read model returned by List; writes go through
// Record/Log. ActorID and TargetID are nullable because "system actor" and "batch action,
// no single target" are genuinely absent, not zero.
type Event struct {
	ID         int64           `json:"id" db:"id"`
	OccurredAt time.Time       `json:"occurred_at" db:"occurred_at"`
	ActorID    *int64          `json:"actor_id" db:"actor_id"`
	ActorEmail string          `json:"actor_email" db:"actor_email"`
	Action     string          `json:"action" db:"action"`
	TargetType string          `json:"target_type" db:"target_type"`
	TargetID   *int64          `json:"target_id" db:"target_id"`
	Detail     json.RawMessage `json:"detail" db:"detail"`
}

// Log records an event, swallowing any error after logging it. Audit writes are
// post-success and non-fatal: a trail-write failure must never fail the user's action.
// This is the recorder the mutation sites use.
func Log(ctx context.Context, actor Actor, action string, targetID int64, detail any) {
	if err := Record(ctx, actor, action, targetID, detail); err != nil {
		fxlog.Errorf("audit: record %s failed: %v", action, err)
	}
}

// Record writes one audit row and returns its error. Prefer Log at callsites; Record
// exists for the path that asserts on the write. target_type is derived from the action
// prefix so the two cannot drift. NULLIF maps the zero sentinels (actor id/email, target
// id) to NULL columns.
func Record(ctx context.Context, actor Actor, action string, targetID int64, detail any) error {
	targetType, _, _ := strings.Cut(action, ".")

	raw := []byte("{}")
	if detail != nil {
		marshaled, err := json.Marshal(detail)
		if err != nil {
			return err
		}
		raw = marshaled
	}

	return data.Exec(ctx, `
		INSERT INTO audit_events (actor_id, actor_email, action, target_type, target_id, detail)
		VALUES (NULLIF($1, 0), NULLIF($2, ''), $3, $4, NULLIF($5, 0), $6::jsonb)`,
		actor.ID, actor.Email, action, targetType, targetID, string(raw))
}

// List returns audit rows newest-first, paginated.
func List(ctx context.Context, pm page.Meta) (*page.Page[*Event], error) {
	out := &page.Page[*Event]{}
	err := page.Select(ctx, out, pm, `
		SELECT * FROM audit_events
		ORDER BY occurred_at DESC, id DESC`)
	return out, err
}
