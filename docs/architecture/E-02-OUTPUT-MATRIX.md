# E-02 — Ma trận output runtime M09–M11

Trạng thái mới: #45 đã review/merge `2cf3bc1`. E-02b trong PR tiếp theo nối [governed_export.go](../../lab/mission-runtime/cmd/demo/governed_export.go) vào serializer mission-demo M10/M11. `TestGovernedDiagnosticExport` và `TestGovernedCanonicalExport` kiểm output bytes: diagnostic riêng, không có gate/authorization/authority; canonical artifact được kiểm và snapshot trước marshal. [E-03 review](E-03-ACCEPTANCE-REVIEW.md) đề nghị nghiệm thu có phạm vi sau review/merge. Các mục “chưa có export” bên dưới mô tả baseline #45, không phải trạng thái E-02b.

> Lịch sử: đoạn dưới từng mô tả [PR #45](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/45) ở trạng thái `IN_REVIEW` trên baseline #44. PR #45 đã merge `2cf3bc1`; trạng thái readiness hiện hành thuộc [REVIEW-REMEDIATION-PLAN.md](../plans/REVIEW-REMEDIATION-PLAN.md) và [READINESS-MATRIX.json](../plans/READINESS-MATRIX.json), không suy từ ma trận output lịch sử này.

## Bằng chứng trực tiếp mới

[output_matrix_test.go](../../lab/mission-runtime/cmd/demo/output_matrix_test.go) marshal output thật rồi gọi decoder đúng profile (decoder kiểm canonical raw schema trước semantic). Không lấy record dựng sẵn thay kết quả executor.

| Hàm/nhóm return | Artifact khác zero | Test và phạm vi |
|---|---|---|
| AuthorizeM09 thành công | ExecutionAuthorization APPROVED_LIVE | TestOutputMatrixExecutors: decoder M09 |
| ExecuteLocalSandbox thành công | SUCCEEDED/PERFORMED | TestOutputMatrixExecutors/success |
| ExecuteLocalSandbox MkdirAll lỗi | FAILED/NOT_PERFORMED | TestOutputMatrixExecutors/failure; path file trong tmp |
| ExecuteLocalSandbox marker tồn tại | RECONCILIATION_REQUIRED/UNKNOWN | TestOutputMatrixExecutors/unknown |
| AuthorizeCanary/AuthorizeProduction thành công | authorization và gate | TestOutputMatrixExecutors success/unknown; đúng profile GOVERNED_CANARY/PRODUCTION |
| ExecuteCanaryLocalSandbox/ExecuteProductionLocalSandbox thành công | SUCCEEDED/PERFORMED | TestOutputMatrixExecutors/success |
| Hai executor thấy marker thiếu durable success | RECONCILIATION_REQUIRED/UNKNOWN | TestOutputMatrixExecutors/unknown; M11 giữ STOP |
| Evaluate/Authorize gate hợp lệ | M10 ALLOW_CANARY/DENY/WAIT/REQUIRE_APPROVAL; M11 thêm STOP/DEGRADE/ALLOW_PRODUCTION | TestOutputMatrixGateDecisions; thiếu allowed executor, rate limit, total budget, sticky stop, telemetry thiếu |
| Authorize từ chối | gate + zero authorization | TestOutputMatrixGateDecisions assert zero; không serialize sentinel thành success artifact |

## Inventory nhánh còn lại và cách đọc phạm vi

| Nhánh | Xử lý / bằng chứng hiện hữu | Trạng thái |
|---|---|---|
| M09 marshal payload lỗi, OpenFile lỗi không phải exists | executionFailureRecord FAILED/NOT_PERFORMED giống nhánh MkdirAll | Shape được phủ qua builder chung; chưa fault-inject từng syscall |
| M09 Write/Sync/Close lỗi | closure uncertain → executionFailureRecord RECONCILIATION_REQUIRED/UNKNOWN | Shape giống marker exists; chưa fault-inject từng syscall |
| M10 Write/Sync/Close, normalize sau effect, persist ledger lỗi | canaryExecutionRecord RECONCILIATION_REQUIRED/UNKNOWN | Shape builder giống nhánh marker; chưa fault-inject từng syscall |
| M11 Write/Sync/Close, normalize/persist sau effect lỗi | productionExecutionRecord RECONCILIATION_REQUIRED/UNKNOWN hoặc zero khi persist STOP thất bại | Shape builder giống marker; chưa chứng minh crash/recovery |
| M09–M11 guard trước effect/authorization lỗi | zero record/authorization + status | Tests guard/persistence/time hiện hữu; không có canonical output để kiểm; caller không được xuất zero như record |
| M10/M11 CANCELLED và FAILED | Các constructor nhận status, boundary tests có fixture CANCELLED | Không thấy executor M10/M11 trả nonzero CANCELLED/FAILED ở baseline; không gán nhãn runtime coverage cho fixture |
| EnforceProductionGate early I/O/nil/missing state | gate chẩn đoán chỉ Decision/Reason, thiếu canonical IDs | Không là canonical gate đầy đủ; không được suy DecodeM11Artifact PASS. Cần quyết định export wrapper nếu đưa ra ngoài |
| Gate Evaluate với input thiếu ID/hash/cost | gate copy dữ liệu không hợp lệ từ input | Case hợp lệ trong test mới không chứng minh mọi diagnostic gate đúng schema; không tự điền fake IDs để lấy PASS |
| ProductionCycleRecord | ValidateProductionClosedCycle kiểm artifact caller cung cấp; m11_chain tests kiểm closed_cycle/resolved_stop | Không có cycle issuer standalone cần giả lập; output demo/fixture không chứng minh mọi cycle runtime |

E-02 **chưa đủ để đóng toàn bộ BR-03**: ma trận đã phân biệt output canonical và diagnostic nhưng chưa có contract export bảo đảm diagnostic không bị gắn nhãn canonical. Đây là phần cần review quyết định, không nới schema hoặc sửa runtime ngầm trong PR test này.

## Tiêu chí chốt tiếp

1. Review phạm vi shape theo builder chung có đủ cho BR-03; lỗi syscall/crash từng điểm thuộc H-03, không claim đã test trực tiếp.
2. Chốt diagnostic gate là local status envelope và kiểm wrapper không export thành canonical, hoặc thêm boundary từ chối xuất khi thiếu schema. Test nil/missing cost/hash/invalid time cho cả M10/M11. Không yêu cầu mọi invalid input tạo ra artifact hợp lệ giả.
3. E-03 chỉ đóng BR-03 sau khi quyết định trên có implementation/test nếu cần và được review. E-01 precision đã merge không thay E-02.

Chạy `go test ./cmd/demo -run TestOutputMatrix -count=1`, toàn suite/vet ba module, 8 validators và 10 Python regressions. Tất cả effect dùng tmp sandbox, không network/affiliate live; không sửa dữ liệu người học.
