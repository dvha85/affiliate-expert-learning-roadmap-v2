# RP-01 Windows ancestor-race audit guard

Status: `VERIFIED_OFFLINE` for evidence-integrity hardening.

## Hosted verification

- PR #464 implementation head:
  `ad38a9a65b32b0800aed3a229e493916eb4c330f`;
- PR #464 squash merge:
  `d339f6d6db218b2a33fb19eda81a5338c0e26004`;
- GitHub reported 13 checks passed on the exact implementation head;
- local readiness audit: PASS, `NOT_READY_FOR_PRODUCTION`, 150 scoped claims;
- local Python readiness-audit suite: 126 tests passed.

## Guarded evidence

The readiness audit now requires the native Windows reparse-point pinning
sources, all four controlled ancestor-junction regressions, the
`windows-latest` CI job and the explicit conservative boundary disclosure.
Isolated negative fixtures remove each evidence class and require audit
failure.

## Boundary

This confirms only evidence integrity around the bounded Windows single-host
fail-closed seam. Windows still lacks a portable `openat`/`mkdirat` equivalent
for arbitrary multi-component traversal in this path, so no POSIX traversal
parity is claimed. Distributed locking, crash/power-loss durability, atomic
multi-file publication, provider/live execution, deployment, pilot, business
outcomes and production readiness remain open.
