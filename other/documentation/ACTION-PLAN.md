# tomv Implementation Plan

## Strategic Implementation Overview

This document outlines the development plan for the tomv library following value-first development principles and the established framework patterns.

## File Structure

```
src/project/tomv/
├── api.go           # Public API (Get, GetInt, GetOr functions)
├── discovery.go     # File finding + conflict resolution
├── parser.go        # TOML parsing + ALL variable substitution
├── cache.go         # File monitoring + smart caching
└── tomv_test.go    # Comprehensive test suite
```

**Design Rationale:**
- **100% standalone library** - zero framework dependencies
- Each file has single responsibility under 250-line limit
- Consolidated architecture (5 files instead of 7) reduces complexity
- `parser.go` handles both `{{section.key}}` and `{{ENV.VAR:-default}}` patterns
- `discovery.go` includes conflict resolution (related functionality)

## Phase-Based Development Strategy

### Phase 1: Core Foundation (Days 1-2)
**Deliverable:** Working `tomv.Get()` with single TOML file (no substitution)

**Success Criteria:**
```go
// This works at end of Phase 1:
port := tomv.GetInt("server.port")     // From config.toml, basic lookup only
host := tomv.Get("database.host")      // String retrieval working
debug := tomv.GetBool("app.debug")     // Type conversion working
```

**Implementation Order:**
1. **`api.go`** - Complete public API skeleton
   - All core functions: `Get()`, `GetInt()`, `GetBool()`, `GetOr()` variants
   - Type conversion integration
   - Standard error message format implementation

2. **`discovery.go`** - Single file discovery only
   - Project root detection via `go.mod` or `.git/`
   - Find first `*.toml` file in project
   - Basic file loading and validation

3. **`cache.go`** - Simple file monitoring
   - File modification time tracking
   - Cache invalidation on file changes
   - Always-current value guarantee (< 500ns timestamp check)

**Error Message Standard:**
```
Error: Variable "database.host" not found

Searched in:
- config.toml

Available variables:
- database.port
- database.name
```

**Testing:** Basic TOML loading and type conversion without any variable substitution

### Phase 2: Internal Variable Substitution (Days 3-5)
**Deliverable:** `{{section.key}}` substitution within single file

**Success Criteria:**
```toml
# This works at end of Phase 2:
[database]
host = "localhost"
port = 5432
url = "postgres://{{database.host}}:{{database.port}}/myapp"

[paths]
base = "/app"
uploads = "{{paths.base}}/uploads"
```

```go
// This API call works:
dbURL := tomv.Get("database.url")  // Returns: "postgres://localhost:5432/myapp"
uploadPath := tomv.Get("paths.uploads")  // Returns: "/app/uploads"
```

**Implementation:**
1. **`parser.go`** - Core substitution engine
   - `{{section.key}}` pattern recognition and parsing
   - Multi-pass dependency resolution algorithm (specified below)
   - Circular dependency detection with clear error messages
   - Cross-section reference support within same file

**Multi-Pass Resolution Algorithm:**
```
Input: TOML map with unresolved {{references}}
Output: TOML map with all variables substituted

Algorithm:
1. Parse all {{variable}} references in values
2. Build dependency graph: variable → referenced variables
3. Detect circular dependencies (error if found)
4. Resolve in topological order:
   - Pass 1: Resolve variables with no dependencies
   - Pass 2: Resolve variables whose dependencies are now resolved
   - Repeat until all resolved or no progress made
5. Error if unresolvable variables remain
```

**Edge Case Handling:**
- **Circular dependency:** `a = "{{b}}" + b = "{{a}}"` → Error: "Circular dependency detected: a → b → a"
- **Undefined reference:** `a = "{{missing}}"` → Error: "Variable 'missing' referenced in 'a' but not found"
- **Self-reference:** `a = "{{a}}/suffix"` → Error: "Variable 'a' cannot reference itself"

**Testing:** Complex nested variable scenarios and all edge cases

### Phase 3: Environment Variable Integration (Days 6-7)
**Deliverable:** `{{ENV.VAR:-default}}` injection using same substitution engine

**Success Criteria:**
```toml
# This works at end of Phase 3:
[server]
port = "{{ENV.PORT:-3000}}"
host = "{{ENV.HOST:-localhost}}"

[database]
url = "postgres://{{ENV.DB_HOST:-localhost}}:{{database.port}}/myapp"
```

```go
// With PORT=8080 in environment:
port := tomv.GetInt("server.port")  // Returns: 8080
// Without PORT in environment:
port := tomv.GetInt("server.port")  // Returns: 3000 (default)
```

**Implementation:**
1. **Enhanced `parser.go`** - Extend substitution engine
   - Add `{{ENV.VAR:-default}}` pattern recognition
   - Environment variable lookup with `os.Getenv()`
   - Default value parsing and application
   - Integration with existing multi-pass resolution

