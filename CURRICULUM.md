# Chương trình hiện hành — Affiliate Intelligence Bot có kiểm soát

**Trạng thái:** chương trình chuẩn đang hoạt động  
**Đối tượng:** Người mới; không mặc định biết terminal (dòng lệnh), Go hay Agent framework (khung phát triển Agent).  
**Mục tiêu:** xây một **Affiliate Intelligence Bot tiến hóa dần tới tự động hóa cao nhưng vẫn kiểm soát được**.

`CURRICULUM.md` là nguồn có thẩm quyền duy nhất cho thứ tự học, thang bằng chứng, trần quyền hạn và mô hình PASS.

## 1. Mục tiêu hệ thống

```text
Evidence (bằng chứng)
→ Deterministic Decision (quyết định tất định)
→ Grounded AI (AI dựa trên bằng chứng) khi cần
→ ActionIntent (ý định hành động dạng đề xuất, không tự có quyền)
→ Deterministic Policy / Risk (chính sách / rủi ro tất định, không tự có quyền)
→ Human Approval / Delegated Authority (phê duyệt / ủy quyền có giới hạn) khi cần
→ ExecutionAuthorization (ủy quyền thực thi cụ thể)
→ Controlled Execution (thực thi có kiểm soát)
→ Outcome (kết quả)
→ Evaluation (đánh giá)
→ Reviewed Improvement (cải tiến đã review)
↺
```

Các bất biến:

```text
AI confidence (độ tin cậy AI) != execution permission (quyền thực thi)
Decision (quyết định) != Approval (phê duyệt) != Execution (thực thi)
Agent proposal (đề xuất của Agent) != authorized action (hành động đã được cấp quyền)
ActionIntent.execution_authorized = false
PolicyDecision.execution_authorized = false
Tool result (kết quả tool) != trusted evidence (bằng chứng đáng tin)
real evidence (bằng chứng thật) != automatic recommendation (khuyến nghị tự động)
real != reliable != current != authoritative != complete
(thật != đáng tin != hiện hành != có thẩm quyền != đầy đủ)
```

## 2. Reality-First (ưu tiên thực tế) nhưng không Publish-First (ưu tiên xuất bản)

M00 bắt đầu bằng public observation (quan sát công khai) thật nhưng không yêu cầu publish (đăng), spend (chi tiền), send (gửi) hoặc mutate account (thay đổi tài khoản). External action (hành động bên ngoài) đầu tiên thuộc M03 và do người thực hiện sau review.

```text
REALITY-FIRST != PUBLISH-FIRST
(ưu tiên thực tế != ưu tiên xuất bản)
```

## 3. Trục Mission chính

`Bằng chứng tối thiểu` là bằng chứng Mission hiện tại phải có. `Readiness target` (mục tiêu sẵn sàng) chỉ nói Mission phải chuẩn bị capability/contract cho mức bằng chứng sau; **readiness không phải evidence đã đạt**. Với M01, E0 chứng minh capability; Reality phải dùng lại E1 context thật từ M00 và không biến fixture thành E1.

