# Architectural decisions

This index is generated from the ADR frontmatter — do not edit by hand.
Load the ADR(s) whose filename matches the area you are touching.

## Index

### Architecture

- [0001 — Route matching is the single shared routing kernel](./0001-routematch-is-the-shared-routing-kernel.md)
- [0002 — One canonical compiled lean renderer shared by every consumer](./0002-one-canonical-compiled-lean-renderer-shared-by-every-consumer.md)
- ~~[0003 — Stable commands run through the Clean Architecture stack](./0003-stable-commands-use-the-clean-architecture-stack.md)~~ — *superseded by ADR 0017*
- [0010 — Executable checks are grep assertions, not commands](./0010-executable-checks-are-grep-assertions-not-commands.md)
- [0011 — adg makes no LLM calls; ADR review runs in a Claude Code subagent](./0011-adg-makes-no-llm-calls-review-runs-in-a-subagent.md)
- [0012 — Routing honors record lifecycle; terminal records govern nothing](./0012-routing-honors-record-lifecycle-terminal-records-govern-nothing.md)
- [0015 — Per-project config lives in .adg.yaml, read at the command edge, and never changes the brief](./0015-per-project-config-lives-in-adg-yaml-read-at-the-command-edge-and-never-changes-the-brief.md) *(amended by ADR 0021)*
- [0017 — Commands are thin cobra adapters over the domain](./0017-commands-are-thin-cobra-adapters-over-the-domain.md)
- [0018 — adg keeps no global user state; the model path is the docs/decisions convention](./0018-adg-keeps-no-global-user-state-the-model-path-is-the-docs-decisions-convention.md)
- [0019 — A fresh context always receives its first brief](./0019-a-fresh-context-always-receives-its-first-brief.md)
- [0020 — Agent hooks conclude with structured output and never trip a permission prompt](./0020-agent-hooks-conclude-with-structured-output-and-never-trip-a-permission-prompt.md)
- [0021 — Inject the session-open digest, capped at 2 KB](./0021-inject-the-session-open-digest-capped-at-2-kb.md)

### ADR formats

- ~~[0004 — MADR and lean are separate user-facing formats, not implementation islands](./0004-madr-and-lean-are-separate-user-facing-formats.md)~~ — *superseded by ADR 0016*
- [0006 — Parser/renderer round-trip stability is an invariant](./0006-parser-renderer-round-trip-is-an-invariant.md)
- [0014 — A lean record's reasoning is a required section, co-equal with Decision and Guidance](./0014-a-lean-record-s-reasoning-is-a-required-section-co-equal-with-decision-and-guidance.md)
- [0016 — Lean is the sole user-facing adg ADR format; madr is shared parsing plumbing](./0016-lean-is-the-sole-user-facing-adg-adr-format-madr-is-shared-parsing-plumbing.md)

### Validation

- [0005 — Validation has enforcement tiers](./0005-validation-has-enforcement-tiers.md)

### Decision model

- [0007 — Supersedes, amends, and links are distinct relationships](./0007-supersedes-amends-and-links-are-distinct.md)
- [0009 — ADR files are the only source of truth — no index or cache](./0009-adr-files-are-the-only-source-of-truth.md)
- [0022 — Write ADR titles as short imperative instructions](./0022-write-adr-titles-as-short-imperative-instructions.md)

### CLI conventions

- [0008 — Commands route machine output to stdout, status to stderr](./0008-route-machine-output-to-stdout-status-to-stderr.md)

### Release

- [0013 — The marketplace tracks main, so a plugin.json version bump must ship with its tag and release](./0013-the-marketplace-tracks-main-so-a-plugin-json-version-bump-must-ship-with-its-tag-and-release.md)
