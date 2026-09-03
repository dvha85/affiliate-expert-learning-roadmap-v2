# BOOT.1 — Chạy, sửa và kiểm thử Bot

**Vai trò:** Tooling bootcamp, không phải Mission PASS gate.

## Mục tiêu

Hiểu vòng kỹ thuật tối thiểu:

```text
baseline run
→ test expectation
→ intentional FAIL
→ minimal change
→ PASS
→ explain what the evidence proves and does not prove
```

## Điều phải hiểu

- `go run` chứng minh behavior runtime đang quan sát được;
- `go test` chứng minh code thỏa expectation hiện có;
- test PASS không chứng minh business decision đúng;
- synthetic fixture không chứng minh market reality;
- một thay đổi phải có failure case đủ mạnh để test có thể phát hiện regression.

## Credit

Nếu đã hoàn thành pilot Bài 0.1 ở repo lịch sử thì không cần học lại BOOT.1. Credit này chỉ xác nhận tooling capability; nó không tạo E1, M00 PASS hoặc quyền automation.
