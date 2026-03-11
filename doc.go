// Package tools provides the tool schema registry for HelixAgent.
//
// This package manages the 21 tools available to HelixAgent and ensures
// consistent tool schema definitions for LLM function calling.
//
// # Available Tools (21)
//
// File Operations:
//   - Read: Read file contents
//   - Write: Write content to file
//   - Edit: Edit file with search/replace
//   - Glob: Find files by pattern
//   - Grep: Search file contents
//
// Code Operations:
//   - Bash: Execute shell commands
//   - Git: Git operations
//
// Web Operations:
//   - WebFetch: Fetch and process web content
//   - WebSearch: Search the web
//
// Communication:
//   - AskUserQuestion: Interactive user prompts
//
// Task Management:
//   - Task: Launch background tasks
//   - TaskOutput: Get task output
//   - TaskCreate/TaskUpdate/TaskGet/TaskList: Task tracking
//
// Session Management:
//   - Skill: Execute skills
//   - NotebookEdit: Edit Jupyter notebooks
//   - KillShell: Terminate background shells
//
// Planning:
//   - EnterPlanMode: Enter planning mode
//   - ExitPlanMode: Exit planning mode
//
// # Tool Schema
//
// All tools follow a consistent schema format:
//
//	type ToolSchema struct {
//	    Name        string                 `json:"name"`
//	    Description string                 `json:"description"`
//	    InputSchema map[string]interface{} `json:"inputSchema"`
//	    Required    []string               `json:"required"`
//	}
//
// # Parameter Naming Convention
//
// All tool parameters use snake_case:
//
//	file_path    - Path to a file
//	old_string   - String to find (Edit tool)
//	new_string   - Replacement string (Edit tool)
//	pattern      - Glob or grep pattern
//	command      - Shell command
//	description  - Command description
//
// # Tool Registry
//
// The registry provides tool schema access:
//
//	registry := tools.NewRegistry()
//
//	// Get tool schema
//	schema, ok := registry.Get("Read")
//
//	// List all tools
//	allTools := registry.List()
//
//	// Validate tool call
//	if err := registry.Validate("Read", params); err != nil {
//	    return err
//	}
//
// # Tool Handlers
//
// Each tool has a corresponding handler:
//
//	type ReadHandler struct {
//	    logger *logrus.Logger
//	}
//
//	func (h *ReadHandler) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
//	    filePath := params["file_path"].(string)
//	    // Read file and return contents
//	}
//
// # Tool Execution
//
// Tools are executed through the handler registry:
//
//	handlerRegistry := tools.NewHandlerRegistry()
//	result, err := handlerRegistry.Execute(ctx, "Read", params)
//
// # Provider Tool Support
//
// Tools are exposed to LLM providers that support function calling:
//
//	capabilities := provider.GetCapabilities()
//	if capabilities.SupportsTools {
//	    request.Tools = registry.GetToolSchemas()
//	}
//
// # Key Files
//
//   - schema.go: Tool schema definitions
//   - handler.go: Tool handlers
//   - registry.go: Tool registry
//   - validation.go: Parameter validation
//
// # Example: Tool Call
//
//	params := map[string]interface{}{
//	    "file_path": "/path/to/file.go",
//	    "offset":    0,
//	    "limit":     100,
//	}
//
//	result, err := handlerRegistry.Execute(ctx, "Read", params)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	content := result.(string)
package tools
