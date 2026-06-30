---
status: accepted
date: "2026-03-20"
category: Python architecture
source: reported to Eyra 2026-03-20; implemented in fork main.py
priority: invariant
applies_to:
    - "**/*.py"
tags:
    - pii-safety
    - error-handling
---

# The ScriptWrapper exception handler is a PII safety boundary

## Decision

`ScriptWrapper.send()` catches all exceptions before they reach Pyodide and routes them through the
consent-gated `error_flow()`. This handler is a PII safety boundary, not mere error handling: it
stops participant data embedded in Python exception messages from reaching the JS logging path,
which forwards unsanitized to the host platform.

## Guidance

- The `except Exception` handler in `ScriptWrapper.send()` must not be removed or narrowed without
  replacing the PII protection it provides.
- New extraction/helper code may raise freely — the boundary catches it; do not add a competing
  JS-side log path that bypasses consent.
- The participant sees the error and chooses whether to donate the traceback.

## Checks

- Confirm `ScriptWrapper.send()` still wraps execution in an `except Exception` that routes to `error_flow()`.
- grep for new `bridge.sendLogs` / JS log paths that could forward exception text without consent.

## Why

Python exception messages routinely contain the data that caused the error (`ValueError` with the
offending input, `KeyError` with the key) — that *is* participant data. The JS path
(worker_engine → LogForwarder → `bridge.sendLogs`) forwards it to mono with no consent mechanism, and
researchers cannot prevent it through careful coding. JS-side sanitization (fragile regex stripping)
and disabling JS logging entirely (loses all crash diagnostics) were the weaker alternatives.
