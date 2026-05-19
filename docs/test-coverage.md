# test-coverage.md — digital.vasic.toolschema

Round 285 symbol → test / Challenge ledger. Every exported symbol of
`digital.vasic.toolschema` MUST appear here with the test(s) and
Challenge(s) that exercise it AND the anti-bluff dimension each
proves. Adding an exported symbol without updating this ledger is a
CONST-048 violation. Per Article XI §11.9, every PASS row MUST carry
positive runtime evidence — the "Evidence" column documents what to
capture during a release-gate sweep.

## Exported symbols — schema layer

| Symbol                          | Kind   | Unit test(s)                                                      | Challenge(s)                          | Anti-bluff dimension                                                                              | Evidence (runtime)                                       |
|---------------------------------|--------|-------------------------------------------------------------------|---------------------------------------|---------------------------------------------------------------------------------------------------|----------------------------------------------------------|
| `ToolSchema`                    | type   | `TestGetToolSchema_*`, `TestGenerateOpenAIToolDefinition_*`       | `runner -all`, `runner -tool=<name>`  | Type identity flows from registry entry to OpenAI tool definition.                                | Challenge stdout `schema=<Name>` line per entry.         |
| `Param`                         | type   | `TestValidateToolArgs_*`                                          | `runner -all`                         | Parameter type metadata enforced in validation path.                                              | Challenge asserts `Type != ""` per parameter.            |
| `ToolExample`                   | type   | `TestGetToolSchema_*` (returns full schema)                       | n/a (metadata only)                   | Example documentation survives round-trip through registry.                                       | Schema introspection in unit tests.                      |
| `ToolSchemaRegistry`            | var    | `TestGetAllToolNames_NonEmpty`                                    | `runner -all` (iterates registry)     | Single source of truth — runner enumerates every entry, no hardcoded list.                        | Challenge stdout `OK schemas=N`.                         |
| `GetToolSchema`                 | func   | `TestGetToolSchema_DirectMatch`, `TestGetToolSchema_AliasMatch`   | `runner -tool=<name>`                 | Direct + alias resolution; unknown name returns `false` (not silent miss).                        | Challenge `coverage` exit-2 on unknown name.             |
| `GetRequiredFields`             | func   | `TestGetRequiredFields_*`                                         | `runner -all` (invariant check)       | Required-field list matches Parameters map declarations.                                          | Challenge `assertSchema` cross-check.                    |
| `ValidateToolArgs`              | func   | `TestValidateToolArgs_MissingRequired`, `TestValidateToolArgs_OK` | `runner -all`                         | Error path AND success path both exercised — symmetric proof.                                     | Challenge prints `validate_tool_args: ... PASS`.         |
| `GetAllToolNames`               | func   | `TestGetAllToolNames_NonEmpty`                                    | `runner -all`                         | Real iteration over registry, not hardcoded list.                                                 | Challenge iterates returned slice.                       |
| `GetToolsByCategory`            | func   | `TestGetToolsByCategory_*`                                        | `runner -all` (category in summary)   | Category filter returns matching entries.                                                         | Per-schema `category=<C>` line.                          |
| `GenerateOpenAIToolDefinition`  | func   | `TestGenerateOpenAIToolDefinition_*`                              | `runner -all`                         | Round-trip: schema → OpenAI fn definition → name match.                                           | Challenge asserts `function.name == schema.Name`.        |
| `GenerateAllToolDefinitions`    | func   | `TestGenerateAllToolDefinitions_LengthMatches`                    | n/a (covered by round-trip per entry) | Bulk generator returns one definition per registry entry.                                         | Length assertion in unit tests.                          |
| `(*ToolSchema).ToJSON`          | method | `TestToolSchema_ToJSON_*`                                         | n/a (serialisation only)              | Schema is JSON-marshalable for transport.                                                         | Unit test JSON-roundtrip.                                |
| `ToolSearchResult`              | type   | `TestSearchTools_*`                                               | `runner -all`                         | Search result carries `Tool *ToolSchema` pointer (not opaque ID).                                 | Challenge dereferences top result.                       |
| `SearchOptions`                 | type   | `TestSearchTools_*`                                               | `runner -all`                         | Query + MaxResults + Categories options honored.                                                  | Challenge calls with explicit `MaxResults: 5`.           |
| `SearchTools`                   | func   | `TestSearchTools_QueryFile_MatchesFilesystem`                     | `runner -all`                         | Real text-search across registry, not a hardcoded "if query == foo" branch.                       | Challenge prints `search_tools: matched=N top=<Name>`.   |
| `SearchByKeywords`              | func   | `TestSearchByKeywords_*`                                          | n/a (composed with SearchTools)       | Multi-keyword filter narrows result set.                                                          | Unit test asserts narrowing.                             |
| `GetToolSuggestions`            | func   | `TestGetToolSuggestions_*`                                        | n/a (prefix-only)                     | Prefix matching returns capped, ranked suggestions.                                               | Unit test asserts cap honored.                           |
| `CategoryCore`, `CategoryFileSystem`, `CategoryVersionControl`, `CategoryCodeIntel`, `CategoryWorkflow`, `CategoryWeb` | const | `TestGetToolsByCategory_*` | `runner -all` | Constants resolve to non-empty categories present in registry. | Per-schema `category=<C>` printed. |

## Exported symbols — handler layer

