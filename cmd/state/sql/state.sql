CREATE TABLE IF NOT EXISTS accounts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  address TEXT UNIQUE NOT NULL,
  balance INTEGER NOT NULL DEFAULT 0 CHECK (balance >= 0),
  nonce INTEGER NOT NULL DEFAULT 0 CHECK (nonce >= 0),
  timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
) STRICT;

-- Trigger to update the timestamp
CREATE TRIGGER IF NOT EXISTS update_accounts_timestamp
AFTER
UPDATE
  ON accounts BEGIN
UPDATE
  accounts
SET
  timestamp = strftime('%s', 'now')
WHERE
  id = NEW.id;

END;

-- For transaction history/audit
CREATE TABLE IF NOT EXISTS balance_changes (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  account_id INTEGER NOT NULL,
  previous_balance INTEGER NOT NULL,
  new_balance INTEGER NOT NULL,
  change_amount INTEGER NOT NULL,
  block_height INTEGER NOT NULL,
  block_hash TEXT NOT NULL,
  tx_hash TEXT NOT NULL,
  timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  FOREIGN KEY(account_id) REFERENCES accounts(id)
) STRICT;

CREATE INDEX IF NOT EXISTS idx_balance_changes_account ON balance_changes(account_id);

CREATE INDEX IF NOT EXISTS idx_balance_changes_block ON balance_changes(block_height);

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
