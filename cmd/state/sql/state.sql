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
