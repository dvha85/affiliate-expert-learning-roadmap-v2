# E-01 — Bảo toàn timestamp M10/M11

Audit #43 đã review/merge `28073d8`. E-01 IN_REVIEW tại [PR #44](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/44), commit triển khai `3fa923e`. Tests/vet ba module, 8 validators và 10 Python regressions local PASS. Chưa merge; BR-03 chưa DONE, E-02/E-03 vẫn mở.

Đổi bốn output Format(RFC3339) thành Format(RFC3339Nano): WindowStartedAt khi normalize M10/M11 và ExpiresAt khi authorize M10/M11. Giữ nguyên phép chọn minimum expiry và guard hết hạn; không làm tròn nới quyền, không đổi schema/hash hoặc rewrite file cũ. Timestamp không có phần lẻ giữ cùng wire format; file cũ đã mất precision không thể phục dựng ngầm.

Regression `timestamp_precision_test.go`: trước sửa FAIL ở authorization còn 1ns nhưng serialize thành hết hạn và window reset mất phần lẻ. Sau sửa kiểm M10 cost expiry/M11 health expiry tại -1ns, đúng ngưỡng, +1ns; authorization hợp lệ serialize rồi qua decoder canonical trước local executor; đọc lại ledger đã persist. Window giữ phần lẻ qua ba lần reset và persist/load; không reset sớm 1ns ở chu kỳ kế tiếp. Fixtures whole-second và tests expiry nguồn hiện hữu tiếp tục được chạy.

Lệnh từ lab/mission-runtime: `go test ./cmd/demo -run TestTimestamp -count=1`, `go test ./...`, `go vet ./...`. Chỉ sandbox tạm, không live. Đây là sửa precision, không chứng minh trusted clock/provenance, crash consistency hoặc toàn bộ branch coverage E-02.
