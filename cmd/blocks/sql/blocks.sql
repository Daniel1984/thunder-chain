CREATE TABLE IF NOT EXISTS blocks (
  hash TEXT PRIMARY KEY,
  prev_hash TEXT NOT NULL,
  merkle_root TEXT NOT NULL,
  height INTEGER NOT NULL UNIQUE DEFAULT 0 CHECK (height >= 0),
  difficulty INTEGER NOT NULL DEFAULT 0 CHECK (nonce >= 0),
  nonce INTEGER NOT NULL DEFAULT 0 CHECK (nonce >= 0),
  timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  transactions TEXT DEFAULT '[]' CHECK (json_valid(transactions))
) STRICT;

CREATE INDEX IF NOT EXISTS idx_blocks_height ON blocks(height);
