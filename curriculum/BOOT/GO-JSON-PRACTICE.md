# BR-07 — Go/JSON tối thiểu để mở rộng cùng Bot

Bài tham khảo kỹ thuật sau [quickstart BR-05](QUICKSTART.md), không thêm Mission/PASS gate. Code tại `lab/affiliate-bot/internal/learning/`, cùng module Bot; **không phải adapter platform hoặc OutcomeRecord**. Chưa có ID/window/provenance đủ cho M03 và chưa ghi vào history. Tất cả dữ liệu giả lập, không cần API key.

## 1. Chạy trước khi đọc đáp án

Từ repo root:

```text
cd lab/affiliate-bot
go test ./internal/learning -v
```

Kỳ vọng các test `TestNullIsNotZero`, `TestReportRejectsBadInput`, `TestNormalizeStatus`, `TestCommittedReportFixture` PASS. Trước khi đọc code, viết dự đoán: `valid_orders:null`, `valid_orders:0`, field bị bỏ hẳn có giống nhau không? Chuỗi `"0"` có phải số 0 không?

## 2. Đọc struct, tag, pointer và map

Trong `internal/learning/report.go`:

```go
type Report struct {
    SourceRef   string `json:"source_ref"`
    Status      string `json:"status"`
    ValidOrders *int   `json:"valid_orders"`
}
```

