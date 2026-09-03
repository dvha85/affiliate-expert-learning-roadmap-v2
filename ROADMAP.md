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
