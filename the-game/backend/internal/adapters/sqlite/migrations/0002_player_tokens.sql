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
