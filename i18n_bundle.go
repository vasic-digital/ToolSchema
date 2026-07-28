// i18n_bundle.go supplies ToolSchema's DEFAULT Translator: a bundle
// compiled directly into the binary via go:embed. This closes a real
// consumer-facing gap (see i18n_cwd_regression_test.go): prior to
// this file, the package-level default Translator was NoopTranslator{}
// — a loud message-ID echo — so ANY consumer that never called
// SetTranslator (HelixAgent's tool_integration test suite among them)
// always saw raw message IDs ("toolschema_git_desc_commit") instead
// of real text ("Create git commit"), unconditionally, in every
// process and every working directory.
//
// go:embed compiles the bundle bytes into the binary at `go build`
// time, so resolving them at runtime has NO dependency whatsoever on
// the caller's current working directory, deployment layout, or how
// the package was packaged/vendored — unlike a relative-path
// os.ReadFile("i18n/bundles/active.en.yaml"), which only succeeds
// when the process cwd happens to equal this module's own source
// directory (true for `go test` invoked on this package directly,
// false for essentially every other real invocation shape: a
// consuming binary running from its own working directory, a test
// binary copied to a temp dir, a packaged release binary, etc).
package tools

import (
	"bytes"
	_ "embed"
	"sync"
	"text/template"

	"gopkg.in/yaml.v3"
)

// embeddedEnBundleYAML holds the raw bytes of the canonical English
// bundle, compiled into the binary. This is the SAME file
// (i18n/bundles/active.en.yaml) that enBundleTranslator() in
// i18n_test.go reads from disk for an independent, source-file-level
// anti-bluff check (TestI18n_EnBundle_ResolvesEveryMigratedID) that
// the committed YAML itself is well-formed — the embed below is what
// production/consumer code actually runs on.
//
//go:embed i18n/bundles/active.en.yaml
var embeddedEnBundleYAML []byte

// bundleMessage mirrors the go-i18n message shape used by
// i18n/bundles/active.en.yaml: a top-level key per message ID with an
// `other` plural form.
type bundleMessage struct {
	Other string `yaml:"other"`
}

// embeddedBundleTranslator implements Translator by resolving message
// IDs against a bundle parsed once from embedded bytes.
type embeddedBundleTranslator struct {
	entries map[string]bundleMessage
}

// T resolves messageID against the embedded bundle, interpolating
// {{.Placeholder}} data via text/template. A messageID with no bundle
// entry, or a template that fails to parse/execute, echoes the
// messageID verbatim — the same loud-echo contract NoopTranslator
// documents, never a silent swallow (CONST-046).
func (b *embeddedBundleTranslator) T(messageID string, data map[string]any) string {
	e, ok := b.entries[messageID]
	if !ok || e.Other == "" {
		return messageID
	}
	if len(data) == 0 {
		return e.Other
	}
	tmpl, err := template.New(messageID).Parse(e.Other)
	if err != nil {
		return messageID
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return messageID
	}
	return buf.String()
}

var (
	defaultTranslatorOnce sync.Once
	defaultTranslatorInst *embeddedBundleTranslator
)

// defaultTranslator lazily parses the embedded bundle exactly once
// and returns the shared instance. Because embeddedEnBundleYAML is
// baked into the binary at compile time, parsing it can never fail
// due to any runtime filesystem/cwd/deployment condition — a parse
// failure here is a genuine bundle-syntax defect that would be caught
// by this package's own test suite (go build embeds the same bytes
// the tests exercise) on every single run, never a field-only
// failure mode.
func defaultTranslator() *embeddedBundleTranslator {
	defaultTranslatorOnce.Do(func() {
		var entries map[string]bundleMessage
		if err := yaml.Unmarshal(embeddedEnBundleYAML, &entries); err != nil {
			panic("digital.vasic.toolschema: embedded i18n bundle i18n/bundles/active.en.yaml is malformed: " + err.Error())
		}
		defaultTranslatorInst = &embeddedBundleTranslator{entries: entries}
	})
	return defaultTranslatorInst
}
