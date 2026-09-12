# Kế hoạch sửa sau review trước merge tại 737e85a

- Mã kế hoạch: PMR-2026-09-09; phiên bản: 1.
- Ngày lập: 09/09/2026.
- Trạng thái: **IN_REVIEW — đã có implementation commits và local evidence; chưa có CI đúng head hoặc review độc lập, chưa đủ điều kiện đề xuất merge**.
- Head được review: `737e85ae008088be6dcb9eb19a269e01c4a4005f`, nhánh `codex/rp-01-path-safety`.
- Base đối chiếu local: `origin/main` tại `12a088aab619106d1206d87dfccc25843d3bcba7`.
- Phạm vi diff đã đối chiếu: 62 commit, 77 file; không chỉ phần M11 của các commit cuối.
- Chưa xác minh số/trạng thái PR và CI mới nhất trên GitHub trong lượt review: truy cập remote bị hệ thống duyệt từ chối do workspace hết credits. Không coi bản soạn PR là PR đã được tạo.
- Kế hoạch cha: [REVIEW-REMEDIATION-PLAN.md](REVIEW-REMEDIATION-PLAN.md). Readiness tổng hợp: [READINESS-MATRIX.json](READINESS-MATRIX.json).

## 1. Mục tiêu, phạm vi và quy tắc nghiệm thu

Khắc phục các lỗi runtime đã tái hiện trước khi đề xuất merge nhánh hiện tại. Tài liệu này là phần bổ sung cho RP-01…RP-09, không thay thế lịch sử triển khai hoặc tự đóng các RP/BR còn mở.

Đợt thực hiện sau kế hoạch chỉ thay đổi Go/Python/schema và fixture cô lập. Không chạy live executor, không dùng tài khoản ACCESSTRADE hoặc dữ liệu khách hàng. Operated M07 local dùng Cockpit loopback với fixture synthetic, không tạo giao dịch affiliate thật. Commit/push chỉ theo yêu cầu chủ repo.

Nguyên tắc bắt buộc:

1. Viết regression gọi đúng core/learner implementation, chứng minh lỗi trên head review rồi chạy lại sau fix. Không dùng parser riêng hoặc kiểm từ khóa làm bằng chứng thay thế.
2. ACK chỉ sau commit bền vững; reject không được thay đổi canonical business artifacts. Lock hoặc staging phục vụ recovery phải có trạng thái và cách xử lý rõ ràng.
3. Không tự sửa, reset, bỏ qua hoặc chọn tùy ý một nhánh ledger khi dữ liệu cũ đã mâu thuẫn.
4. Giữ `NOT_READY_FOR_PRODUCTION`; authorization artifact không đồng nghĩa có quyền hoặc bằng chứng live execution. Hoàn tất fixture offline không chứng minh business outcome, n8n operated run hay beginner pilot.
5. Mỗi gói có trạng thái `TODO → IN_PROGRESS → IN_REVIEW → VERIFIED_OFFLINE`; `BLOCKED` phải ghi nguyên nhân. Chỉ đóng sau khi có implementation SHA, test evidence và reviewer xác nhận. Owner/reviewer hiện chưa phân công.
6. File mới/API/test được đề xuất dưới đây chưa phải implementation đã có. Link mã nguồn là vị trí khảo sát tại head review; line number có thể thay đổi sau fix.

## 2. Sổ phát hiện và bằng chứng baseline

Các ca PMR-01…06 dùng Bot được build từ repo, tái sử dụng setup hợp lệ của `scripts/smoke_br18b_backup_restore.py` trong thư mục tạm và gọi CLI bằng process mới. PMR-07 gọi `core/m07.ValidateAgentOutput` và learner HTTP adapter. Các ca baseline đã được chuyển vào test/smoke trong cập nhật triển khai ngày 09/09; các mục hardening còn lại được ghi rõ ở phần cập nhật, không được suy là đã đóng toàn bộ readiness.

| ID | Ưu tiên | Hành vi đã tái hiện | RP liên quan | Trạng thái |
|---|---|---|---|---|
| PMR-01 | P1 | Lease đã dùng hết 1 execution/4 cost vẫn reserve intent thứ hai bằng ledger ban đầu; gọi ledger-init lại cũng tạo counters bằng 0 | RP-03, RP-07a | regression PASS local; IN_REVIEW |
| PMR-02 | P1 | Health max-age 60 giây, authorize sau 120 giây vẫn APPENDED với execution_authorized=true | RP-03, RP-07a | regression PASS local; IN_REVIEW |
| PMR-03 | P1 | Đổi outcome ID/thời gian, dùng lại execution ledger cũ: outcome thứ hai cùng execution vẫn APPENDED; backup vẫn BACKED_UP | RP-07a, RP-06 | regression PASS local; IN_REVIEW |
| PMR-04 | P1 | Cycle trỏ decision/observations không tồn tại, checksum hợp lệ: restore vẫn RESTORED | RP-06, RP-07a/b | regression PASS local; IN_REVIEW |
| PMR-05 | P1 | Restore GRAPH_FAILED đã ghi file vào đích; process mới chạy mission status tại đó vẫn VALID | RP-01, RP-06 | regression PASS local; IN_REVIEW |
| PMR-06 | P2 | Gate cùng lease/intent được đánh giá lại sau 1 giây bị lỗi ID reused with different content | RP-03, RP-07a | regression PASS local; IN_REVIEW |
| PMR-07 | P2, lỗi cũ | Evidence 9007199254740993 và claim 9007199254740992 được coi là bằng nhau qua float64 | RP-02, RP-05 | regression PASS local; IN_REVIEW |

