# tmvar - TOML Variables Library

## Core Problem Statement

Developers need a way to externalize configuration from environment variables without the complexity of traditional configuration systems. Current solutions require either:
- **Environment variables for everything** (clutters deployment, hard to manage)
- **Heavy configuration frameworks** (Viper, etc. - overkill for simple needs)
- **Hard-coded values** (unmaintainable)

**Gap:** No lightweight system for git-tracked, structured configuration with variable substitution that works as simply as environment variables but with better organization.

## Solution Vision

A **plug-and-play Go library** that works exactly like environment variables but with TOML files:
1. **Internal TOML variable substitution** within TOML files themselves
2. **External Go function calls** to retrieve resolved values  
3. **Zero configuration required** - just import and use
4. **Environment variable override** - standard behavior, no surprises
5. **True simplicity** - works like `os.Getenv()` but with structured configuration

## Package Design Philosophy

### Core Principles
- **Plug-and-play:** Import library, create TOML files, start using - nothing else required
- **Zero project pollution:** Library creates no files in user projects
- **Environment variable behavior:** Works exactly like `os.Getenv()` - no prompts, no modes, just works
- **Maximum simplicity:** Simple syntax, minimal API surface, predictable behavior
- **Always current values:** Smart file monitoring ensures configuration is never stale
- **Conflict-free:** Package name and syntax chosen to avoid collisions

### Package Name: `tmvar`
**Import:** `import "github.com/user/tmvar"`
**Reasoning:** 
- "TOML Variables" - clear, specific meaning
- Short enough to type frequently
- Unique namespace - no conflicts with existing libraries
- More descriptive than generic names like `cfg` or `config`

## Syntax Design

### External API (Go Code)
```go
// Basic value retrieval
port := tmvar.GetInt("server.port")
host := tmvar.Get("database.host")
enabled := tmvar.GetBool("features.login")

// With default values
timeout := tmvar.GetIntOr("api.timeout", 30)
debug := tmvar.GetBoolOr("app.debug", false)

// Type-safe variants
duration := tmvar.GetDuration("cache.ttl")  // Parses "5m", "30s"
ratio := tmvar.GetFloat("display.ratio")
tags := tmvar.GetStringSlice("app.tags")    // Comma-separated values

// Conflict resolution (when same key exists in multiple files)
appPort := tmvar.GetInt("app.server.port")      # From app.toml
dbPort := tmvar.GetInt("database.server.port")  # From database.toml
```

**Syntax Rationale:** Function-style calls are conflict-free, type-safe, and follow Go conventions.

### Internal TOML Variables
```toml
# config/app.toml
[database]
host = "localhost"
port = 5432
name = "myapp"

[computed]
connection_string = "postgres://{{database.host}}:{{database.port}}/{{database.name}}"
backup_url = "{{computed.connection_string}}_backup"

[paths]
base = "/app"
uploads = "{{paths.base}}/uploads"
logs = "{{paths.base}}/logs"
```

**Syntax Rationale:** `{{variable}}` template style is familiar, reads naturally in configuration context, and has no conflicts within TOML files.

## Environment Variable Injection System

### Explicit Environment Variable Integration
Environment variables are explicitly referenced within TOML files using the `{{ENV.VARIABLE}}` syntax:
- **Explicit references:** `{{ENV.PORT:-3000}}` directly in TOML files
- **Default values:** Built-in fallback syntax when environment variable is not set
- **Git-trackable:** All environment variable usage documented in TOML files
- **No magic mapping:** Clear, visible integration without hidden behavior

**Syntax Examples:**
```toml
[server]
port = "{{ENV.PORT:-3000}}"
host = "{{ENV.HOST:-localhost}}"

[database]
url = "{{ENV.DATABASE_URL:-postgres://localhost:5432/myapp}}"
password = "{{ENV.DB_PASSWORD}}"  # No default - required env var
```

**Behavior:** Explicit, self-documenting environment variable integration with built-in default value support.

## File Discovery & Resolution

