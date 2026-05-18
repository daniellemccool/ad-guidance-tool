---
status: proposed
tags:
    - integration
    - external-system
    - interfaces
    - clean-architecture
links:
    precedes: []
    succeeds: []
---

# Integrate external systems

## Context and Problem Statement
How should an external system be integrated in a way that maintains the independence of core business logic and upholds Clean Architecture principles?

## Considered Options
* Use an interface-based adapter pattern to wrap the external system and inject it via dependency inversion.
* Introduce an intermediary service layer that mediates between core logic and the external system.
* Allow direct calls to the external system from the use case layer, with minimal abstraction.

## Decision Drivers
- Degree of decoupling between core logic and the external system
- Complexity of integration and future maintainability
- Adherence to dependency inversion and interface segregation
- Testability of components dependent on external system behavior