| Mission | Kết quả bàn giao | Bằng chứng tối thiểu | Readiness target | Trần quyền hạn |
|---|---|---|---|---|
| O00 | Safe synthetic walkthrough (mô phỏng tổng thể an toàn), không PASS | E0 | — | không side effect |
| M00 | First Real Evidence Packet (gói bằng chứng thật đầu tiên) + Human DecisionPacket (gói quyết định do người lập) | E1 | — | người/read-only |
| M01 | Smallest Deterministic Bot v0.1 (Bot tất định nhỏ nhất) | E0; Reality dùng E1 context từ M00 | — | A0 tất định, không hành động |
| M02 | Trustworthy History + Replay v0.2 (lịch sử đáng tin + phát lại) | E1 | E3 | A0 tất định |
| M03 | First Tracked Human Action + Outcome context (hành động thật đầu tiên do người làm + ngữ cảnh kết quả) | E2→E3 | — | người thực thi |
| M04 | Grounded AI Advisor v0.4 (AI tư vấn dựa trên bằng chứng) | E3 | — | A1 tư vấn, không tool ghi |
| M05 | First Reviewed Improvement (cải tiến đầu tiên có review) | E4 | — | A1 chỉ đề xuất |
| M06 | Reliable Automatic Watcher (bộ theo dõi tự động chỉ đọc đáng tin) | E4 | — | tự động chỉ đọc |
| M07 | Read-only Evidence Agent (Agent bằng chứng chỉ đọc) | E4 | — | A2-RO |
| M08 | Shadow ActionIntent + Policy (ActionIntent chạy bóng + chính sách) | E4 | — | A3-shadow |
| M09 | Durable Approval + Controlled Executor (phê duyệt bền vững + bộ thực thi có kiểm soát) | E4 | E5 | hành động qua cổng phê duyệt |
| M10 | Governed Canary (canary có quản trị) | E5 | — | tự động RISK0/RISK1 trong giới hạn; RISK2 cần phê duyệt |
| M11 | Production Closed Loop (vòng kín production có quản trị) | E6 | — | production có quản trị |

```text
O00 → M00 → M01 → M02 → M03 → M04 → M05 → M06 → M07 → M08 → M09 → M10 → M11
```

Mỗi Mission chỉ tăng một lớp capability (năng lực) hoặc authority (quyền hạn) chính. Không dùng capability của Mission sau để vượt gate (cổng kiểm soát) của Mission trước.

## 4. Lớp bài học cho người học

Bài học dùng ID theo Mission:

```text
BOOT.0
BOOT.1
M00.1
M00.2
M00.3
M01.1 ...
```

`BOOT.0`/`BOOT.1` là onboarding/tooling (làm quen môi trường/công cụ), không phải Mission PASS gate. Danh mục bài học đánh số của repo lịch sử không được chuyển sang learner path (đường học) mới. Kiến thức có giá trị chỉ được đưa sang khi một Mission cụ thể cần đến.

## 5. M00 — First Real Evidence Packet (gói bằng chứng thật đầu tiên)

M00 không cần Go, n8n, AI, API key hoặc affiliate automation (tự động hóa affiliate).

Các thẻ kiến thức đầu tiên:

1. `M00.1` — Affiliate Intelligence Bot đang tối ưu điều gì?
2. `M00.2` — Evidence (bằng chứng), uncertainty (bất định) và missing data (dữ liệu thiếu).
3. `M00.3` — Decision (quyết định) ≠ Approval (phê duyệt) ≠ Execution (thực thi).

Mục tiêu bàn giao:

```text
3+ public observations (quan sát công khai) có observation_id + source + observed_at + access_method
→ provenance/limitation (nguồn gốc/giới hạn) rõ
→ fact / estimate / assumption / unknown
→ Human DecisionPacket bind exact evidence_ids
→ state + reason + missing evidence + next measurement
→ action = null
→ KHÔNG thực thi hành động bên ngoài
```

## 6. Tiến trình quyền hạn

```text
M00 — người/read-only (chỉ đọc)
→ M01–M02 — A0 tất định
→ M03 — người thực thi
→ M04–M05 — A1 tư vấn/chỉ đề xuất
→ M06 — tự động chỉ đọc
→ M07 — A2-RO
→ M08 — A3-shadow (chạy bóng; ActionIntent/PolicyDecision không tự có authority)
→ M09 — thực thi qua per-action human approval
→ M10 — tự động hóa có quản trị trong CanaryGrant giới hạn
→ M11 — production có quản trị trong finite ProductionLease
```

## 7. Nguyên tắc triển khai

