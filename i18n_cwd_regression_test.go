package tools

import (
	"testing"
)

// TestI18n_DefaultTranslator_ResolvesWithoutWiring_RegardlessOfCWD is
// the permanent regression guard for a defect reported against a
// consuming module's test binary: HelixAgent's
// tests/unit/tool_integration.TestToolRegistry_GenerateDefaultArgs
// expects GitHandler.GenerateDefaultArgs to produce the real,
// human-readable description ("Create git commit") but observed the
// raw i18n message ID ("toolschema_git_desc_commit") instead.
//
// The failure is 100% deterministic and is NOT caused by any
// cwd-relative bundle path inside this package: digital.vasic.toolschema
// has no bundle-loading code of any kind in its production source (the
// only relative os.ReadFile of i18n/bundles/active.en.yaml lives in
// enBundleTranslator() in i18n_test.go, a unit-test-only helper that
// Go's test tooling always runs with cwd == this package's own
// directory). The actual cause is that tr() falls back to the
// package-default Translator whenever no consumer has ever called
// SetTranslator — and before this fix, that default was
// NoopTranslator{}, a loud ID-echo stub, not a translator that
// resolves real bundle content.
//
// This test proves the package DEFAULT (i.e. what every unwired
// consumer, including HelixAgent, actually gets) resolves real
// bundle content with ZERO wiring — and does so from a process
// working directory that is deliberately NOT this module's root, to
// rule out any future regression to a relative-path (os.ReadFile
// based) bundle loader, which WOULD break under exactly this
// condition.
//
// PAIRED MUTATION (§1.1): reverting the package default back to
// NoopTranslator{} (or reintroducing a relative-path bundle read)
// makes this test FAIL again with the exact symptom reported above.
func TestI18n_DefaultTranslator_ResolvesWithoutWiring_RegardlessOfCWD(t *testing.T) {
	// Force the package back to its zero-wired state: as if no
	// consumer had ever called SetTranslator. SetTranslator(nil) is
	// the documented "restore default" path.
	SetTranslator(nil)
	defer SetTranslator(nil)

	// Move the process cwd somewhere that is provably NOT this
	// module's root: a bare temp directory has no i18n/bundles
	// subdirectory at all, so any relative-path bundle read would
	// fail here even though it might succeed when `go test`'s own
	// cwd happens to be this package's directory.
	t.Chdir(t.TempDir())

	h := &GitHandler{}
	args := h.GenerateDefaultArgs("commit my changes")

	desc, _ := args["description"].(string)
	const want = "Create git commit"
	if desc != want {
		t.Fatalf(
			"default (unwired) translator must resolve real bundle text independent of the caller's working directory; got %q, want %q "+
				"(this reproduces HelixAgent tests/unit/tool_integration.TestToolRegistry_GenerateDefaultArgs/Git_commit_my_changes)",
			desc, want,
		)
	}
}

// TestI18n_DefaultTranslator_ResolvesWithoutWiring_InPackageRoot is the
// same assertion as above WITHOUT changing cwd, run from this
// package's own root (the directory `go test` uses by default). It
// exists to prove the fix produces IDENTICAL output in both
// directories — i.e. the in-package-root behaviour is unaffected by
// the fix, only the previously-broken out-of-directory case changes.
func TestI18n_DefaultTranslator_ResolvesWithoutWiring_InPackageRoot(t *testing.T) {
	SetTranslator(nil)
	defer SetTranslator(nil)

	h := &GitHandler{}
	args := h.GenerateDefaultArgs("commit my changes")

	desc, _ := args["description"].(string)
	const want = "Create git commit"
	if desc != want {
		t.Fatalf("default (unwired) translator must resolve real bundle text; got %q, want %q", desc, want)
	}
}
