#!/usr/bin/env sh
# adg-session-start — SessionStart hook. Two model-independent jobs, so a session knows
# the governance exists even before any lean record does (the whole-corpus brief is empty
# until records carry applies_to, and a read-only/exploring session meets no other hook):
#   1. Greet — when this repo has an ADR model (docs/decisions/), announce that the
#      write-adr governance is active and where its entry points are.
#   2. Version check — if the system `adg` is missing or older than this plugin ships for,
#      tell the agent to prompt the user to run install.sh.
# SessionStart stdout is added to the agent's context. Fail-open: any error prints nothing.
set -eu

# The plugin is global; only greet where an ADR model actually exists.
[ -d docs/decisions ] || exit 0

echo "This repo is governed by the write-adr plugin: architecture decisions live as lean ADRs in docs/decisions/, enforced by \`adg\` and Claude Code hooks. Entry points — pull the brief for files you'll touch (\`adg lean brief --model docs/decisions <paths>\`); author / migrate / review records with the write-lean-adr skill (\`adg lean new\`); obey an injected brief with follow-adr-governance. If the routing hooks stay silent, the lean model may not be populated yet (records need \`applies_to\` frontmatter) — bootstrap the lean records before relying on the brief."

# Version check.
root="${CLAUDE_PLUGIN_ROOT:-}"
[ -n "$root" ] || exit 0
need=$(sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$root/.claude-plugin/plugin.json" 2>/dev/null | head -1)
[ -n "$need" ] || exit 0
have=$(adg --version 2>/dev/null | awk '{print $NF}')
install='curl -fsSL https://raw.githubusercontent.com/daniellemccool/ad-guidance-tool/main/install.sh | sh'

if [ -z "$have" ]; then
    echo "NOTE: the \`adg\` CLI these hooks depend on is not on PATH — tell the user to install it and reload: \`$install\` (fish: keep \`| sh\`; no VAR=value prefix)."
    exit 0
fi
[ "$have" = "$need" ] && exit 0
older=$(printf '%s\n%s\n' "$have" "$need" | sort -V | head -1)
if [ "$older" = "$have" ]; then
    echo "NOTE: the system \`adg\` is v$have but this plugin ships for v$need — the governance hooks misbehave on the old version. Tell the user to update and reload: \`$install\` (fish: keep \`| sh\`; no VAR=value prefix)."
fi
exit 0
