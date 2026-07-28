package tools

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"text/template"

	"gopkg.in/yaml.v3"
)

// bundleEntry mirrors the go-i18n message shape in active.en.yaml:
// a top-level key per message ID with an `other` plural form.
type bundleEntry struct {
	Other string `yaml:"other"`
}

// enBundleTranslator loads the real i18n/bundles/active.en.yaml file
// and returns a Translator that resolves message IDs against it,
// interpolating {{.Placeholder}} via text/template. This is an
// anti-bluff probe (CONST-035): it proves the committed bundle file
// is a valid, loadable, render-correct resource — a dead or
// malformed bundle would fail the tests that wire this translator.
// Unit-test-only (CONST-050(A)).
func enBundleTranslator() Translator {
	raw, err := os.ReadFile("i18n/bundles/active.en.yaml")
	if err != nil {
		panic("CONST-046: en bundle missing/unreadable: " + err.Error())
	}
	var entries map[string]bundleEntry
	if err := yaml.Unmarshal(raw, &entries); err != nil {
		panic("CONST-046: en bundle malformed YAML: " + err.Error())
	}
	return &bundleTranslator{entries: entries}
}

type bundleTranslator struct {
	entries map[string]bundleEntry
}

// T resolves id against the loaded bundle and renders placeholders.
// A missing ID echoes loudly (never silent) so a stale bundle is
// caught by the wiring tests.
func (b *bundleTranslator) T(id string, data map[string]any) string {
	e, ok := b.entries[id]
	if !ok || e.Other == "" {
		return "MISSING_BUNDLE_ID[" + id + "]"
	}
	if len(data) == 0 {
		return e.Other
	}
	tmpl, err := template.New(id).Parse(e.Other)
	if err != nil {
		return "BAD_TEMPLATE[" + id + "]"
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "BAD_RENDER[" + id + "]"
	}
	return buf.String()
}

// TestI18n_EnBundle_ResolvesEveryMigratedID asserts every message ID
// referenced by a CONST-046-migrated callsite exists in the real
// bundle and renders its placeholder. PAIRED-MUTATION: deleting a
// bundle entry, or reverting a callsite's ID, surfaces here.
func TestI18n_EnBundle_ResolvesEveryMigratedID(t *testing.T) {
	tr := enBundleTranslator()

	cases := []struct {
		id   string
		data map[string]any
		want string
	}{
		{"toolschema_err_unknown_tool", map[string]any{"Tool": "Frob"}, "unknown tool: Frob"},
		{"toolschema_err_invalid_file_path", map[string]any{"Path": "/x"}, "invalid file path: /x"},
		{"toolschema_err_invalid_path", map[string]any{"Path": "/x"}, "invalid path: /x"},
		{"toolschema_err_invalid_git_operation", map[string]any{"Operation": "frob"}, "invalid git operation: frob"},
		{"toolschema_err_invalid_working_dir", map[string]any{"Path": "/x"}, "invalid working directory path: /x"},
		{"toolschema_err_dangerous_arg", map[string]any{"Arg": "x;y"}, "invalid argument contains dangerous characters: x;y"},
		{"toolschema_err_invalid_git_reference", map[string]any{"Reference": "r"}, "invalid git reference: r"},
		{"toolschema_err_invalid_symbol", map[string]any{"Symbol": "s"}, "invalid symbol: s"},
		{"toolschema_err_unsupported_linter", map[string]any{"Linter": "l"}, "unsupported linter: l"},
		{"toolschema_err_dangerous_ignore_pattern", map[string]any{"Pattern": "p"}, "invalid ignore pattern contains dangerous characters: p"},
		{"toolschema_git_desc_commit", nil, "Create git commit"},
		{"toolschema_git_desc_status", nil, "Check git status"},
		{"toolschema_git_desc_push", nil, "Push changes to remote"},
		{"toolschema_git_desc_pull", nil, "Pull changes from remote"},
		{"toolschema_git_desc_branch", nil, "List or create branches"},
		{"toolschema_git_desc_checkout", nil, "Checkout branch or file"},
		{"toolschema_git_desc_merge", nil, "Merge branches"},
		{"toolschema_git_desc_diff", nil, "Show differences"},
		{"toolschema_git_desc_log", nil, "Show commit history"},
		{"toolschema_git_desc_stash", nil, "Stash changes"},
		{"toolschema_test_desc_default", nil, "Run tests"},
		{"toolschema_test_desc_coverage", nil, "Run tests with coverage"},
		{"toolschema_test_desc_unit", nil, "Run unit tests"},
		{"toolschema_test_desc_integration", nil, "Run integration tests"},
		{"toolschema_test_desc_e2e", nil, "Run end-to-end tests"},
		{"toolschema_lint_desc_default", nil, "Run code linting"},
		// Round-371 residual sweep.
		{"toolschema_desc_read_file", nil, "Read file contents"},
		{"toolschema_desc_git_diff", nil, "Show git diff"},
		{"toolschema_desc_tree_view", nil, "Display directory tree"},
		{"toolschema_desc_file_info", nil, "Get file information"},
		{"toolschema_desc_extract_symbols", nil, "Extract code symbols"},
		{"toolschema_desc_find_references", nil, "Find symbol references"},
		{"toolschema_desc_find_definition", nil, "Find symbol definition"},
		{"toolschema_desc_list_issues", nil, "List issues"},
		{"toolschema_desc_list_workflows", nil, "List workflows"},
		{"toolschema_pr_desc_list", nil, "List pull requests"},
		{"toolschema_pr_desc_create", nil, "Create pull request"},
		{"toolschema_pr_desc_merge", nil, "Merge pull request"},
		{"toolschema_pr_desc_view", nil, "View pull request"},
		{"toolschema_err_unknown_action", map[string]any{"Action": "frob"}, "unknown action: frob"},
		{"toolschema_err_dangerous_title", map[string]any{"Title": "x;y"}, "invalid title contains dangerous characters: x;y"},
		{"toolschema_err_dangerous_body", map[string]any{"Body": "x;y"}, "invalid body contains dangerous characters: x;y"},
		{"toolschema_err_invalid_base_branch", map[string]any{"Branch": "b"}, "invalid base branch: b"},
		{"toolschema_err_dangerous_workflow_id", map[string]any{"WorkflowID": "w;y"}, "invalid workflow ID contains dangerous characters: w;y"},
		{"toolschema_err_invalid_branch", map[string]any{"Branch": "b"}, "invalid branch: b"},
	}
	for _, c := range cases {
		got := tr.T(c.id, c.data)
		if got != c.want {
			t.Errorf("bundle ID %q: got %q want %q", c.id, got, c.want)
		}
	}
}

