# ADG (MADR-native fork)

A command-line tool for managing **Architectural Decision Records (ADRs)** in [MADR 4.0](https://adr.github.io/madr/) format. ADRs are grouped into *models* (a directory of files), and the tool helps you create, edit, link, validate, and search them.

This is a fork of [adr/ad-guidance-tool](https://github.com/adr/ad-guidance-tool) — see [Fork rationale](#fork-rationale) for what differs.

## Fork rationale

The upstream tool used a custom Markdown layout with HTML anchor tags and a sidecar `index.yaml`. This fork:

- **Adopts MADR 4.0 as the on-disk format.** Files look like ordinary MADR ADRs and round-trip through `parse → render`.
- **Stores metadata in YAML frontmatter.** Tags, custom links, comments, supersession, and a `legacy-outcome` flag are first-class fields. The body is a projection of frontmatter for sections the tool regenerates (e.g. `## Comments`).
- **Drops `index.yaml`.** ADR files are the only source of truth; `adg rebuild` is gone.
- **Fixes the comment data-loss bug.** Upstream wrote a placeholder count where the comment text should be. Here, comment text is preserved verbatim in frontmatter and re-rendered into the body on every save.
- **Rewrites `adg validate`** in MADR terms. New checks include MADR-section presence, status vocabulary, and bidirectional supersession integrity.
- **Renames `adg view` section flags** to MADR vocabulary: `--context`, `--drivers`, `--options`, `--outcome`, `--comments`.

Stdout/stderr split (machine values on stdout, status on stderr), `adg edit --from-stdin`, `adg supersede` as first-class, and `adg migrate` for legacy ADG → MADR conversion are tracked in follow-up PRs.

## Install

Build from source:

```sh
git clone https://github.com/daniellemccool/ad-guidance-tool.git
cd ad-guidance-tool
go build           # produces ./adg
# or:
go install ./...   # installs to $GOBIN
```

Go 1.22+ is required.

## File format

Each ADR is `NNNN-slug.md` inside a model directory. Example:

```markdown
---
status: proposed
date: "2026-05-18"
tags: [data, infra]
---

# Decision title

## Context and Problem Statement

What problem are we solving?

## Decision Drivers

* {driver 1}
* {driver 2}

## Considered Options

* Option A
* Option B

## Decision Outcome

Chosen option: "Option A", because reasons.

### Consequences

* Good, because ...
* Bad, because ...
```

After `adg comment`, the frontmatter grows a `comments:` list and a `## Comments` body section is regenerated from it. The `links:` map holds custom link tags; `supersedes:` lists predecessor IDs.

## Command reference

| Command | Purpose |
|---|---|
| `init <model>` | Create a new model directory. |
| `add --title <title> [--model <dir>]` | Create a new ADR with the next ID. Refuses titles that slugify to empty. |
| `edit --id <id> [--context ... \| --drivers ... \| --option ...]` | Append to a section. |
| `edit --id <id> --from-stdin\|--from-file <path> [--force]` | Replace the decision body. Status-gated: non-proposed decisions require `--force`. |
| `comment --id <id> --author <name> --text <text>` | Append a comment. Text is preserved verbatim. |
| `decide --id <id> --option <name-or-number> [--rationale <text>]` | Set status to accepted and write the MADR "Chosen option: ..." line. |
| `link --source <id> --target <id> --tag <name> [--reverse-tag <name>]` | Add a custom link. Supersession has its own command — see below. |
| `supersede --new <id> --old <id> [--rationale <text>]` | Mark `new` as superseding `old`. Bidirectional: writes `Supersedes` list on new and `superseded by ADR-N` status on old. Auto-promotes new to `accepted`. |
| `migrate [--dry-run]` | Convert upstream ADG files (`AD\d{4}-slug.md` with HTML anchors and `status: open\|decided`) into MADR shape. In-place; idempotent. |
| `tag --id <id> --tag <name>` | Add a tag. |
| `revise --id <id>` | Clone the decision into a new draft. |
| `view --id <id> [--context\|--drivers\|--options\|--outcome\|--comments]` | Print full body or selected sections. |
| `list [--filter ...]` | List decisions; `--format json\|yaml\|md\|simple`. |
| `validate [--model <dir>]` | Check MADR shape, supersession integrity, comment text. |
| `copy --model <src> --target <dst>` | Clone a model directory (filterable). |
| `import --source <dir> --target <dir>` | Import filtered decisions into an existing model. |
| `merge --model-a <a> --model-b <b> --target <dst>` | Combine two models into a fresh one. |
| `set-config`, `reset-config` | Manage `.adgconfig.yaml`. |

Run `adg <command> -h` for full flag details.

## LLM-friendly editing

For wholesale edits (e.g. an LLM rewrites a draft from scratch), use replace mode:

```sh
adg edit --id 0001 --from-stdin <<'EOF'
# Renamed decision

## Context and Problem Statement

...
EOF
```

Rules:

- The body must parse as MADR and include the three required sections: `Context and Problem Statement`, `Considered Options`, `Decision Outcome`. Otherwise the command refuses and exits non-zero.
- If the input has an `H1`, it overwrites the decision's title; the file is renamed on disk to match the new slug.
- The `## Comments` section in the input is ignored — comments are frontmatter and are regenerated from there on every save.
- A decision whose status is anything other than `proposed` requires `--force`. The intent: accepted/rejected/superseded ADRs are part of the historical record; replacing their body should be deliberate.

`--from-file <path>` reads the same content from a file instead of stdin.

The append flags (`--context`, `--drivers`, `--option`) still exist for incremental edits. Append and replace modes are mutually exclusive on a single invocation.

## Scripting (stdout / stderr / exit codes)

`adg` follows the usual Unix conventions so it's safe to pipe and script:

- **stdout** carries machine-readable values. `add` writes the new ID (one per line for multi-add); `revise` writes the new ID; `list` and `view` write their rendered output.
- **stderr** carries human-readable status (`Decision X (0001) added successfully.`) and all errors.
- **`--quiet`** (global flag) suppresses stderr status messages. Machine values on stdout still flow, and errors on stderr still print.

That means:

```sh
ID=$(adg add --title "Bounded subprocess output")   # captures 0007
adg --quiet add --title "X"                          # only the ID prints, nothing else
adg validate || echo "model has problems"            # exit 1 when issues exist
```

**Exit codes:** `0` on success; `1` on any failure including validation issues. The validate command prints the issue list to stderr; if you only care about the exit code, redirect with `2>/dev/null`.

## Validation rules

`adg validate` reports per-decision issues across:

1. Filename matches `NNNN-slug.md`.
2. H1 title present.
3. Required MADR sections present: `Context and Problem Statement`, `Considered Options`, `Decision Outcome`.
4. `Considered Options` has at least one bullet.
5. When status is `accepted` (and `legacy-outcome: false`), the Decision Outcome contains `Chosen option: "X"` with X appearing in Considered Options.
6. Status matches MADR vocabulary: `proposed`, `rejected`, `accepted`, `deprecated`, or `superseded by ADR-NNNN`.
7. Supersession forward integrity: `superseded by ADR-X` implies ADR-X exists and lists self in its `supersedes:`.
8. Supersession reverse integrity: every `supersedes:` entry points to an ADR whose status references self.
9. Comment text is non-empty and not purely numeric (defends against the legacy placeholder regression).

## Config

```sh
adg set-config         # configure defaults (model path, author, etc.)
adg reset-config       # clear all values
```

Config lives at `~/.adgconfig.yaml` by default; override with `--config-path`.

## Fork design decisions

The fork's own architectural decisions live in [`docs/fork-design/`](./docs/fork-design/) as a self-hosted MADR model. Each ADR documents a deliberate departure from the upstream tool: MADR-on-disk, no `index.yaml`, comments-in-frontmatter, stdout/stderr split, first-class supersede, replace-mode edit. Read them in order if you want the rationale behind everything in this fork.

## Migrating from upstream ADG

If you have an existing model written by the upstream tool, `adg migrate` converts it in place:

```sh
adg migrate --model docs/decisions --dry-run    # preview only
adg migrate --model docs/decisions               # rewrite
```

What changes per file:

- Filename: `AD0001-slug.md` → `0001-slug.md` (drops the `AD` prefix).
- Frontmatter: drops `adr_id`; drops the slug-style `title`; maps `status: open` → `proposed`, `status: decided` → `accepted` (with `legacy-outcome: true` to bypass the strict Chosen-option check); preserves `tags`, `links`, and `comments`.
- Body: strips every `<a name="..."></a>` HTML anchor; renames `Question` → `Context and Problem Statement`, `Options` → `Considered Options` (numbered → bulletized), `Criteria` → `Decision Drivers`, `Outcome` → `Decision Outcome`; removes the `## Comments` H2 since comments are regenerated from frontmatter on every save.
- Comments: best-effort recovery — the §A.1 bug stored placeholder indices in frontmatter while the real prose lived in body anchor blocks. Migrate pairs them by index and pulls the text into `Comments[N].Text`. If an entry can't be paired, a placeholder `(unrecoverable: legacy comment placeholder "N")` is written so `adg validate` flags it for manual repair.

Migrate is faithful: it doesn't synthesize sections the source didn't have. Legacy ADRs in `status: open` had no `## Outcome` section, so after migration they'll report `missing required section: Decision Outcome` from `adg validate` — that's the design. Add the section by hand or via `adg edit --from-stdin`.

`adg migrate` is idempotent: running it again does nothing because already-migrated files don't match the legacy markers.

> **Heads up:** inline body links like `[ADR-0002](AD0002-foo.md)` break after migration because the target filename changes. `adg migrate` doesn't rewrite link targets — grep your corpus for `AD\d{4}-` after migration if you used filename-based internal links.

## Contributing

The codebase follows Clean Architecture (domain → application → adapter → infrastructure). Tests use [testify](https://github.com/stretchr/testify) and [mockery](https://github.com/vektra/mockery). For changes:

1. Start with the use case (interactor) or domain logic.
2. Add cobra command + presenter at the adapter layer.
3. Cover with unit tests; regenerate mocks if interfaces shift.
4. Run `go test ./...` before pushing.

Open an issue or PR against this fork.

## References

- [MADR](https://adr.github.io/madr/) — the file format this fork adopts.
- Upstream tool: [adr/ad-guidance-tool](https://github.com/adr/ad-guidance-tool).
- Original theses behind the upstream tool:
  - [Concept Alternatives for the Management of Architectural Decisions in Clean Architectures](https://eprints.ost.ch/id/eprint/1280/1/MSECS-FS24-CleanArchitectureDecisionsConceptsRS.pdf)
  - [A Command-Line Tool for Managing Recurring Architectural Decisions](https://eprints.ost.ch/id/eprint/1287/1/PA2-Raphael-Schellander.pdf)

## License

Apache License 2.0 — see [LICENSE](./LICENSE).
