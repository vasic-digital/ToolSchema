# digital.vasic.toolschema

A generic, reusable Go module for tool schema definition, validation, and execution. Provides a unified interface for AI agent tool systems with safety and validation as first-class concerns.

## Features

- **Unified Tool Interface**: Consistent `ToolHandler` interface for all tools
- **Built-in Safety**: Path validation, command argument validation, shell injection prevention
- **14+ Built-in Tool Handlers**: Ready-to-use tools for common development tasks
- **Thread-Safe Registry**: Concurrent tool registration and execution
- **Zero Runtime Dependencies**: Pure Go standard library
- **Comprehensive Validation**: Validate paths, symbols, git refs, command arguments

## Installation

```bash
go get digital.vasic.toolschema
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "digital.vasic.toolschema"
)

func main() {
    // Create tool registry
    registry := tools.NewToolRegistry()
    
    // Register built-in handlers
    registry.Register(&tools.ReadFileHandler{})
    registry.Register(&tools.GitHandler{})
    registry.Register(&tools.TestHandler{})
    registry.Register(&tools.LintHandler{})
    
    // Execute a tool
    ctx := context.Background()
    result, err := registry.Execute(ctx, "read_file", map[string]interface{}{
        "file_path": "README.md",
        "offset": 0,
        "limit": 100,
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    if result.Success {
        fmt.Println("Output:", result.Output)
    } else {
        fmt.Println("Error:", result.Error)
    }
}
```

## Built-in Tool Handlers

| Tool | Description | Parameters |
|------|-------------|------------|
| `read_file` | Read file contents with line range support | `file_path` (required), `offset`, `limit` |
| `Git` | Git version control operations | `operation`, `arguments`, `working_dir`, `description` |
| `Test` | Go test execution | `test_path`, `test_type`, `coverage`, `verbose`, `filter`, `timeout` |
| `Lint` | Code linting | `path`, `linter`, `auto_fix`, `config`, `description` |
| `Diff` | Git diff | `file_path`, `mode`, `compare_with`, `context_lines` |
| `TreeView` | Directory tree | `path`, `max_depth`, `show_hidden`, `ignore_patterns` |
| `FileInfo` | File metadata | `file_path`, `include_stats`, `include_git` |
| `Symbols` | Extract code symbols | `file_path`, `recursive` |
| `References` | Find symbol references | `symbol`, `file_path`, `include_declaration` |
| `Definition` | Find symbol definition | `symbol`, `file_path`, `line` |
| `PR` | Pull request management | `action`, `title`, `body`, `base_branch`, `pr_number`, `labels` |
| `Issue` | Issue management | `action`, `title`, `body`, `issue_number`, `labels`, `assignees` |
| `Workflow` | CI/CD workflow management | `action`, `workflow_id`, `branch`, `run_id` |

## Safety Features

### Path Validation
```go
// Prevents path traversal and shell injection
if !tools.ValidatePath(path) {
    return fmt.Errorf("invalid path: %s", path)
}
```

### Command Argument Validation
```go
// Ensures arguments don't contain shell metacharacters
if !tools.ValidateCommandArg(arg) {
    return fmt.Errorf("unsafe argument: %s", arg)
}
```

### Symbol Validation
```go
// Validates symbol names for safe grep patterns
if !tools.ValidateSymbol(symbol) {
    return fmt.Errorf("invalid symbol: %s", symbol)
}
```

### Git Reference Validation
```go
// Validates git branch/tag names
if !tools.ValidateGitRef(ref) {
    return fmt.Errorf("invalid git reference: %s", ref)
}
```

## Creating Custom Tool Handlers

```go
type MyToolHandler struct{}

func (h *MyToolHandler) Name() string { return "my_tool" }

func (h *MyToolHandler) ValidateArgs(args map[string]interface{}) error {
    // Validate arguments using built-in validation
    path, ok := args["path"].(string)
    if !ok || !tools.ValidatePath(path) {
        return fmt.Errorf("invalid path")
    }
    return nil
}

func (h *MyToolHandler) GenerateDefaultArgs(context string) map[string]interface{} {
    return map[string]interface{}{
        "path": ".",
        "description": "My custom tool",
    }
}

func (h *MyToolHandler) Execute(ctx context.Context, args map[string]interface{}) (tools.ToolResult, error) {
    // Execute your tool logic
    return tools.ToolResult{
        Success: true,
        Output: "Tool executed successfully",
    }, nil
}

// Register custom handler
registry.Register(&MyToolHandler{})
```

## Thread Safety

The `ToolRegistry` is thread-safe for concurrent registration and execution:

```go
// Safe for concurrent use
go func() {
    registry.Register(&tools.ReadFileHandler{})
}()

go func() {
    result, _ := registry.Execute(ctx, "read_file", args)
    // ...
}()
```

## Testing

