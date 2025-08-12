package jsjson

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync"
)

// JSONValue is a dynamic JSON wrapper with error handling
type JSONValue struct {
	data interface{}
	err  error
}

// Error types for better error handling
type JSONError struct {
	Op  string
	Err error
}

func (e *JSONError) Error() string {
	return fmt.Sprintf("jsonjs.%s: %v", e.Op, e.Err)
}

var (
	// Object pool for JSONValue instances to reduce GC pressure
	jsonValuePool = sync.Pool{
		New: func() interface{} {
			return &JSONValue{}
		},
	}
	
	// Byte slice pool for buffer reuse
	bytesPool = sync.Pool{
		New: func() interface{} {
			b := make([]byte, 0, 1024)
			return &b
		},
	}
)

// getJSONValue gets a JSONValue from pool
func getJSONValue() *JSONValue {
	return jsonValuePool.Get().(*JSONValue)
}

// putJSONValue returns a JSONValue to pool
func putJSONValue(jv *JSONValue) {
	jv.data = nil
	jv.err = nil
	jsonValuePool.Put(jv)
}

// getBytesBuffer gets a byte slice from pool
func getBytesBuffer() *[]byte {
	return bytesPool.Get().(*[]byte)
}

// putBytesBuffer returns a byte slice to pool
func putBytesBuffer(b *[]byte) {
	*b = (*b)[:0] // reset length but keep capacity
	bytesPool.Put(b)
}

// -------------------- Core JSON API --------------------

// Parse creates a JSONValue from various input types with optional struct destination
// Usage: Parse(data) or Parse(data, &structDest)
func Parse(v interface{}, dest ...interface{}) JSONValue {
	if v == nil {
		return JSONValue{err: &JSONError{Op: "Parse", Err: fmt.Errorf("input is nil")}}
	}

	// Check if destination struct is provided
	var structDest interface{}
	if len(dest) > 0 {
		structDest = dest[0]
		if structDest != nil {
			// Validate that dest is a pointer
			destType := reflect.TypeOf(structDest)
			if destType.Kind() != reflect.Ptr {
				return JSONValue{err: &JSONError{Op: "Parse", Err: fmt.Errorf("destination must be a pointer, got %T", structDest)}}
			}
		}
	}

	var result interface{}
	var err error
	var jsonBytes []byte

	switch val := v.(type) {
	case string:
		if val == "" {
			return JSONValue{err: &JSONError{Op: "Parse", Err: fmt.Errorf("empty string")}}
		}
		jsonBytes = []byte(val)
	case []byte:
		if len(val) == 0 {
			return JSONValue{err: &JSONError{Op: "Parse", Err: fmt.Errorf("empty byte slice")}}
		}
		jsonBytes = val
	case JSONValue:
		// Already a JSONValue, handle struct destination if provided
		if structDest != nil && val.err == nil {
			if unmarshalErr := val.To(structDest); unmarshalErr != nil {
				return JSONValue{err: &JSONError{Op: "Parse", Err: unmarshalErr}}
			}
		}
		return val
	default:
		// For other types, try to marshal then unmarshal
		var marshalErr error
		jsonBytes, marshalErr = json.Marshal(val)
		if marshalErr != nil {
			return JSONValue{err: &JSONError{Op: "Parse", Err: marshalErr}}
		}
	}

	// If struct destination is provided, unmarshal directly into it
	if structDest != nil {
		err = json.Unmarshal(jsonBytes, structDest)
		if err != nil {
			return JSONValue{err: &JSONError{Op: "Parse", Err: err}}
		}
		// Also parse into generic interface{} for JSONValue functionality
		err = json.Unmarshal(jsonBytes, &result)
		if err != nil {
			return JSONValue{err: &JSONError{Op: "Parse", Err: err}}
		}
		return JSONValue{data: result}
	}

	// Standard parsing into interface{}
	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		return JSONValue{err: &JSONError{Op: "Parse", Err: err}}
	}

	return JSONValue{data: result}
}

