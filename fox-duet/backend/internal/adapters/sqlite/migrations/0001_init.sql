-- Games are persisted as a JSON snapshot of the domain aggregate.
-- `version` supports optimistic concurrency (UPDATE ... WHERE version = ?).
CREATE TABLE games (
    id         TEXT PRIMARY KEY,
    status     TEXT    NOT NULL,           -- 'lobby' | 'playing' | 'won' | 'lost' | 'ended'
    state      TEXT    NOT NULL,           -- JSON snapshot of the game aggregate
    version    INTEGER NOT NULL DEFAULT 0,
    created_at TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Token hashes live outside the game JSON snapshot. token_hash is NOT unique:
-- rematch games copy the hashes of the previous game.
CREATE TABLE player_tokens (
    game_id    TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    player_id  TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (game_id, player_id)
);

CREATE INDEX games_updated_at ON games(updated_at);
