---
mission_id: M00
title: First Real Evidence Packet
status: ready
minimum_evidence: E1
authority: human/read-only
external_side_effects: false
---

# Mission M00 — First Real Evidence Packet

## Ship target

```text
3+ public observations E1
→ classify fact / estimate / assumption / unknown
→ Human DecisionPacket
→ state + reason + missing evidence + next measurement
→ NO external execution
```

## Required lessons

- `curriculum/M00/M00.1-affiliate-intelligence-objective.md`
- `curriculum/M00/M00.2-evidence-uncertainty.md`
- `curriculum/M00/M00.3-decision-approval-execution.md`

## Evidence bundle

Mỗi observation tối thiểu có `source_url`, `observed_at`, `access_method`, claim/value, `claim_kind`, limitation.

Human DecisionPacket tối thiểu có:

```text
question
supported_facts
assumptions
unknowns
decision_state: RANK_SCENARIO | GET_MORE_DATA | HUMAN_REVIEW
reason
missing_evidence
next_measurement
action: null
```

## Failure cases

- placeholder URL được gọi là E1;
- assumption ghi thành fact;
- missing đổi thành observed zero;
- sample/synthetic được gọi là market reality;
- real provenance tự nâng output thành `RECOMMEND`;
- DecisionPacket tạo external action.

## PASS

### Capability
- tạo evidence packet có provenance/uncertainty đúng;
- tạo Human DecisionPacket có reason/missing evidence/next measurement.

### Reality
- có ít nhất 3 public observations E1 thật; nếu không thể, ghi `BLOCKED_EXTERNAL` và chưa PASS Reality.

### Operated
- packet/version đủ rõ để M01 dùng làm context cho deterministic baseline.

## Result

`pre-bot`: market truth/context có trước; Bot v0.1 bắt đầu ở M01.
