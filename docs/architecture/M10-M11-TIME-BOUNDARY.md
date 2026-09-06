# BR-03d.4 — Bảo vệ phép tính thời gian M10/M11

IN_REVIEW tại [PR #42](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/42), commit triển khai `03abfba`. Tests/vet ba module, 8 validators và 10 Python regressions local PASS; chưa merge. BR-03d.2 đã merge #40 `1c974ca`; BR-03d.3 đã merge #41 `e4616a2`.

Bốn phép chuyển giây sang duration ở runtime được kiểm trước phép nhân: window M10/M11, health freshness và health expiry M11. Giới hạn là floor(MaxInt64 / 1 giây) = 9223372036 giây. Không clamp hoặc sửa grant/lease. Window vượt giới hạn trả ledger không đổi + false, gate từ chối LEDGER_MISMATCH; health age vượt giới hạn từ chối INVALID_HEALTH_AGE_LIMIT trước khi cấp authorization. Severe STOP hiện hữu vẫn có ưu tiên.

So sánh now với started/observed + duration thay vì Time.Sub để tránh bão hòa khoảng cách. Window reset tại đúng ngưỡng (>=); health stale khi vượt ngưỡng (>); authorization không được cấp khi expiry bằng now. Minimum expiry intent/lease/cost hiện hữu giữ nguyên.

MaxOutcomeAgeSeconds chỉ so sánh số nguyên, không nhân duration. Chain audit M10/M11 dùng Unix seconds và nanosecond trên timestamp RFC3339, không nhân duration; giữ nguyên. Schema canonical không bị siết theo giới hạn riêng runtime: schema/chain PASS không đồng nghĩa runtime cấp quyền với duration quá lớn. Fixture chain dùng runtime đổi sang số giây lớn nhất runtime biểu diễn được.

Regression kiểm âm/zero/giới hạn/sát vượt/MaxInt64; window trước/đúng/sau ngưỡng với nanosecond, không reset khi từ chối; health limit lớn và expiry boundary. Chạy `go test ./...` và `go vet ./...` từ lab/mission-runtime. Giá trị vượt int32 chỉ test trên máy 64-bit, tương ứng field int runtime.

Chỉ fixture/local sandbox. Không giải quyết trusted clock/provenance, rollback snapshot hoặc crash consistency. BR-03 tổng thể chưa hoàn thành.
