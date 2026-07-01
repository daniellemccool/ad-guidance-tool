# ADG — Architectural Decision Guidance

A command-line tool for managing **Architectural Decision Records (ADRs)** and compiling them into
architecture-context **briefs** that a coding agent reads before it edits. ADRs live in a *model* —
a directory of `NNNN-slug.md` files — and `adg` creates, edits, validates, links, and searches them.

`adg` speaks two record formats over one model directory:

- **MADR** — durable decision records ([MADR 4.0](https://adr.github.io/madr/): Context / Considered
  Options / Decision Outcome) with a `decide` / `supersede` / `revise` lifecycle. The archive:
  *what was decided, and why.*
- **Lean** — one-screen Decision/Guidance records with routing frontmatter (`applies_to`, `excludes`,
  `forbids`, `companions`). The active constraint: *what rule governs my next edit, and how do I know
  if I've violated it?* `adg` compiles the records that match a change into a brief and injects it at
  edit time via a Claude Code hook.

This is a fork of [adr/ad-guidance-tool](https://github.com/adr/ad-guidance-tool) — see
[Fork rationale](#fork-rationale) for what differs.

## Fork rationale

The upstream tool managed a single custom-Markdown format with HTML anchor tags and a sidecar
`index.yaml`. This fork made two moves:

1. **MADR on disk, no index.** Files are ordinary MADR records that round-trip through `parse →
   render`; metadata (tags, custom links, comments, supersession) lives in YAML frontmatter; ADR files
   are the only source of truth (`index.yaml` and `adg rebuild` are gone). The upstream comment
   data-loss bug is fixed — comment text is preserved verbatim in frontmatter and re-rendered into the
   body on every save. These departures are recorded in [`docs/fork-design/`](./docs/fork-design/) as a
   self-hosted MADR model.

2. **From ADR management to architecture-context compilation.** A second, *lean* format optimizes for
   agent consumption: small Decision/Guidance records with glob-based routing that `adg` compiles into a
   per-change brief and injects via a Claude Code hook. The tool's own current decisions live in
   [`docs/decisions/`](./docs/decisions/) — themselves lean records.

## Install

**Prebuilt binary (recommended).** Install the latest release into `~/.local/bin` — no Go toolchain:

```sh
curl -fsSL https://raw.githubusercontent.com/daniellemccool/ad-guidance-tool/main/install.sh | sh
```

Pin a version with `ADG_VERSION=v1.1.0` or change the location with `ADG_INSTALL_DIR`. Binaries for
macOS/Linux/Windows (amd64/arm64) are on the
[Releases](https://github.com/daniellemccool/ad-guidance-tool/releases) page.

**From source** (needs Go 1.24+):

```sh
git clone https://github.com/daniellemccool/ad-guidance-tool.git
cd ad-guidance-tool
go build           # produces ./adg
# or:
go install ./...   # installs to $GOBIN
```

## Choosing a format

One model directory, one chronological `NNNN` sequence. Pick the format by what the record is *for*:

| | MADR | Lean |
|---|---|---|
| **Purpose** | Durable record of a decision and its alternatives | Active rule consulted before an edit |
| **Body** | Context / Considered Options / Decision Outcome | Decision / Guidance / Why (+ optional Checks) |
| **Lifecycle** | `add` → `edit` → `decide` → `supersede` / `revise` | `lean new` → validate → route via the brief |
| **Routing** | none | `applies_to` / `excludes` / `forbids` / `companions` globs |
| **Consumed by** | humans, archaeology | the compiled brief + the PreToolUse hook |

The two formats are deliberately separate user-facing surfaces, not implementation variants
(see [ADR-0004](./docs/decisions/0004-madr-and-lean-are-separate-user-facing-formats.md)).

---

## MADR format

Each ADR is `NNNN-slug.md` inside a model directory:

```markdown
---
status: proposed
date: "2026-05-18"
tags: [data, infra]
---

# Decision title

## Context and Problem Statement

What problem are we solving?

## Considered Options

* Option A
* Option B

## Decision Outcome

Chosen option: "Option A", because reasons.

### Consequences

* Good, because ...
* Bad, because ...
```

After `adg comment`, the frontmatter grows a `comments:` list and a `## Comments` body section is
regenerated from it. The `links:` map holds custom link tags; `supersedes:` lists predecessor IDs.

### Command reference

| Command | Purpose |
|---|---|
| `init <model>` | Create a new model directory. |
| `add --title <title> [--id N] [--model <dir>]` | Create a new ADR with the next ID (or `--id`, which fails on collision). Refuses titles that slugify to empty. Prints the bare ID to stdout. |
| `slug "<title>"` | Print the slug `add` would produce, without creating anything. |
| `edit --id <id> --from-stdin\|--from-file <path> [--force]` | Replace the decision body. Status-gated: non-proposed decisions require `--force`. Append flags (`--context`/`--drivers`/`--option`) also exist for incremental edits. |
| `decide --id <id> --option <name-or-number> [--rationale <text>] [--force]` | Set status to accepted and write the MADR "Chosen option: …" line. The option must already exist in Considered Options. `--force` bypasses two guards: re-deciding an accepted ADR, and overwriting an authored Decision Outcome (incl. a nested `### Consequences`). |
| `comment --id <id> --text <text> [--author <name>]` | Append a comment; text preserved verbatim in frontmatter. |
| `supersede --new <id> --old <id> [--rationale <text>]` | Mark `new` as superseding `old`. Bidirectional; auto-promotes `new` to accepted. |
| `link --source <id> --target <id> --tag <name> [--reverse-tag <name>]` | Add a custom within-model link. |
| `revise --id <id>` | Clone a decided ADR into a fresh proposed draft. |
| `tag --id <id> --tag <name>` | Add a tag. |
| `view --id <id> [--context\|--drivers\|--options\|--outcome\|--comments]` | Print full body or selected sections. |
| `list [--filter ...]` | List decisions; `--format json\|yaml\|md\|simple`. |
| `validate [--model <dir>]` | Check MADR shape, supersession integrity, comment text. |
| `migrate [--dry-run]` | Convert upstream ADG files into MADR shape, in place (idempotent). |
| `copy` / `import` / `merge` | Clone, import into, or combine model directories. |
| `set-config` / `reset-config` | Manage `.adgconfig.yaml`. |

Run `adg <command> -h` for full flag details. For whole-body authoring, `edit --from-stdin` reads MADR
markdown from stdin; the body must include the three required sections (Context and Problem Statement,
Considered Options, Decision Outcome) or the command refuses.

---

## Lean format and governance

A lean record is one screen, optimized for the compiled brief. Required sections are `Decision`,
`Guidance`, and `Why` (the reasoning — required on every accepted record, but **record-only**: never
rendered into a brief); `Checks` is optional. Routing lives in frontmatter:

```markdown
---
status: accepted          # proposed | accepted | rejected | deprecated | superseded by ADR-NNNN | amended by ADR-NNNN
category: Extraction      # groups the generated index (not the directory layout)
priority: invariant       # invariant | default — force in the brief
applies_to:
    - port/**/*.py
excludes:
    - "**/port_helpers.py"
---

# Reject unsafe uploads before validation and extraction

## Decision

One to three sentences: what was decided.

## Guidance

- What new code must do, what review rejects, the fix path.

## Why            # required on an accepted record; record-only (never in a brief)
## Checks         # optional; grep targets rolled up into the brief
```

- A path is **governed** iff some `applies_to` glob matches it and no `excludes` glob does.
- `forbids` is negative-space scope — paths that should *not* exist; it routes the brief like
  `applies_to` but warns when it *does* match instead of when it's stale.
- `companions` are expected partner edits (e.g. the TS side of a prop) that the ADR does **not**
  govern; they surface in the brief as "related files," never routed on.
- IDs are a flat global `NNNN` across the model; `category` (not a subfolder) groups the index.
- Globs are forward-slash, repo-root-relative, doublestar (`**`). Brace globs `{a,b}` are rejected —
  write one glob per alternative.

### Command reference

| Command | Purpose |
|---|---|
| `lean new --title <t> [--status …] [--priority …] [--category …] [--applies-to <glob>] [--excludes <glob>] [--from-stdin]` | Author a lean ADR. Validates the candidate and **refuses to write on a hard failure**, so an invalid record never lands on disk. Prints the new ID. |
| `lean index [--write] [--root <tree>] [--overlaps]` | Validate the model and print/write the grouped README. `--root` scope-lints globs against the source tree (wire this into CI — the hook only routes, the index gates). `--overlaps` adds the opt-in default-vs-default hub diagnostic. |
| `lean brief [--hook] <changed-path…>` | Compile the architecture brief for the changed paths: the ADRs that govern them, grouped by force, each with Decision + Guidance + consolidated checks. `--hook` is the PreToolUse entry point. |
| `lean verify [--hook] [<changed-path…>]` | Re-validate the model and re-show the brief + its "Before you finish" footer for files changed this session. `--hook` is a Stop-hook entry point (advisory, non-blocking). |
| `lean check [<changed-path…>]` | Run the executable grep-assertion checks declared in matched ADRs' `## Checks`. |
| `lean review [<adr-file…>] [--since <ref>]` | Emit a deterministic review packet (target ADRs + lint findings) for a reviewer to judge against the rubric. `adg` makes no LLM call — review runs in a Claude Code subagent ([ADR-0011](./docs/decisions/0011-adg-makes-no-llm-calls-review-runs-in-a-subagent.md)). |

### The brief, the hooks, and CI

The same compiled-brief renderer drives the CLI, the hooks, and CI
([ADR-0002](./docs/decisions/0002-one-canonical-compiled-lean-renderer-shared-by-every-consumer.md)):

The `write-adr` plugin bundles a suite of **fail-open** hooks that route the brief across the change
lifecycle — the whole-corpus brief at `SessionStart`, invariants at `Plan`-subagent dispatch, the deduped
file-scoped brief before an edit, a staged-file brief before a commit — plus a **guard** that blocks
hand-creating an ADR record and two **agent** reviewers (code-vs-ADR compliance at commit, ADR-quality on
record change). Only two hard stops exist: a commit that stages a `forbids` violation, and hand-creating a
record; everything else advises.

- **CI** runs `adg lean index --root .` for real enforcement (stale globs, duplicate IDs, brace globs,
  leanness lints). The hooks route and advise; the index gates.

The full suite, the exact hook JSON, and a worked example model live in
[`docs/lean-example/hook-setup.md`](./docs/lean-example/hook-setup.md). Because the hooks fire only on
Claude's tool calls and are fail-open, **"no brief appeared" never means "no rule applies"** —
comprehensive enforcement is CI / review / executable checks.

---

## Claude Code plugin (ADR skills)

The [`write-adr`](./tools/adr-plugin/) plugin ships *with* `adg` so its guidance tracks the CLI in
lockstep. It provides four skills — two for *authoring* (pick the one matching a repo's format), one for
*obeying* lean briefs while changing code, and a *gateway* that routes any ADR task to the right one:

- **using-write-adr** — the gateway: broadly discoverable ("Use when ADRs come up in any way"), it routes
  ADR work to `adg` + the specific skill instead of letting the agent hand-roll it.
- **write-madr-adr** — author durable MADR records with the `decide` / `supersede` / `revise` lifecycle.
- **write-lean-adr** — author/migrate/rewrite/review lean records with routing frontmatter.
- **follow-adr-governance** — a behavior primer for obeying an injected lean brief while editing code
  (the hook and the brief do the real work).

```
/plugin marketplace add daniellemccool/ad-guidance-tool
```

The skills call `adg`, and it **rides along**: the plugin ships a `bin/adg` wrapper that Claude Code
puts on `PATH` while the plugin is enabled, fetching the prebuilt CLI that matches the plugin's version
on first use (no Go toolchain needed). The `d3i-skills` marketplace **lists** this plugin via a
`git-subdir` source pinned to a release tag — a reference to this repo, which stays the canonical
source. (Governed-repo hooks run outside the plugin's PATH and need a system `adg` — see [Install](#install).)

---

## Scripting (stdout / stderr / exit codes)

`adg` follows the usual Unix conventions, so it's safe to pipe and script:

- **stdout** carries machine-readable values: `add` / `revise` / `lean new` write the new ID; `list`,
  `view`, and `lean brief`/`index` write their rendered output.
- **stderr** carries human-readable status and all errors.
- **`--quiet`** (global) suppresses stderr status; machine values on stdout and errors on stderr still
  print.
- **Exit codes:** `0` on success; `1` on any failure including validation issues.

```sh
ID=$(adg add --title "Bounded subprocess output")   # captures 0007
adg --quiet add --title "X"                          # only the ID prints
adg validate || echo "model has problems"            # exit 1 when issues exist
```

This split is an invariant
([ADR-0008](./docs/decisions/0008-route-machine-output-to-stdout-status-to-stderr.md)).

## Config

```sh
adg set-config         # configure defaults (model path, author, etc.)
adg reset-config       # clear all values
```

Config lives at `~/.adgconfig.yaml` by default; override with `--config-path`.

## Migrating from upstream ADG

`adg migrate` converts an existing upstream model into MADR shape in place:

```sh
adg migrate --model docs/decisions --dry-run    # preview only
adg migrate --model docs/decisions               # rewrite
```

It renames `AD0001-slug.md` → `0001-slug.md`, drops the `AD`-era frontmatter, strips HTML anchors,
renames sections to MADR vocabulary, and recovers comment text into frontmatter (flagging any it can't
pair). Migrate is faithful and idempotent — it doesn't synthesize sections the source lacked, so a
legacy `status: open` ADR will report a missing Decision Outcome after migration by design.

> **Heads up:** inline body links like `[ADR-0002](AD0002-foo.md)` break after migration because the
> target filename changes. Grep your corpus for `AD\d{4}-` afterward if you used filename-based links.

## The tool's own decisions

`adg` governs itself. Its current architectural decisions are lean records in
[`docs/decisions/`](./docs/decisions/) (the routing kernel, the canonical renderer, MADR/lean
separation, enforcement tiers, round-trip stability, relationship types, stdout/stderr, no-index, …).
The earlier MADR-fork decisions are in [`docs/fork-design/`](./docs/fork-design/), and a worked lean
example model is in [`docs/lean-example/`](./docs/lean-example/).

## Contributing

The codebase follows Clean Architecture (domain → application → adapter → infrastructure). Tests use
[testify](https://github.com/stretchr/testify) and [mockery](https://github.com/vektra/mockery). For
changes:

1. Start with the use case (interactor) or domain logic.
2. Add the cobra command + presenter at the adapter layer.
3. Cover with unit tests; regenerate mocks if interfaces shift.
4. Run `go test ./...` before pushing.

Stable commands run through the full Clean Architecture stack
([ADR-0003](./docs/decisions/0003-stable-commands-use-the-clean-architecture-stack.md)). Mocks under
`mocks/` are generated by [mockery](https://github.com/vektra/mockery) v2.53.6 from `.mockery.yaml`:
`env GOBIN="$HOME/.local/bin" go install github.com/vektra/mockery/v2@v2.53.6`, then run `mockery` from
the repo root after any interface change.

## References

- [MADR](https://adr.github.io/madr/) — the durable file format this fork adopts.
- Upstream tool: [adr/ad-guidance-tool](https://github.com/adr/ad-guidance-tool).
- Original theses behind the upstream tool:
  - [Concept Alternatives for the Management of Architectural Decisions in Clean Architectures](https://eprints.ost.ch/id/eprint/1280/1/MSECS-FS24-CleanArchitectureDecisionsConceptsRS.pdf)
  - [A Command-Line Tool for Managing Recurring Architectural Decisions](https://eprints.ost.ch/id/eprint/1287/1/PA2-Raphael-Schellander.pdf)

## License

Apache License 2.0 — see [LICENSE](./LICENSE).
