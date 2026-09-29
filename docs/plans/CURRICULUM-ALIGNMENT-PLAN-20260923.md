# Kế hoạch đồng bộ chương trình học sau đợt cập nhật lớn — 23/09/2026

## 1. Trạng thái, phạm vi và thẩm quyền

- Mã kế hoạch: `CA-20260923`.
- Trạng thái: `IMPLEMENTED_PARTIAL_EVIDENCE_GAPS` — CA-01…CA-07 đã sửa theo CR-01…CR-08; full offline sau review PASS 38/38 trên clean snapshot `5972c2b647556454c495d75387ac248d4826fcce` ngày 29/09/2026 (evidence mục 12). Engine/Schedule Trigger fixture đã PASS riêng tại mục 11. CA-07 còn learner walkthrough; readiness vẫn `NOT_READY_FOR_PRODUCTION`.
- Phạm vi được duyệt: thực hiện CA-00…CA-07 trong file này; giữ nguyên Mission gates, learner progress, external-authority boundaries và readiness boundaries.
- Phạm vi Git được duyệt ngày 29/09/2026: hoàn tất phần kỹ thuật có thể thực hiện, commit/push nhánh `codex/curriculum-alignment-20260929` trên repo gốc và tạo PR vào `main`; không tự merge hoặc push trực tiếp `main`. CA-T18 cần người học thao tác thật, không được đóng bằng test chạy hộ.
- Baseline review: `main` tại `277f1322830cf004e1597cfe97876eb5076ffccc`; working tree sạch trước khi tạo kế hoạch.
- Mốc so sánh tiến độ học: `1faf0c1` — ghi nhận M00.1 PASS ngày 04/09/2026.
- Người lập/thực hiện/self-review: Codex. Người phê duyệt triển khai: chủ repo. Reviewer độc lập chưa có; findings và sửa lỗi self-review được ghi tại evidence mục 10.
- Nguồn thẩm quyền: [CURRICULUM.md](../../CURRICULUM.md). Kế hoạch này không thay thứ tự Mission, điều kiện PASS hoặc quyền thực thi.
- Readiness giữ nguyên: `NOT_READY_FOR_PRODUCTION`. Việc sửa đường học không tự đóng readiness gap.
- Trước triển khai, chỉ có file kế hoạch untracked. Baseline learner-walkthrough regression đã chạy RED với 23 lỗi đúng phạm vi review; hai negative tests độc lập PASS.

Việc triển khai tài liệu/test theo kế hoạch không tự cấp quyền dùng provider, đăng nội dung, chạy live executor, sửa tài khoản hoặc merge PR ngoài phạm vi được giao.

## 2. Kết luận review và mục tiêu chỉnh sửa

Chương trình vẫn phù hợp để học và xây Affiliate Intelligence Bot. Trục bằng chứng → Bot tất định → lịch sử → hành động thủ công/đo lường → AI → automation có kiểm soát nên được giữ. Vấn đề chính là hướng dẫn thực hành và cách trình bày phạm vi chưa theo kịp implementation sau cập nhật.

Mục tiêu của đợt này:

1. Người học làm đúng hướng dẫn M06/M07 không bị chặn bởi sai token hoặc sai expected output.
2. Có đường thực hành có thể tìm thấy từ curriculum chính để đưa packet M00 của người học vào cùng Bot M01–M05.
3. Mỗi chặng có input, thư mục làm việc, lệnh, output, failure case và bằng chứng bàn giao rõ ràng.
4. Phân biệt học kỹ năng trong lab với Mission PASS cần evidence thật; phù hợp phạm vi cá nhân hiện tại mà không hạ gate.
5. Bổ sung regression phát hiện tài liệu lệch implementation; CI xanh vẫn không được gọi là learner PASS.

Không cần học lại BOOT.0, BOOT.1, O00.1 hoặc M00.1 chỉ vì repo đã cập nhật. Nội dung chuẩn và M00–M02 không thay đổi trong diff từ mốc tiến độ học đến baseline review.

## 3. Phát hiện và bằng chứng nguồn

Ưu tiên P1 trong tài liệu này nghĩa là chặn hoặc làm sai bài thực hành; không phải tuyên bố sự cố production.

| ID | Ưu tiên | Phát hiện | Bằng chứng hiện tại | Gói xử lý |
|---|---|---|---|---|
| RV-C01 | P1 | M06/M07 hướng dẫn `AFFILIATE_ADAPTER_TOKEN`, nhưng server và blueprint dùng `CANONICAL_ADAPTER_TOKEN` | [M06.3](../../curriculum/M06/M06.3-n8n-readonly-workflow.md), [M07.3](../../curriculum/M07/M07.3-n8n-evidence-agent.md), [Starter M06](../../starter-kits/M06-readonly-watcher/README.md), [watcher.go](../../lab/affiliate-bot/cmd/bot/watcher.go), [blueprint M07](../../lab/n8n/M07-readonly-evidence-agent.blueprint.json) | CA-01, CA-02 |
| RV-C02 | P1 | Bài M06 yêu cầu quan sát `NEW/UNCHANGED/CHANGED`, trong khi workflow báo persistence bằng `APPENDED/EXACT_DUPLICATE` | [Starter M06](../../starter-kits/M06-readonly-watcher/README.md), [checkpoint](../../starter-kits/M06-readonly-watcher/CHECKPOINTS.md), [compatibility ledger](../../lab/n8n/COMPATIBILITY.md), [blueprint M06](../../lab/n8n/M06-readonly-watcher.blueprint.json) | CA-01, CA-03 |
| RV-C03a | P2 | Bước nhập M00 → M01/M02 đã có nhưng chưa nối trực tiếp vào lesson/starter của đường học chính | [bridge hiện có](../../examples/m00-import/README.md), [M01.4](../../curriculum/M01/M01.4-failure-first-operated-proof.md), [Starter M01](../../starter-kits/M01-deterministic-bot/README.md), [Starter M02](../../starter-kits/M02-history-replay/README.md) | CA-04 |
| RV-C03b | P2 | Starter M03–M05 thiên về chạy demo/kiểm file; thiếu chuỗi lệnh dễ theo dõi để người học nối cùng workspace | [Starter M03](../../starter-kits/M03-tracked-human-action/README.md), [Starter M04](../../starter-kits/M04-grounded-ai-advisor/README.md), [Starter M05](../../starter-kits/M05-reviewed-improvement/README.md), [Continuity Gate](../../starter-kits/CONTINUITY-CHECKPOINT.md) | CA-05 |
| RV-C04 | P2 | Phạm vi personal-local chưa được diễn giải thành các chặng học và giới hạn hoàn thành rõ ràng ở entrypoint | [quyết định personal-local](../architecture/EVIDENCE-PERSONAL-LOCAL-SCOPE-20260921.md), [đường học](../../curriculum/README.md), [Mission M03](../../missions/M03-first-tracked-human-action-outcome.md), [Mission M11](../../missions/M11-production-closed-loop.md) | CA-06 |

### 3.1. Chi tiết cần giữ khi sửa M06

