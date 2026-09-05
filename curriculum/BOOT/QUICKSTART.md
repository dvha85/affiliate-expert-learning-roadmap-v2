# BR-05 — Từ máy mới đến chạy, sửa và test Bot

Hướng dẫn thực hành BOOT.0/BOOT.1, **không thêm Mission hoặc PASS gate**. Chưa cần tài khoản affiliate, API key, VPS, Docker hay n8n. Bạn sẽ chạy fixture/test, không phải bot có doanh thu hoặc quyền đăng bài.

## 1. Cài công cụ

Dùng nguồn chính thức: [Git](https://git-scm.com/downloads/), [Go](https://go.dev/doc/install), [VS Code](https://code.visualstudio.com/download). Chọn đúng hệ điều hành/kiến trúc. Go phải đáp ứng `lab/affiliate-bot/go.mod` (hiện **1.27**); không hạ go.mod để né lỗi. macOS dùng gói cài chính thức; Windows dùng MSI; Linux làm theo hướng dẫn của Go phù hợp hệ thống. Không chạy lệnh xóa bản cài cũ khi chưa hiểu đường dẫn/quyền quản trị.

VS Code là editor tùy chọn, không cần tài khoản AI/extension. Mở lại terminal sau khi cài để nhận PATH mới. Terminal nhận lệnh; editor sửa file. Gõ từng dòng rồi đợi kết quả, không gõ ký hiệu `$` hoặc các dòng output.

| Nền tảng | Mở terminal và kiểm tra | Phạm vi kiểm chứng |
|---|---|---|
| macOS Apple Silicon | Terminal hoặc VS Code → Terminal → New Terminal; `pwd`, `git --version`, `go version` | Smoke clone sạch/cache rỗng trên Darwin arm64, Go 1.27.0; chưa chạy lại installer trên máy trắng |
| Linux | Terminal; `pwd`, `git --version`, `go version` | CI Ubuntu chạy script smoke; chỉ claim tại commit CI đã PASS, không claim mọi distro |
| Windows | PowerShell; `Get-Location`, `git --version`, `go version` | Lệnh tham khảo; **chưa smoke Windows**, chưa công bố hỗ trợ đầy đủ |

Ví dụ đã kiểm: `go version go1.27.0 darwin/arm64`. Số bản vá/kiến trúc máy khác có thể khác. “command not found”/“not recognized” là lỗi cài đặt hoặc PATH, không phải business logic.

## 2. Clone đầy đủ repo

Chọn thư mục ghi được bằng file manager; mở nó trong VS Code → File → Open Folder, rồi Terminal → New Terminal. Không chọn thư mục repo đã có dữ liệu riêng để ghi đè.

macOS/Linux terminal và Windows PowerShell dùng cùng các lệnh:

```text
git clone https://github.com/dvha85/affiliate-expert-learning-roadmap-v2.git
cd affiliate-expert-learning-roadmap-v2
git status
git rev-parse --show-toplevel
```

Kỳ vọng `working tree clean`; đường dẫn cuối là **repo root**, có README.md, curriculum, contracts, lab. Nếu yêu cầu credentials bất thường với repo công khai, kiểm URL/kết nối, không dán token vào lệnh hoặc evidence. Nếu đã có clone chứa thay đổi cá nhân, giữ nguyên và làm bài ở clone mới, không reset.

## 3. Root, module và file

```text
affiliate-expert-learning-roadmap-v2/   ← repo root, không có go.mod tại đây
├── contracts/                        ← schema module, phải giữ trong clone
├── lab/affiliate-bot/                 ← module của Bot đang học
│   ├── go.mod
│   ├── cmd/bot/main.go                ← logic baseline và CLI
│   ├── cmd/bot/main_test.go           ← tests
│   └── data/sample-observations.json  ← fixture mặc định
└── lab/mission-runtime/               ← oracle/eval, không phải Bot thứ hai
```

`cd lab/affiliate-bot` từ root vào module. `cd ../..` từ đó về root. `./...` là các package dưới module hiện tại. Không chạy `go mod init` ở root để chữa lỗi sai thư mục. Không copy riêng affiliate-bot vì local replace cần `../../contracts`.

## 4. Chạy và test

Từ **repo root**:

```text
cd lab/affiliate-bot
go version
go mod download
go run ./cmd/bot
go test ./...
go vet ./...
```

Lần tải dependency đầu cần mạng. go.mod/go.sum pin version/checksum; không đổi sang latest hoặc tắt kiểm checksum để né lỗi. Các dòng chính mong đợi từ `go run`:

```text
Affiliate Bot v0.1 — deterministic baseline
Loại bằng chứng (Evidence mode): synthetic
1. Product B [B] | điểm (score)=9.60
2. Product A [A] | điểm (score)=8.00
3. Product C [C] | điểm (score)=8.00
Trạng thái quyết định (Decision state): RANK_SCENARIO
```

`go test` có dòng `ok .../cmd/bot`, có thể kèm `(cached)`; `go vet` thành công thường không in gì. Score chỉ là price × commission_rate trên fixture, không phải conversion/ROI hoặc sản phẩm affiliate tốt nhất.

## 5. Sửa và lưu — intentional FAIL

Trong VS Code mở **repo root**, chọn `lab/affiliate-bot/cmd/bot/main.go`. Tìm đúng một dòng sau, trong nhánh tie-break của `sort.Slice(result.Ranked, ...)`:

```go
return result.Ranked[i].ProductName < result.Ranked[j].ProductName
```

Chỉ đổi `<` thành `>`, không sửa nhánh Score/ProductID. File + dòng logic là vị trí ổn định; script smoke ghi thêm số dòng tại commit được thử. Lưu bằng Cmd+S (macOS) hoặc Ctrl+S (Windows/Linux). Từ `lab/affiliate-bot`:

```text
go test ./cmd/bot -run TestSameInputSameOutput -count=1
```

Phải FAIL chứa `expected deterministic A-first tie break`; code vẫn compile. Đổi `>` lại `<`, lưu, chạy:

```text
go test ./...
git diff -- cmd/bot/main.go
git status
```

Test trở lại `ok`, diff của dòng vừa sửa biến mất. Giữ các thay đổi khác nếu có; không reset cả repo. **Không commit/push intentional regression.** Đọc [BOOT-REFERENCE mục 7](BOOT-REFERENCE.md#7-intentional-fail--vì-sao-fail-có-thể-là-tín-hiệu-tốt) để giải thích test chứng minh gì.

Tạo file bài tập: Explorer → New File → tên đúng đuôi `.json`/`.md`, nhập nội dung rồi Save. Kiểm vị trí trong Explorer, tránh `input.json.txt`. Không lưu secret hoặc private export vào repo.

## 6. O00 và đường học tiếp theo

Từ `lab/affiliate-bot` sau khi test PASS:

```text
cd ../mission-runtime
go test ./...
go run ./cmd/demo O00
```

Output JSON phải có `external_side_effects=false`, `final_state=DRY_RUN_ONLY`. O00 chỉ mô phỏng toàn vòng, không Mission PASS. `cd ../..` về root, theo [CURRICULUM](../../CURRICULUM.md) rồi M00; không lấy fixture thay E1 thật.

## 7. Lỗi thường gặp

| Lỗi | Kiểm tra / xử lý |
|---|---|
| git/go command not found hoặc not recognized | Mở lại terminal; macOS/Linux: `command -v go`, PowerShell: `Get-Command go`. Quay lại trang cài chính thức nếu thiếu; không sửa code |
| go.mod not found / directory prefix does not contain main module | `pwd` hoặc `Get-Location`; vào đúng lab/affiliate-bot từ root, không go mod init |
| requires go >= 1.27 | Kiểm go version và cài toolchain đáp ứng go.mod; không hạ go.mod |
| replacement directory ../../contracts does not exist | Clone toàn repo vào thư mục mới, không copy module riêng |
| destination path already exists | Kiểm đó có phải clone có dữ liệu riêng không; chọn thư mục mới, không xóa mù quáng |
| download timeout / checksum mismatch | Kiểm mạng/proxy; không tắt verification hoặc bỏ go.sum |
| cache permission denied | Kiểm `go env GOCACHE`; dùng cache riêng bạn sở hữu theo chính sách máy, không sudo go test |
| no such file cho input | Path tính từ thư mục đang đứng; kiểm data/sample-observations.json bằng Explorer |
| sửa file mà test không đổi | Save, kiểm đúng clone/module/file, dùng -count=1; không sửa expected để lấy PASS |
| compiler error thay vì tie-break FAIL | Kiểm chỉ đổi toán tử đúng dòng, không xóa ngoặc; ghi lỗi đầu tiên |

Khi hỏi hỗ trợ, gửi lệnh, working directory, git status, go version, expected/observed error đã loại secret. Không gửi token/cookie/dữ liệu khách hàng.

## 8. Bằng chứng và tự kiểm

Hoàn tất vòng BOOT khi bạn chỉ được repo/module, chạy baseline/test, tạo đúng assertion FAIL và sửa về PASS, giải thích synthetic != market truth. Không tự claim mọi bug hết hoặc mọi người mới đều làm được.

Bảo trì có thể dùng [Python 3](https://www.python.org/downloads/) chạy `python3 scripts/smoke_quickstart.py` từ root (Windows tham khảo `py -3 scripts/smoke_quickstart.py`). Script tạo clone/cache tạm, chạy vòng run/test/FAIL/fix/O00, kiểm working tree sạch rồi dọn đúng thư mục tạm; không sửa clone nguồn. Python không cần để chỉ chạy Bot, nhưng CI validators dùng Python 3.12.

[Evidence và phạm vi nền tảng](../../docs/plans/evidence/BR-05-QUICKSTART.md) phân biệt clone sạch/cache rỗng với máy trắng chưa cài công cụ. Chưa có pilot người mới độc lập thì không claim onboarding hoàn chỉnh trên mọi OS.

Sau khi BOOT ổn định, dùng [bài Go/JSON BR-07](GO-JSON-PRACTICE.md) để luyện pointer/null, map lookup, đọc file/error và viết test trước khi mở rộng adapter; đây là tài liệu hỗ trợ, không thay thứ tự Mission.
