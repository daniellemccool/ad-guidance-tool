---
status: accepted
date: "2026-05-18"
comments:
    - author: Danielle McCool
      date: "2026-05-18 18:09:42"
      text: marked decision as decided
---

# Make supersede a first-class bidirectional command

## Context and Problem Statement

Supersession is the most consequential ADR relationship: it says "this decision replaces that one." The upstream tool had no first-class supersede operation. Users wrote `precedes`/`succeeds` links by hand and updated status strings independently. The two could drift — a link present without the status update, or a status string referencing an ADR that never claimed to supersede it. Detecting drift required reading two files. The link command refused tag=`supersedes`/`superseded-by` with an error message pointing at a non-existent `adg supersede`.

## Decision Drivers

* Supersession is bidirectional by definition; one operation should write both ends
* The validator must detect drift — but better is not having drift in the first place
* Auto-promote the new decision's status: superseding *is* the act of accepting a replacement
* Cycles and self-supersession are bugs, not features

## Considered Options

* Keep hand-written links + status; provide a validator-only check for drift
* Add a dedicated `adg supersede --new <id> --old <id>` that writes both ends in one transaction
* Single source: only the new decision's `Supersedes` list; treat the old's status as derived/computed at read time

## Decision Outcome

Chosen option: "Add a dedicated `adg supersede --new <id> --old <id>` that writes both ends in one transaction", because the bidirectional update is one logical operation; the validator's forward+reverse integrity checks still defend against external edits that drift the two ends apart.

## Comments

* **2026-05-18 18:09:42 — @Danielle McCool:** marked decision as decided