PMR-07 đã tồn tại ở base; không mô tả đây là regression mới do nhánh này gây ra. PMR-05 là thiếu sót ở đường restore hiện có, nay được xử lý cùng các graph gates mới.

### Cập nhật triển khai — 09/09/2026

Các thay đổi dưới đây đã được commit/push trên nhánh review nhưng chưa merge; chúng không thay đổi `NOT_READY_FOR_PRODUCTION`.

- **PMR-01:** M11 chỉ cho transition mới dùng ledger head có `updated_at` lớn nhất, duy nhất theo lease. `m11-ledger-init` chỉ idempotent cho genesis chính xác và từ chối reset; generic `m11-register` không thể chèn lifecycle ledger/gate/authorization/execution/evaluation/cycle. Core registry từ chối lịch sử ledger giảm total executions/cost. Khi stopped ledger đã commit nhưng state/STOP chưa ghi xong, preflight dưới runtime lock tái lập durable STOP trước mọi mutation khác. BR-18b tạo intent/approval/cost-bound thứ hai trong runtime đã dùng hết lease: genesis bị từ chối và head trả `BUDGET_EXCEEDED`.
- **PMR-02:** authorize kiểm health/gate không ở tương lai, health chưa đến biên max-age, và bind exact `ledger_artifact_id`/`ledger_content_hash` đã được gate dùng. FAILED/STOP làm gate cũ stale kể cả khi bốn counter budget không đổi. Reserve chặn thời gian trước authorization và không cho timestamp đứng/lùi sau ledger. BR-18b kiểm authorize ở +120 giây và sau FAILED đều bị reject. Vẫn phải kiểm lại toàn bộ policy về clock vận hành nếu có live adapter; fixture clock không là production authority.
- **PMR-03:** M11 outcome store giờ unique theo execution ID trên toàn store trước khi chọn ledger; transition outcome phải dùng ledger head và advance thời gian. Journal `m11-outcome-journal/v1` được fsync trước ledger/outcome, mang exact predecessor artifact ID/content hash, và phải tái tạo đúng transition canonical từ predecessor trước khi replay idempotent dưới runtime lock; chỉ xóa sau hai append. Regression chứng minh journal bị sửa tay bị reject mà không đổi ledger head, inject lỗi đúng trước outcome append, rồi recovery hoàn tất exactly-once. Backup validator cũng từ chối snapshot có hai outcome cho một execution. BR-18b kiểm outcome ID mới với execution cũ bị reject. Journal chỉ bao phủ cặp M11 ledger/outcome; transaction đa-file tổng quát vẫn là hardening mở.
- **PMR-04:** restore resolve/replay canonical history cho every cycle, yêu cầu tập observation chính xác, đồng thời kiểm intent/hash/gate/authorization gắn với execution/evaluation. BR-18b thay decision/observation bằng ID không tồn tại nhưng recompute checksum/hash và nhận `GRAPH_FAILED`. Liên kết M07 proposal được validator M07 riêng kiểm; một cross-store evidence graph tổng quát vẫn là việc mở.
- **PMR-05:** restore copy vào staging sibling, chạy replay/load/graph gates tại đó rồi mới rename publish. Đích phải chưa tồn tại để tránh ghi đè directory rỗng của người dùng; per-target gate serialize các publisher managed trước final rename. Mọi lỗi dọn staging do operation tạo. BR-18b xác nhận `GRAPH_FAILED` không tạo target. Kill/power-loss và writer không tôn trọng managed target gate vẫn cần drill host thật.
- **PMR-06:** Gate ID là digest của lease/intent/health/cost/immutable ledger entry/evaluation time; retry cùng input idempotent, evaluation khác tạo artifact khác. Gate persist exact `ledger_artifact_id` và `ledger_content_hash`; core graph resolve hai refs này. BR-18b kiểm re-evaluate sau một giây không còn collision.
- **PMR-07:** Core M07 dùng decoder exact-number, tool result/evidence/claim/render không đi qua float64. Adapter trả JSON text nguyên vẹn cho workflow; blueprint chuyển context/evidence/model output bằng text thay vì `JSON.stringify`/`JSON.parse` số. Unit/adapter test kiểm `9007199254740993` không match `9007199254740992`. N8n engine operated run với synthetic fixture đã PASS; CI/containerized repeat và provider diversity vẫn PARTIAL.
- **PMR-07 — operated-run update (09/09):** Đã chạy n8n `2.38.1` local, import workflow M07 inactive, tạo canonical history từ synthetic fixture và gọi canonical adapter thật tại loopback. Context `VALID` có ba evidence; request `POST` bị adapter reject trước fetch; execution n8n đi đúng vào `Fetch and Register Tool Adapter` và dừng trước model/persistence khi input không hợp lệ. Lần chạy valid-input tiếp theo phát hiện blueprint có newline thật bên trong JavaScript string của `Read-only Evidence Agent`, khiến engine báo `invalid syntax` trước model call. Blueprint đã đổi sang `\\n` và validator M07 từ chối regression này. Sau fix, run đi tới Agent; mock OpenAI-compatible cục bộ không đáp ứng đầy đủ contract tool-calling/streaming của Agent v3 nên execution bị treo, được recovery sau restart (`database integrity_check=ok`, execution `crashed`). Tại thời điểm cập nhật này chưa có full successful M07 engine run; kết quả Cockpit bên dưới thay thế giới hạn đó.
- **PMR-07 — Cockpit operated-run update (09/09):** Cockpit local gateway tại loopback đã nhận request từ n8n; `gpt-5-mini`, `gpt-5.4-mini` và `gpt-5.3-codex` bị upstream từ chối theo quyền tài khoản, còn `gpt-5.6-luna` trả model output. Tool adapter, registered evidence và canonical context đều pass trước Agent. Grounding adapter từ chối output với `ABSTAIN`/HTTP 400 vì model bỏ `field_or_claim`/`value`, dùng authority sai và trả `proposed_action` là string. Không có proposal nào được persist. Prompt blueprint nay ghi rõ JSON contract này và static validator kiểm các marker; cần import/cập nhật workflow local rồi chạy lại để có bằng chứng successful grounding/persistence.
- **PMR-07 — Cockpit rerun:** Sau khi siết schema, `gpt-5.6-luna` tạo `ABSTAIN` hợp lệ (`A2-RO`, read-only, không claim), grounding adapter trả `VALID` và gate persistence từ chối đúng vì không có `HUMAN_REVIEW` proposal. Prompt trước đó chưa nêu task cụ thể; nay nêu rõ draft chỉ đọc phải tạo `HUMAN_REVIEW` từ scalar canonical evidence nếu có, và chỉ ABSTAIN khi không thể restate evidence. Chưa có ACK persistence; cần rerun sau cập nhật này.
- **PMR-07 — Cockpit rerun 2:** Model tạo `HUMAN_REVIEW` với field/value/IDs/authority/proposed action đúng, nhưng grounding adapter vẫn từ chối. Chạy cùng payload qua CLI ghi rõ mismatch: blueprint yêu cầu answer nối bằng `; `, còn `core/m07.RenderGroundedAnswer` chỉ chấp nhận newline. Blueprint đã đồng bộ delimiter newline và validator kiểm marker này; cần rerun sau cập nhật local để có evidence persistence.
- **PMR-07 — Cockpit rerun 3 (operated PASS):** Với `gpt-5.6-luna` qua Cockpit loopback, synthetic fixture chạy thành công đến `Report Persisted M07 Proposal`: tool trace `ACK`, canonical context `VALID`, grounding `VALID`, proposal register `ACK` và persistence gate `ACK`. Proposal `sha256:afdfbb9116ea609f8032443021956d03a27c8ae8da40199e3a1ea22191c9b7de` được ghi tại canonical proposal store cho record `watch-eda8abfb4d4f9785e469ca43cd061c8a903e752fd5f0d8019fee3291dcca229a`; output giữ `execution_permitted=false`. Đây là bằng chứng engine-operated với fixture synthetic, không phải business outcome hay quyền thực thi. Cần CI/containerized operated test lặp lại để nâng confidence ngoài local host.
- **PMR-07 — Cockpit negative operated PASS:** Workflow local được chạy với tool request `POST` trong registry chỉ allowlist `GET`. Engine dừng tại `Fetch and Register Tool Adapter` với `TOOL_TRANSPORT_REJECTED`; Agent và proposal persistence không xuất hiện trong execution. Proposal inventory giữ nguyên một file đã ACK. Input `GET` fixture được phục hồi và n8n restart sau test. `validate_n8n_m07_operated_execution.py` nay kiểm được cả success ACK và reject-before-persistence capture.
- **PMR-07 — Cockpit rerun 4 (after RP-08 merges):** n8n `2.38.1` local chạy M07 qua Manual Trigger thành công ở execution `12`; mọi node từ tool trace/context/Agent đến grounding/proposal/report đều `success`. Grounding là `VALID/HUMAN_REVIEW`, report ACK cùng proposal `sha256:afdfbb9116ea609f8032443021956d03a27c8ae8da40199e3a1ea22191c9b7de` và `execution_permitted=false`. Sau graceful restart n8n, `/healthz`, adapter health, canonical replay và proposal ID đều PASS. Đây vẫn chỉ là synthetic/read-only local evidence; CI model-success, received redirect transport và provider diversity còn mở.
- **PMR-07 — Cockpit re-verification (12/09):** Một workflow local cũ thiếu cạnh `Require Grounded Proposal -> Persist Agent Proposal Adapter`; Agent/grounding vẫn `success` nhưng không tạo proposal. Blueprint trong repo và static validator đã có cạnh đúng. Đã đồng bộ graph của workflow local, chạy synthetic `GET` bằng Cockpit `gpt-5.6-luna`, và capture `n8n execute --rawOutput` PASS qua `validate_n8n_m07_operated_execution.py`: tool/context/Agent/grounding/persistence/report đều thành công, proposal canonical được resolve và `execution_permitted=false`. Negative `POST` dừng tại `Fetch and Register Tool Adapter`, không gọi Agent/persistence và không đổi inventory proposal. Sau restart n8n và adapter, history replay cùng validator success capture vẫn PASS. Workflow temporary đã bị xóa; workspace quay lại hai workflow. Đây chỉ là xác minh operated synthetic/read-only, không thay đổi trạng thái `PARTIAL`, business-outcome evidence hay execution authority.
- **M06 operated re-verification (12/09):** n8n `2.38.1` local chạy bản workflow temporary từ blueprint M06 hiện hành qua adapter loopback. Capture `--rawOutput` được `validate_n8n_m06_operated_execution.py` kiểm: adapter ACK, persistence ACK/report và record canonical cùng một ID, `execution_permitted=false`; history replay vẫn `MATCH` sau restart n8n/adapter. Fixture source không allowlist dừng tại adapter trước ACK/report và inventory history giữ nguyên. Workflow temporary đã bị xóa; local M06 giữ Schedule Trigger và được đồng bộ field read-only ở report cuối. Đây là fixture synthetic/read-only; không phải source selected, deployment proof hay business evidence, nên BR-13/14 vẫn `PARTIAL`.

