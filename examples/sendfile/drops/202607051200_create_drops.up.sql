CREATE TABLE drops (
    id         BIGSERIAL   PRIMARY KEY,
    token      TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
