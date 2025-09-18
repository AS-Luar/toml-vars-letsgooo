# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**toml-vars-letsgooo** is a Go library that enables **internal variable substitution within TOML files** - allowing variables to be defined and referenced within the same TOML file using `{{variable}}` syntax, similar to template variables but purely within the TOML structure.

### Core Problem Being Solved
- TOML files currently don't support internal variable references
- Developers need Git-tracked configuration with variable substitution
- No existing Go libraries provide this functionality
- Viper is too heavy for simple internal TOML variable substitution needs

### Target API
```go
// External API - works like os.Getenv()
port := tomv.GetInt("server.port")
host := tomv.Get("database.host")
enabled := tomv.GetBool("features.login")

// With defaults
timeout := tomv.GetIntOr("api.timeout", 30)
```

### Variable Syntax Examples
```toml
# Internal TOML variables
[paths]
base_dir = "/home/user/.config"
app_name = "myapp"
config_path = "{{paths.base_dir}}/{{paths.app_name}}"

# Environment variable injection
[server]
port = "{{ENV.PORT:-3000}}"
host = "{{ENV.HOST:-localhost}}"

# Mixed usage
[database]
host = "{{ENV.DB_HOST:-localhost}}"
url = "postgres://{{database.host}}:5432/{{paths.app_name}}"
```

### Learning Objectives
This project serves as a **Go learning vehicle** with these educational goals:
- Master Go data structures (maps, interfaces, type assertions)
- Understand string processing and regular expressions
- Learn recursive algorithms for nested data processing
- Practice Go error handling patterns
- Experience professional package API design
- Build parser-adjacent logic (variable resolution)

## Architecture Overview

This is a **standalone Go library** with clean, simple architecture:

### Library Structure

```
toml-vars-letsgooo/
├── api.go              # Public API (Get, GetInt, GetBool, etc.)
├── cache.go            # File monitoring and smart caching
├── discovery.go        # Multi-file discovery and conflict resolution
├── parser.go           # Variable substitution engine
├── tomv_test.go       # Comprehensive test suite
├── go.mod              # Module definition
├── go.sum              # Dependencies
└── other/              # Documentation and specs
```

### Core Components

- **`api.go`** - External API providing `os.Getenv()`-like interface
  - `Get()`, `GetInt()`, `GetBool()`, `GetDuration()`, etc.
  - `GetOr()` variants with default values
  - `GetStringSlice()`, `GetIntSlice()` for arrays
  - `Exists()` for key existence checking

- **`discovery.go`** - Multi-file TOML discovery and conflict resolution
  - Recursive `*.toml` file discovery
  - Smart conflict detection across files
  - File namespacing: `filename.section.key` syntax
  - Cross-file variable resolution

- **`parser.go`** - Variable substitution engine
  - Internal variables: `{{section.key}}`
  - Environment injection: `{{ENV.VAR:-default}}`
  - Multi-pass resolution with circular dependency detection
  - Cross-file variable references

- **`cache.go`** - Performance optimization
  - File modification time tracking
  - Smart cache invalidation
  - Always-current values with minimal overhead

## Key Architectural Principles

### Standalone Library Standards
- **Zero external dependencies** - Only uses `github.com/BurntSushi/toml`
- **Single purpose focus** - TOML variable substitution only
- **Clean API design** - Works exactly like `os.Getenv()`
- **Self-contained** - No framework dependencies or configuration required

### Multi-File Philosophy
- **Automatic discovery** - Finds all `*.toml` files in project automatically
- **Smart conflict resolution** - No prefix needed unless conflicts exist
- **File namespacing** - `filename.section.key` syntax for explicit references
- **Cross-file variables** - `{{database.host}}` works across any file
- **Zero configuration** - Works out of the box

### Variable Resolution Standards
- **Internal variables** - `{{section.key}}` references within and across files
- **Environment injection** - `{{ENV.VAR:-default}}` with explicit defaults
- **Multi-pass resolution** - Handles forward references and dependencies
- **Circular dependency detection** - Clear error messages for invalid references
- **Performance optimized** - File monitoring with timestamp-based caching

## Development Commands

### Go Development Workflow
```bash
# Module is already initialized
# github.com/DeprecatedLuar/toml-vars-letsgooo

# Run all tests (20 test cases)
go test -v

# Run specific test pattern
go test -v -run TestMultiFile

# Build the library
go build

# Format code
go fmt ./...

# Run linting (if golangci-lint installed)
golangci-lint run
```

### Library Usage
```bash
# Install in another project
go get github.com/DeprecatedLuar/toml-vars-letsgooo
```

```go
// Import and use
import "github.com/DeprecatedLuar/toml-vars-letsgooo"

func main() {
    port := tomv.GetInt("server.port")
    host := tomv.Get("database.host")
    timeout := tomv.GetDurationOr("api.timeout", 30*time.Second)
}
```

## Implementation Status

### ✅ Phase 1: Core External API & File Discovery (Completed)
- ✅ External API: `tomv.Get()`, `tomv.GetInt()`, `tomv.GetBool()`
- ✅ Recursive file discovery system (`*.toml` search)
- ✅ Smart file monitoring with timestamp-based caching
- ✅ Environment variable injection: `{{ENV.VAR:-default}}`
- ✅ Complete type conversion utilities

### ✅ Phase 2: Internal Variable Resolution (Completed)
- ✅ `{{section.variable}}` reference resolution
- ✅ Multi-pass resolution for forward references
- ✅ Circular dependency detection with clear error messages
- ✅ Nested variable resolution (`{{a}}/{{b}}` where `b` references `{{c}}`)
- ✅ Comprehensive error handling