Kết quả kiểm local sau triển khai, không phải nghiệm thu/merge approval:

- PASS trước lần thêm ledger ref: `go test`/`go vet` toàn learner Bot và learner race suite, BR-16a. Đây không phải kết quả CI của worktree cuối cùng.
- PASS tại worktree hiện tại: `go test`/`go vet` cho `contracts`, `core`, `lab/mission-runtime`; full learner Bot và full learner `-race`; BR-16a; learner tests M11/journal/backup được chọn, kể cả `-race`; smoke BR-18b; 15 Python unit tests và các validator M06/M07/adversarial/output; readiness audit; `git diff --check`.
- Lần thử sandbox trước đó không thể bind `::1`/`127.0.0.1`, nhưng đã chạy lại thành công full learner, full learner race và BR-16a ở worktree hiện tại khi môi trường cho phép loopback. N8n M07 engine operated run nay có evidence local; mutation suite và CI remote tại head mới vẫn là merge gate mở.
- Đã rà soát wiring CI: `curriculum-ci.yml` chạy full learner Bot, learner race, BR-16a, BR-18b, M06/M07 validators và readiness audit trên mọi pull request; `mission-agent-path-ci.yml` chạy core/harness test+vet. Đây chỉ là bằng chứng workflow đã khai báo, không thay thế một CI run PASS ở exact head sắp merge.
- PR #95 đã chạy CI cho `a5d3ad0`: ba job PASS, nhưng `Curriculum CI / deterministic-runtime` fail tại BR-12d vì smoke so sánh SHA toàn file baseline đã stale sau thay đổi hợp lệ. Follow-up thay SHA bằng kiểm chứng before/after/rollback theo targeted behavior; BR-12d PASS local. Chưa coi CI là PASS cho đến khi commit follow-up có run hoàn tất.
- Các kết quả local chỉ cho thấy bảy regression hiện có bị chặn trên các đường đã kiểm; không là chứng cứ production readiness hoặc business outcome.