### Discovery Strategy
- **Recursive search:** Find all `*.toml` files in project directory and subdirectories
- **Process on-demand:** Only parse files that contain referenced variables
- **No configuration required:** No need to specify which files to include
- **Ignore unused files:** Regular TOML files that aren't referenced remain untouched

### Conflict Resolution
**When same variable exists in multiple files:**
```
Error: Variable "server.port" found in multiple files:
- config/app.toml
- settings/database.toml

Use explicit syntax:
- tmvar.Get("app.server.port")
- tmvar.Get("database.server.port")
```

**Explicit file reference syntax:** `filename.section.key`
- File extension `.toml` is omitted in reference
- Path separators become dots: `config/app.toml` → `config.app`

### Variable Resolution Process
1. **Parse referenced TOML file(s)**
2. **Resolve internal variables** using `{{}}` syntax
3. **Handle cross-references** between sections
4. **Return final resolved value** to Go code
5. **Cache resolved values** according to caching mode

## Simple Caching Strategy

### Smart File Monitoring (Just Like Environment Variables)
- **Check file timestamps** on every variable access
- **Reload only when files actually change**  
- **Cache resolved values** until files are modified
- **Always return current values** - no stale configuration
- **Minimal performance overhead** (~50ns per access for timestamp check)

**Behavior:** Works exactly like environment variables - always current, good performance, zero user configuration required.

**Performance Characteristics:**
- Unchanged files: Memory lookup only (instant)
- Changed files: Automatic reload and re-resolution  
- File timestamp check: ~50ns overhead (negligible)
- No user decisions about caching modes required

## Error Handling & User Experience

### Error Verbosity Level: Helpful but Not Overwhelming
```
Error: Variable "database.host" not found

Searched in:
- config/app.toml
- settings/database.toml

Available variables:
- database.port
- database.name
```

**Error Principles:**
- **Actionable:** Tell user exactly what to fix
- **Contextual:** Show which files were searched
- **Helpful:** Suggest available alternatives
- **Concise:** No verbose stack traces or debug info
- **Consistent:** Same format for all variable resolution errors

### Internal Variable Resolution Errors
```
Error in config/app.toml: Variable "{{invalid.reference}}" not found

Available in same file:
- database.host
- database.port
- paths.base
```

### File Discovery Errors
- Missing files: Treated as empty (no error - might not exist yet)
- Malformed TOML: Clear syntax error with file and line number
- Permission errors: Clear message about file access issues

## API Specification

### Core Functions
```go
// Basic retrieval (panics if not found - fail fast)
func Get(key string) string
func GetInt(key string) int  
func GetBool(key string) bool
func GetFloat(key string) float64
func GetDuration(key string) time.Duration

// Safe retrieval with defaults (never panics)
func GetOr(key string, defaultValue string) string
func GetIntOr(key string, defaultValue int) int
func GetBoolOr(key string, defaultValue bool) bool
func GetFloatOr(key string, defaultValue float64) float64
func GetDurationOr(key string, defaultValue time.Duration) time.Duration

// Collection types
func GetStringSlice(key string) []string // Parses comma-separated values
func GetIntSlice(key string) []int

// Advanced usage
func Exists(key string) bool // Check if variable exists without retrieving
func GetAll() map[string]interface{} // Get all resolved variables
```

### Type Conversion Rules
- **Strings:** Direct value
- **Integers:** Parse numeric strings, error on invalid format
- **Booleans:** Parse "true"/"false", "yes"/"no", "1"/"0" (case insensitive)
- **Durations:** Parse Go duration format ("5m", "30s", "2h")
- **Floats:** Parse decimal numbers
- **String slices:** Split on comma, trim whitespace from each element

## Internal Variable Processing

### Variable Resolution Order
1. **Parse TOML structure** into nested map
2. **Identify variable references** using `{{section.key}}` pattern
3. **Resolve in dependency order** (definition before use)
4. **Support cross-section references** within same file
5. **Handle nested substitutions** recursively
6. **Return final resolved values**

