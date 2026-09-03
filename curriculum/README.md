# Đường học — chương trình theo Mission

Đây là đường học hiện hành. Không dùng lesson ID legacy (ID bài học cũ) từ repo cũ để quyết định bài tiếp theo.

## Bắt đầu

1. `BOOT.0` nếu máy chưa có environment/tooling (môi trường/công cụ) tối thiểu hoặc chưa quen terminal/Git/Go.
2. `BOOT.1` nếu chưa từng chạy/sửa/test Bot.
3. `M00.1 → M00.2 → M00.3`.
4. Hoàn thành evidence/artifact (bằng chứng/artifact) của M00 trước khi sang M01.
5. Khi M00 PASS: `M01.1 → M01.2 → M01.3 → M01.4`.
6. Khi M01 PASS: `M02.1 → M02.2 → M02.3 → M02.4`.

`BOOT.0` và `BOOT.1` là onboarding/tooling credit (kết quả làm quen môi trường/công cụ), không thay thế E1 evidence hay Mission PASS.

M01 và M02 đã **authoring ready (sẵn sàng về nội dung)**, nhưng trạng thái này không bỏ qua learner gate (cổng của người học) của Mission trước. Người học vẫn tiến theo evidence (bằng chứng) thực tế, không theo độ xa mà repo đã author (soạn).

M02 `ready` chỉ được khôi phục sau corrective planned gate (cổng hiệu chỉnh khi còn planned) có executable eval/runtime PASS; điều này không tự tạo learner M01/M02 PASS.

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

Mỗi lesson/card (bài/thẻ kiến thức) phải phục vụ một câu hỏi cụ thể của Mission hiện tại. Không đọc theory (lý thuyết) chỉ để hoàn thành checklist.

## Quy tắc về mental model (mô hình tư duy)

Một implementation (triển khai) sớm có thể chỉ hỗ trợ subset (tập con) của semantics (ngữ nghĩa) đầy đủ. Khi đó bài học phải nói rõ:

```text
CURRENT IMPLEMENTATION LIMIT
!= FUNDAMENTAL SYSTEM LAW
(giới hạn triển khai hiện tại != quy luật nền tảng của hệ thống)
```

Không được làm biến mất một state/concept (trạng thái/khái niệm) chỉ vì runtime của Mission hiện tại chưa implement (triển khai) nó.

## Thứ tự Mission

```text
O00 → M00 → M01 → M02 → M03 → M04 → M05 → M06 → M07 → M08 → M09 → M10 → M11
```

Authority (quyền hạn) tăng dần. Capability (năng lực) của Mission sau không được dùng để vượt evidence/safety gate (cổng bằng chứng/an toàn) của Mission trước.
