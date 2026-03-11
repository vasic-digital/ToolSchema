# CLAUDE.md - ToolSchema Module

## Overview

`digital.vasic.toolschema` is a generic, reusable Go module for tool schema definition, validation, and execution. It provides a unified interface for defining tool handlers with parameter validation, safe command execution, and result formatting. The module is designed for AI agent tool systems where safety and validation are critical.

**Module**: `digital.vasic.toolschema` (Go 1.24+)

## Build & Test

```bash
go build ./...
go test ./... -count=1 -race
go test ./... -short              # Unit tests only
go test -tags=integration ./...   # Integration tests
```

## Code Style

- Standard Go conventions, `gofmt` formatting
- Imports grouped: stdlib, third-party, internal (blank line separated)
- Line length <= 100 chars
- Naming: `camelCase` private, `PascalCase` exported, acronyms all-caps
- Errors: always check, wrap with `fmt.Errorf("...: %w", err)`
- Tests: table-driven, `testify`, naming `Test<Struct>_<Method>_<Scenario>`

## Package Structure

| Package | Purpose |
|---------|---------|
| `tools` (root) | Core types: ToolHandler interface, ToolRegistry, validation functions, built-in tool handlers (Git, Test, Lint, etc.) |
| `tools/schema` | Schema definition and validation utilities (if extracted) |

## Key Interfaces

- `ToolHandler`: Interface for tool execution with `Name()`, `Execute()`, `ValidateArgs()`, `GenerateDefaultArgs()`
- `ToolRegistry`: Registry for tool handlers with thread-safe registration and lookup
- `ToolResult`: Standardized result structure with success flag, output, error, and data fields

## Safety & Validation

- **Path validation**: Prevents path traversal and shell injection
- **Argument validation**: Validates command arguments for shell safety
- **Symbol validation**: Ensures symbol names are safe for grep patterns
- **Git reference validation**: Validates git branch/tag names
- **Built-in tool handlers**: 14+ safe tool implementations (ReadFile, Git, Test, Lint, Diff, TreeView, FileInfo, Symbols, References, Definition, PR, Issue, Workflow)

## Built-in Tool Handlers

1. **ReadFile**: Read file contents with line range support
2. **Git**: Git version control operations with safe argument validation
3. **Test**: Go test execution with coverage and filtering
4. **Lint**: Code linting with auto-detection and auto-fix
5. **Diff**: Git diff with multiple modes (working, staged, commit, branch)
6. **TreeView**: Directory tree display with depth control
7. **FileInfo**: File metadata with stats and git history
8. **Symbols**: Extract code symbols (functions, types, constants)
9. **References**: Find symbol references in codebase
10. **Definition**: Find symbol definitions
11. **PR**: GitHub/GitLab pull request management via gh CLI
12. **Issue**: Issue management via gh CLI
13. **Workflow**: CI/CD workflow management via gh CLI

## Usage Example

```go
import "digital.vasic.toolschema"

registry := tools.NewToolRegistry()
registry.Register(&tools.ReadFileHandler{})
registry.Register(&tools.GitHandler{})

result, err := registry.Execute(ctx, "read_file", map[string]interface{}{
    "file_path": "README.md",
})
```

## Dependencies

Runtime: None (pure Go standard library)
Test: `github.com/stretchr/testify`

## Thread Safety

- `ToolRegistry` uses `sync.RWMutex` for thread-safe registration and lookup
- Tool handlers are stateless and safe for concurrent execution
- Validation functions are pure functions with no shared state