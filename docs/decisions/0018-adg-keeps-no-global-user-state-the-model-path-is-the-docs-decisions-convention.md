---
status: accepted
date: "2026-07-09"
source: sub-project C global-config teardown; ~/.adgconfig.yaml and set-config/reset-config removed
category: Architecture
applies_to:
    - internal/adapter/command/configutil.go
    - cmd/root.go
priority: invariant
forbids:
    - internal/infrastructure/config/**
---

# adg keeps no global user state; the model path is the docs/decisions convention

## Decision

`adg` reads no per-user state: there is no `~/.adgconfig.yaml` and no global config service. The lean model lives at `docs/decisions` by convention, and the only override is the per-invocation `--model` flag, resolved by one pure helper at the command edge (`ResolveModelPath`).

## Guidance

- A new setting is either a per-invocation flag or a per-project key in `<model-root>/.adg.yaml` read at the command edge; review rejects any config file under `$HOME`, env-var fallback, or revival of `internal/infrastructure/config/` — the fix path is a flag or a `.adg.yaml` key.
- Model resolution stays flag-or-constant in `configutil.go`: no filesystem probing, no upward search, no prompting. A bare command in a repo without `docs/decisions` fails at load with the resolved path and the `--model` hint (`ModelLoadHint`).
- Startup must not depend on any per-user file: nothing in `cmd/root.go` may construct a service whose load failure can exit before a subcommand runs.

## Why

Weakening this reintroduces hidden per-machine drift that neither the brief pipeline nor CI can see — an invocation's behavior would again depend on unversioned local files instead of being fully determined by the repo and its argv.
