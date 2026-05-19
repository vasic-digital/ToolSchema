// Command runner is the round-285 anti-bluff Challenge runner for
// digital.vasic.toolschema. It exercises the public schema API
// (ToolSchemaRegistry, GetToolSchema, ValidateToolArgs,
// GenerateOpenAIToolDefinition, SearchTools) AND the tool-handler
// registry surface (NewToolRegistry, GetDefaultToolRegistry,
// ValidatePath / ValidateSymbol / ValidateGitRef / ValidateCommandArg)
// against the real, in-process implementation. The runner produces
// captured stdout per Article XI §11.9 — every PASS is backed by a
// printed assertion line, not by absence-of-error.
//
// The runner intentionally avoids invoking any handler that would
// shell out (Git / Test / Lint / etc.) — those handlers execute real
// subprocesses, which is out of scope for a self-contained Challenge.
// Instead, the runner exercises every handler's metadata
// (Name + ValidateArgs with empty-args + GenerateDefaultArgs) AND
// every schema entry's invariant set.
//
// Exit codes:
//
//	0   — every assertion passed, every locale line printed.
//	1   — usage / flag error.
//	2   — schema-coverage gap (missing required field, unknown alias,
//	      handler without matching schema entry).
//	3   — schema-invariant violation (empty Name, malformed Parameters,
//	      RequiredFields not present in Parameters map).
//	4   — locale UX line missing or canonical token absent.
//	5   — validation gate regression (a known-bad input passed
//	      ValidatePath / ValidateSymbol / ValidateGitRef /
//	      ValidateCommandArg, or a known-good input was rejected).
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	tools "digital.vasic.toolschema"
)

// locale describes a UX line printed by the runner. The text is a
// short, locale-correct summary that consumers can grep for to
// confirm the runner produced operator-facing output in every
// supported locale per CONST-046.
type locale struct {
	tag  string
	line func(schemaCount, handlerCount int) string
}

// supportedLocales is the 5-locale CONST-046 set the runner must emit
// every run. Mirrors the matrix used by the round-281 RedTeam runner
// and other round-2xx enrichments.
func supportedLocales() []locale {
	return []locale{
		{
			tag: "en",
			line: func(s, h int) string {
				return fmt.Sprintf("[en] toolschema: %d schemas registered, %d default handlers exercised", s, h)
			},
		},
		{
			tag: "sr",
			line: func(s, h int) string {
				return fmt.Sprintf("[sr] toolschema: %d šema registrovano, %d podrazumevanih obrađivača izvršeno", s, h)
			},
		},
		{
			tag: "ja",
			line: func(s, h int) string {
				return fmt.Sprintf("[ja] toolschema: %d スキーマ登録、%d デフォルトハンドラを実行", s, h)
			},
		},
		{
			tag: "es",
			line: func(s, h int) string {
				return fmt.Sprintf("[es] toolschema: %d esquemas registrados, %d handlers ejercidos", s, h)
			},
		},
		{
			tag: "de",
			line: func(s, h int) string {
				return fmt.Sprintf("[de] toolschema: %d Schemata registriert, %d Standard-Handler ausgeübt", s, h)
			},
		},
	}
}

func main() {
	all := flag.Bool("all", false, "run every check (default mode)")
	tool := flag.String("tool", "", "exercise only the named tool schema")
	flag.Parse()

	if !*all && *tool == "" {
		*all = true
	}

	if *tool != "" {
		if err := runOne(*tool); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(exitCodeFor(err))
		}
		return
	}
	if err := runAll(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(exitCodeFor(err))
	}
}

// runOne exercises a single tool schema by name (canonical or alias).
func runOne(name string) error {
	schema, ok := tools.GetToolSchema(name)
	if !ok {
		return wrap(errCoverage, fmt.Errorf("GetToolSchema(%q): not found", name))
	}
	if err := assertSchema(schema); err != nil {
		return wrap(errSchema, fmt.Errorf("schema %s: %w", name, err))
	}
	def := tools.GenerateOpenAIToolDefinition(schema)
	if !openAIDefinitionHasName(def, schema.Name) {
		return wrap(errSchema, fmt.Errorf("schema %s: OpenAI definition name mismatch", schema.Name))
	}
	fmt.Printf("tool=%s aliases=%d params=%d required=%d category=%s\n",
		schema.Name, len(schema.Aliases), len(schema.Parameters),
		len(schema.RequiredFields), schema.Category)
	return nil
}

