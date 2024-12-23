const std = @import("std");
const SqliteService = @import("sqlite-service").SqliteService;
const Transaction = @import("transaction").Transaction;
const Allocator = std.mem.Allocator;

pub const MempoolError = error{
    TransactionAlreadyExists,
    InvalidTransaction,
    MempoolFull,
    InsufficientFee,
    InvalidSize,
};

pub const MemPool = struct {
    db: SqliteService,
    allocator: Allocator,
    mutex: std.Thread.Mutex,
    max_size: usize,

    pub fn init(allocator: Allocator, max_size: usize) !MemPool {
        if (max_size == 0) {
            return MempoolError.InvalidSize;
        }

        var db = try SqliteService.init("./test.db");

        try db.exec(
            \\CREATE TABLE IF NOT EXISTS mempool(
            \\    id BLOB PRIMARY KEY,
            \\    from_addr BLOB PRIMARY KEY,
            \\    to_addr BLOB PRIMARY KEY,
            \\    signature BLOB PRIMARY KEY,
            \\    fee INTEGER NOT NULL,
            \\    amount INTEGER NOT NULL,
            \\    timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
            \\    expires INTEGER NOT NULL DEFAULT (strftime('%s', 'now') + 1500)
            \\)
        , .{}, .{});

        try db.exec("CREATE INDEX IF NOT EXISTS idx_mempool_fee ON mempool(fee DESC)", .{}, .{});
        try db.exec("CREATE INDEX IF NOT EXISTS idx_mempool_expires ON mempool(expires)", .{}, .{});

        return .{
            .allocator = allocator,
            .mutex = .{},
            .max_size = max_size,
            .db = db,
        };
    }

    pub fn deinit(self: *MemPool) void {
        self.db.deinit();
    }

    pub fn addTransaction(self: *MemPool, tx: Transaction) !void {
        const tx_hash = try tx.calculateHash();

        std.debug.print("Adding transaction with hash: {s}\n", .{std.fmt.fmtSliceHexLower(tx_hash[0..])});

        const query =
            \\INSERT INTO mempool 
            \\(id, from_addr, to_addr, signature, fee, amount) 
            \\VALUES 
            \\(:id, :from_addr, :to_addr, :signature, :fee, :amount)
        ;

        var stmt = try self.db.db.prepare(query);
        defer stmt.deinit();

        try stmt.exec(.{}, .{
            .id = tx_hash[0..],
            .from_addr = tx.from[0..],
            .to_addr = tx.to[0..],
            .signature = tx.signature[0..],
            .fee = tx.fee,
            .amount = tx.amount,
        });

        std.debug.print("Transaction added successfully\n", .{});
    }

    // pub fn loadFromDisk(self: *MemPool) !void {
    //     // On startup, load persisted transactions
    //     var stmt = try self.db.prepare("SELECT hash, tx_data FROM mempool");
    //     defer stmt.deinit();

    //     while (try stmt.step()) {
    //         const row = try stmt.row();
    //         const tx = try Transaction.deserialize(row.get("tx_data"));
    //         try self.transactions.put(row.get("hash"), tx);
    //     }
    // }

    // pub fn removeTransaction(self: *MemPool, tx_hash: [32]u8) void {
    //     self.mutex.lock();
    //     defer self.mutex.unlock();
    //     std.log.info("about to remove tx by hash {s}", tx_hash);
    //     // remove tx from DB here
    // }

    // fn validateTransaction(_: *MemPool, tx: Transaction) !void {
    //     // Verify transaction signature
    //     try tx.verify();
    //     // Add more validation as needed
    // }

    // // Get transactions for new block (sorted by fee)
    // pub fn getTransactionsForBlock(self: *MemPool, max_count: usize) ![]Transaction {
    //     self.mutex.lock();
    //     defer self.mutex.unlock();

    //     var result = std.ArrayList(Transaction).init(self.allocator);
    //     defer result.deinit();

    //     var it = self.transactions.iterator();
    //     while (it.next()) |entry| {
    //         if (result.items.len >= max_count) break;
    //         try result.append(entry.value_ptr.*);
    //     }

    //     return result.toOwnedSlice();
    // }
};
