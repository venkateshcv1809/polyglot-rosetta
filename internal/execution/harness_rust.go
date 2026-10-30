package execution

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"polyglot-rosetta/internal/domain/testcase"
)

func prepareRust(testCases []testcase.TestCase, marker string) (map[string]string, preparedProgram, error) {
	var harness strings.Builder
	harness.WriteString("mod solution;\nuse solution::Solution;\nuse std::collections::BTreeMap;\nuse std::time::Instant;\n\n")
	harness.WriteString(rustJSONHelpers)
	fmt.Fprintf(&harness, "const RESULT_MARKER: &str = %q;\n\nfn main() {\n    let index = std::env::args().nth(1).and_then(|arg| arg.parse::<usize>().ok()).expect(\"testcase index\");\n    match index {\n", marker)
	for index, testCase := range testCases {
		input, err := json.Marshal(testCase.Input)
		if err != nil {
			return nil, preparedProgram{}, fmt.Errorf("encode input for testcase %q: %w", testCase.ID, err)
		}
		expected, err := expectedText(testCase)
		if err != nil {
			return nil, preparedProgram{}, err
		}
		_, structured := testCase.Expected.(string)
		fmt.Fprintf(&harness,
			"        %d => {\n            let input_text = std::str::from_utf8(&%s).unwrap();\n            let input = JsonParser::new(input_text).parse().expect(\"valid testcase input\");\n            let started = Instant::now();\n            let actual = Solution::solve(&input);\n            let solution_time_ns = started.elapsed().as_nanos();\n            let expected = std::str::from_utf8(&%s).unwrap();\n            let passed = if %t { actual == expected } else { json_equal(&actual, expected) };\n            println!(\"\\n{}:%d:{}:{}\", RESULT_MARKER, passed, solution_time_ns);\n        },\n",
			index, rustByteArray(input), rustByteArray(expected), structured, index)
	}
	harness.WriteString("        _ => panic!(\"invalid testcase index\"),\n    }\n}\n")

	return map[string]string{"main.rs": harness.String()}, preparedProgram{
		build: []Command{{Name: "rustc", Args: []string{"--edition=2021", "-C", "opt-level=3", "main.rs", "-o", "solution"}}},
		run:   Command{Name: "./solution"},
	}, nil
}

const rustJSONHelpers = `
#[derive(Debug, PartialEq)]
pub enum JsonValue {
    Null,
    Bool(bool),
    Number(f64),
    String(String),
    Array(Vec<JsonValue>),
    Object(BTreeMap<String, JsonValue>),
}

struct JsonParser<'a> { bytes: &'a [u8], position: usize }

impl<'a> JsonParser<'a> {
    fn new(text: &'a str) -> Self { Self { bytes: text.as_bytes(), position: 0 } }
    fn parse(mut self) -> Option<JsonValue> {
        let value = self.value()?;
        self.whitespace();
        if self.position == self.bytes.len() { Some(value) } else { None }
    }
    fn value(&mut self) -> Option<JsonValue> {
        self.whitespace();
        match *self.bytes.get(self.position)? {
            b'n' => { self.literal(b"null")?; Some(JsonValue::Null) }
            b't' => { self.literal(b"true")?; Some(JsonValue::Bool(true)) }
            b'f' => { self.literal(b"false")?; Some(JsonValue::Bool(false)) }
            b'"' => Some(JsonValue::String(self.string()?)),
            b'[' => self.array(),
            b'{' => self.object(),
            b'-' | b'0'..=b'9' => self.number(),
            _ => None,
        }
    }
    fn literal(&mut self, expected: &[u8]) -> Option<()> {
        if self.bytes.get(self.position..self.position + expected.len())? != expected { return None; }
        self.position += expected.len(); Some(())
    }
    fn string(&mut self) -> Option<String> {
        self.position += 1;
        let mut output = String::new();
        while let Some(&byte) = self.bytes.get(self.position) {
            self.position += 1;
            match byte {
                b'"' => return Some(output),
                b'\\' => {
                    let escape = *self.bytes.get(self.position)?; self.position += 1;
                    match escape {
                        b'"' => output.push('"'), b'\\' => output.push('\\'), b'/' => output.push('/'),
                        b'b' => output.push('\u{0008}'), b'f' => output.push('\u{000c}'),
                        b'n' => output.push('\n'), b'r' => output.push('\r'), b't' => output.push('\t'),
                        b'u' => {
                            let digits = std::str::from_utf8(self.bytes.get(self.position..self.position + 4)?).ok()?;
                            self.position += 4;
                            output.push(char::from_u32(u32::from_str_radix(digits, 16).ok()?)?);
                        }
                        _ => return None,
                    }
                }
                0..=31 => return None,
                _ if byte.is_ascii() => output.push(byte as char),
                _ => {
                    self.position -= 1;
                    let text = std::str::from_utf8(&self.bytes[self.position..]).ok()?;
                    let character = text.chars().next()?;
                    output.push(character);
                    self.position += character.len_utf8();
                }
            }
        }
        None
    }
    fn number(&mut self) -> Option<JsonValue> {
        let start = self.position;
        while matches!(self.bytes.get(self.position), Some(b'-' | b'+' | b'.' | b'e' | b'E' | b'0'..=b'9')) { self.position += 1; }
        let text = std::str::from_utf8(&self.bytes[start..self.position]).ok()?;
        Some(JsonValue::Number(text.parse().ok()?))
    }
    fn array(&mut self) -> Option<JsonValue> {
        self.position += 1; let mut values = Vec::new(); self.whitespace();
        if self.take(b']') { return Some(JsonValue::Array(values)); }
        loop { values.push(self.value()?); self.whitespace(); if self.take(b']') { break; } if !self.take(b',') { return None; } }
        Some(JsonValue::Array(values))
    }
    fn object(&mut self) -> Option<JsonValue> {
        self.position += 1; let mut values = BTreeMap::new(); self.whitespace();
        if self.take(b'}') { return Some(JsonValue::Object(values)); }
        loop {
            self.whitespace(); if self.bytes.get(self.position) != Some(&b'"') { return None; }
            let key = self.string()?; self.whitespace(); if !self.take(b':') { return None; }
            if values.insert(key, self.value()?).is_some() { return None; }
            self.whitespace(); if self.take(b'}') { break; } if !self.take(b',') { return None; }
        }
        Some(JsonValue::Object(values))
    }
    fn whitespace(&mut self) { while matches!(self.bytes.get(self.position), Some(b' ' | b'\n' | b'\r' | b'\t')) { self.position += 1; } }
    fn take(&mut self, expected: u8) -> bool { if self.bytes.get(self.position) == Some(&expected) { self.position += 1; true } else { false } }
}

fn json_equal(left: &str, right: &str) -> bool {
    match (JsonParser::new(left).parse(), JsonParser::new(right).parse()) {
        (Some(left), Some(right)) => left == right,
        _ => false,
    }
}

`

func rustByteArray(value []byte) string {
	parts := make([]string, len(value))
	for i, item := range value {
		parts[i] = strconv.Itoa(int(item))
	}
	return "[" + strings.Join(parts, ",") + "]"
}