- `canonicalAdapterTokenFromEnvironment()` chỉ đọc `CANONICAL_ADAPTER_TOKEN`; `runWatcherServer()` dừng nếu token không hợp lệ. Không thêm alias cho tên sai chỉ để tài liệu cũ chạy được.
- Node `Require Canonical Store ACK` chỉ nhận success status `APPENDED` hoặc `EXACT_DUPLICATE`, đồng thời kiểm ACK, persisted flag và `record_id`.
- Đoạn ACK đã được chạy riêng bằng Node trong lượt review: hai status trên được nhận; `NEW`, `UNCHANGED`, `CHANGED` bị từ chối. Đây là kiểm đoạn JavaScript cô lập, **không phải n8n engine execution**.
- [TestWatcherRetryAndRestart](../../lab/affiliate-bot/cmd/bot/watcher_test.go) bảo vệ: retry cùng dữ liệu không đổi history; sửa body giữ identity trả `HANDOFF_ERROR`; đổi thời điểm mà giữ correlation cũng lỗi; observation/event mới hợp lệ cần identity phù hợp.
- `NEW/UNCHANGED/CHANGED` vẫn có ý nghĩa ở normalizer/change detection trong [core M06](../../core/m06/m06.go). Không xóa semantics này và không ánh xạ một-một sang trạng thái persistence.
- [validate_n8n_m06_cases.py](../../scripts/validate_n8n_m06_cases.py) hiện kiểm một số ví dụ canonicalization/ACK offline, không thực thi chuỗi server → engine → history. Không dùng PASS của script này thay walkthrough thực tế.

### 3.2. Giới hạn bằng chứng ở thời điểm lập kế hoạch

- Review trước đó trên cùng baseline: 12 validator và readiness audit PASS; hai kiểm tra runtime M07 chưa chạy được vì máy thiếu Go.
- Trước khi tạo file kế hoạch, lượt này đã chạy lại readiness audit trên checkout sạch: PASS về tính nhất quán, 154 scoped claims, product baseline `8f1c8aaf4524`, vẫn `NOT_READY_FOR_PRODUCTION`.
- Sau khi tạo file, audit dừng đúng tại `product baseline audit requires a clean working tree` vì kế hoạch còn untracked/chưa commit. Không gọi đây là audit PASS của bản thay đổi, không bỏ guard và không tự commit để vượt điều kiện. Kiểm riêng tài liệu: 34 liên kết nội bộ tồn tại, code fences cân bằng, không trailing whitespace.
- Checkout `HEAD` và product baseline là hai mốc khác nhau; không tự thay product baseline chỉ vì tạo kế hoạch.
- Python và Node có trên máy hiện tại; tại thời điểm lập kế hoạch chưa tìm thấy Go trong PATH. Chưa chạy full Go/n8n runtime, fresh-machine pilot hoặc provider/live trong lượt này.
- Cập nhật sau khi chủ repo cài toolchain: Go `go1.27.1 windows/386` tại `C:\Program Files (x86)\Go\bin\go.exe` (chạy test với `GOARCH=amd64`, `CGO_ENABLED=1`, `CC=x86_64-w64-mingw32-gcc`) và n8n `2.35.7` tại npx cache đã được xác minh; engine/schedule fixture, race và smoke chạy được. Readiness audit vẫn yêu cầu working tree sạch.
- Khi triển khai phải ghi kết quả mới theo exact commit; các kết quả runtime trên là bằng chứng working tree local, không phải nghiệm thu merge/production.

## 4. Phạm vi và các bất biến phải giữ

### 4.1. Trong phạm vi

- Sửa lesson, starter, checkpoint, evidence template và runbook trực tiếp liên quan bốn nhóm phát hiện.
- Thêm một walkthrough thực hành dùng lại implementation/fixture có sẵn; liên kết từ entrypoint chính.
- Thêm kiểm tra hợp đồng hướng dẫn với code/blueprint, regression cho chính kiểm tra đó và wiring CI phù hợp.
- Làm rõ chặng học, output gần nhất và cách ghi nhận thiếu evidence bên ngoài.
- Ghi evidence nghiệm thu có phạm vi sau triển khai; đồng bộ readiness bookkeeping **chỉ khi** có thay đổi claim/baseline/tham chiếu thực sự.

### 4.2. Ngoài phạm vi

- Viết lại toàn curriculum, thêm Mission/PASS gate hoặc đảo thứ tự M00–M11.
- Đổi ranking formula, JSON/schema, identity/hash, authorization, persistence hoặc fail-closed behavior để khớp tài liệu sai.
- Thêm framework, database, queue, model/provider, live adapter hay môi trường 24/7 mới.
- Thu thập public evidence, tạo affiliate action hoặc làm bài thay learner trong task chỉnh chương trình.
- Tự đánh dấu `PROGRESS.md`, điền evidence M00 còn trống, reset kết quả PASS hoặc gắn fixture thành dữ liệu thật.
- Yêu cầu một người thứ hai/máy sạch như điều kiện đóng phạm vi personal-local; không biến assisted walkthrough thành independent beginner pilot.
- Tự thay branch protection, công bố production readiness hoặc xóa dữ liệu để chạy smoke.

### 4.3. Bất biến kiểm tra trước và sau

1. `CURRICULUM.md` tiếp tục có thẩm quyền cao nhất; học thử capability không tạo quyền bỏ qua gate.
2. `lab/affiliate-bot` là Bot của người học; `lab/mission-runtime` chỉ đối chiếu contract, không thay tích hợp vào Bot.
3. Synthetic/real, Capability/Reality/Operated và tiến độ phát triển/tiến độ học luôn tách biệt.
4. Read-only đối với nguồn bên ngoài không cấm POST có xác thực tới adapter loopback để ghi canonical history; không diễn giải nhầm thành quyền POST ra nguồn bên ngoài.
5. Không đổi missing/pending thành số 0; không bịa commission để ép ranking.
6. Không ghi token vào blueprint, URL, input JSON, output/evidence được chia sẻ hoặc Git.
7. Test bảo vệ behavior cũ chạy trước sửa và chạy lại sau sửa. Nếu cần đổi behavior, tách đề xuất và xin duyệt phạm vi, không gộp ngầm vào docs fix.

## 5. Danh mục công việc và phụ thuộc

Quy mô S/M/L chỉ dùng chia việc, không phải cam kết thời gian. Trạng thái dưới đây là trạng thái sau lượt triển khai hiện tại; `DONE_DOCS_STATIC` không đồng nghĩa với runtime operated evidence hoặc Mission PASS.

