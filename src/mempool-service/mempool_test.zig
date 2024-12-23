const std = @import("std");
const testing = std.testing;
const Ed25519 = std.crypto.sign.Ed25519;
const MemPool = @import("mempool.zig").MemPool;
const MempoolError = @import("mempool.zig").MempoolError;
const SqliteService = @import("sqlite-service").SqliteService;
const Transaction = @import("transaction").Transaction;

test "MemPool - initialization" {
    const allocator = testing.allocator;

    var pool = try MemPool.init(allocator, 1000);
    defer pool.deinit();

    try testing.expectEqual(@as(usize, 1000), pool.max_size);
}

test "MemPool - initialization with 0 max_size" {
    const allocator = testing.allocator;
    try testing.expectError(MempoolError.InvalidSize, MemPool.init(allocator, 0));
}

test "MemPool - addTransaction" {
    const allocator = testing.allocator;
    var pool = try MemPool.init(allocator, 1000);
    defer pool.deinit();

    // Generate keypair for test transaction
    const kp = try Ed25519.KeyPair.create(null);

    // Create and sign a transaction
    var tx = Transaction.init(kp.public_key.bytes, [_]u8{2} ** 32, 1000);
    tx.fee = 10;
    try tx.sign(kp);

    // Add transaction to mempool
    try pool.addTransaction(tx);

    // // Verify transaction was added by querying the database
    // var stmt = try pool.db.db.prepare(
    //     \\SELECT from_addr, to_addr, amount, fee
    //     \\FROM mempool
    //     \\WHERE id = ?{[]const u8}
    // );
    // defer stmt.deinit();

    // const tx_hash = try tx.calculateHash();
    // const row = try stmt.one(struct {
    //     from_addr: [32]u8,
    //     to_addr: [32]u8,
    //     amount: u64,
    //     fee: u64,
    // }, .{}, .{tx_hash[0..]});

    // // Verify row exists and data matches
    // try testing.expect(row != null);
    // const result = row.?;
    // try testing.expectEqualSlices(u8, &tx.from, &result.from_addr);
    // try testing.expectEqualSlices(u8, &tx.to, &result.to_addr);
    // try testing.expectEqual(tx.amount, result.amount);
    // try testing.expectEqual(tx.fee, result.fee);
}
