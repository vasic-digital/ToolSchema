// i18n.go declares ToolSchema's hardcoded-content abstraction per
// CONST-046 (round-322 §11.4 anti-bluff sweep, 2026-05-20). Mirrors
// the "consumer defines its own Translator interface" pattern of
// every prior CONST-046-migrated package in this codebase (most
// recently helix_code/internal/approval/i18n, round-221).
//
// ToolSchema is a fully decoupled, project-not-aware submodule
// (CONST-051(B)): it declares only the Translator contract + a
// loud-echo NoopTranslator fallback. The consuming binary builds a
// real Translator (e.g. a go-i18n-backed adapter loaded with the
// bundles under i18n/bundles) and injects it via SetTranslator at
// boot. Until wired, the package-level tr() helper falls back to
// NoopTranslator{} — a loud message-ID echo, never a silent swallow
// (which would be a §11.4 PASS-bluff at the i18n layer).
package tools

import "sync"

// Translator is the contract ToolSchema uses for every
// CONST-046-migrated user-facing string. A consuming project wires
// a real implementation that resolves messageID against the active
// locale and interpolates templateData placeholders.
type Translator interface {
	// T resolves messageID against the active locale. templateData
	// supplies named placeholders for go-i18n style interpolation;
	// pass nil when the message has no placeholders.
	T(messageID string, templateData map[string]any) string
}

// NoopTranslator returns the messageID verbatim. SAFETY default for
// unit tests + backward compatibility for callers who have not yet
// wired a real Translator. Production paths SHOULD inject a real
// Translator via SetTranslator at boot.
type NoopTranslator struct{}

// T returns id unchanged (loud echo). Never panics.
func (NoopTranslator) T(id string, _ map[string]any) string { return id }

var (
	trMu         sync.RWMutex
	activeTr     Translator = NoopTranslator{}
)

// SetTranslator installs the active Translator. Passing nil restores
// the loud-echo NoopTranslator. Safe for concurrent use.
func SetTranslator(t Translator) {
	trMu.Lock()
	defer trMu.Unlock()
	if t == nil {
		activeTr = NoopTranslator{}
		return
	}
	activeTr = t
}

// tr resolves messageID via the active Translator. data carries
// named placeholders (nil when none). Package-internal helper used
// by every CONST-046-migrated literal in this package.
func tr(messageID string, data map[string]any) string {
	trMu.RLock()
	t := activeTr
	trMu.RUnlock()
	return t.T(messageID, data)
}

// errInvalidFilePath produces the locale-resolved "invalid file
// path" error message for the given path. CONST-046 helper: keeps
// the message ID + placeholder shape in one place so the four
// handlers that reject a bad file path stay consistent.
func errInvalidFilePath(path string) string {
	return tr("toolschema_err_invalid_file_path", map[string]any{"Path": path})
}

// errInvalidPath produces the locale-resolved "invalid path" error
// message. Used by handlers that validate a directory or generic
// filesystem path argument.
func errInvalidPath(path string) string {
	return tr("toolschema_err_invalid_path", map[string]any{"Path": path})
}

// errInvalidSymbol produces the locale-resolved "invalid symbol"
// error message for the given symbol token.
func errInvalidSymbol(symbol string) string {
	return tr("toolschema_err_invalid_symbol", map[string]any{"Symbol": symbol})
}

// errUnknownAction produces the locale-resolved "unknown action"
// error message. Shared by the PR / Issue / Workflow handlers when an
// agent requests an action outside their allowed-action set.
func errUnknownAction(action string) string {
	return tr("toolschema_err_unknown_action", map[string]any{"Action": action})
}

// errDangerousTitle produces the locale-resolved error for a PR /
// Issue title containing shell-unsafe characters.
func errDangerousTitle(title string) string {
	return tr("toolschema_err_dangerous_title", map[string]any{"Title": title})
}

// errDangerousBody produces the locale-resolved error for a PR /
// Issue body containing shell-unsafe characters.
func errDangerousBody(body string) string {
	return tr("toolschema_err_dangerous_body", map[string]any{"Body": body})
}

// errInvalidBaseBranch produces the locale-resolved error for an
// invalid base-branch git reference passed to the PR handler.
func errInvalidBaseBranch(branch string) string {
	return tr("toolschema_err_invalid_base_branch", map[string]any{"Branch": branch})
}

// errDangerousWorkflowID produces the locale-resolved error for a
// workflow ID containing shell-unsafe characters.
func errDangerousWorkflowID(id string) string {
	return tr("toolschema_err_dangerous_workflow_id", map[string]any{"WorkflowID": id})
}

// errInvalidBranch produces the locale-resolved error for an invalid
// branch git reference passed to the Workflow handler.
func errInvalidBranch(branch string) string {
	return tr("toolschema_err_invalid_branch", map[string]any{"Branch": branch})
}
