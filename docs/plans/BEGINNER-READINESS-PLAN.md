# Kế hoạch hoàn thiện đường tự học Affiliate Intelligence Bot

- Mã kế hoạch: BR-2026-09.
- Ngày lập: 05/09/2026.
- Trạng thái: IN_PROGRESS — BR-06a và BR-03c.1–c.8 đã merge (#27–#36); BR-03c.8b M11 chain audit IN_REVIEW tại PR #37; BR-06b chờ chương trình/kênh, BR-03c tổng thể chưa hoàn thành.
- Bản gốc được đánh giá: commit `7d2a3ab938a609b43174ae5c38f02ff712b931dc`.
- Cơ sở: [Review ngày 05/09/2026](../../REVIEW-2026-09-05.md).
- Người phụ trách từng đầu việc: theo bảng theo dõi; phải điền khi nhận việc.
- Nguồn quyết định thứ tự học và quyền hạn vẫn là [CURRICULUM.md](../../CURRICULUM.md).

## 1. Kết quả cần đạt

Một người chưa biết terminal hoặc Go có thể làm theo tài liệu để xây **cùng một bot** qua các bước: chuẩn bị môi trường, nhập bằng chứng thật, xếp hạng có giới hạn, lưu lịch sử, ghi hành động thủ công, nhập kết quả affiliate, nhận tư vấn có dẫn nguồn, đánh giá thay đổi, rồi vận hành watcher/Agent chỉ đọc.

Sau khi các Mission và bằng chứng tương ứng đạt, học viên có thể mở rộng bot đó sang execution qua approval, canary có giới hạn và production có thời hạn. Không dùng sandbox hoặc một dấu tick để thay bằng chứng thực tế.

Kế hoạch này quản lý **công việc hoàn thiện repo**. Các đợt phát triển dưới đây không thay thứ tự học O00 → M00 → M11. Ví dụ: tác giả có thể viết watcher bằng fixture trước, nhưng học viên vẫn phải đạt các Mission trước M06 mới chuyển tiếp theo curriculum.

## 2. Phạm vi và các quyết định

### Phạm vi đã chốt cho kế hoạch

- Giữ Go làm baseline hiện tại; giữ n8n cho orchestration tham chiếu.
- Giữ các contract, ID/provenance, permission ceiling và phân biệt Capability/Reality/Operated.
- Bổ sung lệnh/API nhận dữ liệu học viên; demo hard-code tiếp tục là minh họa.
- Có một ví dụ sản phẩm xuyên suốt; có mẫu input/output và bài sửa code theo từng bước.
- Sửa sai lệch đã tìm thấy trước khi mở rộng chức năng phụ thuộc.
- Bộ conformance được dùng để kiểm behavior; shared implementation không được làm test mất khả năng phát hiện lỗi.
- PR lưu kế hoạch ban đầu không triển khai các hạng mục; tiến độ triển khai sau đó được ghi ở bảng theo dõi.

### Giả định làm việc

Ví dụ ưu tiên: giá đỡ laptop cho freelancer/remote worker Việt Nam, phù hợp context đã có trong [PROGRESS.md](../../PROGRESS.md). Đây là ngách tham chiếu; không ghi đè dữ liệu hoặc tiến độ cá nhân.

MVP đầu tiên tạo báo cáo có nguồn và hỗ trợ người thao tác thủ công. Watcher chỉ đọc được bổ sung ở M06. Tạo link affiliate trong ví dụ dùng công cụ chính thức của chương trình được chọn; không giả định phải có API tạo link hoặc API đăng nội dung.

### Những lựa chọn cần ghi nhận khi thực hiện

| Quyết định | Thời điểm phải chốt | Người quyết định | Nếu chưa có |
|---|---|---|---|
| Một chương trình affiliate, cách lấy link và báo cáo hợp lệ | BR-04/BR-06 | Chủ repo/học viên có tài khoản phù hợp | Làm fixture và tài liệu trung lập; chưa claim proof từ platform |
| Một kênh đăng thử và cơ chế gắn campaign/sub-ID nếu được hỗ trợ | BR-06 | Chủ repo/học viên | Chưa chạy action thật |
| Một provider/model dùng cho live AI smoke | BR-11 | Người vận hành | Chạy mock cho Capability, để live evidence chưa xác nhận |
| Phiên bản n8n được hỗ trợ | BR-14 | Người triển khai, người review | Ledger giữ UNVERIFIED |
| Target/action rủi ro thấp có thể làm live adapter | BR-17 | Chủ repo/người cấp quyền | Chỉ hoàn thiện sandbox; E5/E6 chưa đạt |
| Nơi chạy 24/7 và giới hạn tài nguyên/chi phí | BR-18 | Chủ repo/người vận hành | Xác minh local deployment; không mặc định mua VPS |

Các lựa chọn trên không chặn việc lưu kế hoạch hoặc các sửa lỗi độc lập. Chỉ công việc phụ thuộc vào một lựa chọn mới phải chờ lựa chọn đó.

## 3. Cách quản lý đầu việc

- Dùng mã `BR-01` … `BR-19` trong tên nhánh, tiêu đề issue/PR và commit khi triển khai.
- Trạng thái hợp lệ: `TODO → IN_PROGRESS → IN_REVIEW → DONE`; dùng `BLOCKED` khi có blocker cụ thể.
- Bảng bên dưới là nguồn theo dõi trạng thái sửa repo. Nếu tạo GitHub issue thì thêm URL vào cột cuối, không duy trì một bảng trạng thái thứ hai.
- Mỗi item lúc bắt đầu phải ghi người thực hiện, reviewer và PR/issue. Chưa phân công không đồng nghĩa đã có người đang làm.
- Chỉ chuyển DONE khi đáp ứng tiêu chí của item, có diff/test/evidence và được review. Nếu chỉ hoàn thành một phần, tách subtask hoặc giữ trạng thái đang làm.
- Không sửa `PROGRESS.md` để phản ánh tiến độ phát triển. Nó chỉ lưu tiến độ học có bằng chứng.
- Với evidence cần tài khoản: lưu dữ liệu riêng tư ở nơi phù hợp; repo chỉ giữ fixture hoặc bản đã loại thông tin riêng tư và tham chiếu có thể review.
- Đầu việc thay đổi schema/behavior phải cập nhật lesson, mission, starter, checkpoint, eval và runtime bị ảnh hưởng trong cùng PR hoặc chuỗi PR có dependency rõ.

P1: chặn đường tự xây MVP hoặc tính đúng đắn quan trọng. P2: cần để tự học/vận hành đáng tin. P3: hoàn thiện cách công bố và quản lý sau các blocker.

Quy mô S/M/L chỉ là ước lượng tương đối để chia PR: S một phần hẹp; M một capability; L nhiều thành phần hoặc cần live proof. Chốt thời gian thực hiện sau BR-04 và BR-08; chưa ấn định ngày production khi chưa có platform/adapter.

## 4. Bảng theo dõi

Cập nhật sau review: PR #24 → #25 → #26 đã merge đúng thứ tự tại `9102f00`, `f438028`, `ccf6c79`; 4/4 checks mỗi head PASS, tests/vet ba module, 8 validators và 10 Python regressions PASS. BR-05/BR-07 DONE ở phạm vi tài liệu/bài tập đã review; installer trên máy trắng, Windows và năng lực học viên chưa được xác minh, tiếp tục theo dõi ở BR-16. Các đoạn “IN_REVIEW/chờ merge” trong bằng chứng triển khai cũ bên dưới là lịch sử trước lần cập nhật này.

PR #27 đã review và merge tại `09a2f50`, 4/4 checks PASS trên head `af7c844`. PR #28 đã review code/tests và không thấy lỗi chặn trong phạm vi M03, đang cập nhật kế hoạch theo main trước lượt CI mới. BR-03c tổng thể vẫn mở; không suy live proof từ fixtures.

| ID | Đợt | Đầu việc | Ưu tiên / quy mô | Phụ thuộc | Trạng thái | Người thực hiện / PR / evidence |
|---|---|---|---|---|---|---|
| BR-01 | A | Sửa CI marker và đồng bộ checkpoint/tên check | P1 / S | — | DONE | Codex; chủ repo đã yêu cầu merge; [PR #20](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/20) đã merge `b17748a`; [evidence](#br-01--bằng-chứng-triển-khai) |
| BR-02 | A | Sửa measurement window M03 | P1 / S | — | DONE | Codex; chủ repo yêu cầu merge; [PR #21](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/21) đã merge `1e94ec5`; [evidence](#br-02--bằng-chứng-triển-khai) |
| BR-03 | A | Đồng bộ schema và validator output | P1 / M | — | IN_PROGRESS | Codex; BR-03a/b/c.1–c.8 đã merge (#36 `fa9b9db`); BR-03c.8b [M11 chain audit](../architecture/M11-CHAIN-AUDIT.md) IN_REVIEW [PR #37](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/37), reviewer chủ repo chưa review; [audit/phần còn lại](../architecture/ARTIFACT-BOUNDARY-AUDIT.md) |
| BR-04 | B | Chốt MVP và case affiliate xuyên suốt | P1 / S | — | DONE | Codex; chủ repo đã yêu cầu merge [PR #23](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/23); [MVP spec](../product/MVP-SPEC.md); chỉ nghiệm thu đặc tả fixture trung lập |
| BR-05 | B | Quickstart từ máy mới | P2 / M | BR-01 | DONE | Codex; đã review/merge #25 `f438028`; [evidence](evidence/BR-05-QUICKSTART.md); giới hạn installer/Windows/pilot giữ mở ở BR-16 |
| BR-06 | B | Hướng dẫn link, campaign và báo cáo thật | P1 / M | BR-04 | IN_PROGRESS | Codex; BR-06a đã review/merge [PR #27](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/27) `09a2f50`; [hướng dẫn](../product/MANUAL-AFFILIATE-LOOP.md); BR-06b BLOCKED: chưa có chương trình/kênh |
| BR-07 | B | Bài Go/JSON tối thiểu để tự viết adapter | P2 / M | BR-05 | DONE | Codex; đã review/merge #26 `ccf6c79`; [bài Go/JSON](../../curriculum/BOOT/GO-JSON-PRACTICE.md); chưa chứng minh năng lực học viên, pilot thuộc BR-16 |
| BR-08 | C | Tổ chức shared core, CLI và store liên tục | P1 / L | BR-03, BR-04 | TODO | Chưa phân công |
| BR-09 | C | Chuyển M00 packet sang M01/M02 | P1 / M | BR-07, BR-08 | TODO | Chưa phân công |
| BR-10 | C | Tích hợp action/outcome và nhập báo cáo M03 | P1 / M | BR-02, BR-06, BR-09 | TODO | Chưa phân công |
| BR-11 | C | Advisor M04 có mock/live adapter | P1 / M | BR-03, BR-10 | TODO | Chưa phân công |
| BR-12 | C | Đóng vòng evaluation/review M05 | P1 / M | BR-10, BR-11 | TODO | Chưa phân công |
| BR-13 | D | Watcher M06 normalize và lưu history thật | P1 / L | BR-09, BR-12 | TODO | Chưa phân công |
| BR-14 | D | n8n M06 import/smoke/static-data đúng | P1 / M | BR-13 | TODO | Chưa phân công |
| BR-15 | D | M07 grounding và tool boundary thực thi được | P1 / L | BR-11, BR-14 | TODO | Chưa phân công |
| BR-16 | E | Kiểm thử xuyên hệ thống và pilot người mới | P1 / L | BR-05…BR-15 | TODO | Chưa phân công |
| BR-17 | F | Tích hợp M08–M11 và một live adapter giới hạn | P2 / L | BR-16 | TODO | Chưa phân công |
| BR-18 | F | Bài triển khai 24/7, backup/restore/recovery | P2 / L | BR-16; phần ghi phụ thuộc BR-17 | TODO | Chưa phân công |
| BR-19 | E/F | Công bố readiness theo bằng chứng, kiểm soát regression | P3 / M | BR-16; bản production cần BR-17, BR-18 | TODO | Chưa phân công |

Có thể làm đồng thời các item không phụ thuộc nhau. Bảng này mô tả dependency công việc, không giao việc cho agent hay tạo lịch tự động.

## 5. Đợt A — Làm rõ và sửa hành vi chuẩn

### BR-01 — CI marker, checklist và tên required checks

Liên quan phát hiện 9, 10 và ghi chú governance trong review.

- [x] Sửa `scripts/validate_artifact_spine.py` để chấp nhận định dạng tương đương của field action trong template; không bỏ invariant action=null.
- [x] Đổi checkpoint M08 về `intent_mode=PROPOSAL_ONLY` và `execution_authorized=false`.
- [x] Rà starter/templates cho `action_id` dùng sai vai trò của EffectRef và tên cost-bound cũ; giữ alias chỉ ở nơi có giải thích implementation.
- [x] Đối chiếu `.github/workflows/*.yml` và `docs/governance/REPOSITORY-GOVERNANCE.md`; cập nhật tên check thực tế. Thay branch protection là thao tác quản trị riêng, không suy ra từ sửa tài liệu.
- [x] Ghi trạng thái baseline CI trước/sau.

Nghiệm thu: 8 validator chạy độc lập đều exit 0; template action non-null vẫn bị phát hiện; checkpoint khớp canonical schema; không hạ tiêu chuẩn kiểm tra để lấy CI xanh.

#### BR-01 — Bằng chứng triển khai

- Ngày: 05/09/2026. Người thực hiện: Codex. Chủ repo đã yêu cầu merge BR-01; PR #20 merge tại `b17748a501736dcdfed6e2b8abb44e00115d40d6`, cả 4 checks PASS trên head `b86f60f` trước merge.
- PR triển khai: [#20](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/20), nhánh `codex/br-01-ci-contract-alignment`; commit code/tests: `4eb51e0adca0753836ea83f78837ac6ee6b99fc0`. Trạng thái CI theo từng commit xem tại [Checks của PR](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/20/checks).
- Baseline: `8db79ebff7fa275329d5b8dae3923171db42fdac` (main sau PR #19). Chạy lại tại local: 7/8 validator exit 0; `validate_artifact_spine.py` exit 1 vì marker ``action: `null` `` không nhận key được bọc backtick. CI PR #19: job `structure-language-and-foundations` lỗi ở validator này, ba job còn lại PASS.
- Sau sửa tại local (Python 3.9.6): cả 8 validator exit 0; 10 regression tests PASS; `git diff --check` exit 0. Kết quả GitHub Actions của PR triển khai cần kiểm riêng, không suy từ local PASS.
- Lệnh tái hiện: `python3 -m unittest discover -s scripts/tests -v`; `for script in scripts/validate_*.py; do python3 "$script" || exit 1; done`; `git diff --check`.
- Regression chạy validator thật trong repo tạm: nhận key/value dạng thường hoặc inline code; từ chối action object, string `"null"`, boolean, số, array, giá trị rỗng/sai, field thiếu/trùng, section thiếu/trùng và null chỉ nằm ở prose/checklist/section khác. Mutation schema action sang object vẫn exit 1. Không sửa template thật để chạy probe.
- CI đã thêm unittest discovery trước các validator; không bỏ check hiện hữu. Validator chỉ nhận đúng một field action=null trong section Human DecisionPacket, không phải trình phân tích Markdown tổng quát.
- Đã đối chiếu `action-intent.schema.json`, `effect-ref.schema.json`, `evaluation-record.schema.json` và `trusted-cost-bound.schema.json`. `action_id` còn lại trong starter/templates là ID của ActionRecord hoặc đích của `effect_ref.effect_id`, không còn thay field EffectRef của Outcome/Evaluation. Không còn `shadow_only`, `dry_run`, `CanaryCostBound` trong hai thư mục đó.
- Phạm vi: validator/tests/workflow, checkpoint/template M03/M05/M08/M10 và governance. Không đổi schema, Go runtime, branch protection, `PROGRESS.md` hoặc kết luận của review lịch sử; BR-02/BR-03 vẫn TODO.
- Rollback: revert PR BR-01 sau khi được merge; không reset lịch sử hoặc sửa dữ liệu học viên.

### BR-02 — Cửa sổ đo M03

Liên quan phát hiện 6.

- [x] Quy định rõ dữ liệu tạm thời/PENDING khác với kết luận chốt của một cửa sổ đo.
- [x] Thêm ca hồi quy: action 03/09 10:00, window end 10/09 10:00, kết luận NO_OBSERVED_OUTCOME lúc 03/09 11:00 phải bị từ chối hoặc giữ PENDING.
- [x] Kiểm mốc đúng bằng window end, sau window, trước action và timestamp khác timezone.
- [x] Sửa `ValidateActionOutcomeLink`, eval M03, schema nếu cần và bài M03.2/M03.3 nhất quán.
- [x] Nêu rõ cách bổ sung record khi báo cáo đến muộn, tránh overwrite bằng chứng cũ.

Nghiệm thu: probe sai hiện tại chuyển thành test bắt được regression; số 0 sau một cửa sổ đo hợp lệ vẫn được nhận; tài liệu không buộc mọi observation tạm thời phải chờ nếu contract đã cho phép PENDING.

#### BR-02 — Bằng chứng triển khai

- Người thực hiện: Codex; chủ repo đã yêu cầu merge PR #21. Merge commit `1e94ec59385f5983c5130b10b889b083852b7b17`; 4 checks PASS trên head `4aef9f0` trước merge. Baseline triển khai: main `b17748a501736dcdfed6e2b8abb44e00115d40d6`.
- PR triển khai: [#21](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/21), nhánh `codex/br-02-measurement-window`; commit code/tests `159d336fc770cbc07d84fcbc811f5c709fe9781c`. Trạng thái CI theo từng commit xem tại [Checks của PR](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/21/checks); không suy CI PASS từ kết quả local.
- Red/green: thêm `TestM03MeasurementWindow` trước khi sửa runtime; 3 subtest (`early_zero_regression`, `just_before_end`, `early_utc`) FAIL vì nhận VALID thay vì MEASUREMENT_WINDOW_OPEN. Sau sửa cả 19 subtest PASS; thêm `TestM05RejectsPrematureNoOutcome` để chứng minh evaluation không nhận kết luận số 0 quá sớm, nhưng vẫn nhận số 0 tại end.
- Eval M03 thêm E08–E15 (8 case, tổng 15); các case cũ không bị bỏ hoặc nới expected. Quy tắc mới chỉ chặn `NO_OBSERVED_OUTCOME` trước end; giữ quan sát tạm/giao dịch hợp lệ. So sánh timestamp bằng `time.Time`, nhận đúng bằng end và offset tương đương.
- Kiểm tại local: `go test ./...` và `go vet ./...` trong cả `lab/mission-runtime` và `lab/affiliate-bot`; 8 Python validator và 10 Python regression tests; `git diff --check`. Tất cả exit 0 (affiliate-bot chạy lại với GOCACHE tạm vì cache mặc định bị sandbox chặn). Lệnh chạy riêng: `go test ./cmd/demo -run 'TestM03(MeasurementWindow|EvalPack)|TestM05RejectsPrematureNoOutcome' -v`.
- Đồng bộ runtime, eval, annotation schema status, bài M03.2/M03.3, mission và starter/checkpoints/evidence template. Không đổi JSON shape/enum; schema đơn lẻ không thể so sánh với ActionRecord nên runtime thực thi điều kiện cross-record.
- Giới hạn: không thêm store/CLI nhập dữ liệu, không cưỡng chế append-only, không claim kết quả kinh doanh hay Operated PASS. Hướng dẫn record bổ sung và chọn snapshot nằm trong M03.2; BR-03/BR-10 vẫn chưa triển khai. Không đổi quyền machine execution.
- Rollback: revert PR BR-02 nếu được merge; lưu ý việc đó đưa lỗi kết luận sớm trở lại.

### BR-03 — Output đúng schema trước semantic validation

BR-03c.6 M09: Codex thực hiện, reviewer chủ repo chưa review; nhánh `codex/br-03c-m09-boundary`, [phạm vi/evidence](../architecture/M09-JSON-BOUNDARY.md). #32 đã review/merge `9aae6ad`, CI 4/4 trên head `c407a43`, tests/vet mission-runtime PASS. M09 lần này chỉ raw profile/consistency audit, không cấp authorization mới hoặc thay persistence/executor. Các trạng thái chờ merge cũ là lịch sử.

BR-03c.5 M08: Codex thực hiện, chủ repo chưa review; nhánh `codex/br-03c-m08-boundary`, [phạm vi/evidence](../architecture/M08-JSON-BOUNDARY.md). #31 đã review và merge `d759fc0`, CI 4/4 trên head `0fca001`, tests/vet mission-runtime PASS. Các trạng thái chờ merge cũ là lịch sử. BR-03 tổng thể còn mở.

BR-03c.4 M07 đang triển khai: Codex, reviewer chủ repo chưa review, nhánh `codex/br-03c-m07-boundary`; [phạm vi/evidence](../architecture/M07-JSON-BOUNDARY.md). #30 đã review và merge `891562a`, 4/4 CI trên head `72ba4e8`, tests/vet mission-runtime PASS. Không thấy lỗi chặn trong phạm vi normalizer offline; ID/provenance không áp dụng migration store. Các trạng thái cũ bên dưới là lịch sử.

BR-03c.3: Codex thực hiện, reviewer chủ repo chưa review, nhánh `codex/br-03c-m06-boundary`; [M06 boundary/evidence](../architecture/M06-JSON-BOUNDARY.md). #29 đã review và merge `7a0f6c4`, 4/4 checks PASS trên head `50f3200`; tests/vet M03–M11 PASS. M06 chỉ schema/profile/normalizer offline, không thay BR-13/14. Các trạng thái chờ review trước đó bên dưới là lịch sử.

BR-03c.2: Codex thực hiện, reviewer chủ repo chưa review, nhánh `codex/br-03c-m05-boundary`; [phạm vi/evidence M05](../architecture/M05-JSON-BOUNDARY.md). Nối raw EvaluationRecord/ImprovementProposal/ReviewRecord và chuỗi M03 vào CLI chỉ đọc; không apply/persist. BR-03c.1 đã merge `d89a04c` sau review, giải quyết conflict kế hoạch với #27 và chạy lại tests/vet/validators; 4/4 checks PASS trên head `1f4c5bc`. #27 merge trước tại `09a2f50`. Những đoạn chờ review/merge trước đó là lịch sử.

BR-03c.1: Codex thực hiện, reviewer chủ repo (chưa review), nhánh `codex/br-03c-m03-boundary`; [thiết kế/evidence](../architecture/M03-JSON-BOUNDARY.md). Phạm vi raw M03 action/outcome + CLI read-only + raw eval + output schema; không sửa alias M11 hoặc triển khai store. BR-03c tổng thể vẫn IN_PROGRESS, không DONE chỉ từ M03.

Liên quan phát hiện 7, 8.

- [x] Tái hiện AdvisorOutput ADVISE reason rỗng hiện vẫn SUPPORTED; bổ sung test bắt lỗi này.
- [x] Lập bảng type Go ↔ schema JSON ↔ input/output CLI cho các artifact được dùng (bản audit, không claim các dòng đã conformance).
- [ ] Kiểm required field, enum, null/array, unique IDs, timestamp và field ngoài schema tại boundary.
- [x] BR-03b đề xuất giữ M02 projection + adapter sang DecisionPacket, có kiểm schema/replay và ngoại lệ ranked:null cũ; chờ review thiết kế trong PR triển khai.
- [ ] Ghi rõ kiểm schema không thay kiểm liên kết, freshness hoặc tính hỗ trợ của nội dung bằng chứng.
- [ ] Chọn cách validation nhỏ nhất đáp ứng contract; nếu thêm dependency phải pin và ghi cách cài.

Nghiệm thu: output sai schema bị từ chối trước khi SUPPORTED; output thực của runtime được kiểm bằng schema tương ứng; fixture history hiện tại vẫn replay hoặc có quyết định chuyển đổi rõ.

#### BR-03a — Advisor boundary (đã merge PR #22 tại 487ed73; item cha chưa hoàn thành)

- PR: [#22](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/22); code/tests commit `274e60676533fe3fd05f7630e4b0dd70331c8068`. [CI theo commit](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/22/checks).

- Baseline `1e94ec5`: `TestAdvisorRejectsEmptyReason` FAIL vì nhận SUPPORTED thay vì INVALID_SCHEMA; sau sửa PASS.
- Đã thêm raw JSON boundary cho M04, giữ required/null/duplicate/unknown information trước typed decode; enum/type/unique IDs/reason/constant false được kiểm. Guard typed không thay raw boundary. Test snapshot canonical schema buộc review khi schema đổi; đây là validator riêng cho M04, không phải engine schema tổng quát.
- Có lệnh `advisor-check` nhận output/context từ file, xuất artifact đã kiểm cùng semantic result; CI chạy smoke fixture. 6 case eval mới; eval giữ JSON gốc. Test CLI handler kiểm artifact serialize lại, lỗi không được xuất success envelope; test freshness/linkage/provenance vẫn giữ độc lập.
- Kiểm chứng local: Go tests/vet + CLI fixture ở mission-runtime; 8 Python validators, 10 Python regression tests và diff check PASS. Không đổi history M02; adapter/replay conformance sâu hơn vẫn thuộc BR-03b.
- [Audit artifact và quyết định bảo toàn M02](../architecture/ARTIFACT-BOUNDARY-AUDIT.md) ghi BR-03b/c còn TODO. Không đánh dấu DONE chỉ vì M04 PASS.
- Chủ repo xác nhận chưa có chương trình affiliate: BR-04/BR-06 sẽ dùng fixture trung lập, chưa có platform/operated proof; không tự đăng ký chương trình hoặc chọn kênh đăng thay chủ repo.

#### BR-03b — History schema + DecisionPacket adapter (đã merge)

Review và merge PR #24 tại `9102f00` theo yêu cầu chủ repo; 4/4 head checks PASS. Các mô tả chờ review bên dưới là lịch sử triển khai, không phải trạng thái hiện tại. BR-03c vẫn TODO.

- [PR #24](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/24); commit triển khai `a9d80726b1f66f1de42d652db89b1a1e161e1464`; [CI theo commit](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/24/checks).

- Baseline: `eacbd27`. Người thực hiện: Codex; reviewer: chủ repo, chờ review. [Thiết kế/mapping](../architecture/M02-DECISION-ADAPTER.md); [schema runtime](../../contracts/README.md).
- Package contracts nhúng schema canonical, pin jsonschema/v6 v6.0.2 và go.sum; format assertions bật, không tải schema ngoài. Compile schema không thay việc nối runtime từng Mission (BR-03c).
- Capture/load/append M02 kiểm raw schema + field exact/duplicate/unsupported trước semantic ID/hash/time checks. Context xuất packet bắt buộc tường minh, không tự tạo fact; packet phải qua canonical schema và history phải replay=MATCH.
- Record mới serialize ranked rỗng thành []; chỉ ranked:null lịch sử được kiểm qua view [] và so sánh replay tương đương, giữ file cũ/input hash. Không nới canonical schema, không rewrite history. Thiếu ranked và các dạng sai khác vẫn bị từ chối.
- Kiểm chứng: tests/vet của contracts, affiliate-bot và mission-runtime PASS; 8 Python validators và 10 regression tests PASS; diff check PASS. Cache Go mặc định bị sandbox chặn nên chạy lại với GOCACHE riêng ở tmp.
- Smoke binary mới: capture sample → APPENDED, replay → MATCH, history decision → packet giữ exact obs-a-1/obs-b-1, state RANK_SCENARIO và action:null. CI thêm cùng smoke export. Tests kiểm context/raw mutations, strict schema, output serialization, read-only export, legacy replay và mutation không alias dữ liệu gốc.
- Không claim M03–M11 đã dùng bộ kiểm schema mới; BR-03c và item cha BR-03 còn mở. Rollback bằng revert PR sau merge, không reset dữ liệu học viên.

## 6. Đợt B — Người mới bắt đầu được và hiểu sản phẩm đang xây

### BR-04 — Chốt MVP và ví dụ tham chiếu

Liên quan phát hiện 2, 11.

- [x] Viết `docs/product/MVP-SPEC.md` với audience, câu hỏi, input, output hữu ích, giới hạn và thước đo thành công (chủ repo đã yêu cầu merge đặc tả).
- [x] Dùng context giá đỡ laptop làm ví dụ mặc định; chủ repo xác nhận chưa có chương trình, nên dùng fixture trung lập, chưa chọn platform/kênh hoặc claim quyền truy cập.
- [x] Mô tả kết quả tối thiểu: báo cáo có source/time/limitation, link tới hành động thủ công, số liệu đo và evaluation.
- [x] Phân biệt tiêu chí kỹ thuật với KPI kinh doanh; không yêu cầu phải có doanh thu để chứng minh code hoạt động.
- [x] Chốt các trường và ID cần xuyên suốt case trong bản đặc tả.

Nghiệm thu: một reviewer đọc spec có thể nói được bot giúp ai làm việc gì, đầu vào lấy ở đâu, đầu ra trông thế nào và thế nào là hoàn thành MVP.

Evidence BR-04: [MVP-SPEC](../product/MVP-SPEC.md) và [REFERENCE-CASE](../product/REFERENCE-CASE.md). Chỉ là đặc tả/mẫu nội dung, không phải pipeline/CLI đã triển khai. Chủ repo xác nhận chưa có chương trình affiliate; BR-06 sẽ hướng dẫn trung lập và giữ platform proof chưa xác nhận. Chủ repo đã yêu cầu xử lý conflict và merge PR #23; BR-04 DONE ở phạm vi đặc tả, không phải MVP đã chạy hoàn chỉnh.

PR: [#23](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/23), commit đặc tả `7fc5baa2636feb618cc6c88cacfd90af1f0ddddc`; 8 validator và diff check local PASS. [CI theo commit](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/23/checks). PR độc lập với [#22 — BR-03a](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/22); cập nhật BR-02 DONE và phần còn lại BR-03 nằm ở #22, không ghi đè tiến độ đó khi gộp kế hoạch.

### BR-05 — Quickstart từ máy chưa cài công cụ

Liên quan phát hiện 11.

- [x] Viết quickstart có link cài Git/Go/editor, phiên bản được kiểm và nhánh lệnh theo macOS/Windows/Linux; phân biệt smoke clone/cache với cài toolchain trên máy trắng, Windows chưa smoke.
- [x] Có lệnh clone đầy đủ repo, cd, git status, go version, go run, go test và output dự kiến.
- [x] Giải thích vị trí repo root, module root, path, cách tạo/sửa/lưu file.
- [x] Có xử lý các lỗi command not found, PATH, sai thư mục/module và Go version.
- [x] Dẫn tới BOOT-REFERENCE cho bài intentional failure, đồng thời nêu chính xác file/dòng logic cần sửa; harness kiểm lỗi đó trong clone riêng.

Nghiệm thu: trên môi trường sạch thuộc profile hỗ trợ, người thử đi từ clone đến test PASS theo tài liệu mà không phải hỏi thêm lệnh bị thiếu. Không đánh dấu OS chưa thử là đã hỗ trợ đầy đủ.

Triển khai BR-05: [quickstart](../../curriculum/BOOT/QUICKSTART.md), [smoke/evidence](evidence/BR-05-QUICKSTART.md). macOS clone sạch/cache rỗng PASS vòng run→test→intentional assertion FAIL→fix→PASS→O00; CI thêm smoke Ubuntu. Installer và pilot người mới chưa được kiểm độc lập, Windows chưa smoke. PR dựa trên BR-03b để hướng dẫn đủ contracts module; không merge trước PR #24. Giữ IN_REVIEW, không tự claim full-machine onboarding hoặc Mission PASS.

[PR #25](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/25), base là nhánh BR-03b của [#24](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/24); commit triển khai `59fb37ba950abe32eb2a76f0d4aa11bfc9891809`. [CI theo commit](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/25/checks). Sau merge #24, đổi base #25 về main và kiểm CI trước khi merge tiếp.

### BR-06 — Một vòng affiliate thủ công có số liệu

Liên quan phát hiện 2.

Tách phạm vi để không dùng fixture thay live proof:

- BR-06a: bản trung lập [MANUAL-AFFILIATE-LOOP](../product/MANUAL-AFFILIATE-LOOP.md), [fixture](../../examples/affiliate-manual/manual-loop.json), test `TestManualAffiliateFixture` trong contracts. Người thực hiện Codex; reviewer chủ repo (chưa review); nhánh `codex/br-06-manual-affiliate-loop`. Không có importer/runtime mới.
- BR-06b: chương trình/kênh cụ thể, màn hình export và case thật — BLOCKED vì chủ repo xác nhận chưa có chương trình. Không tự tạo account hoặc hành động ngoài hệ thống.
- Test kiểm canonical action/outcome schemas và mapping fixture, thêm 7 mutation regressions: missing→0, zero trước end, orphan action, update sai, field lạ, số âm, duplicate snapshot. Đây không phải semantic runtime M03 hoặc proof affiliate.
- Kiểm local BR-06a: tests/vet ba module PASS, 8 validators và 10 Python regressions PASS; diff check PASS. Test fixture được job `deterministic-runtime` hiện hữu chạy qua `go test ./...` trong contracts. Không suy trạng thái CI từ kết quả local.
- Checklist gốc dưới đây giữ mở cho nghiệm thu đầy đủ trên chương trình đã chọn. BR-10 có thể dùng fixture trung lập nhưng không được claim live proof.

- [ ] Hướng dẫn dùng chương trình được chốt: điều kiện tài khoản, mở sản phẩm/offer, tạo link chính thức và kiểm link thuộc tài khoản đúng.
- [ ] Giải thích campaign/sub-ID hoặc cơ chế tương đương chỉ khi chương trình hỗ trợ; không giả định UTM thay thế affiliate attribution.
- [ ] Có ví dụ nội dung/kênh đăng thủ công, disclosure phù hợp và cách ghi ActionRecord.
- [ ] Chỉ cụ thể màn hình hoặc chức năng xuất báo cáo, timezone, kỳ đo, trạng thái đơn và hoa hồng.
- [ ] Lưu mẫu export đã loại dữ liệu riêng tư hoặc fixture được gắn nhãn synthetic; có bảng map sang OutcomeRecord.
- [ ] Mô tả báo cáo rỗng, số 0, pending, đơn hủy/hoàn và commission chưa thanh toán.

Nghiệm thu: một case đi từ link đến báo cáo được review; mỗi metric nói rõ nguồn và giới hạn. Nếu chưa có tài khoản/nguồn thật, phần hướng dẫn có thể hoàn thành bản nháp nhưng live proof chưa được đóng.

### BR-07 — Kiến thức Go/JSON đủ để mở rộng bot

Liên quan phát hiện 1, 11.

- [x] Thêm bài ngắn gắn trực tiếp bot: struct và JSON tag, array/map, pointer/null, hàm, error, đọc file và test assertion.
- [x] Một bài chỉnh input rồi đọc error; một bài thêm field clicks và hàm normalize nhỏ.
- [x] Mỗi bài có phiên bản trước/sau, vị trí chỉnh, lệnh chạy và lỗi cố ý; source giữ bản trước để học viên tự viết thay đổi.
- [x] Cho biết bài nào là prerequisite của adapter M03/M04/M06.

Nghiệm thu: học viên tự sửa được một adapter nhỏ và viết test phân biệt null/0; không chỉ chép toàn bộ solution mà không hiểu input/output.

Triển khai BR-07: [GO-JSON-PRACTICE](../../curriculum/BOOT/GO-JSON-PRACTICE.md), package `lab/affiliate-bot/internal/learning` trong cùng module (không phải Bot/platform adapter thứ hai). Fixtures synthetic, chưa là OutcomeRecord; không đổi production behavior. Go tests/vet toàn affiliate-bot và 8 validators PASS. Tests 4 nhóm kiểm null/absent/0/1, input lỗi/type/negative/unknown/duplicate, normalize status và đọc fixture. Bài học yêu cầu tự dự đoán, tạo lỗi, thêm Clicks kèm test trước/sau và explain-back; chưa có learner pilot độc lập nên giữ IN_REVIEW, không ghi PROGRESS hoặc claim năng lực học viên từ CI. PR phụ thuộc BR-05 (#25), sau BR-03b (#24).

Smoke bài tập trong bản sao tạm riêng: bài A đổi valid_orders từ 0 sang string "0" làm TestCommittedReportFixture FAIL với cannot unmarshal string into int. Bài B thêm TestClicksExercise trước code hỗ trợ làm FAIL unknown field clicks; sau khi thêm pointer field, guard số âm, metric có điều kiện và đổi test unknown cũ sang typo_clicks, full tests PASS (gồm clicks null/absent/0/-1/string). Không đưa solution sau bài B vào source mẫu; source vẫn để học viên tự thực hành, không sửa file cá nhân để chạy probe.

[PR #26](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/26), base nhánh BR-05 (#25), commit triển khai `36e47f1f054399512a708b70253592b3ca56be03`; [CI theo commit](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/26/checks). Review/merge thứ tự #24 → #25 → #26 và đổi base về main sau khi PR trước đã merge; không gộp ngược để né review dependency.

## 7. Đợt C — Cùng một bot từ M00 đến M05

### BR-08 — Shared core, entrypoint và quyền sở hữu state

Liên quan phát hiện 1, 8.

- [ ] Viết ADR trước khi refactor: giữ một module hay tách shared module/package; chỉ rõ import graph, store owner và conformance boundary.
- [ ] Đưa capability cần dùng lại ra khỏi `cmd/demo/package main` hoặc tạo adapter có boundary rõ; tránh copy toàn bộ runtime thành bản thứ hai.
- [ ] Giữ expected outcomes/cases độc lập; không kiểm một hàm bằng cách so nó với chính nó.
- [ ] Thiết kế CLI thống nhất nhận file/context, output có cấu trúc, exit code rõ và không phụ thuộc dữ liệu hard-code.
- [ ] Tách lệnh demo, validate, persist và execute; tên lệnh không gây hiểu lầm về side effect.
- [ ] Có hướng dẫn từng bước để học viên mở rộng cùng workspace từ baseline.

Nghiệm thu: lệnh demo cũ vẫn dùng được hoặc có cập nhật đường học đồng bộ; một capability M03 nhận file của người học qua entrypoint thật; full regression qua; store owner được ghi rõ.

Các tên lệnh dưới đây là **giao diện đề xuất, chưa tồn tại**, chốt tại ADR trước khi đưa vào bài học:

    bot evidence import <input>
    bot history capture|list|replay ...
    bot action validate|record <input>
    bot outcome import <input>
    bot advisor run --context <input>
    bot evaluation record <input>
    bot review record <input>

### BR-09 — M00 evidence sang M01/M02 có hướng dẫn và converter

Liên quan phát hiện 8.

- [ ] Thêm một packet mẫu đầy đủ trong khu vực examples mới, có nguồn/gắn nhãn đúng và không lẫn file evidence cá nhân.
- [ ] Công bố bảng map subject_id/product_id, field claim/value, source/time và dữ liệu domain.
- [ ] Cung cấp converter/importer; giữ unknown, pending và missing theo semantics, không tự điền commission.
- [ ] Chỉ rõ khi nào nhiều field observation được nhóm thành product input và provenance từng field được giữ thế nào.
- [ ] Xuất recorded_result/DecisionPacket qua adapter đã chốt ở BR-03.
- [ ] Bài thực hành hai lần quan sát t1/t2 và restart/list/replay dùng chính dữ liệu đã nhập.

Nghiệm thu: packet → input → history → quyết định resolve được ID; thiếu commission trả trạng thái trung thực; duplicate/conflict và replay có test; có ví dụ lệnh lẫn output đầy đủ.

### BR-10 — M03 action/outcome importer vào bot

Liên quan phát hiện 1, 2, 6.

- [ ] Lệnh ghi ActionRecord chỉ nhận decision_id tồn tại; phân biệt human record với machine execution.
- [ ] Import format báo cáo của BR-06 qua adapter riêng, lưu source/time/window.
- [ ] Nối EffectRef với đúng action; không đồng nhất đơn phát sinh với đơn hợp lệ/payment.
- [ ] Chạy các case trước window, số 0, pending, late update, duplicate và orphan ID.
- [ ] Hướng dẫn file cần sửa, cách chạy validator và cách đọc lỗi theo case affiliate.

Nghiệm thu: action và outcome thật hoặc fixture được gắn nhãn đi vào cùng store; list/query chứng minh continuity; proof thực tế do học viên thực hiện riêng, không được sinh ra từ fixtures.

### BR-11 — M04 advisor có adapter thực

Liên quan phát hiện 1, 4, 7.

- [ ] Tạo mock provider chạy offline để bài học không cần credential ngay.
- [ ] Thêm một live provider adapter có cấu hình mẫu không chứa secret; provider/model phải ghi lại khi smoke.
- [ ] Tạo context từ history/outcome hiện có, có evidence payload, exact IDs, source, as_of, max_age và limitation.
- [ ] Output có cấu trúc → schema validation → reference/freshness validation → ADVISE/HUMAN_REVIEW/ABSTAIN.
- [ ] Có bounded timeout/retry và thông báo lỗi cấu hình; không có write tool ở M04.
- [ ] Giải thích ID hợp lệ chưa chứng minh mọi phát biểu đúng; case kiểm nội dung vẫn cần expected evidence và review.

Nghiệm thu: chạy được một câu hỏi với artifact của BR-10; hallucinated ID, stale/future evidence, reason rỗng, malformed output và write request không được SUPPORTED. Live proof có model/prompt/context version; mock không được thay live proof.

### BR-12 — M05 evaluation, đề xuất và review

Liên quan phát hiện 1, 2.

- [ ] Tạo EvaluationRecord từ exact decision/effect/outcome chain.
- [ ] Sinh hoặc nhập ImprovementProposal với version, benefit, risk và rollback.
- [ ] Lệnh nhập Human ReviewRecord; giữ auto_apply=false.
- [ ] Viết một bài thay đổi nhỏ đã review: regression FAIL trước sửa, PASS sau sửa và rollback về version trước.
- [ ] Có case dữ liệu ít → INCONCLUSIVE/thu thập thêm, không ép kết luận “cải tiến hiệu quả”.

Nghiệm thu: đi trọn M00 → M05 trong một workspace với artifact links resolve được; proposal không tự thay code/policy; lưu diff, review và kết quả đo riêng với kết quả test.

## 8. Đợt D — Watcher và Agent được nối và kiểm chứng

### BR-13 — M06 fetch/normalize/history handoff thực

Liên quan phát hiện 3.

- [ ] Chọn một nguồn public/được phép và parser cụ thể; có fixture cho response hợp lệ, lỗi và thiếu field.
- [ ] Map raw response vào Observation và input domain mà không gán mọi seller claim thành kết luận đáng tin.
- [ ] Làm endpoint/CLI handoff theo ADR BR-08; core validate và ghi history rồi trả record ID.
- [ ] Quy định identity/idempotency cho lần quan sát mới so với retry cùng lần quan sát.
- [ ] Handoff thất bại không được báo đã persist; cache watcher không thay canonical store.
- [ ] Chỉ rõ hướng dẫn cấu hình, chạy thủ công, restart/query/replay cho học viên.

Nghiệm thu: watcher output vào đúng history của BR-09, đọc lại sau restart; retry không tạo record mơ hồ; nguồn/method không được phép, malformed response và history sink lỗi có kết quả rõ.

### BR-14 — n8n M06 compatibility và static data

Liên quan phát hiện 3, 5.

- [ ] Chốt một n8n engine version thực sự thử, ghi node versions trong COMPATIBILITY.
- [ ] Sửa hướng dẫn: không yêu cầu static data persist giữa manual test nếu engine không hỗ trợ.
- [ ] Tạo quy trình smoke bằng trigger/webhook production-mode trên nguồn kiểm thử giới hạn, hoặc dùng test cache có persistence phù hợp.
- [ ] Nối node history handoff tới adapter BR-13, giữ lại acknowledgment.
- [ ] Test NEW → UNCHANGED → CHANGED, đổi thứ tự key vẫn UNCHANGED, oversized response, sink failure và retry.
- [ ] Test restart và mất watcher cache: có thể trở lại NEW nhưng history không bị reset.

Nghiệm thu: có engine version, ngày thử, import result, execution IDs và output/history refs; JSON parse hợp lệ một mình không đủ. Không đổi UNVERIFIED thành verified chỉ từ code review.

### BR-15 — M07 grounded Agent và tool enforcement

Liên quan phát hiện 4.

- [ ] Truyền evidence payload thật từ cùng store vào Agent; loại placeholder e1 khỏi đường vận hành.
- [ ] Parse ID/claim do Agent thực sự trả, không tự gắn ID context để tạo cảm giác grounded.
- [ ] Nối deterministic validators vào workflow trước khi output được dùng tiếp.
- [ ] Tool registry được thực thi trước request: method, host, timeout và redirect ngoài allowlist phải có cách xử lý rõ.
- [ ] Fetch mới phải normalize/register evidence rồi mới được viện dẫn; registry không chỉ là chuỗi JSON để log.
- [ ] Chạy prompt injection, ID bịa, thiếu context, tool/host lạ, redirect và write request qua đúng integration path.
- [ ] Dạy học viên đọc validation result và phân biệt HUMAN_REVIEW với việc đã chứng minh claim đúng.

Nghiệm thu: model trả ID bịa bị reject/abstain có reason; raw tool text không đổi quyền; normal case có evidence resolve được; có output/trace từ instance thực theo profile đã ghi.

## 9. Đợt E — Kiểm thử khả năng tự học

### BR-16 — Một bài end-to-end và pilot người mới

Liên quan phát hiện 1–8, 11.

- [ ] Tạo offline smoke xuyên CLI/API thật: import → history → action → outcome → advisor → evaluation → review → watcher → Agent.
- [ ] Dùng output bước trước làm input bước sau, không tạo bộ ID mới độc lập ở mỗi demo.
- [ ] Rerun/restart, wrong ID, missing data, zero, timeout và deliberate failure đều có kết quả dự kiến.
- [ ] Viết walkthrough rõ phần dùng fixture và phần cần học viên thao tác thực.
- [ ] Mời một người chưa biết repo làm thử theo tài liệu; ghi thời gian, câu hỏi phát sinh, bước cần trợ giúp và lỗi chưa được tài liệu giải thích.
- [ ] Sửa các điểm vướng và thử lại đoạn liên quan; ghi residual gaps.

Nghiệm thu bản beginner MVP: người thử hoàn thành đường được công bố mà không cần tác giả cung cấp lệnh/code bị thiếu; giải thích được ít nhất input, output, một failure case và giới hạn của bot. Người được hỗ trợ về thiết kế/code phải được ghi rõ, chưa tính self-service PASS.

Không dùng số lần gọi model, test PASS hoặc việc đã tạo account thay cho affiliate outcome thực tế.

## 10. Đợt F — Execution và vận hành 24/7

### BR-17 — Tích hợp M08–M11 và live adapter có giới hạn

Liên quan phát hiện 1, 10, 12. Item lớn này triển khai bằng các PR theo Mission; không gom một PR khổng lồ.

- [ ] BR-17a: M08 dùng decision/evidence hiện có để tạo intent/policy shadow; không nối executor.
- [ ] BR-17b: M09 có đường human approval được kiểm chứng nguồn gốc, persist/restart/revalidate và executor profile riêng.
- [ ] BR-17c: M10 nối grant/trusted cost/ledger/outcome; chọn một action ngoài hệ thống có rủi ro thấp được review.
- [ ] BR-17d: M11 nối promotion/lease/activation/health/STOP/reconciliation/closed-cycle vào entrypoint canonical.
- [ ] Dùng đúng canonical enforcement functions, tránh gọi helper bỏ qua trusted provenance, sticky STOP hoặc EffectRef validation.
- [ ] Adapter thật phải chứng minh idempotency/atomic reservation theo yêu cầu; adapter không đáp ứng thì chưa được activate.
- [ ] Có failure drills: hết hạn, đổi policy, kill switch sau approval, missing ledger, duplicate, unknown side effect và recovery.
- [ ] Bài thực hành ghi đúng target/action/outcome, credential setup và cách dừng; không hard-code human approval để claim live proof.

Nghiệm thu: offline integration PASS từng Mission; E5 có external side effect thật trong grant đã review và outcome link đúng; E6 có ít nhất 3 closed cycles thật cùng recovery/review evidence. Chưa có E5/E6 thì ghi phần triển khai đã xong riêng, item không được coi là hoàn tất live proof.

### BR-18 — Deployment và runbook có thể làm theo

Liên quan phát hiện 12.

- [ ] Có cấu hình triển khai tham chiếu: Compose hoặc service được chọn, pin version, biến môi trường mẫu không có secret, volumes và các thành phần thực sự cần.
- [ ] Lệnh build/start/status/logs/stop; health endpoint và expected response.
- [ ] Read-only profile của M06 có thể triển khai sau BR-16; profile ghi chỉ mở sau gate BR-17.
- [ ] Backup canonical store và restore sang môi trường trống; kiểm record count/ID/replay/budget/approval state.
- [ ] Restart host/process không reset state hoặc tự mở quyền; có quan sát disk/resource và lỗi dependency.
- [ ] Runbook cho token hết hạn, nguồn đổi format, model unavailable, history sink lỗi và STOP/unknown execution.
- [ ] Chốt window thử always-on theo cadence use case; khuyến nghị ít nhất 24 giờ cho bản tham chiếu đầu, đây là tiêu chí deployment bổ sung chứ không thay E6.
- [ ] Ghi môi trường và giới hạn tải đã thử; local smoke không được gọi là VPS production proof.

Nghiệm thu: một người khác triển khai được cấu hình theo tài liệu, đọc health, dừng bot và restore dữ liệu; báo cáo window vận hành có restart drill và không mất canonical state.

### BR-19 — Công bố readiness theo từng lớp

Liên quan các tuyên bố ready hiện tại và toàn bộ review.

- [ ] Tách trạng thái nội dung, runtime integration, live compatibility và beginner pilot; không để một chữ ready bao trùm mọi mức bằng chứng.
- [ ] Cập nhật README/curriculum index/mission manifest/validator đồng bộ nếu đổi mô hình trạng thái; không làm sai semantics PASS của học viên.
- [ ] Có bảng supported profile cho OS, Go, n8n, provider và deployment version đã thử.
- [ ] Chạy CI integration offline ở mỗi PR liên quan; live smoke theo profile và khi nâng version, không ép public CI có secret của học viên.
- [ ] Lưu release evidence và known limitations; các item mở vẫn hiển thị trong kế hoạch.
- [ ] Tách release “beginner MVP đến M07” với release “production path M08–M11” nếu phần sau chưa đủ bằng chứng.
- [ ] Rà lại required check names và bằng chứng nhánh bảo vệ khi maintainer quyết định áp dụng.

Nghiệm thu: mọi tuyên bố readiness có artifact/test/pilot tương ứng, không còn dùng demo/sandbox thay live proof; PROGRESS của học viên không thay đổi nếu chưa có việc học mới.

## 11. Mapping review → đầu việc

Số phát hiện theo [review gốc](../../REVIEW-2026-09-05.md).

| Phát hiện | Đầu việc xử lý |
|---|---|
| 1. Thiếu integration từ M03 | BR-07, BR-08, BR-10, BR-11, BR-12, BR-16, BR-17 |
| 2. Thiếu use case affiliate end-to-end | BR-04, BR-06, BR-10, BR-12, BR-16 |
| 3. M06 handoff chưa persist | BR-13, BR-14 |
| 4. M07 chưa validate grounding | BR-11, BR-15 |
| 5. Static-data manual test sai | BR-14 |
| 6. Measurement window M03 | BR-02, BR-10 |
| 7. Advisor/schema validation | BR-03, BR-11 |
| 8. Mapping M00/M02/DecisionPacket | BR-03, BR-08, BR-09 |
| 9. CI marker fail | BR-01 |
| 10. Checkpoint field cũ | BR-01, BR-17 |
| 11. Onboarding chưa đủ | BR-05, BR-07, BR-16 |
| 12. Chưa có deployment chạy được | BR-17, BR-18 |
| Ghi chú tên required checks | BR-01, BR-19 |

## 12. Tiêu chí đóng từng đợt

| Đợt | Điều kiện đóng |
|---|---|
| A | Các lỗi tái hiện có regression test; 8 validator PASS; checkpoint/schema khớp |
| B | Quickstart smoke trên profile hỗ trợ; use case và định dạng nguồn đã chốt; bài Go/JSON có bài làm kiểm được |
| C | M00–M05 nối cùng state/IDs; CLI nhận input người học; mẫu nghiệp vụ có thể đi hết vòng |
| D | n8n version được thử, watcher persist, Agent validate; có integration evidence thật |
| E | Offline E2E và pilot người mới có kết quả; phạm vi beginner readiness được công bố trung thực |
| F | Adapter live + deployment/restore/recovery chứng minh được; E5/E6 giữ đúng gate hiện hành |

Khi một blocker ngoài repo chưa giải quyết, đóng phần code/docs có bằng chứng và ghi phần live/pilot còn mở; không đánh dấu toàn đợt DONE.

## 13. Kiểm chứng khi triển khai

Từ repo root, chạy từng lệnh và kiểm exit code riêng:

    python3 scripts/validate_missions.py
    python3 scripts/validate_repo.py
    python3 scripts/validate_artifact_spine.py
    python3 scripts/validate_continuity.py
    python3 scripts/validate_language_policy.py
    python3 scripts/validate_agent_semantics.py
    python3 scripts/validate_semantic_contracts.py
    python3 scripts/validate_m11.py

Tại mỗi module Go hiện tại:

    go test ./...
    go vet ./...

Sau BR-08, cập nhật danh sách module/lệnh theo cấu trúc thực tế. Sau BR-16, thêm đúng lệnh smoke đã triển khai; không đưa lệnh dự kiến vào tài liệu như thể đã chạy được.

- Thay runtime: regression và eval liên quan, schema check output thực.
- Thay schema: compatibility/replay và producer/consumer liên quan.
- Thay n8n: import + execution smoke trên engine đã pin; static checks không thay engine execution.
- Thay tài liệu thuần túy: link/path, nhất quán trường/tên lệnh và validator phù hợp; không cần test không liên quan.
- Mỗi lỗi mới phải có expected/observed và bằng chứng sửa được; chỉ bổ sung test khi nó bảo vệ behavior có ý nghĩa.

## 14. Mẫu cập nhật một đầu việc

Khi nhận việc hoặc mở PR, điền:

    ID:
    Trạng thái:
    Người thực hiện:
    Reviewer:
    Issue/PR:
    Commit baseline:
    Phụ thuộc đã đạt:
    File/API thay đổi:
    Tiêu chí nghiệm thu đã đạt:
    Lệnh kiểm tra và exit code:
    Evidence:
    Known limitations:
    Blocker và người cần quyết định:
    Rollback hoặc cách trở lại baseline:
    Ngày cập nhật:

Mỗi PR nên gắn một item hoặc một nhóm thay đổi phụ thuộc chặt. BR-17 dùng subtask a–d và trạng thái riêng trong PR; item cha chỉ DONE khi các phần bắt buộc đã đạt.

## 15. Ba PR thực thi đầu tiên đề xuất

1. **BR-01:** sửa CI marker + checkpoint stale + tên check trong governance; không thay domain behavior.
2. **BR-02:** quy tắc cửa sổ đo + regression tests + cập nhật M03 lesson/eval.
3. **BR-03:** required-field/schema validation cho output, bắt đầu Advisor reason và quyết định adapter M02.

BR-04 có thể được chốt song song với ba PR này. Sau đó mới chốt cấu trúc shared core và use case platform; tránh viết adapter lớn khi chưa biết input/output mục tiêu.

## 16. Nhật ký kế hoạch

| Ngày | Thay đổi | Bằng chứng |
|---|---|---|
| 05/09/2026 | Lập kế hoạch 6 đợt, 19 đầu việc, mapping toàn bộ review | Review tại commit 7d2a3ab; các item triển khai đều TODO |
| 05/09/2026 | Triển khai BR-01, chuyển IN_REVIEW; 18 item còn lại TODO | [Bằng chứng BR-01](#br-01--bằng-chứng-triển-khai); chờ human review |
| 05/09/2026 | Ghi nhận BR-01 DONE theo yêu cầu merge của chủ repo; triển khai BR-02 IN_REVIEW; 17 item còn lại TODO | PR #20 đã merge; [bằng chứng BR-02](#br-02--bằng-chứng-triển-khai) |
| 05/09/2026 | BR-02 DONE theo yêu cầu merge; BR-03 IN_PROGRESS, phần a chờ review, b/c TODO | PR #21 merge `1e94ec5`; audit artifact và regression M04 |

Việc commit kế hoạch không có nghĩa các lỗi đã sửa hoặc Mission của học viên đã PASS.
