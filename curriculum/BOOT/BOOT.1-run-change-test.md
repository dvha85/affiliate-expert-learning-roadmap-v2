# BOOT.1 — Chạy, sửa và kiểm thử Bot

**Vai trò:** Tooling bootcamp, không phải Mission PASS gate.

Nếu terminal/Git/Go environment chưa ổn định, hoàn thành `BOOT.0` trước. BOOT.1 không dùng việc debug môi trường để thay thế việc học test semantics.

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
- một thay đổi phải có failure case đủ mạnh để test có thể phát hiện regression;
- command failure do sai working directory/toolchain khác với test failure do behavior không đúng expectation.

## Failure classification tối thiểu

Khi gặp FAIL, trước tiên phân loại:

```text
environment/tooling failure
!= compile/build failure
!= test assertion failure
!= business/evidence limitation
```

Không sửa business logic chỉ để chữa một lỗi environment.

## Credit

Nếu đã hoàn thành pilot Bài 0.1 ở repo lịch sử thì không cần học lại BOOT.1. Credit này chỉ xác nhận tooling capability; nó không tạo E1, M00 PASS hoặc quyền automation.

## PASS

Bạn chạy được baseline/test, tạo được một intentional failure hợp lệ, sửa tối thiểu để PASS lại, và giải thích được test evidence chứng minh gì/không chứng minh gì.
