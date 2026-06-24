-- vim: filetype=SQL
CREATE TABLE audit_events (
  id          BIGSERIAL PRIMARY KEY,
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  actor_id    BIGINT,
  actor_email TEXT,

  action      TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id   BIGINT,

  detail      JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_audit_events_occurred ON audit_events (occurred_at DESC, id DESC);
