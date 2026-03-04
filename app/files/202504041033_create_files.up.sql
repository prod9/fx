-- vim: filetype=SQL
CREATE TABLE files (
	id         SERIAL PRIMARY KEY,
	kind       TEXT NOT NULL,
	owner_id   BIGINT NOT NULL,
	owner_type TEXT NOT NULL,

	original_name  TEXT NOT NULL,
	content_type   TEXT NOT NULL,
	content_length BIGINT NOT NULL,
	created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_files_owner ON files (owner_type, owner_id);
CREATE INDEX idx_files_kind_owner ON files (kind, owner_type, owner_id);
