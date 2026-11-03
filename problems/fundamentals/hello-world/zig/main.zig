const std = @import("std");

pub const Solution = struct {
    pub fn solve(allocator: std.mem.Allocator, input_data: std.json.Value) ![]const u8 {
        const name = if (input_data == .object) blk: {
            const value = input_data.object.get("name") orelse break :blk "World";
            break :blk if (value == .string) value.string else "World";
        } else "World";
        return std.mem.concat(allocator, u8, &.{ "Hello ", name });
    }
};
