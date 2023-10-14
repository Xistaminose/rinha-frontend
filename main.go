package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"syscall/js"
)

type JsonValue struct {
	Key    string
	Value  interface{}
	Indent int
}

func parse(input string) (JsonValue, error) {
	var jsValue interface{}
	err := json.Unmarshal([]byte(input), &jsValue)
	if err != nil {
		return JsonValue{}, err
	}
	return JsonValue{Value: jsValue, Indent: 0}, nil
}

func mountRows(jsValue JsonValue, result *[]string) {
	switch v := jsValue.Value.(type) {
	case map[string]interface{}:
		for key, val := range v {
			*result = append(*result, formatRow(JsonValue{Key: key, Value: val, Indent: jsValue.Indent}))
			mountRows(JsonValue{Key: key, Value: val, Indent: jsValue.Indent + 1}, result)
		}
	case []interface{}:
		*result = append(*result, formatRow(JsonValue{Key: "", Value: "[", Indent: jsValue.Indent}))
		for i, val := range v {
			*result = append(*result, formatRow(JsonValue{Key: strconv.Itoa(i), Value: val, Indent: jsValue.Indent + 1}))
			mountRows(JsonValue{Key: strconv.Itoa(i), Value: val, Indent: jsValue.Indent + 2}, result)
		}
	default:
		// base case (int, string, etc.) - do nothing
	}
}

func formatRow(jsValue JsonValue) string {
	return fmt.Sprintf("%s\x1F%s\x1F%d", jsValue.Key, fmt.Sprintf("%v", jsValue.Value), jsValue.Indent)
}

// ParseAndFormatJSON accepts JSON string, parses, and formats it, then passes it to JavaScript function by callback.
func ParseAndFormatJSON(this js.Value, p []js.Value) interface{} {
	callback := p[1]
	inputJSON := p[0].String()

	jsValue, err := parse(inputJSON)
	if err != nil {
		callback.Invoke(err.Error(), js.Null())
		return nil
	}

	var result []string
	mountRows(jsValue, &result)
	callback.Invoke(js.Null(), strings.Join(result, "\x1E"))
	return nil
}

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("parseAndFormatJSON", js.FuncOf(ParseAndFormatJSON))
	<-c
}
