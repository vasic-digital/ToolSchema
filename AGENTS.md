# AGENTS.md - ToolSchema Module

## Module Overview

`digital.vasic.toolschema` is a generic, reusable Go module providing tool schema definition, validation, and execution for AI agent tool systems. It provides a unified interface for defining tool handlers with parameter validation, safe command execution, and result formatting. The module has zero external runtime dependencies beyond Go standard library.

**Module path**: `digital.vasic.toolschema`
**Go version**: 1.24+
**Dependencies**: `github.com/stretchr/testify` (test only)

## Package Responsibilities

| Package | Path | Responsibility |
|---------|------|----------------|
| `tools` | `./` | Core types: `ToolHandler` interface, `ToolRegistry`, validation functions (`ValidatePath`, `ValidateSymbol`, `ValidateGitRef`, `ValidateCommandArg`, `SanitizePath`), 14+ built-in tool handlers (ReadFile, Git, Test, Lint, Diff, TreeView, FileInfo, Symbols, References, Definition, PR, Issue, Workflow). This is the only package with no internal dependencies. |

## Dependency Graph

```
tools (self-contained)
```

The module is a single package with no internal dependencies. All validation functions are pure and stateless.

## Key Files

| File | Purpose |
|------|---------|
| `handler.go` | ToolHandler interface, ToolRegistry, all tool handler implementations |
| `schema.go` | Schema definition and validation utilities |
| `validation.go` | Validation functions for paths, symbols, git refs, command arguments |
| `search.go` | Search utilities (if any) |
| `handler_test.go` | Handler package unit tests |
| `schema_test.go` | Schema package unit tests |
| `search_test.go` | Search package unit tests |
| `go.mod` | Module definition and dependencies |
| `CLAUDE.md` | AI coding assistant instructions |
| `README.md` | User-facing documentation with quick start |

## Agent Coordination Guide

### Division of Work

When multiple agents work on this module simultaneously, divide work by tool handler boundaries:

1. **Core Agent** -- Owns `ToolHandler` interface, `ToolRegistry`, validation functions. Changes to core types affect all tool handlers. Must coordinate with all other agents before modifying the `ToolHandler` interface or `ToolResult` struct.
2. **Tool Handler Agents** -- Each agent can own one or more tool handlers (Git, Test, Lint, etc.). Changes to a specific tool handler only affect that handler.
3. **Validation Agent** -- Owns validation functions. Changes to validation logic affect all tool handlers that use those functions.

### Coordination Rules

- **ToolHandler interface changes** require all agents to update. The interface is the shared contract.
- **ToolResult struct changes** require all agents to update. This is the shared output format.
- **Validation function changes** affect all tool handlers that use them. Must be coordinated with tool handler owners.
- **New tool handlers** can be added independently without coordination, as long as they implement the existing interface.
- **Test isolation**: Each tool handler should have its own test cases in `handler_test.go`.

### Safe Parallel Changes

These changes can be made simultaneously without coordination:
- Adding a new tool handler (implementing existing ToolHandler interface)
- Adding new tests for existing tool handlers
- Updating documentation
- Adding new validation helper functions (if they don't break existing signatures)

### Changes Requiring Coordination

- Modifying the `ToolHandler` interface methods
- Changing `ToolResult` struct fields
- Modifying validation function signatures or behavior
- Changing `ToolRegistry` thread-safety mechanisms

## Build and Test Commands

```bash
# Build all packages
go build ./...

# Run all tests with race detection
go test ./... -count=1 -race

# Run unit tests only (short mode)
go test ./... -short

# Run integration tests
go test -tags=integration ./...

# Run a specific test
go test -v -run TestReadFileHandler ./...

# Format code
gofmt -w .

# Vet code
go vet ./...
```

## Commit Conventions

Follow Conventional Commits with tool scope:

```
feat(tools): add new validation function for URLs
feat(git): add support for git stash operations
feat(test): add benchmark test support
fix(validation): prevent path traversal in ValidatePath
test(readfile): add edge case tests for empty files
docs(toolschema): update API reference
refactor(registry): improve thread safety with RWMutex
```

## Thread Safety Notes

- `ToolRegistry` is thread-safe using `sync.RWMutex`. Registration and lookup are protected.
- Tool handlers are stateless and safe for concurrent execution.
- Validation functions are pure functions with no shared state, safe for concurrent invocation.
- Command execution uses `exec.CommandContext` with validated arguments to prevent shell injection.