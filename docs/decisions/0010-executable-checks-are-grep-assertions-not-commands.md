---
status: accepted
date: "2026-06-29"
category: Architecture
applies_to:
    - internal/domain/decision/lean/check.go
    - internal/adapter/command/lean/check.go
priority: invariant
checks:
    - desc: the grep-assertion check runner executes no commands
      grep: exec\.Command
      in:
          - internal/domain/decision/lean/check.go
      expect: absent
---

# Executable checks are grep assertions, not commands

## Decision

An ADR's executable `checks` are grep assertions — a regexp that must be absent (default) or present within a glob scope — run by `adg lean check`. ADRs do not execute arbitrary commands.

## Guidance

- Express an automatable check as a frontmatter `checks` entry (`grep` / `in` / `except` / `expect`); keep non-automatable checks as prose in `## Checks`.
- Do not add a command- or script-execution check kind; a tree-derived grep assertion covers the common "this pattern must not appear outside X" case.
- Run `adg lean check` in CI for code-level enforcement; the brief footer points contributors at it.

## Why

Running commands declared in version-controlled ADRs would turn the brief/check pipeline into an arbitrary-code-execution surface — any committed ADR could run anything in CI or a contributor tree. Grep assertions are inert and cover the common case; command execution is a deliberate, separate decision if ever needed.
