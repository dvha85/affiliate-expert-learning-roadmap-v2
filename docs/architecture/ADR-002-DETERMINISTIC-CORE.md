# ADR-002 — Deterministic Core + implementation flexibility

**Status:** Accepted  
**Date:** 2026-09-03

## Decision

Canonical ownership thuộc **semantics/contracts/tests**, không thuộc một language/vendor.

```text
DETERMINISTIC CORE FIRST != CODE FIRST
```

- Go là deterministic reference/fallback khi code làm behavior rõ hơn.
- Visual rule engine có thể implement deterministic logic nếu parity/version/audit/fail-closed gate đạt.
- n8n sở hữu orchestration plumbing, không sở hữu truth/policy mặc định.
- AgentRuntime sở hữu research/reasoning/proposal trong permission ceiling, không sở hữu authorization.

## Fail-safe invariant

```text
Deterministic Policy Authority unavailable / invalid / unverified
→ no consequential execution
```

## Non-goals

ADR này không bắt buộc Go, n8n, OPA, MCP hay bất kỳ Agent framework nào để PASS một Mission nếu contract có thể đạt bằng implementation đơn giản hơn.