```bash
# Run all tests (resource-capped per universal mandate)
GOMAXPROCS=2 nice -n 19 ionice -c 3 go test -count=1 -race -p 1 ./...

# Run unit tests only
go test ./... -short

# Run integration tests
go test -tags=integration ./...

# Run specific tool tests
go test -v -run TestGitHandler ./...

# Run the round-285 anti-bluff Challenge runner against the real
# in-process registry + handler + validation surfaces. Exits 0 on
# success with captured stdout for every schema, handler, validation
# gate, and 5-locale UX line. See docs/test-coverage.md for the
# symbol → test / Challenge ledger.
go run ./challenges/runner -all

# Paired-mutation evidence (CONST-035). Plants a known schema
# violation in a scratch module, asserts the runner-style invariant
# check surfaces it. Exits 99 (sentinel) on correctly detected
# mutation; exits 1 if the mutation slipped through (Challenge is
# a bluff).
bash challenges/toolschema_describe_challenge.sh --mutate
```

## Anti-Bluff Guarantees (round-285)

This module ships with the round-285 anti-bluff Challenge surface
mandated by the parent constitution submodule (Article XI §11.9 +
CONST-035 + CONST-048). The Challenge runner exercises the **real**,
**in-process** schema registry, handler registry, and validation
gates — there are no mocks beyond the unit-test layer (CONST-050(A))
and no metadata-only PASS results (every Challenge assertion is
backed by a printed line on stdout that consumers can grep for).

Specifically, the runner proves:

1. **Registry completeness.** `GetAllToolNames()` is iterated; every
   returned name resolves via `GetToolSchema()`; no entry is skipped.
   A registry mutation that drops an entry surfaces as a coverage
   exit-2 with the offending name printed.
2. **Schema invariants.** For every schema, every `RequiredFields`
   entry MUST appear in `Parameters`; every `Parameters` entry
   marked `Required: true` MUST appear in `RequiredFields`; every
   parameter MUST declare a non-empty `Type`. Mutation that breaks
   any of these surfaces as schema exit-3.
3. **OpenAI definition round-trip.** Every schema is round-tripped
   through `GenerateOpenAIToolDefinition()`; the inner
   `function.name` field MUST match `schema.Name`. Mutation in the
   generator surfaces as schema exit-3.
4. **ValidateToolArgs error + success symmetry.** The runner asserts
   `ValidateToolArgs("Read", {})` returns a non-nil error (missing
   required field MUST be caught) AND `ValidateToolArgs("Read",
   {file_path: "/tmp/x"})` returns nil (valid args MUST be
   accepted). Either-direction regression surfaces.
5. **Default handler registry.** `GetDefaultToolRegistry()` is asked
   for each of its 13 built-in handler names; each handler's
   `Name()` MUST be non-empty and `GenerateDefaultArgs("")` MUST
   return a non-nil map. Mutation that drops a handler surfaces.
6. **Validation gates (4× symmetric).** `ValidatePath`,
   `ValidateSymbol`, `ValidateGitRef`, `ValidateCommandArg` are each
   tested with a known-good input (MUST accept) AND a known-bad
   input (MUST reject). Mutation in either direction surfaces as
   validation exit-5.
7. **Real search across registry.** `SearchTools(Query: "file",
   MaxResults: 5)` is invoked; result count MUST be ≥ 1; the top
   result's `Tool` pointer MUST be non-nil. A hardcoded search
   would not honour `MaxResults` cap and would fail this check.
8. **5-locale CONST-046 UX.** Five locale lines (`en`, `sr`, `ja`,
   `es`, `de`) are printed, each containing the canonical
   `toolschema:` token. Missing locale → exit-4.

### Paired-mutation evidence

`challenges/toolschema_describe_challenge.sh --mutate` constructs a
self-contained scratch Go module in `$TMPDIR`, vendors a mutated
copy of the schema-invariant assertion that declares
`RequiredFields: ["file_path"]` with an empty `Parameters` map, and
runs the assertion. The mutation MUST be detected (exit 99). A run
that exits 0 means the Challenge itself is a bluff and the wrapper
fails the Challenge with exit 1. This is the round-285 paired
mutation gate (CONST-035 mutation-bluff prevention).

### Exit-code map

| Code | Meaning                                                          |
|------|------------------------------------------------------------------|
| 0    | Every assertion passed; every locale line printed.               |
| 1    | Usage / flag error (or mutation slipped through in `--mutate`).  |
| 2    | Coverage gap (registry empty, unknown name, missing handler).    |
| 3    | Schema-invariant violation (RequiredField↔Parameters drift).     |
| 4    | Locale UX line missing or canonical token absent.                |
| 5    | Validation gate regression (accept-bad or reject-good).          |
| 99   | (mutation mode) Mutation correctly surfaced — sentinel PASS.     |

## License

MIT