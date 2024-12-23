const std = @import("std");
const testing = std.testing;
const crypto = std.crypto;
const Ed25519 = crypto.sign.Ed25519;
const Transaction = @import("./transaction.zig").Transaction;
const TransactionError = @import("./transaction.zig").TransactionError;

test "Transaction - initialization" {
    // Setup test data
    const from = [_]u8{1} ** 32;
    const to = [_]u8{2} ** 32;
    const amount: u64 = 1000;

    // Create transaction
    const tx = Transaction.init(from, to, amount);

    // Verify initial values
    try testing.expectEqualSlices(u8, &from, &tx.from);
    try testing.expectEqualSlices(u8, &to, &tx.to);
    try testing.expectEqual(amount, tx.amount);
    try testing.expectEqualSlices(u8, &[_]u8{0} ** 64, &tx.signature);
    try testing.expect(tx.timestamp > 0);
}

test "Transaction - hash calculation" {
    // Setup
    const from = [_]u8{1} ** 32;
    const to = [_]u8{2} ** 32;
    const amount: u64 = 1000;

    var tx = Transaction.init(from, to, amount);

    // Calculate hash
    const hash1 = try tx.calculateHash();

    // Modify transaction
    tx.amount += 1;

    // Calculate new hash
    const hash2 = try tx.calculateHash();

    // Hashes should be different
    try testing.expect(!std.mem.eql(u8, &hash1, &hash2));
}

test "Transaction - signing with valid key" {
    // Generate keypair
    const kp = try Ed25519.KeyPair.create(null);

    // Create transaction with correct public key
    var tx = Transaction.init(kp.public_key.bytes, [_]u8{2} ** 32, 1000);

    // Sign transaction
    try tx.sign(kp);

    // Verify signature is not zero
    const zero_signature = [_]u8{0} ** 64;
    try testing.expect(!std.mem.eql(u8, &zero_signature, &tx.signature));
}

test "Transaction - signing with invalid key" {
    // Generate keypair
    const kp = try Ed25519.KeyPair.create(null);

    // Create transaction with DIFFERENT public key/from
    var tx = Transaction.init([_]u8{1} ** 32, [_]u8{2} ** 32, 1000);

    // Attempt to sign - should fail
    try testing.expectError(TransactionError.InvalidPublicKey, tx.sign(kp));
}

test "Transaction - verification of valid signature" {
    // Generate keypair
    const kp = try Ed25519.KeyPair.create(null);

    // Create and sign transaction
    var tx = Transaction.init(kp.public_key.bytes, [_]u8{2} ** 32, 1000);
    try tx.sign(kp);

    // Verify signature
    try testing.expect(try tx.verify());
}

test "Transaction - verification of tampered transaction" {
    // Generate keypair
    const kp = try Ed25519.KeyPair.create(null);

    // Create and sign transaction
    var tx = Transaction.init(kp.public_key.bytes, [_]u8{2} ** 32, 1000);
    try tx.sign(kp);

    // Tamper with the transaction
    tx.amount += 1;

    // Verification should fail
    try testing.expectError(TransactionError.InvalidSignature, tx.verify());
}

test "Transaction - verification of invalid signature" {
    // Generate keypair
    const kp = try Ed25519.KeyPair.create(null);

    // Create and sign transaction
    var tx = Transaction.init(kp.public_key.bytes, [_]u8{2} ** 32, 1000);
    try tx.sign(kp);

    // Tamper with the signature
    tx.signature[0] += 1;

    // Verification should fail
    try testing.expectError(TransactionError.InvalidSignature, tx.verify());
}