## 3. Thứ tự triển khai

| Đợt | Nội dung | Phụ thuộc | Điều kiện kết thúc |
|---|---|---|---|
| A | PMR-01: ledger hiện hành và budget toàn lease | Chốt contract ledger/recovery | Hai intent không vượt cap; restart/concurrency không reset usage |
| B | PMR-06 rồi PMR-02: identity gate và hiệu lực authorization | A | Gate có thể đánh giá lại; stale health và authority bị từ chối |
| C | PMR-03: outcome duy nhất và commit an toàn | A | Không tạo hai outcome cho cùng execution; retry/crash có kết quả rõ |
| D | PMR-04: kiểm graph đầy đủ | A, B, C | Orphan và liên kết chéo bị reject, graph hợp lệ replay được |
| E | PMR-05: restore staging và publish an toàn | Có thể viết staging song song A–D; nghiệm thu sau D | Không công bố runtime lỗi; fail/retry/concurrency được kiểm |
| F | PMR-07: exact-number toàn đường M07 | Độc lập, có thể làm song song A–E | Giá trị khác không được grounded; giá trị đúng không bị làm tròn |
| G | Tích hợp regression, CI, cập nhật plan/readiness và review lại | A–F | Có bằng chứng local + CI đúng head và review lại các P1/P2 |

Mỗi đợt là một changeset có thể review độc lập, không mặc định tạo một PR GitHub mới. Nếu tiếp tục trên nhánh hiện tại, mô tả PR phải bao quát toàn diff; không dùng tiêu đề chỉ nói M11 backup cho thay đổi M06–M11. PMR-06 được làm trước PMR-02 để việc refresh health/gate không bị collision ID cản trở.

## 4. Chi tiết từng gói

### PMR-01 — Ledger hiện hành, budget toàn lease và chống fork

Vị trí: [m11_registry.go](../../lab/affiliate-bot/cmd/bot/m11_registry.go), [runtime_gate.go](../../lab/affiliate-bot/cmd/bot/runtime_gate.go), [core registry](../../core/m11/artifact_registry.go), [mission_command.go](../../lab/affiliate-bot/cmd/bot/mission_command.go).

Thay đổi cần làm:

- Chốt identity theo lease ID/version/hash và biểu diễn ledger head, revision, predecessor. Head phải được xác định từ chuỗi commit hợp lệ, không chỉ lấy timestamp lớn nhất hoặc artifact cuối file.
- Thêm resolver/reducer dùng chung: gate, reserve, record execution, outcome, reconcile, STOP và restore đều dùng trạng thái hiện hành. Caller không được dùng snapshot cũ để mở lại budget.
- Đặt thao tác resolve/check/append/commit dưới cùng khóa runtime. So sánh predecessor/revision để ngăn hai authorization khác nhau cùng tiêu một số dư.
- `m11-ledger-init` chỉ tạo genesis khi lease chưa có ledger; retry đúng genesis trả duplicate, timestamp khác không tạo genesis mới. Không reset totals khi qua cửa sổ thời gian, rebind intent hoặc restart.
- Giữ total executions/cost không giảm; chỉ counters theo window được chuyển cửa sổ theo contract. Outcome/reconciliation không hoàn budget đã charge nếu chưa có quy tắc hoàn được duyệt riêng.
- Hạn chế `m11-register` với ledger/artifact chuyển trạng thái: không được nhập một ledger zero hoặc fork để né writer có guard. Luồng import phục hồi nếu cần phải riêng biệt và validate toàn graph trước publish.
- Với registry legacy: hỗ trợ replay chỉ khi suy ra chuỗi duy nhất, nhất quán; nếu không đủ metadata thì fail closed và hướng dẫn recovery được review, không tự bổ sung hoặc ghi đè artifact bất biến.

