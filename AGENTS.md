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
<!-- BEGIN host-power-management addendum (CONST-033) -->

## Host Power Management — Hard Ban (CONST-033)

**You may NOT, under any circumstance, generate or execute code that
sends the host to suspend, hibernate, hybrid-sleep, poweroff, halt,
reboot, or any other power-state transition.** This rule applies to:

- Every shell command you run via the Bash tool.
- Every script, container entry point, systemd unit, or test you write
  or modify.
- Every CLI suggestion, snippet, or example you emit.

**Forbidden invocations** (non-exhaustive — see CONST-033 in
`CONSTITUTION.md` for the full list):

- `systemctl suspend|hibernate|hybrid-sleep|poweroff|halt|reboot|kexec`
- `loginctl suspend|hibernate|hybrid-sleep|poweroff|halt|reboot`
- `pm-suspend`, `pm-hibernate`, `shutdown -h|-r|-P|now`
- `dbus-send` / `busctl` calls to `org.freedesktop.login1.Manager.Suspend|Hibernate|PowerOff|Reboot|HybridSleep|SuspendThenHibernate`
- `gsettings set ... sleep-inactive-{ac,battery}-type` to anything but `'nothing'` or `'blank'`

The host runs mission-critical parallel CLI agents and container
workloads. Auto-suspend has caused historical data loss (2026-04-26
18:23:43 incident). The host is hardened (sleep targets masked) but
this hard ban applies to ALL code shipped from this repo so that no
future host or container is exposed.

**Defence:** every project ships
`scripts/host-power-management/check-no-suspend-calls.sh` (static
scanner) and
`challenges/scripts/no_suspend_calls_challenge.sh` (challenge wrapper).
Both MUST be wired into the project's CI / `run_all_challenges.sh`.

**Full background:** `docs/HOST_POWER_MANAGEMENT.md` and `CONSTITUTION.md` (CONST-033).

<!-- END host-power-management addendum (CONST-033) -->



<!-- CONST-035 anti-bluff addendum (cascaded) -->

## CONST-035 — Anti-Bluff Tests & Challenges (mandatory; inherits from root)

Tests and Challenges in this submodule MUST verify the product, not
the LLM's mental model of the product. A test that passes when the
feature is broken is worse than a missing test — it gives false
confidence and lets defects ship to users. Functional probes at the
protocol layer are mandatory:

- TCP-open is the FLOOR, not the ceiling. Postgres → execute
  `SELECT 1`. Redis → `PING` returns `PONG`. ChromaDB → `GET
  /api/v1/heartbeat` returns 200. MCP server → TCP connect + valid
  JSON-RPC handshake. HTTP gateway → real request, real response,
  non-empty body.
- Container `Up` is NOT application healthy. A `docker/podman ps`
  `Up` status only means PID 1 is running; the application may be
  crash-looping internally.
- No mocks/fakes outside unit tests (already CONST-030; CONST-035
  raises the cost of a mock-driven false pass to the same severity
  as a regression).
- Re-verify after every change. Don't assume a previously-passing
  test still verifies the same scope after a refactor.
- Verification of CONST-035 itself: deliberately break the feature
  (e.g. `kill <service>`, swap a password). The test MUST fail. If
  it still passes, the test is non-conformant and MUST be tightened.

## CONST-033 clarification — distinguishing host events from sluggishness

Heavy container builds (BuildKit pulling many GB of layers, parallel
podman/docker compose-up across many services) can make the host
**appear** unresponsive — high load average, slow SSH, watchers
timing out. **This is NOT a CONST-033 violation.** Suspend / hibernate
/ logout are categorically different events. Distinguish via:

- `uptime` — recent boot? if so, the host actually rebooted.
- `loginctl list-sessions` — session(s) still active? if yes, no logout.
- `journalctl ... | grep -i 'will suspend\|hibernate'` — zero broadcasts
  since the CONST-033 fix means no suspend ever happened.
- `dmesg | grep -i 'killed process\|out of memory'` — OOM kills are
  also NOT host-power events; they're memory-pressure-induced and
  require their own separate fix (lower per-container memory limits,
  reduce parallelism).

A sluggish host under build pressure recovers when the build finishes;
a suspended host requires explicit unsuspend (and CONST-033 should
make that impossible by hardening `IdleAction=ignore` +
`HandleSuspendKey=ignore` + masked `sleep.target`,
`suspend.target`, `hibernate.target`, `hybrid-sleep.target`).

If you observe what looks like a suspend during heavy builds, the
correct first action is **not** "edit CONST-033" but `bash
challenges/scripts/host_no_auto_suspend_challenge.sh` to confirm the
hardening is intact. If hardening is intact AND no suspend
broadcast appears in journal, the perceived event was build-pressure
sluggishness, not a power transition.