| ID | Công việc | Ưu tiên / quy mô | Phụ thuộc | Trạng thái | Điều kiện hoàn tất chính |
|---|---|---|---|---|---|
| CA-00 | Chụp baseline, xác nhận scope và toolchain | P1 / S | Phê duyệt triển khai | DONE_DOCS_STATIC | Baseline và local/hosted blockers có evidence |
| CA-01 | Regression chống lệch hướng dẫn và wiring kiểm tra | P1 / M | CA-00; mở rộng theo từng gói | DONE_DOCS_STATIC | Guard có ca âm, bắt được drift thật, không chỉ đếm marker |
| CA-02 | Sửa token/setup M06/M07 | P1 / S | CA-00, test liên quan CA-01 | DONE_DOCS_STATIC | Lệnh/setup và token wiring thống nhất, không lộ secret |
| CA-03 | Sửa semantics/output/bài thử M06 | P1 / M | CA-00, CA-02, test liên quan CA-01 | DONE_DOCS_STATIC | Phân biệt change detection, identity conflict và persistence ACK |
| CA-04 | Nối thực hành M00–M02 vào đường học | P2 / M | CA-00, test liên quan CA-01 | DONE_DOCS_STATIC | Packet → Bot → history/replay có hướng dẫn và giới hạn rõ |
| CA-05 | Nối thực hành M03–M05 trong cùng workspace | P2 / M | CA-04 | DONE_DOCS_STATIC | Walkthrough và smoke chạy cùng history t1/t2 qua learner CLI, không lấy demo rời thay integration |
| CA-06 | Chia chặng học và diễn giải personal-local | P2 / S | CA-04, CA-05; rà CA-03 | DONE_DOCS_STATIC | Entry points thống nhất, không hạ Mission gate hoặc sửa progress |
| CA-07 | Nghiệm thu, walkthrough và lưu evidence | P1 / M | CA-01…CA-06 | PARTIAL_EVIDENCE_GAPS | Full offline/clean-snapshot acceptance sau review PASS 38/38; Engine/Schedule Trigger fixture PASS riêng; còn learner walkthrough CA-T18 |

Thứ tự đề xuất: `CA-00 → CA-01/02/03 → CA-04 → CA-05 → CA-06 → CA-07`. CA-01 là công việc xuyên suốt: thêm test liên quan **trước** từng nhóm sửa, không merge một gói guard bắt buộc còn đỏ cho các nhóm chưa triển khai.

## 6. Chi tiết từng gói

### CA-00 — Baseline và chuẩn bị

**Việc thực hiện**

1. Ghi branch/commit, working tree, diff so với baseline kế hoạch; bảo toàn thay đổi người dùng nếu có.
2. Đọc lại các file đích và chốt lại finding trên HEAD thực tế; bỏ hạng mục đã được sửa ở commit mới nếu có bằng chứng.
3. Ghi OS/shell và phiên bản Go/Python/Node/n8n khả dụng. Không tự cài hoặc nâng toolchain chỉ để che blocker.
4. Chạy Python/static/audit baseline; Go/engine trên host hoặc CI được phép và đủ toolchain. Ghi `BLOCKED_TOOLCHAIN` nếu chưa chạy được, không ghi PASS.
5. Chuẩn bị workspace lab mới, tách dữ liệu synthetic và dữ liệu learner thật; không dùng n8n instance hoặc history cá nhân đang hoạt động cho mutation test.

**Nghiệm thu:** có bảng lệnh/working directory/expected/observed/commit; biết rõ kiểm tra nào chưa chạy và ở đâu sẽ chạy tiếp.

### CA-01 — Regression cho hợp đồng hướng dẫn

**File dự kiến**

- Mới: `scripts/validate_learner_walkthroughs.py`, `scripts/tests/test_learner_walkthroughs.py`.
- Tích hợp: `scripts/run_offline_checks.py`, `scripts/tests/test_run_offline_checks.py`, `.github/workflows/curriculum-ci.yml`.
- Rà ảnh hưởng nếu cần: `scripts/validate_continuity.py`, `scripts/n8n_change_scope.py`, `scripts/tests/test_n8n_change_scope.py`, `scripts/tests/test_audit_readiness.py`.

**Cách làm**

1. Dùng parser/kiểm tra nhỏ, danh sách tài liệu hiện hành tường minh; không xây hệ thống quản lý khóa học mới.
2. Kiểm tên biến trong **hướng dẫn cấu hình/code block đang áp dụng**, đối chiếu code và blueprint. Không fail vì tài liệu review/lịch sử trích tên sai để giải thích lỗi.
3. Parse blueprint để kiểm token expression, node/endpoint được lesson nhắc và đường ACK/report. Không coi có từ khóa trong văn bản là đủ chứng minh execution.
4. Kiểm link tới walkthrough, input/output/evidence sections và boundary bắt buộc. Quy tắc phải cho phép diễn đạt tương đương, không khóa toàn văn Markdown bằng snapshot dễ vỡ.
5. Thêm negative tests trên bản sao fixture tạm: đổi tên token, xóa bridge link, làm sai expected status, bỏ boundary synthetic/real hoặc chèn hướng dẫn vượt gate; chỉ case liên quan được fail.
6. Trước sửa docs, lưu RED của lỗi đang tái hiện; sau sửa, guard và regression cũ phải GREEN. Assertion không được lấy expected từ cùng văn bản đang kiểm.
7. Nối validator vào offline runner/CI. Nếu thêm file được audit tham chiếu, cập nhật fixture `ReadinessAuditTests.setUp()` tương ứng; không tự thêm claim đã verified.

**Nghiệm thu:** guard bắt được biến thể sai có chủ đích, tài liệu hợp lệ qua được; thay một đoạn lịch sử hoặc ví dụ negative không gây false positive. Job skipped không được báo thành runtime PASS.

### CA-02 — Setup/token thống nhất cho M06/M07

**File sửa chính**

- `curriculum/M06/M06.3-n8n-readonly-workflow.md`.
- `curriculum/M07/M07.3-n8n-evidence-agent.md`.
- `starter-kits/M06-readonly-watcher/README.md`.
- Rà liên kết/setup trong `starter-kits/M07-readonly-evidence-agent/README.md`, `lab/affiliate-bot/README.md`, `docs/architecture/BR-14-M06-OPERATED-RUNBOOK.md`, `docs/architecture/BR-15D-M07-OPERATED-RUNBOOK.md`.

**Thay đổi cụ thể**

1. Dùng `CANONICAL_ADAPTER_TOKEN` nhất quán. Không đổi runtime, không thêm alias `AFFILIATE_ADAPTER_TOKEN`.
2. Ghi rõ chạy từ repo root hay `lab/affiliate-bot`; lệnh start adapter, địa chỉ loopback, đường dẫn history lab và điều kiện thư mục.
3. Hướng dẫn tạo token ngẫu nhiên hợp lệ ở local; adapter và process n8n phải nhận **cùng token**, không tạo lại token khác ở terminal thứ hai.
4. Có hướng dẫn shell macOS và PowerShell riêng cho bước environment/path; không bắt người học Windows copy `export` hoặc `/tmp` mà không giải thích.
5. Phân biệt adapter token với credential model. M06 không cần provider key; M07 thử mặc định theo fixture/model stub trước, credential thật cần cấu hình/quyền riêng.
6. Giải thích n8n expression access tới environment theo runtime đã pin; không tắt guard toàn cục của instance cá nhân/production. Nếu cần cấu hình đặc thù, chỉ dùng instance lab cô lập và nêu rủi ro.
7. Troubleshooting phải phân biệt thiếu/sai token, adapter không chạy, sai cwd/path và model failure; không log token để debug.

**Kiểm tra:** guard CA-01; các test auth/admission hiện có; positive request hợp lệ, missing/wrong Bearer bị từ chối trước mutation; M06/M07 engine fixture với shared token.

