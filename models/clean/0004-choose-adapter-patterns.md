---
status: proposed
tags:
    - pattern
    - translation
    - interface-adapter
    - clean-architecture
links:
    precedes:
        - "0005"
    succeeds:
        - "0003"
---

# Choose adapter patterns

## Context and Problem Statement
Which pattern should be used to handle data translation between a use case and an external system, in order to avoid tight coupling and maintain separation of concerns?

## Considered Options
* Use the Data Mapper pattern to convert between internal domain models and external representations.
* Apply the Adapter (or Anti-Corruption Layer) pattern to fully isolate external models behind interface boundaries.
* Embed translation logic directly within the use case or controller to reduce indirection.

## Decision Drivers
- Degree of decoupling from external systems
- Maintainability and testability of translation logic
- Complexity introduced by the pattern
- Reusability of transformation logic across use cases
