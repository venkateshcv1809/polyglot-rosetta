package execution

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Rust solutions retain the existing JsonValue interface. Construct its values
// from validated catalog JSON instead of shipping another JSON parser in Rust.
func rustInput(input []byte) (string, error) {
	var value any
	decoder := json.NewDecoder(bytes.NewReader(input))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return "", err
	}
	return rustValue(value), nil
}

func rustValue(value any) string {
	switch value := value.(type) {
	case nil:
		return "JsonValue::Null"
	case bool:
		return fmt.Sprintf("JsonValue::Bool(%t)", value)
	case json.Number:
		return "JsonValue::Number(" + value.String() + "_f64)"
	case string:
		return "JsonValue::String(" + rustString(value) + ")"
	case []any:
		items := make([]string, len(value))
		for i, item := range value {
			items[i] = rustValue(item)
		}
		return "JsonValue::Array(vec![" + strings.Join(items, ",") + "])"
	case map[string]any:
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		items := make([]string, len(keys))
		for i, key := range keys {
			items[i] = "(" + rustString(key) + "," + rustValue(value[key]) + ")"
		}
		return "JsonValue::Object(BTreeMap::from([" + strings.Join(items, ",") + "]))"
	default:
		panic(fmt.Sprintf("unexpected decoded JSON type %T", value))
	}
}

func rustString(value string) string {
	return "String::from_utf8(vec![" + byteList([]byte(value)) + "]).unwrap()"
}

func byteList(value []byte) string {
	items := make([]string, len(value))
	for i, b := range value {
		items[i] = strconv.Itoa(int(b))
	}
	return strings.Join(items, ",")
}