**Nghiệm thu:** không còn tên biến sai trong hướng dẫn thực thi hiện hành thuộc scope; walkthrough setup khớp implementation; không có secret literal được commit. macOS/Windows chỉ được ghi là đã chạy khi có evidence đúng OS, không suy từ bản dịch lệnh.

### CA-03 — Sửa bài M06 theo hai lớp semantics

**File sửa chính**

- `curriculum/M06/M06.1-readonly-watcher-contract.md`, `M06.2-idempotency-correlation-observability.md`, `M06.3-n8n-readonly-workflow.md` trong cùng thư mục.
- `starter-kits/M06-readonly-watcher/README.md`, `CHECKPOINTS.md`, `M06-OPERATED-EVIDENCE-TEMPLATE.md`.
- `lab/n8n/COMPATIBILITY.md`; rà `docs/architecture/BR-14-M06-OPERATED-RUNBOOK.md` và `scripts/validate_n8n_m06_cases.py`.
- Code/blueprint hiện tại là đối tượng đối chiếu, không mặc định là file cần sửa.

**Tách rõ trong lesson và checklist**

| Lớp | Nội dung học | Bằng chứng cần quan sát |
|---|---|---|
| Change detection | So sánh nội dung và state `NEW/UNCHANGED/CHANGED` | Core normalizer/conformance case, source và input comparison context rõ |
| Identity/idempotency | Retry cùng event khác với observation mới; reuse identity sai bị chặn | Exact IDs, timestamps, correlation, số record và bytes trước/sau |
| Canonical persistence | Lưu mới hoặc đã lưu chính xác, chỉ ACK khi resolve/replay đạt | `APPENDED/EXACT_DUPLICATE`, record_id, ACK/persisted flags, replay `MATCH` |

**Bài tập cần có**

1. Import fixture hợp lệ lần đầu: `APPENDED`, một record, ACK/persisted true.
2. Retry chính xác cùng event: `EXACT_DUPLICATE`, không tăng record hoặc đổi history.
3. Sửa body nhưng giữ identity của event cũ: reject theo contract hiện tại (`HANDOFF_ERROR` ở learner adapter), không sửa history. Không ghi expected `CHANGED` cho ca này.
4. Tạo observation/event mới với timestamp/correlation/identity hợp lệ: append record mới, giữ record cũ; không suy `APPENDED` có nghĩa giá đã thay đổi.
5. Chạy riêng case change detection với comparison context; phân biệt thay content với đổi thứ tự key theo contract của lớp đang kiểm.
6. Restart và retry; adapter unavailable, token sai, source ngoài profile, input lỗi: không có success ACK/report giả.

Cập nhật evidence template để có trường riêng cho change-detection result nếu đã thực sự chạy, persistence result, record IDs, history count/hash, replay, auth/failure và limitation. Chỉ dùng fixture profile hiện có; selected-source là bài/phạm vi riêng, không đổi URL thành seller source tùy ý. `POST` loopback tới adapter không mâu thuẫn GET-only đối với nguồn bên ngoài.

**Nghiệm thu:** lesson → starter → checkpoint → template → compatibility/runbook thống nhất với output thực tế. Guard không ép runtime trả enum cũ; toàn bộ test identity, retry, auth và no-mutation cũ vẫn PASS.

### CA-04 — Cầu nối thực hành M00–M02

**File dự kiến**

- Mới: `curriculum/PRACTICE-M00-M05.md` — một walkthrough phụ trợ, không phải Mission/lesson ID hoặc gate mới; gói này viết phần M00–M02.
- Liên kết từ `curriculum/README.md`, M01.1/M01.4, M02.4 và README/CHECKPOINTS của starter M01/M02.
- Tái sử dụng `examples/m00-import/README.md`, `docs/architecture/BR-09-EVIDENCE-IMPORT.md`, `docs/architecture/M02-DECISION-ADAPTER.md`, `curriculum/BOOT/GO-JSON-PRACTICE.md`.

**Nội dung walkthrough**

1. Tách fixture rehearsal và dữ liệu thật. M00 vẫn thu thập tối thiểu 3 public observations và lập Human DecisionPacket; không đòi Go/API key để PASS M00.
2. Sau gate M00, giải thích bản chép có cấu trúc `m00-input/v1`: importer không tự parse Markdown hoặc fetch URL; learner tự kiểm mapping từ evidence gốc.
3. Giới hạn v1 chỉ price/commission_rate; evidence khác vẫn giữ trong packet/context, không nhét sai field. Thiếu commission dùng missing/unknown/null kèm provenance, không bịa 0.
4. Tách M01: `evidence import` → kiểm envelope/status → lấy artifact → chạy Bot. `GET_MORE_DATA` hợp lệ khi evidence chưa đủ; chưa bắt learner hoàn tất M02 history để PASS M01.
5. Sau gate M01, phần M02 mới capture t1/t2 → list → restart → replay → xuất DecisionPacket. Giữ subject ổn định, IDs mới đúng từng observation và nguồn/thời gian thực.
6. Giải thích envelope không phải input array, input/history/output phải khác đường dẫn; không redirect đè dữ liệu gốc.
7. Mỗi đoạn có cwd, input sample synthetic, lệnh macOS/PowerShell đã kiểm, expected output, failure case, nơi lưu evidence ngoài dữ liệu tracked.
8. Đưa Go/JSON vào đúng lúc cần hiểu pointer/null/map/JSON và test-first; đây là tài liệu hỗ trợ, không buộc học lại BOOT hoặc phát sinh PASS gate.

**Kiểm tra:** smoke BR-09 và baseline M01/M02; invalid profile, duplicate/source-ID conflict, projection tamper, null-vs-zero, repeat/restart/replay. Kiểm người đọc từ lesson tìm được walkthrough mà không tự tìm trong architecture archive.

**Nghiệm thu:** một rehearsal synthetic chạy thông suốt và một checklist để learner tự nhập evidence thật. Rehearsal không điền M00 working file, không đổi `PROGRESS.md` và không chứng minh learner đã học.

### CA-05 — Thực hành M03–M05 trên cùng learner Bot

**File dự kiến**

- Mở rộng `curriculum/PRACTICE-M00-M05.md`; không tạo một walkthrough thứ hai trùng nội dung.
- README/CHECKPOINTS/evidence templates của starter M03, M04, M05; liên kết tại M03.3, M04.3, M05.2 và `starter-kits/CONTINUITY-CHECKPOINT.md` nếu cần.
- Tái sử dụng [action store](../architecture/BR-10A-ACTION-STORE.md), [outcome store](../architecture/BR-10B-OUTCOME-STORE.md), [mock advisor](../architecture/BR-11A-MOCK-ADVISOR.md), [evaluation store](../architecture/BR-12B-EVALUATION-STORE.md), [proposal/review store](../architecture/BR-12C-PROPOSAL-REVIEW-STORE.md), [acceptance BR-12d](../architecture/BR-12D-ACCEPTANCE.md).

**Chuỗi thực hành**

