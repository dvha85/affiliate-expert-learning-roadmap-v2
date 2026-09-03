# Learner Bot Continuity — một Bot tiến hóa xuyên M01→M11

## Mục đích

Repo có nhiều reference surface (bề mặt tham chiếu), nhưng learner chỉ xây **một hệ thống Affiliate Intelligence Bot tiến hóa liên tục**. Mission sau phải nối artifact/state/capability của Mission trước; không được PASS bằng một demo rời không gắn lại vào learner Bot.

```text
M01 learner Bot baseline
→ M02 history/replay
→ M03 human action/outcome adapter
→ M04 grounded advisor
→ M05 evaluation/reviewed improvement
→ M06 read-only watcher adapter
→ M07 read-only Agent adapter
→ M08 shadow ActionIntent/policy
→ M09 approved executor
→ M10 governed canary
→ M11 governed production loop
```

## Vai trò của các thư mục

### `lab/affiliate-bot/`

Reference learner baseline (mốc Bot tham chiếu cho người học) của M01–M02. Từ M03 trở đi, learner **mở rộng chính Bot/workspace của mình** từ baseline này; repo không copy toàn bộ M03–M11 vào đây vì như vậy sẽ tạo một implementation thứ hai phải giữ parity.

### `lab/mission-runtime/`

Conformance oracle/harness (bộ đối chiếu chuẩn), không phải Bot thứ hai và không phải nơi learner “PASS bằng cách chạy demo”. Runtime này chứng minh contract/failure semantics offline để learner so implementation của mình với behavior chuẩn.

### `lab/n8n/`

Adapter/orchestration reference. n8n trigger/workflow không sở hữu canonical business state, policy hay authority. Workflow output phải map về contracts của Deterministic Core.

### `contracts/`

Machine-readable boundary chuẩn. Learner implementation có thể dùng Go, n8n hay framework khác nhưng artifact phải giữ canonical identity/linkage.

## Continuity Gate bắt buộc từ M03

Reality/Operated PASS của M03–M11 cần một **Integration Evidence (bằng chứng tích hợp)** ghi tối thiểu:

```text
learner_bot_commit_or_version:
previous_mission_artifact_refs:
new_capability_entrypoint:
canonical_contracts_used:
conformance_cases_run:
real_operated_evidence_refs:
authority_ceiling_verified:
known_gap_or_limitation:
```

`previous_mission_artifact_refs` phải resolve được tới artifact thật đã dùng; một ảnh chụp demo hoặc fixture mới độc lập không đủ.

## Matrix tích hợp

| Mission | Capability mới | Phải nối lại vào hệ thống trước bằng |
|---|---|---|
| M03 | human action + outcome | `DecisionPacket.decision_id` → `ActionRecord` → `OutcomeRecord` |
| M04 | grounded advisor | exact evidence/history IDs từ Bot hiện tại; advisor không sở hữu truth |
| M05 | evaluation + proposal | exact action/outcome/evaluation chain; proposal `auto_apply=false` |
| M06 | watcher tự động chỉ đọc | watcher emit canonical `Observation`; history write thuộc Deterministic Core |
| M07 | read-only Agent | Agent nhận exact evidence IDs + Tool Registry; tool output vẫn untrusted |
| M08 | shadow ActionIntent/policy | intent bind current decision/evidence; policy shadow-only |
| M09 | controlled executor | approval/auth/execution bind exact current intent/policy |
| M10 | governed canary | grant/ledger/cost/outcome bind executor state hiện hành |
| M11 | production loop | promotion/lease/health/ledger/cycle nối E5 thật từ M10 |

## Không được tạo hai nguồn sự thật

```text
mission-runtime state != learner canonical state
n8n static data != canonical history
Agent memory != canonical evidence/history
telemetry backend != canonical audit/business state
```

Nếu một adapter cần cache để retry/change detection, cache đó phải được đặt tên và documented (ghi rõ) là cache. Mất cache có thể làm giảm khả năng phát hiện thay đổi nhưng không được âm thầm reset canonical history/budget/approval state.

## Khi nào continuity đạt

Một Mission đạt continuity khi learner có thể trả lời được:

1. artifact mới dùng exact artifact ID nào từ Mission trước;
2. capability mới chạy ở entrypoint nào trong cùng learner system;
3. conformance runtime/eval nào đã dùng để kiểm failure case;
4. state nào là canonical và component nào chỉ là adapter/cache;
5. rollback/recovery đưa Bot về version/state nào.

Nếu không trả lời được 5 câu này, Mission có thể có Capability proof nhưng chưa đủ Operated continuity.
