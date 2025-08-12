package jsjson

import (
	"encoding/json"
	"testing"

	gojson "github.com/goccy/go-json"
	"github.com/tidwall/gjson"
	jsoniter "github.com/json-iterator/go"
)

// Test data for benchmarks
var (
	smallJSON = `{"name":"John","age":30,"active":true}`
	
	mediumJSON = `{
		"id": 12345,
		"name": "John Doe",
		"email": "john@example.com",
		"age": 30,
		"active": true,
		"score": 95.5,
		"tags": ["developer", "golang", "json"],
		"metadata": {
			"created": "2023-01-01",
			"updated": "2023-12-01",
			"version": 2
		}
	}`
	
	largeJSON = `{
		"users": [
			{
				"id": 1,
				"name": "John Doe",
				"email": "john@example.com",
				"age": 30,
				"active": true,
				"score": 95.5,
				"preferences": {
					"theme": "dark",
					"language": "en",
					"notifications": true,
					"privacy": {
						"public": false,
						"shareable": true
					}
				},
				"scores": [95, 87, 92, 88, 94],
				"tags": ["developer", "golang", "json", "api"],
				"metadata": {
					"created": "2023-01-01",
					"updated": "2023-12-01",
					"version": 2,
					"flags": {
						"premium": true,
						"verified": false
					}
				}
			},
			{
				"id": 2,
				"name": "Jane Smith",
				"email": "jane@example.com",
				"age": 25,
				"active": false,
				"score": 88.2,
				"preferences": {
					"theme": "light",
					"language": "es",
					"notifications": false,
					"privacy": {
						"public": true,
						"shareable": false
					}
				},
				"scores": [88, 92, 84, 90, 87],
				"tags": ["designer", "ui", "ux"],
				"metadata": {
					"created": "2023-02-15",
					"updated": "2023-11-20",
					"version": 1,
					"flags": {
						"premium": false,
						"verified": true
					}
				}
			}
		],
		"pagination": {
			"page": 1,
			"limit": 10,
			"total": 2,
			"hasNext": false,
			"hasPrev": false
		},
		"meta": {
			"timestamp": 1702934400,
			"requestId": "req_123456789",
			"apiVersion": "v1.2.3"
		}
	}`

	// Pre-parsed objects for Get benchmarks
	smallObj  = Parse(smallJSON)
	mediumObj = Parse(mediumJSON)
	largeObj  = Parse(largeJSON)

	// Standard library comparison data
	smallStd  map[string]interface{}
	mediumStd map[string]interface{}
	largeStd  map[string]interface{}
)

func init() {
	// Pre-parse for standard library comparison
	json.Unmarshal([]byte(smallJSON), &smallStd)
	json.Unmarshal([]byte(mediumJSON), &mediumStd)
	json.Unmarshal([]byte(largeJSON), &largeStd)
}

// ==================== PARSE BENCHMARKS ====================

func BenchmarkParse_Small(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Parse(smallJSON)
	}
}

func BenchmarkParse_Medium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Parse(mediumJSON)
	}
}

func BenchmarkParse_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Parse(largeJSON)
	}
}

// Standard library comparison
func BenchmarkParseStdLib_Small(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var result interface{}
		json.Unmarshal([]byte(smallJSON), &result)
	}
}

func BenchmarkParseStdLib_Medium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var result interface{}
		json.Unmarshal([]byte(mediumJSON), &result)
	}
}

func BenchmarkParseStdLib_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var result interface{}
		json.Unmarshal([]byte(largeJSON), &result)
	}
}

// ==================== GET BENCHMARKS ====================

func BenchmarkGet_SimpleKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		smallObj.Get("name")
	}
}

func BenchmarkGet_NestedKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mediumObj.Get("metadata", "version")
	}
}

func BenchmarkGet_ArrayIndex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		mediumObj.Get("tags", 0)
	}
}

func BenchmarkGet_DeepNested(b *testing.B) {
	for i := 0; i < b.N; i++ {
		largeObj.Get("users", 0, "preferences", "privacy", "public")
	}
}

func BenchmarkGet_ArrayOfObjects(b *testing.B) {
	for i := 0; i < b.N; i++ {
		largeObj.Get("users", 1, "scores", 2)
	}
}

// ==================== TYPE CONVERSION BENCHMARKS ====================