Regression bắt buộc:

- Lease cap=1/cost=4; intent A đã reserve/failed/outcome, intent B có approval hợp lệ: dùng genesis hoặc head đều không được reserve vượt cap.
- Re-init sau usage; rebind A→B→A; nhập ledger zero; chọn predecessor cũ; thiếu head; hai head cạnh tranh đều không mở lại ngân sách.
- Hai process với authorization khác nhau, barrier đồng bộ, cap=1: đúng một commit; process thua nhận conflict/BUSY rõ ràng và retry không vượt cap.
- Crash trước append, sau append, trước/sau commit head: không có ACK giả; recovery không double-charge hoặc bỏ charge. Chạy lại bằng process mới và sau backup/restore.

Nghiệm thu: có một authoritative ledger chain trên mỗi lease; usage khớp các reservation đã commit; không phục hồi quyền bằng fixture timestamp hay chỉnh input ledger.

### PMR-06 — Identity gate và lịch sử đánh giá

Vị trí: [m11_registry.go](../../lab/affiliate-bot/cmd/bot/m11_registry.go), [artifact.go](../../core/m11/artifact.go), [artifact_registry.go](../../core/m11/artifact_registry.go).

Thay đổi cần làm:

- Thay gate ID chỉ ghép lease/intent bằng identity có version dựa trên canonical evaluation inputs: intent/hash, policy, lease/hash, health/hash, cost/hash, ledger head/revision và thời điểm đánh giá.
- Retry cùng evaluation trả đúng artifact cũ; thay health, cost, ledger hoặc thời điểm tạo evaluation mới, không overwrite gate cũ.
- Gate phải lưu liên kết đủ để resolve ledger đã dùng. Authorization trỏ đúng gate ID/hash và được kiểm lại theo PMR-02; không tự coi mọi gate ALLOW trong quá khứ là còn hiệu lực.
- Đọc artifact cũ theo phiên bản rõ ràng; thiếu binding cần thiết thì không dùng để cấp quyền mới. Không tái hash âm thầm artifact đã lưu.

Regression: retry nguyên input; đánh giá lại sau 1 giây; health từ stale sang fresh; ALLOW→DENY sau tiêu budget; DENY→ALLOW khi điều kiện hợp lệ thực sự thay đổi; resolve từng gate cũ sau restart; auth không đổi tham chiếu theo gate mới.

Nghiệm thu: các evaluation khác nhau tồn tại song song, immutable; exact retry vẫn idempotent và không làm mới authority.

### PMR-02 — Hiệu lực health, gate và authorization tại lúc dùng

Vị trí: [m11_registry.go](../../lab/affiliate-bot/cmd/bot/m11_registry.go), [mission_command.go](../../lab/affiliate-bot/cmd/bot/mission_command.go), [core/m11](../../core/m11/artifact.go).

Thay đổi cần làm:

- Tách shared validity guard dùng tại gate/authorize/reserve và trước ghi attempt: STOP, activation, current intent/policy/approval, scope/executor, lease, health, cost và ledger head.
- Kiểm thời gian theo thứ tự nhân quả; health không ở tương lai, không quá tuổi; authorization không trước gate; reservation không trước authorized_at; attempt không trước reservation.
- Chốt điều kiện biên expiry/max-age trong contract và áp dụng thống nhất. Đề xuất khoảng hiệu lực nửa mở: tại expires_at hoặc observed_at + max_age thì không cấp operation mới; cập nhật các consumer nếu contract hiện có khác.
- Giới hạn expiry của authorization bởi lease, intent, approval, cost và health; vẫn recheck tại lúc dùng để STOP/rebind/revocation không bị authorization cũ bỏ qua.
- Clock fixture chỉ ở profile offline rõ ràng; đường vận hành không nhận thời gian do caller chọn để lùi đồng hồ mở quyền. Giữ ability replay lịch sử hết hạn mà không cấp quyền mới.

Regression: health tuổi 59/60/61/120 giây; authorize khi fresh rồi reserve sau khi stale; health/gate ở tương lai; reserve trước authorization; attempt trước reserve; approval/cost/lease hết hạn; rebind intent; STOP giữa gate và reserve. Kiểm cả process mới và runtime đã restore.

Nghiệm thu: ca baseline 120 giây bị reject trước mutation; mọi ACK authorization/reservation đều có dependencies còn hiệu lực ở thời điểm dùng, không chỉ hash đúng.

### PMR-03 — Một outcome cho mỗi execution và retry/crash an toàn

Vị trí: [mission_command.go](../../lab/affiliate-bot/cmd/bot/mission_command.go), [m11_registry.go](../../lab/affiliate-bot/cmd/bot/m11_registry.go), [backup_command.go](../../lab/affiliate-bot/cmd/bot/backup_command.go).

Thay đổi cần làm:

- Resolve uniqueness trên toàn store theo execution ID lẫn outcome ID; một execution đã có terminal fixture outcome không được nhận outcome ID khác.
- Giữ exact retry của cùng outcome là duplicate; ID cũ/content khác hoặc execution cũ/ID mới đều conflict. Không tự thay thế outcome đã commit.
- Outcome transition dùng ledger head theo PMR-01; không chọn execution ledger lịch sử để tạo nhánh mới. Loader/backup/restore phải phát hiện dữ liệu duplicate đã có, không chỉ writer mới.
- Chốt giao dịch cho outcome store và ledger: commit marker/journal hoặc cơ chế tương đương có recovery rõ. Retry sau một trong hai lần ghi thất bại không tạo orphan, double-decrement hoặc false ACK.
- Xác định rõ counters consecutive failures với FAILED/NOT_PERFORMED + CANCELLED: không mặc định outcome đã nhận nghĩa là execution thành công. Giữ phân biệt lỗi dispatch, outcome receipt và business outcome.

