CREATE TABLE IF NOT EXISTS mempool(
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  hash TEXT,
  from_addr TEXT NOT NULL,
  to_addr TEXT NOT NULL,
  signature TEXT NOT NULL,
  fee INTEGER NOT NULL,
  amount INTEGER NOT NULL,
  nonce INTEGER NOT NULL,
  timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  expires INTEGER NOT NULL DEFAULT (strftime('%s', 'now') + 1500)
);

CREATE INDEX IF NOT EXISTS idx_mempool_hash ON mempool(hash);
CREATE INDEX IF NOT EXISTS idx_mempool_fee ON mempool(fee DESC);
CREATE INDEX IF NOT EXISTS idx_mempool_expires ON mempool(expires);