**Key Features:**
- Reuses existing `{{}}` pattern matching infrastructure
- Environment and internal variables can reference each other
- Clear error messages for malformed ENV syntax
- Git-trackable environment variable usage

**Testing:** Mixed environment and internal variable scenarios

### Phase 4: Multi-File Discovery & Conflict Resolution (Days 8-9)
**Deliverable:** Multiple TOML files with conflict handling

**Success Criteria:**
```go
// This works at end of Phase 4:
appPort := tomv.GetInt("app.server.port")          // From config/app.toml
dbPort := tomv.GetInt("database.server.port")      // From config/database.toml

// Automatic conflict detection:
port := tomv.GetInt("server.port")  // Error if exists in multiple files
```

**Conflict Resolution Error Format:**
```
Error: Variable "server.port" found in multiple files:
- config/app.toml
- config/database.toml

Use explicit syntax:
- tomv.Get("app.server.port")
- tomv.Get("database.server.port")
```

**Implementation:**
1. **Enhanced `discovery.go`** - Multi-file discovery + conflict resolution
   - Recursive directory traversal for `*.toml` files
   - Duplicate key detection across files
   - `filename.section.key` syntax implementation
   - File path normalization: `config/app.toml` → `app`

**Key Features:**
- Automatic discovery of all TOML files in project
- Integrated conflict detection and resolution
- Clear error messages with explicit syntax suggestions
- Performance-optimized file processing

**Testing:** Multi-file scenarios with various conflict situations

### Phase 5: Polish & Performance (Days 10-11)
**Deliverable:** Production-ready library

**Implementation:**
1. **Advanced Types** - Extended type conversion
   - `GetDuration()` with Go duration format parsing
   - `GetStringSlice()` with comma-separated value parsing
   - `GetIntSlice()` for integer arrays
   - `Exists()` for key existence checking

2. **Performance Optimization**
   - LRU cache for resolved values
   - Optimized file monitoring with minimal overhead
   - Memory usage optimization
   - Concurrent file processing where applicable

3. **Enhanced Error Handling**
   - Comprehensive error messages with context
   - Helpful suggestions for common mistakes
   - Clear indication of available variables
   - File location information in errors

4. **Full Test Coverage**
   - Unit tests for all functions
   - Integration tests for complex scenarios
   - Performance benchmarks
   - Edge case coverage

## Development Workflow

### 1. Dependency Management
- Initialize Go module: `go mod init github.com/username/tomv`
- Add required dependency: `github.com/BurntSushi/toml`
- Maintain minimal external dependencies

### 2. Standalone Library Strategy
- **100% standalone implementation** - zero framework dependencies
- Standard library only (except `github.com/BurntSushi/toml`)
- Self-contained error handling and logging
- Framework integration removed from scope

### 3. Testing Strategy
- Create test TOML files in `other/testing/`
- Test each phase incrementally
- Include both positive and negative test cases
- Performance testing for large configuration sets

### 4. Documentation Standards
- Update `CLAUDE.md` with implementation progress
- Maintain inline documentation for complex algorithms
- Create usage examples for each major feature
- Document performance characteristics

## Quality Standards

### Code Quality
- Follow Go conventions and best practices
- Maintain single responsibility per file
- Use clear, descriptive function and variable names
- Implement comprehensive error handling

### Performance Targets (Realistic)
- File discovery: < 10ms for typical projects (< 100 TOML files)
- Variable resolution: < 1ms for cached lookups
- File monitoring overhead: < 500ns per access (realistic timestamp check)
- Memory usage: < 5MB for typical configuration sets

### Error Handling
- Clear, actionable error messages
- Context information for debugging
- Graceful handling of edge cases
- No silent failures or unexpected behavior

## Success Metrics

### Phase Completion Criteria
Each phase must meet all success criteria before proceeding to the next phase. This ensures a solid foundation and prevents technical debt accumulation.

### Final Success Metrics
- **Zero learning curve:** Works exactly like `os.Getenv()`
- **Zero configuration required:** Import and use immediately
- **Always current values:** Smart file monitoring without user intervention
- **Clear error messages:** Configuration problems are obvious and fixable
- **Environment variable compatibility:** Standard override behavior

## Risk Mitigation

### Technical Risks
- **File system performance:** Implement efficient caching and monitoring
- **Memory usage:** Optimize data structures and implement cleanup
- **Cross-platform compatibility:** Test on Windows, macOS, and Linux
- **Concurrent access:** Handle multiple goroutines safely

### Development Risks
- **Scope creep:** Stick to phase boundaries and success criteria
- **Over-engineering:** Maintain focus on core value proposition
- **Framework coupling:** Ensure library works standalone
- **Performance regressions:** Regular benchmarking throughout development

This implementation plan balances thorough feature development with incremental value delivery, ensuring each phase produces a working, testable improvement to the library.