// ParseInto directly parses JSON data into a struct with better performance
// This is more efficient than Parse + To for struct unmarshaling
func ParseInto(data interface{}, dest interface{}) error {
	if dest == nil {
		return &JSONError{Op: "ParseInto", Err: fmt.Errorf("destination cannot be nil")}
	}

	destType := reflect.TypeOf(dest)
	if destType.Kind() != reflect.Ptr {
		return &JSONError{Op: "ParseInto", Err: fmt.Errorf("destination must be a pointer, got %T", dest)}
	}

	var jsonBytes []byte
	var err error

	switch val := data.(type) {
	case string:
		if val == "" {
			return &JSONError{Op: "ParseInto", Err: fmt.Errorf("empty string")}
		}
		jsonBytes = []byte(val)
	case []byte:
		if len(val) == 0 {
			return &JSONError{Op: "ParseInto", Err: fmt.Errorf("empty byte slice")}
		}
		jsonBytes = val
	case JSONValue:
		if val.err != nil {
			return &JSONError{Op: "ParseInto", Err: val.err}
		}
		return val.To(dest)
	default:
		jsonBytes, err = json.Marshal(val)
		if err != nil {
			return &JSONError{Op: "ParseInto", Err: err}
		}
	}

	err = json.Unmarshal(jsonBytes, dest)
	if err != nil {
		return &JSONError{Op: "ParseInto", Err: err}
	}

	return nil
}

// MustParse is like Parse but panics on error
func MustParse(v interface{}, dest ...interface{}) JSONValue {
	result := Parse(v, dest...)
	if result.err != nil {
		panic(result.err)
	}
	return result
}

// MustParseInto is like ParseInto but panics on error
func MustParseInto(data interface{}, dest interface{}) {
	if err := ParseInto(data, dest); err != nil {
		panic(err)
	}
}

// Stringify converts a value to JSON string
func Stringify(v interface{}) (string, error) {
	if v == nil {
		return "null", nil
	}

	// Handle JSONValue type
	if jv, ok := v.(JSONValue); ok {
		if jv.err != nil {
			return "", jv.err
		}
		v = jv.data
	}

	// Use buffer pool for better performance
	buffer := getBytesBuffer()
	defer putBytesBuffer(buffer)

	// Reset buffer and grow if needed
	if cap(*buffer) < 512 {
		*buffer = make([]byte, 0, 1024)
	}

	encoder := json.NewEncoder(&bytesWriter{buffer})
	err := encoder.Encode(v)
	if err != nil {
		return "", &JSONError{Op: "Stringify", Err: err}
	}

	// Remove trailing newline that encoder adds
	result := *buffer
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}

	return string(result), nil
}

// StringifyPretty converts a value to pretty-printed JSON string
func StringifyPretty(v interface{}, indent string) (string, error) {
	if v == nil {
		return "null", nil
	}

	if jv, ok := v.(JSONValue); ok {
		if jv.err != nil {
			return "", jv.err
		}
		v = jv.data
	}

	bytes, err := json.MarshalIndent(v, "", indent)
	if err != nil {
		return "", &JSONError{Op: "StringifyPretty", Err: err}
	}
	return string(bytes), nil
}

// -------------------- JSONValue Methods --------------------

// IsValid checks if the JSONValue is valid (no errors)
func (j JSONValue) IsValid() bool {
	return j.err == nil
}

// Error returns the error if any
func (j JSONValue) Error() error {
	return j.err
}

// Get allows nested access with error propagation
func (j JSONValue) Get(keys ...interface{}) JSONValue {
	if j.err != nil {
		return j // Propagate existing error
	}

	if len(keys) == 0 {
		return j
	}

	current := j.data
	for i, key := range keys {
		if current == nil {
			return JSONValue{err: &JSONError{
				Op:  "Get",
				Err: fmt.Errorf("cannot access key %v on nil value at position %d", key, i),
			}}
		}

		switch c := current.(type) {
		case map[string]interface{}:
			keyStr, ok := key.(string)
			if !ok {
				return JSONValue{err: &JSONError{
					Op:  "Get",
					Err: fmt.Errorf("key must be string for object access, got %T at position %d", key, i),
				}}
			}
			var exists bool
			current, exists = c[keyStr]
			if !exists {
				return JSONValue{err: &JSONError{
					Op:  "Get",
					Err: fmt.Errorf("key %q not found at position %d", keyStr, i),
				}}
			}

		case []interface{}:
			idx, err := convertToIndex(key)
			if err != nil {
				return JSONValue{err: &JSONError{
					Op:  "Get",
					Err: fmt.Errorf("invalid array index %v at position %d: %v", key, i, err),
				}}
			}
			if idx < 0 || idx >= len(c) {
				return JSONValue{err: &JSONError{
					Op:  "Get",
					Err: fmt.Errorf("array index %d out of bounds (length: %d) at position %d", idx, len(c), i),
				}}
			}
			current = c[idx]

		default:
			return JSONValue{err: &JSONError{
				Op:  "Get",
				Err: fmt.Errorf("cannot access key %v on type %T at position %d", key, current, i),
			}}
		}
	}

	return JSONValue{data: current}
}

