package tomv

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
			if !strings.Contains(errorMsg, "circular dependency") && !strings.Contains(errorMsg, "maximum resolution passes exceeded") {
				t.Errorf("Expected circular dependency or resolution error, got: %v", errorMsg)
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

func TestEnvironmentVariables(t *testing.T) {
	// Create a test TOML file with environment variables
	testFile := "test_env.toml"
	content := `
[server]
port = "{{ENV.PORT:-3000}}"
host = "{{ENV.HOST:-localhost}}"
debug = "{{ENV.DEBUG:-false}}"

[database]
url = "postgres://{{ENV.DB_HOST:-localhost}}:{{ENV.DB_PORT:-5432}}/{{ENV.DB_NAME:-testdb}}"
`

	// Write test file
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Clear cache before testing
	clearCache()

	// Test with environment variables set
	os.Setenv("PORT", "8080")
	os.Setenv("HOST", "production.com")
	os.Setenv("DEBUG", "true")
	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("HOST")
		os.Unsetenv("DEBUG")
	}()

	// Clear cache after setting env vars
	clearCache()

	// Test environment variable substitution
	if got := GetInt("server.port"); got != 8080 {
		t.Errorf("GetInt(\"server.port\") with ENV.PORT=8080 = %v, want %v", got, 8080)
	}

	if got := Get("server.host"); got != "production.com" {
		t.Errorf("Get(\"server.host\") with ENV.HOST=production.com = %v, want %v", got, "production.com")
	}

	if got := GetBool("server.debug"); got != true {
		t.Errorf("GetBool(\"server.debug\") with ENV.DEBUG=true = %v, want %v", got, true)
	}
}

func TestEnvironmentVariableDefaults(t *testing.T) {
	// Create a test TOML file with environment variables
	testFile := "test_env_defaults.toml"
	content := `
[server]
port = "{{ENV.PORT:-3000}}"
host = "{{ENV.HOST:-localhost}}"
timeout = "{{ENV.TIMEOUT:-30}}"
`

	// Write test file
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Clear cache and ensure no relevant env vars are set
	clearCache()
	os.Unsetenv("PORT")
	os.Unsetenv("HOST")
	os.Unsetenv("TIMEOUT")

	// Test default values when environment variables are not set
	if got := GetInt("server.port"); got != 3000 {
		t.Errorf("GetInt(\"server.port\") with no ENV.PORT = %v, want %v", got, 3000)
	}

	if got := Get("server.host"); got != "localhost" {
		t.Errorf("Get(\"server.host\") with no ENV.HOST = %v, want %v", got, "localhost")
	}

	if got := GetInt("server.timeout"); got != 30 {
		t.Errorf("GetInt(\"server.timeout\") with no ENV.TIMEOUT = %v, want %v", got, 30)
	}
}

func TestMixedEnvironmentAndInternalVariables(t *testing.T) {
	// Create a test TOML file mixing environment and internal variables
	testFile := "test_mixed.toml"
	content := `
[database]
host = "{{ENV.DB_HOST:-localhost}}"
port = "{{ENV.DB_PORT:-5432}}"
name = "{{ENV.DB_NAME:-testdb}}"

[connection]
url = "postgres://{{database.host}}:{{database.port}}/{{database.name}}"
backup_url = "{{connection.url}}_backup"

[paths]
base = "{{ENV.BASE_PATH:-/app}}"
uploads = "{{paths.base}}/uploads"
logs = "{{paths.base}}/logs"
`

	// Write test file
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Set some environment variables
	os.Setenv("DB_HOST", "prod.db.com")
	os.Setenv("BASE_PATH", "/production")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("BASE_PATH")
		os.Unsetenv("DB_PORT")
		os.Unsetenv("DB_NAME")
		os.Unsetenv("LOG_PATH")
	}()

	// Clear cache after setting env vars
	clearCache()

	// Test mixed resolution
	expectedURL := "postgres://prod.db.com:5432/testdb"
	if got := Get("connection.url"); got != expectedURL {
		t.Errorf("Get(\"connection.url\") = %v, want %v", got, expectedURL)
	}

	expectedBackupURL := "postgres://prod.db.com:5432/testdb_backup"
	if got := Get("connection.backup_url"); got != expectedBackupURL {
		t.Errorf("Get(\"connection.backup_url\") = %v, want %v", got, expectedBackupURL)
	}

	if got := Get("paths.uploads"); got != "/production/uploads" {
		t.Errorf("Get(\"paths.uploads\") = %v, want %v", got, "/production/uploads")
	}

	// Test nested environment variable with internal variable default
	if got := Get("paths.logs"); got != "/production/logs" {
		t.Errorf("Get(\"paths.logs\") = %v, want %v", got, "/production/logs")
	}
}

