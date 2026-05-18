---
status: proposed
tags:
    - interfaces
    - dependency-inversion
    - use-cases
    - clean-architecture
links:
    precedes:
        - "0010"
    succeeds:
        - "0003"
---

# Define usecase interfaces

## Context and Problem Statement
How should the interface contracts between the use case layer and the outer layers (e.g., interface adapters, frameworks) be defined to support dependency inversion and layer independence?

## Considered Options
* Let the use case layer define abstract interfaces that are implemented by external layers (e.g., `UserRepository` interface).
* Define interfaces in a shared contract layer decoupled from both use case and infrastructure.
* Allow outer layers to define concrete APIs and let use cases depend directly on them (inversion not enforced).

## Decision Drivers
- Clarity of architectural boundaries and ownership
- Degree of adherence to dependency inversion principle
- Ease of mocking or substituting components in tests
- Maintainability and discoverability of contracts across modules
