---
status: proposed
tags:
    - entities
    - domain-model
    - business-rules
    - clean-architecture
links:
    precedes:
        - "0003"
    succeeds:
        - "0001"
---

# Decide entity boundaries

## Context and Problem Statement
How should the boundaries of core entities be defined to encapsulate business rules in a stable and reusable way?

## Considered Options
* Define dedicated domain entities representing key business concepts, independent of frameworks or infrastructure.
* Use simple data structures or records (e.g., DTOs or structs) and keep business logic external.
* Postpone entity modeling until implementation details are clearer (emergent modeling).

## Decision Drivers
- Stability and reusability across different applications or bounded contexts
- Degree of encapsulation and abstraction from external systems
- Clarity and alignment with core business rules
- Development effort and complexity of maintaining separation
