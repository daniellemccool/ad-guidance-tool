---
name: write-adr
description: >
  Create and manage Architectural Decision Records (ADRs) in MADR format using the
  `adg` CLI. Use whenever you record, document, or revise an architectural decision,
  capture a design choice or a rejected alternative, or are asked "should this be an
  ADR". Also use proactively after implementing a significant architectural pattern
  that isn't yet recorded. Works in any repo with a `docs/decisions/` model (and
  bootstraps one if absent).
---

# write-adr — record an architectural decision with `adg`

ADRs are MADR-format markdown managed by the `adg` CLI. Three rules carry the whole skill:

1. **Drive everything through `adg`. Never hand-edit ADR files.** The tool enforces the
   shape, round-trips the body, and regenerates the `## Comments` section from frontmatter
   — hand-edits are lost or rejected.
2. **One flat model: `docs/decisions/`.** Files are `NNNN-slug.md` with bare 4-digit IDs.
   No per-domain sub-models — a single chronological sequence per repo.
3. **Keep ADRs small.** See `references/form-factor.md`. The validator enforces the floor
   (no scaffolding, no empty sections); you hold the ceiling — don't pad.

Examples use bare `adg` so they work in any repo. `--model docs/decisions` is shown
explicitly; once you've run `adg set-config` it can be omitted. An optional repo-local
wrapper removes the boilerplate — see "Optional ergonomics".

## Setup (once per repo)

```bash
# Bootstrap the model if docs/decisions/ doesn't exist yet:
adg init docs/decisions

# Install the enforcement hook so `adg validate` gates every commit.
# (Copy from this skill's assets/githooks/pre-commit.)
mkdir -p .githooks && cp assets/githooks/pre-commit .githooks/pre-commit
chmod +x .githooks/pre-commit
git config core.hooksPath .githooks
```

## The authoring loop: add → edit → decide → validate

```bash
# 1. Create a proposed stub. The bare ID prints to stdout (capture it); status to stderr.
ID=$(adg add --title "Use bounded VecDeque for stdout capture" --model docs/decisions)

#    Need a specific ID (e.g. a plan that committed to 0022)? Pass it — fails on collision:
#      ID=$(adg add --title "..." --id 22 --model docs/decisions)
#    Preview the filename before the ADR exists:  adg slug "Use bounded VecDeque ..."

# 2. Fill the whole body via stdin. Required sections: Context and Problem Statement,
#    Considered Options, Decision Outcome (adg rejects a body missing any). The H1 is
#    optional — adg prepends it from the stored title. Leave Decision Outcome as the {...}
#    placeholder; `decide` writes the real outcome line.
adg edit --id "$ID" --from-stdin --model docs/decisions <<'EOF'
## Context and Problem Statement

The subprocess can emit unbounded stdout; we need the tail without risking OOM.

## Considered Options

* Streaming reader into a bounded VecDeque
* Unbounded read_to_end with post-hoc tail slicing

## Decision Outcome

{...}
EOF

# 3. Decide. --option takes a 1-based index OR the exact option text; prefer the index
#    for options with quotes/backticks. Status flips to accepted.
adg decide --id "$ID" --option 1 --rationale "bounds peak memory, not just retained bytes" --model docs/decisions

# 4. Validate (the pre-commit hook runs this too).
adg validate --model docs/decisions
```

### Two failure modes to know up front

- **`decide` refuses to overwrite an *authored* Decision Outcome.** It only fills a
  *placeholder* outcome — an empty section, `{...}`, or the unedited `adg add` template.
  If you wrote real prose there (even the literal word "placeholder"), `decide` blocks;
  `--force` overrides and also replaces any nested `### Consequences`. So in step 2, leave
  `{...}` and let `decide` do the writing.
- **`validate` rejects an *accepted* ADR that still has scaffolding** — a leftover
  `{...}` / `{option 1}` / `{driver 1}` token, or an empty required section. Proposed
  stubs are exempt (they're in progress). Fix the content; never bypass the hook.

## Amend, supersede, annotate

```bash
adg edit --id 0042 --from-stdin --model docs/decisions < body.md            # re-edit a proposed ADR
adg edit --id 0042 --from-stdin --force --model docs/decisions < body.md    # non-proposed needs --force
adg comment --id 0042 --text "follow-up: see commit abc123" --model docs/decisions   # frontmatter-stored
adg tag --id 0042 --tag worker-protocol --model docs/decisions
adg link --from 0042 --to 0050 --model docs/decisions                       # within-model precedence
adg supersede --new 0050 --old 0042 --rationale "approach evolved" --model docs/decisions  # bidirectional
adg revise --id 0042 --model docs/decisions                                 # decided ADR -> fresh proposed copy
```

## Conventions

- **Titles carry the load.** There is no summary field, and `adg list` is the lookup
  surface — a title should encode the WHAT and often the WHY in ~10–15 words. Not a question.
- **Comments are frontmatter.** The `## Comments` body section regenerates on every save;
  never hand-edit it — use `adg comment`.
- **Validate early.** Don't wait for the hook; run `adg validate` after mutating.

## Optional ergonomics

`assets/scripts/adr` is a thin wrapper that hardcodes the model and takes positional args
(`adr new`, `adr edit <id> < body`, `adr decide <id> <opt> [<r>] [--force]`). Install it
per-repo for token savings; drop to bare `adg` for any flag it doesn't expose. It honours
`ADR_MODEL` (default `docs/decisions`).

## Reference files

- `references/adg-reference.md` — full `adg` command surface and notes
- `references/form-factor.md` — the small-form spec and exactly what the validator enforces
- `assets/madr-template.md` — the minimal body to author against
- `assets/scripts/adr`, `assets/githooks/pre-commit` — optional wrapper and enforcement hook
