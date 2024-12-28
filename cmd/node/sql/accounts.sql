CREATE TABLE IF NOT EXISTS accounts (
  address TEXT PRIMARY KEY,
  balance INTEGER NOT NULL DEFAULT 0 CHECK (balance >= 0),
  nonce INTEGER NOT NULL DEFAULT 0 CHECK (nonce >= 0),
  updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
) STRICT;

-- For transaction history/audit
CREATE TABLE IF NOT EXISTS balance_changes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  address TEXT NOT NULL,
  previous_balance INTEGER NOT NULL,
  new_balance INTEGER NOT NULL,
  change_amount INTEGER NOT NULL,
  block_height INTEGER NOT NULL,
  block_hash TEXT NOT NULL,
  tx_hash TEXT NOT NULL,
  timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  FOREIGN KEY(address) REFERENCES accounts(address)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_balance_changes_address ON balance_changes(address);
CREATE INDEX IF NOT EXISTS idx_balance_changes_block ON balance_changes(block_height);
