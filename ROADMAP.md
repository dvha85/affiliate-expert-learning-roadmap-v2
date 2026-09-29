# Lộ trình — Affiliate Intelligence Bot v2

`CURRICULUM.md` là nguồn có thẩm quyền. File này chỉ là bản tóm tắt dễ đọc của trục Mission chính.

## Giai đoạn A — Ground truth (sự thật nền) trước automation (tự động hóa)

- **O00** — Safe synthetic walkthrough (mô phỏng tổng thể an toàn).
- **M00** — First Real Evidence Packet (gói bằng chứng thật đầu tiên) + Human DecisionPacket (gói quyết định do người lập).
- **M01** — Smallest Deterministic Bot v0.1 (Bot tất định nhỏ nhất).
- **M02** — Trustworthy History + Replay v0.2 (lịch sử đáng tin + phát lại).
- **M03** — First Tracked Human Action + Outcome context (hành động thật đầu tiên do người thực hiện + ngữ cảnh kết quả).

## Giai đoạn B — AI có grounding (bám bằng chứng) nhưng chưa có quyền hành động

- **M04** — Grounded AI Advisor v0.4 (AI tư vấn dựa trên bằng chứng).
- **M05** — First Reviewed Improvement (cải tiến đầu tiên có review).

## Giai đoạn C — Automation chỉ đọc và Agent

- **M06** — Reliable Automatic Watcher (bộ theo dõi tự động chỉ đọc đáng tin).
- **M07** — Read-only Evidence Agent (Agent thu thập bằng chứng chỉ đọc).

## Giai đoạn D — Hành động có kiểm soát

- **M08** — Shadow ActionIntent + deterministic policy (ActionIntent chạy bóng + chính sách tất định).
- **M09** — Durable Approval + Controlled Executor (phê duyệt bền vững + bộ thực thi có kiểm soát).
- **M10** — Governed Canary (canary có quản trị).
- **M11** — Production Closed Loop (vòng kín production có quản trị).

## Nguyên tắc chuyển Mission

Không chuyển Mission chỉ vì code đã viết xong. Phải đạt đúng bằng chứng + quyền hạn + operated gate (cổng chứng minh đã tự vận hành) của Mission hiện tại.

```text
Capability PASS (năng lực đạt)
+ Reality PASS (thực tế đạt) khi được yêu cầu
+ Operated PASS (đã tự vận hành đạt) khi được yêu cầu
→ Mission PASS
```

`ready` của nội dung nghĩa là lesson + mission contract + starter/checkpoints + executable eval/runtime + CI guard đã đủ để learner tự thực hành. Blueprint/placeholder chỉ minh họa không đủ để gọi Mission learner-operable ready.

Từ M02 trở đi, artifact phải nối được với artifact trước đó; ID mồ côi không được dùng để claim Reality/Operated PASS.

Technology (công nghệ) không quyết định thứ tự học. Go/n8n/Agent/MCP/Temporal/OPA chỉ được đưa vào khi Mission hiện tại có nhu cầu và adoption gate (cổng áp dụng) đạt.

## Chặng thực hành sau cập nhật

Đường nối implementation hiện hành được mô tả trong
[walkthrough M00–M05](curriculum/PRACTICE-M00-M05.md). Đây là tài liệu thực
hành trên **một learner Bot**, không tạo Mission hoặc PASS gate mới.

```text
M00–M02: real evidence → deterministic ranking → durable local history/replay
M03–M05: human action → measured outcome → advisor/evaluation/reviewed proposal
M06–M07: read-only watcher/Agent → canonical history, không external write
M08–M11: shadow → approval → governed canary → finite production lease
```

Phạm vi `PERSONAL_LOCAL_SCOPE_ACCEPTED` hiện chỉ hỗ trợ lab local
synthetic/read-only và operated bounded lifecycle của maintainer. Nó giúp luyện
capability/continuity, nhưng không đóng Reality/Operated evidence thật cho
Mission cần E1–E6 và không thay `NOT_READY_FOR_PRODUCTION`.

Trạng thái đó không ngăn tiếp tục M00–M02 theo [PROGRESS.md](PROGRESS.md).
Khi chưa có quyền/kênh/outcome thật cho M03–M05, làm rehearsal có input/output
và failure case trong walkthrough, ghi rõ gap và giữ gate. M08–M11 là chặng
nâng cao dài hạn; phạm vi cá nhân không xóa các Mission này hoặc tự cấp PASS.
