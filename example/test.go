package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ktbsomen/jsjson" 
)

// -------------------- Example Structs --------------------

type User struct {
	Name    string   `json:"name"`
	Age     int      `json:"age"`
	Active  bool     `json:"active"`
	Tags    []string `json:"tags"`
	Profile Profile  `json:"profile"`
}

type Profile struct {
	Email    string `json:"email"`
	Bio      string `json:"bio"`
	Location string `json:"location,omitempty"`
}

type Company struct {
	Name      string `json:"name"`
	Founded   int    `json:"founded"`
	Employees []User `json:"employees"`
	Address   struct {
		Street  string `json:"street"`
		City    string `json:"city"`
		Country string `json:"country"`
	} `json:"address"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
	Meta    struct {
		Page       int `json:"page"`
		TotalPages int `json:"total_pages"`
	} `json:"meta"`
}

// -------------------- Example Data --------------------

const userJSON = `{
	"name": "John Doe",
	"age": 30,
	"active": true,
	"tags": ["developer", "golang", "json"],
	"profile": {
		"email": "john@example.com",
		"bio": "Senior Go developer with 5+ years experience",
		"location": "San Francisco, CA"
	}
}`

const companyJSON = `{
	"name": "Tech Corp",
	"founded": 2020,
	"address": {
		"street": "123 Tech Street",
		"city": "San Francisco",
		"country": "USA"
	},
	"employees": [
		{
			"name": "Alice Smith",
			"age": 28,
			"active": true,
			"tags": ["frontend", "react"],
			"profile": {
				"email": "alice@techcorp.com",
				"bio": "Frontend specialist"
			}
		},
		{
			"name": "Bob Johnson",
			"age": 35,
			"active": false,
			"tags": ["backend", "golang", "docker"],
			"profile": {
				"email": "bob@techcorp.com",
				"bio": "DevOps engineer"
			}
		}
	]
}`

const apiResponseJSON = `{
	"success": true,
	"message": "Data retrieved successfully",
	"data": {
		"users": [
			{"name": "User 1", "age": 25},
			{"name": "User 2", "age": 30}
		]
	},
	"meta": {
		"page": 1,
		"total_pages": 5
	}
}`

// -------------------- Example Functions --------------------

func example1_BasicStructParsing() {
	fmt.Println("=== Example 1: Basic Struct Parsing ===")

	// Method 1: Parse with struct destination 
	var user User
	jv := jsjson.Parse(userJSON, &user)
	
	if !jv.IsValid() {
		log.Printf("Parse error: %v", jv.Error())
		return
	}

	fmt.Printf("Parsed User Struct: %+v\n", user)
	fmt.Printf("Profile: %+v\n", user.Profile)

	// You can still use JSONValue methods
	name := jv.Get("name").StringOr("Unknown")
	fmt.Printf("Name from JSONValue: %s\n", name)
}

func example2_HighPerformanceParsing() {
	fmt.Println("\n=== Example 2: High-Performance Direct Parsing ===")

	var user User
	err := jsjson.ParseInto(userJSON, &user)
	if err != nil {
		log.Printf("ParseInto error: %v", err)
		return
	}

	fmt.Printf("User: %s, Age: %d, Email: %s\n", 
		user.Name, user.Age, user.Profile.Email)
	fmt.Printf("Tags: %v\n", user.Tags)
}

func example3_NestedStructs() {
	fmt.Println("\n=== Example 3: Complex Nested Structures ===")

	var company Company
	jv := jsjson.Parse(companyJSON, &company)
	
	if !jv.IsValid() {
		log.Printf("Parse error: %v", jv.Error())
		return
	}

	fmt.Printf("Company: %s (founded %d)\n", company.Name, company.Founded)
	fmt.Printf("Address: %s, %s, %s\n", 
		company.Address.Street, company.Address.City, company.Address.Country)
	
	fmt.Printf("Employees (%d):\n", len(company.Employees))
	for i, emp := range company.Employees {
		fmt.Printf("  %d. %s (%d) - %s - Active: %t\n", 
			i+1, emp.Name, emp.Age, emp.Profile.Email, emp.Active)
		fmt.Printf("     Tags: %v\n", emp.Tags)
	}
}

func example4_DynamicJSONAccess() {
	fmt.Println("\n=== Example 4: Dynamic JSON Access (without structs) ===")

	jv := jsjson.Parse(companyJSON)
	if !jv.IsValid() {
		log.Printf("Parse error: %v", jv.Error())
		return
	}

	// Access nested values dynamically
	companyName := jv.Get("name").StringOr("Unknown Company")
	founded := jv.Get("founded").IntOr(0)
	city := jv.Get("address", "city").StringOr("Unknown City")
	
	fmt.Printf("Company: %s, Founded: %d, City: %s\n", companyName, founded, city)

	// Access array elements
	firstEmployee := jv.Get("employees", 0, "name").StringOr("No employee")
	fmt.Printf("First Employee: %s\n", firstEmployee)

	// Check if keys exist
	fmt.Printf("Has employees: %t\n", jv.Has("employees"))
	fmt.Printf("Has office: %t\n", jv.Has("office"))

	// Iterate through array
	employees, err := jv.Get("employees").Array()
	if err == nil {
		fmt.Printf("Employees via iteration:\n")
		for i, emp := range employees {
			name := emp.Get("name").StringOr("Unknown")
			active := emp.Get("active").BoolOr(false)
			fmt.Printf("  %d. %s (Active: %t)\n", i+1, name, active)
		}
	}
}

func example5_ErrorHandling() {
	fmt.Println("\n=== Example 5: Error Handling ===")

	// Invalid JSON
	invalidJSON := `{"name": "John", "age": 30,}` // trailing comma
	var user User
	jv := jsjson.Parse(invalidJSON, &user)
	
	if !jv.IsValid() {
		fmt.Printf("Expected error for invalid JSON: %v\n", jv.Error())
	}

	// Valid JSON but accessing non-existent keys
	jv2 := jsjson.Parse(userJSON)
	nonExistent := jv2.Get("non_existent_key")
	if !nonExistent.IsValid() {
		fmt.Printf("Expected error for non-existent key: %v\n", nonExistent.Error())
	}

	// Type conversion errors
	ageStr := jv2.Get("name").IntOr(-1) // trying to convert string to int
	fmt.Printf("Name as int (with default): %d\n", ageStr)

	// Safe access with defaults
	location := jv2.Get("profile", "location").StringOr("Not specified")
	fmt.Printf("Location: %s\n", location)
}

func example6_APIResponseHandling() {
	fmt.Println("\n=== Example 6: API Response Handling ===")

	var response APIResponse
	jv := jsjson.Parse(apiResponseJSON, &response)
	
	if !jv.IsValid() {
		log.Printf("Parse error: %v", jv.Error())
		return
	}

	fmt.Printf("API Response - Success: %t, Message: %s\n", 
		response.Success, response.Message)
	fmt.Printf("Pagination: Page %d of %d\n", 
		response.Meta.Page, response.Meta.TotalPages)

	// Access dynamic data
	users := jv.Get("data", "users")
	if users.IsValid() {
		userArray, err := users.Array()
		if err == nil {
			fmt.Printf("Users in response:\n")
			for i, user := range userArray {
				name := user.Get("name").StringOr("Unknown")
				age := user.Get("age").IntOr(0)
				fmt.Printf("  %d. %s (age %d)\n", i+1, name, age)
			}
		}
	}
}

func example7_JSONGeneration() {
	fmt.Println("\n=== Example 7: JSON Generation ===")

	// Create a user struct
	user := User{
		Name:   "Jane Doe",
		Age:    28,
		Active: true,
		Tags:   []string{"manager", "leadership"},
		Profile: Profile{
			Email: "jane@example.com",
			Bio:   "Team lead with excellent communication skills",
		},
	}

	// Convert to JSON string
	jsonStr, err := jsjson.Stringify(user)
	if err != nil {
		log.Printf("Stringify error: %v", err)
		return
	}
	fmt.Printf("Generated JSON: %s\n", jsonStr)

	// Pretty print
	prettyJSON, err := jsjson.StringifyPretty(user, "  ")
	if err != nil {
		log.Printf("StringifyPretty error: %v", err)
		return
	}
	fmt.Printf("Pretty JSON:\n%s\n", prettyJSON)
}

func example8_PerformanceComparison() {
	fmt.Println("\n=== Example 8: Performance Comparison ===")

	iterations := 1000
	
	// Test ParseInto performance
	start := time.Now()
	for i := 0; i < iterations; i++ {
		var user User
		jsjson.ParseInto(userJSON, &user)
	}
	parseIntoDuration := time.Since(start)

	// Test Parse + To performance
	start = time.Now()
	for i := 0; i < iterations; i++ {
		var user User
		jv := jsjson.Parse(userJSON)
		jv.To(&user)
	}
	parseAndToDuration := time.Since(start)

	fmt.Printf("Performance comparison (%d iterations):\n", iterations)
	fmt.Printf("  ParseInto: %v\n", parseIntoDuration)
	fmt.Printf("  Parse+To:  %v\n", parseAndToDuration)
	fmt.Printf("  ParseInto is %.2fx faster\n", 
		float64(parseAndToDuration)/float64(parseIntoDuration))
}

func example9_ChainedOperations() {
	fmt.Println("\n=== Example 9: Chained Operations ===")

	jv := jsjson.Parse(companyJSON)
	
	// Chain multiple operations
	result := jv.Get("employees").
		Get(0).
		Get("profile").
		Get("email").
		StringOr("no-email@example.com")
	
	fmt.Printf("First employee's email: %s\n", result)

	// Check types
	employeesType := jv.Get("employees").Type()
	fmt.Printf("Employees type: %s\n", employeesType)

	// Clone and modify
	cloned := jv.Clone()
	fmt.Printf("Cloned successfully: %t\n", cloned.IsValid())
}

func example10_EdgeCases() {
	fmt.Println("\n=== Example 10: Edge Cases ===")

	// Empty JSON
	emptyObj := jsjson.Parse("{}")
	fmt.Printf("Empty object valid: %t\n", emptyObj.IsValid())

	// Null values
	nullJSON := `{"name": null, "age": 25}`
	jv := jsjson.Parse(nullJSON)
	
	name := jv.Get("name")
	fmt.Printf("Name is null: %t\n", name.IsNull())
	fmt.Printf("Name type: %s\n", name.Type())
	
	// Array access
	arrayJSON := `[1, 2, 3, "four", true]`
	arr := jsjson.Parse(arrayJSON)
	
	fmt.Printf("Array length check via iteration:\n")
	if arrValues, err := arr.Array(); err == nil {
		for i, val := range arrValues {
			fmt.Printf("  [%d]: %v (type: %s)\n", i, val.Raw(), val.Type())
		}
	}
}

// -------------------- Main Function --------------------

func main() {
	fmt.Println("🚀 jsjson Library Examples")
	fmt.Println("==========================")

	example1_BasicStructParsing()
	example2_HighPerformanceParsing()
	example3_NestedStructs()
	example4_DynamicJSONAccess()
	example5_ErrorHandling()
	example6_APIResponseHandling()
	example7_JSONGeneration()
	example8_PerformanceComparison()
	example9_ChainedOperations()
	example10_EdgeCases()

	fmt.Println("\n✅ All examples completed successfully!")
}

// --------------------Custom Usage Patterns --------------------

// Example of a custom helper function
func parseUserSafely(jsonData string) (*User, error) {
	var user User
	if err := jsjson.ParseInto(jsonData, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user: %w", err)
	}
	
	// Validate required fields
	if user.Name == "" {
		return nil, fmt.Errorf("user name is required")
	}
	if user.Profile.Email == "" {
		return nil, fmt.Errorf("user email is required")
	}
	
	return &user, nil
}

// Example of working with dynamic configuration
func parseConfig(configJSON string) map[string]interface{} {
	jv := jsjson.Parse(configJSON)
	if !jv.IsValid() {
		return map[string]interface{}{
			"error": jv.Error().Error(),
		}
	}

	config := make(map[string]interface{})
	
	// Extract common config values with defaults
	config["debug"] = jv.Get("debug").BoolOr(false)
	config["port"] = jv.Get("server", "port").IntOr(8080)
	config["host"] = jv.Get("server", "host").StringOr("localhost")
	config["timeout"] = jv.Get("timeout").IntOr(30)
	
	return config
}