Regression: khác outcome ID nhưng cùng execution với predecessor cũ/head; exact retry; conflict content; hai process ghi hai outcome ID; crash giữa ledger/outcome/commit; backup/restart/restore dữ liệu hợp lệ và dữ liệu duplicate checksum-valid.

Nghiệm thu: duy nhất một terminal fixture outcome và một committed ledger transition cho mỗi execution; không double-decrement pending hoặc âm pending; reject không mutate store.

### PMR-04 — Graph M11 nối đúng canonical history và toàn chuỗi authority

Vị trí: [backup_command.go](../../lab/affiliate-bot/cmd/bot/backup_command.go), [canonical_resolver.go](../../lab/affiliate-bot/cmd/bot/canonical_resolver.go), [core registry](../../core/m11/artifact_registry.go).

Thay đổi cần làm:

- Dùng canonical resolver để kiểm cycle.decision_id và từng observation_id trong history; yêu cầu unique resolution, integrity và replay MATCH.
- Kiểm tập observation đúng với decision/context tạo intent, không chỉ từng ID tồn tại ở đâu đó. Kiểm liên kết proposal M07 khi intent là agent proposal.
- Kiểm các cạnh đúng cùng một chuỗi: cycle → intent/decision → gate → authorization → execution → outcome → evaluation; lease/version/hash, correlation, thời gian và ledger revision phải khớp cha thực.
- Tái sử dụng validator này khi close-cycle/register, load và backup/restore; không để generic register né kiểm các link bên ngoài registry M11.
- Dữ liệu legacy thiếu cạnh bắt buộc phải được phân loại rõ; không tự điền ID hoặc báo restore đầy đủ khi thực chất không chứng minh được graph.

Regression: ID decision/observation không tồn tại; ID có thật nhưng thuộc decision khác; duplicate history ID; replay DRIFT; cycle trỏ auth/gate của execution khác; evaluation/outcome sai execution; lease/hash/correlation sai. Mỗi backup âm phải có manifest và inner hashes hợp lệ để chắc chắn semantic validator gây reject, không phải checksum gate.

Nghiệm thu: baseline orphan history không còn RESTORED; chuỗi hợp lệ vẫn close/replay/restore được; bằng chứng full chain dùng output thật từ bước trước, không tạo ID tách rời.

### PMR-05 — Restore vào staging, validate rồi mới publish

Vị trí: [backup_command.go](../../lab/affiliate-bot/cmd/bot/backup_command.go), [backup tests](../../lab/affiliate-bot/cmd/bot/backup_command_test.go), [runtime gate](../../lab/affiliate-bot/cmd/bot/runtime_gate.go).

Thay đổi cần làm:

- Tạo staging riêng trong parent đích để cùng filesystem; validate target identity/path, chống symlink/hardlink alias và ghi đè. Không copy canonical artifacts trực tiếp vào runtime đích.
- Copy chính bytes đã được checksum/inventory verify hoặc kiểm lại digest trên bản copy; không bỏ qua lỗi đọc file, không đọc source đã thay đổi rồi tiếp tục dùng manifest cũ.
- Chạy loader/replay và mọi graph gate M00–M11/ACCESSTRADE trên staging, gồm PMR-01/03/04. Sync dữ liệu và metadata cần thiết trước publish.
- Publish atomic, không clobber target xuất hiện hoặc bị thay giữa lúc check và publish. Chốt rõ hỗ trợ đích chưa tồn tại và đích rỗng; nếu phải thay đổi CLI contract thì cập nhật runbook/tests cùng changeset.
- Serialize hai restore cùng đích và không cho consumer coi staging là runtime sẵn dùng. Failure/crash không để một thư mục runtime nửa hợp lệ ở đích.
- Cleanup chỉ staging do operation sở hữu; nếu giữ lại để điều tra thì đặt nhãn incomplete/quarantine và hướng dẫn recovery. Không xóa đích của người dùng hoặc tự gỡ stale runtime lock.

Regression: GRAPH_FAILED/REPLAY_FAILED/STATE_FAILED và lỗi copy giữ đích chưa tồn tại hoặc rỗng như trước; retry bản đúng vào cùng đích; kill ở copy/validate/trước-sau publish; source bị thay đổi; target symlink hoặc xuất hiện đồng thời; hai restore cạnh tranh; happy path status/replay/budget/STOP qua process mới.

Nghiệm thu: không có canonical runtime tại đích nếu chưa qua đủ validators; restore lỗi không để `mission status=VALID` cho dữ liệu bị từ chối; restore thành công không mất durable STOP/approval/usage.

### PMR-07 — Exact-number trong evidence, grounding và render M07

Vị trí: [core/m07](../../core/m07/m07.go), [learner M07](../../lab/affiliate-bot/cmd/bot/m07.go), [watcher adapter](../../lab/affiliate-bot/cmd/bot/watcher.go), [M07 blueprint](../../lab/n8n/M07-readonly-evidence-agent.blueprint.json).

Thay đổi cần làm:

