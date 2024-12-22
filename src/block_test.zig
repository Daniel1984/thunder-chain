const std = @import("std");
const testing = std.testing;
const Block = @import("./block.zig").Block;
const Transaction = @import("./transaction.zig").Transaction;
const Ed25519 = std.crypto.sign.Ed25519;

// Test helpers
fn createDummyTransaction() Transaction {
    return Transaction.init([_]u8{1} ** 32, // from
        [_]u8{2} ** 32, // to
        1000 // amount
    );
}

fn createUniqueTransaction(value: u8) Transaction {
    return Transaction.init([_]u8{value} ** 32, // from
        [_]u8{value +% 1} ** 32, // to
        1000 // amount
    );
}

test "Block - initialization" {
    const allocator = testing.allocator;

    var block = try Block.init(allocator);
    defer block.deinit();

    // Test initial values
    try testing.expectEqual(@as(u64, 0), block.height);
    try testing.expectEqual(@as(u64, 0), block.nonce);
    try testing.expect(block.transactions.items.len == 0);

    // Test hash initialization (should be all zeros)
    const expected_hash = [_]u8{0} ** 32;
    try testing.expectEqualSlices(u8, &expected_hash, &block.hash);
    try testing.expectEqualSlices(u8, &expected_hash, &block.merkle_root);
}

test "Block - hash calculation" {
    const allocator = testing.allocator;

    var block = try Block.init(allocator);
    defer block.deinit();

    // Calculate initial hash
    const initial_hash = try block.calculateHash();

    // Modify block data
    block.nonce += 1;

    // Calculate new hash
    const new_hash = try block.calculateHash();

    // Hashes should be different after modification
    try testing.expect(!std.mem.eql(u8, &initial_hash, &new_hash));
}

test "Block - transaction management" {
    const allocator = testing.allocator;

    var block = try Block.init(allocator);
    defer block.deinit();

    // Add a transaction using addTransaction
    const tx = createDummyTransaction();
    try block.addTransaction(tx);

    // Verify transaction was added
    try testing.expect(block.transactions.items.len == 1);
    try testing.expectEqual(tx.amount, block.transactions.items[0].amount);

    // Verify merkle root was updated
    const zero_hash = [_]u8{0} ** 32;
    try testing.expect(!std.mem.eql(u8, &zero_hash, &block.merkle_root));
}

test "Block - merkle root with single transaction" {
    const allocator = testing.allocator;

    var block = try Block.init(allocator);
    defer block.deinit();

    // Add one transaction
    const tx = createDummyTransaction();
    try block.addTransaction(tx);

    // Calculate merkle root
    const first_root = try block.calculateMerkleRoot();

    // Calculate it again - should get same result
    const second_root = try block.calculateMerkleRoot();

    // Verify deterministic behavior
    try testing.expectEqualSlices(u8, &first_root, &second_root);

    // Verify root changes with different transaction
    try block.addTransaction(createUniqueTransaction(5));
    const new_root = try block.calculateMerkleRoot();
    try testing.expect(!std.mem.eql(u8, &first_root, &new_root));
}

test "Block - merkle root with multiple transactions" {
    const allocator = testing.allocator;

    var block = try Block.init(allocator);
    defer block.deinit();

    // Add transactions with different data
    try block.addTransaction(createUniqueTransaction(1));
    try block.addTransaction(createUniqueTransaction(2));
    try block.addTransaction(createUniqueTransaction(3));
    try block.addTransaction(createUniqueTransaction(4));

    // Get first merkle root
    const first_root = try block.calculateMerkleRoot();

    // Add another transaction
    try block.addTransaction(createUniqueTransaction(5));

    // Get second merkle root
    const second_root = try block.calculateMerkleRoot();

    // Roots should be different
    try testing.expect(!std.mem.eql(u8, &first_root, &second_root));
}

test "Block - merkle root with odd number of transactions" {
    const allocator = testing.allocator;

    var block = try Block.init(allocator);
    defer block.deinit();

    // Add three transactions
    try block.addTransaction(createUniqueTransaction(1));
    try block.addTransaction(createUniqueTransaction(2));
    try block.addTransaction(createUniqueTransaction(3));

    // Calculate merkle root
    const root = try block.calculateMerkleRoot();

    // Root should not be zeros
    const zero_hash = [_]u8{0} ** 32;
    try testing.expect(!std.mem.eql(u8, &zero_hash, &root));
}

test "Block - hash pair calculation" {
    const allocator = testing.allocator;

    var block = try Block.init(allocator);
    defer block.deinit();

    const left = [_]u8{1} ** 32;
    const right = [_]u8{2} ** 32;

    // Calculate hash pair twice
    const hash1 = try Block.hashPair(left, right);
    const hash2 = try Block.hashPair(left, right);

    // Same input should produce same hash
    try testing.expectEqualSlices(u8, &hash1, &hash2);

    // Different order should produce different hash
    const hash3 = try Block.hashPair(right, left);
    try testing.expect(!std.mem.eql(u8, &hash1, &hash3));
}

test "Block - memory management with transactions" {
    const allocator = testing.allocator;

    // Test that all memory is freed properly
    {
        var block = try Block.init(allocator);
        defer block.deinit();

        // Add several transactions
        try block.addTransaction(createUniqueTransaction(1));
        try block.addTransaction(createUniqueTransaction(2));
        try block.addTransaction(createUniqueTransaction(3));

        // Calculate merkle root multiple times
        _ = try block.calculateMerkleRoot();
        _ = try block.calculateMerkleRoot();

        // Modify and calculate hash
        block.nonce += 1;
        _ = try block.calculateHash();
    }

    // If we get here without memory leaks, the test passes
}