// GetOr returns the value at the given keys or the default value if not found/error
func (j JSONValue) GetOr(defaultValue interface{}, keys ...interface{}) interface{} {
	result := j.Get(keys...)
	if result.err != nil {
		return defaultValue
	}
	return result.data
}

// Has checks if a key path exists
func (j JSONValue) Has(keys ...interface{}) bool {
	return j.Get(keys...).IsValid()
}

// -------------------- Type Conversion Methods --------------------

// String returns the value as string with error handling
func (j JSONValue) String() (string, error) {
	if j.err != nil {
		return "", j.err
	}

	switch v := j.data.(type) {
	case string:
		return v, nil
	case nil:
		return "", nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// StringOr returns the value as string or default if error/not string
func (j JSONValue) StringOr(defaultVal string) string {
	s, err := j.String()
	if err != nil || s == "" {
		return defaultVal
	}
	return s
}

// Int returns the value as int
func (j JSONValue) Int() (int, error) {
	if j.err != nil {
		return 0, j.err
	}

	switch v := j.data.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	case string:
		if i, err := strconv.Atoi(v); err == nil {
			return i, nil
		}
		return 0, &JSONError{Op: "Int", Err: fmt.Errorf("cannot convert string %q to int", v)}
	case nil:
		return 0, nil
	default:
		return 0, &JSONError{Op: "Int", Err: fmt.Errorf("cannot convert %T to int", v)}
	}
}

// IntOr returns the value as int or default if error/conversion fails
func (j JSONValue) IntOr(defaultValue int) int {
	if i, err := j.Int(); err == nil {
		return i
	}
	return defaultValue
}

// Float64 returns the value as float64
func (j JSONValue) Float64() (float64, error) {
	if j.err != nil {
		return 0, j.err
	}

	switch v := j.data.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, nil
		}
		return 0, &JSONError{Op: "Float64", Err: fmt.Errorf("cannot convert string %q to float64", v)}
	case nil:
		return 0, nil
	default:
		return 0, &JSONError{Op: "Float64", Err: fmt.Errorf("cannot convert %T to float64", v)}
	}
}

// Float64Or returns the value as float64 or default if error/conversion fails
func (j JSONValue) Float64Or(defaultValue float64) float64 {
	if f, err := j.Float64(); err == nil {
		return f
	}
	return defaultValue
}

// Bool returns the value as bool
func (j JSONValue) Bool() (bool, error) {
	if j.err != nil {
		return false, j.err
	}

	switch v := j.data.(type) {
	case bool:
		return v, nil
	case string:
		if b, err := strconv.ParseBool(v); err == nil {
			return b, nil
		}
		return false, &JSONError{Op: "Bool", Err: fmt.Errorf("cannot convert string %q to bool", v)}
	case float64:
		return v != 0, nil
	case nil:
		return false, nil
	default:
		return false, &JSONError{Op: "Bool", Err: fmt.Errorf("cannot convert %T to bool", v)}
	}
}

// BoolOr returns the value as bool or default if error/conversion fails
func (j JSONValue) BoolOr(defaultValue bool) bool {
	if b, err := j.Bool(); err == nil {
		return b
	}
	return defaultValue
}

// Array returns the value as []JSONValue for iteration
func (j JSONValue) Array() ([]JSONValue, error) {
	if j.err != nil {
		return nil, j.err
	}

	arr, ok := j.data.([]interface{})
	if !ok {
		return nil, &JSONError{Op: "Array", Err: fmt.Errorf("value is not an array, got %T", j.data)}
	}

	result := make([]JSONValue, len(arr))
	for i, item := range arr {
		result[i] = JSONValue{data: item}
	}
	return result, nil
}

