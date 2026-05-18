---
status: proposed
tags:
    - use-case
    - application-logic
    - workflow
    - clean-architecture
links:
    precedes:
        - "0004"
    succeeds:
        - "0002"
---

# Select use cases

## Context and Problem Statement
How should the use case for handling a specific workflow be structured to encapsulate the application-specific business logic while remaining independent of external systems?

## Considered Options
* Define a dedicated use case class or module that orchestrates the workflow and contains all relevant business logic.
* Implement the workflow as part of a service layer shared across use cases.
* Integrate the workflow logic directly within a controller or adapter component to reduce indirection.

## Decision Drivers
- Separation of application logic from infrastructure and delivery concerns
- Maintainability and clarity of workflow responsibilities
- Ease of testing and mocking in isolation
- Reusability of the use case logic across interfaces or delivery mechanisms
