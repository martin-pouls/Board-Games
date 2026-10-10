-- Games are persisted as a JSON snapshot of the domain aggregate.
-- `version` supports optimistic concurrency (UPDATE ... WHERE version = ?).
CREATE TABLE games (
    id         TEXT PRIMARY KEY,
    variant    TEXT    NOT NULL,           -- 'classic' | 'extreme'
    status     TEXT    NOT NULL,           -- 'lobby' | 'running' | 'won' | 'lost'
    state      TEXT    NOT NULL,           -- JSON snapshot of the game aggregate
    version    INTEGER NOT NULL DEFAULT 0,
    created_at TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
);
