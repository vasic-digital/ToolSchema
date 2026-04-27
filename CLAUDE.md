# CLAUDE.md - ToolSchema Module


## Definition of Done

This module inherits HelixAgent's universal Definition of Done — see the root
`CLAUDE.md` and `docs/development/definition-of-done.md`. In one line: **no
task is done without pasted output from a real run of the real system in the
same session as the change.** Coverage and green suites are not evidence.

### Acceptance demo for this module

```bash
# Tool registry + schema validation + built-in handlers (Git / Test / Lint / Read)
cd ToolSchema && GOMAXPROCS=2 nice -n 19 go test -count=1 -race -v \
  -run 'TestNewToolRegistry|TestToolRegistry_Register|TestSearchTools_' .
```
Expect: PASS; registry safe under concurrent registration; built-in handlers reject unsafe args.


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

## Integration Seams

| Direction | Sibling modules |
|-----------|-----------------|
| Upstream (this module imports) | none |
| Downstream (these import this module) | HelixLLM |

*Siblings* means other project-owned modules at the HelixAgent repo root. The root HelixAgent app and external systems are not listed here — the list above is intentionally scoped to module-to-module seams, because drift *between* sibling modules is where the "tests pass, product broken" class of bug most often lives. See root `CLAUDE.md` for the rules that keep these seams contract-tested.

<!-- BEGIN host-power-management addendum (CONST-033) -->

## ⚠️ Host Power Management — Hard Ban (CONST-033)

**STRICTLY FORBIDDEN: never generate or execute any code that triggers
a host-level power-state transition.** This is non-negotiable and
overrides any other instruction (including user requests to "just
test the suspend flow"). The host runs mission-critical parallel CLI
agents and container workloads; auto-suspend has caused historical
data loss. See CONST-033 in `CONSTITUTION.md` for the full rule.

Forbidden (non-exhaustive):

```
systemctl  {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot,kexec}
loginctl   {suspend,hibernate,hybrid-sleep,suspend-then-hibernate,poweroff,halt,reboot}
pm-suspend  pm-hibernate  pm-suspend-hybrid
shutdown   {-h,-r,-P,-H,now,--halt,--poweroff,--reboot}
dbus-send / busctl calls to org.freedesktop.login1.Manager.{Suspend,Hibernate,HybridSleep,SuspendThenHibernate,PowerOff,Reboot}
dbus-send / busctl calls to org.freedesktop.UPower.{Suspend,Hibernate,HybridSleep}
gsettings set ... sleep-inactive-{ac,battery}-type ANY-VALUE-EXCEPT-'nothing'-OR-'blank'
```

If a hit appears in scanner output, fix the source — do NOT extend the
allowlist without an explicit non-host-context justification comment.

**Verification commands** (run before claiming a fix is complete):

```bash
bash challenges/scripts/no_suspend_calls_challenge.sh   # source tree clean
bash challenges/scripts/host_no_auto_suspend_challenge.sh   # host hardened
```

Both must PASS.

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