// recordingTranslator is a unit-test-only fake (CONST-050(A): fakes
// permitted in *_test.go) that captures every message ID + payload
// it is asked to resolve and returns a deterministic rendering. It
// lets the paired-mutation tests prove that CONST-046-migrated
// callsites actually route through the Translator seam rather than
// emitting a hardcoded literal.
type recordingTranslator struct {
	calls map[string]map[string]any
}

func newRecordingTranslator() *recordingTranslator {
	return &recordingTranslator{calls: make(map[string]map[string]any)}
}

// T records the call and returns a sentinel-prefixed rendering so a
// test can assert the resolved string really came from the seam.
func (r *recordingTranslator) T(id string, data map[string]any) string {
	r.calls[id] = data
	out := "TR[" + id + "]"
	for k, v := range data {
		out += " " + k + "=" + toStr(v)
	}
	return out
}

func toStr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// TestI18n_NoopTranslator_EchoesID verifies NoopTranslator's loud-echo
// contract: an explicitly-wired NoopTranslator returns the message ID
// verbatim, never an empty string (a silent swallow would be a §11.4
// PASS-bluff). NoopTranslator is opt-in (see i18n.go) — it is no
// longer what SetTranslator(nil) restores to, so it is wired
// explicitly here rather than via nil.
func TestI18n_NoopTranslator_EchoesID(t *testing.T) {
	SetTranslator(NoopTranslator{})
	defer SetTranslator(nil)

	got := tr("toolschema_err_unknown_tool", map[string]any{"Tool": "Frob"})
	if got != "toolschema_err_unknown_tool" {
		t.Fatalf("NoopTranslator must echo the message ID; got %q", got)
	}
}

// TestI18n_DefaultTranslator_UnknownID_EchoesLoud proves the PACKAGE
// DEFAULT (what SetTranslator(nil) restores to since the go:embed
// fix — see i18n_bundle.go) preserves the same never-silent-swallow
// guarantee for a message ID that genuinely does not exist in the
// bundle: it must echo the raw ID, never return an empty string.
func TestI18n_DefaultTranslator_UnknownID_EchoesLoud(t *testing.T) {
	SetTranslator(nil) // restore package default (embedded bundle translator)
	defer SetTranslator(nil)

	const unknownID = "toolschema_this_id_does_not_exist_in_any_bundle"
	got := tr(unknownID, nil)
	if got != unknownID {
		t.Fatalf("default translator must echo an unknown message ID loudly, never silently swallow it; got %q, want %q", got, unknownID)
	}
}

