const std = @import("std");
const Transaction = @import("./transaction.zig").Transaction;
const TransactionPool = @import("./mempool.zig").TransactionPool;
const RPCServer = @import("./rpc.zig").RPCServer;

pub const Node = struct {
    mempool: *TransactionPool,
    rpc_server: *RPCServer,
    // ... other fields ...

    // Handle new transaction from RPC
    pub fn handleRpcTransaction(self: *Node, tx: Transaction) !void {
        // 1. Validate transaction
        try tx.verify();

        // 2. Add to mempool
        try self.mempool.addTransaction(tx);

        // 3. Broadcast to peers
        try self.broadcastTransaction(tx);
    }

    // Handle transaction received from network
    pub fn handleNetworkTransaction(self: *Node, tx: Transaction) !void {
        // 1. Validate transaction
        try tx.verify();

        // 2. Check if already in mempool
        const tx_hash = try tx.calculateHash();
        if (self.mempool.transactions.contains(tx_hash)) {
            return;
        }

        // 3. Add to mempool
        try self.mempool.addTransaction(tx);

        // 4. Relay to other peers
        try self.broadcastTransaction(tx);
    }
};