- Dùng decoder giữ json.Number/RawMessage thay vì json.Unmarshal vào any qua float64, ở cả registered tool result → evidence, claim comparison và deterministic render.
- Chốt numeric equivalence: 1, 1.0, 1e0 có được coi tương đương hay không; nếu có thì chuẩn hóa bằng biểu diễn chính xác, không float64. Áp dụng thống nhất với canonical hashing/versioning hiện có; không đổi hash artifact cũ âm thầm.
- Kiểm nested objects/arrays, số nguyên lớn và decimal dài. Dùng strict decoder để reject duplicate keys và dữ liệu sai contract trước khi persistence.
- Audit đường n8n JSON.parse/JSON.stringify và HTTP/model output: JavaScript Number có thể làm tròn trước khi đến Go. Bảo toàn raw JSON hoặc dùng encoding số chính xác được version hóa; nếu chưa hỗ trợ thì reject rõ trước ACK, không báo grounded cho số đã đổi.
- Regression phải đi qua CLI/adapter/proposal persistence/restart, không chỉ helper core. Khi thay blueprint, thêm ca chạy blueprint thật ở môi trường n8n hỗ trợ; nếu chưa có môi trường thì giữ phần n8n chưa nghiệm thu.

Regression: evidence 9007199254740993 vs claim 9007199254740992 phải reject; cùng số lớn phải pass và render nguyên vẹn; số âm lớn/decimal dài/nested values; policy numeric equivalence; tool body raw → store → context → model output → validator; proposal được reload không đổi giá trị/hash.

Nghiệm thu: không có hai giá trị khác nhau được grounded vì làm tròn; raw tool evidence/proposal bảo toàn qua các đường đã hỗ trợ; giới hạn n8n nếu còn phải được ghi rõ.

## 5. Tích hợp test, CI và tài liệu

Mỗi đợt đưa test vào suite ngay, không đợi đợt G mới viết test. Đợt G tổng hợp:

- Mở rộng `core/m07/m07_test.go`, `core/m11/artifact_test.go`, learner mission/backup/fault tests; có thể tách file theo chủ đề để tránh dồn vào smoke.
- Bổ sung ca CLI nhiều process và restart vào `scripts/smoke_br18b_backup_restore.py`; dùng một shared fixture chain cho success path. Ca âm fork workspace từ chính artifacts runtime tạo, không thay bằng JSON giả tối giản.
- Mở rộng `scripts/smoke_br16a_offline.py` để chuỗi chung còn đi hết M00–M11/restore sau thay đổi schema/ledger; kiểm observation/history/authorization/outcome references thực.
- Kiểm `.github/workflows/curriculum-ci.yml` thực sự chạy tests mới, core/learner/harness parity, race và hai smoke. Không suy core tests được chạy chỉ vì learner import core; dependency test suites cần được gọi rõ.
- Mutation một guard mỗi lần trong workspace test cô lập: bỏ head check, bỏ health age, bỏ execution uniqueness, bỏ history link, đổi lại direct publish hoặc float64 comparison phải làm đúng test/job fail. Không merge mutation hoặc test đỏ vào main.
- Cập nhật CLI help, schema version, compatibility/recovery notes, runbook và các plan/readiness refs bị ảnh hưởng. Audit phải giữ các code/test/external gaps còn mở, không tự nâng trạng thái theo tên file test.

Lệnh nghiệm thu dự kiến, chạy từ repo root trong môi trường có quyền loopback (đây không phải log PASS):

```bash
for module in contracts core lab/affiliate-bot lab/mission-runtime; do
  (cd "$module" && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...) || exit 1
done
(cd lab/affiliate-bot && GOWORK=off go test -race -count=1 ./...)
python3 -m unittest discover -s scripts/tests -v
python3 scripts/smoke_br16a_offline.py
python3 scripts/smoke_br18b_backup_restore.py
python3 scripts/validate_n8n_m06.py
python3 scripts/validate_n8n_m07.py
python3 scripts/validate_n8n_m07_adversarial.py
python3 scripts/validate_n8n_m07_output_cases.py
python3 scripts/audit_readiness.py
git diff --check
```

Các script n8n ở trên không tự khởi tạo credential/provider và vì vậy không thay thế engine run. Operated M07 local ngày 09/09 đã PASS synthetic success path qua model, grounding và persistence; `validate_n8n_m07_operated_execution.py` kiểm capture/ACK/proposal store mà không đọc secret. Đây chưa là CI/containerized provider coverage hay full M07 acceptance ngoài fixture. Nếu sandbox không cho mở loopback, ghi test bị chặn, chạy phần không cần port và nghiệm thu phần còn lại trên môi trường được cấp quyền; không sửa test để bỏ qua guard.

**Cập nhật RP-08 (2026-09-09):** `scripts/run_n8n_engine_regression.py`
đã chạy n8n `2.38.1` trong runtime disposable và được wire vào
`n8n-engine-regression` trên GitHub Actions. Nó import copy workflow với
ID/loopback URL và Execute Workflow Trigger chỉ cho CLI, giữ nguyên blueprint
gốc. Regression kiểm M06 append/duplicate/replay và sink failure; M07 dùng
record M06 thật, ép `POST`, registry bật redirect, hoặc adapter không khả dụng;
tất cả phải dừng tại fetch trước Agent/proposal persistence. Ca POST còn yêu
cầu lỗi `TOOL_TRANSPORT_REJECTED`; registry redirect bị strict decoder reject.
Không có credential hoặc provider trong job này, nên M07 success qua model và
received-redirect transport parity vẫn là gap mở.

