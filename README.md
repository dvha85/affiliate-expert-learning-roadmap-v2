# Affiliate Intelligence Bot — lộ trình học v2

Repo chính để học và xây một **Affiliate Intelligence Bot tiến hóa dần từ bằng chứng thật tới tự động hóa có kiểm soát**.

> Repo lịch sử: `dvha85/affiliate-expert-learning-roadmap`.

## Bắt đầu

1. Đọc `CURRICULUM.md` — nguồn có thẩm quyền duy nhất về thứ tự Mission, bằng chứng, mức tự động hóa và mô hình PASS.
2. Đọc `curriculum/README.md` — đường học hiện hành cho người học.
3. Nếu bắt đầu từ máy chưa chuẩn bị môi trường, học `BOOT.0`.
4. Nếu chưa từng chạy/sửa/test Bot, học `BOOT.1`.
5. Bắt đầu M00 tại `curriculum/M00/M00.1-affiliate-intelligence-objective.md`.

## Lộ trình chuẩn

```text
O00 — walkthrough tổng thể bằng dữ liệu synthetic (mô phỏng), không có side effect
→ M00 — real evidence (bằng chứng thật)
→ M01 — deterministic Bot (Bot tất định)
→ M02 — trustworthy history/replay (lịch sử đáng tin + phát lại)
→ M03 — tracked human action + measurement (hành động người làm + đo lường)
→ M04 — grounded AI advisor (AI tư vấn dựa trên bằng chứng)
→ M05 — reviewed improvement (cải tiến có review)
→ M06 — automatic read-only watcher (bộ theo dõi tự động chỉ đọc)
→ M07 — read-only evidence Agent (Agent thu thập bằng chứng chỉ đọc)
→ M08 — shadow ActionIntent + policy (ý định hành động chạy bóng + chính sách)
→ M09 — approval-gated executor (thực thi qua cổng phê duyệt)
→ M10 — governed canary (tự động hóa canary có kiểm soát)
→ M11 — production closed loop (vòng kín production có quản trị)
```

## Các bất biến kiểm soát

```text
AI confidence (độ tin cậy AI) != execution permission (quyền thực thi)
Decision (quyết định) != Approval (phê duyệt) != Execution (thực thi)
Agent proposal (đề xuất của Agent) != authorized ActionIntent (ActionIntent đã được cấp quyền)
Tool result (kết quả tool) != trusted evidence (bằng chứng đáng tin)
real evidence (bằng chứng thật) != automatic recommendation (khuyến nghị tự động)
real != reliable != current != authoritative != complete
(thật != đáng tin != hiện hành != có thẩm quyền != đầy đủ)
```

## Nguyên tắc công nghệ

```text
MISSION xác định CAPABILITY (năng lực) + AUTHORITY (quyền hạn)
TECHNOLOGY PROFILE xác định cách IMPLEMENTATION (triển khai)

Tool thay đổi != Curriculum thay đổi
```

Go, n8n, MCP, OpenTelemetry, Langfuse, Playwright, OpenAI Agents SDK, Hermes Agent, Windmill, Temporal, OPA và rule engine chỉ được áp dụng khi giải quyết bottleneck (điểm nghẽn) đã quan sát được và không vượt authority ceiling (trần quyền hạn) của Mission.

## Chính sách repo

Repo v2 không mang lớp compatibility/migration (tương thích/chuyển đổi) từ repo cũ. Legacy syllabus, numeric lesson map, Mission trùng lặp, migration script và runtime lịch sử chỉ tồn tại ở repo lịch sử.
