# Conventions

## Commits

Format: `feature(scope): short imperative summary`

- Subject line only. No body, no bullet list, no explanation.
- No `Co-Authored-By` and no tool attribution of any kind.
- Scope is the package or app: `relay`, `cli`, `protocol`, `web`, `mcp`.
- Present tense, lowercase after the colon, no trailing period.

Example: `feature(protocol): add tool.blocked event`, `fix(mcp): tool calling log errorz`

## Files

- Hard limit: 300 lines. At 300, split by responsibility, not by
  line count — a 150-line file and a 150-line file that only make
  sense together is worse than one 300-line file.
- One exported concept per file. If you cannot name the file after
  what it exports, the split is wrong.
- No `utils.ts` / `helpers.go`. Name the thing.

## Functions

- If a function needs a comment to explain what it does, rename it.
- Comments explain why, never what.

## Comments

- Never write comments in the code, only write them if necessary and one line
  to make something easy to understand or separate logic.