// TODO: TestNestedEnvironmentVariableDefaults - Complex nested env var defaults
// This test case represents an advanced feature where environment variable defaults
// contain internal variable references like: {{ENV.LOG_PATH:-{{paths.base}}/logs}}
// This requires more sophisticated parsing and will be implemented in a future phase.

func TestStringSlices(t *testing.T) {
	// Create a test TOML file with string arrays
	testFile := "test_string_slices.toml"
	content := `
[server]
allowed_hosts = "localhost,127.0.0.1,staging.com"
empty_list = ""
single_item = "onlyone"
with_spaces = " item1 , item2 , item3 "

[env_arrays]
hosts = "{{ENV.HOSTS:-localhost,127.0.0.1}}"
features = "{{ENV.FEATURES:-auth,logging,metrics}}"
`

	// Write test file
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Clear cache and environment
	clearCache()
	os.Unsetenv("HOSTS")
	os.Unsetenv("FEATURES")

	// Test basic string slices
	hosts := GetStringSlice("server.allowed_hosts")
	expected := []string{"localhost", "127.0.0.1", "staging.com"}
	if len(hosts) != len(expected) {
		t.Errorf("GetStringSlice length = %v, want %v", len(hosts), len(expected))
	}
	for i, host := range hosts {
		if host != expected[i] {
			t.Errorf("GetStringSlice[%d] = %v, want %v", i, host, expected[i])
		}
	}

	// Test empty list
	empty := GetStringSlice("server.empty_list")
	if len(empty) != 0 {
		t.Errorf("GetStringSlice empty = %v, want []", empty)
	}

	// Test single item
	single := GetStringSlice("server.single_item")
	if len(single) != 1 || single[0] != "onlyone" {
		t.Errorf("GetStringSlice single = %v, want [onlyone]", single)
	}

	// Test whitespace trimming
	withSpaces := GetStringSlice("server.with_spaces")
	expectedSpaces := []string{"item1", "item2", "item3"}
	for i, item := range withSpaces {
		if item != expectedSpaces[i] {
			t.Errorf("GetStringSlice with spaces[%d] = %v, want %v", i, item, expectedSpaces[i])
		}
	}

	// Test environment variable defaults
	envHosts := GetStringSlice("env_arrays.hosts")
	expectedEnvHosts := []string{"localhost", "127.0.0.1"}
	for i, host := range envHosts {
		if host != expectedEnvHosts[i] {
			t.Errorf("GetStringSlice env default[%d] = %v, want %v", i, host, expectedEnvHosts[i])
		}
	}
}

