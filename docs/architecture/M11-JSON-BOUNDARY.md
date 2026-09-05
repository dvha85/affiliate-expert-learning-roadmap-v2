# BR-03c.8 — M11: kiểm từng artifact production

Evidence local: tests/vet ba module, 8 Python validators, 10 Python regressions và diff check PASS. Smoke CLI activation trả ARTIFACT_VALID_UNVERIFIED với ba flag false. CI cần kiểm trên head PR riêng.

#35 đã review/merge `07c1b8a`, CI 4/4 PASS trên head `7568f59`, tests/vet mission-runtime PASS. Bài M11 này chỉ kiểm từng artifact; chưa kiểm liên kết toàn chain.

## Bài offline

Từ repo root:

```text
cd lab/mission-runtime
go run ./cmd/demo m11-check activation testdata/m11-activation.json
go test ./cmd/demo -run 'TestM11(Raw|Audit)' -count=1 -v
```

CLI nhận `m11-check KIND FILE.json`. Có 11 KIND: lease, approval, health, cost, ledger, gate, authorization, execution, activation, resolution, cycle. Cost reuse boundary trusted-cost-bound của M10; các profile còn lại dùng canonical schema production/execution tương ứng và exact typed decode. Không chấp nhận field authorization/execution của M09/M10 trong profile production.

Exit 0 chỉ xuất `ARTIFACT_VALID_UNVERIFIED`, artifact_type, `provenance_authenticated:false`, `chain_validated:false`, `execution_permitted:false`. Exit 1 khi args/read/schema/hash/time/semantic sai; lỗi input không có summary. Writer errors không bị bỏ qua. CLI không gọi gate, authorize, initialize, executor, reconciliation hoặc persistence; không network, không thay input.

Fixture activation có hash placeholder đúng định dạng, không có lease tương ứng, không phải chứng cứ activation. Không đưa file vào runtime hoặc dùng để khởi tạo ledger. Một approval ghi human hoặc source_e5_refs không chứng minh người duyệt/E5 thật; resolution PERFORMED không chứng minh side effect. Chưa có chương trình/kênh affiliate thật.

## Ràng buộc đã kiểm

- Required/null/type/unknown/duplicate/trailing/enum/const theo raw schema trước typed decode. Optional field được bỏ thì hợp lệ nhưng null vẫn bị chặn nếu schema không cho phép.
- Lease/health tính lại hash bằng thuật toán runtime, không seal hoặc sửa input. Lease có window dương, review không sau valid_from, per-window limit không vượt total, không wildcard action/executor.
- Authorization phải GOVERNED_PRODUCTION, window dương. Gate luôn execution_authorized=false. Cycle closed_at không trước opened_at; không resolve các ID trong cycle.
- Ledger: pending khớp số ID và không vượt executions total, in-window/success keys không vượt total, links không trùng hoặc đồng thời pending, timestamps không sau updated_at. Không kiểm đầy đủ kế toán lịch sử hoặc ngân sách còn đủ.
- MarshalJSON ledger runtime đổi bốn collection nil thành array rỗng, không mutate state. Raw null vẫn bị từ chối; không nối decoder vào reader persistence cũ.

Tests serialize lease/approval/health/cost/ledger từ runtime, authorization/gate từ AuthorizeProduction, execution CANCELLED/NOT_PERFORMED từ helper runtime. Activation/resolution/cycle là record synthetic; cycle có ID outcome/evaluation chưa resolve, không phải closed-loop evidence. Không gọi executor trong test mới. Tests required/null/type/duplicate/stage/authority/hash/time/offset/int64 precision/overflow/CLI read-only/replay/errors; regression executor/STOP/reconciliation/activation/closed-loop hiện hữu giữ nguyên.

## Không được suy từ PASS

Health DEGRADED có thể đúng schema/hash; ledger STOPPED có thể đúng artifact. PASS không có nghĩa healthy, lease còn hiệu lực hiện tại, hay hệ thống được resume. Các file đều PASS riêng vẫn có thể liên kết sai. Không xác thực promotion source, E5, reviewer, trusted health/cost, activation provenance, resolution, sticky STOP hoặc external effect. Đọc lại file không chứng minh chống replay/budget reset.

Bước tiếp theo là audit liên kết toàn chain M11 với snapshot và profile rõ ràng. Trusted persistence/executor integration vẫn là phần riêng còn mở trong BR-03. Không tự cập nhật PROGRESS học viên hoặc đánh dấu BR-03 hoàn tất.
