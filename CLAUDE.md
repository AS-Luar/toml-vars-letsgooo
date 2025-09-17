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
port := tmvar.GetInt("server.port")
host := tmvar.Get("database.host")
enabled := tmvar.GetBool("features.login")

// With defaults
timeout := tmvar.GetIntOr("api.timeout", 30)
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

This is a **framework-based project** with strict architectural separation:

### Primary Architecture (`src/`)

- **`src/modules/`** - Reusable framework components (cross-project utilities)
  - `paths/` - Central path management with TOML-based configuration
  - `logging/` - Emoji-based logging system with environment awareness
  - `errors/` - Standardized error handling
  - `server/` - Server infrastructure components

- **`src/project/`** - Application-specific business logic (tmvar library implementation)
  - External API (`tmvar.Get()`, `tmvar.GetInt()`, etc.)
  - File discovery system (recursive *.toml search)
  - Environment variable injection (`{{ENV.VAR:-default}}`)
  - Internal variable resolution (`{{section.key}}`)
  - Smart file monitoring and caching
  - Type conversion utilities

- **`src/settings/`** - Centralized configuration management
  - `settings.toml` - Main configuration file
  - `.env` - Environment variables

### Supporting Structure (`other/`)

- **`other/documentation/`** - Project specifications and design docs
- **`other/testing/`** - Testing utilities and infrastructure
- **`other/report/`** - Development workflow tracking
  - `active/` - Current ongoing tasks
  - `complete/` - Historical completed work

## Key Architectural Principles

### Framework Module Standards
- **Zero project dependencies** - Modules must be reusable across projects
- **250-line file limit** - Enforced for maintainability
- **Single responsibility** - Each module has one clear purpose
- **Dependency injection patterns** - Clean interfaces between components
- **Standardized responses** - `{success, data, error}` format

### Path Management Philosophy
- **Zero hardcoded paths** - All paths via centralized TOML configuration
- **Variable substitution** - `{base.variable}` and `{runtime_variable}` patterns
- **Project root discovery** - Automatic detection via `go.mod`, `.git/`, etc.
- **Performance caching** - LRU cache with 1000 entry maximum

### Logging Standards
- **Single logger per project** - No multiple logging libraries
- **Emoji-based categorization** - Visual efficiency (🔵 info, ❌ error, ✅ success, etc.)
- **Environment awareness** - All logs in development, silent in production
- **Component tagging** - `[ComponentName]` prefix for all log entries

## Development Commands

### Go Development Workflow
```bash
# Initialize Go module (if not already done)
go mod init github.com/yourusername/toml-vars-letsgooo

# Add required dependencies
go get github.com/BurntSushi/toml

# Run tests
go test ./...

# Run specific test
go test ./src/project -v

# Build the library
go build ./...

# Run linting (install golangci-lint first)
golangci-lint run

# Format code
go fmt ./...
```

### Framework Validation Commands
```bash
# Validate framework structure (custom script needed)
# Should verify all .purpose.md files exist and framework compliance

# Check path configuration
# Validate paths.toml against actual directory structure

# Verify logging integration
# Ensure all modules use centralized logging system
```

## Implementation Strategy

### Phase 1: Core External API & File Discovery (Week 1)
1. Implement external API: `tmvar.Get()`, `tmvar.GetInt()`, `tmvar.GetBool()`
2. Build recursive file discovery system (*.toml search)
3. Add smart file monitoring with timestamp-based caching
4. Basic environment variable injection: `{{ENV.VAR:-default}}`
5. Type conversion utilities

### Phase 2: Internal Variable Resolution (Week 2)
1. Implement `{{section.variable}}` reference resolution
2. Multi-pass resolution for forward references
3. Circular dependency detection and clear error messages
4. Nested variable resolution (`{{a}}/{{b}}` where `b` references `{{c}}`)
5. Comprehensive error handling

### Phase 3: Advanced Features & Conflict Resolution (Week 3)
1. File conflict resolution (`filename.section.key` syntax)
2. Advanced type handling (`GetDuration`, slices)
3. `GetOr` functions with defaults
4. Performance optimization for large projects
5. Comprehensive test suite

### Phase 4: Framework Integration & Polish (Week 4)
1. Integrate with modules/paths for configuration file handling
2. Use modules/logging for debug output
3. Apply modules/errors for standardized error handling
4. Professional API design and documentation
5. Package structure for potential open-source release

## Critical Implementation Notes

### Variable Resolution Rules
- **Environment variables**: `{{ENV.VAR:-default}}` syntax with explicit defaults
- **Internal variables**: `{{section.variable}}` reference resolution
- **Multi-pass resolution**: Support forward references with dependency resolution
- **Circular dependency detection**: Clear error messages for `A → B → A` cycles
- **File conflict resolution**: `filename.section.key` syntax for explicit file references

### Technical Approach
- **External API first**: `tmvar.Get()` functions that work like `os.Getenv()`
- **File discovery**: Recursive search for *.toml files in project directory
- **Smart caching**: Timestamp-based file monitoring for always-current values
- **Two-phase resolution**: Environment injection → Internal variable resolution
- **BurntSushi/toml foundation**: Use proven TOML parser, focus on library logic

### Framework Compliance Requirements
- All source files in `src/project/` for business logic
- Use `modules/logging` for all debug output (no direct fmt.Println)
- Leverage `modules/paths` for any file system operations
- Follow standardized error responses from `modules/errors`
- Maintain 250-line file limit for all modules

### Dependencies Strategy
- **Core TOML parsing**: `github.com/BurntSushi/toml`
- **No external logging libraries** - use framework modules/logging
- **No external path libraries** - use framework modules/paths
- **Minimal dependencies** - prefer standard library when possible

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

## File Structure Guidelines

### Purpose File System
Every directory contains `.purpose.md` explaining its intent, scope, and dependencies. Always read these files to understand architectural context before making changes.

### Documentation Integration
- Main project specification: `other/documentation/TOML-VARIABLE-SYSTEM.md`
- Module specifications: `src/modules/{module}/doc.md`
- Purpose definitions: `{directory}/.purpose.md`

### Development Workflow
- Use `other/report/active/` for tracking ongoing implementation tasks
- Move completed analysis to `other/report/complete/`
- Update documentation as features are implemented

## Competitive Positioning

**vs. Viper**: 10x lighter, focused solely on internal TOML variable substitution
**vs. External templates**: No build step, no external tools required
**vs. Environment variables**: Everything in one git-trackable TOML file
**vs. Manual string replacement**: Type-safe, structured, maintainable

## Success Metrics

### Technical Success
- Clean API that works with any TOML structure
- Comprehensive variable substitution without external dependencies
- Clear error messages for all edge cases
- Zero breaking changes when upgrading from simple to flexible resolution

### Learning Success
- Deep understanding of Go's type system and interface patterns
- Mastery of string processing and regular expressions
- Experience with recursive data structure processing
- Professional Go package development skills

This project represents both a legitimate gap-filling Go library and an intensive Go learning experience, built with professional framework architecture for long-term maintainability and potential open-source distribution.