Mở rộng [smoke continuity `scripts/smoke_br12d.py`](../../scripts/smoke_br12d.py)
để import cả packet t1/t2, chạy M01, capture/replay cùng một history rồi nối
action/outcome/advisor/evaluation/proposal/review M03–M05 trên chính history đó.
Đây là rehearsal harness tạm, không thay lệnh learner tự chạy hoặc cấp learner credit.

1. Tiếp tục history/decision từ phần trước, không khởi tạo fixture rời rồi gọi là continuity.
2. M03: `action record/list` → `outcome import/list`; chỉ ghi action thật sau khi người có quyền thực sự làm và review compliance. Trong rehearsal mọi action/outcome giữ synthetic rõ ràng.
3. Giải thích cửa sổ đo, `PENDING`, số 0 đã đo và snapshot báo cáo đến muộn; không dùng pending để claim kết quả đã chốt hoặc hoàn thành Mission.
4. M04: `advisor mock` với cùng history/action/outcome và as_of/max_age rõ; minh họa stale/future/orphan/hallucinated ID. Mock không yêu cầu key, không thay provider evidence.
5. M05: `evaluation create/list` → `proposal import/list` → `review import/list`; chọn outcome IDs đúng effect, không cộng trùng snapshot. Baseline evaluation có thể `INCONCLUSIVE`, không hứa thuật toán chứng minh hiệu quả kinh doanh.
6. Review record trong rehearsal là synthetic; review thật cần người thực sự review. `APPENDED` không có nghĩa proposal được duyệt và approval không tự apply.
7. Giao một thay đổi nhỏ do learner tự viết, có prediction → regression FAIL → sửa → PASS → diff/rollback proof. Không sửa policy/authority, không làm thay learner rồi tự cấp credit.
8. Mỗi điểm nối ghi previous IDs, learner Bot commit, entrypoint, store owner, expected state và failure result. Demo/harness vẫn dùng để đối chiếu, không thay chuỗi learner CLI.

Các tài liệu BR lịch sử chứa baseline/giới hạn tại commit cũ: xác minh command và behavior với source hiện tại trước khi đưa vào walkthrough; không sao chép trạng thái lịch sử như xác nhận hiện hành.

**Nghiệm thu:** rehearsal M00–M05 giữ cùng artifact lineage, process restart vẫn resolve, không execution bên ngoài; learner biết chỗ nào phải dừng để bổ sung E2/E3/E4 thật. Bài thay đổi nhỏ có bằng chứng tự thực hành riêng, không suy từ CI.

### CA-06 — Chặng học và phạm vi personal-local

**File sửa chính:** `curriculum/README.md`, `ROADMAP.md`, `README.md`; chỉ bổ sung giải thích vào `CURRICULUM.md` nếu thật sự cần và không đổi semantics. Các plan/readiness lịch sử không phải đường học mới.

| Chặng trình bày | Output hữu ích | Điều kiện/giới hạn |
|---|---|---|
| M00–M02 — nền tảng gần nhất | Packet thật, Bot có giới hạn, history/replay | E1 thật; không cần provider, publish hay VPS |
| M03–M05 — vòng học từ kết quả | Hành động người làm → outcome → evaluation → reviewed proposal | Có quyền/kênh/nguồn đo và E2/E3/E4 tương ứng; fixture chỉ rehearsal |
| M06–M07 — automation chỉ đọc | Watcher/Agent nối cùng canonical history | Giữ gate trước đó; source/provider/operated evidence tách khỏi fixture |
| M08–M11 — nâng cao | Shadow/policy → approval → canary → production có giới hạn | Không phải mục tiêu tức thời của lab cá nhân; không coi là đã PASS hoặc bỏ khỏi curriculum |

**Việc thực hiện**

1. Entry point phải trả lời: học bài nào tiếp, mở file nào, tạo output gì, test gì và evidence nào còn thiếu.
2. Giữ [PROGRESS.md](../../PROGRESS.md) là nguồn tiến độ cá nhân; không chép trạng thái M00.1 cố định vào nhiều README dễ lỗi thời.
3. Phân biệt trạng thái nội dung/test của repo với Capability/Reality/Operated của learner. Rehearsal không nâng Mission hiện tại, không mở authority mới.
4. Khi thiếu account/source/outcome, cho biết bài offline hỗ trợ đang làm và gate còn mở; không nói có thể tự chuyển Mission chính thức dù chưa đạt điều kiện.
5. Personal-local scope không đồng nghĩa đổi toàn bộ mục tiêu dài hạn sang synthetic. Nếu muốn đổi thứ tự/gate hoặc mục tiêu chương trình phải là quyết định riêng ngoài kế hoạch này.
6. Giải thích `NOT_READY_FOR_PRODUCTION` không ngăn học M00–M02; việc nội dung đã soạn/CI xanh không chứng minh người mới tự hoàn thành toàn bộ.
7. Không biến independent beginner pilot/máy trắng thành blocker cho sửa tài liệu cá nhân; vẫn giữ các gap đó nếu nói về readiness rộng hơn.

**Nghiệm thu:** ba entrypoint thống nhất với curriculum authority; không xuất hiện nhãn PASS mới, reset progress, yêu cầu credential sớm hoặc chỉ dẫn vượt gate.

### CA-07 — Nghiệm thu và bàn giao

1. Chạy toàn bộ guard/static/regression liên quan; self-review diff theo từng RV-C01…RV-C04.
2. Chạy các command mới trong workspace synthetic cô lập trên môi trường đã công bố. Chỉ ghi shell/OS thực sự được kiểm; host thiếu Go không được claim runtime PASS.
3. Chạy n8n engine fixture theo topology/pin được repo hỗ trợ cho CA-02/03; kiểm restart/ACK/auth/negative paths. Đoạn JavaScript cô lập không thay bước này.
   Điều chỉnh disposable runtime nằm trong `scripts/n8n_runtime_env.py` và
   `scripts/run_n8n_m06_schedule_regression.py`; giữ SQLite pool hợp lệ và đợi
   đủ thời gian startup theo n8n `2.38.1`.
4. Đề nghị chủ repo walkthrough phần M00–M02 khi sẵn sàng; ghi assisted/self-operated rõ. Đây là nghiệm thu trải nghiệm cá nhân, không independent beginner pilot và không tự tạo learner credit.
5. Lưu evidence triển khai dự kiến tại `docs/architecture/EVIDENCE-CURRICULUM-ALIGNMENT-20260923.md`: baseline/exact tested commit, OS/toolchain, lệnh, output đã loại secret, finding→test→file, limitation và phần chưa chạy.
6. Nếu phát sinh readiness claim/ref mới: cập nhật plan/matrix/graph/audit fixture đồng bộ. Nếu chỉ docs alignment, không tự tăng claim count, đổi product baseline hoặc đóng RP/BR ngoài scope.
7. Nếu được phép dùng PR workflow: mỗi PR cần self-review + exact-head checks; sau merge kiểm lại diff/baseline theo contract repo. Không lấy CI trước merge làm bằng chứng sau merge.

