# Tiến độ — trạng thái người học

Repo v2 bắt đầu theo trục Mission mới. File này chỉ ghi **tiến độ học thực tế trong repo v2**; kết quả từ repo lịch sử được tách riêng để tránh nhầm với bài đã học trong chương trình mới.

## Trạng thái học v2 hiện tại

- Trạng thái: **IN PROGRESS (đang học)**
- Mission hiện tại: **M00 — First Real Evidence Packet (gói bằng chứng thật đầu tiên), chưa bắt đầu Reality evidence (bằng chứng thực tế)**
- Bài đã hoàn thành trong repo v2: **BOOT.0 — PASS; BOOT.1 — PASS; O00.1 — ORIENTATION COMPLETED (hoàn thành định hướng)**
- Bài tiếp theo: **M00.1 — Affiliate Intelligence Bot đang tối ưu điều gì?**
- Tài liệu tham khảo BOOT: **`curriculum/BOOT/BOOT-REFERENCE.md`**
- Trạng thái Reality (thực tế) của M00: **chưa bắt đầu**

## Evidence học tập cho BOOT

### BOOT.0 — PASS

Learner đã:

- xác định được local environment (môi trường local);
- chạy được Git;
- chạy được Go `1.27.0` trên `darwin/arm64`;
- chạy `go test ./...` thành công trong `lab/affiliate-bot`;
- giải thích được vì sao API key/token/credential là secret (bí mật) và không được commit vào Git repo.

BOOT.0 PASS chỉ chứng minh environment/tooling basics (nền tảng môi trường/công cụ), không tạo E1 evidence hoặc Mission PASS.

### BOOT.1 — PASS

Learner đã:

- chạy baseline bằng `go run ./cmd/bot` và quan sát deterministic runtime output (output runtime tất định);
- phân biệt được output ranking theo công thức hiện tại với business truth (sự thật kinh doanh);
- tạo intentional regression (hồi quy lỗi có chủ đích) ở tie-break (quy tắc phá hòa);
- quan sát `TestSameInputSameOutput` và eval case phát hiện `[b a]` thay vì `[a b]`;
- phân loại đúng đây là test assertion failure (lỗi assertion của test), không phải environment hay compile failure;
- sửa tối thiểu để đưa behavior về đúng expectation;
- giải thích được `go run` chứng minh Bot chạy/runtime behavior hiện tại, còn `go test ./...` PASS chỉ chứng minh code thỏa các expectation được test hiện tại kiểm tra; không chứng minh sản phẩm đứng #1 là sản phẩm Affiliate tốt nhất.

BOOT.1 PASS chỉ xác nhận tooling/test semantics capability (năng lực công cụ/ngữ nghĩa kiểm thử); không tạo E1, M00 PASS hoặc quyền automation.

## O00.1 — ORIENTATION COMPLETED

Learner đã chạy:

```bash
cd lab/mission-runtime
go run ./cmd/demo O00
```

và quan sát được chuỗi artifact (đối tượng dữ liệu):

```text
Observation
→ DecisionPacket
→ AdvisorOutput
→ ActionIntent
→ PolicyDecision
→ ApprovalReview
→ ExecutionRecord
→ OutcomeRecord
→ EvaluationRecord
```

Learner đã giải thích được các boundary (ranh giới) chính:

- `Observation (quan sát)` là dữ liệu/sự kiện được ghi nhận, **không phải recommendation (khuyến nghị)**;
- `DecisionPacket (gói quyết định)` là quyết định/kết luận dựa trên evidence (bằng chứng);
- `ActionIntent = PROPOSAL_ONLY (chỉ là đề xuất)` nghĩa là Bot **chưa có execution permission (quyền thực thi)**;
- `ExecutionRecord = DRY_RUN_ONLY (chỉ chạy mô phỏng)` nghĩa là execution (thực thi) chỉ được mô phỏng và không tạo hành động thật; dry-run/live execution là trục khác với synthetic/real evidence;
- `external_side_effects=false (không có tác động bên ngoài)` nghĩa là runtime (chương trình khi chạy) **đã thực sự chạy**, nhưng không publish (đăng), send (gửi), spend (chi tiền), mutate account (sửa tài khoản) hoặc gọi write API (API ghi) ra hệ thống bên ngoài.

O00.1 chỉ hoàn thành orientation (định hướng). Nó dùng synthetic data (dữ liệu mô phỏng), không tạo E1 evidence, Reality PASS, Mission PASS hoặc quyền automation.

## Kết quả từ repo lịch sử

- `BOOT.1 — Chạy, sửa và kiểm thử Bot`: learner từng có **historical PASS credit (kết quả PASS lịch sử có thể được công nhận)**, nhưng hiện learner đã **thực sự học lại và PASS BOOT.1 trong repo v2**, nên không cần dùng historical credit cho trạng thái hiện tại.
- Historical credit không tạo E1 evidence (bằng chứng E1), M00 Reality PASS hoặc quyền automation (tự động hóa).

## Đường học hiện tại

```text
BOOT.0 PASS
→ BOOT.1 PASS
→ O00.1 ORIENTATION COMPLETED
→ M00.1
→ M00.2
→ M00.3
→ M00 PASS
→ M01 ...
```

`BOOT.0` và `BOOT.1` là onboarding/tooling (làm quen môi trường/công cụ), không phải Mission PASS gate. `O00.1` là orientation (định hướng) bằng dữ liệu synthetic (mô phỏng), không tạo Mission PASS. Mission đầu tiên tạo Reality evidence (bằng chứng thực tế) là M00.

## Quy tắc cập nhật tiến độ

- Chỉ đánh dấu một bài trong repo v2 là đã hoàn thành khi learner thực sự học/thực hành bài đó hoặc chủ động áp dụng historical credit khi curriculum cho phép.
- Progress (tiến độ) của Mission chỉ được cập nhật khi có evidence (bằng chứng) tương ứng.
- CI xanh, fixture PASS, sample/synthetic data (dữ liệu mẫu/mô phỏng) hoặc trạng thái authoring `ready` không tự tạo learner PASS hay Reality PASS.
