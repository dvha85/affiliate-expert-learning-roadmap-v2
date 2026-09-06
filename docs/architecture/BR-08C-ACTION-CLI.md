# BR-08c — M03 qua learner CLI

BR-08b đã review/merge #48 `56a1ccf`, CI 4/4 PASS. BR-08c nối `cmd/bot` → `internal/app` → `core/m03` → contracts; không gọi harness hoặc copy validation. Store/record/import thuộc BR-08d/BR-10, chưa triển khai.

## Chạy trong cùng Bot

Từ lab/affiliate-bot, build binary (Windows dùng bot.exe):

```sh
go build -o bot ./cmd/bot
./bot action validate ../mission-runtime/testdata/m03-action.json ../mission-runtime/testdata/m03-outcome.json
```

Hai file trên là fixture synthetic để thử lệnh, không phải evidence cá nhân. Sau khi thử, truyền đường dẫn ACTION.json và OUTCOME.json của người học. Không chỉnh fixture chung hoặc PROGRESS để lấy PASS. CLI nhận đúng hai file, không dùng dữ liệu hard-code khi validate.

| Exit | status | stdout/stderr |
|---|---|---|
| 0 | VALID | Một JSON envelope gồm command, status, artifact.action/outcome/result và execution_permitted=false; stderr rỗng |
| 2 | USAGE_ERROR | JSON không artifact; stderr chỉ cách dùng |
| 1 | IO_ERROR | JSON không artifact; stderr mô tả lỗi đọc file |
| 1 | INVALID_SCHEMA hoặc semantic status như BROKEN_LINK/MEASUREMENT_WINDOW_OPEN | JSON không artifact; stderr nêu trạng thái |
| 1 | OUTPUT_INVALID / lỗi ghi stdout | Không xuất success artifact sai; nếu stdout hỏng không bảo đảm nhận đủ JSON |

Exit của `go run` có thể được Go wrapper chuyển thành 1; kiểm exit 2 bằng binary đã build. `VALID` chỉ là schema/semantic giữa pair, không resolve decision trong store, không chứng minh báo cáo đúng, không approval/execute/operated. Artifact output được kiểm schema lại trước serialize. Không tạo directory/history hoặc sửa file đầu vào.

Các lệnh file-input/ranking và history capture/list/replay/decision cũ giữ đường xử lý hiện hữu. Từ khóa action nay là subcommand; nếu có file observation tên action trong cwd, dùng `./action` để chỉ file. Không thay output/exit của lệnh legacy.

## Regression

internal/app/action_test.go: valid, schema sai, duplicate, measurement window mở, usage và missing file; kiểm status, artifact vắng khi lỗi và input bytes không đổi. cmd/bot/action_cli_test.go build binary thật rồi kiểm dispatch/exit 0/1/2/stdout/stderr, fixture canonical, no store write và đường ranking/history legacy. Fixture expected statuses độc lập với core; tests history và harness cũ tiếp tục chạy.

BR-08 tổng thể chưa DONE: store seam và hướng dẫn tích hợp/acceptance BR-08d/e còn mở. Không thêm action record/outcome import/execution hoặc nguồn affiliate live.
