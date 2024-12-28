CREATE TABLE IF NOT EXISTS blocks (
  hash TEXT PRIMARY KEY,
  prev_hash TEXT NOT NULL,
  merkle_root TEXT NOT NULL,
  height INTEGER NOT NULL UNIQUE DEFAULT 0 CHECK (height >= 0),
  difficulty INTEGER NOT NULL DEFAULT 0 CHECK (nonce >= 0),
  nonce INTEGER NOT NULL DEFAULT 0 CHECK (nonce >= 0),
  timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  transactions JSON DEFAULT '[]' CHECK (json_valid(transactions))
) STRICT;

CREATE INDEX IF NOT EXISTS idx_blocks_prev_hash ON blocks(prev_hash);


-- CREATE TABLE chain_metadata (
--     key TEXT PRIMARY KEY,
--     value TEXT
-- );

-- -- For quick header validation/synchronization
-- CREATE TABLE block_headers (
--     hash TEXT PRIMARY KEY,
--     height INTEGER UNIQUE,
--     prev_hash TEXT,
--     timestamp INTEGER
-- );

-- -- Optional: track forks/orphans
-- CREATE TABLE orphan_blocks (
--     hash TEXT PRIMARY KEY,
--     height INTEGER,
--     prev_hash TEXT
-- );
