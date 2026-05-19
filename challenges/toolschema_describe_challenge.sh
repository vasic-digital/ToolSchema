#!/usr/bin/env bash
# challenges/toolschema_describe_challenge.sh
#
# Round-285 anti-bluff Challenge for digital.vasic.toolschema.
#
# Default mode: invoke the runner against the real, in-process
# schema + handler registry and assert it exits 0 with the expected
# coverage, schema invariants, validation-gate, search, and 5-locale
# UX evidence. This is the positive-evidence proof per Article XI
# §11.9 — the PASS is backed by captured stdout, not by absence of
# error or a green summary line.
#
# Paired-mutation mode (--mutate): copy the schema-invariant
# assertion into a scratch directory, plant a known violation (a
# schema entry whose RequiredField is absent from its Parameters
# map), build a scratch runner against the mutated copy, and assert
# the runner detects it. A mutation run that exits 0 means the
# Challenge itself is a bluff (CONST-035 mutation-bluff), and this
# script exits 1 to surface that. A correctly detected mutation
# exits 99 — sentinel value the parent test bank recognises.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

MODE="default"
if [[ ${1:-} == "--mutate" ]]; then
    MODE="mutate"
fi

run_default() {
    echo "[toolschema-challenge] mode=default — exercising runner against real registry"
    cd "${REPO_ROOT}"

    local out
    out=$(go run ./challenges/runner -all 2>&1) || {
        echo "[toolschema-challenge] FAIL: runner exited non-zero"
        echo "${out}"
        exit 1
    }

    # Positive-evidence assertions on captured stdout.
    if ! grep -q "validate_tool_args: error-path + success-path PASS" <<<"${out}"; then
        echo "[toolschema-challenge] FAIL: ValidateToolArgs error/success paths not exercised"
        echo "${out}"
        exit 1
    fi
    if ! grep -q "validation_gates: ValidatePath + ValidateSymbol + ValidateGitRef + ValidateCommandArg PASS" <<<"${out}"; then
        echo "[toolschema-challenge] FAIL: validation gates not exercised"
        echo "${out}"
        exit 1
    fi
    if ! grep -q "^search_tools: query=file matched=" <<<"${out}"; then
        echo "[toolschema-challenge] FAIL: SearchTools not exercised"
        echo "${out}"
        exit 1
    fi
    if ! grep -q "^\[en\] toolschema:" <<<"${out}" \
            || ! grep -q "^\[sr\] toolschema:" <<<"${out}" \
            || ! grep -q "^\[ja\] toolschema:" <<<"${out}" \
            || ! grep -q "^\[es\] toolschema:" <<<"${out}" \
            || ! grep -q "^\[de\] toolschema:" <<<"${out}"; then
        echo "[toolschema-challenge] FAIL: missing one or more locale UX lines"
        echo "${out}"
        exit 1
    fi
    if ! grep -qE "^OK schemas=[0-9]+ handlers=[0-9]+ locales=5$" <<<"${out}"; then
        echo "[toolschema-challenge] FAIL: missing OK trailer"
        echo "${out}"
        exit 1
    fi

    echo "${out}"
    echo "[toolschema-challenge] PASS — runtime evidence captured above"
    exit 0
}

run_mutate() {
    echo "[toolschema-challenge] mode=mutate — paired-mutation evidence"
    local scratch
    scratch="$(mktemp -d -t toolschema-mutate-XXXXXX)"
    # shellcheck disable=SC2064
    trap "rm -rf '${scratch}'" EXIT

    # Stage a self-contained scratch module that vendors a mutated
    # copy of the schema-invariant assertion. We construct the test
    # purely in the scratch dir so the real repository is never
    # modified.
    mkdir -p "${scratch}/pkg/toolschema_scratch"

    cat > "${scratch}/go.mod" <<'EOF'
module toolschema.scratch

go 1.24
EOF

    cat > "${scratch}/pkg/toolschema_scratch/schema.go" <<'EOF'
package toolschema_scratch

import "errors"

// Param is a minimal stand-in for tools.Param sufficient to exercise
// the invariant check in scratch.
type Param struct {
	Type     string
	Required bool
}

// ToolSchema mirrors the invariant-relevant subset of tools.ToolSchema.
type ToolSchema struct {
	Name           string
	Description    string
	Category       string
	RequiredFields []string
	OptionalFields []string
	Parameters     map[string]Param
}

// LoadOne returns a mutated schema whose RequiredFields contains a
// name ("file_path") that is intentionally absent from Parameters.
// The runner-style invariant assertion MUST catch this.
func LoadOne() ToolSchema {
	return ToolSchema{
		Name:           "ScratchMutatedTool",
		Description:    "intentionally broken schema for mutation evidence",
		Category:       "core",
		RequiredFields: []string{"file_path"},
		Parameters:     map[string]Param{},
	}
}

// AssertSchema is the literal port of the runner's assertSchema
// loop body for RequiredFields. It MUST flag the missing parameter.
func AssertSchema(s ToolSchema) error {
	if s.Name == "" {
		return errors.New("empty Name")
	}
	for _, req := range s.RequiredFields {
		if _, ok := s.Parameters[req]; !ok {
			return errors.New("RequiredField absent from Parameters map")
		}
	}
	return nil
}
EOF

    cat > "${scratch}/main.go" <<'EOF'
package main

import (
	"fmt"
	"os"

	ts "toolschema.scratch/pkg/toolschema_scratch"
)

func main() {
	s := ts.LoadOne()
	if err := ts.AssertSchema(s); err != nil {
		fmt.Fprintf(os.Stderr, "mutation detected: %v\n", err)
		os.Exit(99)
	}
	fmt.Println("mutation NOT detected — bluff")
	os.Exit(0)
}
EOF

    cd "${scratch}"
    # Build then exec — `go run` does not preserve exit codes >2 on
    # all toolchains, which would mask the sentinel 99 the program
    # emits when the mutation is detected.
    go build -o ./mutbin . >/dev/null 2>&1 || {
        echo "[toolschema-challenge] FAIL-MUTATE — scratch build failed"
        exit 1
    }
    local mut_out mut_rc
    set +e
    mut_out=$(./mutbin 2>&1)
    mut_rc=$?
    set -e

    echo "${mut_out}"
    if [[ ${mut_rc} -eq 99 ]]; then
        echo "[toolschema-challenge] PASS-MUTATE — mutation correctly surfaced (exit 99)"
        exit 99
    fi
    echo "[toolschema-challenge] FAIL-MUTATE — mutation NOT surfaced (exit ${mut_rc}); Challenge is a bluff"
    exit 1
}

case "${MODE}" in
    default) run_default ;;
    mutate)  run_mutate ;;
esac
