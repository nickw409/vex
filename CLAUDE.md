# Vex

Test coverage auditor that verifies tests fully cover intended behavior described in a spec. Designed for AI agent consumption, not direct human use.

## Quick Reference

```bash
go test ./...
go build -o vex ./cmd/vex/
```

## Project Structure

```
cmd/vex/          Entry point (main.go)
internal/
  cli/            Cobra command definitions
  config/         vex.yaml parsing
  provider/       LLM provider abstraction (claude-cli)
  spec/           vexspec.yaml parsing, validation, generation
  check/          Core gap detection engine (two-pass)
  diff/           Git diff and drift detection
  lang/           Language detection and test file discovery
  report/         JSON output formatting
```

## Key Conventions

- Go 1.24, module `github.com/nwiley/vex`
- CLI built with `spf13/cobra`
- Tests use stdlib `testing` only — no external test frameworks
- Test files live alongside source (`*_test.go`)
- JSON output to stdout by default (agents consume it, not humans)
- Exit code 0 = no gaps, exit code 1 = gaps found, exit code 2 = fatal error

## Commands

```bash
vex check                                        # check test coverage against spec
vex check --section Config                       # check single section
vex check --drift                                # only check sections changed since last check
vex validate                                     # validate spec completeness
vex spec "description"                           # generate spec sections from task
vex spec "description" --extend Config           # add behaviors to existing section
vex drift                                        # check for code changes since last check
vex init                                         # create vex.yaml config
vex guide                                        # print agent instructions for writing specs
```

## Design Principles

- **Spec-driven** — spec is the source of truth, not the code
- **Two-pass check** — pass 1 sends only tests (cheap triage), pass 2 sends source only for uncovered behaviors
- **Drift-aware** — `--drift` skips clean sections, converging cost toward zero for stable code
- **Language agnostic** — auto-detects Go, TypeScript, JavaScript, Python, Java
- **Agent-first** — JSON output, config files over CLI flags, guide command for agent instructions
- **Bounded** — spec defines the scope, no infinite nitpicking
