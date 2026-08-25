#!/usr/bin/env bash
set -euo pipefail

out="${1:-/tmp/coop-probe.jsonl}"
scratch="$(mktemp -d)"
: > "$out"

mkdir -p "$scratch/.claude"
cat > "$scratch/.claude/settings.local.json" <<EOF
{
  "hooks": {
    "SessionStart": [{"hooks": [{"type": "command", "command": "cat >> $out"}]}],
    "SessionEnd": [{"hooks": [{"type": "command", "command": "cat >> $out"}]}],
    "UserPromptSubmit": [{"hooks": [{"type": "command", "command": "cat >> $out"}]}],
    "PreToolUse": [{"hooks": [{"type": "command", "command": "cat >> $out"}]}],
    "PostToolUse": [{"hooks": [{"type": "command", "command": "cat >> $out"}]}],
    "Stop": [{"hooks": [{"type": "command", "command": "cat >> $out"}]}]
  }
}
EOF

echo "scratch dir:  $scratch"
echo "capture file: $out"
echo
echo "Run a normal claude session there, do a few representative things"
echo "(read a file, edit a file, run a shell command), then exit:"
echo
echo "  cd $scratch && claude"
echo
echo "Each captured line is one hook's raw JSON payload. Compare it against"
echo "the field names documented in .agents/rules/harnesses.md and correct"
echo "anything that's drifted."