// Object returns the value as map[string]JSONValue for iteration
func (j JSONValue) Object() (map[string]JSONValue, error) {
	if j.err != nil {
		return nil, j.err
	}

	obj, ok := j.data.(map[string]interface{})
	if !ok {
		return nil, &JSONError{Op: "Object", Err: fmt.Errorf("value is not an object, got %T", j.data)}
	}

	result := make(map[string]JSONValue, len(obj))
	for key, value := range obj {
		result[key] = JSONValue{data: value}
	}
	return result, nil
}

// Raw returns the underlying Go value
func (j JSONValue) Raw() interface{} {
	if j.err != nil {
		return nil
	}
	return j.data
}

// IsNull checks if the value is null
func (j JSONValue) IsNull() bool {
	return j.err == nil && j.data == nil
}

// Type returns the JSON type as a string
func (j JSONValue) Type() string {
	if j.err != nil {
		return "error"
	}

	switch j.data.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case float64:
		return "number"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return "unknown"
	}
}

// -------------------- Enhanced To Method --------------------

// To unmarshals the JSONValue data into the provided destination with improved performance
func (j JSONValue) To(dest interface{}) error {
	if j.err != nil {
		return &JSONError{Op: "To", Err: j.err}
	}

	if dest == nil {
		return &JSONError{Op: "To", Err: fmt.Errorf("destination cannot be nil")}
	}

	// Direct assignment for simple cases to avoid marshal/unmarshal overhead
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return &JSONError{Op: "To", Err: fmt.Errorf("destination must be a pointer, got %T", dest)}
	}

	destElem := destValue.Elem()
	
	// Try direct assignment for compatible types
	if j.data != nil && destElem.CanSet() {
		srcValue := reflect.ValueOf(j.data)
		if srcValue.Type().AssignableTo(destElem.Type()) {
			destElem.Set(srcValue)
			return nil
		}
	}

	// Fall back to JSON marshal/unmarshal for complex types
	buffer := getBytesBuffer()
	defer putBytesBuffer(buffer)

	// Reset buffer
	*buffer = (*buffer)[:0]

	// Use a bytes writer for better performance
	encoder := json.NewEncoder(&bytesWriter{buffer})
	if err := encoder.Encode(j.data); err != nil {
		return &JSONError{Op: "To", Err: fmt.Errorf("failed to marshal data: %w", err)}
	}

	// Remove trailing newline
	if len(*buffer) > 0 && (*buffer)[len(*buffer)-1] == '\n' {
		*buffer = (*buffer)[:len(*buffer)-1]
	}

	// Unmarshal into the destination
	if err := json.Unmarshal(*buffer, dest); err != nil {
		return &JSONError{Op: "To", Err: fmt.Errorf("failed to unmarshal into destination: %w", err)}
	}

	return nil
}

// MustTo is like To but panics on error
func (j JSONValue) MustTo(dest interface{}) {
	if err := j.To(dest); err != nil {
		panic(err)
	}
}

// -------------------- Utility Functions --------------------

// convertToIndex converts various types to array index
func convertToIndex(key interface{}) (int, error) {
	switch v := key.(type) {
	case int:
		return v, nil
	case string:
		return strconv.Atoi(v)
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("cannot convert %T to array index", key)
	}
}

// bytesWriter implements io.Writer for efficient byte slice writing
type bytesWriter struct {
	buf *[]byte
}

func (w *bytesWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}

// -------------------- Convenience Functions --------------------

// Valid creates a JSONValue from a Go value (no parsing)
func Valid(data interface{}) JSONValue {
	return JSONValue{data: data}
}

// Invalid creates a JSONValue with an error
func Invalid(err error) JSONValue {
	return JSONValue{err: &JSONError{Op: "Invalid", Err: err}}
}

// Clone creates a deep copy of the JSONValue
func (j JSONValue) Clone() JSONValue {
	if j.err != nil {
		return j
	}

	// Use buffer pool for better performance
	buffer := getBytesBuffer()
	defer putBytesBuffer(buffer)

	*buffer = (*buffer)[:0]
	encoder := json.NewEncoder(&bytesWriter{buffer})
	if err := encoder.Encode(j.data); err != nil {
		return JSONValue{err: &JSONError{Op: "Clone", Err: err}}
	}

	// Remove trailing newline
	if len(*buffer) > 0 && (*buffer)[len(*buffer)-1] == '\n' {
		*buffer = (*buffer)[:len(*buffer)-1]
	}

	return Parse(*buffer)
}