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

CREATE TABLE IF NOT EXISTS receipts (
  tx_hash TEXT PRIMARY KEY,
  block_hash TEXT NOT NULL,
  status TEXT NOT NULL,
  gas_used INTEGER NOT NULL,
  logs TEXT DEFAULT '[]' CHECK (json_valid(logs)),
  timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  FOREIGN KEY (block_hash) REFERENCES blocks(hash) ON DELETE CASCADE
) STRICT;

CREATE INDEX IF NOT EXISTS idx_receipts_block_hash ON receipts(block_hash);
CREATE INDEX IF NOT EXISTS idx_receipts_tx_hash ON receipts(tx_hash);
