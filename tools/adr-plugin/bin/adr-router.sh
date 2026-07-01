#!/usr/bin/env sh
# adr-router — UserPromptSubmit hook. When a prompt is about ADRs / the adg decision
# model, print a pointer (UserPromptSubmit stdout is added to the agent's context) so ADR
# work goes through the write-adr skills + adg instead of being reinvented by hand. This
# is the deterministic version of "load the skill when ADRs come up" — a grep fires it,
# not the model's unreliable skill auto-discovery. Fail-open: any error prints nothing.
set -eu

payload=$(cat 2>/dev/null) || exit 0

# ADR (the acronym) is matched case-sensitively with word boundaries so it does not fire
# on substrings like "Madrid"; the phrase keywords are case-insensitive.
if printf '%s' "$payload" | grep -qE '\b(ADR|ADRs)\b' \
   || printf '%s' "$payload" | grep -qiE 'architecture decision|docs/decisions|lean (adr|record)|supersed|adg lean'; then
    :
else
    exit 0
fi

cat <<'MSG'
This task touches ADRs / the architecture-decision model. This repo governs ADRs with the `adg` CLI and the write-adr skills — do not hand-roll ADR work:
- Author / migrate / rewrite / review a record -> use the write-lean-adr skill: `adg lean new` (add `--date <original>` to preserve a migrated record's date), `adg lean review`, `adg lean index`.
- Obey the architecture brief while changing code -> the follow-adr-governance skill; pull the brief with `adg lean brief --model docs/decisions <paths>`.
`adg` is on PATH. If an `adg` command errors, surface it — do not fall back to editing records by hand.
MSG
exit 0