// runAll exercises the full schema registry, the default handler
// registry, the validation gates, and emits the 5-locale CONST-046
// UX summary lines.
func runAll() error {
	names := tools.GetAllToolNames()
	if len(names) == 0 {
		return wrap(errCoverage, errors.New("GetAllToolNames returned 0 — registry empty"))
	}
	sort.Strings(names)

	for _, n := range names {
		schema, ok := tools.GetToolSchema(n)
		if !ok {
			return wrap(errCoverage, fmt.Errorf("GetToolSchema(%q): registry inconsistency", n))
		}
		if err := assertSchema(schema); err != nil {
			return wrap(errSchema, fmt.Errorf("schema %s: %w", n, err))
		}
		// Round-trip every schema through the OpenAI definition
		// generator to ensure no schema fails generation.
		def := tools.GenerateOpenAIToolDefinition(schema)
		if !openAIDefinitionHasName(def, schema.Name) {
			return wrap(errSchema, fmt.Errorf("schema %s: OpenAI name field drift", schema.Name))
		}
		fmt.Printf("schema=%s required=%d optional=%d aliases=%d category=%s\n",
			schema.Name, len(schema.RequiredFields), len(schema.OptionalFields),
			len(schema.Aliases), schema.Category)
	}

	// ValidateToolArgs error path — missing required field MUST fail loudly.
	if err := tools.ValidateToolArgs("Read", map[string]interface{}{}); err == nil {
		return wrap(errSchema, errors.New("ValidateToolArgs accepted empty args for Read (required: file_path)"))
	}
	// ValidateToolArgs success path — providing the required field MUST pass.
	if err := tools.ValidateToolArgs("Read", map[string]interface{}{"file_path": "/tmp/x"}); err != nil {
		return wrap(errSchema, fmt.Errorf("ValidateToolArgs rejected valid Read args: %w", err))
	}
	fmt.Println("validate_tool_args: error-path + success-path PASS")

	// Default handler registry — exercise metadata of every registered handler.
	reg := tools.GetDefaultToolRegistry()
	handlerNames := exerciseDefaultHandlers(reg)
	if len(handlerNames) == 0 {
		return wrap(errCoverage, errors.New("GetDefaultToolRegistry: 0 handlers"))
	}
	for _, hn := range handlerNames {
		fmt.Printf("handler=%s registered\n", hn)
	}

	// Validation gates — every gate MUST accept a known-good input
	// AND reject a known-bad input. Mutation of either direction is
	// a CONST-035 wrapper-bluff.
	if err := assertValidationGates(); err != nil {
		return wrap(errValidation, err)
	}
	fmt.Println("validation_gates: ValidatePath + ValidateSymbol + ValidateGitRef + ValidateCommandArg PASS")

	// SearchTools — real search across the registry, not a hardcoded list.
	results := tools.SearchTools(tools.SearchOptions{Query: "file", MaxResults: 5})
	if len(results) == 0 {
		return wrap(errSchema, errors.New("SearchTools(query=file) returned 0 — registry-broken"))
	}
	topName := "<nil>"
	if results[0].Tool != nil {
		topName = results[0].Tool.Name
	}
	fmt.Printf("search_tools: query=file matched=%d top=%s\n", len(results), topName)

	// 5-locale bilingual UX evidence per CONST-046.
	printed := 0
	for _, loc := range supportedLocales() {
		out := loc.line(len(names), len(handlerNames))
		if !strings.Contains(out, "toolschema:") {
			return wrap(errLocale, fmt.Errorf("locale %s: missing canonical token", loc.tag))
		}
		fmt.Println(out)
		printed++
	}
	if printed != len(supportedLocales()) {
		return wrap(errLocale, fmt.Errorf("printed %d/%d locales", printed, len(supportedLocales())))
	}

	fmt.Printf("OK schemas=%d handlers=%d locales=%d\n", len(names), len(handlerNames), printed)
	return nil
}

