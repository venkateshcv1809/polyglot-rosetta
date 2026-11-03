use crate::JsonValue;

pub struct Solution;

impl Solution {
    pub fn solve(input_data: &JsonValue) -> String {
        let name = match input_data {
            JsonValue::Object(fields) => match fields.get("name") {
                Some(JsonValue::String(name)) => name.as_str(),
                _ => "World",
            },
            _ => "World",
        };
        let mut greeting = String::with_capacity("Hello ".len() + name.len());
        greeting.push_str("Hello ");
        greeting.push_str(name);
        greeting
    }
}
