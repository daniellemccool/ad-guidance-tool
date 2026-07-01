#!/usr/bin/env sh
# adg-version-check — SessionStart hook. Compare the system `adg` on PATH against the
# version this plugin ships for. If it is missing or older, print a note (SessionStart
# stdout is added to the agent's context) telling the agent to prompt the user to update.
# Silent when adg is current. Must not depend on any particular adg version — a stale
# adg is exactly the case it detects — so it is plain POSIX sh, not an `adg` subcommand.
set -eu

root="${CLAUDE_PLUGIN_ROOT:-}"
[ -n "$root" ] || exit 0
need=$(sed -n 's/.*"version"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$root/.claude-plugin/plugin.json" 2>/dev/null | head -1)
[ -n "$need" ] || exit 0

have=$(adg --version 2>/dev/null | awk '{print $NF}')
install='curl -fsSL https://raw.githubusercontent.com/daniellemccool/ad-guidance-tool/main/install.sh | sh'

if [ -z "$have" ]; then
    echo "NOTE for the agent: the \`adg\` CLI that the write-adr governance hooks depend on is not on PATH. Tell the user to install it and reload the session — run: \`$install\` (fish users: keep the \`| sh\`; do NOT use a VAR=value prefix). Until then the hooks emit nothing."
    exit 0
fi

[ "$have" = "$need" ] && exit 0
older=$(printf '%s\n%s\n' "$have" "$need" | sort -V | head -1)
if [ "$older" = "$have" ]; then
    echo "NOTE for the agent: the system \`adg\` is v$have but the write-adr plugin ships for v$need. The governance hooks (session brief, ADR guard, commit review) rely on flags newer adg provides, so they misbehave on the old version. Tell the user to update and reload the session — run: \`$install\` (fish: keep \`| sh\`; no VAR=value prefix)."
fi
exit 0