// TestI18n_SetTranslator_RoutesThroughSeam proves a wired Translator
// receives the call. PAIRED-MUTATION: if a future edit reverts a
// migrated callsite back to a hardcoded literal, this test still
// passes for tr() directly — so the handler-level tests below are
// the real mutation guard. This one guards the seam plumbing.
func TestI18n_SetTranslator_RoutesThroughSeam(t *testing.T) {
	rec := newRecordingTranslator()
	SetTranslator(rec)
	defer SetTranslator(nil)

	got := tr("toolschema_err_invalid_symbol", map[string]any{"Symbol": "x;rm"})
	if !strings.HasPrefix(got, "TR[toolschema_err_invalid_symbol]") {
		t.Fatalf("tr() must route through the wired Translator; got %q", got)
	}
	if _, ok := rec.calls["toolschema_err_invalid_symbol"]; !ok {
		t.Fatalf("wired Translator did not receive the message ID")
	}
}

// TestI18n_UnknownTool_UsesSeam is a PAIRED-MUTATION test for the
// CONST-046 migration of Execute's "unknown tool" error. If the
// callsite is reverted to fmt.Sprintf("unknown tool: %s", ...) the
// recorded ID disappears and this test FAILS — that is the mutation
// detector the §11.4 sweep requires.
func TestI18n_UnknownTool_UsesSeam(t *testing.T) {
	rec := newRecordingTranslator()
	SetTranslator(rec)
	defer SetTranslator(nil)

	reg := NewToolRegistry()
	res, _ := reg.Execute(context.Background(), "NoSuchTool", nil)

	if res.Success {
		t.Fatalf("expected failure for unknown tool")
	}
	data, ok := rec.calls["toolschema_err_unknown_tool"]
	if !ok {
		t.Fatalf("Execute() unknown-tool error did not route through the i18n seam — CONST-046 regression")
	}
	if data["Tool"] != "NoSuchTool" {
		t.Fatalf("unknown-tool error lost its Tool placeholder; got %v", data["Tool"])
	}
	if !strings.Contains(res.Error, "toolschema_err_unknown_tool") {
		t.Fatalf("ToolResult.Error must carry the seam-resolved string; got %q", res.Error)
	}
}

// TestI18n_GitHandler_DescriptionUsesSeam is a PAIRED-MUTATION test
// for the GenerateDefaultArgs git-operation descriptions. Reverting
// any description back to a literal drops the recorded message ID.
func TestI18n_GitHandler_DescriptionUsesSeam(t *testing.T) {
	rec := newRecordingTranslator()
	SetTranslator(rec)
	defer SetTranslator(nil)

	h := &GitHandler{}
	args := h.GenerateDefaultArgs("please commit my work")

	if _, ok := rec.calls["toolschema_git_desc_commit"]; !ok {
		t.Fatalf("GitHandler commit description did not route through the i18n seam — CONST-046 regression")
	}
	desc, _ := args["description"].(string)
	if !strings.Contains(desc, "toolschema_git_desc_commit") {
		t.Fatalf("git description must carry the seam-resolved string; got %q", desc)
	}
}

// TestI18n_TestHandler_DescriptionUsesSeam is a PAIRED-MUTATION
// test for the TestHandler default-arg descriptions.
func TestI18n_TestHandler_DescriptionUsesSeam(t *testing.T) {
	rec := newRecordingTranslator()
	SetTranslator(rec)
	defer SetTranslator(nil)

	h := &TestHandler{}
	args := h.GenerateDefaultArgs("run the unit tests")

	if _, ok := rec.calls["toolschema_test_desc_unit"]; !ok {
		t.Fatalf("TestHandler unit description did not route through the i18n seam — CONST-046 regression")
	}
	desc, _ := args["description"].(string)
	if !strings.Contains(desc, "toolschema_test_desc_unit") {
		t.Fatalf("test description must carry the seam-resolved string; got %q", desc)
	}
}

