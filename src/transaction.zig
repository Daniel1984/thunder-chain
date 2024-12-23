const std = @import("std");
const crypto = std.crypto;
const Ed25519 = crypto.sign.Ed25519;

pub const TransactionError = error{
    InvalidSignature,
    InvalidPublicKey,
    SigningError,
    InvalidDataLength,
};

pub const Transaction = struct {
    from: [32]u8, // Sender's public key
    to: [32]u8, // Recipient's public key
    amount: u64,
    fee: u64,
    signature: [64]u8, // Ed25519 signature
    timestamp: i64,
    expires: i64,

    pub fn init(from: [32]u8, to: [32]u8, amount: u64) Transaction {
        return Transaction{
            .from = from,
            .to = to,
            .amount = amount,
            .fee = 0,
            .signature = [_]u8{0} ** 64,
            .timestamp = std.time.timestamp(),
            .expires = 0,
        };
    }

    pub fn calculateHash(self: *const Transaction) ![32]u8 {
        var hasher = crypto.hash.sha2.Sha256.init(.{});

        hasher.update(&self.from);
        hasher.update(&self.to);
        hasher.update(std.mem.asBytes(&self.amount));
        hasher.update(std.mem.asBytes(&self.timestamp));

        var hash: [32]u8 = undefined;
        hasher.final(&hash);
        return hash;
    }

    pub fn sign(self: *Transaction, key_pair: Ed25519.KeyPair) !void {
        // Verify the public key matches the from address
        if (!std.mem.eql(u8, &key_pair.public_key.bytes, &self.from)) {
            return TransactionError.InvalidPublicKey;
        }

        // Calculate the message hash
        const message_hash = try self.calculateHash();

        // Sign the message hash as a slice
        const signature = try key_pair.sign(message_hash[0..], null);
        // std.mem.copy(u8, &self.signature, &signature.bytes);
        @memcpy(&self.signature, &signature.toBytes());
    }

    pub fn verify(self: *const Transaction) !bool {
        // Convert from address to public key
        var public_key: Ed25519.PublicKey = undefined;
        @memcpy(&public_key.bytes, &self.from);

        // Calculate the message hash
        const message_hash = try self.calculateHash();

        // Convert byte array to Signature
        var sig = Ed25519.Signature.fromBytes(self.signature);

        // Verify the signature
        sig.verify(&message_hash, public_key) catch |err| {
            std.debug.print("Signature verification failed: {}\n", .{err});
            return TransactionError.InvalidSignature;
        };

        return true;
    }
};
