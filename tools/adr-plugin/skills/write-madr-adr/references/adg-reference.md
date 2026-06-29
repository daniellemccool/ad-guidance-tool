# `adg` command reference

The fork's CLI surface, as used for ADR authoring. All commands take `--model <path>`
(omit if set via `adg set-config`). Bare 4-digit IDs; filenames are `NNNN-slug.md`.

## Lifecycle

```bash
adg init <model-path>                       # create an empty model directory
adg add --title "<title>" [--id N]          # create a proposed stub; prints bare ID to stdout
adg slug "<title>"                          # print the slug `add` would generate (no side effect)
adg edit --id N --from-stdin                # replace the whole body with MADR markdown from stdin
adg edit --id N --from-file <path>          # ... or from a file
adg decide --id N --option <idx|text> [--rationale "<why>"]   # write the outcome; status -> accepted
adg revise --id N                           # copy a decided ADR into a fresh proposed one
adg supersede --new N --old M [--rationale "<why>"]           # bidirectional supersession
adg validate                                # check the whole model; non-zero exit on any issue
adg list [--status ...] [--tag ...] [--title <regex>] [--id <ranges>] [--format simple|yaml|json|md]
adg view --id N [--context|--options|--outcome|--drivers|--comments]
adg comment --id N --text "<text>" [--author "<name>"]
adg tag --id N --tag <tag>                  # repeatable
adg link --from N --to M [--tag <t> --reverse-tag <r>]        # within-model link
```

## `edit`: whole-body vs. per-section

`--from-stdin` / `--from-file` replace the **entire** body and are the default authoring
path. A handful of additive flags also exist for targeted appends — `--context`,
`--drivers`, `--option` (repeatable) — but prefer whole-body input: it's one call, easy to
review, and adg validates the result. Editing a non-proposed ADR requires `--force`.

## `decide`: the outcome guard

`decide` writes `Chosen option: "X", because Y.` into Decision Outcome and flips status to
`accepted`. Two guards, both bypassed by `--force`:

1. It refuses to **overwrite an authored Decision Outcome** — it only fills a placeholder
   (empty, `{...}`, or the unedited `adg add` template). This protects hand-written prose
   and any nested `### Consequences`.
2. It refuses to **re-decide an already-accepted** ADR.

`--option` accepts a 1-based index into the Considered Options bullets, or the option's
exact text (case-insensitive, trimmed — not fuzzy). Prefer the index for options containing
quotes or backticks.

## What `adg validate` enforces

`validate` exits non-zero on any issue and, on success, prints a per-check breakdown.
Checks include:

- filenames match `NNNN-slug.md`; H1 titles present
- required sections present (Context, Considered Options, Decision Outcome)
- Considered Options has bullets; accepted ADRs name a valid Chosen option
- **accepted ADRs carry no leftover template placeholders** (`{...}`, `{driver 1}`,
  `{option 1/2}`, `{option title}`, `{justification}`)
- **accepted ADRs have non-empty required sections**
- status vocabulary; supersession links (forward + reverse integrity)
- comments well-formed (non-empty, non-placeholder)

Proposed ADRs are exempt from the accepted-only checks — they are work in progress.

## Status vocabulary

MADR statuses: `proposed` → `accepted` → `deprecated` → `superseded by ADR-NNNN`.
`adg add` writes `proposed`; `adg decide` writes `accepted`. There is no manual status
patching and no separate index to rebuild — the model is the directory of files.

## stdout vs. stderr

`adg add` writes the new ID to **stdout** (for `$(...)` capture); human-readable status
goes to **stderr**. `--quiet` suppresses status without affecting machine values or errors.

## Cross-reference notes

- Links (`adg link`) are within-model. Reference other context in prose or via `adg comment`.
- `adg migrate` (with `--dry-run`) rewrites legacy ADRs toward the current format; `import`,
  `copy`, and `merge` move decisions between models. These are maintenance commands, not
  part of the everyday authoring loop.