### Supported Reference Patterns
```toml
# Same section reference
[database]
host = "localhost"
url = "postgres://{{database.host}}/myapp"

# Cross-section reference  
[paths]
base = "/app"

[uploads]
dir = "{{paths.base}}/uploads"

# Nested references
[computed]
db_url = "postgres://{{database.host}}:{{database.port}}/{{database.name}}"
backup_db = "{{computed.db_url}}_backup"
```

### Variable Resolution Constraints
- **No forward references:** Variables must be defined before use
- **No circular dependencies:** A → B → A relationships detected and reported
- **Single file scope:** `{{}}` references only resolve within same TOML file
- **String values only:** Internal variables resolve to strings (type conversion happens at Go API level)

## Implementation Phases

### Phase 1: Core Foundation (Week 1)
**Goals:** Basic external resolution with file discovery
- Implement `tmvar.Get()`, `tmvar.GetInt()`, `tmvar.GetBool()`
- Basic recursive TOML file discovery 
- Smart file monitoring (timestamp-based reload)
- Environment variable override behavior
- Level 2 error handling

**Success Criteria:**
```go
port := tmvar.GetInt("server.port")  // Works with smart caching
host := tmvar.Get("database.host")   // Always current values
// Internal variables - not yet
```

### Phase 2: Internal Templates (Week 2)  
**Goals:** TOML internal variable processing
- Implement `{{variable}}` resolution within TOML files
- Cross-section references (`{{section.key}}`)
- Dependency resolution (definition-before-use)
- Error handling for invalid references

**Success Criteria:**
```toml
# This works with automatic reload when changed:
connection = "postgres://{{db.host}}:{{db.port}}/{{db.name}}"
```

### Phase 3: Advanced Features (Week 3)
**Goals:** Complete API and conflict resolution
- Conflict resolution (`filename.section.key`)
- Advanced type handling (`GetDuration`, slices)
- `GetOr` functions with defaults
- Comprehensive error messages

**Success Criteria:**
- All API functions work correctly
- File conflicts resolved clearly
- Environment variable overrides work seamlessly

### Phase 4: Production Polish (Week 4)
**Goals:** Performance optimization and edge cases
- Performance optimization for file monitoring
- Edge case handling and error resilience
- Documentation and examples
- Cross-platform compatibility testing

## Usage Patterns & Examples

### Simple Configuration
```toml
# config.toml
[server]
port = 3000
host = "localhost"

[database]  
url = "postgres://localhost:5432/myapp"

[features]
login_required = true
debug_mode = false
```

```go
// main.go
import "github.com/user/tmvar"

func main() {
    port := tmvar.GetIntOr("server.port", 8080)
    dbURL := tmvar.Get("database.url")
    loginRequired := tmvar.GetBool("features.login_required")
    
    fmt.Printf("Server starting on port %d\n", port)
}
```

### Advanced Variable Substitution
```toml
# config/app.toml
[database]
host = "{{ENV.DB_HOST:-localhost}}"
port = "{{ENV.DB_PORT:-5432}}"
name = "{{ENV.DB_NAME:-myapp}}"
user = "{{ENV.DB_USER:-admin}}"

[connection]
primary = "postgres://{{database.user}}@{{database.host}}:{{database.port}}/{{database.name}}"
backup = "{{connection.primary}}_backup"
readonly = "{{connection.primary}}?sslmode=require"

[paths]
base = "{{ENV.APP_BASE_PATH:-/app}}"
uploads = "{{paths.base}}/uploads"
logs = "{{paths.base}}/logs"
cache = "{{paths.base}}/cache"
```

```go
// Usage
dbURL := tmvar.Get("connection.primary")
// Returns: "postgres://admin@localhost:5432/myapp"

uploadPath := tmvar.Get("paths.uploads") 
// Returns: "/app/uploads"
```

### Multi-File Organization
```
project/
├── config/
│   ├── app.toml          # App-specific settings
│   ├── database.toml     # Database configuration  
│   └── features.toml     # Feature flags
├── settings/
│   └── deployment.toml   # Deployment-specific
└── main.go
```

