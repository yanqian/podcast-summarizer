<!--
Sync Impact Report
Version change: N/A → 1.0.0
Modified principles: New → Code Quality & Maintainability; New → Testing Discipline; New → User Experience Consistency; New → Performance & Efficiency
Added sections: Baseline Engineering Standards; Workflow & Review Process
Removed sections: Placeholder Principle 5
Templates requiring updates: ✅ .specify/templates/plan-template.md; ✅ .specify/templates/spec-template.md; ✅ .specify/templates/tasks-template.md
Follow-up TODOs: None
-->

# Podcast Summarizer Constitution

## Core Principles

### Code Quality & Maintainability
Code must remain small, cohesive units with clear ownership, automated formatting and linting must pass
pre-merge, and every change requires peer review plus concise inline comments where logic is non-obvious.
Dependencies must be justified, pinned, and removed when unused; public interfaces demand docs updates
alongside code. Rationale: disciplined structure keeps the summarizer reliable and easy to evolve.

### Testing Discipline
Every change must add or update automated tests that fail before implementation and pass after: unit tests
for logic branches, integration/contract tests for external I/O, and regression tests for discovered bugs.
Touched code must reach at least 85% coverage, and CI must block merges on failing or missing tests.
Rationale: fast, repeatable feedback prevents regressions in summarization quality.

### User Experience Consistency
All user-facing outputs (CLI, API responses, docs) must follow a consistent interaction model: predictable
flags/arguments, stable output schemas, actionable errors with remediation steps, and accessibility-aware
text (clear language, no ambiguous abbreviations). UX acceptance notes belong in specs and PRs to confirm
that behavior matches existing patterns. Rationale: consistent flows keep the product trustworthy.

### Performance & Efficiency
Each feature must declare measurable performance budgets in its spec (default: p95 of core operations
under 200ms outside long-running model inference; background jobs document throughput/latency targets).
Code must include lightweight instrumentation or benchmarks to verify budgets, and any regression over 5%
from baseline needs remediation or an approved exception. Rationale: predictable performance protects user
experience and cost.

## Baseline Engineering Standards

- Definition of Done: lint/format clean, principle-aligned tests added and passing, UX notes verified, and
  performance checks executed or benchmark evidence provided.
- Observability: add structured logs around external calls and performance-critical paths to support the
  performance and testing principles.
- Documentation: update usage examples, CLI/API help, and changelog entries alongside functional changes.
- Risk control: isolate experiments behind flags, and record fallback behavior for external dependencies.

## Workflow & Review Process

- Planning: every feature needs a spec and implementation plan that map requirements to tests, UX checks,
  and performance budgets before coding starts.
- Reviews: PRs must cite how each principle is met (tests added, UX consistency check, performance evidence)
  and must not merge with failing automation.
- Releases: include a brief verification note summarizing test results, UX validation, and performance
  measurements for the change scope.

## Governance

This constitution supersedes conflicting practices. Amendments require a PR describing the change, the
version bump rationale (semver), and how existing templates/docs stay in sync. Versioning follows: MAJOR
for principle removals or incompatible rewrites, MINOR for new principles or material expansions, PATCH for
clarifications. Compliance is reviewed at planning (Constitution Check), PR review (evidence linked), and
release sign-off (results recorded).

**Version**: 1.0.0 | **Ratified**: 2025-11-29 | **Last Amended**: 2025-11-29
