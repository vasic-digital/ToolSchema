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
# Run all tests
go test ./... -count=1 -race

# Run unit tests only
go test ./... -short

# Run integration tests
go test -tags=integration ./...

# Run specific tool tests
go test -v -run TestGitHandler ./...
```

## License

MIT