```go
// Automatic discovery - no configuration needed
serverPort := tmvar.GetInt("server.port")      // Finds in any .toml file
dbHost := tmvar.Get("database.host")           // Searches all files

// Explicit file reference if conflicts
appPort := tmvar.GetInt("app.server.port")     # From config/app.toml  
deployPort := tmvar.GetInt("deployment.server.port") # From settings/deployment.toml
```

## Integration Guidelines

### Environment Variable Integration Examples
```bash
# TOML file has: port = "{{ENV.PORT:-3000}}"
# Environment variable provides value
PORT=8080 go run main.go
# tmvar.GetInt("server.port") returns 8080

# Standard application environment variables work as expected
NODE_ENV=production DATABASE_URL=postgres://... go run main.go

# Docker - explicit environment variable usage
docker run -e NODE_ENV=production -e DATABASE_URL=postgres://prod... myapp

# No environment variables - uses TOML defaults
go run main.go
# tmvar.GetInt("server.port") returns 3000 (default from {{ENV.PORT:-3000}})
```

### Simple Usage Pattern
```go
// Environment variables explicitly referenced in TOML
port := tmvar.GetInt("server.port")    // Resolves {{ENV.PORT:-3000}}
dbURL := tmvar.Get("database.url")     // Resolves {{ENV.DATABASE_URL:-default}}
debug := tmvar.GetBool("app.debug")    // Resolves {{ENV.DEBUG:-false}}

// All environment variable usage visible in TOML files
// No magic overrides - everything explicit and git-tracked
```

### Project Structure Recommendations
- **Configuration files:** Any directory structure works (`config/`, `settings/`, root, etc.)
- **Naming convention:** Descriptive TOML filenames (`app.toml`, `database.toml`, `features.toml`)
- **Organization:** Group related settings in same file, use separate files for distinct concerns
- **Version control:** Commit TOML files (unlike `.env` files), exclude sensitive data

### Error Handling Patterns
```go
// Fail-fast approach (recommended for required configuration)
port := tmvar.GetInt("server.port") // Panics if missing - catches config errors early

// Safe approach (for optional configuration)  
timeout := tmvar.GetIntOr("api.timeout", 30) // Provides sensible default

// Check existence before use
if tmvar.Exists("features.experimental") {
    experimental := tmvar.GetBool("features.experimental")
    // Use experimental features
}
```

## Technical Requirements

### Dependencies
- **Standard library only** for core functionality
- **github.com/BurntSushi/toml** for TOML parsing (proven, stable)
- **No external dependencies** for terminal detection or file operations
- **Cross-platform compatibility** (Windows, macOS, Linux)

### Performance Targets  
- **File discovery:** < 10ms for typical projects (< 100 TOML files)
- **Variable resolution:** < 1ms for cached lookups (unchanged files)
- **File monitoring overhead:** < 100ns per access (timestamp check only)
- **Memory usage:** < 5MB for typical configuration sets
- **Startup impact:** < 50ms additional startup time

### Compatibility Requirements
- **Go version:** Go 1.19+ (for modern file system APIs)
- **Terminal compatibility:** Works in all major terminals, IDEs, and non-interactive environments
- **Docker compatibility:** Seamless operation in containerized environments
- **CI/CD compatibility:** No hanging or user interaction required in automation

## Success Metrics

### User Experience Goals
- **Zero learning curve:** Works exactly like environment variables - developers productive immediately  
- **Zero configuration required:** Import and use immediately, no setup
- **Predictable behavior:** Same as `os.Getenv()` behavior users already understand
- **Clear error messages:** Configuration problems obvious and fixable
- **No surprises:** Always current values, environment variables override TOML

### Technical Goals  
- **Environment variable simplicity:** `tmvar.Get()` works just like `os.Getenv()`
- **Always current values:** Smart file monitoring without user intervention
- **Conflict resolution:** Clear path forward when configuration conflicts arise
- **Standard override behavior:** Environment variables work exactly as expected

This library fills the gap between environment variables (good for secrets, bad for structured config) and heavy configuration frameworks (powerful but complex). It provides structured, git-tracked configuration with exactly the same simplicity as `os.Getenv()` but with the power of variable substitution and structured data.