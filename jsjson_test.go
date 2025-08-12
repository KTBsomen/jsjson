package jsjson_test

import (
	"fmt"
	"testing"
	"time"

	JSON "github.com/ktbsomen/jsjson"
)

// Example usage demonstrating the improved API
func ExampleUsage() {
	// Parse JSON with error handling
	jsonStr := `{
		"name": "John",
		"age": 30,
		"active": true,
		"scores": [95, 87, 92],
		"address": {
			"street": "123 Main St",
			"city": "New York"
		},
		"metadata": null
	}`

	// Parse with error checking
	obj := JSON.Parse(jsonStr)
	if !obj.IsValid() {
		fmt.Printf("Parse error: %v\n", obj.Error())
		return
	}

	// Safe access with error handling
	name, err := obj.Get("name").String()
	if err != nil {
		fmt.Printf("Error getting name: %v\n", err)
		return
	}
	fmt.Printf("Name: %s\n", name)

	// Or use the "Or" methods for defaults
	age := obj.Get("age").IntOr(0)
	fmt.Printf("Age: %d\n", age)

	// Nested access
	city := obj.Get("address", "city").StringOr("Unknown")
	fmt.Printf("City: %s\n", city)

	// Array access
	firstScore := obj.Get("scores", 0).IntOr(0)
	fmt.Printf("First score: %d\n", firstScore)

	// Check if keys exist
	if obj.Has("metadata") {
		fmt.Println("Has metadata key")
	}

	// Iterate over arrays
	if scores, err := obj.Get("scores").Array(); err == nil {
		fmt.Print("Scores: ")
		for _, score := range scores {
			fmt.Printf("%d ", score.IntOr(0))
		}
		fmt.Println()
	}

	// Stringify with error handling
	if jsonStr, err := JSON.Stringify(obj.Raw()); err == nil {
		fmt.Printf("JSON: %s\n", jsonStr)
	}
}

// Benchmark tests
func BenchmarkParse(b *testing.B) {
	jsonStr := `{"name":"John","age":30,"scores":[95,87,92],"active":true}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		JSON.Parse(jsonStr)
	}
}

func BenchmarkGet(b *testing.B) {
	obj := JSON.Parse(`{"user":{"profile":{"name":"John","scores":[95,87,92]}}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj.Get("user", "profile", "scores", 1)
	}
}

func BenchmarkStringify(b *testing.B) {
	data := map[string]interface{}{
		"name":   "John",
		"age":    30,
		"scores": []int{95, 87, 92},
		"active": true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		JSON.Stringify(data)
	}
}

// Test functions
func TestParseEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{"nil input", nil, true},
		{"empty string", "", true},
		{"empty bytes", []byte{}, true},
		{"invalid json", "{invalid}", true},
		{"valid json", `{"key":"value"}`, false},
		{"valid struct", struct{ Name string }{"John"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := JSON.Parse(tt.input)
			if tt.wantErr && result.IsValid() {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && !result.IsValid() {
				t.Errorf("Expected no error but got: %v", result.Error())
			}
		})
	}
}

func TestGetEdgeCases(t *testing.T) {
	obj := JSON.Parse(`{"array":[1,2,3],"object":{"nested":"value"},"null":null}`)

	tests := []struct {
		name    string
		keys    []interface{}
		wantErr bool
	}{
		{"valid object key", []interface{}{"object", "nested"}, false},
		{"valid array index", []interface{}{"array", 1}, false},
		{"invalid object key", []interface{}{"object", "missing"}, true},
		{"array index out of bounds", []interface{}{"array", 10}, true},
		{"string key on array", []interface{}{"array", "invalid"}, true},
		{"access on null", []interface{}{"null", "key"}, true},
		{"negative array index", []interface{}{"array", -1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := obj.Get(tt.keys...)
			if tt.wantErr && result.IsValid() {
				t.Error("Expected error but got none")
			}
			if !tt.wantErr && !result.IsValid() {
				t.Errorf("Expected no error but got: %v", result.Error())
			}
		})
	}
}

