---
status: accepted
date: "2026-03-17"
category: Extraction
priority: default
applies_to:
    - "**/flow_builder.py"
    - "**/uploads.py"
tags:
    - safety
    - uploads
---

# Reject unsafe uploads before DDP validation and extraction

## Decision

FlowBuilder enforces upload safety before DDP validation: materialize → safety check → validate →
extract. Safety checks are per-upload concerns that apply to every platform, so they belong in
FlowBuilder (which owns the per-platform flow), not in study-level `script.py`.

## Guidance

- `helpers/uploads.py` provides `check_file_safety(path)`, which raises `FileTooLargeError` (>2GB, OOM
  risk in the Pyodide worker) or `ChunkedExportError` (exactly 2GB incomplete export).
- FlowBuilder catches these and renders a safety error page via `port_helpers` before any extraction work.
- DDP validation may assume a structurally safe file — it never runs on an unsafe upload.

## Checks

- Confirm FlowBuilder calls `check_file_safety(path)` before any DDP validation or extraction call.
- Search for extraction or validation invoked ahead of the safety check in any per-platform flow.