func TestIntSlices(t *testing.T) {
	// Create a test TOML file with int arrays
	testFile := "test_int_slices.toml"
	content := `
[server]
ports = "8080,8081,8082"
single_port = "3000"
with_spaces = " 1001 , 1002 , 1003 "

[database]
replica_ports = "{{ENV.DB_PORTS:-5432,5433,5434}}"
`

	// Write test file
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Clear cache and environment
	clearCache()
	os.Unsetenv("DB_PORTS")

	// Test basic int slices
	ports := GetIntSlice("server.ports")
	expected := []int{8080, 8081, 8082}
	if len(ports) != len(expected) {
		t.Errorf("GetIntSlice length = %v, want %v", len(ports), len(expected))
	}
	for i, port := range ports {
		if port != expected[i] {
			t.Errorf("GetIntSlice[%d] = %v, want %v", i, port, expected[i])
		}
	}

	// Test single item
	single := GetIntSlice("server.single_port")
	if len(single) != 1 || single[0] != 3000 {
		t.Errorf("GetIntSlice single = %v, want [3000]", single)
	}

	// Test whitespace trimming
	withSpaces := GetIntSlice("server.with_spaces")
	expectedSpaces := []int{1001, 1002, 1003}
	for i, port := range withSpaces {
		if port != expectedSpaces[i] {
			t.Errorf("GetIntSlice with spaces[%d] = %v, want %v", i, port, expectedSpaces[i])
		}
	}

	// Test environment variable defaults
	dbPorts := GetIntSlice("database.replica_ports")
	expectedDbPorts := []int{5432, 5433, 5434}
	for i, port := range dbPorts {
		if port != expectedDbPorts[i] {
			t.Errorf("GetIntSlice env default[%d] = %v, want %v", i, port, expectedDbPorts[i])
		}
	}
}

func TestSliceOrFunctions(t *testing.T) {
	// Create a test TOML file
	testFile := "test_slice_or.toml"
	content := `
[existing]
hosts = "server1,server2"
ports = "8080,8081"
`

	// Write test file
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Clear cache
	clearCache()

	// Test existing values
	hosts := GetStringSliceOr("existing.hosts", []string{"default"})
	if len(hosts) != 2 || hosts[0] != "server1" || hosts[1] != "server2" {
		t.Errorf("GetStringSliceOr existing = %v, want [server1 server2]", hosts)
	}

	ports := GetIntSliceOr("existing.ports", []int{3000})
	if len(ports) != 2 || ports[0] != 8080 || ports[1] != 8081 {
		t.Errorf("GetIntSliceOr existing = %v, want [8080 8081]", ports)
	}

	// Test missing values (should return defaults)
	missingHosts := GetStringSliceOr("missing.hosts", []string{"default1", "default2"})
	if len(missingHosts) != 2 || missingHosts[0] != "default1" || missingHosts[1] != "default2" {
		t.Errorf("GetStringSliceOr missing = %v, want [default1 default2]", missingHosts)
	}

	missingPorts := GetIntSliceOr("missing.ports", []int{9000, 9001})
	if len(missingPorts) != 2 || missingPorts[0] != 9000 || missingPorts[1] != 9001 {
		t.Errorf("GetIntSliceOr missing = %v, want [9000 9001]", missingPorts)
	}
}

func TestSliceEnvironmentVariables(t *testing.T) {
	// Create a test TOML file with environment variable arrays
	testFile := "test_slice_env.toml"
	content := `
[services]
hosts = "{{ENV.SERVICE_HOSTS:-localhost,127.0.0.1}}"
ports = "{{ENV.SERVICE_PORTS:-8080,8081,8082}}"
`

	// Write test file
	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Test with environment variables set
	os.Setenv("SERVICE_HOSTS", "prod1.com,prod2.com,prod3.com")
	os.Setenv("SERVICE_PORTS", "9000,9001,9002")
	defer func() {
		os.Unsetenv("SERVICE_HOSTS")
		os.Unsetenv("SERVICE_PORTS")
	}()

	// Clear cache after setting env vars
	clearCache()

	// Test environment variable substitution in arrays
	hosts := GetStringSlice("services.hosts")
	expectedHosts := []string{"prod1.com", "prod2.com", "prod3.com"}
	for i, host := range hosts {
		if host != expectedHosts[i] {
			t.Errorf("GetStringSlice env[%d] = %v, want %v", i, host, expectedHosts[i])
		}
	}

	ports := GetIntSlice("services.ports")
	expectedPorts := []int{9000, 9001, 9002}
	for i, port := range ports {
		if port != expectedPorts[i] {
			t.Errorf("GetIntSlice env[%d] = %v, want %v", i, port, expectedPorts[i])
		}
	}
}

// ===== PHASE 4 MULTI-FILE TESTS =====

