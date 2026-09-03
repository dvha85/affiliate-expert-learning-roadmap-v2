# Đường học — chương trình theo Mission

Đây là đường học hiện hành. Không dùng lesson ID legacy (ID bài học cũ) từ repo cũ để quyết định bài tiếp theo.

## Bắt đầu

1. `BOOT.0` nếu máy chưa có environment/tooling (môi trường/công cụ) tối thiểu hoặc chưa quen terminal/Git/Go.
2. `BOOT.1` nếu chưa từng chạy/sửa/test Bot.
3. Chạy `O00.1` để nhìn toàn hệ thống bằng dữ liệu mô phỏng; O00 chỉ orientation (định hướng), không tạo PASS.
4. `M00.1 → M00.2 → M00.3`, hoàn thành evidence/artifact của M00 trước M01.
5. Khi M00 PASS: `M01.1 → M01.2 → M01.3 → M01.4`.
6. Khi M01 PASS: `M02.1 → M02.2 → M02.3 → M02.4`.
7. Khi M02 PASS: `M03.1 → M03.2 → M03.3`.
8. Khi M03 PASS: `M04.1 → M04.2 → M04.3`.
9. Khi M04 PASS: `M05.1 → M05.2 → M05.3`.
10. Khi M05 PASS: `M06.1 → M06.2 → M06.3`.
11. Khi M06 PASS: `M07.1 → M07.2 → M07.3`.

`BOOT.0` và `BOOT.1` là onboarding/tooling credit (kết quả làm quen môi trường/công cụ), không thay thế evidence hay Mission PASS. O00 chỉ là bản đồ hệ thống bằng synthetic data (dữ liệu mô phỏng).

M01–M07 hiện **authoring ready và learner-operable (sẵn sàng về nội dung và đường thực hành)**: có lesson, Mission contract, starter/checkpoints/evidence template, executable eval/runtime và CI guard. M06/M07 có workflow n8n thực tế nhưng learner vẫn phải cấu hình nguồn/credential an toàn và tự tạo Reality/Operated evidence.

## Vòng học của learner

```text
TRY (thử)
→ OBSERVE GAP (quan sát khoảng thiếu)
→ PULL 1–3 SMALL KNOWLEDGE CARDS (lấy 1–3 thẻ kiến thức nhỏ)
→ BUILD / APPLY (xây / áp dụng)
→ TEST FAILURE CASE (kiểm thử ca lỗi)
→ SAVE EVIDENCE (lưu bằng chứng)
→ EXPLAIN LIMITS (giải thích giới hạn)
→ NEXT MEASUREMENT (phép đo tiếp theo)
```

## Trục Mission

```text
O00 → M00 → M01 → M02 → M03 → M04 → M05 → M06 → M07 → M08 → M09 → M10 → M11
```

Authority (quyền hạn) tăng dần. Capability (năng lực) của Mission sau không được dùng để vượt evidence/safety gate của Mission trước.

Từ M02 trở đi, ID trong artifact phải resolve được về artifact trước đó khi contract yêu cầu. Schema-valid nhưng provenance bị đứt không tạo Reality/Operated PASS.

## Boundary quan trọng từ M03–M07

```text
M03: external action đầu tiên do human_only thực hiện
M04: AI tư vấn dựa trên evidence; không write tool
M05: evaluation → proposal → human review; không auto-apply
M06: automation tự động nhưng chỉ GET/HEAD trên allowlist, normalize về canonical Observation
M07: Agent + tool nhưng vẫn read-only; tool output là untrusted data
```
