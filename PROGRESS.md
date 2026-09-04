# Tiến độ — trạng thái người học

Repo v2 bắt đầu theo trục Mission mới. File này chỉ ghi **tiến độ học thực tế trong repo v2**; kết quả từ repo lịch sử được tách riêng để tránh nhầm với bài đã học trong chương trình mới.

## Trạng thái học v2 hiện tại

- Trạng thái: **NOT STARTED (chưa bắt đầu)**
- Mission hiện tại: **chưa có**
- Bài đã hoàn thành trong repo v2: **chưa có**
- Bài tiếp theo: **BOOT.0 — Chuẩn bị môi trường cho người mới**
- Sau BOOT.0: **BOOT.1 → O00.1 → M00.1 → M00.2 → M00.3**
- Trạng thái Reality (thực tế) của M00: **chưa bắt đầu**

## Kết quả từ repo lịch sử

- `BOOT.1 — Chạy, sửa và kiểm thử Bot`: learner đã có **historical PASS credit (kết quả PASS lịch sử có thể được công nhận)** nếu muốn áp dụng cơ chế carry-over (chuyển kết quả) của curriculum.
- Historical credit không có nghĩa learner đã học `BOOT.1` trong repo v2.
- Nếu learner chọn học repo v2 từ đầu, vẫn học `BOOT.0` và `BOOT.1` theo thứ tự hiện hành.
- Dù có áp dụng historical credit, kết quả BOOT.1 chỉ xác nhận tooling capability (năng lực sử dụng công cụ); không tạo E1 evidence (bằng chứng E1), M00 Reality PASS hoặc quyền automation (tự động hóa).

## Đường học bắt đầu

```text
BOOT.0
→ BOOT.1
→ O00.1
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