// TestI18n_Round371_DescriptionsUseSeam is a PAIRED-MUTATION test for
// the round-371 GenerateDefaultArgs description migrations. Reverting
// any description to a hardcoded literal drops the recorded ID and
// fails the matching subtest — CONST-046 regression detector.
func TestI18n_Round371_DescriptionsUseSeam(t *testing.T) {
	cases := []struct {
		name    string
		gen     func() map[string]interface{}
		wantID  string
	}{
		{"read_file", func() map[string]interface{} { return (&ReadFileHandler{}).GenerateDefaultArgs("read it") }, "toolschema_desc_read_file"},
		{"git_diff", func() map[string]interface{} { return (&DiffHandler{}).GenerateDefaultArgs("show diff") }, "toolschema_desc_git_diff"},
		{"tree_view", func() map[string]interface{} { return (&TreeViewHandler{}).GenerateDefaultArgs("show tree") }, "toolschema_desc_tree_view"},
		{"file_info", func() map[string]interface{} { return (&FileInfoHandler{}).GenerateDefaultArgs("file info") }, "toolschema_desc_file_info"},
		{"symbols", func() map[string]interface{} { return (&SymbolsHandler{}).GenerateDefaultArgs("extract symbols") }, "toolschema_desc_extract_symbols"},
		{"references", func() map[string]interface{} { return (&ReferencesHandler{}).GenerateDefaultArgs("find refs") }, "toolschema_desc_find_references"},
		{"definition", func() map[string]interface{} { return (&DefinitionHandler{}).GenerateDefaultArgs("find definition") }, "toolschema_desc_find_definition"},
		{"issue_list", func() map[string]interface{} { return (&IssueHandler{}).GenerateDefaultArgs("list issues") }, "toolschema_desc_list_issues"},
		{"workflow_list", func() map[string]interface{} { return (&WorkflowHandler{}).GenerateDefaultArgs("list workflows") }, "toolschema_desc_list_workflows"},
		{"pr_list", func() map[string]interface{} { return (&PRHandler{}).GenerateDefaultArgs("list prs") }, "toolschema_pr_desc_list"},
		{"pr_create", func() map[string]interface{} { return (&PRHandler{}).GenerateDefaultArgs("create a pr") }, "toolschema_pr_desc_create"},
		{"pr_merge", func() map[string]interface{} { return (&PRHandler{}).GenerateDefaultArgs("merge the pr") }, "toolschema_pr_desc_merge"},
		{"pr_view", func() map[string]interface{} { return (&PRHandler{}).GenerateDefaultArgs("view the pr") }, "toolschema_pr_desc_view"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := newRecordingTranslator()
			SetTranslator(rec)
			defer SetTranslator(nil)

			args := c.gen()
			if _, ok := rec.calls[c.wantID]; !ok {
				t.Fatalf("%s description did not route through the i18n seam — CONST-046 regression", c.name)
			}
			desc, _ := args["description"].(string)
			if !strings.Contains(desc, c.wantID) {
				t.Fatalf("%s description must carry the seam-resolved string; got %q", c.name, desc)
			}
		})
	}
}

// TestI18n_Round371_ErrorsUseSeam is a PAIRED-MUTATION test for the
// round-371 PR / Issue / Workflow handler error migrations. Each
// subtest drives a handler into a rejection path and asserts the
// error string routed through the i18n seam.
func TestI18n_Round371_ErrorsUseSeam(t *testing.T) {
	cases := []struct {
		name   string
		exec   func() ToolResult
		wantID string
	}{
		{
			name:   "pr_unknown_action",
			exec:   func() ToolResult { r, _ := (&PRHandler{}).Execute(context.Background(), map[string]interface{}{"action": "frobnicate"}); return r },
			wantID: "toolschema_err_unknown_action",
		},
		{
			name:   "pr_dangerous_title",
			exec:   func() ToolResult { r, _ := (&PRHandler{}).Execute(context.Background(), map[string]interface{}{"action": "create", "title": "x; rm -rf /"}); return r },
			wantID: "toolschema_err_dangerous_title",
		},
		{
			name:   "pr_dangerous_body",
			exec:   func() ToolResult { r, _ := (&PRHandler{}).Execute(context.Background(), map[string]interface{}{"action": "create", "body": "y && curl evil"}); return r },
			wantID: "toolschema_err_dangerous_body",
		},
		{
			name:   "pr_invalid_base_branch",
			exec:   func() ToolResult { r, _ := (&PRHandler{}).Execute(context.Background(), map[string]interface{}{"action": "create", "base_branch": "bad;branch"}); return r },
			wantID: "toolschema_err_invalid_base_branch",
		},
		{
			name:   "workflow_dangerous_id",
			exec:   func() ToolResult { r, _ := (&WorkflowHandler{}).Execute(context.Background(), map[string]interface{}{"action": "run", "workflow_id": "w; rm"}); return r },
			wantID: "toolschema_err_dangerous_workflow_id",
		},
		{
			name:   "workflow_invalid_branch",
			exec:   func() ToolResult { r, _ := (&WorkflowHandler{}).Execute(context.Background(), map[string]interface{}{"action": "run", "workflow_id": "ci.yml", "branch": "bad;branch"}); return r },
			wantID: "toolschema_err_invalid_branch",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rec := newRecordingTranslator()
			SetTranslator(rec)
			defer SetTranslator(nil)

			res := c.exec()
			if res.Success {
				t.Fatalf("%s: expected a rejection result", c.name)
			}
			if _, ok := rec.calls[c.wantID]; !ok {
				t.Fatalf("%s error did not route through the i18n seam — CONST-046 regression", c.name)
			}
			if !strings.Contains(res.Error, c.wantID) {
				t.Fatalf("%s: ToolResult.Error must carry the seam-resolved string; got %q", c.name, res.Error)
			}
		})
	}
}