**Trạng thái bàn giao phân biệt:** `DOCS_ALIGNED` cho nội dung đã đồng bộ; `VERIFIED_FIXTURE_RUNTIME` cho rehearsal runtime đã chạy. Các nhãn này chỉ dùng trong evidence của kế hoạch, không thay Mission PASS hoặc production readiness. Gói còn thiếu runtime giữ `IN_REVIEW/BLOCKED_TOOLCHAIN`, không đóng DONE toàn bộ.

## 7. Ma trận test và tiêu chí chấp nhận

| Test ID | Ca kiểm | Expected / bằng chứng | Mức kiểm |
|---|---|---|---|
| CA-T01 | Hướng dẫn dùng token sai tên | Guard fail trên bản sao docs sai; pass sau sửa; không sửa server để nhận alias | Static + regression guard |
| CA-T02 | Token đúng, thiếu, sai; caller không tin cậy | Success đúng scope; thiếu/sai bị chặn trước mutation; không lộ token | Go auth/admission + engine |
| CA-T03 | First import và exact retry | APPENDED rồi EXACT_DUPLICATE, một record, ACK/persisted true | Learner CLI/HTTP + n8n |
| CA-T04 | Sửa content giữ identity | HANDOFF_ERROR/reject theo boundary, history bytes không đổi | Go + walkthrough negative |
| CA-T05 | Timestamp mới nhưng correlation cũ / event mới hợp lệ | Case sai bị chặn; event mới hợp lệ append, giữ record cũ | Go + fixture |
| CA-T06 | Change detection với context rõ | NEW/UNCHANGED/CHANGED đúng core contract; không gán nhãn đó cho ACK | Core + docs guard |
| CA-T07 | Adapter fail, false persisted, thiếu record_id | Dừng trước success report; không nhận ACK giả | ACK code + engine negative |
| CA-T08 | Restart adapter/n8n và retry | Exact record resolve, replay MATCH, không duplicate | Engine fixture |
| CA-T09 | Packet v1 → input → Bot | VALID envelope; lấy đúng artifact; missing commission dẫn GET_MORE_DATA | BR-09 + learner runtime |
| CA-T10 | Import field/ID/projection sai, null-vs-zero | Reject/giữ semantics hiện tại; không fabricate dữ liệu | Baseline M00/M01/M02 tests |
| CA-T11 | Capture t1/t2, duplicate/conflict, restart | History giữ cả snapshot hợp lệ; source-ID conflict bị chặn; replay MATCH | BR-09 + M02 |
| CA-T12 | M03 action/outcome lệch decision/effect/window | Reject case orphan/window sai; PENDING không thành zero đã chốt | BR-10a/b/c |
| CA-T13 | M04 mock grounded/stale/future/orphan | Output/abstention/reject đúng; upstream không bị ghi | BR-11a |
| CA-T14 | M05 outcome selection/proposal/review | Links resolve, auto_apply=false, approval không execute, INCONCLUSIVE không thành business success | BR-12d |
| CA-T15 | Entry point/link/prerequisite bị xóa hoặc trỏ sai | Guard fail đúng lỗi; link hợp lệ và boundary được giữ | Docs regression |
| CA-T16 | Rehearsal bị mô tả như Mission PASS/production | Review/guard chặn claim sai; PROGRESS và evidence learner không đổi | Boundary review + diff |
| CA-T17 | Các validator/audit/CI cũ | Không suy yếu assertion để làm xanh; matrix/graph vẫn thống nhất | Python/static/CI |
| CA-T18 | Người học làm lại theo tài liệu | Ghi thao tác tự làm, điểm phải hỏi, expected/observed, limitation | Walkthrough cá nhân, không suy independent pilot |

Không yêu cầu mọi ca đều được chứng minh bằng một loại test. Static checks kiểm hợp đồng tài liệu; Go/engine kiểm behavior; walkthrough kiểm khả năng theo hướng dẫn. Không thay một loại bằng loại khác.

## 8. Lệnh kiểm chứng khi triển khai

Các lệnh dưới dành cho maintainer trong workspace đang áp dụng RTK; RTK không trở thành điều kiện tiên quyết của learner. Mỗi nhóm phải chạy đúng working directory. Dấu `<...>` là placeholder phải thay bằng đường dẫn executable được xác minh, không copy chạy nguyên văn.

`audit_readiness.py` yêu cầu working tree sạch; full offline runner cũng có kiểm tra snapshot tương ứng. Chạy baseline trước sửa và acceptance trên checkout sạch của commit đã được tạo theo workflow được phép. Trong lúc còn diff/untracked, chạy guard/static/regression phù hợp và ghi audit chưa đủ điều kiện; không stage/commit, ignore file hoặc nới audit chỉ để lấy PASS.

Khi chủ repo yêu cầu clean snapshot (lượt 29/09/2026), có thể tạo commit chỉ trong clone tạm ngoài repo chính, chứa đủ tracked diff và file untracked không bị ignore. Phải đối chiếu inventory/hash với working tree, giữ ancestry của product baseline, kiểm snapshot sạch trước/sau và ghi exact snapshot commit. Không stage/commit/push checkout chính; không coi snapshot này là merged-main hoặc hosted CI evidence.

### 8.1. Repo root — baseline và kiểm tài liệu

```text
rtk git status --short --branch
rtk git rev-parse HEAD
rtk proxy python -m unittest discover -s scripts/tests -v
rtk proxy python scripts/validate_repo.py
rtk proxy python scripts/validate_missions.py
rtk proxy python scripts/validate_artifact_spine.py
rtk proxy python scripts/validate_continuity.py
rtk proxy python scripts/validate_language_policy.py
rtk proxy python scripts/validate_n8n_m06.py
rtk proxy python scripts/validate_n8n_m06_cases.py
rtk proxy python scripts/validate_n8n_m07.py
rtk proxy python scripts/audit_readiness.py .
rtk git diff --check
```

Sau khi CA-01 tạo file mới, bổ sung:

```text
rtk proxy python -m unittest scripts.tests.test_learner_walkthroughs -v
rtk proxy python scripts/validate_learner_walkthroughs.py
rtk proxy python scripts/run_offline_checks.py --list
```

### 8.2. `lab/affiliate-bot` — bảo vệ behavior M06/M07

```text
rtk proxy go test ./cmd/bot -count=1 -run "Test(CanonicalAdapterAuthRequiresBearerToken|WatcherAdmissionRejectsUntrustedCallerBeforeMutation|WatcherRetryAndRestart|WatcherRejectAndMissing|M06HTTPAdapterBuildsAndResolvesCanonicalHistory|M07HTTPAdapterRegistersThenResolvesToolEvidence)$"
rtk proxy go test ./...
rtk proxy go vet ./...
```

Trong module `core`, chạy thêm `rtk proxy go test ./m06 -count=1` để bảo vệ change-detection semantics độc lập persistence.

### 8.3. Repo root — smoke liên tục và kiểm đầy đủ

```text
rtk proxy python scripts/smoke_br09.py
rtk proxy python scripts/smoke_br10a.py
rtk proxy python scripts/smoke_br10b.py
rtk proxy python scripts/smoke_br10c.py
rtk proxy python scripts/smoke_br11a.py
rtk proxy python scripts/smoke_br12d.py
rtk proxy python scripts/run_offline_checks.py
```

