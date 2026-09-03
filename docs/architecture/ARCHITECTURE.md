# Kiến trúc chuẩn — Governed Affiliate Intelligence Bot

## 1. Architecture authority

Curriculum quyết định **khi nào** capability/authority được mở. Architecture này quyết định **ranh giới hệ thống**; vendor/framework không sở hữu semantics.

```text
Observation / Evidence
        │
        ▼
Deterministic Domain Core
        │
        ├── DecisionPacket
        │
        ├── History / Replay
        │
        └── Grounding validation
        │
        ▼
ActionIntent (M08+)
        │
        ▼
Deterministic Policy / Risk
        │
        ├── DENY / WAIT / GET_MORE_DATA / HUMAN_REVIEW
        └── ALLOW
                │
                ▼
       Approval khi required
                │
                ▼
       Controlled Execution
                │
                ▼
     Outcome → Evaluation → Reviewed Improvement
```

## 2. Ownership

### Deterministic Domain / Governance Core

Sở hữu contract/behavior của:

- evidence schema và validation;
- identity/provenance/freshness;
- canonical history/replay;
- deterministic decision states;
- `DecisionPacket`, `ActionIntent`, `PolicyDecision`;
- risk/authorization semantics;
- audit/correlation invariants.

### Orchestration

Sở hữu trigger/schedule/integration/retry/approval routing và bounded execution plumbing. Orchestrator không tự trở thành policy authority.

### AgentRuntime

Sở hữu unstructured research/reasoning/proposal trong permission ceiling. Agent không sở hữu truth hoặc authorization.

## 3. Invariants

```text
Decision != Approval != Execution
AI confidence != execution permission
Tool result != trusted evidence
Agent proposal != authorized ActionIntent
Deterministic Policy unavailable/invalid/unverified
→ no consequential execution
```

## 4. Implementation flexibility

```text
DETERMINISTIC CORE FIRST != CODE FIRST
```

Go là deterministic reference/fallback khi code làm behavior rõ hơn. Visual rule engine có thể implement deterministic semantics nếu parity/version/audit/fail-closed gate đạt. n8n có thể orchestration. Agent runtime có thể thay đổi mà không đổi authority model.

## 5. External action boundary

External action đầu tiên ở M03 và do human thực hiện. Machine ActionIntent bắt đầu shadow ở M08. Machine execution chỉ mở ở M09 qua deterministic policy + approval khi required; bounded auto-action bắt đầu M10.
