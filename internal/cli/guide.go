package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const guideText = `# Vex Agent Guide

Vex is a spec-driven test coverage auditor. You write a spec describing
intended behaviors, then vex checks whether your tests cover them.

## Core Concept
The spec is the source of truth — not the code. Write the spec from the
task description BEFORE or DURING implementation. NEVER write a spec by
reading existing code.

## Quick Start

  # Generate spec from task description
  vex spec "Add JWT authentication with login, refresh, and token validation"

  # Validate spec for missing behaviors
  vex validate

  # Check test coverage
  vex check

  # Read results (stdout may be truncated — always read the file)
  cat .vex/report.json

## Spec Format

The spec lives at .vex/vexspec.yaml. All paths are absolute from project root.

  project: MyApp
  description: |
    One-line project description.

  shared:
    - name: error-handling
      description: |
        Behaviors that apply across multiple sections.

  sections:
    - name: Auth
      path: internal/auth
      description: |
        JWT authentication module.
      shared: [error-handling]
      behaviors:
        - name: login
          description: |
            POST /login accepts credentials.
            Returns JWT on success. Returns 401 on invalid credentials.
      subsections:
        - name: Token Refresh
          file: internal/auth/refresh.go
          behaviors:
            - name: refresh
              description: |
                POST /refresh returns new access token.

## Scoping Sections to Code

Use "path" for directories (walks for all source and test files):
  path: internal/auth                      # single directory
  path: [internal/auth, internal/session]  # multiple directories

Use "file" for specific files (classified as source or test automatically):
  file: hn_jobs.py                         # single file
  file: [src/auth.py, tests/test_auth.py]  # source + test files

Both work on sections and subsections. You can combine them:
  path: tests/           # walk for test files
  file: src/auth.py      # just this source file

Vex auto-detects all languages in a directory and discovers files across
all of them. For multi-language projects (e.g., Rust + CUDA), all source
and test files are found regardless of language.

Files listed under "file" are classified as source or test using language
patterns (e.g., test_*.py, *_test.go). This works with any project layout:

  # Tests alongside source
  file: [auth.py, test_auth.py]

  # Separate test directory
  file: [src/auth.py, tests/test_auth.py]

  # Walk a directory + specific files
  path: tests/
  file: [src/auth.py, src/session.py]

Key rules:
- "path" walks a directory tree for all matching files
- "file" includes exactly the files listed (no walking)
- subsections use "path" or "file", not both
- shared behaviors are referenced by name in a section's "shared" list
- both "path" and "file" accept a string or list: "path: foo" or "path: [foo, bar]"

## Writing Behaviors

A behavior IS:
- One observable external behavior with input -> output
- Error cases included inline: "Returns 401 on invalid credentials"
- Side effects stated: "Writes session to database"
- A mathematical formula or equation that the code must implement correctly

A behavior is NOT:
- A data structure ("Report contains these fields")
- An interface definition ("Provider must implement Complete()")
- A list of values ("Supports Go, Python, Java")
- Implementation details ("Uses bcrypt for hashing")

Be specific: "returns 401" not "handles errors". Use kebab-case names.

## Formulas and Equations

When a behavior involves a formula, include the formula directly in the
description. Vex will verify that tests assert mathematical correctness
(known inputs/outputs, boundary conditions, convergence properties),
not just that the function runs without error.

Example:

  behaviors:
    - name: geometric-brownian-motion
      description: |
        Simulates asset price paths using GBM.
        S(t+dt) = S(t) * exp((mu - sigma^2/2)*dt + sigma*sqrt(dt)*Z)
        where Z ~ N(0,1). Must reproduce expected drift and volatility
        over large sample sizes.
    - name: black-scholes-call
      description: |
        Computes European call option price.
        C = S*N(d1) - K*e^(-rT)*N(d2)
        d1 = (ln(S/K) + (r + sigma^2/2)*T) / (sigma*sqrt(T))
        d2 = d1 - sigma*sqrt(T)
        Must match known analytical values for standard inputs.

## Understanding Reports

### Check Report (.vex/report.json)

  {
    "behaviors_checked": 10,
    "gaps": [
      {
        "behavior": "login",
        "detail": "No test for invalid credentials returning 401",
        "suggestion": "TestLoginInvalidCredentials401"
      }
    ],
    "covered": [
      {
        "behavior": "login",
        "detail": "Valid credentials return JWT",
        "test_file": "auth_test.go",
        "test_name": "TestLoginSuccess"
      }
    ],
    "summary": {
      "total_behaviors": 10,
      "fully_covered": 7,
      "gaps_found": 5
    }
  }

A behavior can appear in BOTH gaps and covered (partial coverage).
"fully_covered" counts behaviors in covered but NOT in gaps.

### UNSPECIFIED Gaps

If check finds significant code behavior not in the spec, it reports:

  {"behavior": "UNSPECIFIED", "detail": "Rate limiting logic exists but is not in spec"}

Action: add the missing behavior to the spec, then write a test for it.

### Validation Report (.vex/validation.json)

  {
    "complete": false,
    "suggestions": [
      {
        "section": "Auth",
        "behavior_name": "logout",
        "description": "What happens when user logs out",
        "relation": "new"
      }
    ]
  }

"relation" is "new" for missing behaviors or "extends <name>" for
incomplete ones. "remove: not a behavior" flags non-behavioral entries.

## Exit Codes
- 0: clean (no gaps, spec complete)
- 1: gaps or suggestions found
- 2: fatal error (file not found, invalid config)

## Section Sizing

Keep sections under 10 behaviors. Sections are both the concurrency
boundary and the drift boundary — each section runs as its own LLM
call in parallel, and drift detection skips entire clean sections.

A large section with many subsections still runs as one LLM call and
drifts as one unit. Splitting into separate sections gives you both
faster checks (parallel LLM calls) and finer drift granularity (only
changed sections are re-checked). If a section grows past 10 behaviors,
prefer splitting into independent sections over adding subsections.

Use subsections when behaviors share the same code paths and must be
evaluated together. Use separate sections when behaviors are independent.

## Workflow

### First time (spec authoring)

1. Receive task description
2. Run: vex spec "description"
3. Review .vex/vexspec.yaml — edit if needed
4. Run: vex validate — read .vex/validation.json
5. Address suggestions, repeat step 4 until complete: true

### Implementation loop

6. Implement code and tests
7. Run: vex check — read .vex/report.json
8. For each gap: write the missing test
9. Repeat steps 7-8 until exit code 0

Drift detection is on by default: vex skips sections where neither the
code files nor the spec content have changed since the last check. Use
"vex check --drift=false" to force a full re-check of all sections.

### Ongoing development

Use validate regularly as you evolve the spec:

  vex validate        # ensure spec is complete
  # fix any suggestions
  vex validate        # confirm complete: true

Then check when tests are ready:

  vex check              # drift detection is on by default
  vex check --drift=false  # force full re-check of all sections

## Adding to an Existing Spec

Append new sections:
  vex spec "Add billing with Stripe integration"

Add behaviors to an existing section:
  vex spec "Add password reset flow" --extend Auth

## Commands
  vex check                          # check (drift detection on by default)
  vex check --section "Name"         # check one section
  vex check --drift=false            # force full re-check
  vex validate                       # validate spec completeness
  vex spec "description"             # generate spec sections
  vex spec "desc" --extend "Name"    # add to existing section
  vex drift                          # show which sections changed
  vex init                           # create vex.yaml config
  vex lang add rust --test-patterns "*_test.rs" --source-patterns "*.rs"
  vex lang list                      # list available languages
  vex lang remove rust               # remove a configured language
  vex update                         # update to latest version
  vex guide                          # print this guide

## Output Files
- .vex/vexspec.yaml    — project spec (source of truth, committed to git)
- .vex/report.json     — check report (gaps + covered)
- .vex/validation.json — validation results
- .vex/drift.json      — drift detection results

## Installing & Updating

Install:
  go install github.com/nickw409/vex/cmd/vex@latest

Update to latest:
  vex update

Check current version:
  vex version

IMPORTANT: Always read the full report from these files.
Stdout output may be truncated by your environment.
`

func newGuideCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "guide",
		Short: "Print instructions for writing vexspecs",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprint(cmd.OutOrStdout(), guideText)
		},
	}
}
