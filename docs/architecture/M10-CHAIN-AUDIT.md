# BR-03c.7b — Audit chain M10 chỉ đọc

Trạng thái IN_REVIEW tại [PR #35](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/35), commit triển khai `9dbd9a0`. [CI theo head PR](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/35/checks). Chưa merge.

Kiểm local: tests/vet ba module, 8 Python validators, 10 Python regressions và diff check PASS. Tests CLI dùng chín file synthetic trong thư mục tạm, xác nhận summary không cấp quyền và input không thay đổi. CI cần kiểm theo head PR riêng.

#34 đã review/merge `1d6594c`, CI 4/4 PASS trên head `732cb71`, tests/vet mission-runtime PASS. Bài này nối các boundary của #34 với intent/policy M08; không thay `m10-check` kiểm từng file.

## Chạy và hợp đồng đầu vào

Từ `lab/mission-runtime`, chạy bài synthetic có tạo đủ chín file trong thư mục test tạm:

```text
go test ./cmd/demo -run TestM10Chain -count=1 -v
```

Khi đã có file bài tập của riêng bạn, thứ tự CLI là:

```text
go run ./cmd/demo m10-chain-check intent.json policy.json grant.json approval.json cost.json pre-gate-ledger.json gate.json authorization.json execution.json
```

Các tên trên minh họa file người học cung cấp, không phải fixture đã có ở repo root. Fixture `testdata/m10-approval.json` của bài cũ cố ý không có grant tương ứng, không dùng cho chain PASS.

Profile chỉ audit một issuance `ALLOW_CANARY/CANARY_ELIGIBLE` với ledger snapshot **trước gate**, authorization GOVERNED_CANARY và execution record tương ứng. Gate evaluation và authorization phải cùng instant (profile AuthorizeCanary hiện hữu); không nhận delayed issuance. Không dùng ledger sau execution. DENY/WAIT/REQUIRE_APPROVAL là record có thể hợp lệ riêng nhưng không hợp lệ trong chain có authorization này.

Exit 0 xuất summary `CONSISTENT_UNVERIFIED`, intent/grant/execution ID và `provenance_authenticated:false`, `execution_permitted:false`. Exit 1 khi args/read/schema/hash/link/scope/time/ledger/budget sai; input lỗi không có summary. Không echo authorization, không gọi policy/gate/authorize/executor, không ghi/normalize ledger hoặc gọi network. Writer errors được trả lại.

## Kiểm liên kết và ràng buộc

- Chín raw JSON qua boundary tương ứng; không seal, sửa hash, thêm default hay trust map từ file.
- Hash intent/grant/cost tính lại. Policy, approval, ledger, gate, authorization và execution phải khớp exact ID/version/hash; policy version, risk, currency, cost amount, executor, correlation và idempotency phải khớp nơi contract liên kết. Grant correlation riêng không bị ép bằng intent correlation.
- Approval ref, approver, approved_at phải khớp grant; timestamp so theo instant. `human` vẫn là claim, không phải xác thực danh tính.
- RISK0 cần ALLOW; RISK1 cần HUMAN_REVIEW/RISK1_REQUIRES_REVIEW. Grant phải bao gồm risk/action/host/executor; không wildcard action/executor. Đây là scope khai báo, không xác thực executor profile bên ngoài.
- Intent/policy/grant/cost phải có thời gian phù hợp tại gate; authorization expiry không vượt intent/grant/cost. Execution không trước authorization; claim PERFORMED phải trước expiry, đúng bằng expiry bị chặn. FAILED/CANCELLED/UNKNOWN có thể ghi nhận attempt sau expiry, không cấp quyền execute.
- Snapshot ledger không đến từ tương lai; gate counters phải khớp snapshot sau phép reset số lần trong window tại gate. Tính elapsed bằng Unix seconds/nanoseconds, không nhân WindowSeconds thành time.Duration gây tràn. Không mutate snapshot.
- Chặn reconciliation, idempotency đã thành công, execution đã pending/linked. Kiểm total/rate/pending/cost budget; cost dùng phép trừ sau kiểm bound, không cộng int64 có thể overflow.

## Bài sửa lỗi và giới hạn

Sửa riêng approval.grant_id hoặc execution.authorization_id: schema từng file có thể PASS nhưng chain trả BROKEN_LINK. Sửa counter của gate: LEDGER_SNAPSHOT_MISMATCH. Thử cost vượt ngân sách, scope không được grant, hoặc PERFORMED đúng expiry. Khôi phục input gốc, không sửa expected hoặc seal lại để che tampering. Tests còn có RISK1, rate-window reset, timezone offset, integer overflow, đọc lại không cấp quyền, file không đổi và writer error.

Đây là consistency của một chain lịch sử từ file **không tin cậy**. Không xác thực người duyệt, source_ref, tính đầy đủ/độ mới ledger, policy hiện hành, revocation, kill switch thật, external executor profile hoặc side effect. Không resolve nội dung Decision/Evidence/Outcome; không chứng minh đúng kế toán toàn lịch sử. Một chain giả nhất quán vẫn có thể PASS nhưng hai flag quyền luôn false. Lặp CLI không chứng minh chống replay khi execute.

Không tích hợp reader persistence hoặc thay runtime gate/executor. Regression executor/restart/reconciliation cũ giữ nguyên. Chưa có chương trình affiliate/kênh thật nên không có live proof; BR-03 tổng thể còn mở. Hoàn tất review phần này trước khi chuyển M11; không tự cập nhật PROGRESS của học viên.