**Mở rộng M07 model-stub CI (2026-09-09):** runner import một credential OpenAI
disposable chỉ trỏ loopback và chạy endpoint OpenAI-compatible in-process. Bản
copy test thay fetch network bằng `register-tool-result` với fixture synthetic,
nhưng giữ nguyên node Agent, canonical context, grounding và proposal persistence
của blueprint. Stub phải nhận canonical context cùng registered-tool evidence và
trả JSON contract được grounding; runner restart adapter, replay history và gọi
validator bằng proposal/tool trace đã persist. Đây là coverage Agent thật không
secret/provider; không phải provider diversity hay received-redirect transport
evidence.

**M07 received-redirect transport regression (2026-09-09):** production wrapper
luôn tự dựng transport DNS/proxy/redirect guard. Helper nội bộ chỉ nhận
`RoundTripper` trong test để trả response `302 Location` có kiểm soát sau khi
preflight DNS public giả lập. Adapter phải trả `TOOL_TRANSPORT_REJECTED`, không
follow URL thứ hai và không ghi tool-result; test gọi adapter thật, không thay
bằng parser độc lập hoặc endpoint Internet. Đây không là evidence cho một nguồn
redirect live trong n8n engine.

**Mở rộng M06 engine regression (2026-09-09):** cùng runner disposable nay
thực thi copy của M06 blueprint với fixture có JSON key reorder, content đổi
nhưng reuse correlation, URL ngoài profile, và content đổi với correlation mới.
Key reorder phải trả `EXACT_DUPLICATE`; hai fixture không hợp lệ phải dừng tại
HTTP adapter trước ACK/report và bytes canonical history không đổi; fixture mới
hợp lệ phải `APPENDED` với record ID khác rồi replay `MATCH`. Đây là CI coverage
cho fixed synthetic profile, chưa là schedule admission, parser nguồn thật hay
operated selected-source evidence.

**M06 Schedule Trigger regression (2026-09-12):** `run_n8n_m06_schedule_regression.py`
tạo SQLite n8n cô lập, import một copy fixture và chỉ đổi cadence thành một giây,
rồi tạo active-version graph hợp lệ trong database tạm trước khi khởi động `n8n
start`. Regression phải thấy `execution_entity.mode=trigger` thành công, đọc
report thật từ payload n8n, bind ACK/history/record ID, giữ
`execution_permitted=false`, và replay `MATCH`. Ca adapter loopback không chạy
phải tạo trigger execution lỗi tại canonical handoff, không ACK/report và không
ghi history. CI chạy cả hai ca; blueprint checked-in vẫn inactive, fixture
synthetic/read-only và không có provider/source/credential/executor. Đây đóng
gap schedule admission của coverage offline, không phải deployment hay source
selected operated evidence.

## 6. Compatibility, bàn giao và merge gate

Trước khi sửa schema/store, mỗi changeset phải chốt: version mới nếu có, loader hỗ trợ bản nào, cách xử lý snapshot cũ, migration read-only hay explicit command, rollback code sau khi đã ghi format mới có an toàn không. Không tự migrate dữ liệu của người dùng trong lúc review/test. Thay đổi PMR-02 thêm ledger ref bắt buộc vào gate: gate cũ không đủ proof để authorize và sẽ fail closed; tạo lại gate từ ledger hiện hành qua luồng review thay vì rewrite artifact cũ. Bản backup cũ không đủ graph proof không được tự nhận là restore đầy đủ; runtime lịch sử chỉ đọc không được tự cấp quyền mới.

Mỗi gói cập nhật bảng evidence sau; để trống bằng `—` khi chưa có, không điền SHA/CI URL dự đoán:

| Gói | Owner / reviewer | Implementation SHA | Regression trước/sau | CI run / head | Trạng thái |
|---|---|---|---|---|---|
| PMR-01 | — / — | uncommitted | BR-18b: reset/fork/budget | local only | IN_REVIEW |
| PMR-02 | — / — | uncommitted | BR-18b: stale health + stale ledger head; core/harness schema parity | local only | IN_REVIEW |
| PMR-03 | — / — | uncommitted | BR-18b + journal fault/recovery | local only | IN_REVIEW |
| PMR-04 | — / — | uncommitted | BR-18b: broken history graph | local only | IN_REVIEW |
| PMR-05 | — / — | uncommitted | BR-18b + target-gate unit test | local only | IN_REVIEW |
| PMR-06 | — / — | uncommitted | BR-18b + core ledger-link graph | local only | IN_REVIEW |
| PMR-07 | — / — | uncommitted | core + adapter exact-number | local only | IN_REVIEW |

Chỉ đề xuất merge khi:

- [ ] Bảy ca baseline có regression và kết quả sau fix; toàn P1 đã được reviewer xác nhận đóng.
- [ ] PMR-06/07 đã sửa, hoặc có quyết định defer rõ từ chủ repo với giới hạn chức năng và remaining gap công khai; không tự defer.
- [ ] Các ca concurrency, expiry, rebind, duplicate, fault injection, restart và restore chạy trên đúng implementation; reject không làm mở quyền hoặc mất dữ liệu.
- [ ] Full suites và shared chain PASS trong môi trường đủ quyền; CI được xác minh trên đúng head sắp merge, không dùng kết quả commit cũ.
- [ ] Mô tả PR phản ánh toàn diff; plan/readiness/runbook không mâu thuẫn với kết quả thật; không còn ghi chỉ cần external evidence khi vẫn có code gap.
- [ ] Có review lại sau fix; merge/push chỉ thực hiện khi được yêu cầu. Merge không đồng nghĩa production readiness.

Nếu remote vẫn bị chặn, tiếp tục phần local trong phạm vi được giao và ghi rõ thiếu CI/head evidence; không vượt cơ chế duyệt để xác nhận hoặc merge bằng đường khác.
