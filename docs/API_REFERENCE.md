# ToolSchema API Reference

## Overview

`digital.vasic.toolschema` provides tool schema definition, validation, and execution for AI agent tool systems.

## Core Interface

### ToolHandler

```go
type ToolHandler interface {
    HandleToolCall(ctx context.Context, call ToolCall) (*ToolResult, error)
    GetSchema() *ToolSchema
    Validate(params map[string]interface{}) error
}
```

### ToolSchema

Defines the schema for a tool that AI agents can invoke:

```go
type ToolSchema struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  map[string]Parameter   `json:"parameters"`
    Required    []string               `json:"required"`
}
```

### Parameter

```go
type Parameter struct {
    Type        string   `json:"type"`        // string, number, boolean, array, object
    Description string   `json:"description"`
    Enum        []string `json:"enum,omitempty"`
    Default     any      `json:"default,omitempty"`
}
```

## Validation

All parameters use **snake_case** naming convention. The validator checks:
- Required parameters are present
- Parameter types match schema definitions
- Enum values are within allowed set
- Nested object schemas are validated recursively

## SQL Definitions

See `docs/sql-definitions.md` for database schema used by the tool registry.
