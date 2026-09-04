# BOOT Reference — tài liệu tham khảo trước O00.1

**Vai trò:** tài liệu recap/reference (ôn tập/tham khảo) cho người học sau `BOOT.0` và `BOOT.1`; **không phải bài học mới và không tạo PASS gate mới**.

Tài liệu này tóm tắt vòng hướng dẫn thực hành đầu tiên: chuẩn bị môi trường, chạy Bot, chạy test, tạo intentional failure (lỗi có chủ đích), đọc FAIL, sửa tối thiểu và hiểu giới hạn của test evidence (bằng chứng kiểm thử).

## 1. Mục tiêu sau BOOT

Người học cần có thể tự đi qua vòng kỹ thuật tối thiểu:

```text
terminal (dòng lệnh)
→ đúng working directory (thư mục làm việc)
→ Git
→ Go toolchain (bộ công cụ Go)
→ baseline run (chạy mốc ban đầu)
→ test
→ intentional FAIL (FAIL có chủ đích)
→ minimal fix (sửa tối thiểu)
→ PASS
→ giải thích bằng chứng chứng minh gì / không chứng minh gì
```

BOOT chỉ xác nhận environment/tooling basics (nền tảng môi trường/công cụ). Nó **không tạo E1 evidence (bằng chứng E1), Mission PASS hoặc quyền automation (tự động hóa)**.

## 2. Kiểm tra môi trường tối thiểu

Từ terminal:

```bash
pwd
git --version
go version
```

Ý nghĩa:

- `pwd`: biết chính xác đang đứng ở thư mục nào;
- `git --version`: xác nhận Git khả dụng;
- `go version`: xác nhận Go toolchain khả dụng.

Không copy command (lệnh) khi chưa biết working directory hiện tại.

Repo hiện dùng Go version theo `lab/affiliate-bot/go.mod`; nếu `go` không tồn tại hoặc version không phù hợp, xử lý environment trước. Không sửa business logic hoặc `go.mod` chỉ để né lỗi môi trường.

## 3. Mở repo bằng VS Code

Có thể dùng VS Code làm editor (trình soạn thảo) chính.

Mở repo root, sau đó dùng:

```text
Terminal → New Terminal
```

Từ repo root:

```bash
cd lab/affiliate-bot
pwd
git status
go test ./...
```

Nếu `go test ./...` chạy được, environment và working directory cơ bản đã đúng.

## 4. Chạy Bot baseline

Từ `lab/affiliate-bot`:

```bash
go run ./cmd/bot
```

Baseline hiện đọc fixture (dữ liệu kiểm thử) mặc định và in kết quả xếp hạng scenario (kịch bản) theo công thức hiện tại.

Điều cần hiểu:

```text
go run ./cmd/bot
→ chứng minh runtime behavior (hành vi khi chạy) hiện tại có thể quan sát được
```

Nó **không chứng minh**:

- dữ liệu thị trường là thật;
- business decision (quyết định kinh doanh) là đúng;
- sản phẩm đứng #1 là sản phẩm Affiliate tốt nhất;
- Bot được phép recommend (khuyến nghị), approve (phê duyệt) hoặc execute (thực thi).

## 5. Cách đọc baseline hiện tại

Baseline dùng công thức:

```text
score = price × commission_rate
```

Ví dụ fixture hiện tại có thể tạo kết quả:

```text
Product B → 9.60
Product A → 8.00
Product C → 8.00
```

Kết luận hợp lệ chỉ là:

```text
Product B đứng #1 theo input + công thức + version hiện tại.
```

Không được suy rộng thành:

```text
Product B là sản phẩm Affiliate tốt nhất.
```

Vì baseline chưa tự chứng minh các yếu tố như demand (nhu cầu), product–audience fit (độ phù hợp sản phẩm–khán giả), conversion (chuyển đổi), competition (cạnh tranh), refund risk (rủi ro hoàn tiền), seller quality (chất lượng người bán) hay content potential (tiềm năng nội dung).

## 6. Synthetic evidence không phải market reality

Nếu output ghi:

```text
Evidence mode: synthetic
```

thì input chỉ là synthetic/test fixture (dữ liệu mô phỏng/kiểm thử).

