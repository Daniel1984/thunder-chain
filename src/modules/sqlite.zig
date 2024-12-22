const std = @import("std");
// const sqlite = @import("sqlite");
const c = @cImport({
    @cInclude("sqlite3.h");
});

// pub const SqliteService = struct {
//     db: sqlite.Db,

//     pub fn init(path: [:0]const u8) !SqliteService {
//         var db_instance = try sqlite.Db.init(.{
//             .mode = sqlite.Db.Mode{ .File = path },
//             .open_flags = .{
//                 .write = true,
//                 .create = true,
//             },
//             .threading_mode = .MultiThread,
//         });

//         // Get pointer to db instance
//         const db = &db_instance;

//         _ = try db.pragma([128:0]u8, .{}, "journal_mode", "wal");
//         _ = try db.pragma([128:0]u8, .{}, "txlock", "immediate");
//         _ = try db.pragma([128:0]u8, .{}, "busy_timeout", "5000");
//         _ = try db.pragma([128:0]u8, .{}, "cache_size", "1000000000");

//         return .{
//             .db = db,
//         };
//     }
// };

// Define our error set
pub const SqliteError = error{
    OpenFailed,
    ExecFailed,
    PrepareStatementFailed,
    QueryFailed,
    ConnectionClosed,
};

pub const SqliteService = struct {
    db: *c.sqlite3,

    pub fn init(path: [:0]const u8) !SqliteService {
        var db: ?*c.sqlite3 = null;
        const rc = c.sqlite3_open(path.ptr, &db);

        if (rc != c.SQLITE_OK) {
            return SqliteError.OpenFailed;
        }

        // Set pragmas
        var err_msg: [*c]u8 = null;
        _ = c.sqlite3_exec(db, "PRAGMA journal_mode=WAL", null, null, &err_msg);
        _ = c.sqlite3_exec(db, "PRAGMA synchronous=NORMAL", null, null, &err_msg);
        _ = c.sqlite3_exec(db, "PRAGMA cache_size=64000", null, null, &err_msg);
        _ = c.sqlite3_exec(db, "PRAGMA temp_store=MEMORY", null, null, &err_msg);

        return .{
            .db = db.?,
        };
    }

    pub fn deinit(self: *SqliteService) void {
        _ = c.sqlite3_close(self.db);
    }

    pub fn exec(self: *SqliteService, query: [:0]const u8) !void {
        var err_msg: [*c]u8 = null;
        const rc = c.sqlite3_exec(self.db, query.ptr, null, null, &err_msg);

        if (rc != c.SQLITE_OK) {
            c.sqlite3_free(err_msg);
            return SqliteError.ExecFailed;
        }
    }
};
