---
status: accepted
date: "2026-06-29"
category: CLI conventions
source: docs/fork-design/0004-route-machine-values-to-stdout-and-human-status-to-stderr.md
priority: invariant
applies_to:
    - cmd/**/*.go
    - internal/adapter/command/**/*.go
    - internal/adapter/printer/**/*.go
---

# Commands route machine output to stdout, status to stderr

## Decision

A command writes machine-consumable values (IDs, generated content, the compiled brief/index) to
**stdout**, and human-facing status, prompts, progress, and errors to **stderr**. Validation failure
exits non-zero. This keeps `adg` output composable — `ID=$(adg add ...)` captures only the ID.

## Guidance

- Write the primary machine value to `cmd.OutOrStdout()`; write status, errors, and issue lists to
  `cmd.ErrOrStderr()`. Never interleave human chatter into stdout.
- A command that can fail validation sets `SilenceErrors` / `SilenceUsage` and returns a sentinel error
  so main exits non-zero with issues already printed to stderr (see `adg lean index` / `adg validate`).
- New output in a command must pick the right stream — this is the easiest convention to violate by
  accident (a stray `fmt.Println`).

## Why

This looks like a style preference but it protects automation: `ID=$(adg add ...)` works only because
the ID is the *sole* thing on stdout while status goes to stderr. A single stray status line on stdout
silently corrupts every script that captures command output — so this is an invariant, not a nicety.

## Checks

- A new command's stdout writes are only the machine value; status/error text goes to stderr (grep the
  command for `fmt.Print`/`Println` that isn't the machine value).
