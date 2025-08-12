# 🚀 jsjson - JavaScript-like JSON for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/KTBsomen/jsjson.svg)](https://pkg.go.dev/github.com/KTBsomen/jsjson)
[![Go Report Card](https://goreportcard.com/badge/github.com/KTBsomen/jsjson)](https://goreportcard.com/report/github.com/KTBsomen/jsjson)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A high-performance JSON library for Go that brings JavaScript-like JSON manipulation to Go with elegant error handling, type safety, and chainable operations.

## ✨ Why jsjson?

Coming from JavaScript? Tired of verbose Go JSON handling? **jsjson** bridges that gap:

```go
// ❌ Traditional Go way
var data map[string]interface{}
json.Unmarshal([]byte(jsonStr), &data)
users, ok := data["users"].([]interface{})
if !ok || len(users) == 0 {
    return "", errors.New("no users")
}
user, ok := users[0].(map[string]interface{})
if !ok {
    return "", errors.New("invalid user")
}
name, ok := user["name"].(string)
if !ok {
    name = "Unknown"
}

// ✅ jsjson way
name := Parse(jsonStr).Get("users", 0, "name").StringOr("Unknown")
```

## 🎯 Features

- **🔗 Chainable API** - JavaScript-like method chaining
- **🛡️ Built-in Error Handling** - No more manual type assertions
- **⚡ High Performance** - Competitive with popular Go JSON libraries
- **🎭 Type Safety** - Graceful type conversions with defaults
- **🔍 Deep Access** - Navigate nested structures effortlessly
- **🚫 Zero Dependencies** - Uses only Go standard library
- **♻️ Memory Efficient** - Object pooling for reduced GC pressure
- **🧵 Concurrent Safe** - Safe for concurrent read operations

## 📦 Installation

```bash
go get github.com/KTBsomen/jsjson
```

## 🚀 Quick Start

```go
package main

import (
    "fmt"
    "github.com/KTBsomen/jsjson"
)

func main() {
    jsonStr := `{
        "user": {
            "name": "John Doe",
            "age": 30,
            "hobbies": ["reading", "coding", "gaming"],
            "address": {
                "city": "New York",
                "zipcode": "10001"
            }
        }
    }`

    // Parse JSON
    obj := jsonjs.Parse(jsonStr)
    
    // Access nested values with defaults
    name := obj.Get("user", "name").StringOr("Unknown")
    age := obj.Get("user", "age").IntOr(0)
    city := obj.Get("user", "address", "city").StringOr("N/A")
    firstHobby := obj.Get("user", "hobbies", 0).StringOr("None")
    
    fmt.Printf("Name: %s, Age: %d, City: %s, Hobby: %s\n", 
               name, age, city, firstHobby)
    // Output: Name: John Doe, Age: 30, City: New York, Hobby: reading
}
```

## 📚 Complete API Reference

### Core Functions

#### Parse
```go
func Parse(v interface{}) JSONValue
```
Parses JSON from string, []byte, or any Go value. Returns JSONValue with error handling.

```go
// From JSON string
obj := Parse(`{"name": "John", "age": 30}`)

// From byte slice
data := []byte(`{"active": true}`)
obj := Parse(data)

// From Go struct/map
user := map[string]interface{}{"id": 123}
obj := Parse(user)
```

#### MustParse
```go
func MustParse(v interface{}) JSONValue
```
Like Parse but panics on error. Use when you're certain the JSON is valid.

```go
obj := MustParse(`{"valid": "json"}`)
```

#### Stringify
```go
func Stringify(v interface{}) (string, error)
```
Converts any value back to JSON string.

```go
data := map[string]interface{}{"hello": "world"}
jsonStr, err := Stringify(data)
// Output: {"hello":"world"}
```

#### StringifyPretty
```go
func StringifyPretty(v interface{}, indent string) (string, error)
```
Pretty-prints JSON with custom indentation.

```go
jsonStr, err := StringifyPretty(data, "  ")
// Output:
// {
//   "hello": "world"
// }
```

### JSONValue Methods

#### Error Checking
```go
func (j JSONValue) IsValid() bool
func (j JSONValue) Error() error
```

```go
obj := Parse(`invalid json`)
if !obj.IsValid() {
    fmt.Println("Error:", obj.Error())
}
```

#### Navigation
```go
func (j JSONValue) Get(keys ...interface{}) JSONValue
func (j JSONValue) GetOr(defaultValue interface{}, keys ...interface{}) interface{}
func (j JSONValue) Has(keys ...interface{}) bool
```

```go
// Get with error propagation
name := obj.Get("user", "name")
if name.IsValid() {
    fmt.Println(name.StringOr(""))
}

// Get with default
email := obj.GetOr("no-email@example.com", "user", "email")

// Check existence
if obj.Has("user", "premium") {
    // Handle premium user
}
```

#### Type Conversions

All conversion methods come in two flavors:
- `Type()` - Returns `(value, error)`
- `TypeOr(default)` - Returns `value` or `default` on error

```go
// String conversions
name, err := obj.Get("name").String()
name := obj.Get("name").StringOr("Unknown")

// Numeric conversions
age, err := obj.Get("age").Int()
age := obj.Get("age").IntOr(0)

score, err := obj.Get("score").Float64()
score := obj.Get("score").Float64Or(0.0)

// Boolean conversions
active, err := obj.Get("active").Bool()
active := obj.Get("active").BoolOr(false)
```

#### Collection Operations
```go
func (j JSONValue) Array() ([]JSONValue, error)
func (j JSONValue) Object() (map[string]JSONValue, error)
```

```go
// Iterate over array
tags, err := obj.Get("tags").Array()
if err == nil {
    for i, tag := range tags {
        fmt.Printf("Tag %d: %s\n", i, tag.StringOr(""))
    }
}

// Iterate over object
metadata, err := obj.Get("metadata").Object()
if err == nil {
    for key, value := range metadata {
        fmt.Printf("%s: %v\n", key, value.Raw())
    }
}
```

#### Utility Methods
```go
func (j JSONValue) Raw() interface{}
func (j JSONValue) IsNull() bool
func (j JSONValue) Type() string
func (j JSONValue) Clone() JSONValue
```

```go
// Get underlying Go value
raw := obj.Get("data").Raw()

// Check for null
if obj.Get("optional").IsNull() {
    // Handle null case
}

// Get JSON type
switch obj.Get("value").Type() {
case "string":
    // Handle string
case "number":
    // Handle number
case "array":
    // Handle array
}

// Deep clone
copy := obj.Clone()
```

## 🔥 Real-World Examples

### API Response Processing
```go
func processUserAPI(jsonResponse string) {
    api := Parse(jsonResponse)
    
    // Extract pagination info
    currentPage := api.Get("pagination", "page").IntOr(1)
    totalPages := api.Get("pagination", "total_pages").IntOr(1)
    
    // Process users array
    users, err := api.Get("data", "users").Array()
    if err != nil {
        log.Printf("No users found: %v", err)
        return
    }
    
    for i, user := range users {
        id := user.Get("id").IntOr(0)
        name := user.Get("profile", "name").StringOr("Anonymous")
        email := user.Get("contact", "email").StringOr("N/A")
        isActive := user.Get("status", "active").BoolOr(false)
        
        fmt.Printf("User %d: %s <%s> Active: %t\n", 
                   id, name, email, isActive)
    }
    
    fmt.Printf("Page %d of %d\n", currentPage, totalPages)
}
```

### Configuration File Parsing
```go
func loadConfig(configJSON string) AppConfig {
    config := Parse(configJSON)
    
    return AppConfig{
        Database: DatabaseConfig{
            Host:     config.Get("database", "host").StringOr("localhost"),
            Port:     config.Get("database", "port").IntOr(5432),
            Name:     config.Get("database", "name").StringOr("app"),
            SSL:      config.Get("database", "ssl").BoolOr(false),
            Timeout:  config.Get("database", "timeout").IntOr(30),
        },
        Redis: RedisConfig{
            Host: config.Get("cache", "redis", "host").StringOr("localhost"),
            Port: config.Get("cache", "redis", "port").IntOr(6379),
            DB:   config.Get("cache", "redis", "db").IntOr(0),
        },
        Features: extractFeatures(config.Get("features")),
    }
}

func extractFeatures(featuresValue JSONValue) []string {
    features, err := featuresValue.Array()
    if err != nil {
        return []string{}
    }
    
    result := make([]string, 0, len(features))
    for _, feature := range features {
        if name := feature.StringOr(""); name != "" {
            result = append(result, name)
        }
    }
    return result
}
```

### Error Handling Patterns
```go
func safeJSONProcessing(jsonStr string) error {
    obj := Parse(jsonStr)
    
    // Check if parsing succeeded
    if !obj.IsValid() {
        return fmt.Errorf("failed to parse JSON: %w", obj.Error())
    }
    
    // Safe navigation with error checking
    userObj := obj.Get("user")
    if !userObj.IsValid() {
        return fmt.Errorf("user object not found: %w", userObj.Error())
    }
    
    // Type conversion with error handling
    age, err := userObj.Get("age").Int()
    if err != nil {
        log.Printf("Warning: invalid age, using default: %v", err)
        age = 0
    }
    
    // Safe array access
    hobbies, err := userObj.Get("hobbies").Array()
    if err != nil {
        log.Printf("No hobbies array found: %v", err)
        hobbies = []JSONValue{}
    }
    
    processUser(age, hobbies)
    return nil
}
```

### Dynamic JSON Generation
```go
func createUserResponse(users []User) (string, error) {
    response := map[string]interface{}{
        "status": "success",
        "data": map[string]interface{}{
            "users": users,
            "count": len(users),
        },
        "pagination": map[string]interface{}{
            "page":       1,
            "per_page":   len(users),
            "total":      len(users),
            "total_pages": 1,
        },
        "meta": map[string]interface{}{
            "timestamp":   time.Now().Unix(),
            "api_version": "v1.0.0",
        },
    }
    
    return Stringify(response)
}
```

## 📊 Performance Benchmarks

Benchmarks comparing **jsjson** against popular Go JSON libraries on Intel i3-1220P (12th Gen):

### Parse Performance

| Library | Small JSON (42B) | Medium JSON (312B) | Large JSON (1.6KB) |
|---------|-------------------|--------------------|--------------------|
| **jsjson** | **855 ns/op** | **4.0 μs/op** | **18.8 μs/op** |
| Standard Library | 834 ns/op | 3.5 μs/op | 18.8 μs/op |
| go-json | **726 ns/op** | **3.5 μs/op** | **15.7 μs/op** |
| json-iterator | **728 ns/op** | **2.8 μs/op** | **16.1 μs/op** |
| gjson | **12.9 ns/op** ⚡ | **13.5 ns/op** ⚡ | **13.7 ns/op** ⚡ |

*Note: gjson's extremely fast parsing is because it doesn't fully parse JSON upfront*

### Memory Allocations (Parse)

| Library | Small JSON | Medium JSON | Large JSON |
|---------|------------|-------------|------------|
| **jsjson** | 616 B/14 allocs | 1,576 B/46 allocs | 9,412 B/192 allocs |
| Standard Library | 600 B/13 allocs | 1,560 B/45 allocs | 9,396 B/191 allocs |
| go-json | **520 B/12 allocs** | 1,585 B/45 allocs | 10,089 B/217 allocs |
| json-iterator | **496 B/13 allocs** | **1,608 B/53 allocs** | 10,202 B/244 allocs |
| gjson | **0 B/0 allocs** ⚡ | **0 B/0 allocs** ⚡ | **0 B/0 allocs** ⚡ |

### Value Access Performance

| Operation | jsjson | Standard Library | gjson |
|-----------|--------|------------------|-------|
| Simple Get | **11.2 ns/op** | **4.4 ns/op** ⚡ | 46.3 ns/op |
| Nested Get | **48.6 ns/op** | **39.9 ns/op** ⚡ | 362 ns/op |

### Marshal/Stringify Performance

| Library | Time | Memory |
|---------|------|--------|
| **jsjson** | 3.5 μs/op | 1,216 B/26 allocs |
| Standard Library | 3.5 μs/op | 1,008 B/25 allocs |
| go-json | **1.9 μs/op** ⚡ | **592 B/2 allocs** ⚡ |
| json-iterator | **1.4 μs/op** ⚡ | **464 B/7 allocs** ⚡ |

### Real-World Scenario (Parse + Multiple Gets + Type Conversions)

| Library | Time | Memory | Zero Allocs |
|---------|------|--------|-------------|
| **jsjson** | **3.9 μs/op** | 1,576 B/46 allocs | ❌ |
| Standard Library | 5.0 μs/op | 1,696 B/54 allocs | ❌ |
| gjson | **895 ns/op** ⚡ | **0 B/0 allocs** ⚡ | ✅ |

### Key Takeaways

- **jsjson** offers competitive performance with JavaScript-like API convenience
- **gjson** excels in read-only scenarios with zero allocations
- **json-iterator** and **go-json** are fastest for marshal operations
- **jsjson** provides the best balance of performance and ease-of-use for JavaScript developers

### Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. -benchmem

# Run library comparison only
go test -bench=BenchmarkLibraryComparison -benchmem

# Run with multiple iterations for more accurate results
go test -bench=BenchmarkLibraryComparison -benchmem -count=5
```

## 🏆 Comparison with Other Libraries

| Feature | jsjson | Standard Library | gjson | go-json | json-iterator |
|---------|--------|------------------|-------|---------|---------------|
| **Ease of Use** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐ |
| **Performance** | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Memory Usage** | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐ |
| **Error Handling** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐⭐ | ⭐⭐ |
| **Type Safety** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐⭐ | ⭐⭐ |
| **API Consistency** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ |
| **Modification** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ❌ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

### When to Use jsjson

✅ **Perfect for:**
- JavaScript/Node.js developers transitioning to Go
- Rapid prototyping and development
- Complex nested JSON processing
- Applications requiring robust error handling
- Code that benefits from chainable operations

🤔 **Consider alternatives for:**
- Maximum performance critical applications (use json-iterator/go-json)
- Read-only JSON processing (use gjson)
- Simple struct marshaling/unmarshaling (use standard library)

## 🛠️ Advanced Usage

### Custom Error Handling
```go
func processWithCustomErrors(jsonStr string) {
    obj := Parse(jsonStr)
    
    // Custom error wrapper
    if !obj.IsValid() {
        if jsonErr, ok := obj.Error().(*JSONError); ok {
            log.Printf("JSON operation '%s' failed: %v", jsonErr.Op, jsonErr.Err)
        }
        return
    }
    
    // Process...
}
```

### Working with Large JSON
```go
func processLargeJSON(jsonStr string) {
    obj := Parse(jsonStr)
    
    // Use array iteration for memory efficiency
    items, err := obj.Get("items").Array()
    if err != nil {
        return
    }
    
    // Process in chunks
    const chunkSize = 100
    for i := 0; i < len(items); i += chunkSize {
        end := i + chunkSize
        if end > len(items) {
            end = len(items)
        }
        
        chunk := items[i:end]
        processChunk(chunk)
    }
}
```

### Type Validation Patterns
```go
func validateUserData(userData JSONValue) error {
    // Required fields
    if !userData.Has("email") {
        return errors.New("email is required")
    }
    
    if !userData.Has("name") {
        return errors.New("name is required")
    }
    
    // Type validation
    if userData.Get("age").Type() != "number" {
        return errors.New("age must be a number")
    }
    
    // Range validation
    age := userData.Get("age").IntOr(0)
    if age < 0 || age > 150 {
        return errors.New("age must be between 0 and 150")
    }
    
    return nil
}
```

## 🧪 Testing with jsjson

```go
func TestUserProcessing(t *testing.T) {
    testJSON := `{
        "user": {
            "id": 123,
            "name": "John Doe",
            "email": "john@example.com",
            "preferences": {
                "theme": "dark",
                "notifications": true
            }
        }
    }`
    
    obj := Parse(testJSON)
    assert.True(t, obj.IsValid(), "JSON should be valid")
    
    // Test individual fields
    assert.Equal(t, 123, obj.Get("user", "id").IntOr(0))
    assert.Equal(t, "John Doe", obj.Get("user", "name").StringOr(""))
    assert.Equal(t, "dark", obj.Get("user", "preferences", "theme").StringOr(""))
    
    // Test error cases
    invalidField := obj.Get("user", "nonexistent")
    assert.False(t, invalidField.IsValid())
    assert.Equal(t, "default", invalidField.StringOr("default"))
}
```

## 🔧 Configuration and Optimization

### Memory Optimization
The library uses object pooling internally to reduce GC pressure. For additional optimization:

```go
// For high-frequency parsing, consider reusing JSONValue objects
obj := Parse(jsonStr)
defer func() {
    // Object will be returned to pool automatically
    // when it goes out of scope
}()
```

### Concurrent Usage
```go
func concurrentProcessing(jsonStrs []string) {
    var wg sync.WaitGroup
    results := make(chan string, len(jsonStrs))
    
    for _, jsonStr := range jsonStrs {
        wg.Add(1)
        go func(js string) {
            defer wg.Done()
            
            obj := Parse(js) // Safe for concurrent use
            name := obj.Get("name").StringOr("Unknown")
            results <- name
        }(jsonStr)
    }
    
    go func() {
        wg.Wait()
        close(results)
    }()
    
    for name := range results {
        fmt.Println("Processed:", name)
    }
}
```

## 🤝 Contributing

We welcome contributions! Here's how you can help:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** with tests
4. **Run the test suite**: `go test ./...`
5. **Run benchmarks**: `go test -bench=. -benchmem`
6. **Commit your changes**: `git commit -m 'Add amazing feature'`
7. **Push to the branch**: `git push origin feature/amazing-feature`
8. **Open a Pull Request**

### Development Setup

```bash
# Clone the repository
git clone https://github.com/KTBsomen/jsjson.git
cd jsjson

# Run tests
go test -v ./...

# Run benchmarks
go test -bench=. -benchmem

# Run with race detection
go test -race ./...
```

### Code Style

- Follow standard Go conventions
- Add tests for new features
- Update benchmarks for performance-related changes
- Keep API consistent with JavaScript JSON patterns

## 📜 License

This project is licensed under the GPL-3.0 License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by JavaScript's JSON API design
- Thanks to the Go community for excellent JSON libraries that set the performance bar
- Special thanks to contributors and users providing feedback

## 📞 Support

- 📫 **Issues**: [GitHub Issues](https://github.com/KTBsomen/jsjson/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/KTBsomen/jsjson/discussions)
- 📖 **Documentation**: [pkg.go.dev](https://pkg.go.dev/github.com/KTBsomen/jsjson)

---

⭐ **If you find jsjson helpful, please star the repository!** ⭐

Made with ❤️ for the Go community
