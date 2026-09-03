# BOOT.0 — Chuẩn bị môi trường cho người mới

**Vai trò:** onboarding kỹ thuật tối thiểu trước BOOT.1; không phải Mission PASS gate.

## Mục tiêu

Người mới có thể tự đi từ máy cá nhân tới trạng thái chạy được repo mà chưa cần biết Go, Agent hay automation.

```text
computer
→ terminal
→ folder/path
→ Git
→ clone repo
→ Go toolchain
→ editor
→ environment variable / secret
→ first command
```

## 1. Terminal và đường dẫn

Hiểu tối thiểu:

- terminal là nơi chạy command;
- `pwd`/`cd` dùng để biết và đổi thư mục;
- repository là một folder có lịch sử Git;
- command phải chạy đúng working directory.

Không copy command khi chưa biết mình đang đứng ở folder nào.

## 2. Git tối thiểu

Learner cần biết:

```text
git clone
git status
git pull
```

Chưa cần branch/rebase ở BOOT.0. Mục tiêu chỉ là lấy đúng source hiện hành và nhìn được trạng thái local.

## 3. Go toolchain

Từ repo root, kiểm tra:

```bash
go version
```

Sau đó:

```bash
cd lab/affiliate-bot
go test ./...
```

Nếu `go` không tồn tại, dừng và cài đúng Go version mà `go.mod` yêu cầu trước khi tiếp tục. Không đổi `go.mod` chỉ để né lỗi môi trường.

## 4. Secret và environment variable

Foundation M00–M02 chưa cần API key. Tuy vậy learner phải biết boundary ngay từ đầu:

```text
secret/API key != source code
secret/API key != evidence
secret/API key != file được commit
```

Không commit token, cookie, credential, account export hoặc personal/customer data vào repo. Nếu Mission sau cần credential, dùng secret store hoặc environment variable phù hợp.

## 5. Failure-first onboarding

Nếu một command fail, ghi lại:

```text
command:
working_directory:
expected:
observed_error:
what I checked:
```

Không sửa ngẫu nhiên nhiều thứ cùng lúc. Mỗi lần chỉ thay một giả thuyết nhỏ rồi chạy lại.

## PASS

Bạn có thể:

1. chỉ ra repo local đang ở đâu;
2. chạy `git status`;
3. chạy `go version`;
4. vào `lab/affiliate-bot` và chạy `go test ./...`;
5. giải thích vì sao credential không được commit vào repo.

PASS BOOT.0 chỉ chứng minh environment/tooling basics, không tạo E1 evidence hoặc Mission PASS.
