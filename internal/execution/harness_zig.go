package execution

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"polyglot-rosetta/internal/domain/testcase"
)

func prepareZig(testCases []testcase.TestCase, marker string) (map[string]string, preparedProgram, error) {
	var harness strings.Builder
	harness.WriteString(`const std = @import("std");
const Solution = @import("solution.zig").Solution;

`)
	fmt.Fprintf(&harness, "const result_marker = %q;\n\n", marker)
	harness.WriteString(`
fn jsonEqual(left: std.json.Value, right: std.json.Value) bool {
    if (std.meta.activeTag(left) != std.meta.activeTag(right)) {
        const left_number = switch (left) { .integer => |value| @as(f64, @floatFromInt(value)), .float => |value| value, else => return false };
        const right_number = switch (right) { .integer => |value| @as(f64, @floatFromInt(value)), .float => |value| value, else => return false };
        return left_number == right_number;
    }
    return switch (left) {
        .null => true,
        .bool => |value| value == right.bool,
        .integer => |value| value == right.integer,
        .float => |value| value == right.float,
        .number_string => |value| std.mem.eql(u8, value, right.number_string),
        .string => |value| std.mem.eql(u8, value, right.string),
        .array => |values| blk: {
            if (values.items.len != right.array.items.len) break :blk false;
            for (values.items, right.array.items) |left_item, right_item| {
                if (!jsonEqual(left_item, right_item)) break :blk false;
            }
            break :blk true;
        },
        .object => |values| blk: {
            if (values.count() != right.object.count()) break :blk false;
            var iterator = values.iterator();
            while (iterator.next()) |entry| {
                const right_value = right.object.get(entry.key_ptr.*) orelse break :blk false;
                if (!jsonEqual(entry.value_ptr.*, right_value)) break :blk false;
            }
            break :blk true;
        },
    };
}

fn jsonTextEqual(allocator: std.mem.Allocator, left: []const u8, right: []const u8) bool {
    const left_parsed = std.json.parseFromSlice(std.json.Value, allocator, left, .{}) catch return false;
    defer left_parsed.deinit();
    const right_parsed = std.json.parseFromSlice(std.json.Value, allocator, right, .{}) catch return false;
    defer right_parsed.deinit();
    return jsonEqual(left_parsed.value, right_parsed.value);
}

fn monotonicNs() i128 {
    var timestamp: std.c.timespec = undefined;
    if (std.c.clock_gettime(.MONOTONIC, &timestamp) != 0) @panic("monotonic clock unavailable");
    return @as(i128, timestamp.sec) * 1_000_000_000 + @as(i128, timestamp.nsec);
}

pub fn main(init: std.process.Init.Minimal) !void {
    var arena = std.heap.ArenaAllocator.init(std.heap.page_allocator);
    defer arena.deinit();
    const allocator = arena.allocator();
    var args = std.process.Args.Iterator.init(init.args);
    _ = args.skip();
    const index = try std.fmt.parseUnsigned(usize, args.next() orelse return error.MissingTestCaseIndex, 10);
    switch (index) {
`)
	for index, testCase := range testCases {
		input, err := json.Marshal(testCase.Input)
		if err != nil {
			return nil, preparedProgram{}, fmt.Errorf("encode input for testcase %q: %w", testCase.ID, err)
		}
		expected, err := expectedText(testCase)
		if err != nil {
			return nil, preparedProgram{}, err
		}
		_, plainText := testCase.Expected.(string)
		fmt.Fprintf(&harness, "        %d => {\n", index)
		fmt.Fprintf(&harness, "    const parsed_%d = try std.json.parseFromSlice(std.json.Value, allocator, &%s, .{});\n", index, zigByteArray(input))
		fmt.Fprintf(&harness, "    defer parsed_%d.deinit();\n    const started_%d = monotonicNs();\n", index, index)
		fmt.Fprintf(&harness, "    const actual_%d = try Solution.solve(allocator, parsed_%d.value);\n", index, index)
		fmt.Fprintf(&harness, "    const solution_time_ns_%d = monotonicNs() - started_%d;\n", index, index)
		fmt.Fprintf(&harness, "    const expected_%d = &%s;\n", index, zigByteArray(expected))
		fmt.Fprintf(&harness, "    const passed_%d = if (%t) std.mem.eql(u8, actual_%d, expected_%d) else jsonTextEqual(allocator, actual_%d, expected_%d);\n", index, plainText, index, index, index, index)
		fmt.Fprintf(&harness, "    std.debug.print(\"\\n{s}:%d:{}:{d}\\n\", .{result_marker, passed_%d, solution_time_ns_%d});\n", index, index, index)
		harness.WriteString("        },\n")
	}
	harness.WriteString("        else => return error.InvalidTestCaseIndex,\n    }\n}\n")

	return map[string]string{"main.zig": harness.String()}, preparedProgram{
		build: []Command{{Name: "zig", Args: []string{"build-exe", "main.zig", "-O", "ReleaseFast", "-lc", "-femit-bin=solution"}}},
		run:   Command{Name: "./solution"},
	}, nil
}
func zigByteArray(value []byte) string {
	parts := make([]string, len(value))
	for i, item := range value {
		parts[i] = strconv.Itoa(int(item))
	}
	return "[_]u8{" + strings.Join(parts, ",") + "}"
}
