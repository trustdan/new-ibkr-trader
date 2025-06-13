# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records - documents that capture important architectural decisions made during the development of the IBKR Spread Automation system.

## What is an ADR?

An ADR is a document that captures an important architectural decision made along with its context and consequences. Each ADR describes a single decision and its rationale.

## ADR Template

```markdown
# ADR-NNN: Title

## Status
[Proposed | Accepted | Deprecated | Superseded by ADR-NNN]

## Context
What is the issue that we're seeing that is motivating this decision?

## Decision
What is the decision that we're making?

## Rationale
Why are we making this decision?

## Consequences
What are the positive and negative consequences of this decision?

## References
Links to relevant documentation, discussions, or related ADRs

## Decision Date
YYYY-MM-DD

## Participants
Who was involved in making this decision?
```

## Current ADRs

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [ADR-001](ADR-001-async-first-architecture.md) | Async-First Architecture | Accepted | 2025-01-13 |

## Process

1. **Proposing an ADR**: Create a new file following the template
2. **Discussion**: Share with team for feedback
3. **Decision**: Update status to Accepted/Rejected
4. **Superseding**: Link to new ADR if decision changes

## Why ADRs Matter

- **Documentation**: Captures the "why" behind decisions
- **Onboarding**: Helps new team members understand the architecture
- **History**: Provides context for future changes
- **Transparency**: Makes decision-making process visible