Full offline runner cần đủ toolchain và checkout phù hợp; đọc plan `--list` trước. Smoke tạo fixture/runtime riêng, không dùng dữ liệu learner. Full runner không thay n8n engine hoặc operated validators.

### 8.4. Repo root — engine và operated evidence

```text
rtk proxy python scripts/run_n8n_engine_regression.py --n8n-cli <n8n-cli-path> --n8n-node <node-path>
rtk proxy python scripts/run_n8n_m06_schedule_regression.py --n8n-cli <n8n-cli-path> --n8n-node <node-path>
```

Dùng versions/topology đang pin trong CI/compatibility ledger, không tự upgrade. Hai operated validators chỉ chạy sau khi có execution capture thật và đúng history/proposal store, theo [runbook M06](../architecture/BR-14-M06-OPERATED-RUNBOOK.md) và [runbook M07](../architecture/BR-15D-M07-OPERATED-RUNBOOK.md).

CI phải thực thi guard mới ở required gate. Nếu path filter làm engine skip khi sửa hướng dẫn token/output, ghi rõ skip và chạy smoke được phép trên exact revision; nếu cần mở rộng selector thì chỉ thêm các path liên quan và regression tương ứng, không giả engine đã chạy.

## 9. Chia đợt triển khai/PR đề xuất

Chỉ áp dụng khi được giao triển khai và cho phép workflow Git tương ứng; task lập kế hoạch không tạo PR.

Chủ repo đã giao commit/push và tạo PR ngày 29/09/2026. Các đợt A/B/C đã được triển khai cùng nhau nên bàn giao bằng một PR trên nhánh `codex/curriculum-alignment-20260929`, tránh tách guard khỏi tài liệu tương ứng. PR phải nêu riêng CA-T18 còn chờ learner; không ghi CA-07 DONE toàn bộ. Kết quả hosted CI phải đối chiếu đúng SHA của PR, không chuyển snapshot local thành CI PASS.

| Đợt | Nội dung | Điều kiện trước khi kết thúc |
|---|---|---|
| A | CA-00 + phần CA-01 cho M06/M07 + CA-02/03 | Auth/output/identity docs thống nhất; RED→GREEN guard; Go/engine evidence hoặc blocker công khai, không merge required check đỏ |
| B | CA-04 + CA-05 + guard/link checks liên quan | M00–M05 rehearsal liên tục; có walkthrough phân chặng và không đổi learner progress |
| C | CA-06 + CA-07 + kiểm đồng bộ cuối | Scope/entrypoint rõ; toàn bộ findings có mapping; evidence không nâng claim quá mức |

Không merge test đỏ độc lập trước PR sửa docs tương ứng. Nếu HEAD có thay đổi khác, rebase/đối chiếu và chạy lại test ảnh hưởng; không reset worktree người dùng.

## 10. Rủi ro, quyết định và rollback

| Rủi ro | Cách kiểm soát |
|---|---|
| Sửa văn bản để khớp output nhưng làm mất ý nghĩa change detection | Tách ba lớp tại CA-03; giữ core tests và identity/no-mutation tests |
| Guard chỉ kiểm marker, không phát hiện câu lệnh hỏng | Parse phần có cấu trúc/blueprint; có negative tests và runtime walkthrough riêng |
| Script bài tập ghi đè dữ liệu thật | Workspace riêng; input/output/store paths tách biệt; không reset/delete history để lấy PASS |
| Hướng dẫn token làm lộ secret hoặc nới quyền n8n | Random token local, secret không ra log, instance cô lập, không đổi write capability |
| Dùng tài liệu BR lịch sử như behavior hiện hành | Đối chiếu CLI/source/test trước khi tái sử dụng; không sửa evidence lịch sử để giả hiện tại |
| Lộ trình lab bị hiểu thành cách vượt Mission gate | Entry point và checklist ghi rõ rehearsal != Mission PASS != authority |
| Thiếu Go/n8n ở local | Ghi BLOCKED_TOOLCHAIN; dùng host/CI được phép; không suy PASS từ static output |
| Readiness bookkeeping lệch sau thêm evidence | Chỉ cập nhật khi claim/ref thay đổi; đồng bộ fixture audit; không tăng status do docs-only |

Quyết định mặc định cho kế hoạch: giữ toàn bộ Mission/gate, lấy macOS personal-local đã được chốt làm phạm vi walkthrough cá nhân, cung cấp chỉ dẫn PowerShell với nhãn kiểm chứng đúng thực tế. Muốn thay mục tiêu dài hạn, mở source/provider/live hoặc bắt buộc hỗ trợ thêm topology phải có yêu cầu riêng.

Rollback dự kiến cho đợt triển khai: revert có chọn lọc commit docs/test đã thay, giữ nguyên dữ liệu/evidence người dùng và runtime contracts. Không dùng rollback tài liệu để xóa history, reset STOP, đổi approval hoặc che kết quả test. Sau rollback chạy lại guard/validator/audit và ghi rõ finding nào mở lại.

## 11. Checklist đóng kế hoạch

- [x] Có phê duyệt triển khai và exact baseline mới trước khi sửa.
- [x] Người thực hiện/self-review và evidence sửa lỗi được ghi; đây không phải independent review.
- [ ] CA-07: còn learner walkthrough CA-T18; full offline/clean-snapshot và engine fixture sau review đã PASS. Commit/PR/reviewer độc lập chỉ áp dụng khi workflow Git được giao theo mục 9; không thêm người thứ hai làm gate cho phạm vi học personal-local.
- [x] Chuẩn bị phiếu bàn giao CA-T18 với expected/observed, trợ giúp, explain-back và tiêu chí review tại evidence mục 13; phiếu trống không phải learner PASS.
- [x] RV-C01: mọi hướng dẫn token hiện hành thuộc scope khớp server/blueprint.
- [x] RV-C02: change detection, identity conflict và persistence ACK không còn bị trộn.
- [x] RV-C03a: từ lesson M01/M02 tìm và chạy được bridge; phần M01/M02 không tạo gate mới.
- [x] RV-C03b: M03–M05 có input/command/output/failure/evidence trong cùng learner workspace.
- [x] RV-C04: chặng học/personal-local được giải thích mà không đổi quyền hoặc điều kiện PASS.
- [x] Acceptance kỹ thuật sau review: full offline runner PASS 38/38 trên clean snapshot `5972c2b647556454c495d75387ac248d4826fcce`, gồm đủ 49 file thay đổi, 211 Python tests, Go test/vet/race, 15 validators, 11 smoke, readiness audit và whitespace check (evidence mục 12). n8n engine/Schedule Trigger fixture PASS riêng tại mục 11; không gộp thành learner walkthrough.
- [x] Go/n8n/runtime evidence ghi đúng host/commit; không gọi static/Node-isolated là engine PASS. Snapshot offline ngày 29/09 được ghi ở evidence mục 12, thay kết quả lịch sử cho diff sau review; audit vẫn `NOT_READY_FOR_PRODUCTION` với 154 claims. Engine và Schedule Trigger fixture PASS trên n8n `2.38.1` + Node `24.21.0`; release admission vẫn `UNVERIFIED`, learner walkthrough và external/production evidence còn mở.
- [x] PowerShell walkthrough import synthetic M00→M01 đã chạy trên Windows: envelope `VALID`, input UTF-8 không BOM, output `RANK_SCENARIO`; không ghi learner evidence.
- [x] Review lại CR-01…CR-08: sửa setup/status/placeholder, guard cấu trúc, teardown, entrypoint và evidence; chạy trực tiếp command blocks PowerShell M00–M05 từ Markdown trong workspace synthetic.
- [x] Không có secret, dữ liệu riêng tư, learner credit tự tạo hoặc thay đổi `PROGRESS.md` ngoài yêu cầu riêng.
- [x] Không nâng readiness, đóng external gap hoặc đòi independent pilot ngoài scope personal-local.
- [x] Các link có thật; file mới được wiring vào runner/CI/audit fixture khi cần.
- [x] Bàn giao nêu phần đã làm, chưa chạy, blocker và bài học tiếp theo; không yêu cầu learner học lại từ đầu.