func BenchmarkString_Conversion(b *testing.B) {
	nameValue := smallObj.Get("name")
	for i := 0; i < b.N; i++ {
		nameValue.String()
	}
}

func BenchmarkInt_Conversion(b *testing.B) {
	ageValue := smallObj.Get("age")
	for i := 0; i < b.N; i++ {
		ageValue.Int()
	}
}

func BenchmarkFloat_Conversion(b *testing.B) {
	scoreValue := mediumObj.Get("score")
	for i := 0; i < b.N; i++ {
		scoreValue.Float64()
	}
}

func BenchmarkBool_Conversion(b *testing.B) {
	activeValue := smallObj.Get("active")
	for i := 0; i < b.N; i++ {
		activeValue.Bool()
	}
}

// With default values (Or methods)
func BenchmarkStringOr_Conversion(b *testing.B) {
	nameValue := smallObj.Get("name")
	for i := 0; i < b.N; i++ {
		nameValue.StringOr("default")
	}
}

func BenchmarkIntOr_Conversion(b *testing.B) {
	ageValue := smallObj.Get("age")
	for i := 0; i < b.N; i++ {
		ageValue.IntOr(0)
	}
}

// ==================== STRINGIFY BENCHMARKS ====================

func BenchmarkStringify_Small(b *testing.B) {
	data := smallObj.Raw()
	for i := 0; i < b.N; i++ {
		Stringify(data)
	}
}

func BenchmarkStringify_Medium(b *testing.B) {
	data := mediumObj.Raw()
	for i := 0; i < b.N; i++ {
		Stringify(data)
	}
}

func BenchmarkStringify_Large(b *testing.B) {
	data := largeObj.Raw()
	for i := 0; i < b.N; i++ {
		Stringify(data)
	}
}

// Standard library comparison
func BenchmarkStringifyStdLib_Small(b *testing.B) {
	for i := 0; i < b.N; i++ {
		json.Marshal(smallStd)
	}
}

func BenchmarkStringifyStdLib_Medium(b *testing.B) {
	for i := 0; i < b.N; i++ {
		json.Marshal(mediumStd)
	}
}

func BenchmarkStringifyStdLib_Large(b *testing.B) {
	for i := 0; i < b.N; i++ {
		json.Marshal(largeStd)
	}
}

// ==================== ITERATION BENCHMARKS ====================

func BenchmarkArray_Iteration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if arr, err := mediumObj.Get("tags").Array(); err == nil {
			for _, item := range arr {
				_,_ = item.String()
			}
		}
	}
}

func BenchmarkObject_Iteration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if obj, err := mediumObj.Get("metadata").Object(); err == nil {
			for _, value := range obj {
				_ = value.Raw()
			}
		}
	}
}

// ==================== ERROR HANDLING BENCHMARKS ====================

func BenchmarkGet_InvalidKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := smallObj.Get("nonexistent")
		_ = result.IsValid() // Check if error occurred
	}
}

func BenchmarkGet_InvalidIndex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := mediumObj.Get("tags", 999)
		_ = result.IsValid()
	}
}

func BenchmarkGet_TypeError(b *testing.B) {
	for i := 0; i < b.N; i++ {
		result := smallObj.Get("name", "invalid") // String doesn't have keys
		_ = result.IsValid()
	}
}

// ==================== MEMORY ALLOCATION BENCHMARKS ====================

func BenchmarkMemoryAllocation_Parse(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Parse(mediumJSON)
	}
}

func BenchmarkMemoryAllocation_Get(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mediumObj.Get("metadata", "version")
	}
}

func BenchmarkMemoryAllocation_Stringify(b *testing.B) {
	b.ReportAllocs()
	data := mediumObj.Raw()
	for i := 0; i < b.N; i++ {
		Stringify(data)
	}
}

// ==================== REALISTIC USAGE BENCHMARKS ====================

func BenchmarkRealisticUsage_UserProcessing(b *testing.B) {
	// Simulates real-world usage: parse JSON, extract multiple fields, do calculations
	for i := 0; i < b.N; i++ {
		obj := Parse(largeJSON)
		if !obj.IsValid() {
			continue
		}
		
		users, err := obj.Get("users").Array()
		if err != nil {
			continue
		}
		
		totalScore := 0.0
		activeUsers := 0
		
		for _, user := range users {
			if user.Get("active").BoolOr(false) {
				activeUsers++
				totalScore += user.Get("score").Float64Or(0)
			}
		}
		
		if activeUsers > 0 {
			_ = totalScore / float64(activeUsers) // Average score
		}
	}
}

