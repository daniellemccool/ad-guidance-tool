---
name: write-madr-adr
description: >
  Use when recording, documenting, or revising an architecture decision in a repo whose
  ADRs use the MADR format, when capturing a design choice or a rejected alternative, when
  asked "should this be an ADR", or proactively after implementing a significant pattern
  that isn't yet recorded. For lean Decision/Guidance records with routing frontmatter use
  write-lean-adr instead. Works in any repo with a docs/decisions/ model (and bootstraps
  one if absent).
---

# write-madr-adr — record a durable decision in MADR format with `adg`

ADRs are MADR-format markdown managed by the `adg` CLI. Three rules carry the whole skill:

1. **Drive everything through `adg`. Never hand-edit ADR files.** The tool enforces the
   shape, round-trips the body, and regenerates the `## Comments` section from frontmatter
   — hand-edits are lost or rejected.
2. **One flat *active* model: `docs/decisions/`** (unless `adg set-config` points elsewhere).
   Files are `NNNN-slug.md` with bare 4-digit IDs; one chronological sequence per repo, no
   per-domain sub-models. Operate only on that active model — other ADR-like files in the tree (a
   historical record under `docs/fork-design/`, a lean model, seed/template models) are **not** your
   model; don't scan every `NNNN-*.md` in the repo. Pick this skill vs write-lean-adr by the active
   model's format (full MADR sections here; Decision/Guidance + routing frontmatter for lean).
3. **Keep ADRs small.** See `references/form-factor.md`. The validator enforces the floor
   (no scaffolding, no empty sections); you hold the ceiling — don't pad.

Examples use bare `adg` so they work in any repo. `--model docs/decisions` is shown
explicitly; once you've run `adg set-config` it can be omitted. An optional repo-local
wrapper removes the boilerplate — see "Optional ergonomics".

## Setup (once per repo)

```bash
# Bootstrap the model if docs/decisions/ doesn't exist yet:
adg init docs/decisions

# Install the enforcement hook so `adg validate` gates every commit. Copy this skill's
# assets/githooks/pre-commit into the repo — substitute this skill's actual install
# directory for <skill-dir> (the asset path is not relative to the target repo):
mkdir -p .githooks
cp <skill-dir>/assets/githooks/pre-commit .githooks/pre-commit
chmod +x .githooks/pre-commit
git config core.hooksPath .githooks
```

The pre-commit hook runs `adg` in git's hook context, outside Claude Code — so it needs a **system**
`adg` on `PATH` (install it with the `curl … install.sh` one-liner from the main README), not just the
plugin's bundled `bin/adg` wrapper.

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
`ADR_MODEL` (default `docs/decisions`). Like the git hook, it calls bare `adg`, so it needs a
system `adg` on `PATH` (see the main README's install section).

## Reference files

- `references/adg-reference.md` — full `adg` command surface and notes
- `references/form-factor.md` — the small-form spec and exactly what the validator enforces
- `assets/madr-template.md` — the minimal body to author against
- `assets/scripts/adr`, `assets/githooks/pre-commit` — optional wrapper and enforcement hook
