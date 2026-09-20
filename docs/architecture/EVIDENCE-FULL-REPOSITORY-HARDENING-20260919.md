# Full repository hardening evidence — 2026-09-19

Status: `PARTIAL`, bounded offline/fixture/read-only implementation evidence.

The implementation branch adds limit+1 request-body guards and `413` responses
before canonical history mutation, constant-time bearer authentication for
stateful loopback adapter endpoints, and matching headers in n8n blueprints and
offline runners. GitHub Actions are SHA-pinned, Ubuntu is explicit `24.04`, Go
cache dependency paths are declared, n8n path coverage includes all module
changes, and `curriculum-gate`/`mission-gate` fail closed. Security scanning,
Dependabot, CODEOWNERS, editor attributes, and contributor/security guidance are
also checked in.

Local verification: the readiness audit remains
`NOT_READY_FOR_PRODUCTION`, 154 scoped claims; all 174 Python tests pass, JSON
parsing, Python compile, static n8n validators and `git diff --check` pass. Go
and real n8n execution were not run because those executables are unavailable on
the host. Hosted PR #475 checks passed Curriculum, Mission (including Go,
Windows and n8n jobs), Security and CodeQL; provider/live and production
verification remain required.

This evidence does not establish branch protection, provider/live execution,
deployment recovery, beginner pilot acceptance, business outcomes, distributed
locking, power-loss durability, or production readiness.


The current graph count also includes the F-08 M11 process-clock authority evidence recorded at `docs/architecture/EVIDENCE-F08-M11-PROCESS-CLOCK-20260920.md`; this remains bounded offline/fixture/read-only evidence.
