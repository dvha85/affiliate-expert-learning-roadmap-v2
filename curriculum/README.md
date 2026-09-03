# Learner path — Mission-based curriculum

Đây là đường học hiện hành. Không dùng lesson ID legacy từ repo cũ để quyết định bài tiếp theo.

## Bắt đầu

1. `BOOT.0` nếu máy chưa có environment/tooling tối thiểu hoặc chưa quen terminal/Git/Go.
2. `BOOT.1` nếu chưa từng chạy/sửa/test Bot.
3. `M00.1 → M00.2 → M00.3`.
4. Hoàn thành evidence/artifact của M00 trước khi sang M01.
5. Khi M00 PASS: `M01.1 → M01.2 → M01.3 → M01.4`.
6. Khi M01 PASS: `M02.1 → M02.2 → M02.3 → M02.4`.

`BOOT.0` và `BOOT.1` là onboarding/tooling credit, không thay thế E1 evidence hay Mission PASS.

M01 và M02 đã **authoring ready**, nhưng trạng thái này không bỏ qua learner gate của Mission trước. Learner hiện tại vẫn học theo progress evidence thực tế, không theo độ xa mà repo đã author.

M02 `ready` chỉ được khôi phục sau corrective planned gate có executable eval/runtime PASS; điều này không tạo learner M01/M02 PASS tự động.

## Learner loop

```text
TRY
→ OBSERVE GAP
→ PULL 1–3 SMALL KNOWLEDGE CARDS
→ BUILD / APPLY
→ TEST FAILURE CASE
→ SAVE EVIDENCE
→ EXPLAIN LIMITS
→ NEXT MEASUREMENT
```

Mỗi lesson/card phải phục vụ một câu hỏi cụ thể của Mission hiện tại. Không đọc theory chỉ để hoàn thành checklist.

## Mental-model rule

Một implementation sớm có thể chỉ hỗ trợ subset của semantics đầy đủ. Khi đó lesson phải nói rõ:

```text
CURRENT IMPLEMENTATION LIMIT
!= FUNDAMENTAL SYSTEM LAW
```

Không được làm biến mất một state/concept chỉ vì runtime Mission hiện tại chưa implement nó.

## Mission sequence

```text
O00 → M00 → M01 → M02 → M03 → M04 → M05 → M06 → M07 → M08 → M09 → M10 → M11
```

Authority tăng dần. Capability của Mission sau không được dùng để vượt evidence/safety gate của Mission trước.