func BenchmarkRealisticUsage_ConfigProcessing(b *testing.B) {
	// Simulates configuration file processing
	configJSON := `{
		"database": {
			"host": "localhost",
			"port": 5432,
			"name": "myapp",
			"ssl": true,
			"timeout": 30
		},
		"cache": {
			"redis": {
				"host": "localhost",
				"port": 6379,
				"db": 0
			}
		},
		"features": ["feature1", "feature2", "feature3"]
	}`
	
	for i := 0; i < b.N; i++ {
		config := Parse(configJSON)
		if !config.IsValid() {
			continue
		}
		
		// Extract configuration values
		dbHost := config.Get("database", "host").StringOr("localhost")
		dbPort := config.Get("database", "port").IntOr(5432)
		dbSSL := config.Get("database", "ssl").BoolOr(false)
		timeout := config.Get("database", "timeout").IntOr(10)
		
		redisHost := config.Get("cache", "redis", "host").StringOr("localhost")
		redisPort := config.Get("cache", "redis", "port").IntOr(6379)
		
		// Process features array
		if features, err := config.Get("features").Array(); err == nil {
			featureCount := len(features)
			_ = featureCount
		}
		
		// Use the extracted values (prevent optimization)
		_ = dbHost + redisHost
		_ = dbPort + redisPort + timeout
		_ = dbSSL
	}
}

// ==================== CONCURRENT ACCESS BENCHMARKS ====================

func BenchmarkConcurrentRead(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			largeObj.Get("users", 0, "name").StringOr("default")
		}
	})
}

func BenchmarkConcurrentParse(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Parse(mediumJSON)
		}
	})
}

// ==================== LIBRARY COMPARISON BENCHMARKS ====================

// Parse Performance Comparison
func BenchmarkLibraryComparison_Parse_Small_jsJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Parse(smallJSON)
	}
}

func BenchmarkLibraryComparison_Parse_Small_StdLib(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result interface{}
		json.Unmarshal([]byte(smallJSON), &result)
	}
}

func BenchmarkLibraryComparison_Parse_Small_GoJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result interface{}
		gojson.Unmarshal([]byte(smallJSON), &result)
	}
}

func BenchmarkLibraryComparison_Parse_Small_JsonIter(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result interface{}
		jsoniter.Unmarshal([]byte(smallJSON), &result)
	}
}

func BenchmarkLibraryComparison_Parse_Small_Gjson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		gjson.Parse(smallJSON)
	}
}

// Medium JSON Parse Comparison
func BenchmarkLibraryComparison_Parse_Medium_jsJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Parse(mediumJSON)
	}
}

func BenchmarkLibraryComparison_Parse_Medium_StdLib(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result interface{}
		json.Unmarshal([]byte(mediumJSON), &result)
	}
}

func BenchmarkLibraryComparison_Parse_Medium_GoJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result interface{}
		gojson.Unmarshal([]byte(mediumJSON), &result)
	}
}

func BenchmarkLibraryComparison_Parse_Medium_JsonIter(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result interface{}
		jsoniter.Unmarshal([]byte(mediumJSON), &result)
	}
}

func BenchmarkLibraryComparison_Parse_Medium_Gjson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		gjson.Parse(mediumJSON)
	}
}

// Large JSON Parse Comparison
func BenchmarkLibraryComparison_Parse_Large_jsJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Parse(largeJSON)
	}
}

func BenchmarkLibraryComparison_Parse_Large_StdLib(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result interface{}
		json.Unmarshal([]byte(largeJSON), &result)
	}
}

func BenchmarkLibraryComparison_Parse_Large_GoJson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result interface{}
		gojson.Unmarshal([]byte(largeJSON), &result)
	}
}

func BenchmarkLibraryComparison_Parse_Large_JsonIter(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result interface{}
		jsoniter.Unmarshal([]byte(largeJSON), &result)
	}
}

func BenchmarkLibraryComparison_Parse_Large_Gjson(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		gjson.Parse(largeJSON)
	}
}

