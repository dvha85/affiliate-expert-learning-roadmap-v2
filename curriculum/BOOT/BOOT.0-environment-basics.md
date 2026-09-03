# BOOT.0 — Chuẩn bị môi trường cho người mới

**Vai trò:** onboarding (làm quen môi trường) kỹ thuật tối thiểu trước BOOT.1; không phải Mission PASS gate (cổng PASS của Mission).

## Mục tiêu

Người mới có thể tự đi từ máy cá nhân tới trạng thái chạy được repo mà chưa cần biết Go, Agent hay automation (tự động hóa).

```text
computer (máy tính)
→ terminal (dòng lệnh)
→ folder/path (thư mục/đường dẫn)
→ Git
→ clone repo (tải bản sao repo)
→ Go toolchain (bộ công cụ Go)
→ editor (trình soạn thảo)
→ environment variable / secret (biến môi trường / bí mật)
→ first command (lệnh đầu tiên)
```

## 1. Terminal và đường dẫn

Hiểu tối thiểu:

- terminal là nơi chạy command (lệnh);
- `pwd`/`cd` dùng để biết và đổi thư mục;
- repository là một folder (thư mục) có lịch sử Git;
- command phải chạy đúng working directory (thư mục làm việc).

Không copy command khi chưa biết mình đang đứng ở folder nào.

## 2. Git tối thiểu

Người học cần biết:

```text
git clone
git status
git pull
```

Chưa cần branch/rebase (nhánh/tái đặt commit) ở BOOT.0. Mục tiêu chỉ là lấy đúng source (mã nguồn) hiện hành và nhìn được trạng thái local (trên máy).

## 3. Go toolchain (bộ công cụ Go)

Từ repo root (thư mục gốc của repo), kiểm tra:

```bash
go version
```

Sau đó:

```bash
cd lab/affiliate-bot
go test ./...
```

Nếu `go` không tồn tại, dừng và cài đúng Go version mà `go.mod` yêu cầu trước khi tiếp tục. Không đổi `go.mod` chỉ để né lỗi môi trường.

## 4. Secret và environment variable (bí mật và biến môi trường)

Foundation (phần nền tảng) M00–M02 chưa cần API key. Tuy vậy người học phải biết boundary (ranh giới) ngay từ đầu:

```text
secret/API key (bí mật/khóa API) != source code (mã nguồn)
secret/API key != evidence (bằng chứng)
secret/API key != file được commit
```

Không commit token, cookie, credential (thông tin xác thực), account export (dữ liệu xuất từ tài khoản) hoặc personal/customer data (dữ liệu cá nhân/khách hàng) vào repo. Nếu Mission sau cần credential, dùng secret store (kho bí mật) hoặc environment variable phù hợp.

## 5. Failure-first onboarding (làm quen theo hướng quan sát lỗi trước)

Nếu một command fail (lệnh thất bại), ghi lại:

```text
command: lệnh đã chạy
working_directory: thư mục làm việc
expected: kết quả mong đợi
observed_error: lỗi quan sát được
what_i_checked: những gì đã kiểm tra
```

Không sửa ngẫu nhiên nhiều thứ cùng lúc. Mỗi lần chỉ thay một giả thuyết nhỏ rồi chạy lại.

## PASS

Bạn có thể:

1. chỉ ra repo local đang ở đâu;
2. chạy `git status`;
3. chạy `go version`;
4. vào `lab/affiliate-bot` và chạy `go test ./...`;
5. giải thích vì sao credential (thông tin xác thực) không được commit vào repo.

PASS BOOT.0 chỉ chứng minh environment/tooling basics (kiến thức môi trường/công cụ cơ bản), không tạo E1 evidence hoặc Mission PASS.
