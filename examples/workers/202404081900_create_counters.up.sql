-- vim: filetype=SQL
CREATE TABLE counters (
	name  TEXT NOT NULL PRIMARY KEY,
	count BIGINT NOT NULL DEFAULT 0
);
