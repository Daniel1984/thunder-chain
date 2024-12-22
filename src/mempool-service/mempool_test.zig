const std = @import("std");
const testing = std.testing;
const MemPool = @import("mempool.zig").MemPool;
const MempoolError = @import("mempool.zig").MempoolError;
const SqliteService = @import("sqlite-service").SqliteService;

test "MemPool - initialization" {
    const allocator = testing.allocator;

    var pool = try MemPool.init(allocator, 1000);
    defer pool.deinit();

    try testing.expectEqual(@as(usize, 1000), pool.max_size);
    // try testing.expectEqual(@as(usize, 0), pool.transactions.count());
}

test "MemPool - initialization with 0 max_size" {
    const allocator = testing.allocator;
    try testing.expectError(MempoolError.InvalidSize, MemPool.init(allocator, 0));
}

// test "MemPool - database setup" {
//     const allocator = testing.allocator;

//     var pool = try MemPool.init(allocator, 1000);
//     defer pool.deinit();

//     // Verify table exists by attempting to insert a test record
//     try pool.db.exec(
//         \\INSERT INTO mempool (id, fee, timestamp, tx_data)
//         \\VALUES (X'DEADBEEF', 100, strftime('%s','now'), X'CAFEBABE')
//     );

//     // Verify index exists by checking execution plan
//     var stmt = try pool.db.prepare("EXPLAIN QUERY PLAN SELECT * FROM mempool ORDER BY fee DESC");
//     defer stmt.deinit();

//     var uses_index = false;
//     while (try stmt.step()) {
//         const plan = try stmt.row();
//         const detail = plan.get("detail");
//         if (std.mem.indexOf(u8, detail, "USING INDEX idx_mempool_fee") != null) {
//             uses_index = true;
//             break;
//         }
//     }
//     try testing.expect(uses_index);
// }

// test "MemPool - cleanup on deinit" {
//     const allocator = testing.allocator;

//     {
//         _ = try MemPool.init(allocator, 1000);
//         // Let pool go out of scope and deinit
//     }

//     // Try to open the same database file and verify it's not locked
//     var db = try SqliteService.init("./test.db");
//     defer db.deinit();
//     try testing.expect(db.isOpen());
// }
