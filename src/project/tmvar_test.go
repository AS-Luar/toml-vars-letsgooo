package tmvar

import (
	"os"
	"testing"
	"time"
)

func TestBasicValueRetrieval(t *testing.T) {
	// Create a test TOML file
	testFile := "test_config.toml"
	content := `
[server]
port = 3000
host = "localhost"
debug = true

[database]
url = "postgres://localhost:5432/test"
timeout = "30s"
ratio = 1.5

[features]
enabled = true
`

	// Write test file
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Clear cache before testing
	clearCache()

	// Test Get
	if got := Get("server.host"); got != "localhost" {
		t.Errorf("Get(\"server.host\") = %v, want %v", got, "localhost")
	}

	// Test GetInt
	if got := GetInt("server.port"); got != 3000 {
		t.Errorf("GetInt(\"server.port\") = %v, want %v", got, 3000)
	}

	// Test GetBool
	if got := GetBool("server.debug"); got != true {
		t.Errorf("GetBool(\"server.debug\") = %v, want %v", got, true)
	}

	// Test GetFloat
	if got := GetFloat("database.ratio"); got != 1.5 {
		t.Errorf("GetFloat(\"database.ratio\") = %v, want %v", got, 1.5)
	}

	// Test GetDuration
	expected := 30 * time.Second
	if got := GetDuration("database.timeout"); got != expected {
		t.Errorf("GetDuration(\"database.timeout\") = %v, want %v", got, expected)
	}
}

func TestGetOrFunctions(t *testing.T) {
	// Create a test TOML file
	testFile := "test_config.toml"
	content := `
[server]
port = 3000
`

	// Write test file
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Clear cache before testing
	clearCache()

	// Test existing value
	if got := GetIntOr("server.port", 8080); got != 3000 {
		t.Errorf("GetIntOr(\"server.port\", 8080) = %v, want %v", got, 3000)
	}

	// Test default value for missing key
	if got := GetIntOr("server.missing", 8080); got != 8080 {
		t.Errorf("GetIntOr(\"server.missing\", 8080) = %v, want %v", got, 8080)
	}

	// Test GetOr with string
	if got := GetOr("server.missing", "default"); got != "default" {
		t.Errorf("GetOr(\"server.missing\", \"default\") = %v, want %v", got, "default")
	}
}

func TestExists(t *testing.T) {
	// Create a test TOML file
	testFile := "test_config.toml"
	content := `
[server]
port = 3000
`

	// Write test file
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Clear cache before testing
	clearCache()

	// Test existing key
	if got := Exists("server.port"); got != true {
		t.Errorf("Exists(\"server.port\") = %v, want %v", got, true)
	}

	// Test missing key
	if got := Exists("server.missing"); got != false {
		t.Errorf("Exists(\"server.missing\") = %v, want %v", got, false)
	}
}

func TestCaching(t *testing.T) {
	// Create a test TOML file
	testFile := "test_config.toml"
	content1 := `
[server]
port = 3000
`
	content2 := `
[server]
port = 4000
`

	// Write initial file
	err := os.WriteFile(testFile, []byte(content1), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Clear cache before testing
	clearCache()

	// First read
	if got := GetInt("server.port"); got != 3000 {
		t.Errorf("First read: GetInt(\"server.port\") = %v, want %v", got, 3000)
	}

	// Simulate file change with a brief delay to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)
	err = os.WriteFile(testFile, []byte(content2), 0644)
	if err != nil {
		t.Fatalf("Failed to update test file: %v", err)
	}

	// Second read should detect change and return new value
	if got := GetInt("server.port"); got != 4000 {
		t.Errorf("After file change: GetInt(\"server.port\") = %v, want %v", got, 4000)
	}
}

func TestErrorHandling(t *testing.T) {
	// Test with no TOML files
	clearCache()

	// This should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for missing variable, but didn't panic")
		}
	}()

	Get("nonexistent.key")
}