CREATE TABLE IF NOT EXISTS mempool(
  id BLOB PRIMARY KEY,
  from_addr BLOB NOT NULL,
  to_addr BLOB NOT NULL,
  signature BLOB NOT NULL,
  fee INTEGER NOT NULL,
  amount INTEGER NOT NULL,
  timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  expires INTEGER NOT NULL DEFAULT (strftime('%s', 'now') + 1500)
);

CREATE INDEX IF NOT EXISTS idx_mempool_fee ON mempool(fee DESC);
CREATE INDEX IF NOT EXISTS idx_mempool_expires ON mempool(expires);