| Symbol                       | Kind      | Unit test(s)                                                          | Challenge(s)                  | Anti-bluff dimension                                                                                 | Evidence (runtime)                                  |
|------------------------------|-----------|-----------------------------------------------------------------------|-------------------------------|------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| `ToolHandler` (interface)    | interface | `TestNewToolRegistry`, `TestToolRegistry_Register`                    | `runner -all`                 | Interface contract honored by every built-in.                                                        | Challenge resolves every handler by `Name()`.        |
| `ToolResult`                 | struct    | `TestToolRegistry_Execute_*`                                          | n/a (handler-internal)        | Standard result shape (Success / Output / Error / Data).                                             | Unit-test structural assertions.                    |
| `ToolRegistry`               | struct    | `TestNewToolRegistry`, `TestToolRegistry_Register_Concurrent`         | `runner -all`                 | Concurrent-safe map of handlers; RWMutex guards reads.                                               | `-race` runs clean.                                 |
| `NewToolRegistry`            | func      | `TestNewToolRegistry`                                                 | `runner -all`                 | Returns ready-to-use registry; no init-order bugs.                                                   | Challenge resolves immediately after construction.  |
| `(*ToolRegistry).Register`   | method    | `TestToolRegistry_Register`, `TestToolRegistry_Register_Concurrent`   | `runner -all` (default reg)   | Registration is total; duplicates replace deterministically.                                         | `-race` PASS; per-handler line.                     |
| `(*ToolRegistry).Get`        | method    | `TestToolRegistry_Get_*`                                              | `runner -all`                 | Lookup returns `(handler, true)` for registered names, `(_, false)` for unknown.                     | Challenge `handler=<N> registered` lines.           |
| `(*ToolRegistry).Execute`    | method    | `TestToolRegistry_Execute_*`                                          | n/a (covered by handler tests)| Dispatches to registered handler with arg validation.                                                | Unit tests cover dispatch + ValidateArgs path.      |
| `GetDefaultToolRegistry`     | func      | n/a (composed)                                                        | `runner -all`                 | Pre-populated registry contains all 13 built-in handlers; sync.Once ensures single init.             | Challenge enumerates 13 handlers, exits OK.         |
| 13 `*Handler` types          | struct    | `TestReadFileHandler_*`, `TestGitHandler_*`, `TestTestHandler_*`, etc.| `runner -all`                 | Each handler implements `Name()`, `ValidateArgs()`, `GenerateDefaultArgs()`, `Execute()`.            | Per-handler line in Challenge stdout.               |

## Exported symbols — validation layer

| Symbol                  | Kind   | Unit test(s)                       | Challenge(s)  | Anti-bluff dimension                                                                                   | Evidence (runtime)                                                              |
|-------------------------|--------|------------------------------------|---------------|--------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------------------|
| `ValidatePath`          | func   | `TestValidatePath_*`               | `runner -all` | Accept-known-safe + reject-known-unsafe pair; mutation flips either direction → CONST-035 surfaces.    | Challenge `validation_gates: ... PASS` line.                                    |
| `ValidateSymbol`        | func   | `TestValidateSymbol_*`             | `runner -all` | Accept identifier + reject shell metachar.                                                             | Same as above.                                                                  |
| `SanitizePath`          | func   | `TestSanitizePath_*`               | n/a           | Composed with `ValidatePath` — returns cleaned path + ok=false on unsafe input.                        | Unit-test composition assertions.                                               |
| `ValidateGitRef`        | func   | `TestValidateGitRef_*`             | `runner -all` | Accept canonical branch name + reject space/metachar refs.                                             | Same as above.                                                                  |
| `ValidateCommandArg`    | func   | `TestValidateCommandArg_*`         | `runner -all` | Accept plain word + reject shell metachar arg.                                                         | Same as above.                                                                  |

## Anti-bluff dimensions covered

| Dimension                                                                | Where proved                                                                       |
|--------------------------------------------------------------------------|------------------------------------------------------------------------------------|
| Real registry iteration (not hardcoded list)                             | `GetAllToolNames` → `for name := range ToolSchemaRegistry`                         |
| Real OpenAI definition round-trip (not metadata-only assertion)          | Runner generates definition AND asserts `function.name == schema.Name` per schema  |
| Schema invariant enforcement (RequiredField ↔ Parameters bijection)      | `assertSchema` in runner walks every entry; mutation surfaces missing parameter    |
| Symmetric validation gate proof (accept-good + reject-bad)               | `assertValidationGates` exercises both directions for each gate                    |
| Default handler registry — all 13 handlers reachable                     | `exerciseDefaultHandlers` enumerates registry by canonical name                    |
| Real search across registry (not hardcoded query branch)                 | `SearchTools(query=file)` returns ≥1 result, top result has non-empty Name         |
| 5-locale bilingual UX (CONST-046)                                        | Runner prints `en/sr/ja/es/de` summary lines, each containing canonical `toolschema:` token |
| Paired-mutation evidence (CONST-035)                                     | `--mutate` flag in describe-Challenge plants RequiredField bug, asserts exit 99    |
| Concurrent registration safety                                           | `-race` PASS on `TestToolRegistry_Register_Concurrent`                             |
| Error-path proof for ValidateToolArgs                                    | Runner asserts empty-args call returns non-nil error before asserting success path |

## Maintenance

Every CL that touches `schema.go`, `handler.go`, or `validation.go`
(adds/removes/renames an exported symbol, alters `ToolSchema` /
`Param` shape, registers a new built-in handler, adjusts a validation
gate) MUST update this file in the SAME commit. Drift is a CONST-048
violation. The Challenge runner asserts schema invariants at runtime
— adding a built-in tool without extending the runner's coverage is
a paired CONST-035 violation.
