const std = @import("std");
const Transaction = @import("./transaction.zig").Transaction;
const Allocator = std.mem.Allocator;
const Sha256 = std.crypto.hash.sha2.Sha256;

pub const Block = struct {
    // Block header
    hash: [32]u8, // Fixed size array for SHA-256 hash
    prev_hash: [32]u8, // Previous block hash
    timestamp: i64, // Unix timestamp
    height: u64, // Block height (unsigned)
    nonce: u64, // For proof of work
    merkle_root: [32]u8, // Merkle tree root hash

    // Block data
    transactions: std.ArrayList(Transaction), // List of transactions

    // Memory allocator for dynamic allocations
    allocator: Allocator,

    // Constructor
    pub fn init(allocator: Allocator) !Block {
        return Block{
            .hash = [_]u8{0} ** 32,
            .prev_hash = [_]u8{0} ** 32,
            .timestamp = std.time.timestamp(),
            .height = 0,
            .nonce = 0,
            .merkle_root = [_]u8{0} ** 32,
            .transactions = std.ArrayList(Transaction).init(allocator),
            .allocator = allocator,
        };
    }

    // Clean up resources
    pub fn deinit(self: *Block) void {
        self.transactions.deinit();
    }

    // Updated block hash calculation to include transactions
    pub fn calculateHash(self: *Block) ![32]u8 {
        // First calculate the merkle root
        self.merkle_root = try self.calculateMerkleRoot();

        var hasher = Sha256.init(.{});

        // Hash block header fields
        hasher.update(&self.prev_hash);
        hasher.update(std.mem.asBytes(&self.timestamp));
        hasher.update(std.mem.asBytes(&self.height));
        hasher.update(std.mem.asBytes(&self.nonce));
        hasher.update(&self.merkle_root);

        var hash: [32]u8 = undefined;
        hasher.final(&hash);
        return hash;
    }

    // Hash a pair of 32-byte hashes together
    pub fn hashPair(left: [32]u8, right: [32]u8) ![32]u8 {
        var hasher = Sha256.init(.{});

        hasher.update(&left);
        hasher.update(&right);

        var hash: [32]u8 = undefined;
        hasher.final(&hash);
        return hash;
    }

    // Add a transaction to the block
    pub fn addTransaction(self: *Block, transaction: Transaction) !void {
        try self.transactions.append(transaction);
        // Update merkle root when transaction is added
        self.merkle_root = try self.calculateMerkleRoot();
    }

    // Calculate merkle root of all transactions
    pub fn calculateMerkleRoot(self: *Block) ![32]u8 {
        // Handle empty block case
        if (self.transactions.items.len == 0) {
            return [_]u8{0} ** 32;
        }

        var arena = std.heap.ArenaAllocator.init(self.allocator);
        defer arena.deinit();
        const arena_allocator = arena.allocator();

        // First, calculate hashes of all transactions
        var current_level = std.ArrayList([32]u8).init(arena_allocator);
        defer current_level.deinit();

        // Get hash for each transaction
        for (self.transactions.items) |transaction| {
            const tx_hash = try transaction.calculateHash();
            try current_level.append(tx_hash);
        }

        // If odd number of transactions, duplicate last one
        if (current_level.items.len % 2 == 1) {
            try current_level.append(current_level.items[current_level.items.len - 1]);
        }

        // Keep hashing pairs until we get to the root
        while (current_level.items.len > 1) {
            var next_level = std.ArrayList([32]u8).init(arena_allocator);

            var i: usize = 0;
            while (i < current_level.items.len) : (i += 2) {
                // Hash pair of hashes together
                const combined_hash = try hashPair(current_level.items[i], current_level.items[i + 1]);
                try next_level.append(combined_hash);
            }

            // Update current level
            current_level.deinit();
            current_level = next_level;

            // If odd number of hashes, duplicate last one
            if (current_level.items.len % 2 == 1 and current_level.items.len > 1) {
                try current_level.append(current_level.items[current_level.items.len - 1]);
            }
        }

        return current_level.items[0];
    }
};