// Value Access Comparison
func BenchmarkLibraryComparison_Get_Simple_jsJson(b *testing.B) {
	obj := Parse(smallJSON)
	for i := 0; i < b.N; i++ {
		obj.Get("name").StringOr("")
	}
}

func BenchmarkLibraryComparison_Get_Simple_StdLib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if val, ok := smallStd["name"].(string); ok {
			_ = val
		}
	}
}

func BenchmarkLibraryComparison_Get_Simple_Gjson(b *testing.B) {
	for i := 0; i < b.N; i++ {
		gjson.Get(smallJSON, "name").String()
	}
}

// Nested Value Access Comparison
func BenchmarkLibraryComparison_Get_Nested_jsJson(b *testing.B) {
	obj := Parse(largeJSON)
	for i := 0; i < b.N; i++ {
		obj.Get("users", 0, "preferences", "privacy", "public").BoolOr(false)
	}
}

func BenchmarkLibraryComparison_Get_Nested_StdLib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if users, ok := largeStd["users"].([]interface{}); ok && len(users) > 0 {
			if user, ok := users[0].(map[string]interface{}); ok {
				if prefs, ok := user["preferences"].(map[string]interface{}); ok {
					if privacy, ok := prefs["privacy"].(map[string]interface{}); ok {
						if public, ok := privacy["public"].(bool); ok {
							_ = public
						}
					}
				}
			}
		}
	}
}

func BenchmarkLibraryComparison_Get_Nested_Gjson(b *testing.B) {
	for i := 0; i < b.N; i++ {
		gjson.Get(largeJSON, "users.0.preferences.privacy.public").Bool()
	}
}

// Marshal/Stringify Comparison
func BenchmarkLibraryComparison_Marshal_jsJson(b *testing.B) {
	data := mediumStd
	for i := 0; i < b.N; i++ {
		Stringify(data)
	}
}

func BenchmarkLibraryComparison_Marshal_StdLib(b *testing.B) {
	data := mediumStd
	for i := 0; i < b.N; i++ {
		json.Marshal(data)
	}
}

func BenchmarkLibraryComparison_Marshal_GoJson(b *testing.B) {
	data := mediumStd
	for i := 0; i < b.N; i++ {
		gojson.Marshal(data)
	}
}

func BenchmarkLibraryComparison_Marshal_JsonIter(b *testing.B) {
	data := mediumStd
	for i := 0; i < b.N; i++ {
		jsoniter.Marshal(data)
	}
}

// Real-world Scenario: Parse + Multiple Gets + Type Conversions
func BenchmarkLibraryComparison_Scenario_jsJson(b *testing.B) {
	for i := 0; i < b.N; i++ {
		obj := Parse(mediumJSON)
		name := obj.Get("name").StringOr("")
		email := obj.Get("email").StringOr("")
		age := obj.Get("age").IntOr(0)
		score := obj.Get("score").Float64Or(0.0)
		active := obj.Get("active").BoolOr(false)
		version := obj.Get("metadata", "version").IntOr(1)
		
		// Use values to prevent optimization
		_ = name + email
		_ = age + version
		_ = score
		_ = active
	}
}

func BenchmarkLibraryComparison_Scenario_StdLib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var data map[string]interface{}
		json.Unmarshal([]byte(mediumJSON), &data)
		
		name, _ := data["name"].(string)
		email, _ := data["email"].(string)
		age := int(data["age"].(float64))
		score, _ := data["score"].(float64)
		active, _ := data["active"].(bool)
		
		metadata, _ := data["metadata"].(map[string]interface{})
		version := int(metadata["version"].(float64))
		
		// Use values to prevent optimization
		_ = name + email
		_ = age + version
		_ = score
		_ = active
	}
}

func BenchmarkLibraryComparison_Scenario_Gjson(b *testing.B) {
	for i := 0; i < b.N; i++ {
		name := gjson.Get(mediumJSON, "name").String()
		email := gjson.Get(mediumJSON, "email").String()
		age := gjson.Get(mediumJSON, "age").Int()
		score := gjson.Get(mediumJSON, "score").Float()
		active := gjson.Get(mediumJSON, "active").Bool()
		version := gjson.Get(mediumJSON, "metadata.version").Int()
		
		// Use values to prevent optimization
		_ = name + email
		_ = int(age) + int(version)
		_ = score
		_ = active
	}
}