Nó phù hợp để chứng minh deterministic behavior (hành vi tất định), plumbing (luồng kỹ thuật) hoặc test expectation (kỳ vọng kiểm thử), nhưng không tự tạo market reality (thực tế thị trường).

```text
synthetic evidence
!= real evidence
!= market truth
```

## 7. Intentional FAIL — vì sao FAIL có thể là tín hiệu tốt

BOOT.1 yêu cầu một intentional failure (lỗi có chủ đích) để chứng minh test thực sự có khả năng phát hiện regression (hồi quy lỗi).

Ví dụ an toàn trong baseline: tie-break (quy tắc phá hòa) khi hai sản phẩm cùng score.

Behavior đúng có đoạn:

```go
return result.Ranked[i].ProductName < result.Ranked[j].ProductName
```

Trong bài thực hành, có thể tạm đổi thành:

```go
return result.Ranked[i].ProductName > result.Ranked[j].ProductName
```

Sau đó chạy:

```bash
go test ./...
```

Nếu test trả về FAIL kiểu:

```text
expected deterministic A-first tie break
...
expected ranking [a b], got [b a]
```

thì đây là **test assertion failure (lỗi assertion của test)** hợp lệ: code vẫn compile, test vẫn chạy, nhưng behavior không còn đúng với expectation.

Sau bài thực hành phải sửa lại đúng behavior ban đầu và chạy lại:

```bash
go test ./...
```

để trở về PASS. Không commit intentional regression lên repo.

## 8. Phân loại FAIL trước khi sửa

Khi command thất bại, phân loại trước:

```text
environment/tooling failure (lỗi môi trường/công cụ)
!= compile/build failure (lỗi biên dịch/build)
!= test assertion failure (lỗi kỳ vọng test)
!= business/evidence limitation (giới hạn business/bằng chứng)
```

Ví dụ:

- `go: command not found` → environment/tooling failure;
- code không compile → compile/build failure;
- test chạy nhưng `expected ... got ...` → test assertion failure;
- test PASS nhưng dữ liệu chỉ synthetic → business/evidence limitation vẫn còn.

Không sửa business logic để chữa lỗi environment.

## 9. `go test ./...` PASS chứng minh gì?

Kết luận đúng:

```text
go test ./... PASS
→ code hiện tại thỏa những expectation mà test hiện tại đang kiểm tra.
```

PASS không tự chứng minh:

```text
mọi bug đã được bắt
market data là thật
business decision là đúng
sản phẩm #1 là tốt nhất
Bot có quyền hành động
```

Một test suite chỉ bảo vệ những behavior/expectation mà nó thực sự encode (mã hóa thành test).

## 10. Secret / credential boundary

Không commit các giá trị như:

```text
API key
access token
auth token
cookie
password
credential
private account export
personal/customer data
```

vào source repository.

```text
secret/API key != source code
secret/API key != evidence
```

Nếu secret từng bị commit, việc xóa khỏi file hiện tại chưa chắc đã đủ vì nó có thể còn trong Git history (lịch sử Git). Credential đã lộ phải được coi là compromised (đã bị lộ/rủi ro), sau đó revoke/rotate (thu hồi/đổi) theo hệ thống tương ứng.

## 11. Checklist trước khi sang O00.1

Người học nên tự trả lời được:

1. Repo local nằm ở đâu?
2. `git status` dùng để quan sát gì?
3. `go version` chứng minh gì?
4. `go run ./cmd/bot` chứng minh gì và không chứng minh gì?
5. `go test ./...` PASS chứng minh gì và không chứng minh gì?
6. Vì sao intentional FAIL có giá trị?
7. Phân biệt được environment failure, compile failure và test assertion failure chưa?
8. Vì sao secret/credential không được commit?
9. Vì sao Product đứng #1 theo baseline chưa phải là sản phẩm Affiliate tốt nhất?

Khi các câu trên đã rõ, learner sẵn sàng sang `O00.1 — Safe System Walkthrough (đi một vòng hệ thống an toàn)` để nhìn toàn bộ hệ thống bằng synthetic data (dữ liệu mô phỏng) trước khi M00 bắt đầu thu thập real evidence (bằng chứng thật).
