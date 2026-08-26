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
- Hand-written code does not live as a flat dump of many files in
  one directory. Group by responsibility into subfolders as a
  package grows — a directory you can't summarize in one sentence
  is a directory that needs splitting.
- Exception: fully generated code (anything produced by a
  `go:generate` directive or equivalent, e.g. `internal/db/ent/`)
  is exempt. Its layout is dictated by the generator and gets
  overwritten on every regeneration — do not hand-reorganize it,
  the next `go generate` will just flatten it again.

## Functions

- If a function needs a comment to explain what it does, rename it.
- Comments explain why, never what.

## Comments

- Do not write comments. No exceptions for restating what a line
  does, labeling a section, or narrating a change.
- The single allowed case: one line explaining a genuinely
  non-obvious *why* — a hidden constraint, a workaround for a
  specific bug, behavior that would surprise a reader. If removing
  it wouldn't confuse someone, don't write it.
- Before finishing any change, grep the files you touched for
  stray comments and remove them.
