# Kiến trúc chuẩn — Governed Affiliate Intelligence Bot

## 1. Architecture authority

Curriculum quyết định **khi nào** capability/authority được mở. Architecture này quyết định **ranh giới hệ thống**; vendor/framework không sở hữu semantics.

```text
                         ┌──────────────────────────────┐
                         │  Read-only Watcher / Agent   │
                         │  M06 / M07                   │
                         └──────────────┬───────────────┘
                                        │
                                        ▼
Observation / Evidence ───────────────► Normalize + Validate
                                        │
                                        ▼
                               Canonical History / Replay
                                        │
                                        ▼
                              Deterministic Domain Core
                                        │
                     ┌──────────────────┴──────────────────┐
                     │                                     │
                     ▼                                     ▼
              DecisionPacket                       Grounded AI Advisor
                                                        M04/M05
                     │                                     │
                     └──────────────────┬──────────────────┘
                                        ▼
                              ActionIntent (M08+)
                                        │
                                        ▼
                          Deterministic Policy / Risk
                                        │
                   ┌────────────────────┴───────────────────┐
                   │                                        │
     DENY / WAIT / GET_MORE_DATA / HUMAN_REVIEW           ALLOW
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

M03 là ngoại lệ có chủ đích trước machine `ActionIntent`: người thực hiện external action thật, hệ thống chỉ ghi/validate `ActionRecord` và `OutcomeRecord`.

## 2. Ownership

### Deterministic Domain / Governance Core

Sở hữu contract/behavior của:

- evidence schema và validation;
- identity/provenance/freshness;
- canonical history/replay;
- deterministic decision states;
- `DecisionPacket`, `ActionRecord`, `OutcomeRecord`, `EvaluationRecord`, `ImprovementProposal`, `ReviewRecord`, `ActionIntent`, `PolicyDecision`;
- risk/authorization semantics;
- audit/correlation invariants;
- cross-artifact linkage khi contract yêu cầu.

### Orchestration

Sở hữu trigger/schedule/integration/retry/approval routing và bounded execution plumbing. Orchestrator không tự trở thành policy authority. n8n output phải map về canonical contract thay vì tạo data model song song.

### AgentRuntime

Sở hữu unstructured research/reasoning/proposal trong permission ceiling. Agent không sở hữu truth hoặc authorization. Tool output là untrusted data cho tới khi qua deterministic validation/grounding.

## 3. Invariants

```text
Decision != Approval != Execution
AI confidence != execution permission
Tool result != trusted evidence
Agent proposal != authorized ActionIntent
Schema-valid reference != resolved provenance
Deterministic Policy unavailable/invalid/unverified
→ no consequential execution
```

## 4. Implementation flexibility

```text
DETERMINISTIC CORE FIRST != CODE FIRST
```

Go là deterministic reference/fallback khi code làm behavior rõ hơn. Visual rule engine có thể implement deterministic semantics nếu parity/version/audit/fail-closed gate đạt. n8n có thể orchestration. Agent runtime có thể thay đổi mà không đổi authority model.

`lab/mission-runtime` là conformance/integration harness (bộ kiểm tương thích/tích hợp), không phải một Bot thứ hai thay thế `lab/affiliate-bot`. Mission sau phải reuse canonical contracts/history hoặc chứng minh adapter rõ ràng.

## 5. External action boundary

External action đầu tiên ở M03 và do human thực hiện. Machine ActionIntent bắt đầu shadow ở M08. Machine execution chỉ mở ở M09 qua deterministic policy + approval khi required; bounded auto-action bắt đầu M10.
