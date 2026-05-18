---
status: proposed
tags:
    - modularity
    - use-cases
    - feature-modules
    - organization
links:
    precedes:
        - "0015"
    succeeds:
        - "0013"
---

# Structure usecase modules

## Context and Problem Statement
How should feature modules be structured within the use case layer to maintain clarity, modularity, and scalability of the application?

## Considered Options
* Organize use cases by feature verticals (e.g., `user/register`, `invoice/generate`) with dedicated subdirectories per feature.
* Group use cases by technical function (e.g., all use cases, all interfaces) across the application.
* Follow domain-driven design aggregates or bounded contexts to group related business capabilities.

## Decision Drivers
- Discoverability and navigation of use case logic
- Encapsulation and independence of features or domains
- Reusability and testability of logic across modules
- Scalability as the application grows
- Alignment with business concepts and responsibilities
