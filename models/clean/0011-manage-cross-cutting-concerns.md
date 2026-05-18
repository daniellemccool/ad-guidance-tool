---
status: proposed
tags:
    - cross-cutting
    - logging
    - monitoring
    - auth
    - clean-architecture
links:
    precedes:
        - "0012"
    succeeds:
        - "0010"
---

# Manage cross cutting concerns

## Context and Problem Statement
How should cross-cutting concerns such as logging, monitoring, or authentication be handled across layers while preserving the architectural integrity of the system?

## Considered Options
* Use middleware or decorator patterns in the outer layers (e.g., interface adapters) to inject cross-cutting logic.
* Apply aspect-oriented techniques (e.g., interceptors, annotations) to weave in concerns dynamically.
* Embed cross-cutting logic directly into use cases and infrastructure code where needed.

## Decision Drivers
- Degree of decoupling and modularity
- Ease of applying concerns uniformly across layers
- Testability of use cases without concern-side effects
- Visibility and traceability of system behavior (e.g., logging context)
- Performance and overhead introduced by concern injection