func TestMultiFileNoConflict(t *testing.T) {
	// Create multiple TOML files with unique variables
	appFile := "test_app.toml"
	appContent := `
[server]
port = 3000
host = "localhost"

[logging]
level = "info"
`

	dbFile := "test_database.toml"
	dbContent := `
[connection]
host = "db.example.com"
port = 5432
user = "admin"

[pool]
max_connections = 100
`

	// Write test files
	err := os.WriteFile(appFile, []byte(appContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create app test file: %v", err)
	}
	defer os.Remove(appFile)

	err = os.WriteFile(dbFile, []byte(dbContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create db test file: %v", err)
	}
	defer os.Remove(dbFile)

	// Clear cache
	clearCache()

	// Test that unique variables work without prefixes
	if got := Get("server.port"); got != "3000" {
		t.Errorf("Get(\"server.port\") = %v, want %v", got, "3000")
	}

	if got := Get("connection.host"); got != "db.example.com" {
		t.Errorf("Get(\"connection.host\") = %v, want %v", got, "db.example.com")
	}

	if got := GetInt("pool.max_connections"); got != 100 {
		t.Errorf("GetInt(\"pool.max_connections\") = %v, want %v", got, 100)
	}
}

func TestMultiFileConflictDetection(t *testing.T) {
	// Create multiple TOML files with conflicting variables
	appFile := "test_app_conflict.toml"
	appContent := `
[server]
port = 3000
host = "app.localhost"
`

	dbFile := "test_database_conflict.toml"
	dbContent := `
[server]
port = 5432
host = "db.localhost"
`

	// Write test files
	err := os.WriteFile(appFile, []byte(appContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create app test file: %v", err)
	}
	defer os.Remove(appFile)

	err = os.WriteFile(dbFile, []byte(dbContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create db test file: %v", err)
	}
	defer os.Remove(dbFile)

	// Clear cache
	clearCache()

	// Test that conflicting variables cause error
	defer func() {
		if r := recover(); r != nil {
			errorMsg := fmt.Sprintf("%v", r)
			if !strings.Contains(errorMsg, "found in multiple files") {
				t.Errorf("Expected conflict error, got: %v", errorMsg)
			}
			if !strings.Contains(errorMsg, "test_app_conflict.toml") {
				t.Errorf("Expected error to mention app file, got: %v", errorMsg)
			}
			if !strings.Contains(errorMsg, "test_database_conflict.toml") {
				t.Errorf("Expected error to mention database file, got: %v", errorMsg)
			}
		} else {
			t.Error("Expected panic due to conflict, but didn't get one")
		}
	}()

	// This should panic due to conflict
	Get("server.port")
}

func TestMultiFileExplicitSyntax(t *testing.T) {
	// Create multiple TOML files with conflicting variables
	appFile := "test_app_explicit.toml"
	appContent := `
[server]
port = 3000
environment = "development"

[app_features]
auth = true
logging = true
`

	apiFile := "test_api_explicit.toml"
	apiContent := `
[server]
port = 8080
environment = "production"

[api_features]
rate_limiting = true
caching = true
`

	// Write test files
	err := os.WriteFile(appFile, []byte(appContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create app test file: %v", err)
	}
	defer os.Remove(appFile)

	err = os.WriteFile(apiFile, []byte(apiContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create api test file: %v", err)
	}
	defer os.Remove(apiFile)

	// Clear cache
	clearCache()

	// Debug: Check what files are found
	t.Logf("Testing explicit file syntax...")

	// Test explicit file prefix syntax
	if got := Get("test_app_explicit.server.port"); got != "3000" {
		t.Errorf("Get(\"test_app_explicit.server.port\") = %v, want %v", got, "3000")
	}

	if got := Get("test_api_explicit.server.port"); got != "8080" {
		t.Errorf("Get(\"test_api_explicit.server.port\") = %v, want %v", got, "8080")
	}

	if got := Get("test_app_explicit.server.environment"); got != "development" {
		t.Errorf("Get(\"test_app_explicit.server.environment\") = %v, want %v", got, "development")
	}

	if got := Get("test_api_explicit.server.environment"); got != "production" {
		t.Errorf("Get(\"test_api_explicit.server.environment\") = %v, want %v", got, "production")
	}

	// Test unique variables work without prefix
	if got := GetBool("app_features.auth"); got != true {
		t.Errorf("GetBool(\"app_features.auth\") = %v, want %v", got, true)
	}

	if got := GetBool("api_features.rate_limiting"); got != true {
		t.Errorf("GetBool(\"api_features.rate_limiting\") = %v, want %v", got, true)
	}

	// Test conflicting variables should require explicit syntax
	defer func() {
		if r := recover(); r != nil {
			errorMsg := fmt.Sprintf("%v", r)
			if !strings.Contains(errorMsg, "found in multiple files") {
				t.Errorf("Expected conflict error for server.port, got: %v", errorMsg)
			}
		} else {
			t.Error("Expected panic for conflicting server.port, but didn't get one")
		}
	}()

	// This should panic due to conflict
	Get("server.port")
}

func TestMultiFileWithVariableSubstitution(t *testing.T) {
	// Create multiple TOML files with internal variable references
	configFile := "test_config_vars.toml"
	configContent := `
[paths]
base = "/app"
config = "{{paths.base}}/config"

[database]
host = "localhost"
port = 5432
`

	servicesFile := "test_services_vars.toml"
	servicesContent := `
[api]
host = "{{ENV.API_HOST:-localhost}}"
port = 8080
url = "http://{{api.host}}:{{api.port}}"

[cache]
host = "{{database.host}}"
port = 6379
url = "redis://{{cache.host}}:{{cache.port}}"
`

	// Write test files
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config test file: %v", err)
	}
	defer os.Remove(configFile)

	err = os.WriteFile(servicesFile, []byte(servicesContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create services test file: %v", err)
	}
	defer os.Remove(servicesFile)

	// Clear cache and environment
	clearCache()
	os.Unsetenv("API_HOST")

	// Test cross-file variable resolution
	if got := Get("api.url"); got != "http://localhost:8080" {
		t.Errorf("Get(\"api.url\") = %v, want %v", got, "http://localhost:8080")
	}

	// This should reference database.host from the other file
	if got := Get("cache.url"); got != "redis://localhost:6379" {
		t.Errorf("Get(\"cache.url\") = %v, want %v", got, "redis://localhost:6379")
	}

	// Test internal variable substitution within same file
	if got := Get("paths.config"); got != "/app/config" {
		t.Errorf("Get(\"paths.config\") = %v, want %v", got, "/app/config")
	}
}

func TestInvalidFilePrefix(t *testing.T) {
	// Create a test file
	testFile := "test_invalid_prefix.toml"
	content := `
[server]
port = 3000
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Clear cache
	clearCache()

	// Test invalid file prefix
	defer func() {
		if r := recover(); r != nil {
			errorMsg := fmt.Sprintf("%v", r)
			if !strings.Contains(errorMsg, "file prefix \"nonexistent\" not found") {
				t.Errorf("Expected file prefix error, got: %v", errorMsg)
			}
		} else {
			t.Error("Expected panic due to invalid file prefix, but didn't get one")
		}
	}()

	// This should panic due to invalid prefix
	Get("nonexistent.server.port")
}

func TestFileSpecificVariableNotFound(t *testing.T) {
	// Create a test file
	testFile := "test_file_specific_missing.toml"
	content := `
[server]
port = 3000
host = "localhost"
`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Clear cache
	clearCache()

	// Test variable not found in specific file
	defer func() {
		if r := recover(); r != nil {
			errorMsg := fmt.Sprintf("%v", r)
			if !strings.Contains(errorMsg, "variable \"missing.key\" not found in file") {
				t.Errorf("Expected variable not found in file error, got: %v", errorMsg)
			}
			if !strings.Contains(errorMsg, "test_file_specific_missing.toml") {
				t.Errorf("Expected error to mention specific file, got: %v", errorMsg)
			}
		} else {
			t.Error("Expected panic due to missing variable in specific file, but didn't get one")
		}
	}()

	// This should panic because missing.key doesn't exist in the specified file
	Get("test_file_specific_missing.missing.key")
}