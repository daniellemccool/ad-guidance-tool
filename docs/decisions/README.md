# Architectural decisions

This index is generated from the ADR frontmatter — do not edit by hand.
Load the ADR(s) whose filename matches the area you are touching.

## Index

### Architecture

- [0001 — Route matching is the single shared routing kernel](./0001-routematch-is-the-shared-routing-kernel.md)
- [0002 — One canonical compiled lean renderer, shared by CLI, hook, CI, and tools](./0002-one-canonical-compiled-lean-renderer-shared-by-cli-hook-ci-and-tools.md)
- [0003 — Stable commands run through the Clean Architecture stack](./0003-stable-commands-use-the-clean-architecture-stack.md)
- [0010 — Executable checks are grep assertions, not commands](./0010-executable-checks-are-grep-assertions-not-commands.md)
- [0011 — adg makes no LLM calls; ADR review runs in a Claude Code subagent](./0011-adg-makes-no-llm-calls-review-runs-in-a-subagent.md)

### ADR formats

- [0004 — MADR and lean are separate user-facing formats, not implementation islands](./0004-madr-and-lean-are-separate-user-facing-formats.md)
- [0006 — Parser/renderer round-trip stability is an invariant](./0006-parser-renderer-round-trip-is-an-invariant.md)

### Validation

- [0005 — Validation has enforcement tiers](./0005-validation-has-enforcement-tiers.md)

### Decision model

- [0007 — Supersedes, amends, and links are distinct relationships](./0007-supersedes-amends-and-links-are-distinct.md)
- [0009 — ADR files are the only source of truth — no index or cache](./0009-adr-files-are-the-only-source-of-truth.md)

### CLI conventions

- [0008 — Commands route machine output to stdout, status to stderr](./0008-route-machine-output-to-stdout-status-to-stderr.md)
