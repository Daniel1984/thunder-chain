const std = @import("std");
const SqliteService = @import("sqlite-service").SqliteService;
const Transaction = @import("../transaction.zig").Transaction;
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

        try db.exec("CREATE TABLE IF NOT EXISTS mempool(id BLOB PRIMARY KEY, fee INTEGER NOT NULL, timestamp INTEGER NOT NULL, tx_data BLOB NOT NULL)", .{}, .{});
        try db.exec("CREATE INDEX IF NOT EXISTS idx_mempool_fee ON mempool(fee DESC)", .{}, .{});

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
        // Add to in-memory map
        const tx_hash = try tx.calculateHash();
        try self.transactions.put(tx_hash, tx);

        // Persist to SQLite
        try self.db.exec("INSERT INTO mempool (hash, tx_data) VALUES (?, ?)", .{ tx_hash, tx.serialize() });
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
