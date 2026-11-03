const std = @import("std");

pub const Solution = struct {
    pub fn solve(allocator: std.mem.Allocator, input_data: std.json.Value) ![]const u8 {
        // Use input_data.object.get("value"); return plain text for strings or JSON bytes otherwise.
        _ = allocator;
        _ = input_data;
        return "";
    }
};
