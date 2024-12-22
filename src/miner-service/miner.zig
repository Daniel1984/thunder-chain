const std = @import("std");
const Block = @import("../block.zig").Block;
const MemPool = @import("../mempool-service/mempool.zig").MemPool;

pub const Miner = struct {
    chain: *Chain,
    mempool: *MemPool,
    mining_address: [32]u8,
    is_mining: std.atomic.Bool,

    pub fn init(chain: *Chain, mempool: *MemPool, mining_address: [32]u8) Miner {
        // Initialize miner
    }

    pub fn startMining(self: *Miner) !void {
        while (self.is_mining.load(.Acquire)) {
            var new_block = try self.createNewBlock();
            try self.mineBlock(&new_block);
            try self.chain.addBlock(new_block);
            // Broadcast to network
        }
    }

    fn createNewBlock(self: *Miner) !Block {
        // Get transactions from mempool
        // Create coinbase transaction
        // Create new block
    }
};
