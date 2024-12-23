const std = @import("std");
const sqlite = @import("sqlite");

pub const SqliteService = struct {
    db: sqlite.Db,

    pub fn init(path: [:0]const u8) !SqliteService {
        var db = try sqlite.Db.init(.{
            .mode = sqlite.Db.Mode{ .File = path },
            .open_flags = .{
                .write = true,
                .create = true,
            },
            .threading_mode = .MultiThread,
        });

        _ = try db.pragma([128:0]u8, .{}, "journal_mode", "wal");
        _ = try db.pragma([128:0]u8, .{}, "txlock", "immediate");
        _ = try db.pragma([128:0]u8, .{}, "busy_timeout", "5000");
        _ = try db.pragma([128:0]u8, .{}, "cache_size", "1000000000");

        return .{
            .db = db,
        };
    }

    pub fn deinit(self: *SqliteService) void {
        self.db.deinit();
    }

    pub fn exec(self: *SqliteService, comptime query: []const u8, options: anytype, values: anytype) !void {
        try self.db.exec(query, options, values);
    }

    pub fn execRuntime(self: *SqliteService, query: []const u8, options: anytype, values: anytype) !void {
        try self.db.exec(query, options, values);
    }
};