### ✅ Phase 3: Environment Variable Integration (Completed)
- ✅ `{{ENV.VAR:-default}}` injection using same substitution engine
- ✅ Mixed environment and internal variable scenarios
- ✅ Git-trackable environment variable usage
- ✅ Advanced type handling (`GetDuration`, slices)
- ✅ `GetOr` functions with defaults

### ✅ Phase 4: Multi-File Discovery & Conflict Resolution (Completed)
- ✅ Multiple TOML files with conflict handling
- ✅ File namespacing: `filename.section.key` syntax
- ✅ Smart conflict detection and resolution
- ✅ Cross-file variable resolution
- ✅ Professional API design and documentation
- ✅ **Production-ready standalone library**

## Library Features

### Variable Resolution Rules
- **Environment variables**: `{{ENV.VAR:-default}}` syntax with explicit defaults
- **Internal variables**: `{{section.variable}}` reference resolution within and across files
- **Multi-pass resolution**: Support forward references with automatic dependency resolution
- **Circular dependency detection**: Clear error messages for `A → B → A` cycles
- **File conflict resolution**: `filename.section.key` syntax for explicit file references
- **Smart conflict detection**: No prefix needed unless actual conflicts exist

### Multi-File Support
- **Automatic discovery**: Finds all `*.toml` files in project recursively
- **Cross-file variables**: `{{database.host}}` works from any file to any file
- **Conflict resolution**: Clear error messages with suggested explicit syntax
- **File namespacing**: `app.toml` becomes `app.section.key` for conflicts
- **Zero configuration**: Works immediately without setup

### Standalone Architecture
- **Single dependency**: Only `github.com/BurntSushi/toml` for parsing
- **Self-contained**: No framework dependencies or external configuration
- **Standard library**: Uses Go's built-in file system, regex, and string processing
- **Clean imports**: Direct import at root level `github.com/DeprecatedLuar/toml-vars-letsgooo`
- **Production ready**: Comprehensive test coverage (20+ test scenarios)

## Learning Focus Areas

### Go Concepts to Master Through This Project
- **Data structures**: Working with `map[string]interface{}` and type assertions
- **Regular expressions**: Pattern matching for `{{variable}}` detection
- **Recursion**: Processing nested TOML structures
- **Error handling**: Go's explicit error patterns and custom error types
- **Interface design**: Creating clean, reusable APIs
- **Testing patterns**: Unit tests for parser logic and edge cases

### Development Approach
1. **AI-generated code** → **Line-by-line manual review** → **Deep understanding**
2. **Iterative development**: Start simple, add complexity gradually
3. **Question everything**: Understand every Go concept encountered
4. **Real-world testing**: Use actual configuration files

## Testing Strategy

### Test Organization
- **Unit tests**: Place alongside source files (`*_test.go`)
- **Integration tests**: Use `other/testing/` for shared test utilities
- **Test data**: TOML files with various variable patterns in `other/testing/`

### Test Coverage Areas
- Basic variable substitution: `{{var}}` → `value`
- Cross-section references: `{{section.var}}` → `value`
- Nested variable references: `{{a}}/{{b}}` where `b` is `{{c}}`
- Forward references: `{{a}}` defined after `{{b}}` that uses it
- Circular dependency detection: `{{a}}` → `{{b}}` → `{{a}}`
- Error condition handling: undefined variables, malformed syntax
- Performance with large TOML files and many variables

## Usage Examples

### Basic Usage
```toml
# config.toml
[server]
port = 3000
host = "localhost"

[database]
url = "postgres://{{server.host}}:5432/myapp"
```

```go
port := tomv.GetInt("server.port")           // Returns: 3000
dbURL := tomv.Get("database.url")            // Returns: "postgres://localhost:5432/myapp"
```

### Multi-File with Conflicts
```toml
# app.toml
[server]
port = 3000

# api.toml  
[server]
port = 8080
```

```go
// This causes helpful error:
tomv.Get("server.port")  // Error: "found in multiple files, use app.server.port or api.server.port"

// Explicit syntax works:
appPort := tomv.GetInt("app.server.port")    // Returns: 3000
apiPort := tomv.GetInt("api.server.port")    // Returns: 8080
```

### Environment Variables
```toml
[server]
port = "{{ENV.PORT:-3000}}"
host = "{{ENV.HOST:-localhost}}"
```

## Library Distribution

### Installation
```bash
go get github.com/DeprecatedLuar/toml-vars-letsgooo
```

### Import
```go
import "github.com/DeprecatedLuar/toml-vars-letsgooo"
```

## Competitive Positioning

**vs. Viper**: 10x lighter, focused solely on internal TOML variable substitution
**vs. External templates**: No build step, no external tools required  
**vs. Environment variables**: Everything in git-trackable TOML files with env override capability
**vs. Manual string replacement**: Type-safe, structured, maintainable with automatic conflict detection

## Technical Achievements

### Production Ready Features
- ✅ Clean API that works like `os.Getenv()`
- ✅ Multi-file TOML support with smart conflict resolution
- ✅ Cross-file variable references
- ✅ Environment variable injection with defaults
- ✅ Comprehensive error messages for all edge cases
- ✅ Zero configuration required
- ✅ Complete type system (string, int, bool, duration, slices)
- ✅ 20+ comprehensive test scenarios

### Code Quality
- ✅ Single external dependency (`github.com/BurntSushi/toml`)
- ✅ Professional Go package structure
- ✅ Clean, readable, maintainable codebase
- ✅ Follows Go conventions and best practices
- ✅ Ready for open-source distribution

This library successfully fills a gap in the Go ecosystem by providing simple, efficient internal TOML variable substitution with multi-file support and smart conflict resolution.