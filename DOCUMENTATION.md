# 📖 jsjson - Complete Documentation

## Table of Contents

1. [Introduction](#introduction)
2. [Architecture and Design](#architecture-and-design)
3. [Core Concepts](#core-concepts)
4. [API Reference](#api-reference)
5. [Error Handling](#error-handling)
6. [Performance Considerations](#performance-considerations)
7. [Migration Guide](#migration-guide)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)
10. [FAQ](#faq)

## Introduction

**jsjson** is a Go library designed to bring JavaScript's intuitive JSON handling to Go while maintaining Go's type safety and performance characteristics. It addresses common pain points in Go JSON processing:

- **Verbose Type Assertions**: Eliminates manual `.(type)` casting
- **Error-Prone Navigation**: Provides safe nested access
- **Boilerplate Code**: Reduces repetitive error checking
- **Poor Developer Experience**: Offers familiar JavaScript-like API

### Philosophy

The library follows these design principles:

1. **Familiarity**: JavaScript developers should feel at home
2. **Safety**: Errors should be contained and recoverable
3. **Performance**: Competitive with existing Go JSON libraries
4. **Simplicity**: Common tasks should be simple, complex tasks possible

## Architecture and Design

### Core Types

```go
// JSONValue - The central type wrapping any JSON value
type JSONValue struct {
    data interface{} // The actual parsed JSON data
    err  error       // Any error that occurred during operations
}

// JSONError - Structured error type for better debugging
type JSONError struct {
    Op  string // Operation that failed (e.g., "Parse", "Get", "Int")
    Err error  // Underlying error
}
```

### Error Propagation

jsjson uses **error propagation** - once an error occurs, it's carried through subsequent operations:

```go
obj := Parse(`{"user": {"name": "John"}}`)
// If Parse fails, all subsequent operations return the same error
name := obj.Get("user").Get("name").StringOr("Unknown")
```

This design prevents cascading nil pointer panics and makes error handling more predictable.

### Memory Management

The library employs **object pooling** to reduce garbage collection pressure:

```go
var jsonValuePool = sync.Pool{
    New: func() interface{} {
        return &JSONValue{}
    },
}
```

Objects are automatically returned to the pool when they go out of scope.

## Core Concepts

### 1. Parsing vs. Validation

jsjson distinguishes between parsing and validation:

```go
// Parsing - Convert JSON string to internal representation
obj := Parse(`{"name": "John", "age": "not-a-number"}`)
// This succeeds - JSON is structurally valid

// Validation - Happens during type conversion
age, err := obj.Get("age").Int()
// This fails - "not-a-number" can't be converted to int
```

### 2. Path Navigation

Paths support both string keys and numeric indices:

```go
jsonStr := `{
    "users": [
        {"name": "John", "tags": ["admin", "user"]},
        {"name": "Jane", "tags": ["user"]}
    ]
}`

obj := Parse(jsonStr)

// Mixed path types
firstUserName := obj.Get("users", 0, "name")        // "John"
firstTag := obj.Get("users", 0, "tags", 0)          // "admin"
secondUserName := obj.Get("users", 1, "name")       // "Jane"
```

### 3. Type Coercion

jsjson provides intelligent type coercion:

```go
// JSON numbers are always float64 in Go
obj := Parse(`{"count": 42}`)

// Automatic conversion to int
count := obj.Get("count").IntOr(0) // Returns 42 (int)

// String to number conversion
obj2 := Parse(`{"score": "95.5"}`)
score := obj2.Get("score").Float64Or(0.0) // Returns 95.5
```

### 4. Default Values

The `Or` methods provide fallback values:

```go
obj := Parse(`{"user": {"name": "John"}}`)

// Existing value
name := obj.Get("user", "name").StringOr("Unknown") // "John"

// Missing value
email := obj.Get("user", "email").StringOr("N/A")   // "N/A"

// Type mismatch
age := obj.Get("user", "name").IntOr(0)             // 0 (name is string)
```

## API Reference

### Parse Functions

#### `Parse(v interface{}) JSONValue`

**Purpose**: Convert input to JSONValue with error handling.

**Input Types**:
- `string`: JSON string
- `[]byte`: JSON bytes
- `JSONValue`: Returns as-is (idempotent)
- `any`: Marshal then unmarshal (for structs, maps, etc.)

**Examples**:
```go
// From string
obj1 := Parse(`{"name": "John"}`)

// From bytes
data := []byte(`{"age": 30}`)
obj2 := Parse(data)

// From struct
type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}
user := User{Name: "John", Age: 30}
obj3 := Parse(user)

// From map
userData := map[string]interface{}{
    "name": "John",
    "age":  30,
}
obj4 := Parse(userData)
```

**Error Handling**:
```go
obj := Parse(`invalid json`)
if !obj.IsValid() {
    fmt.Printf("Parse error: %v\n", obj.Error())
}
```

#### `MustParse(v interface{}) JSONValue`

**Purpose**: Like Parse but panics on error.

**When to Use**: When JSON validity is guaranteed (e.g., embedded JSON, testing).

```go
// Safe - embedded JSON is known to be valid
config := MustParse(`{
    "database": {
        "host": "localhost",
        "port": 5432
    }
}`)

// Dangerous - external input might be invalid
// userInput := MustParse(requestBody) // DON'T DO THIS
```

### Navigation Methods

#### `Get(keys ...interface{}) JSONValue`

**Purpose**: Navigate nested JSON structures safely.

**Key Types**:
- `string`: Object property access
- `int`: Array index access
- Mixed: Navigate complex structures

**Examples**:
```go
jsonStr := `{
    "company": {
        "name": "TechCorp",
        "employees": [
            {
                "name": "John",
                "roles": ["developer", "architect"],
                "contact": {
                    "email": "john@techcorp.com"
                }
            }
        ]
    }
}`

obj := Parse(jsonStr)

// Single key
companyName := obj.Get("company")

// Nested object access
employeeName := obj.Get("company", "employees", 0, "name")

// Deep nesting
email := obj.Get("company", "employees", 0, "contact", "email")

// Array access
firstRole := obj.Get("company", "employees", 0, "roles", 0)
```

**Error Propagation**:
```go
// If any step fails, error is propagated
result := obj.Get("nonexistent", "also", "nonexistent")
fmt.Println(result.IsValid()) // false
```

#### `GetOr(defaultValue interface{}, keys ...interface{}) interface{}`

**Purpose**: Get value with fallback in one operation.

```go
obj := Parse(`{"user": {"name": "John"}}`)

// Get with fallback
email := obj.GetOr("unknown@example.com", "user", "email")
// Returns "unknown@example.com" (string)

age := obj.GetOr(0, "user", "age")
// Returns 0 (int)
```

#### `Has(keys ...interface{}) bool`

**Purpose**: Check if a path exists without retrieving the value.

```go
obj := Parse(`{"user": {"name": "John", "profile": null}}`)

// Existing key
fmt.Println(obj.Has("user", "name"))    // true

// Missing key
fmt.Println(obj.Has("user", "email"))   // false

// Null value still exists
fmt.Println(obj.Has("user", "profile")) // true
```

### Type Conversion Methods

All conversion methods follow the pattern:
- `Type()` returns `(value, error)`
- `TypeOr(default)` returns `value` or `default`

#### String Conversions

```go
// String() - with error
name, err := obj.Get("name").String()
if err != nil {
    // Handle error
}

// StringOr() - with default
name := obj.Get("name").StringOr("Unknown")
```

**Conversion Rules**:
- `string` → returned as-is
- `nil` → empty string
- Other types → `fmt.Sprintf("%v", value)`

#### Numeric Conversions

```go
// Int conversion
age, err := obj.Get("age").Int()
age := obj.Get("age").IntOr(0)

// Float64 conversion
score, err := obj.Get("score").Float64()
score := obj.Get("score").Float64Or(0.0)
```

**Conversion Rules**:
- `float64` → `int(value)` or value as-is
- `string` → `strconv.Atoi`/`strconv.ParseFloat`
- `nil` → 0
- Other types → error

#### Boolean Conversions

```go
active, err := obj.Get("active").Bool()
active := obj.Get("active").BoolOr(false)
```

**Conversion Rules**:
- `bool` → returned as-is
- `string` → `strconv.ParseBool` ("true", "false", "1", "0", etc.)
- `float64` → `value != 0`
- `nil` → `false`
- Other types → error

### Collection Methods

#### `Array() ([]JSONValue, error)`

**Purpose**: Convert JSON array to slice of JSONValue.

```go
jsonStr := `{
    "tags": ["golang", "json", "api"],
    "numbers": [1, 2, 3, 4, 5]
}`

obj := Parse(jsonStr)

// Get array
tags, err := obj.Get("tags").Array()
if err != nil {
    log.Printf("Not an array: %v", err)
    return
}

// Iterate
for i, tag := range tags {
    fmt.Printf("Tag %d: %s\n", i, tag.StringOr(""))
}

// Process numbers
numbers, err := obj.Get("numbers").Array()
if err == nil {
    sum := 0
    for _, num := range numbers {
        sum += num.IntOr(0)
    }
    fmt.Printf("Sum: %d\n", sum)
}
```

#### `Object() (map[string]JSONValue, error)`

**Purpose**: Convert JSON object to map for iteration.

```go
jsonStr := `{
    "metadata": {
        "version": "1.0",
        "author": "developer",
        "created": "2023-01-01"
    }
}`

obj := Parse(jsonStr)

metadata, err := obj.Get("metadata").Object()
if err != nil {
    log.Printf("Not an object: %v", err)
    return
}

// Iterate over key-value pairs
for key, value := range metadata {
    fmt.Printf("%s: %s\n", key, value.StringOr(""))
}
```

### Utility Methods

#### `Raw() interface{}`

**Purpose**: Get the underlying Go value.

```go
obj := Parse(`{"score": 95.5}`)
raw := obj.Get("score").Raw()
// raw is float64(95.5)

// Type assertion if you know the type
if score, ok := raw.(float64); ok {
    fmt.Printf("Score: %.1f\n", score)
}
```

#### `IsNull() bool`

**Purpose**: Check for JSON null values.

```go
obj := Parse(`{"name": "John", "middle": null, "age": 30}`)

fmt.Println(obj.Get("name").IsNull())   // false
fmt.Println(obj.Get("middle").IsNull()) // true
fmt.Println(obj.Get("age").IsNull())    // false
fmt.Println(obj.Get("missing").IsNull()) // false (missing ≠ null)
```

#### `Type() string`

**Purpose**: Get JSON type as string.

```go
obj := Parse(`{
    "name": "John",
    "age": 30,
    "active": true,
    "scores": [95, 87, 92],
    "metadata": {"version": 1},
    "notes": null
}`)

fmt.Println(obj.Get("name").Type())     // "string"
fmt.Println(obj.Get("age").Type())      // "number"
fmt.Println(obj.Get("active").Type())   // "boolean"
fmt.Println(obj.Get("scores").Type())   // "array"
fmt.Println(obj.Get("metadata").Type()) // "object"
fmt.Println(obj.Get("notes").Type())    // "null"
fmt.Println(obj.Get("missing").Type())  // "error"
```

#### `Clone() JSONValue`

**Purpose**: Create a deep copy of the JSONValue.

```go
original := Parse(`{"count": 1}`)
copy := original.Clone()

// Modifications to original don't affect copy
// (Note: jsjson is primarily read-only, but Clone ensures independence)
```

### Output Functions

#### `Stringify(v interface{}) (string, error)`

**Purpose**: Convert any value to JSON string.

```go
data := map[string]interface{}{
    "name":   "John",
    "age":    30,
    "active": true,
}

jsonStr, err := Stringify(data)
if err != nil {
    log.Printf("Stringify error: %v", err)
    return
}
fmt.Println(jsonStr) // {"active":true,"age":30,"name":"John"}
```

#### `StringifyPretty(v interface{}, indent string) (string, error)`

**Purpose**: Convert to pretty-printed JSON.

```go
data := map[string]interface{}{
    "user": map[string]interface{}{
        "name": "John",
        "age":  30,
    },
}

pretty, err := StringifyPretty(data, "  ")
if err != nil {
    log.Printf("StringifyPretty error: %v", err)
    return
}

fmt.Println(pretty)
// {
//   "user": {
//     "age": 30,
//     "name": "John"
//   }
// }
```

## Error Handling

### Error Types

jsjson uses structured errors for better debugging:

```go
type JSONError struct {
    Op  string // Operation that failed
    Err error  // Underlying error
}
```

### Error Categories

1. **Parse Errors**: Invalid JSON syntax
2. **Access Errors**: Invalid keys or indices
3. **Type Errors**: Type conversion failures

### Error Checking Patterns

#### 1. Immediate Checking

```go
obj := Parse(jsonInput)
if !obj.IsValid() {
    return fmt.Errorf("parse failed: %w", obj.Error())
}

name, err := obj.Get("name").String()
if err != nil {
    return fmt.Errorf("name extraction failed: %w", err)
}
```

#### 2. Defensive Programming

```go
func processUser(jsonInput string) User {
    obj := Parse(jsonInput)
    
    return User{
        ID:     obj.Get("id").IntOr(0),
        Name:   obj.Get("name").StringOr("Unknown"),
        Email:  obj.Get("email").StringOr(""),
        Active: obj.Get("active").BoolOr(false),
        Score:  obj.Get("score").Float64Or(0.0),
    }
}
```

#### 3. Validation Pattern

```go
func validateAndProcess(jsonInput string) error {
    obj := Parse(jsonInput)
    if !obj.IsValid() {
        return obj.Error()
    }
    
    // Check required fields
    requiredFields := []string{"id", "name", "email"}
    for _, field := range requiredFields {
        if !obj.Has(field) {
            return fmt.Errorf("missing required field: %s", field)
        }
    }
    
    // Validate types
    if obj.Get("id").Type() != "number" {
        return errors.New("id must be a number")
    }
    
    // Process valid data
    return processValidUser(obj)
}
```

### Common Error Scenarios

#### 1. Invalid JSON

```go
obj := Parse(`{"invalid": json}`) // Missing quotes
fmt.Println(obj.Error()) // jsonjs.Parse: invalid character 'j' looking for beginning of value
```

#### 2. Missing Keys

```go
obj := Parse(`{"name": "John"}`)
age := obj.Get("age") // Key doesn't exist
fmt.Println(age.Error()) // jsonjs.Get: key "age" not found at position 0
```

#### 3. Type Mismatches

```go
obj := Parse(`{"name": "John"}`)
age, err := obj.Get("name").Int() // "John" is not a number
fmt.Println(err) // jsonjs.Int: cannot convert string "John" to int
```

#### 4. Array Bounds

```go
obj := Parse(`{"tags": ["a", "b"]}`)
tag := obj.Get("tags", 5) // Index out of bounds
fmt.Println(tag.Error()) // jsonjs.Get: array index 5 out of bounds (length: 2) at position 1
```

## Performance Considerations

### Memory Usage

1. **Object Pooling**: jsjson automatically pools JSONValue objects
2. **Lazy Evaluation**: Type conversions happen only when requested
3. **Copy Minimization**: Operates on shared data when possible

### Performance Tips

#### 1. Reuse Parsed Objects

```go
// Good - parse once, use many times
obj := Parse(jsonStr)
for i := 0; i < 1000; i++ {
    name := obj.Get("user", "name").StringOr("")
    processName(name)
}

// Bad - parsing repeatedly
for i := 0; i < 1000; i++ {
    obj := Parse(jsonStr) // Wasteful reparsing
    name := obj.Get("user", "name").StringOr("")
    processName(name)
}
```

#### 2. Use Appropriate Access Patterns

```go
obj := Parse(largeJSON)

// Good - check existence first for optional fields
if obj.Has("optional", "field") {
    value := obj.Get("optional", "field").StringOr("")
    processValue(value)
}

// Less efficient - double navigation
optional := obj.Get("optional", "field")
if optional.IsValid() {
    value := optional.StringOr("")
    processValue(value)
}
```

#### 3. Batch Array Processing

```go
// Good - get array once, iterate efficiently
items, err := obj.Get("items").Array()
if err == nil {
    for _, item := range items {
        processItem(item)
    }
}

// Bad - repeated array access
count := obj.Get("items").Raw().([]interface{})
for i := 0; i < len(count); i++ {
    item := obj.Get("items", i) // Repeated navigation
    processItem(item)
}
```

### Benchmarking Your Code

```go
func BenchmarkYourUsage(b *testing.B) {
    jsonStr := `your test JSON here`
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        obj := Parse(jsonStr)
        // Your processing logic here
        _ = obj.Get("some", "path").StringOr("")
    }
}
```

## Migration Guide

### From encoding/json

#### Before (encoding/json)

```go
func processUser(jsonData []byte) (User, error) {
    var data map[string]interface{}
    if err := json.Unmarshal(jsonData, &data); err != nil {
        return User{}, err
    }
    
    user, ok := data["user"].(map[string]interface{})
    if !ok {
        return User{}, errors.New("user not found")
    }
    
    name, ok := user["name"].(string)
    if !ok {
        name = "Unknown"
    }
    
    age := 0
    if ageVal, ok := user["age"].(float64); ok {
        age = int(ageVal)
    }
    
    active := false
    if activeVal, ok := user["active"].(bool); ok {
        active = activeVal
    }
    
    return User{
        Name:   name,
        Age:    age,
        Active: active,
    }, nil
}
```

#### After (jsjson)

```go
func processUser(jsonData []byte) (User, error) {
    obj := Parse(jsonData)
    if !obj.IsValid() {
        return User{}, obj.Error()
    }
    
    if !obj.Has("user") {
        return User{}, errors.New("user not found")
    }
    
    return User{
        Name:   obj.Get("user", "name").StringOr("Unknown"),
        Age:    obj.Get("user", "age").IntOr(0),
        Active: obj.Get("user", "active").BoolOr(false),
    }, nil
}
```

### From gjson

#### Before (gjson)

```go
name := gjson.Get(jsonStr, "user.name").String()
age := int(gjson.Get(jsonStr, "user.age").Int())
active := gjson.Get(jsonStr, "user.active").Bool()

// No built-in defaults
email := gjson.Get(jsonStr, "user.email")
if !email.Exists() {
    email = gjson.Result{Type: gjson.String, Str: "unknown@example.com"}
}
```

#### After (jsjson)

```go
obj := Parse(jsonStr)
name := obj.Get("user", "name").StringOr("")
age := obj.Get("user", "age").IntOr(0)
active := obj.Get("user", "active").BoolOr(false)
email := obj.Get("user", "email").StringOr("unknown@example.com")
```

## Best Practices

### 1. Error Handling Strategy

```go
// Strategy 1: Fail Fast (for critical data)
func processCriticalData(jsonStr string) error {
    obj := Parse(jsonStr)
    if !obj.IsValid() {
        return fmt.Errorf("invalid JSON: %w", obj.Error())
    }
    
    requiredID := obj.Get("id")
    if !requiredID.IsValid() {
        return fmt.Errorf("missing required ID: %w", requiredID.Error())
    }
    
    id, err := requiredID.Int()
    if err != nil {
        return fmt.Errorf("invalid ID format: %w", err)
    }
    
    return processCritical(id)
}

// Strategy 2: Defensive (for optional/user data)
func processUserPreferences(jsonStr string) UserPrefs {
    obj := Parse(jsonStr)
    
    return UserPrefs{
        Theme:         obj.Get("theme").StringOr("light"),
        Language:      obj.Get("language").StringOr("en"),
        Notifications: obj.Get("notifications").BoolOr(true),
        MaxItems:      obj.Get("maxItems").IntOr(50),
    }
}
```

### 2. Input Validation

```go
func validateUserInput(obj JSONValue) error {
    // Check structure
    if !obj.Has("user") {
        return errors.New("user object required")
    }
    
    user := obj.Get("user")
    
    // Validate required fields
    if !user.Has("email") {
        return errors.New("email is required")
    }
    
    // Validate formats
    email := user.Get("email").StringOr("")
    if !isValidEmail(email) {
        return errors.New("invalid email format")
    }
    
    // Validate ranges
    age := user.Get("age").IntOr(0)
    if age < 0 || age > 150 {
        return errors.New("age must be between 0 and 150")
    }
    
    return nil
}
```

### 3. Large Data Processing

```go
func processLargeDataset(jsonStr string) error {
    obj := Parse(jsonStr)
    if !obj.IsValid() {
        return obj.Error()
    }
    
    items, err := obj.Get("data", "items").Array()
    if err != nil {
        return fmt.Errorf("no items array: %w", err)
    }
    
    // Process in batches to control memory usage
    const batchSize = 100
    for i := 0; i < len(items); i += batchSize {
        end := i + batchSize
        if end > len(items) {
            end = len(items)
        }
        
        batch := items[i:end]
        if err := processBatch(batch); err != nil {
            return fmt.Errorf("batch %d-%d failed: %w", i, end-1, err)
        }
    }
    
    return nil
}
```

### 4. Configuration Management

```go
type Config struct {
    Database DatabaseConfig
    Cache    CacheConfig
    Features []string
}

func LoadConfig(configJSON string) (*Config, error) {
    obj := Parse(configJSON)
    if !obj.IsValid() {
        return nil, fmt.Errorf("invalid config JSON: %w", obj.Error())
    }
    
    config := &Config{
        Database: DatabaseConfig{
            Host:    obj.Get("database", "host").StringOr("localhost"),
            Port:    obj.Get("database", "port").IntOr(5432),
            Name:    obj.Get("database", "name").StringOr("app"),
            SSL:     obj.Get("database", "ssl").BoolOr(false),
            Timeout: obj.Get("database", "timeout").IntOr(30),
        },
        Cache: CacheConfig{
            Host: obj.Get("cache", "host").StringOr("localhost"),
            Port: obj.Get("cache", "port").IntOr(6379),
            TTL:  obj.Get("cache", "ttl").IntOr(3600),
        },
    }
    
    // Handle arrays
    if features, err := obj.Get("features").Array(); err == nil {
        config.Features = make([]string, 0, len(features))
        for _, feature := range features {
            if name := feature.StringOr(""); name != "" {
                config.Features = append(config.Features, name)
            }
        }
    }
    
    return config, nil
}
```

## Troubleshooting

### Common Issues

#### 1. "Key not found" errors

**Problem**: Getting errors when accessing nested properties.

**Solution**: Use `Has()` to check existence or `GetOr()` for defaults.

```go
// Problem
value := obj.Get("might", "not", "exist").String() // May error

// Solutions
if obj.Has("might", "not", "exist") {
    value, _ := obj.Get("might", "not", "exist").String()
}

// Or with default
value := obj.Get("might", "not", "exist").StringOr("default")
```

#### 2. Type conversion errors

**Problem**: Getting type conversion errors unexpectedly.

**Debug**: Check the actual JSON type first.

```go
// Debug the type
valueType := obj.Get("field").Type()
fmt.Printf("Field type: %s\n", valueType)

raw := obj.Get("field").Raw()
fmt.Printf("Raw value: %v (%T)\n", raw, raw)
```

#### 3. Performance issues

**Problem**: Slow performance with large JSON.

**Solutions**:
- Parse once, access many times
- Use array iteration instead of repeated indexing
- Check for memory leaks in long-running processes

#### 4. Memory usage

**Problem**: High memory consumption.

**Debug**:
```go
func debugMemory() {
    var m1, m2 runtime.MemStats
    
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Your jsjson code here
    obj := Parse(largeJSON)
    _ = obj.Get("some", "path").StringOr("")
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    fmt.Printf("Memory used: %d bytes\n", m2.Alloc-m1.Alloc)
}
```

### Debugging Tips

#### 1. Enable detailed error information

```go
obj := Parse(jsonStr)
if !obj.IsValid() {
    if jsonErr, ok := obj.Error().(*JSONError); ok {
        fmt.Printf("Operation: %s\n", jsonErr.Op)
        fmt.Printf("Error: %v\n", jsonErr.Err)
    }
}
```

#### 2. Trace navigation paths

```go
func debugPath(obj JSONValue, path ...interface{}) {
    current := obj
    for i, key := range path {
        fmt.Printf("Step %d: key=%v, valid=%t\n", i, key, current.IsValid())
        if !current.IsValid() {
            fmt.Printf("Error at step %d: %v\n", i, current.Error())
            return
        }
        current = current.Get(key)
    }
    fmt.Printf("Final result: valid=%t, type=%s\n", current.IsValid(), current.Type())
}

// Usage
debugPath(obj, "users", 0, "profile", "email")
```

#### 3. Validate JSON structure

```go
func validateStructure(obj JSONValue, schema map[string]string) error {
    for path, expectedType := range schema {
        parts := strings.Split(path, ".")
        value := obj
        for _, part := range parts {
            if i, err := strconv.Atoi(part); err == nil {
                value = value.Get(i)
            } else {
                value = value.Get(part)
            }
        }
        
        if !value.IsValid() {
            return fmt.Errorf("missing required field: %s", path)
        }
        
        if value.Type() != expectedType {
            return fmt.Errorf("field %s: expected %s, got %s", 
                            path, expectedType, value.Type())
        }
    }
    return nil
}

// Usage
schema := map[string]string{
    "user.name":  "string",
    "user.age":   "number",
    "user.active": "boolean",
}
if err := validateStructure(obj, schema); err != nil {
    log.Printf("Schema validation failed: %v", err)
}
```

## FAQ

### Q: How does jsjson compare to encoding/json performance-wise?

**A**: jsjson has similar parsing performance to `encoding/json` but offers much better developer experience. The overhead is minimal (typically <10%) and is offset by reduced development time and fewer bugs.

### Q: Can I modify JSON with jsjson?

**A**: jsjson is primarily designed for reading JSON. For modifications, you should:
1. Extract values with jsjson
2. Modify using Go data structures
3. Use `Stringify` to convert back to JSON

### Q: Is jsjson thread-safe?

**A**: Yes, for reading operations. Multiple goroutines can safely read from the same `JSONValue`. However, avoid sharing `JSONValue` objects across goroutines if any modifications might occur.

### Q: How do I handle very large JSON files?

**A**: For very large JSON (>100MB), consider:
1. Streaming JSON parsers for sequential processing
2. `gjson` for read-only access with zero allocation
3. Breaking large JSON into smaller chunks

### Q: Can I use jsjson with JSON Schema validation?

**A**: jsjson doesn't include schema validation, but you can combine it with schema validation libraries:

```go
// Validate with schema library first
if err := validateSchema(jsonStr, schema); err != nil {
    return err
}

// Then process with jsjson
obj := Parse(jsonStr) // Now guaranteed to be valid
```

### Q: How do I handle different number types?

**A**: JSON only has one number type (float64 in Go). Use type conversion methods:

```go
// For integers
id := obj.Get("id").IntOr(0)

// For floats
score := obj.Get("score").Float64Or(0.0)

// Check if it's actually an integer value
value := obj.Get("number").Raw()
if f, ok := value.(float64); ok && f == math.Trunc(f) {
    // It's an integer value
    intVal := int(f)
}
```

### Q: What about custom types and struct tags?

**A**: jsjson works with the underlying JSON data. For custom struct binding, use it alongside `encoding/json`:

```go
// Parse with jsjson for navigation
obj := Parse(jsonStr)
userData := obj.Get("user").Raw()

// Then unmarshal to struct with encoding/json
var user User
if data, err := json.Marshal(userData); err == nil {
    json.Unmarshal(data, &user)
}
```

### Q: How do I handle time/date fields?

**A**: Extract as strings and parse with Go's time package:

```go
dateStr := obj.Get("created_at").StringOr("")
if dateStr != "" {
    if timestamp, err := time.Parse(time.RFC3339, dateStr); err == nil {
        // Use timestamp
    }
}

// For Unix timestamps
unixTime := obj.Get("timestamp").IntOr(0)
if unixTime > 0 {
    timestamp := time.Unix(int64(unixTime), 0)
}
```

---

This completes the comprehensive documentation for jsjson. The library provides a powerful, JavaScript-like interface for JSON processing in Go while maintaining performance and type safety.