`struct` nhóm field; JSON tag nối tên Go `ValidOrders` với key `valid_orders`. `*int` là pointer: nil biểu diễn chưa có số, pointer tới 0 biểu diễn có số bằng 0. Không đọc `*report.ValidOrders` khi pointer nil. Với decoder của bài này, missing và null đều nil; muốn phân biệt riêng hai trường hợp phải kiểm raw field presence như boundary BR-03, không thể suy lại từ pointer sau decode. Xem [encoding/json chính thức](https://pkg.go.dev/encoding/json#Unmarshal).

`Metrics` trả `map[string]int`, ánh xạ tên metric tới số đã biết. Map lookup cần cả hai kết quả:

```go
value, observed := Metrics(report)["valid_orders"]
```

Nếu chỉ đọc `value`, key không có cũng trả 0 và bạn sẽ nhầm missing với observed zero. `[]string` là slice các chuỗi; bảng test `[]struct{...}` chứa nhiều input/expected để chạy cùng một assertion. Slice khác map: slice giữ thứ tự phần tử, map dùng key.

## 3. Hàm, error và đọc file

`ReadReport(path string) (Report, error)` nhận path, trả object hoặc lỗi. Nó gọi `os.ReadFile`, strict decode, kiểm source/status/số âm. Mỗi bước `if err != nil { return ..., err }` giữ lỗi cho caller; không bỏ lỗi rồi tự trả zero value như dữ liệu thật.

`NormalizeStatus` chỉ chuẩn hóa chữ hoa/khoảng trắng của enum được phép trong fixture. `"success"` phải lỗi, không map tùy tiện sang PAID. Đây không phải bảng mapping status của bất kỳ chương trình affiliate nào.

`contracts.DecodeStrict` giữ exact field name, từ chối field lạ/duplicate/trailing JSON trước khi decoder làm mất thông tin. Bài này không có canonical schema riêng; không dùng nó như đường tắt bỏ schema tại boundary thật của M03/M04/M06.

## 4. Bài A — Chỉnh input và đọc lỗi

Vị trí: `internal/learning/testdata/report.json`. Chỉ làm trên bản bài tập, kiểm git status trước khi sửa.

| Trước | Sau có chủ đích | Kỳ vọng |
|---|---|---|
| `"valid_orders": 0` | `"valid_orders": "0"` | Decode báo không đưa string vào int được; không phải observed zero |
| `"status": " pending "` | `"status": "success"` | unknown fixture status, không tự cấp trạng thái thành công |
| `"valid_orders": 0` | `"valid_orders": null` | ReadReport vẫn hợp lệ nhưng test fixture mong observed zero phải FAIL |

Sau mỗi lần Save, chạy:

```text
go test ./internal/learning -run TestCommittedReportFixture -count=1 -v
```

Ghi error đầu tiên và giải thích **vì sao** FAIL. Đổi file lại như ban đầu rồi chạy full tests về PASS. Không sửa expected để giấu lỗi, không commit mutation gây FAIL.

## 5. Bài B — Thêm field clicks và assertion riêng

Đừng chép toàn bộ solution. Đầu tiên tự viết test trong `internal/learning/report_test.go`: input có `clicks:0` phải thành metric observed zero. Chạy test để thấy decoder từ chối field chưa được hỗ trợ. Trạng thái trước: struct chỉ có SourceRef, Status, ValidOrders.

Sau đó sửa đúng ba vị trí trong `report.go`:

1. Thêm vào struct: `Clicks *int` với JSON tag `json:"clicks"`.
2. Sau kiểm ValidOrders âm, thêm cùng kiểu kiểm cho Clicks âm; không chấp nhận negative hoặc string.
3. Trong Metrics, chỉ thêm key clicks khi pointer không nil:

```go
if report.Clicks != nil {
    metrics["clicks"] = *report.Clicks
}
```

Mẫu assertion để tự hoàn thiện trong test mới (package learning):

```go
report, err := ReadReport(reportFile(t, `{"source_ref":"synthetic","status":"pending","clicks":0}`))
if err != nil { t.Fatal(err) }
value, observed := Metrics(report)["clicks"]
if !observed || value != 0 { t.Fatalf("clicks zero must be observed") }
```

Test cũ `TestReportRejectsBadInput` có case `clicks:0` để minh họa unknown field; **chỉ sau khi chính thức thêm field vào contract bài tập**, đổi riêng case unknown đó sang `typo_clicks:0`. Không bỏ toàn bộ test unknown fields. Tự thêm case clicks null, absent, -1 và `"0"`; giải thích vì sao null/absent không có key metrics.

```text
go test ./internal/learning -count=1 -v
go test ./...
git diff -- internal/learning
```

Đây là bài thay đổi contract có chủ đích với tests trước/sau. Không đổi canonical OutcomeRecord/schema hoặc ghi vào history; tích hợp đó là BR-10. Mẫu repo giữ phiên bản **trước** có ValidOrders để người học tự làm bước thêm Clicks.

## 6. Bài C — Lỗi logic mà compiler không bắt

Vị trí: nhánh `if report.ValidOrders != nil` trong Metrics. Tạm viết map luôn có `valid_orders:0`, không quan tâm pointer. Chạy `go test ./internal/learning -run TestNullIsNotZero -count=1 -v`: case null/absent phải FAIL. Compiler thành công không có nghĩa logic đúng. Khôi phục hàm bằng cách sửa lại đúng điều kiện, không reset các thay đổi bài B của bạn.

## 7. Tiền đề cho adapter nào?

| Bài / kỹ năng | Dùng trước bước nào |
|---|---|
| A + pointer/map lookup + source/error | M03/BR-10: nhập report, không biến missing thành 0, vẫn cần effect_ref/window checks |
| Required/raw vs typed + string/array | M04/BR-11: output JSON, reason/evidence IDs; schema hợp lệ chưa chứng minh grounding |
| NormalizeStatus + error propagation | M06/BR-13: normalize dữ liệu nguồn có quyền truy cập; status/field mapping phải dựa contract thật |
| B/C + assertion độc lập | Mọi adapter: kiểm input/output và failure cases, không so một hàm với chính nó để gọi là conformance |

Tham khảo [Go pointers](https://go.dev/tour/moretypes/1) và [viết test chính thức](https://go.dev/doc/tutorial/add-a-test). Kết quả học cần lưu: dự đoán trước chạy, diff tự viết, output FAIL/PASS và giải thích. Tests repo PASS không chứng minh học viên đã tự sửa adapter; reviewer cần xem phần tự làm và explain-back. Không cập nhật PROGRESS.md chỉ vì CI xanh.
