# Antibody memory

## Overview
- Repository francogalfre/coop.
- Written in go, javascript, shell, typescript.
- 166 harness files existed when Antibody first ran.

## Stack
- Language: go.
- Language: javascript.
- Language: shell.
- Language: typescript.
- Tool: claude_code (coding_agent), seen in .claude/CLAUDE.md.
- Tool: codex (coding_agent), seen in AGENTS.md.
- Tool: github_actions (continuous_integration), seen in .github/workflows/ci.yml.
- Tool: go_test (test_runner), seen in apps/relay/go.mod.
- Tool: greptile (reviewer), seen in greptile.json.
- Tool: oxlint (linter), seen in .oxlintrc.json.
- Tool: typescript (type_checker), seen in tsconfig.base.json.
- Tool: vitest (test_runner), seen in apps/web/vitest.config.ts.

## Commands
- `dev` runs `bash scripts/dev.sh` (package.json script).
- `build` runs `turbo run build` (package.json script).
- `lint` runs `oxlint .` (package.json script).
- `typecheck` runs `turbo run typecheck` (package.json script).
- `test` runs `turbo run test` (package.json script).
- `bash scripts/check.sh` runs the full fast verification: JS lint, typecheck, tests plus gofmt/go vet/go test for apps/relay and packages/cli. Use it before finishing any change.

## Conventions

## Decisions

## Rejected proposals

## Open questions
