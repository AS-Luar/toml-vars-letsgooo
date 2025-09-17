package tmvar

import (
	"fmt"
	"os"
	"strings"
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

func TestVariableSubstitution(t *testing.T) {
	// Create a test TOML file with variable references
	testFile := "test_variables.toml"
	content := `
[database]
host = "localhost"
port = 5432
name = "testdb"

[connection]
primary = "postgres://{{database.host}}:{{database.port}}/{{database.name}}"
backup = "{{connection.primary}}_backup"

[paths]
base = "/app"
uploads = "{{paths.base}}/uploads"
logs = "{{paths.base}}/logs"
`

	// Write test file
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Clear cache before testing
	clearCache()

	// Test basic variable substitution
	if got := Get("connection.primary"); got != "postgres://localhost:5432/testdb" {
		t.Errorf("Get(\"connection.primary\") = %v, want %v", got, "postgres://localhost:5432/testdb")
	}

	// Test nested variable substitution
	if got := Get("connection.backup"); got != "postgres://localhost:5432/testdb_backup" {
		t.Errorf("Get(\"connection.backup\") = %v, want %v", got, "postgres://localhost:5432/testdb_backup")
	}

	// Test cross-section references
	if got := Get("paths.uploads"); got != "/app/uploads" {
		t.Errorf("Get(\"paths.uploads\") = %v, want %v", got, "/app/uploads")
	}
}

func TestCircularDependencyDetection(t *testing.T) {
	// Create a test TOML file with circular dependencies
	testFile := "test_circular.toml"
	content := `
[circular]
a = "{{circular.b}}"
b = "{{circular.a}}"
`

	// Write test file
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Clear cache before testing
	clearCache()

	// This should panic with circular dependency error
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for circular dependency, but didn't panic")
		} else {
			// Check that the error message mentions circular dependency
			errorMsg := fmt.Sprintf("%v", r)
			if !strings.Contains(errorMsg, "circular dependency") {
				t.Errorf("Expected circular dependency error, got: %v", errorMsg)
			}
		}
	}()

	Get("circular.a")
}

func TestUndefinedVariableReference(t *testing.T) {
	// Create a test TOML file with undefined variable reference
	testFile := "test_undefined.toml"
	content := `
[test]
value = "{{missing.variable}}"
existing = "valid"
`

	// Write test file
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Clear cache before testing
	clearCache()

	// This should panic with undefined variable error
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for undefined variable, but didn't panic")
		} else {
			// Check that the error message mentions the missing variable
			errorMsg := fmt.Sprintf("%v", r)
			if !strings.Contains(errorMsg, "missing.variable") {
				t.Errorf("Expected undefined variable error, got: %v", errorMsg)
			}
		}
	}()

	Get("test.value")
}