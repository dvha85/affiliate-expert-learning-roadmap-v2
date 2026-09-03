# Chương trình hiện hành — Affiliate Intelligence Bot có kiểm soát

**Trạng thái:** Active canonical curriculum  
**Đối tượng:** Người mới; không mặc định biết terminal, Go hay Agent framework.  
**Mục tiêu:** xây một **Affiliate Intelligence Bot tiến hóa dần tới tự động hóa cao nhưng vẫn kiểm soát được**.

`CURRICULUM.md` là authority duy nhất cho learner sequence, evidence ladder, authority ceiling và PASS model.

## 1. Mục tiêu hệ thống

```text
Evidence
→ Deterministic Decision
→ Grounded AI khi cần
→ ActionIntent
→ Deterministic Policy / Risk
→ Human Approval khi cần
→ Controlled Execution
→ Outcome
→ Evaluation
→ Reviewed Improvement
↺
```

Invariant:

```text
AI confidence != execution permission
Decision != Approval != Execution
Agent proposal != authorized ActionIntent
Tool result != trusted evidence
real evidence != automatic recommendation
real != reliable != current != authoritative != complete
```

## 2. Reality-First nhưng không Publish-First

M00 bắt đầu bằng public observation thật nhưng không yêu cầu publish, spend, send hoặc mutate account. External action đầu tiên thuộc M03 và do human thực hiện sau review.

```text
REALITY-FIRST != PUBLISH-FIRST
```

## 3. Mission spine

| Mission | Outcome bàn giao | Evidence tối thiểu | Authority ceiling |
|---|---|---|---|
| O00 | Safe synthetic walkthrough, không PASS | E0 | no side effect |
| M00 | First Real Evidence Packet + Human DecisionPacket | E1 | human/read-only |
| M01 | Smallest Deterministic Bot v0.1 | E0 + E1 support | A0 deterministic, no action |
| M02 | Trustworthy History + Replay v0.2 | E1/E3-ready | A0 deterministic |
| M03 | First Tracked Human Action + Outcome context | E2→E3 | human executes |
| M04 | Grounded AI Advisor v0.4 | E3 | A1 advisory, no tools/write |
| M05 | First Reviewed Improvement | E4 | A1 propose only |
| M06 | Reliable Automatic Watcher | E4 | automatic read-only |
| M07 | Read-only Evidence Agent | E4 | A2-RO |
| M08 | Shadow ActionIntent + Policy | E4 | A3-shadow |
| M09 | Durable Approval + Controlled Executor | E4/E5-ready | approval-gated action |
| M10 | Governed Canary | E5 | bounded RISK0/RISK1 auto; RISK2 approval |
| M11 | Production Closed Loop | E6 | governed production |

```text
O00 → M00 → M01 → M02 → M03 → M04 → M05 → M06 → M07 → M08 → M09 → M10 → M11
```

Mỗi Mission chỉ tăng một lớp capability/authority chính. Không dùng capability Mission sau để vượt gate Mission trước.

## 4. Learner lesson layer

Learner-facing lesson dùng ID theo Mission:

```text
BOOT.0
BOOT.1
M00.1
M00.2
M00.3
M01.1 ...
```

`BOOT.0`/`BOOT.1` là onboarding/tooling, không phải Mission PASS gate. Numeric lesson inventory của repo lịch sử không được migrate thành learner path. Knowledge có giá trị chỉ được đưa sang khi được pull bởi Mission cụ thể.

## 5. M00 — First Real Evidence Packet

M00 không cần Go, n8n, AI, API key hoặc affiliate automation.

Knowledge cards đầu tiên:

1. `M00.1` — Affiliate Intelligence Bot đang tối ưu điều gì?
2. `M00.2` — Evidence, uncertainty và missing data.
3. `M00.3` — Decision ≠ Approval ≠ Execution.

Ship target:

```text
3+ public observations có source + observed_at + access method
→ provenance/limitation rõ
→ fact / estimate / assumption / unknown
→ Human DecisionPacket
→ state + reason + missing evidence + next measurement
→ NO external execution
```

## 6. Authority progression

```text
M00 human/read-only
→ M01–M02 A0 deterministic
→ M03 human executes
→ M04–M05 A1 advisory/propose
→ M06 automatic read-only
→ M07 A2-RO
→ M08 A3-shadow
→ M09 approval-gated execution
→ M10 bounded governed automation
→ M11 governed production
```

## 7. Implementation principles

```text
DETERMINISTIC CORE FIRST != CODE FIRST
NO-CODE WHEN AUDITABLE
CODE WHEN IT REDUCES AMBIGUITY OR FAILURE SURFACE
AGENT WHEN DETERMINISTIC LOGIC IS NOT ENOUGH
AUTOMATION ONLY AFTER EVIDENCE + POLICY + AUDIT + RECOVERY
CURRENT IMPLEMENTATION LIMIT != FUNDAMENTAL SYSTEM LAW
```

Mission semantics không phụ thuộc vendor/framework. Technology profile có thể đổi mà không đổi authority ceiling.

## 8. Real Evidence Ladder

| Level | Bằng chứng |
|---|---|
| E0 | synthetic/test/replay; chỉ chứng minh plumbing/behavior |
| E1 | public observation thật có source + observed_at + access method + limitation/provenance |
| E2 | human external action thật có ActionRecord |
| E3 | outcome/analytics/export thật, kể cả observed value = 0 |
| E4 | Decision → Action → Outcome → Evaluation → reviewed proposal |
| E5 | bounded governed canary có policy/audit/kill switch |
| E6 | production loop qua observation window + recovery + reviewed improvement |

Sample không thể thay E1–E6. `real` chỉ mô tả origin; không tự chứng minh source reliable/current/authoritative/complete.

## 9. PASS model

Mission chỉ PASS khi contract của Mission đạt các lớp áp dụng:

```text
Capability + Reality + Operated
```

`draft`, `ready`, CI xanh hoặc fixture PASS không tự tạo Reality PASS.

## 10. Authority order trong repo v2

1. `CURRICULUM.md` — sequence/evidence/authority/PASS.
2. `curriculum/` — learner path hiện hành.
3. `missions/` — execution contract.
4. `docs/architecture/` và `docs/technology/` — implementation/safety detail.
5. `contracts/` — machine-readable boundaries.
6. `starter-kits/`, `evals/`, `lab/` — implementation/evidence support.

Không tạo migration mapping layer để che conflict. Nếu file authority thấp mâu thuẫn với file cao hơn, sửa hoặc xóa file thấp hơn.