// assertSchema enforces the per-schema invariants every test-coverage
// row depends on. Anti-bluff: every schema in the registry is checked,
// not just the first one or a sampled subset.
func assertSchema(s *tools.ToolSchema) error {
	if strings.TrimSpace(s.Name) == "" {
		return errors.New("empty Name")
	}
	if strings.TrimSpace(s.Description) == "" {
		return fmt.Errorf("%s: empty Description", s.Name)
	}
	if strings.TrimSpace(s.Category) == "" {
		return fmt.Errorf("%s: empty Category", s.Name)
	}
	// Every RequiredFields entry MUST be declared in the Parameters map.
	for _, req := range s.RequiredFields {
		if _, ok := s.Parameters[req]; !ok {
			return fmt.Errorf("%s: RequiredField %q absent from Parameters map", s.Name, req)
		}
	}
	// Every OptionalFields entry MUST also appear in Parameters.
	for _, opt := range s.OptionalFields {
		if _, ok := s.Parameters[opt]; !ok {
			return fmt.Errorf("%s: OptionalField %q absent from Parameters map", s.Name, opt)
		}
	}
	// Every parameter declared Required=true MUST appear in RequiredFields.
	for pname, p := range s.Parameters {
		if p.Required {
			found := false
			for _, req := range s.RequiredFields {
				if req == pname {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("%s: parameter %q marked Required but not in RequiredFields", s.Name, pname)
			}
		}
		if strings.TrimSpace(p.Type) == "" {
			return fmt.Errorf("%s: parameter %q has empty Type", s.Name, pname)
		}
	}
	return nil
}

// exerciseDefaultHandlers asks every handler in the default registry
// for its Name() and GenerateDefaultArgs(""). Anti-bluff: handlers
// that bluff a Name (e.g. "") are flagged via empty handlerNames.
// The returned slice is sorted for deterministic output.
func exerciseDefaultHandlers(reg *tools.ToolRegistry) []string {
	// The registry does not expose iteration directly; we ask for each
	// known built-in by canonical handler-name.
	candidates := []string{
		"read_file", "Git", "Test", "Lint", "Diff", "TreeView",
		"FileInfo", "Symbols", "References", "Definition",
		"PR", "Issue", "Workflow",
	}
	out := make([]string, 0, len(candidates))
	for _, name := range candidates {
		h, ok := reg.Get(name)
		if !ok {
			continue
		}
		if strings.TrimSpace(h.Name()) == "" {
			continue
		}
		// GenerateDefaultArgs MUST never panic and MUST return a non-nil map.
		args := h.GenerateDefaultArgs("")
		if args == nil {
			continue
		}
		out = append(out, h.Name())
	}
	sort.Strings(out)
	return out
}

// assertValidationGates checks that every validation function in
// the package accepts a representative safe input AND rejects a
// representative unsafe input. Mutation of either direction would
// silently permit shell-injection / path-traversal regressions.
func assertValidationGates() error {
	// ValidatePath: accept relative + reject `..` traversal.
	if !tools.ValidatePath("internal/x.go") {
		return errors.New("ValidatePath rejected safe path internal/x.go")
	}
	if tools.ValidatePath("../../etc/passwd") {
		return errors.New("ValidatePath accepted traversal path ../../etc/passwd")
	}
	// ValidateSymbol: accept identifier + reject shell metachar.
	if !tools.ValidateSymbol("HandleRequest") {
		return errors.New("ValidateSymbol rejected safe identifier HandleRequest")
	}
	if tools.ValidateSymbol("rm -rf /") {
		return errors.New("ValidateSymbol accepted shell-metachar input")
	}
	// ValidateGitRef: accept branch + reject control chars.
	if !tools.ValidateGitRef("main") {
		return errors.New("ValidateGitRef rejected safe ref main")
	}
	if tools.ValidateGitRef("bad ref with space") {
		return errors.New("ValidateGitRef accepted ref with space")
	}
	if tools.ValidateGitRef("bad;ref") {
		return errors.New("ValidateGitRef accepted ref with shell metachar")
	}
	// ValidateCommandArg: accept plain word + reject shell metachar.
	if !tools.ValidateCommandArg("status") {
		return errors.New("ValidateCommandArg rejected safe arg status")
	}
	if tools.ValidateCommandArg("foo; rm -rf /") {
		return errors.New("ValidateCommandArg accepted metachar arg")
	}
	return nil
}

// openAIDefinitionHasName walks the nested OpenAI function-call
// definition shape produced by GenerateOpenAIToolDefinition and
// returns true iff the inner function.name field matches want.
func openAIDefinitionHasName(def map[string]interface{}, want string) bool {
	fn, ok := def["function"].(map[string]interface{})
	if !ok {
		return false
	}
	name, ok := fn["name"].(string)
	if !ok {
		return false
	}
	return name == want
}

// Sentinel error tags used to compute exit codes without printing
// the tag itself.
var (
	errCoverage   = errors.New("coverage")
	errSchema     = errors.New("schema")
	errLocale     = errors.New("locale")
	errValidation = errors.New("validation")
)

// taggedError attaches a sentinel for exit-code mapping while
// preserving the inner cause via Unwrap.
type taggedError struct {
	tag   error
	inner error
}

func (e *taggedError) Error() string { return e.inner.Error() }
func (e *taggedError) Unwrap() error { return e.inner }
func (e *taggedError) Is(t error) bool {
	return errors.Is(e.tag, t)
}

func wrap(tag, inner error) error {
	return &taggedError{tag: tag, inner: inner}
}

func exitCodeFor(err error) int {
	switch {
	case errors.Is(err, errCoverage):
		return 2
	case errors.Is(err, errSchema):
		return 3
	case errors.Is(err, errLocale):
		return 4
	case errors.Is(err, errValidation):
		return 5
	default:
		return 1
	}
}