```text
DETERMINISTIC CORE FIRST != CODE FIRST
NO-CODE WHEN AUDITABLE
CODE WHEN IT REDUCES AMBIGUITY OR FAILURE SURFACE
AGENT WHEN DETERMINISTIC LOGIC IS NOT ENOUGH
READ-ONLY AUTOMATION MAY START AT M06 UNDER A BOUNDED SAFE PROFILE
(tự động hóa chỉ đọc có thể bắt đầu ở M06 dưới hồ sơ an toàn bị giới hạn)
CONSEQUENTIAL / WRITE AUTOMATION ONLY AFTER EVIDENCE + POLICY + AUDIT + RECOVERY
(chỉ tự động hóa ghi/có hậu quả sau khi có bằng chứng + chính sách + kiểm toán + phục hồi)
CURRENT IMPLEMENTATION LIMIT != FUNDAMENTAL SYSTEM LAW
```

Semantics (ngữ nghĩa) của Mission không phụ thuộc vendor/framework. Technology Profile (hồ sơ công nghệ) có thể đổi mà không đổi trần quyền hạn.

## 8. Tính liên tục của artifact

Từ M02 trở đi, learner không xây các demo rời. ID tham chiếu trong artifact phải resolve được về artifact trước đó khi Mission contract yêu cầu:

```text
Observation / HistoryRecord
→ DecisionPacket
→ ActionRecord (human) hoặc ActionIntent (proposal)
→ ExecutionRecord khi có machine execution
→ EffectRef(HUMAN_ACTION | MACHINE_EXECUTION)
→ OutcomeRecord
→ EvaluationRecord
→ ImprovementProposal
→ ReviewRecord
→ automated Observation
→ grounded Advisor/Agent output
```

`OutcomeRecord` và `EvaluationRecord` dùng `EffectRef` để không trộn human `action_id` với machine `execution_id`.

```text
HUMAN_ACTION     → effect_id = ActionRecord.action_id
MACHINE_EXECUTION → effect_id = ExecutionRecord.execution_id
```

Schema hợp lệ nhưng tham chiếu mồ côi không đủ để claim Reality/Operated PASS.

## 9. Thang bằng chứng thực tế

| Mức | Bằng chứng |
|---|---|
| E0 | synthetic/test/replay (mô phỏng/kiểm thử/phát lại); chỉ chứng minh plumbing/behavior (luồng kỹ thuật/hành vi) |
| E1 | quan sát công khai thật có source + observed_at + access_method + limitation/provenance |
| E2 | hành động bên ngoài thật do người thực hiện có `ActionRecord` |
| E3 | outcome/analytics/export (kết quả/phân tích/dữ liệu xuất) thật, kể cả observed value = 0 |
| E4 | Decision → Action → Outcome → Evaluation → reviewed proposal (quyết định → hành động → kết quả → đánh giá → đề xuất đã review) |
| E5 | governed canary (canary có quản trị) trong giới hạn, có policy/audit/kill switch |
| E6 | production loop (vòng production) qua observation window + recovery + reviewed improvement |

Sample (mẫu) không thể thay E1–E6. `real` chỉ mô tả origin (nguồn gốc); không tự chứng minh source đáng tin, hiện hành, có thẩm quyền hoặc đầy đủ.

```text
readiness_target = X
!= evidence level X achieved
!= permission to skip Reality/Operated proof
```

## 10. Mô hình PASS

Mission chỉ PASS khi contract của Mission đạt các lớp áp dụng:

```text
Capability (năng lực)
+ Reality (thực tế)
+ Operated (đã tự vận hành/chạy chứng minh)
```

`draft`, `ready`, CI xanh hoặc fixture PASS không tự tạo Reality PASS.

## 11. Thứ tự thẩm quyền trong repo v2

1. `CURRICULUM.md` — thứ tự/bằng chứng/quyền hạn/PASS.
2. `curriculum/` — đường học hiện hành.
3. `missions/` — contract thực hiện Mission.
4. `docs/architecture/` và `docs/technology/` — chi tiết triển khai/an toàn.
5. `contracts/` — ranh giới máy có thể đọc.
6. `starter-kits/`, `evals/`, `lab/` — hỗ trợ triển khai/bằng chứng.

Không tạo migration mapping layer (lớp ánh xạ chuyển đổi) để che conflict (xung đột). Nếu file thẩm quyền thấp mâu thuẫn với file cao hơn, sửa hoặc xóa file thấp hơn.