func TestTypeConversions(t *testing.T) {
	obj := JSON.Parse(`{
		"string": "hello",
		"number": "42.5",
		"integer": 42,
		"null": null,
		"boolean":true,
		"stringNumber": "123",
		"stringBool": "true"
	}`)

	// Test string conversion
	if s, err := obj.Get("string").String(); err != nil || s != "hello" {
		t.Errorf("String conversion failed: got %v, err: %v", s, err)
	}

	// Test number conversions
	if f, err := obj.Get("number").Float64(); err != nil || f != 42.5 {
		t.Errorf("Float64 conversion failed: got %v, err: %v", f, err)
	}

	if i, err := obj.Get("integer").Int(); err != nil || i != 42 {
		t.Errorf("Int conversion failed: got %v, err: %v", i, err)
	}

	// Test boolean conversion
	if b, err := obj.Get("boolean").Bool(); err != nil || !b {
		t.Errorf("Bool conversion failed: got %v, err: %v", b, err)
	}

	// Test string to number conversion
	if i, err := obj.Get("stringNumber").Int(); err != nil || i != 123 {
		t.Errorf("String to int conversion failed: got %v, err: %v", i, err)
	}

	// Test string to bool conversion
	if b, err := obj.Get("stringBool").Bool(); err != nil || !b {
		t.Errorf("String to bool conversion failed: got %v, err: %v", b, err)
	}

	// Test null handling
	if s := obj.Get("null").StringOr("default"); s != "default" {
		t.Errorf("Null handling failed: got %v", s)
	}
}

func TestMemoryUsage(t *testing.T) {
	// Test that we don't leak memory with many operations
	jsonStr := `{"data": [1,2,3,4,5]}`

	for i := 0; i < 1000; i++ {
		obj := JSON.Parse(jsonStr)
		obj.Get("data", 0).Int()
		obj.Get("data", 1).Int()
		obj.Get("data", 2).Int()
	}
	// No memory assertions here, but this tests for obvious leaks
}

func TestConcurrency(t *testing.T) {
	jsonStr := `{"counter": 0, "data": [1,2,3,4,5]}`
	obj := JSON.Parse(jsonStr)

	// Test concurrent access (should not panic)
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 100; j++ {
				obj.Get("data", j%5).Int()
				obj.Get("counter").Int()
			}
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// Performance comparison example
func ExamplePerformanceComparison() {
	jsonStr := `{
		"users": [
			{"name": "John", "age": 30, "scores": [95, 87, 92]},
			{"name": "Jane", "age": 25, "scores": [88, 92, 94]},
			{"name": "Bob", "age": 35, "scores": [76, 89, 91]}
		]
	}`

	start := time.Now()

	// Parse once, use many times (efficient)
	obj := JSON.Parse(jsonStr)
	if !obj.IsValid() {
		fmt.Printf("Parse error: %v\n", obj.Error())
		return
	}

	// Access nested data efficiently
	users, err := obj.Get("users").Array()
	if err != nil {
		fmt.Printf("Array access error: %v\n", err)
		return
	}

	totalScore := 0
	userCount := 0

	for _, user := range users {
		name := user.Get("name").StringOr("Unknown")
		age := user.Get("age").IntOr(0)

		scores, err := user.Get("scores").Array()
		if err != nil {
			continue
		}

		userScore := 0
		for _, score := range scores {
			userScore += score.IntOr(0)
		}

		avgScore := userScore / len(scores)
		totalScore += avgScore
		userCount++

		fmt.Printf("User: %s, Age: %d, Avg Score: %d\n", name, age, avgScore)
	}

	if userCount > 0 {
		fmt.Printf("Overall average: %d\n", totalScore/userCount)
	}

	elapsed := time.Since(start)
	fmt.Printf("Processing time: %v\n", elapsed)
}