## 12. Nhật ký kế hoạch

| Ngày | Hoạt động | Kết quả |
|---|---|---|
| 23/09/2026 | Lập kế hoạch từ review chương trình sau cập nhật; đối chiếu source/token/status/identity và chạy lại readiness audit | Tạo tài liệu; triển khai CA-00…CA-07 chưa bắt đầu; giữ `DRAFT_PENDING_APPROVAL` |
| 23/09/2026 | Kiểm tra file kế hoạch và phạm vi thay đổi | Markdown/link checks đạt; chỉ thêm file kế hoạch. Audit sau tạo file dừng vì working tree chưa sạch; không sửa guard hoặc tự commit |
| 23/09/2026 | Triển khai CA-00…CA-06 và phần static của CA-07 theo phê duyệt; thêm learner walkthrough guard, bridge M00–M05 và evidence | 189 Python tests PASS; offline plan và M03–M05 learner CLI markers có assertion; learner/static/link/language checks PASS; Node ACK isolated PASS; Go/n8n runtime `BLOCKED_TOOLCHAIN`; lưu [EVIDENCE-CURRICULUM-ALIGNMENT-20260923.md](../architecture/EVIDENCE-CURRICULUM-ALIGNMENT-20260923.md); chưa commit/PR |
| 23/09/2026 | Kiểm tra lại sau khi cài Go và n8n CLI; sửa teardown Windows có kiểm thử hồi quy trước/sau | Go `go1.27.1` với `GOARCH=amd64`: contracts/core/affiliate-bot/mission-runtime test+vet PASS; smoke BR-08…BR-18b PASS; n8n engine và Schedule Trigger regression PASS bằng n8n 2.35.7/Node 22.23.2; Python suite `190 PASS`; offline runner dừng ở `-race requires cgo` vì chưa có gcc; learner walkthrough thực tế chưa chạy; cập nhật evidence, chưa commit/PR |
| 23/09/2026 | Cài LLVM-MinGW và kiểm tra lại cgo/race/full offline | `CGO_ENABLED=1`, `CC=x86_64-w64-mingw32-gcc`, `GOARCH=amd64`: race PASS; full offline steps 1–36 PASS, dừng bước 37 tại `product baseline audit requires a clean working tree`; cập nhật evidence, chưa commit/PR |
| 23/09/2026 | Audit checkout sạch và rerun theo release pin n8n | Checkout tạm sạch `HEAD=277f1322830cf004e1597cfe97876eb5076ffccc`; `audit_readiness.py .` PASS về mặt kiểm tra với kết luận `NOT_READY_FOR_PRODUCTION`, 154 scoped claims, mở `BR-13, BR-14, BR-15, BR-16a, BR-17, BR-18b, BR-19`. Engine và Schedule Trigger regression PASS trên n8n `2.38.1` + Node `24.21.0`; cập nhật timeout/pool của disposable harness sau false-negative startup timeout; chưa commit/PR |
| 23/09/2026 | Bổ sung liên kết CA-04/05 và continuity rehearsal M00–M05 | Guard bắt 21 entrypoints/hrefs, input/output/failure/evidence table, sai status M06 và bypass instruction; `smoke_br12d.py` nối t1/t2→M01/M02 rồi tiếp tục M03–M05 trên cùng history; PowerShell M00→M01 PASS; chưa ghi learner credit |
| 23/09/2026 | Đồng bộ product-baseline audit và nghiệm thu toàn bộ | Audit chỉ nhận `ROADMAP.md`, compatibility ledger và các script test-harness khi được nêu trong đúng CA-01/05/07; 7 regression tests giữ fail đối với drift không khai báo. Full offline runner PASS 38/38 trên clean snapshot `1531b8f`; audit vẫn `NOT_READY_FOR_PRODUCTION` với 154 claims; 199 Python tests và 101 relative links trên 34 tài liệu PASS; chưa commit/PR |
| 23/09/2026 | Review lại toàn bộ diff theo từng yêu cầu CA, sửa CR-01…CR-08 | Sửa setup n8n/token, output M06, input/cwd/lineage M00–M05, guard cấu trúc/false positives, Windows crash teardown và section-scope audit. Command blocks PowerShell từ Markdown, BR-12d, Go auth/core và regression PASS; kết quả 38/38/n8n cũ được gắn snapshot lịch sử, không claim cho diff mới. Chi tiết và sự cố cô lập ở lượt debug trước: evidence mục 10. Chưa commit/PR |
| 29/09/2026 | Rerun real n8n CLI M06/M07 sau review | Engine regression và M06 Schedule Trigger regression PASS trên n8n `2.38.1` + Node `24.21.0`, SQLite/home tạm, loopback adapter/model stub; chi tiết ở evidence mục 11. Không phải clean-tree/full-offline, learner credit hay production evidence; giữ release admission `UNVERIFIED` và `NOT_READY_FOR_PRODUCTION`. Chưa commit/PR |
| 29/09/2026 | Chốt full offline audit trên clean snapshot chứa toàn bộ diff | Snapshot đầu `591651c` dừng ở bước 15 vì hai dòng English-only trong compatibility ledger; dịch ghi chú sang tiếng Việt, giữ nguyên semantics/guard, rồi rerun từ đầu trên `5972c2b647556454c495d75387ac248d4826fcce`: PASS 38/38, 211 Python tests, 154 scoped claims. Inventory/hash 693/693 file khớp working tree và không đổi sau audit. Chỉ sửa ledger và đồng bộ plan/evidence; không đổi PROGRESS, product baseline, matrix/graph hay source/runtime. CA-T18 còn mở; không stage/commit/push repo chính hoặc tạo PR |
| 29/09/2026 | Chủ repo giao hoàn tất phần còn lại và bàn giao qua Git/PR | Rà lại CA-07: CA-T18 vẫn cần thao tác thực tế của learner; thêm phiếu bàn giao, giữ PROGRESS/gates. Phát hiện hướng dẫn lưu evidence nhầm vào `learner/` không được ignore; thêm regression RED trước sửa và đổi sang `workspace/learner/`. Git workflow được duyệt ở mục 9; exact-head checks/CI phải ghi trên PR, không suy từ snapshot cũ |
