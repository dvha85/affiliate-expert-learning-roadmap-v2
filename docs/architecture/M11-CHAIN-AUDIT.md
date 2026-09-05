# BR-03c.8b — Audit chain M11, không cấp quyền

Evidence local: tests/vet ba module, 8 Python validators, 10 Python regressions và diff check PASS. Tests CLI và chuyển trạng thái sandbox PASS. CI phải kiểm theo head PR riêng.

#36 đã review/merge `fa9b9db`, CI 4/4 PASS trên head `0f98edc`, tests/vet mission-runtime PASS. Bài này nối các artifact boundary M11/M08/M03/M05 trong một bundle lịch sử. Không thay `m11-check` kiểm từng file.

## Chạy bài offline

Từ `lab/mission-runtime`:

```text
go test ./cmd/demo -run TestM11Chain -count=1 -v
```

Tests tạo bundle synthetic trong thư mục tạm. Khi có file bài tập riêng:

```text
go run ./cmd/demo m11-chain-check bundle.json
```

`bundle.json` là tên file bạn cung cấp, không phải fixture có sẵn. Bundle là object vận chuyển local, không phải trusted state hoặc một canonical artifact mới. Mỗi giá trị artifact bên trong phải qua canonical schema và strict decode. Không nhận unknown/duplicate/trailing JSON; thiếu/null artifact bắt buộc bị chặn, optional field muốn bỏ phải bỏ hẳn, không ghi null.

## Hai profile tách biệt

Cả hai có `profile` cùng các key bắt buộc: `intent`, `policy`, `lease`, `approval`, `health`, `cost`, `activation`, `pre_ledger`, `gate`, `authorization`, `execution`, `post_ledger`.

- `profile: "closed_cycle"`: thêm `cycle`, `outcome`, `evaluation`; có thể thêm `proposal`, `review`. Review không được có nếu thiếu proposal. Execution phải SUCCEEDED/PERFORMED, outcome VALID. Không proposal thì cycle CLOSED; proposal chưa review thì REVIEW_PENDING; có review thì CLOSED. Không nhận resolution hoặc stop_ledger ở profile này.
- `profile: "resolved_stop"`: thêm `stop_ledger`, `resolution`; không nhận cycle/outcome/evaluation/proposal/review. Profile chỉ mô tả UNKNOWN **chưa commit bộ đếm execution vào ledger**, rồi human resolution. stop_ledger phải giữ nguyên pre_ledger ngoài STOPPED/RECONCILIATION_REQUIRED và updated_at. Resolution NOT_PERFORMED không tăng bộ đếm; PERFORMED tăng đúng một execution/cost/pending/key. post_ledger vẫn STOPPED/RECOVERY_REVIEW_REQUIRED. Không nhận biến thể effect đã được hạch toán trước stop; cần profile riêng nếu muốn hỗ trợ.

Đây là profile cho một issuance ALLOW_PRODUCTION, gate và authorization cùng instant. pre_ledger là snapshot trước gate, không STOPPED, không reconciliation và không có resolution history của lease cũ. Gate DENY/STOP/WAIT/DEGRADE có thể đúng artifact nhưng không phù hợp với chain issuance này. post_ledger của closed_cycle là snapshot ngay sau ghi outcome; cycle/evaluation/review có thể được hoàn tất sau đó. Không trộn ledger của nhiều cycle đồng thời.

## Kiểm liên kết và chuyển trạng thái

- Exact lease ID/version/hash giữa approval/health/activation/ledger/gate/auth/execution/cycle hoặc resolution. Promotion review, source canary ID/version/hash, reviewer và thời điểm review phải khớp giữa lease và approval. Không resolve E5 hoặc canary payload từ các ref đó.
- Intent/policy/hash/version, health/cost ref/hash, risk, amount/currency, executor, correlation/idempotency phải khớp đúng nơi contract liên kết. Hash intent/lease/health/cost tính lại, không seal hoặc sửa input. Scope risk/action/host/executor và risk đã được promotion review phải bao gồm action.
- Kiểm thời gian tại gate, độ mới health, expiry auth không vượt intent/lease/cost, attempt trước expiry. Activation trong lease window và không sau gate. Rate window có thể bắt đầu trước activation, đúng cách runtime hiện hữu lưu initial ledger; không suy từ đó được phép tạo lại ledger.
- Chặn pre-ledger STOPPED/reconciliation/failure threshold, health degraded/incomplete/compliance/failure/outcome-stale. Không biến file health thành trusted source hoặc thực hiện safety action.
- Gate counters khớp pre_ledger sau rate-window reset. Elapsed dùng Unix seconds/nanoseconds; cost dùng phép trừ kiểm bound trước khi cộng, tránh overflow. Chặn total/rate/pending/cost budget, idempotency đã thành công, execution đã pending/linked.
- closed_cycle: post-ledger tăng đúng một execution/cost, append đúng successful key và outcome link; giữ pending/history/resolution IDs, không reset ngân sách. Canonical MACHINE_EXECUTION EffectRef nối outcome/evaluation với execution; cycle nối đúng ID, evidence/outcome/evaluation tập đơn theo profile, timestamp outcome → evaluation → review/close. Proposal auto_apply bị chặn; approval của review không áp dụng thay đổi.
- resolved_stop: kiểm snapshot STOP trung gian và post-resolution đúng chuyển trạng thái nêu trên, không nhận NORMAL hoặc mất history sau resolution. Không khởi tạo ledger, không resume lease cũ.

## Output và giới hạn

Exit 0 chỉ xuất `CONSISTENT_UNVERIFIED`, profile, lease/execution ID và ba flag false: `provenance_authenticated`, `execution_permitted`, `resume_permitted`. Exit 1 khi args/read/schema/profile/link/time/scope/budget/transition sai; lỗi input không có summary. Writer errors được trả lại. CLI không gọi policy/gate/authorize/executor/initialize/reconciliation hoặc persistence, không network, không thay file. Đọc lại bundle chỉ kiểm lại, không chứng minh chống replay khi execute.

PASS không xác thực danh tính reviewer/resolver, source canary/E5, Decision/Evidence payload, trusted health/cost, độ đầy đủ/độ mới ledger, revocation/kill switch bên ngoài hoặc side effect. Một bộ file giả nhất quán có thể PASS; mọi flag quyền vẫn false. Không chứng minh activation chỉ xảy ra một lần trên hệ thống thật, không chứng minh budget toàn lịch sử. Trusted persistence/executor integration và các profile ngoài hai profile trên còn mở trong BR-03.

Tests kiểm schema/missing/null/stage/link, health/scope/time, budget, window reset, cycle/review/auto-apply, STOP không resume, CLI read-only/replay/writer. Test chuyển trạng thái thực gọi runtime trong thư mục sandbox tạm: execution local + outcome và UNKNOWN + trusted resolution synthetic, không external adapter. Không có chương trình/kênh affiliate thật và không có E5/E6 live proof. Không cập nhật PROGRESS của học viên hoặc đánh dấu BR-03 hoàn tất.
