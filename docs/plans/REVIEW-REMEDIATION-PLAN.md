# Kế hoạch sửa sau review toàn repo tại ece6a32

<!-- readiness-as-of: 2026-09-11 -->
<!-- readiness-main-baseline: c11d6eea32bf82ddbfdd9751373543899fbfb4ae -->

> Reconcile 10/09/2026: đây là tracker hiện tại của `main` tại baseline trên.
> Xem [kế hoạch pre-merge tại 737e85a](PRE-MERGE-REMEDIATION-737E85A.md) cho
> phát hiện PMR-01…07 ban đầu. Các ghi chú cũ chỉ có giá trị lịch sử; matrix và
> bảng gói dưới đây là nguồn trạng thái hiện hành.

- Mã: RR-2026-09-07; phiên bản kế hoạch: 2.
- Ngày lập kế hoạch: 08/09/2026; mã kế hoạch theo ngày review baseline.
- Baseline: `ece6a32619e5b9a05d0599b87f50023f38931cb9`.
- Trạng thái: **CURRENT_MAIN_TRACKER** — có implementation/test offline đã
  merge; không criterion nào được coi production-ready hoặc pilot-complete.
- Cơ sở: [sổ phát hiện và bằng chứng baseline](evidence/REVIEW-ECE6A32.md).
- Liên kết kế hoạch gốc: [BR-2026-09](BEGINNER-READINESS-PLAN.md); trạng thái tích hợp BR: [readiness matrix](READINESS-MATRIX.json).
- Người lập kế hoạch: Codex theo yêu cầu chủ repo. Người triển khai/reviewer từng đợt: chưa phân công; điền khi nhận việc. Không mặc định chủ repo đã nghiệm thu kế hoạch hoặc cho phép live operation.

Cập nhật sau review PR #94: bổ sung dependency RP-05 cho RP-06, tách nghiệm thu
restore M00–M10 khỏi phần mở rộng M11 bắt buộc ở RP-07a, và giao rõ việc tạo/lưu/
kiểm chuỗi cost-bound, gate, authorization, execution và EffectRef. Các phạm vi
còn lại được ghi cụ thể `PARTIAL`/`OPEN`, không suy từ test PASS sang nghiệm thu
toàn bộ R01–R16.

## 1. Mục tiêu và giới hạn

Sửa những đường chạy đã tái hiện sai; thống nhất canonical implementation giữa learner Bot, harness và n8n; chứng minh một chuỗi artifact M00–M11 bằng runtime thật trong môi trường offline. Sau đó mới tổ chức pilot và kiểm chứng n8n/host/provider được chọn.

Kế hoạch gốc không tự cấp quyền sửa Go/Python/blueprint/schema, không đổi
required checks/branch protection, không chạy live executor, không tạo tài khoản,
không mua host hoặc dùng API trả phí. Các implementation đã merge vẫn giữ các
giới hạn đó; công việc live/operated cần authority riêng.

Không đổi `PROGRESS.md` hoặc công nhận Mission PASS cho học viên. `execution_permitted=false`, evidence synthetic và trạng thái `NOT_READY_FOR_PRODUCTION` vẫn được giữ. Hoàn tất offline không chứng minh business outcome, human pilot hoặc production readiness.

Các từ trạng thái trong tài liệu:

- `PARTIAL`: có implementation/test được nêu rõ nhưng còn acceptance hoặc evidence thiếu.
- `OPEN`: chưa bắt đầu, hoặc đòi authority/evidence ngoài repo.
- R01–R16 chỉ đóng khi matrix/evidence graph ghi phạm vi, diff/regression và reviewer xác nhận baseline/head.
- Trạng thái BR tổng hợp thuộc matrix; bảng BR trong kế hoạch gốc là bản trình bày của cùng trạng thái, không phải nguồn độc lập. RP là gói công việc, không phải mức readiness mới.

## 2. Quyết định thiết kế để triển khai

1. **Một implementation cho mỗi quy tắc.** Tách logic pure và decoder cần thiết từ `lab/mission-runtime/cmd/demo` vào `core` theo từng module; CLI/harness gọi cùng API. Store adapter giữ I/O và giao dịch. Không coi harness là oracle tuyệt đối: đối chiếu contract và test expected outcomes độc lập.
2. **Một canonical builder/resolver.** M06 gửi input observation/provenance, không tự tính hash/result bằng JavaScript. Consumer M07/M08 dùng chung bước resolve đúng một record, kiểm integrity và replay, xuất cả aggregate ID và field ID thật.
3. **Grounding có phạm vi có thể kiểm.** Ưu tiên claim có cấu trúc và diễn giải được render từ claim đã xác minh. Free-text của model chỉ là nội dung chưa xác minh, không nhận nhãn grounded. Không dùng một LLM judge mới làm điều kiện nghiệm thu duy nhất.
4. **Tool trace do adapter sở hữu.** Registry được kiểm tại chỗ thực hiện mỗi request; response được phân loại, đăng ký và nhận ACK trước khi trở thành evidence có thể cite. Model không tự khai trace hoặc tự cấp ID.
5. **Ledger bền vững, grant bất biến.** Intent đang bind không sở hữu tổng usage. Grant version/hash và approval gắn chặt; reservation có ID/idempotency, không reset qua bind/restart/restore. Không nâng giới hạn chỉ bằng sửa JSON cùng grant ID.
6. **An toàn ghi dữ liệu.** Kiểm path/file identity; khóa toàn giao dịch read/check/write; ghi atomic, sync theo nền tảng; ACK chỉ sau commit. Thiếu ledger, state không nhất quán hoặc STOP không đọc được phải fail closed, không tự tạo lại.
7. **Clock rõ ràng.** Domain nhận clock để test; adapter vận hành dùng thời gian thực đáng tin trong process. Fixture clock không được dùng như override mở quyền trên đường vận hành.
8. **Restore kiểm cả nội dung và quan hệ.** Manifest inventory đầy đủ; snapshot nhất quán; validate trong staging, chạy canonical loaders/replay/cross-store links trước khi công bố đích khôi phục sẵn dùng. Checksum không thay semantic validation.

Các tên package/endpoint/test dưới đây xuất phát từ kế hoạch ban đầu; implementation
refs hiện hữu được liệt kê trong matrix. Phần chưa có ref vẫn là đề xuất và phải
được version/test trong PR riêng.

## 3. Bao phủ 16 phát hiện

Các hàng dưới đây giữ acceptance target. Trạng thái hiện tại nằm ở matrix và
bảng gói; “Kiểm chứng bắt buộc” không tự chứng minh test đã PASS.

| Review / ưu tiên | Lỗi cần đóng | Gói chủ trì | Kiểm chứng bắt buộc |
|---|---|---|---|
| R01 / P1 | Diễn giải bịa đi kèm field/value đúng | RP-05 | Câu bảo đảm lợi nhuận không được xuất thành grounded answer; field claim đúng vẫn dùng được |
| R02 / P1 | M06 hash/result lệch core | RP-04 | ASCII, A&B, Unicode, thiếu giá/commission đều có hành vi canonical đúng; record được ACK phải replay MATCH |
| R03 / P1 | ID n8n/CLI/M08 không thống nhất | RP-04 | ID field thật đi qua cả ba đường; ID `#price` chưa đăng ký bị từ chối ở mọi đường |
| R04 / P1 | HTTP/M08 bỏ replay và unique-record gate | RP-04 | DRIFT, UNREPLAYABLE, duplicate record ID không thành context/intent hợp lệ |
| R05 / P1 | Tool result không được đăng ký, call hợp lệ bị chặn | RP-05 | Registry → request → evidence ACK → cite; trace thực hợp lệ qua được, trace giả bị chặn |
| R06 / P1 | M08 khác schema/policy harness | RP-02 | Thiếu proposal, null parameters, thời gian tương lai/không hợp lệ có cùng kết quả trên CLI/harness |
| R07 / P2 | JSON number bị làm tròn trước hash | RP-02 | `9007199254740993` giữ nguyên qua decode/hash/store/restart hoặc reject tường minh trước ghi |
| R08 / P1 | Nới/reset budget qua reimport/bind | RP-03 | Grant đã cạn không dùng lại qua tăng cap, đổi currency, bind qua lại hoặc replay approval |
| R09 / P1 | Approval hết hạn vẫn reserve | RP-03 | Before/at/after expiry, process mới và restore; hết hạn chặn operation |
| R10 / P1 | Concurrent reserve vượt cap | RP-03 | Barrier đồng bộ 24 process, cap=1: đúng một reservation, ledger khớp ACK; crash injection không làm mở budget |
| R11 / P1 | Output ghi đè history đầu vào | RP-01 | Same path, symlink, hardlink, output tồn tại: reject trước ghi; input bytes không đổi |
| R12 / P1 | Backup bỏ nested artifact | RP-06 + RP-07a | Restore đủ inventory M00–M10 và phần mở rộng M11; layout chưa hỗ trợ bị từ chối, không bỏ artifact |
| R13 / P1 | RESTORED dù graph bị hỏng | RP-06 + RP-07a | Orphan action/outcome/proposal, cost-bound/gate/authorization/execution và artifact M11 bị từ chối; chỉ đóng toàn phạm vi sau gate restore RP-07a |
| R14 / P1 | Chain M00–M11 thiếu chặng/artifact | RP-07 | Không bỏ chặng; proposal M07 → M08 → cost-bound/gate/authorization/execution → outcome có EffectRef thật; lifecycle M11 và restore resolve được |
| R15 / P2 | CI chưa chạy smoke/blueprint implementation | RP-08 | Required workflow gọi smoke và code blueprint thật; mutation implementation làm đúng job fail |
| R16 / P2 | Readiness mâu thuẫn và thiếu code gaps | RP-09 | Matrix/plan/evidence thiếu hoặc mâu thuẫn làm audit fail; không suy “file tồn tại” thành nghiệm thu |

## 4. Chia PR và dependency

RP là package logic, không phải số PR GitHub. Bảng dưới phản ánh `main` tại
baseline; mỗi thay đổi mới phải cập nhật matrix/evidence graph, baseline/head
và regression tương ứng.

| Gói | Phạm vi | Phụ thuộc trước khi merge | Quy mô | Trạng thái |
|---|---|---|---|---|
| RP-00 | Lưu kế hoạch/baseline, hạ tuyên bố quá mức | — | S | MERGED (`main` `12a088a`) |
| RP-01 | Bảo vệ đường dẫn và file đầu vào | RP-00 | S | PARTIAL — M08/M10/M11 output paths reject alias/overwrite and preserve canonical state before portable output; inventory of every writer remains open |
| RP-02 | Shared M08 decoder/policy, exact-number/hash contract | RP-01 | M | PARTIAL — learner and harness share M08 decoding/policy with exact-number regression; broader M09 conformance/migration remains open |
| RP-03 | Shared M09/M10 guard, cost-bound/gate/authorization/execution, ledger và STOP | RP-02 | L; chia 03a/03b | PARTIAL — offline approval/grant/cost/gate/authorization/reservation and no-side-effect records exist; expiry/rebind, complete EC coverage and multi-file crash proof remain open |
| RP-04 | Canonical M06 builder và resolver M07/M08/HTTP | RP-01; tích hợp M08 sau RP-02 | M | PARTIAL — shared fixture builder/resolver and n8n engine regression exist; governed selected-source profile and operated run remain open |
| RP-05 | M07 grounded output và tool-result lifecycle | RP-04 | L; chia 05a/05b | PARTIAL — adapter-owned trace/proposal persistence and n8n stub path exist; selected-source/provider operated evidence remains open |
| RP-06 | Snapshot/restore và graph M00–M10, gồm proposal M07 và execution chain | RP-03, RP-04, RP-05 | M | PARTIAL — v3 typed inventory, graph validation and cross-process gate exist; M11 fixture outcome/ledger links are now checked both ways, while broader semantic orphan and crash/host proof remain open |
| RP-07 | M11 lifecycle + mở rộng restore (07a), rồi full chain/walkthrough (07b) | 07a sau RP-02…RP-06; 07b sau gate lifecycle/restore của 07a | L; chia 07a/07b | PARTIAL — learner lifecycle, UNKNOWN→STOP/reconciliation, admission, shared M00–M11 smoke and restore exist; each M11 gate follows activation and retains its exact budget snapshot, each authorization has a prior ALLOW gate/health snapshot and must reserve against that unchanged ledger, each attempt resolves its prior normal reservation ledger and historical authorization lifetime, each offline evaluation cites exactly its fixture outcome, each M11 chain closes with exact lease/correlation lineage, and an UNKNOWN attempt may have only one post-attempt human `NOT_PERFORMED` reconciliation resolution, while expiry/rebind and multi-file crash seams remain open |
| RP-08 | CI parity/mutation/cross-process coverage | Bắt đầu cùng RP-01; đóng sau RP-07 | M, xuyên các PR | PARTIAL — required offline smokes and disposable M06/M07 n8n engine regressions run in CI; mutation breadth and operated parity remain open |
| RP-09 | Readiness audit có dữ liệu/evidence, chốt offline acceptance | RP-06, RP-07, RP-08 | M | PARTIAL — matrix/graph/plan/CI audit is structured; remote CI and external evidence remain outside local audit |
| RP-10 | n8n operated run, pilot máy sạch, deployment drill | RP-09 và lựa chọn môi trường/quyền cần thiết | M/L | OPEN — requires selected environment, authority and independently recorded operated evidence |

Luồng ưu tiên: RP-01 → RP-02 → RP-03; RP-04 có thể làm song song trên file độc lập. RP-06 chỉ merge sau RP-03/RP-04/RP-05 để kiểm proposal đã persist và execution chain thật. RP-06 nghiệm thu inventory M00–M10; RP-07a bổ sung artifact M11 và phải mở rộng manifest/loader/restore tests trong cùng gói, rồi RP-07b mới nghiệm thu toàn chuỗi. Không thêm dependency RP-07 ngược vào RP-06 gây vòng lặp. RP-08 đưa test vào từng PR, không đợi cuối dự án mới bật gate. Không đặt ngày production trước khi chốt điều kiện RP-10.

### Runtime gap được chọn tiếp theo — RP-03 authority expiry/rebind

Phạm vi tiếp theo là M09/M10 authority lifetime sau **process mới và
backup/restore**. Đây là gap runtime lớn nhưng có nghiệm thu offline rõ ràng:
không cần executor/provider, và không được biến fixture timestamp thành quyền
vận hành.

- Dựng authority hợp lệ qua learner Bot: intent/policy → approval → canary
  grant → trusted cost-bound → gate/authorization/reservation.
- Kiểm ranh giới `before` / `at` / `after` expiry cho intent, approval, grant
  và cost-bound bằng clock seam do runtime kiểm soát; đường CLI vận hành không
  được nhận caller timestamp để lùi thời gian cấp quyền.
- Lặp lại mỗi boundary trong process mới và runtime đã `backup restore`.
  Attempt hết hạn phải trả trạng thái đóng, không tăng usage, không tạo
  reservation/authorization/execution artifact và không ghi portable output.
- Thử rebind/reimport artifact cùng ID nhưng expiry/cap/currency đã thay đổi:
  phải reject; chỉ authority mới có identity/hash và human review đúng luồng
  mới được xét tiếp.
- Regression phải snapshot bytes/registry/state trước và sau reject. Không gọi
  executor, không tạo business outcome và không suy thành power-loss proof.

**Cập nhật implementation RP-03 — IN PROGRESS (chưa đổi `PARTIAL`):** learner
Bot hiện dùng clock do runtime sở hữu (seam chỉ nằm trong Go test, không có cờ
CLI hay environment override). M08 → M10 regression tạo intent/policy/approval/
grant/cost/gate thật, rồi chứng minh `m10-authorize` reject tại và sau expiry
của từng authority mà không tạo portable output hay thay đổi
`mission-state`/M10 registry. Cùng regression chặn rebind ID khi đổi expiry,
cap hoặc currency; shared smoke M00–M11 dùng cùng runtime đã kiểm rebind và
no-mutation, còn backup/restore bao phủ đủ bốn authority bằng binary Bot ở
process mới. Combined subprocess smoke dùng test-only clock seam (không có ở
binary Bot vận hành), backup/restore từng authority rồi kiểm `before` / `at` /
`after` expiry; reject không ghi portable output hay state. Coverage
execution/ledger/cost state rộng hơn vẫn thiếu; không suy slice này thành
complete acceptance.

**Nghiệm thu chọn scope:** một shared smoke dùng runtime thật, subprocess và
restore; mỗi ca trả status xác định và kiểm no-mutation. Sau đó matrix vẫn
`PARTIAL` cho tới khi EC-01…EC-05 breadth và multi-file crash seams được xử lý.

### Snapshot lịch sử trước merge — 2026-09-08

Các ghi chú dưới đây mô tả branch `codex/rp-01-path-safety` tại `df16783`
trước khi các foundation được merge qua PR #95 và các PR sau đó. Chúng chỉ giữ
ngữ cảnh quyết định; **bảng trạng thái ở trên và metadata baseline mới là trạng
thái hiện tại của `main`**.

- **RP-01, phần M08:** commit `b48cea7` chặn cùng path, symlink/hardlink và
  overwrite artifact khác; retry byte-identical trả `EXACT_DUPLICATE`. Test
  trực tiếp learner command kiểm history/request/policy không bị mutate. Chưa
  audit hết các writer khác (backup/restore) nên không đóng RP-01.
- **RP-02, phần M08:** commit `573d4ed` thêm `core/m08`; learner và harness
  dùng chung hash/policy/decoder cho ActionIntent/PolicyDecision. Có regression
  số `9007199254740993`, `parameters:null`, duplicate key và agent proposal
  chưa resolve. Proposal resolver/store thực vẫn thuộc RP-04/RP-05; chưa có
  full conformance matrix nên không đóng RP-02.
- **RP-03 foundation:** commit `1eb452c` thêm state atomic, lock fail-closed và
  expiry/binding checks; `b68144c` thêm reservation ID/ledger retry;
  `12d54cc` thêm race smoke; `149128e` thêm `core/m10` TrustedCostBound;
  `df16783` thêm registry/resolution learner. Smoke `scripts/smoke_br16a_offline.py`
  chạy register → reserve và các reject unregistered/tampered/expired, restart,
  STOP và race cap. Grant/approval/ledger hiện chưa là graph canonical M09/M10,
  chưa có execution stub/outcome linkage hay fault-injection đầy đủ; RP-03 vẫn
  mở.
- **RP-03 canonical grant/gate foundation:** `core/m10` nay decode/hash
  `CanaryGrant` theo contract, recheck approval/policy/risk/host/correlation/
  expiry khi learner nhận grant, và recheck hash khi reload mission state.
  `m10-gate` resolve cost-bound đã register rồi emit immutable,
  non-authorizing `CanaryGateDecision` từ snapshot budget với
  `EVALUATED_AT` explicit để retry byte-identical. Smoke shared chain cover
  grant sealed, gate ALLOW/retry và cost-bound tamper reject. Grant còn embedded
  trong mutable state, gate chưa được registry/authorization/execution resolve
  và không có side effect; RP-03 vẫn mở.
- **RP-03 authorization foundation:** `m10-authorize` load gate artifact,
  cost-bound registered và state hiện tại, buộc `AUTHORIZED_AT` khớp gate rồi
  evaluate lại snapshot trước khi emit immutable `ExecutionAuthorization`.
  Authorization bind grant/gate/cost/intent/executor và hạn dùng là minimum của
  intent/grant/cost expiry; không reserve budget hay gọi executor. Smoke cover
  authorization retry và reject gate stale sau reserve. Artifact chưa được
  đăng ký trong graph state và execution record/effect link còn mở.
- **RP-03 cancelled-execution foundation:** learner `m10-cancel` resolve và
  kiểm `ExecutionAuthorization` với mission state hiện tại rồi ghi artifact
  bất biến `ExecutionRecord` chỉ có `CANCELLED`/`NOT_PERFORMED`. Core validator
  và BR-16a smoke kiểm retry byte-identical; đường này không gọi executor,
  không reserve budget hay tạo EffectRef. Execution thành công/thất bại, ledger
  link, outcome và graph persistence vẫn thuộc RP-03/RP-06/RP-07, do đó RP-03
  vẫn mở.
- **RP-03 authorization-reservation foundation:** `m10-reserve-authorization`
  resolve authorization/cost bound đã registry cấp, kiểm expiry/binding rồi
  charge đúng bound và persist reservation một-lần với `authorization_id`.
  `m10-cancel` giờ cần reservation đó và ACK chỉ sau khi ghi `execution_id`
  ngược vào state; retry chỉ cho cùng ID/artifact. `m10-reserve` cũ được giữ
  compatibility-only, không thể tạo execution binding. Chưa có transaction
  multi-artifact crash recovery, execution result/outcome/EffectRef hay ledger
  canonical ngoài learner state, nên RP-03 vẫn mở.
- **RP-03 no-side-effect execution/outcome foundation:** `m10-record-failed`
  tạo đúng `ExecutionRecord` `FAILED`/`NOT_PERFORMED` cho reservation governed
  và không có executor/network. `m10-outcome` chỉ append fixture outcome
  `CANCELLED`, metrics rỗng, `fixture:m10-outcome/…`, sau khi `EffectRef`
  `MACHINE_EXECUTION` resolve record đã registry cấp và state xác nhận bind
  reservation. Smoke cover record retry và forged EffectRef reject. Đây không
  phải execution success, business outcome, cross-store canonical outcome graph
hay proof side effect; RP-03/RP-06/RP-07 vẫn mở.
- **RP-03 canonical binding before portable output (2026-09-10):**
  `m10-record-failed` và `m10-cancel` bind `ExecutionID` vào reservation state
  sau canonical registry append nhưng trước khi ghi output path do caller chọn.
  BR-16a cố tình tạo output conflict, kiểm response `CONFLICT` vẫn có record
  canonical, state đã bind exact execution, rồi retry sang path mới thành công.
  Không coi portable output là commit boundary, không chứng minh transaction
  registry/state khi crash và không cấp execution authority.
- **RP-03 M10 registry foundation:** state directory nay có registry append-only
  `m10-artifacts.jsonl`; core canonicalize/hash envelope và learner chỉ ACK
  `CanaryGrant`, trusted cost-bound, gate, authorization hoặc cancellation
  record sau khi registry validate toàn bộ link grant → cost/gate → authorization
  → record. `m10-resolve` chỉ đọc artifact canonical theo kind/ID/(tùy chọn)
  content hash. Gate/authorization hợp schema nhưng không có entry chính xác bị
  smoke BR-16a từ chối. Registry chưa có reservation-to-attempt link, execution
  result/outcome/EffectRef, migration/restore graph hoặc fault injection đầy
  đủ, vì vậy không đóng RP-03/RP-06/RP-07.
- **RP-04 foundation:** thay đổi sau mốc head ở trên thêm một learner resolver
  read-only dùng chung cho M07 context và M08 intent: record phải resolve đúng
  một lần từ history và replay `MATCH`. M06 watcher handoff và HTTP GET cũng
  resolve lại record từ store sau append trước khi ACK/return artifact. Chưa
  có core M06 builder chung, chưa đổi n8n blueprint/HTTP adapter, và chưa có
  proposal resolver; RP-04 vẫn mở.
- **RP-04 M06 shared fixture path:** `core/m06` nay owns strict
  `br13-offer-fixture/v1` decoding, timestamp/identity/provenance normalization
  và M00 packet construction. Learner CLI local/pinned fetch cùng n8n endpoint
  `/v1/m06/fixture-import` dùng profile này; endpoint append rồi resolve/replay
  canonical record trước ACK. Blueprint không còn GET/parse/hash/build history
  bằng JavaScript. Unit test cover adapter APPENDED/EXACT_DUPLICATE/replay và
  reject không mutate history. Đây chỉ là fixed synthetic profile; generic
  source profile, n8n operated execution và full shared HistoryRecord type vẫn
  còn mở, nên RP-04 chưa đóng.
- **RP-04 canonical fingerprint + M08 field links:** M06 dùng canonical JSON
  fingerprint cho body có cấu trúc (key order không tạo observation mới), còn
  raw byte hash giữ riêng cho HTTPS fixture pinning. CLI/HTTP adapter retry
  cùng body reordered là `EXACT_DUPLICATE`; price/commission thiếu tạo field
  `missing` và HistoryRecord vẫn replay `MATCH`. M08 resolve cùng canonical
  record rồi dùng đúng field IDs mà M07 context công bố; ID tự dựng và record
  `DRIFT` đều bị reject. Regression chạy implementation learner thật, không
  dùng parser Python thay thế. Chưa có generic source profile hoặc n8n engine
  execution evidence, nên RP-04 vẫn `IN PROGRESS`.
- **RP-05 foundation:** core/learner M07 nay kiểm output thực: chỉ
  `HUMAN_REVIEW`/`ABSTAIN`, claim/evidence/value và `answer`/`claim.text` phải
  là render deterministic; prose tự do, ID dư/giả, quyền ghi và `tool_calls`
  tự khai bị reject. `m07 register-tool-result` preflight/validate response,
  ghi artifact immutable có trace hash rồi `m07 validate` resolve lại hash và
  record binding trước khi body `unknown` được cite. Test learner trực tiếp,
  regression M07 và smoke shared workspace tạo/cite trace rồi reject trace giả.
  `m07 register-proposal` persist raw validated output/proposed action với
  canonical digest và record binding; M08 agent intent/policy resolve lại
  artifact và reject target/parameters đổi ngoài proposal. Chưa có canonical
  tool-evidence store/transport seam và blueprint n8n chưa gọi adapter; RP-05
  vẫn mở, R01/R05 chưa đóng toàn phạm vi.
- **RP-05 HTTP adapter foundation:** watcher có endpoint loopback
  `/v1/m07/register-tool-result`, `/v1/m07/validate` và
  `/v1/m07/register-proposal`. Chúng resolve history canonical, lưu artifact
  trace/proposal bất biến dưới store dẫn xuất từ history và ACK sau validation;
  validate chỉ resolve tool trace bằng ID đã persist. Test handler cover ACK,
  validate, proposal persistence và trace ID giả. Blueprint n8n đã chuyển flow
  sang preflight/full-response/no-redirect/register/context/validate/proposal
  endpoints và static validator kiểm wiring, nhưng chưa có n8n instance chạy
  workflow hay parity execution thật nên chưa là operated evidence.
- **RP-05 adapter-owned transport:** watcher thêm `fetch-and-register`; n8n
  không còn gọi remote HTTP trực tiếp. Adapter validate registry trước fetch,
  resolve DNS trước request và mỗi dial, reject non-public/mixed IP, proxy,
  redirect và response vượt 256 KiB; timeout đến từ registry. Unit test cover
  private/CGNAT/mixed DNS và body quá cỡ; chưa có n8n import/run, controlled
  public-source integration hoặc sink-failure parity nên không đóng RP-05.
- **RP-05 blueprint persistence boundary:** blueprint M07 dùng adapter URL và
  registry review cố định, không lấy hai policy boundary này từ event input.
  Trace tool, context canonical, grounding `HUMAN_REVIEW` có draft và proposal
  persistence đều phải nhận ACK ở node riêng trước khi node sau chạy. Learner
  HTTP regression kiểm proposal ACK và xác nhận output không grounded không
  thể sửa artifact proposal đã persist. Đây vẫn là fixture/offline evidence:
  chưa có n8n engine, credential model hay provider integration được vận hành.
- **BR-16a continuity:** smoke shared workspace nay dùng `M07
  register-proposal → M08 agent intent → M08 policy`, có ca target bị thay đổi
  bị reject. Số `9007199254740993` đi qua proposal/intent/bind/state; `bind`
  và state loader dùng shared decoder/`UseNumber` để không làm tròn trước khi
  kiểm intent hash. Chuỗi vẫn chưa có execution/EffectRef/M11 graph đầy đủ,
  nên không đổi trạng thái RP-02/RP-07 hay BR-16a.

### RP-01 — Không làm mất input/store khi ghi artifact

**Chạm tới:** `lab/affiliate-bot/cmd/bot/mission_command.go`; tái sử dụng seam kiểm đường dẫn từ các action/improvement store. Rà các output của intent/policy/backup để không tái tạo lỗi cùng loại.

- Chốt semantics create/retry/conflict thay cho status APPENDED nhưng overwrite tùy ý.
- So sánh đường dẫn chuẩn hóa và file identity; kiểm parent, symlink/hardlink theo threat model hiện có; từ chối output trùng bất kỳ input/store bảo vệ nào trước khi tạo/truncate file.
- Tách helper ghi artifact khỏi cập nhật mutable state; không sửa quyền hoặc nội dung file người dùng khi validation thất bại.
- Test cùng path HISTORY=OUT, REQUEST=OUT, INTENT=policy OUT; alias, existing output, permission failure. So digest input trước/sau mọi reject.

**Nghiệm thu:** R11 reject trước ghi; happy path/retry có hợp đồng rõ; quickstart và M00–M05 không regression. **Rollback:** revert code chưa migration; không xóa/ghi đè store để làm test xanh.

### RP-02 — Tái sử dụng M08 và bảo toàn dữ liệu được hash

**Chạm tới:** `core` (package M08 dự kiến), `contracts/validate.go`, learner mission, harness `m08*.go` và các fixture liên quan. Không nới schema để hợp thức hóa decoder thiếu field.

- Tách decoder, intent hash/seal và pure policy từ harness; hai entrypoint gọi cùng implementation. Kiểm original bytes trước typed decode: duplicate/unknown/null/case-variant keys theo canonical contract.
- Giữ JSON numbers chính xác; định nghĩa canonicalization của evidence-ID ordering, action normalization và parameters. Hash khác version cũ phải được nhận biết, không tự reseal intent đã duyệt.
- Agent intent phải có proposal_ref resolve được; RP-04 cung cấp resolver chung, RP-05 cung cấp persisted proposal. Trước đó không tự tạo proposal giả để vượt gate.
- Policy kiểm created_at/expires_at, schema, authority, risk, proposal/evidence/decision links và idempotency context. Thiếu dependency phải trả trạng thái đóng, không ALLOW.
- Conformance tests cùng payload/expected result cho CLI và harness: thiếu proposal, proposal không tồn tại, parameters null, future intent, expired, risk không biết, authority tamper, duplicate keys, exact large number qua restart.

**Nghiệm thu:** R06/R07 đóng; output vẫn đúng schema và proposal-only. **Migration:** báo version/hash không hỗ trợ; migration có lệnh riêng, backup và human review; không rewrite approval cũ để khớp hash mới.

### RP-03a/03b — Guard approval, ledger và STOP có tính bền vững

**Chạm tới:** shared M09/M10 logic dự kiến, learner mission/store adapters, canonical approval/grant/ledger contracts, harness persistence; phần STOP dùng lại khi làm M11.

**03a: đúng semantics và lifetime.**

- Tái sử dụng decoder/gate đã có, bổ sung parity tests; không duy trì LearnerCanary giản lược mà gọi đó là canonical grant.
- Bind giữ ledger/grant/consumed-approval/idempotency history tách khỏi current intent. Grant version/hash và cap/currency là bất biến; thay đổi cần grant mới được phê duyệt, không tái nhập cùng ID.
- Approval không đồng nghĩa grant budget approval. Kiểm tất cả liên kết, decision và hiệu lực intent/policy/approval/grant trước mỗi operation; grant không kéo dài quyền quá hạn dependency.
- Retry cùng reservation ID trả kết quả idempotent, không charge hai lần; retry khác ID sau hết cap bị chặn; consumed marker không bị mất qua bind/restart/restore.
- RP-03 sở hữu entrypoint/store cho M09 `ApprovalRecord → ExecutionAuthorization → ExecutionRecord` và M10 `CanaryGrant/CanaryGrantApproval + TrustedCostBound + pre-gate ledger → CanaryGateDecision → reservation/ExecutionAuthorization → ExecutionRecord → post-ledger`. Tái sử dụng [authorization schema](../../contracts/execution-authorization.schema.json), [execution schema](../../contracts/execution-record.schema.json), [cost-bound schema](../../contracts/trusted-cost-bound.schema.json) và [canary gate schema](../../contracts/canary-gate-decision.schema.json). Không dùng grant approval thay per-action approval của M09 hoặc ngược lại.
- Cost bound phải resolve tới đúng intent/hash, currency, source và thời hạn; gate/authorization/execution giữ đúng cost-bound ID/hash/amount, grant ID/version/hash, executor, correlation và idempotency. Không lấy số cost caller tự nhập làm trusted bound. Reservation phải link tới attempt/execution identity; `RESERVED` không phải bằng chứng đã thực thi.
- Chạy executor stub cô lập để sinh execution result đúng schema và side-effect state; ghi provenance fixture trong envelope/bundle, giữ `execution_permitted=false` và không có external side effect. Các authority fields bên trong artifact mô phỏng tuân theo canonical schema, không được hiểu thành quyền live.
- Loader/adapter outcome của máy phải resolve [EffectRef](../../contracts/effect-ref.schema.json) `MACHINE_EXECUTION` tới đúng `ExecutionRecord.execution_id`, rồi kiểm authorization/intent/ledger của nó. Không dùng reservation ID, approval ID hoặc human action ID thay execution ID; không tạo outcome thành công chỉ từ ACK reservation.

**03b: transaction và crash safety.**

- Khóa toàn read/validate/reserve/commit theo store; chọn transaction backend hoặc lock + atomic replace + file/directory sync. Chốt hỗ trợ OS/filesystem; nền tảng không bảo đảm primitive phải fail closed hoặc ghi rõ chưa hỗ trợ.
- Commit ledger/reservation trước ACK, kiểm STOP bên trong cùng giao dịch; không refund tự động cho attempt có side effect UNKNOWN. Không tự khởi tạo ledger bị mất.
- Fault injection tại trước/sau write, sync, rename/commit, ACK và STOP; cross-process barrier cap=1, cost boundary/overflow, đồng thời STOP/reserve, process restart, malformed/missing state, recovery chưa review.
- Inject clock cho unit tests; adapter production không nhận timestamp caller để lùi thời gian. Có ít nhất một smoke process mới sau expiry thật ngắn; không dùng hạn 2099 thay kiểm boundary.

**Ca bắt buộc cho execution chain:**

- `EC-01`: happy path M09 và M10 sinh/lưu/load lại đầy đủ artifact bằng runtime + stub; kiểm cùng graph bằng checker shared/harness, không tự viết JSON để bỏ qua bước issuance.
- `EC-02`: thiếu/sai/hết hạn cost bound, sai currency/source/intent/hash hoặc sửa amount → reject trước reservation/authorization/stub; usage không đổi.
- `EC-03`: thiếu/sai/hết hạn authorization, gate bị DENY/tamper, sai executor/grant/correlation/idempotency, execution tham chiếu authorization khác → reject; không có simulated execution thành công.
- `EC-04`: thiếu execution record, outcome trỏ reservation ID/approval ID/execution không tồn tại hoặc sai effect kind → reject; không đóng outcome/reconciliation/cycle.
- `EC-05`: execution UNKNOWN hoặc thiếu cost/result giữ reservation và reconciliation mở; retry không double-reserve, không tự refund hoặc tạo success outcome. Có cross-process restart để kiểm lại ID và usage.

**Nghiệm thu:** R08/R09/R10 có regression guard/state/ledger/ACK và STOP sau restart; EC-01…EC-05 PASS trên đường learner/harness dùng chung. Kiểm lại các bất biến này qua restore là gate bổ sung bắt buộc ở RP-06, không phải dependency ngược để bắt đầu RP-03. Test kill-process không tự được gọi là power-loss proof. **Migration/rollback:** snapshot trước migration, dừng writer, giữ STOP; không downgrade ledger format bằng binary cũ hay xóa usage.

### RP-04 — Builder và resolver canonical dùng chung

**Chạm tới:** `history.go`, `watcher.go`, `m07.go`, mission intent, `core/m06`, blueprint M06/M07 và schemas/API version cần thiết.

- Adapter nhận observation/source metadata trong input versioned, gọi canonical builder/evaluate. Blueprint không tính input_hash, recorded_result hoặc evidence IDs thay core.
- Giữ source URL/role, access method, timestamps, evidence_kind/use_context, limitation và field-level claim_kind. HTTP thành công không biến synthetic thành real hay seller claim thành fact. Reject unsupported status/body/profile trước ghi.
- Resolver trả đúng một record replay MATCH và đầy đủ original field IDs/provenance. CLI M07, HTTP context và M08 dùng resolver này. Lịch sử DRIFT có thể đọc bằng đường diagnostic riêng, không thành verified context.
- Append/ACK kiểm record ID/hash/correlation tương ứng; dedup/retry không sinh observation mới từ retry ngẫu nhiên. Chốt change-detection canonical key-order; update fingerprint sau ACK, không để sink failure được coi như persistence.
- Test missing price/commission, A&B/Unicode/escape, key reorder, malformed/oversize/error response, seller claims, synthetic HTTPS fixture, duplicate IDs, DRIFT/UNREPLAYABLE, forged field ID và sink failure. Test HTTP adapter thật bằng loopback và cùng payload qua CLI.

**Nghiệm thu:** R02/R03/R04 đóng; record hợp lệ được ACK phải replay MATCH; field ID thật đi từ context đến intent. **Compatibility:** giữ endpoint lịch sử nếu cần diagnostic nhưng tách rõ envelope/trust level; không tự sửa history đã DRIFT.

### RP-05a/05b — Grounding và vòng đời tool evidence M07

**Chạm tới:** `core/m07/m07.go`, learner M07 entrypoint, blueprint M07, proposal/output contracts, registry/request adapter và tests.

**05a: output contract.**

- Version output contract; trusted result gồm claim field/value/evidence link đã kiểm. Render diễn giải từ claim canonical, hoặc tách free-text vào phần UNVERIFIED không bàn giao như grounded advice.
- Không sửa output model bằng cách gắn IDs từ context; unknown/duplicate IDs, thiếu claim, authority/write request, claim-value mismatch phải reject/abstain theo contract.
- Persist AgentProposal với proposal ID, decision/evidence refs, raw-output digest, validation result/version, provenance và authority ceiling. Chỉ proposal hợp lệ được resolve cho M08; HUMAN_REVIEW không tự thành approval.
- Ca bắt buộc: lợi nhuận bịa với snapshot đúng; free-text trái field value; prompt injection trong tool body; false write_permission thiếu/null/string; hợp lệ ABSTAIN và field-only claim.

**05b: tool enforcement.**

- Fetch nằm sau policy enforcement tại mỗi request, kể cả redirect hop nếu profile hỗ trợ; mặc định không redirect. Method/host/scheme/port/userinfo/timeout/response-size và private-address policy được chốt bằng allowlist profile, không lấy host mới từ model.
- Normalize/register tool response vào store, ACK rồi mới refresh context và cho model cite. Trace do adapter ghi, gắn request ID/evidence IDs/content digest, không lấy tool_calls tự khai làm sự thật.
- Test local HTTP fixture qua transport seam chỉ dùng trong test: GET allow, method/host/redirect/timeout deny, sink fail thì không cite; allowed trace thật qua boundary, trace giả/write/injection bị chặn. Không chỉ test tên case hoặc parser viết lại.
- Chạy code node được lấy trực tiếp từ blueprint; endpoint/adapter dùng cùng core; có happy path và failure path parity với learner CLI.

**Nghiệm thu:** R01/R05 đóng; không yêu cầu provider trả phí để chứng minh guard offline. **Rollback:** vô hiệu hóa tool/proposal handoff khi version không tương thích, không quay lại behavior “tự gắn evidence IDs”.

### RP-06 — Backup/restore M00–M10 và graph có thể dùng lại

**Chạm tới:** `backup_command.go`, runtime store inventory, canonical loaders, manifest version và deployment runbook.

**Cập nhật thực hiện (2026-09-09, phạm vi hẹp):** learner backup đã lên
`affiliate-bot-backup/v3`: mỗi artifact ghi path, kind, kích thước, SHA-256 và
profile inventory; restore đòi inventory trùng khớp tuyệt đối. Runtime có canary phải có
M10 registry; runtime có `FAILED` execution phải có fixture outcome store.
Restore gọi canonical loader để kiểm reservation → execution và FAILED
execution → `MACHINE_EXECUTION` EffectRef trước khi trả `RESTORED`. Smoke tạo
governed M10 chain bằng Bot, kiểm manifest thiếu artifact, checksum-hợp-lệ
nhưng orphan outcome, replay/resolve sau process mới, budget và STOP. Đây chỉ
là một lát cắt RP-06; các hạng mục còn lại bên dưới vẫn mở.

**Cập nhật M07 snapshot (2026-09-09):** manifest v3 inventory đệ quy theo
relative path chuẩn hóa và đưa `history.jsonl.m07/` vào backup khi adapter đã
persist tool trace/proposal. Trace lưu registry đã được review; loader replay
chính validator M07 cho method/host/redirect, trace digest, canonical history,
grounding và link AgentProposal → M08 intent trước `RESTORED`. Test tạo trace
và proposal bằng HTTP adapter của learner Bot, restore sang runtime trống rồi
thử proposal checksum-hợp-lệ nhưng hỏng. Backup từ symlink, writer lock hoặc
target nằm trong runtime bị từ chối. Đây chưa thay thế snapshot transaction,
coverage full M00–M10 hoặc n8n/model operated evidence.

**Cập nhật snapshot gate (2026-09-09):** writer của learner runtime và
`backup create` cùng giữ một runtime gate cross-process. M06/M07 history and
sidecar, M08–M11 mission state, action/outcome/evaluation/review và import
ACCESSTRADE trả `BUSY` khi snapshot hoặc writer khác đang giữ gate; stale gate
không tự bị xóa. Manifest chỉ publish sau toàn bộ copy thành công. Regression
giữ gate rồi thử backup/history write, và inject lỗi copy để xác nhận target
partial không có manifest/không restore được. Chưa có filesystem snapshot
transaction đa-host, kill/power-loss proof hoặc coverage mọi external store.

**Cập nhật graph M00–M05 và expiry (2026-09-09):** backup profile bây giờ
derive inventory action → outcome → evaluation → proposal → review từ file có
thực, dùng canonical loader/link validator ở source và restored target. Một
outcome orphan dù manifest checksum đã cập nhật bị reject. Loader state chỉ
kiểm toàn vẹn/binding lịch sử khi restore; intent/approval/grant hết hạn vẫn
replay được để audit nhưng gate/reservation mới kiểm thời gian thực và bị chặn.
Regression tạo M03–M05 qua CLI, backup/restore, resolve review, rồi cover
orphan và expired authority. Manifest v3 từ chối layout lạ, file known nhưng
không nằm trong inventory, và kind sai; mọi artifact runtime hiện hỗ trợ đều có
kind/size/SHA-256/profile. Kill/power-loss đa file, source ngoài runtime và
evidence operated vẫn mở.

**Cập nhật consistency gate (2026-09-09):** trước khi copy, backup chụp typed
inventory của source; sau canonical validation chụp lại để phát hiện thay đổi
trong lúc chuẩn bị, rồi đối chiếu từng file đã copy với snapshot ban đầu. Lệch
một file trả `SNAPSHOT_CONFLICT`, không publish manifest và không thể restore.
Regression mô phỏng đồng thời thay đổi mission state cùng M07 proposal giữa
copy; đây là fault seam nội bộ, không phải quyền cho writer vượt runtime gate.

**Cập nhật M10 active-canary registry link (2026-09-10):** backup create và
restore staging đều canonicalize `state.Canary.CanaryGrant`, rồi yêu cầu đúng
entry bất biến `CANARY_GRANT` (cùng ID, hash và bytes) trong M10 registry.
Regression từ chối source thiếu link, và từ chối backup đã cập nhật
checksum/manifest nhưng xóa entry trước khi staging có thể trở thành runtime.
RP-06 vẫn `PARTIAL`: các loại semantic orphan M10 khác cùng crash/host proof
vẫn chưa được chứng minh.

**Cập nhật M10 reservation-authorization link (2026-09-10):** reservation
`GOVERNED_AUTHORIZATION` đã tiêu budget nhưng chưa có execution record cũng
phải resolve authorization bất biến với cùng grant, intent và cost-bound
binding. Backup create và restore staging cùng từ chối authorization bị xóa;
regression sửa manifest checksum sau khi tamper để kiểm graph gate, không chỉ
checksum. Các orphan state-to-registry còn lại vẫn mở.

**Cập nhật M10 reservation-execution link (2026-09-10):** khi reservation đã
ghi `ExecutionID`, ID đó phải resolve execution record registry cùng
`AuthorizationID`; một ID tùy ý trong mutable state không đủ để giữ budget đã
tiêu. Cả source và restore staging từ chối backup tamper có checksum/manifest
hợp lệ. Các match semantic sâu hơn và crash/host proof vẫn mở.

**Cập nhật M10 canary usage ledger (2026-09-10):** `ExecutionsUsed` và
`CostUsedMinor` trong mutable canary state phải đúng bằng số/tổng cost của mọi
reservation đã persist; một state hạ counter sau reservation không thể mở lại
budget. Backup create và restore manifest-checksum-hợp-lệ đều reject mismatch.
Đây chưa là chứng minh crash/power-loss hoặc đầy đủ ledger external.

**Cập nhật M10 reservation lifetime (2026-09-10):** timestamp reservation
governed phải nằm từ `AuthorizedAt` đến trước `ExpiresAt` của authorization
bất biến; state đã sửa timestamp sau expiry bị reject ở source và restore
staging dù manifest/checksum đã cập nhật. Historical expiry vẫn được replay;
đây chỉ chặn issuance timestamp không thể có.

**Cập nhật M10 reservation cost-bound link (2026-09-10):** mọi reservation
có `CostBoundID`/hash phải resolve trusted-cost-bound registry đúng ID, hash,
intent và amount, kể cả đường legacy compatibility. Regression xóa bound cùng
gate để registry còn internally valid, rồi kiểm source và staging restore đều
reject sau khi manifest/checksum được cập nhật.

**Cập nhật M10 fixture-outcome cardinality (2026-09-10):** learner command và
loader chỉ cho một local no-side-effect outcome cho mỗi `MACHINE_EXECUTION`.
Regression dựng authorization → reservation → failed execution bằng Bot, chặn
outcome thứ hai lúc append, rồi kiểm source/restore reject JSONL duplicate sau
khi manifest/checksum được cập nhật. Đây không là ingestion business outcome.

**Cập nhật M10 execution lifetime (2026-09-10):** core terminal fixture record
và M10 registry graph cùng đòi `AttemptedAt` từ `AuthorizedAt` đến trước
`ExpiresAt`. Learner command từ chối ghi record đúng thời điểm expiry mà không
mutation; regression backup thay record trong registry, cập nhật checksum và
manifest, rồi bị reject. Không biến fixture no-side-effect thành live execution.

**Cập nhật M10 execution-reservation order (2026-09-10):** terminal record
phải có `AttemptedAt` không sớm hơn `ReservedAt`, vì reservation mới là commit
budget trước execution. Learner command reject record đến trước reservation mà
không mutation; backup/restore reject registry tamper có checksum/manifest hợp
lệ. Không suy điều này thành proof cho multi-file crash hoặc executor thật.

**Cập nhật M10 registry/state retry seam (2026-09-10):** fault test dừng
`m10-record-failed` sau khi registry append nhưng trước atomic commit
`mission-state`. Lúc dở dang, portable output không xuất hiện; journal M10
bền yêu cầu `status` và `m10-resolve` fail-closed. Lệnh writer có lock kế tiếp
canonical-validate journal, append/replay record nếu cần, bind đúng reservation
rồi mới xử lý retry exact và cho backup hợp lệ. Backup create chạy cùng recovery
trước inventory; regression xác nhận sau recovery nó chỉ bị chặn bởi thiếu
fixture outcome, không phải journal pending. Regression journal có record hợp lệ nhưng
`reservation_id` không resolve chứng minh writer và backup trả
`RECOVERY_REQUIRED`, giữ nguyên journal/state/registry; tamper không được đoán
hoặc bỏ qua. Đây là recovery cục bộ có giới hạn, **không** là transaction đa-file hay
power-loss proof.

- Định nghĩa inventory M00–M10: history, human action/outcome, evaluation, improvement/review, AgentProposal đã persist từ RP-05, intent/policy/per-action approval, canary grant/grant approval, TrustedCostBound, gate decision, ExecutionAuthorization, ExecutionRecord, machine outcome/EffectRef, reservation/pre-post ledger/consumed markers/STOP, và nested advisor bundle. Các artifact từ RP-03/RP-05 phải được tạo qua runtime trong test snapshot, không dựng file placeholder. Khai báo store ngoài runtime root; không tự gom secret hoặc gọi toàn bộ home là backup.
- Manifest dùng relative paths chuẩn hóa và checksum/size/type/version. Backup đệ quy theo inventory; layout chưa hỗ trợ phải reject rõ thay vì skip thư mục. Reject symlink/path traversal và đích backup nằm trong source gây self-inclusion.
- Quiesce writer/lock snapshot hoặc dùng snapshot transaction. Copy bytes nhất quán với ledger/state/STOP; chỉ publish manifest sau snapshot hoàn tất.
- Restore vào staging trống, verify manifest rồi load từng artifact bằng canonical decoder và resolver; replay history, kiểm full graph trong inventory M00–M10, gồm proposal→intent và cost-bound→gate→authorization→execution→outcome/EffectRef, usage, consumed approval, expiry và STOP. Không emit RESTORED trước các gate.
- Test bundle từ `advisor fixture-run`, AgentProposal từ RP-05, execution chain EC-01 từ RP-03; orphan từng loại artifact, corrupted payload có checksum đúng do backup create, expired approval, missing ledger, nonempty target, interrupted copy và concurrent writer. Sau restore chạy lại EC-02…EC-05 bằng process mới, kiểm list/replay/status/reserve/resolve, không chỉ so JSON field. Hết hạn tại thời điểm restore không làm mất historical artifact hợp lệ: vẫn đọc/replay được, nhưng operation mới phải bị chặn; không gia hạn hoặc reissue quyền.
- Profile manifest phải nêu rõ phạm vi/version. Trước phần mở rộng RP-07a, gặp artifact M11 chưa được hỗ trợ thì reject tường minh; không skip hoặc báo restore toàn M00–M11. Không tạo lease/activation placeholder để đạt test.

**Nghiệm thu RP-06:** snapshot/restore M00–M10 PASS và giữ budget/STOP, proposal/authorization/execution/EffectRef resolve đúng. R12/R13 được ghi evidence cho phạm vi này nhưng **chưa đóng toàn phạm vi M00–M11**; chỉ đóng sau gate mở rộng restore RP-07a dưới đây. **Rollback:** giữ source/backup nguyên trạng; staging lỗi không trở thành active runtime; người vận hành quyết định cleanup/recovery, không tự ghi đè đích cũ.

### RP-07a/07b — M11 lifecycle và chain learner đầy đủ

**Cập nhật thực hiện (2026-09-08, artifact spine):** `core/m11` nay sở hữu
decoder/semantic checks cho M11 artifact; harness gọi decoder này thay vì parser
semantic riêng. Learner có `m11-register`/`m11-resolve` và registry append-only
`m11-artifacts.jsonl`, reject hash/schema/link hỏng, và backup v3 snapshot/verify
registry nếu có. `m11-activate` chỉ tạo activation từ lease + approval đã
register, trong thời hạn và khi chưa STOP; `m11-ledger-init` chỉ tạo ledger
rỗng sau activation này và không reset được bằng retry. Smoke backup tạo
lease/approval/activation/ledger qua learner, restore rồi resolve bằng process
mới. `m11-gate` resolve exact lease/activation/health/cost/ledger, persist
gate không-authorizing và fail closed với scope, time, budget hoặc health lỗi.
`m11-authorize` tạo quyền governed có expiry bị chặn bởi lease/intent/approval/
cost từ gate ALLOW đã persist, nhưng chưa gọi executor. `m11-reserve-authorization`
chỉ tạo pre-ledger immutable một lần cho authorization đó, charge bound và giữ
pending execution; `m11-record-failed` chỉ có đường fixture `FAILED`/
`NOT_PERFORMED` và không gọi executor. Pending chỉ được giải phóng bởi
`m11-outcome` khi fixture `CANCELLED` có đúng `MACHINE_EXECUTION` EffectRef
được ghi; lệnh này tạo post-ledger và vẫn hoạt động để đóng observation sau STOP.
Business outcome, reconciliation và reviewed recovery bên dưới vẫn còn bắt buộc
trước khi đóng.

**Cập nhật reconciliation (2026-09-08):** learner có `m11-record-unknown` cho
fixture effect không xác định. Nó ghi `RECONCILIATION_REQUIRED`/`UNKNOWN`,
ledger STOPPED và durable STOP trước khi trả kết quả. `m11-reconcile` chỉ nhận
resolution đã nằm trong registry, do `human` xác nhận đúng unknown execution và
chỉ chốt ledger ở `RECOVERY_REVIEW_REQUIRED`; lệnh không thể kích hoạt lại lease
hoặc cấp authorization mới. `smoke_br18b_backup_restore.py` chạy runtime riêng
qua UNKNOWN → STOP → resolution → backup/restore → restart và xác nhận lease
cũ vẫn bị reject; smoke cũng sửa stopped ledger với checksum manifest hợp lệ và
xác nhận restore trả `GRAPH_FAILED`. Cần thêm trace fault-injection/
concurrent-writer và reviewed recovery bằng **lease mới** trước khi coi RP-07a
hoàn tất.

**Cập nhật recovery handoff (2026-09-08):** `m11-recovery-export` chỉ đọc
stopped ledger + registered human resolution và xuất proof có
`requires_new_runtime=true`, `requires_new_lease=true`,
`execution_permitted=false`. Nó không reset STOP, không copy state sang runtime
mới và không gọi executor. Handoff đã có smoke; workflow tạo runtime mới, lease
mới và approval mới vẫn phải được thiết kế như một canonical artifact boundary
riêng, không thể suy ra chỉ từ export proof.

**Cập nhật recovery admission (2026-09-09):** `m11-recovery-admit` đã persist
`PRODUCTION_RECOVERY_ADMISSION` trong registry của **runtime mới**. Adapter
đọc runtime cũ fail-closed, canonicalize handoff và yêu cầu STOP + reviewed
resolution/ledger còn nguyên; runtime mới phải có state, lease/approval khác
identity/hash, activation và NORMAL ledger riêng. New approval chỉ hợp lệ sau
resolution cũ, admission review sau approval mới. Admission bind absolute
runtime paths, không nhận parent/child hoặc path trùng, không ghi vào runtime
cũ và luôn `execution_permitted=false`; health/cost/gate/authorization vẫn là
đường duy nhất tới fixture operation. Smoke BR-16a cover valid admission,
old-runtime lease append reject, prior lease/approval reuse reject, resolve và
backup/restore admission. Đây không phải live recovery/executor proof; graph
backup của runtime mới không thể tự chứng minh availability liên tục của
runtime cũ ngoài artifact handoff đã bind.

**Cập nhật admission restore drill (2026-09-09):** BR-18b nay thực hiện lại
admission từ **old runtime đã restore** và vẫn STOPPED sang runtime mới với
lease/approval/activation/NORMAL-ledger riêng. Smoke backup/restore runtime
mới, resolve `PRODUCTION_RECOVERY_ADMISSION`, rồi sửa `new_lease_hash` trong
backup và đồng bộ checksum manifest. Smoke cũng sửa `lease_hash` trong new
`PRODUCTION_LEASE_APPROVAL` với checksum registry/manifest hợp lệ. Cả hai
restore trả `VERIFY_FAILED` trước publish vì loader runtime mới phát hiện
admission hoặc approval không còn link exact tới new lease. Đây chỉ là
validation deterministic của snapshot; không suy diễn atomic multi-file crash
recovery hoặc availability của old runtime ngoài handoff persisted.

**Cập nhật UNKNOWN→STOP journal (2026-09-09):** `m11-record-unknown` tạo
`m11-unknown-stop-journal/v1` trước khi append execution và stopped ledger.
Replay kiểm exact predecessor ledger, UNKNOWN execution và stopped transition;
chỉ sau đó mới repair durable mission STOP/marker rồi xóa journal. Nếu journal
malformed, stale hoặc competing thì mutation và `status` fail closed
`RECOVERY_REQUIRED`, không tự đoán side effect. Fault test tiêm lỗi ngay trước
và ngay sau stopped-ledger write sau execution append, cùng lỗi atomic rename
của mission state hoặc STOP marker; restart replay append ledger/STOP
exactly-once và retry trả duplicate. Backup create cũng chạy recovery dưới lock
trước inventory; regression tạo journal pending thật rồi backup/restore runtime
và kiểm STOP vẫn bền. Journal là kế hoạch replay có fsync cho một
transition, **không
phải** transaction đa-file/power-loss proof.

**Cập nhật M11 outcome-journal ACK fault (2026-09-10):** regression gọi trực
tiếp `recoverM11OutcomeJournal` với fault sau khi JSONL outcome đã append nhưng
trước khi caller nhận ACK. Journal phải còn lại, outcome và ledger transition
đã ghi phải resolve đúng một lần; replay sau restart-style retry chỉ nhận
`EXACT_DUPLICATE`, xóa journal và không thêm outcome/ledger mới. Đây là bằng
chứng deterministic cho đúng recovery seam đó, không chứng minh atomicity
đa-file khi mất điện hoặc external business outcome.
`backup create` cũng được test recovery journal này trước inventory, rồi
restore xác nhận cùng outcome và post-outcome ledger từ runtime mới.

**Cập nhật complete lease-approval snapshot link (2026-09-10):** registry có
thể append lease trước approval trong flow cấp phát, nhưng backup/restore yêu
cầu graph hoàn chỉnh: mọi lease đã persist phải resolve đúng approval theo ID,
version/hash, promotion/canary refs và reviewer/time. Smoke tạo backup runtime
mới thật, xóa approval M11 rồi cập nhật checksum/manifest; restore trả
`GRAPH_FAILED` trước publish. Đây chỉ chứng minh semantic snapshot validation
cho link đó, không là external review evidence hoặc atomic multi-file proof.

**Cập nhật ledger-activation snapshot link (2026-09-10):** lease có thể chưa
active trong registry, nhưng ledger đã persist phải resolve activation cùng
lease/version/hash với thời điểm activation không muộn hơn ledger. Smoke xóa
activation của runtime có ledger, cập nhật checksum/manifest và restore trả
`GRAPH_FAILED` trước publish. Đây chỉ là kiểm graph deterministic, không suy
ra live activation, external execution hay crash atomicity.

**Cập nhật fixture-outcome execution link M11 (2026-09-10):** mọi fixture
outcome M11 trong snapshot phải có `MACHINE_EXECUTION` EffectRef resolve tới
execution record đã persist, kể cả khi chưa có evaluation/cycle. Smoke đổi
EffectRef thành execution không tồn tại rồi cập nhật checksum/manifest; restore
trả `GRAPH_FAILED` trước publish. Đây chỉ chặn orphan trong graph offline, không
chứng minh business outcome hay external effect.

**Cập nhật read/export boundary của M11 journal (2026-09-10):** trong khi bất
kỳ journal `FAILED`, `UNKNOWN→STOP` hoặc outcome nào còn pending,
`m11-resolve` và `m11-recovery-export` trả `RECOVERY_REQUIRED` trước khi đọc
registry, resolve artifact hoặc tạo handoff. Fault regression tạo journal
UNKNOWN→STOP thật sau lỗi ghi stopped ledger, rồi chứng minh cả resolver lẫn
export đều không xuất partial lifecycle data hay file handoff; locked writer
recovery vẫn là đường duy nhất để khôi phục. Guard này chỉ ngăn read-path quan
sát trạng thái dở dang, **không** chứng minh atomicity đa-file, power-loss
recovery hay external execution/outcome.

**Cập nhật coverage ba journal read boundary (2026-09-10):** fault regressions
giờ giữ pending journal trên từng transition `FAILED`, `UNKNOWN→STOP` và
outcome, rồi gọi CLI `m11-resolve` vào artifact đã tồn tại. Cả ba phải trả
`RECOVERY_REQUIRED`; test không dùng parser mô phỏng. UNKNOWN vẫn kiểm thêm
recovery export/admission không tạo handoff/admission dở dang. Đây chỉ mở rộng
bằng chứng cho guard đọc chung, không là bằng chứng replay atomic hay crash
coverage ngoài các fault seam cụ thể.

**Cập nhật recovery-admission source boundary (2026-09-10):**
`m11-recovery-admit` và helper handoff giờ kiểm journal M11 của **runtime cũ**
trước khi đọc lifecycle/handoff hoặc ghi admission vào runtime mới. Regression
giữ UNKNOWN→STOP journal pending thật rồi gọi admission với input còn thiếu:
nó phải trả `RECOVERY_REQUIRED` thay vì đọc source/inputs khác, và registry mới
không được có partial admission. Đây chỉ ràng buộc source đọc của handoff; nó
không khóa snapshot đa-runtime, không là transaction giữa hai runtime và không
chứng minh power-loss/external recovery.

**Cập nhật M11 failed-execution journal (2026-09-10):** `m11-record-failed`
ghi `m11-failed-execution-journal/v1` trước cặp immutable
`FAILED/NOT_PERFORMED` execution và post-execution ledger. Recovery chỉ replay
khi execution và ledger đúng transition của predecessor; journal malformed hoặc
head cạnh tranh vẫn chặn mutation. Fault test cover trước/sau append ledger,
`status=RECOVERY_REQUIRED`, recovery exact-once và retry `EXACT_DUPLICATE`.
`backup create` cũng replay journal trước inventory; regression yêu cầu outcome
fixture sau recovery rồi backup/restore runtime thật.
Đây không biến hai file thành transaction power-loss hoặc chứng minh external
execution/business outcome.

**Cập nhật M11 reconciliation head guard (2026-09-10):** `m11-reconcile`
chỉ tạo transition resolution từ stopped ledger đang là head. Retry exact của
resolution đã nằm trên head vẫn trả `EXACT_DUPLICATE`, nhưng resolution khác
trỏ predecessor cũ bị từ chối trước append, không thể fork ledger để làm mất
link resolution trước đó. Đây là guard lineage offline; không cho phép nối lại
lease cũ hay chứng minh recovery/execution ngoài fixture.

**Cập nhật M11 health-after-activation boundary (2026-09-10):** `m11-gate`
từ chối health snapshot được quan sát trước activation của lease, dù hash,
scope và tuổi snapshot còn hợp lệ; `m11-authorize` kiểm lại cùng quan hệ để
không dùng gate cũ/đã persist sai làm quyền mới. Backup/restore cũng từ chối
snapshot có checksum và manifest hợp lệ nhưng health của gate có timestamp trước
activation. BR-18b gọi learner Bot cho ca gate âm và mutation restore thực tế.
Đây chỉ là bảo toàn lineage thời gian cho fixture offline, không là telemetry
thật, business outcome hay execution authority.

**Cập nhật M11 activation lease-window graph (2026-09-10):** canonical M11
registry decoder giờ reject activation có timestamp trước `ValidFrom` hoặc tại/
sau `ExpiresAt` của exact lease, nên backup create và restore staging cùng fail
closed với artifact checksum/manifest hợp lệ nhưng admission lifecycle không
thể xảy ra. Core regression mutate activation trực tiếp; BR-18b mutate backup
qua learner Bot. Đây chỉ kiểm graph timestamp của fixture runtime, không chứng
minh clock đáng tin cậy, quyền executor hay production activation.

**Cập nhật M11 ledger-after-activation graph (2026-09-10):** khi activation
đã được đăng ký, canonical graph reject ledger có `window_started_at` hoặc
`updated_at` trước activation; BR-18b mutate post-ledger checksum hợp lệ và
restore fail-closed. Registry draft chưa activation vẫn đọc được; backup hoàn
chỉnh đã yêu cầu activation riêng. Đây chỉ là invariant fixture offline.

**Cập nhật M11 recovery-admission restore graph (2026-09-10):** backup graph
decode `PRODUCTION_RECOVERY_ADMISSION` và yêu cầu exact new lease, approval,
activation cùng normal non-reconciliation ledger; admission luôn non-authorizing.
Không đọc hay suy diễn runtime cũ từ path. BR-18b xóa new ledger trong backup
checksum-hợp-lệ và restore `GRAPH_FAILED`; không chứng minh recovery đa-runtime.

**Cập nhật M11 recovery-admission cardinality (2026-09-10):** learner và
backup graph reject admission thứ hai dùng cùng new lease hoặc cùng
prior-runtime/resolution lineage. BR-18b gọi command thật với ID admission khác
và nhận `REJECTED`; đây chỉ chống audit ambiguity offline.

**Cập nhật M11 lease expiry/rebind regression (2026-09-11):** learner
reject activation tại exact `ExpiresAt` mà không mutation registry; same
lease ID với payload/hash canonical khác cũng bị reject. Chưa là coverage
expiry/rebind đầy đủ cho mọi M11 artifact hay production authority.

**Cập nhật M11 health lease-window graph (2026-09-11):** canonical graph
reject health snapshot trước `ValidFrom` hoặc tại/sau `ExpiresAt` của exact
lease. Core regression dùng health tại expiry; BR-18b mutate backup checksum
hợp lệ và restore `GRAPH_FAILED`. Đây chỉ là time-lineage offline, không là
telemetry đáng tin cậy hay production health operation.

**Cập nhật M11 health-after-activation graph (2026-09-11):** khi activation
đã tồn tại, canonical registry reject health snapshot trước activation. BR-18b
gọi learner để reject đăng ký này và dùng health mới đúng activation cho ca
gate-before-activation; mutation restore vẫn fail closed. Draft chưa activation
vẫn không bị suy diễn là active runtime. Đây chỉ là lineage fixture offline.

**Cập nhật M11 gate lease-window graph (2026-09-11):** canonical graph reject
gate trước `ValidFrom` hoặc tại/sau `ExpiresAt` của exact lease. Core regression
và BR-18b mutation backup checksum-hợp-lệ tại expiry đều fail closed. Đây không
xác nhận provider clock hay production gate operation.

**Cập nhật M11 authorization lease-window graph (2026-09-11):** canonical
graph reject authorization trước `ValidFrom`, tại/sau `ExpiresAt` của lease,
hoặc mang expiry vượt lease. Core regression và BR-18b checksum-valid mutation
fail closed. Đây không là proof authority/executor production.

**Cập nhật M11 gate cost-bound window graph (2026-09-11):** canonical graph
đòi gate evaluation từ `TrustedCostBound.observed_at` đến trước bound expiry,
và currency bound khớp lease. Core regression và BR-18b checksum-valid mutation
tại cost expiry fail closed; không là provider-cost hay production proof.

**Cập nhật M11 authorization gate/health freshness graph (2026-09-11):**
canonical authorization phải bind gate `ALLOW_PRODUCTION`, không trước gate/
health, và health age nhỏ hơn lease limit tại thời điểm authorize. Core regression
chặn non-ALLOW gate và stale health; BR-18b checksum-valid mutation chặn stale
health. Đây không chứng minh health provider hay authority production.

**Cập nhật M11 execution authorization-lifetime graph (2026-09-11):**
canonical core registry đòi `AttemptedAt` từ `AuthorizedAt` đến trước
authorization expiry, đồng bộ backup/runtime guard. Core regression và BR-18b
set auth expiry đúng execution attempt đều fail closed. Không là executor hoặc
atomic crash proof.

**Cập nhật M11 reconciliation semantic graph (2026-09-11):** canonical core
registry nay chỉ nhận resolution của `human`, `NOT_PERFORMED`, cho execution
`RECONCILIATION_REQUIRED`/`UNKNOWN`, và resolution không thể có trước attempt.
Core regression tạo một UNKNOWN fixture hợp lệ rồi reject resolution cho FAILED,
`PERFORMED`, hoặc trước attempt. Điều này đồng bộ graph core với command/restore
boundary; chưa chứng minh reconciliation bên ngoài, STOP graph hoàn chỉnh hay
khả năng recovery/execution production.

**Cập nhật M11 reconciliation cardinality graph (2026-09-11):** canonical
registry chỉ nhận tối đa một resolution cho mỗi UNKNOWN execution, đồng bộ với
restore graph và head-guard của learner. Core regression thêm artifact resolution
khác ID nhưng cùng execution và bị reject. Đây không thay thế human review,
không chứng minh external reconciliation, và không làm durable STOP có thể mở lại.

**Cập nhật M11 evaluation/cycle temporal graph (2026-09-11):** canonical
registry yêu cầu offline evaluation không trước execution attempt và cycle chỉ
đóng từ evaluation đã tồn tại, tại/sau `evaluated_at`. Core regression reject
evaluation trước attempt và cycle đóng trước evaluation. Đây chỉ đồng bộ temporal
lineage core với restore checker; không chứng minh outcome business hay lifecycle
ngoài fixture.

**Cập nhật M11 cycle closure-status graph (2026-09-11):** canonical registry
chỉ coi `ProductionCycleRecord` là lifecycle closure khi `status=CLOSED`, đồng
bộ với restore graph. Core regression reject record `REVIEW_PENDING` dù các
reference/time khác hợp lệ. Đây không thêm workflow pending/review hay chứng
minh operation/outcome bên ngoài fixture.

**Cập nhật M11 authorization/execution cardinality graph (2026-09-11):**
canonical registry chỉ nhận một immutable `ProductionExecutionRecord` cho mỗi
authorization. Core regression append execution thứ hai với ID khác nhưng giữ
mọi authorization binding hợp lệ và bị reject. Đây khớp single-reservation
learner boundary, không chứng minh idempotency của executor ngoài fixture hay
transaction đa-file.

**Cập nhật M11 evaluation/cycle cardinality graph (2026-09-11):** canonical
registry chỉ nhận một offline evaluation và một closed cycle cho mỗi execution,
đồng bộ với learner command. Core regression append evaluation/cycle thứ hai có
ID khác nhưng cùng execution rồi reject. Đây không mở rộng evidence outcome,
không thay thế review workflow và không chứng minh business lifecycle.

**Cập nhật M11 authorization cost-bound window graph (2026-09-11):**
canonical authorization phải được cấp trong exact `TrustedCostBound` window và
hết hạn không sau bound expiry. Core regression thay bound/gate/auth thành một
lineage checksum-valid, rồi reject authorization còn sống sau cost expiry. Đây
không chứng minh cost provider, clock vận hành hay production authority.

**Cập nhật M11 execution/reservation lineage graph (2026-09-11):** sau khi
decode toàn bộ inventory append-only, canonical registry yêu cầu mọi execution
có một NORMAL ledger trước/equal attempt, cùng lease/version/hash và giữ ID đó
trong `pending_execution_ids`. Core regression dùng pre-ledger hợp lệ rồi reject
inventory bỏ ledger này. Đây chỉ đồng bộ core với backup/learner fixture; không
chứng minh atomic transaction, budget provider hay external execution.

**Cập nhật M11 gate/ledger state graph (2026-09-11):** canonical gate
phải bind exact NORMAL, non-reconciliation ledger cùng lease lineage, và bốn
budget counters trên gate phải khớp immutable ledger. Core regression reject
ALLOW gate từ stopped ledger và gate có counter drift. Đây chỉ đồng bộ core
với learner/restore checker; không là production budget hay gate operation proof.

**Cập nhật M11 gate activation boundary (2026-09-11):** canonical gate
phải resolve exact activation cùng lease lineage và `evaluated_at` không trước
`activated_at`. Core regression giữ health sau activation nhưng đặt gate sớm hơn
rồi reject; BR-18b gọi command thật và xác nhận registry chặn candidate trước
khi persist. Đây chỉ là temporal lineage fixture, không chứng minh health/gate
operation production.

**Cập nhật M11 authorization executor scope graph (2026-09-11):**
canonical authorization chỉ hợp lệ nếu `executor_id` nằm trong immutable
`lease.executor_ids`. Core regression dùng authorization cùng gate/health/cost
nhưng executor lạ và bị reject. Đây là scope lineage offline, không chứng minh
danh tính executor, provider hoặc execution production.

**Cập nhật M11 gate/health temporal order (2026-09-11):** canonical gate
không thể evaluate trước `health.observed_at` của exact snapshot. Core regression
giữ hash/lease/cost/activation hợp lệ nhưng dời health sau gate và bị reject.
Đây chỉ là fixture chronology, không là telemetry freshness hay production health
proof.

**Cập nhật M11 ALLOW gate health safety graph (2026-09-11):** canonical
`ALLOW_PRODUCTION` gate chỉ hợp lệ với health snapshot telemetry complete,
dependency `HEALTHY`, không alert/reconciliation và nằm trong các ngưỡng lease.
Core regression giữ lineage/hash hợp lệ nhưng đổi dependency sang `DEGRADED`
rồi reject. Đây không là telemetry provider hay production health proof.

**Cập nhật M11 ALLOW gate budget graph (2026-09-11):** canonical
`ALLOW_PRODUCTION` gate chỉ hợp lệ khi referenced ledger còn execution/window/
pending/cost capacity cho exact cost bound. Journal fixtures nay tạo gate trên
pre-ledger rỗng rồi append reservation ledger, đúng command order; core regression
reject ALLOW gate có ledger exhausted. Đây không chứng minh external budget,
concurrency hay transaction đa-file.

**Cập nhật M11 ALLOW gate risk scope (2026-09-11):** canonical
`ALLOW_PRODUCTION` gate phải có `risk_class` nằm đồng thời trong immutable
`lease.allowed_risk_classes` và `lease-approval.validated_risk_classes` exact
của lease. Core regression chặn gate RISK1 dưới lease RISK0 và approval không
validate RISK0; learner/backup dùng chính graph validator này trước append/publish.
Đây là scope lineage offline, không là human approval hoặc production authority.

**Cập nhật M11 reservation/authorization ordering (2026-09-11):** canonical
execution chỉ resolve reservation ledger NORMAL có `updated_at` từ thời điểm
immutable authorization được cấp đến execution attempt. Core regression dời
authorization sau pre-ledger và reject, nên một pending execution ID được ghi
trước authorization không thể hợp thức hoá attempt về sau. Đây là sequencing
lineage offline, không chứng minh lock hoặc transaction đa-file.

**Cập nhật M11 ledger window continuity (2026-09-11):** canonical ledger
history cho cùng lease không được đổi `window_started_at` hoặc giảm
`executions_in_window`; counters total/cost vẫn monotonic như trước. Core
regression append ledger có window reset và counter window giảm, rồi reject.
Điều này giữ fixed-window budget lineage của runtime hiện tại; không triển khai
rolling-window reset hoặc chứng minh concurrent accounting.

**Cập nhật M11 pending reservation continuity (2026-09-11):** giữa hai ledger
entries của cùng lease, số pending execution ID mới phải đúng bằng delta
`executions_total`, và `executions_in_window` phải tăng cùng delta. Core
regression thay pending ID mà không tạo reservation/counter mới rồi reject.
Đây chặn forged ledger swap trong deterministic graph; outcome/reconciliation
transitions và crash atomicity đa-file vẫn là các phạm vi riêng.

**Cập nhật reservation concurrency (2026-09-08):** `m11-reserve-authorization`
quét canonical registry trước khi append. Exact retry của cùng artifact trả
`EXACT_DUPLICATE`; request cùng authorization nhưng timestamp/ledger artifact
khác bị reject, nên không thể charge budget hai lần chỉ bằng cách dùng lại
pre-ledger. Smoke BR-18b chạy trực tiếp ca âm này. Fault injection giữa nhiều
file append vẫn là việc riêng, chưa được coi là transaction đa-file. Smoke cũng
giả lập partial reconciliation write (UNKNOWN execution + resolution còn nhưng
reviewed stopped-ledger mất) với checksum manifest hợp lệ; restore fail-closed
`GRAPH_FAILED`. Chưa có crash hook thực thi tại từng `write/sync/rename`.

**Cập nhật multi-process reservation (2026-09-08):** BR-18b chạy hai process
cùng `m11-reserve-authorization` với cùng authorization và timestamps khác.
Chỉ một process được `APPENDED`; process còn lại `BUSY` hoặc `REJECTED`, và retry
sau `BUSY` bị reject. Đây kiểm lock + canonical registry guard; không thay CAS
hoặc transaction đa-file.

**Cập nhật fault seam (2026-09-08):** registry M11 có hook nội bộ chỉ dùng trong
test (không nhận từ CLI/env). Test inject lỗi sau `write` nhưng trước khi caller
nhận success: loader vẫn đọc được artifact hoàn chỉnh và retry trả
`EXACT_DUPLICATE`. Đây chứng minh recovery cho một append artifact; không suy
ra transaction cho cặp registry/ledger/outcome/STOP hoặc lỗi trước/giữa partial
filesystem write.

**Cập nhật atomic state seam (2026-09-08):** `writeJSONAtomic` có hook test
trước rename. Test chứng minh lỗi tại điểm đó giữ nguyên state cũ, không để
temporary file và retry commit semantic state mới. Hook không có đường kích hoạt
từ runtime; nó không chứng minh atomicity giữa mission-state và registry/STOP.

**Cập nhật stopped-ledger crash seam (2026-09-08):** loader fail-closed nếu
M11 registry có ledger `STOPPED` nhưng mutable state chưa STOP. BR-18b tạo
snapshot UNKNOWN/STOP rồi sửa manifest hợp lệ để mất state/marker STOP, giữ
stopped ledger; restore trả `VERIFY_FAILED`. Đây là recovery guard, không phải
transaction đa-file hay power-loss proof.

**07a:** tách/reuse M11 lease activation, health gate, ledger/reconciliation, STOP, reviewed recovery từ harness; thêm learner entrypoint và store links. Offline executor stub không được gọi là live execution. Persist lifecycle artifact và version, kiểm time/authority/unknown usage/cycle closure; không chỉ thêm STOP/status alias.

RP-07a sở hữu graph M11: source canary/promotion review → lease + lease approval → activation; intent/policy + health snapshot + TrustedCostBound + pre-ledger → ProductionGateDecision → reservation/ExecutionAuthorization → ExecutionRecord → outcome có `MACHINE_EXECUTION` EffectRef → evaluation/cycle/post-ledger; trường hợp UNKNOWN/STOP có reconciliation resolution và reviewed recovery riêng. Tái sử dụng [M11 chain checker](../../lab/mission-runtime/cmd/demo/m11_chain.go), [production gate](../../contracts/production-gate-decision.schema.json) và canonical schemas; không bỏ gate/authorization/execution chỉ vì chạy offline.

**Gate restore M11 bắt buộc trước khi đóng RP-07a:**

- Mở rộng inventory/manifest version, decoder, graph resolver và snapshot/restore của RP-06 trong cùng gói RP-07a. Bao phủ lease/lease approval/promotion refs, activation, health, cost bound, gate, authorization, execution, pre/post/STOP ledgers, outcome/evaluation/cycle và reconciliation/recovery artifacts; giữ nguyên các artifact M00–M10.
- Sinh artifact bằng learner + executor stub, snapshot, restore vào đích trống rồi process mới replay/resolve đầy đủ. Kiểm `closed_cycle` và `resolved_stop` theo harness; STOP vẫn chặn, expiry không mở quyền mới, usage/consumed markers không reset. Mở rộng EC-01…EC-05 cho production profile.
- Negative cases: thiếu hoặc sai lease/activation/health/cost/gate/authorization/execution/EffectRef/cycle/resolution; dữ liệu bị sửa nhưng checksum backup hợp lệ; phiên bản manifest không hỗ trợ. Reject graph hỏng, không publish runtime sẵn dùng. Lịch sử đã hết hạn vẫn phục hồi được ở chế độ không cấp quyền; recovery chưa review phải bị chặn.
- Chỉ ghi đóng R12/R13 toàn phạm vi khi cả evidence RP-06 và gate restore M11 này PASS trên head tương thích. RP-07b và RP-09 không được nghiệm thu full chain/readiness nếu gate này chưa đạt. Dependency là RP-06 → RP-07a → RP-07b, không có vòng lặp.

**Cập nhật backup negative cases (2026-09-09):** `smoke_br18b_backup_restore.py`
tạo evaluation/cycle bằng learner Bot rồi restore các bản sao có checksum hợp lệ
nhưng outcome bị orphan khỏi evaluation, cycle trỏ evaluation không tồn tại, hoặc
`closed_at` sớm hơn evaluation. Tất cả bị `GRAPH_FAILED`; bước inventory chỉ đọc
envelope hợp lệ trước, còn graph chỉ quyết định sau staging restore.

**Cập nhật M10 temporal restore (2026-09-10):** regression learner tạo một
backup M10 hợp lệ, sau đó sửa `m10-outcomes.jsonl` để `observed_at` sớm hơn
`ExecutionRecord.attempted_at` và cập nhật lại checksum/size trong manifest.
Restore trả `GRAPH_FAILED`; checksum hợp lệ không thay cho kiểm tra quan hệ
outcome → execution. Đây chỉ bổ sung một negative temporal seam, không đóng
những malformed-graph hoặc crash seam còn lại của RP-06.

**07b:** thay smoke BR-16a bằng một workspace chung và cùng evidence/decision lineage:

**Cập nhật shared chain (2026-09-09):** `smoke_br16a_offline.py` hiện tạo M00
history, M07 registered proposal, M08 agent-bound intent/policy, M09 approval,
M10 cost/gate/authorization/fixture outcome rồi dùng **chính các artifact đó**
để tạo M11 lease, activation, health gate, authorization, reservation,
`FAILED`/`NOT_PERFORMED` fixture và M11 `MACHINE_EXECUTION` outcome. Smoke
resolve execution sau đó. Cùng workspace tiếp tục tạo recovery lease đã review
riêng (vì authorization đầu là one-time) nhưng giữ proposal/intent/policy/grant/
cost-bound gốc, rồi cover `UNKNOWN` → `RECONCILIATION_REQUIRED` durable STOP →
human reconciliation idempotent → read-only recovery handoff; activation lại
bị reject trước reconciliation. M07 tool/proposal đi qua loopback adapter và
persist sidecar trong chính runtime. Smoke backup/restore runtime đó vào đích
trống, replay history và resolve M07 context, M10 execution, M11 UNKNOWN,
resolution và stopped ledger trước khi xác nhận STOP tiếp tục chặn activation.
Mỗi lệnh là process mới nên STOP được đọc lại từ store. Nhánh
`FAILED`/`NOT_PERFORMED` fixture nay tạo `ProductionOutcomeEvaluation`
`OFFLINE_FIXTURE`/`FIXTURE_NO_SIDE_EFFECT`, rồi chỉ đóng audit cycle khi các
link outcome → execution → lease/gate/authorization/history đều exact; restore
còn từ chối outcome bị sửa dù manifest checksum đã được cập nhật. Đây không
phải business outcome/evaluation thật và live proof vẫn chưa có.

1. M00 packet → M01 evaluation → M02 history/decision.
2. M03 human action → outcome snapshots → M04 advisor dùng đúng history/action/outcome.
3. M05 evaluation → improvement proposal → human review/isolated regression; không auto-apply.
4. M06 watcher cập nhật đúng subject, đưa record ID mới vào context tiếp theo.
5. M07 adapter/model fixture → validated persisted AgentProposal → M08 intent tham chiếu proposal và original evidence IDs.
6. M09: M08 intent/policy → human approval fixture → ExecutionAuthorization → executor stub → ExecutionRecord → outcome/EffectRef. M10: intent/policy và grant/grant approval + TrustedCostBound + pre-ledger → gate → reservation/authorization → stub/execution → outcome/EffectRef → reconciliation/post-ledger. Mỗi attempt/profile có ID/idempotency riêng nhưng giữ lineage về proposal/evidence; không dùng lại one-time approval đã consumed.
7. M11: lease/approval/activation + intent/policy/health/cost/pre-ledger → gate → reservation/authorization → stub/execution → outcome/EffectRef/evaluation/cycle/post-ledger → STOP → restart → reviewed recovery bị giới hạn. Backup/restore cùng workspace phải qua gate mở rộng M11 của RP-07a trước khi công nhận chain hoàn tất.

Mỗi bước phải đọc artifact từ bước trước, không tạo ID/context thay thế để test pass. Khi vòng đời cần một human input mới, lưu reviewer fixture riêng và link đúng artifact; không giả làm approval thật. Kiểm toàn graph sau restart, có negative break-link tại từng seam và no-write assertions khi reject.

Walkthrough phải có lệnh build, input paths/fixtures được version control, expected outputs, retry/conflict, deliberate FAIL/fix, restart/STOP/restore và cleanup có phạm vi rõ. Người mới không phải tự đọc smoke Python để suy ra input JSON.

**Nghiệm thu:** R14 đóng ở phạm vi offline chỉ khi gate lifecycle/restore RP-07a và chain RP-07b đều PASS, mọi authorization/execution/cost/gate/EffectRef resolve được sau restart/restore. Report liệt kê rõ fixture/model stub/executor stub; `RESERVED` không thay ExecutionRecord, không in “M00–M11 PASS” nếu bỏ chặng/artifact. **Rollback:** không promote state chưa đóng cycle; giữ STOP và artifacts để chẩn đoán.

### RP-08 — CI bắt được implementation sai

- Baseline tests phải tiếp tục chạy. Mỗi PR trên thêm regression vào cùng workflow trước khi merge; không để tests chỉ nằm trong thư mục tạm của reviewer.
- Wire `smoke_br16a_offline.py` và `smoke_br18b_backup_restore.py` vào `.github/workflows/`; test blueprint đọc và thực thi chính `jsCode` hoặc gọi adapter shared, không tự viết lại canonical()/validator trong test.

**Cập nhật CI (2026-09-08):** hai smoke trên đã được wired vào job
`deterministic-runtime` của `curriculum-ci.yml`, chạy cho pull request và push
vào `main`. Đây là regression CI cho shared M00–M11 fixture lineage và M11
backup/reconciliation/restore; chưa phải bằng chứng GitHub Actions ở head cho
đến khi remote workflow hoàn tất, và không thay mutation/fault-injection bên
dưới.

**Cập nhật walkthrough giữ artifact (2026-09-10):** BR-16a nhận
`--workspace` chỉ với thư mục trống, không xóa workspace caller-owned và ghi
`walkthrough-result.json` có đường dẫn M08–M11 cùng các kiểm PASS/MATCH/STOP.
CI chạy chính mode này rồi kiểm report và restored history tồn tại. Walkthrough
BR-16b dùng report để learner replay/resolve artifact thật và xác nhận STOP
reject canary; điều này không thay clean-machine pilot hoặc chứng minh người
mới tự hoàn thành không trợ giúp.

**Cập nhật STOP response contract (2026-09-10):** sau khi load durable STOP,
`mission m11-activate` trả envelope `STOPPED` (và exit non-zero) thay vì
`REJECTED`; regression tạo lease/approval thật, STOP rồi thử lại activation.
Runbook kiểm chính lệnh này sau restore. Đây chỉ là contract fail-closed của
CLI, không làm lease fixture hợp lệ hay cấp authority.

**Cập nhật STOP gate boundary (2026-09-10):** `m11-gate` giờ dừng trước resolve
bất cứ supplied ID nào và trước registry append khi runtime đã STOP; nó trả
`STOPPED`, không ghi gate `STOP` mới. Go regression và BR-16a shared smoke giữ
snapshot registry để kiểm no-write. Chỉ reconciliation được phép append vào
runtime STOPPED; thay đổi không mở lại lease hay execution authority.

**Cập nhật STOP lifecycle boundary (2026-09-10):** cùng guard nay áp dụng cho
ledger initialization, authorization/reservation, failed/unknown fixture record,
fixture outcome, evaluation và cycle closure. Mọi lệnh này trả `STOPPED` sau
STOP và regression table kiểm cả registry lẫn outcome store không thay đổi;
reconciliation/handoff vẫn là đường recovery được giới hạn. Đây không chứng
minh recovery tự động hay business outcome.

**Cập nhật STOP admission boundary (2026-09-10):** `m11-recovery-admit` kiểm
runtime mới trước khi đọc handoff/admission input, trả `STOPPED` nếu đích đã
dừng và không tạo admission artifact. Regression dùng input paths không tồn tại
để chứng minh STOP preempts input access. Điều này không thay thế review cho
runtime/lease mới đang hoạt động.

**Cập nhật race CI (2026-09-08):** `deterministic-runtime` chạy thêm
`go test -race ./...` cho learner Bot. Local race suite PASS. Race detector là
phủ trợ cho smoke multi-process, không chứng minh transaction đa-file hoặc
thay thế barrier/fault hook quyết định.

**Cập nhật engine CI (2026-09-09):** job `n8n-engine-regression` trong
`mission-agent-path-ci.yml` cài n8n `2.38.1` vào runtime disposable, build
canonical adapter thật và import một bản copy có ID/loopback URL. Vì `n8n
execute` chỉ chấp nhận `Execute Workflow Trigger`, bản copy thay trigger
Schedule/Manual bằng entrypoint này; blueprint review gốc không đổi. Job kiểm
M06 `APPENDED → EXACT_DUPLICATE → replay=MATCH`, sink failure không đến report,
và M07 `POST`, registry bật `follow_redirects`, hoặc adapter không khả dụng đều
dừng ở `Fetch and Register Tool Adapter` trước Agent/proposal persistence.
Không cài credential, không gọi affiliate/provider và không coi đó là M07
model-success hoặc received-redirect transport coverage/evidence business.

**Mở rộng M06 engine CI (2026-09-09):** runner còn chạy key-order retry,
same-correlation content conflict, unsupported source và changed event. Hai ca
reject phải dừng ở adapter trước ACK/report và không được thay đổi bytes history;
changed event mới phải persist record riêng và replay `MATCH`. Phạm vi vẫn là
fixture synthetic; schedule admission và selected-source profile không được suy
ra là đã nghiệm thu.

**Mở rộng M07 model-stub CI (2026-09-09):** cùng job import credential disposable
chỉ trỏ endpoint OpenAI-compatible loopback, thay network fetch ở bản copy test
bằng fixture `register-tool-result`, rồi giữ Agent/context/grounding/proposal
nodes của blueprint. Stub phải nhận canonical context và tool evidence; output
được validate, persist, rồi revalidate sau restart adapter. Điều này không đưa
secret/provider vào CI và không là bằng chứng received-redirect hoặc provider
diversity.

**M07 received-redirect regression (2026-09-09):** transport helper có seam
`RoundTripper` chỉ dùng test; wrapper production vẫn tạo transport guard của
chính nó. Response `302 Location` có kiểm soát phải dừng tại adapter với
`TOOL_TRANSPORT_REJECTED`, không follow request thứ hai và không persist trace.
Test chạy HTTP adapter thật cùng core transport logic, không cần endpoint public
hoặc fixture loopback bị policy cấm. Nó không thay operated evidence cho một
redirect source live.
- Test vận hành HTTP bằng loopback; policy transport test không gọi internet/provider. Pinned HTTPS smoke hiện có giữ profile nguồn đã ghim, phân biệt lỗi network với guard reject.
- Cross-process tests có barrier/fault hook và timeout hữu hạn; không trông chờ xác suất race hoặc sleep dài. `go test -race` bổ sung, không thay test nhiều process.
- Mutation proof trong checkout tạm: bỏ expiry gate, bỏ lock, cho overwrite input, tự thêm ID hoặc skip nested bundle phải làm đúng test/job fail; restore checkout tạm sau test, không sửa worktree người dùng.
- Job logs ghi commit, case IDs, expected/actual, exit status và artifact refs đã loại secret. Test ref tồn tại không đủ: phải có run/head evidence.

**Nghiệm thu:** R15 đóng; không tắt check để làm CI xanh. Thay branch protection/required-check setting cần phê duyệt quản trị riêng, không ngầm nằm trong sửa workflow.

### RP-09 — Audit readiness và nghiệm thu offline

**Cập nhật audit (2026-09-08):** `audit_readiness.py` đọc matrix thay vì tìm từ
khóa đơn lẻ: kiểm version/IDs/status, implementation/test refs tồn tại, remaining
evidence bắt buộc cho non-final status, `IMPLEMENTED` không được còn gap, và hai
smoke M00–M11/M11 restore đã wired trong CI. Nó cũng reject claim
`ready for production` không có phủ định khi matrix giữ NOT_READY. Năm isolated
negative fixtures chạy implementation thực cho ref hỏng, status final còn gap,
CI regression mất, và prose overclaim. Audit vẫn chưa parse toàn bộ ngữ nghĩa
mọi tài liệu/PR hoặc xác nhận remote CI run; các phần đó còn mở.

**Cập nhật evidence graph (2026-09-09):**
`READINESS-EVIDENCE-GRAPH.json` chia mỗi BR thành các claim implementation,
test và operated/external, với scope bắt buộc. Claim implementation/test link
ngược vào field refs của matrix; mọi claim link marker trong plan, còn test đã
`VERIFIED_OFFLINE` phải chỉ đúng command trong workflow CI. Audit từ chối graph
thiếu partition, ID trùng, marker/CI command không resolve, hoặc criterion
`IMPLEMENTED` còn claim `PARTIAL`/`MISSING`. Negative fixtures chạy audit thật
trên isolated copy. Graph chỉ là evidence có trong checkout; nó không thay API
GitHub, remote CI run hay independent/external evidence.

- Matrix mở rộng tiêu chí theo từng gap và phân loại `implementation_gaps`, `test_gaps`, `external_evidence_gaps`; implementation/test/evidence refs có scope/version/commit và trạng thái rõ. Migrate version của schema/audit cùng lúc.
- Tự sinh hoặc kiểm bảng BR từ matrix. Audit phát hiện thiếu R01–R16 mapping, ref hỏng, thiếu evidence của claim đã đóng, status mâu thuẫn và prose đang tuyên bố cao hơn trạng thái được chấp nhận.
- Có negative fixtures của chính audit: empty refs, stale/incorrect commit, implemented nhưng thiếu regression, plan cao hơn matrix, toàn offline nhưng claim production. Không cố định một câu output rồi gọi là readiness computation.
- Reviewer đối chiếu các PR, rerun suite/cases theo baseline/head, ghi remaining gaps cụ thể. Current planning PR chỉ sửa thông tin sai; R16 chưa đóng cho tới khi audit mới bắt được chúng.

**Nghiệm thu:** R16 đóng ở mức cơ chế kiểm chứng; giữ overall NOT_READY_FOR_PRODUCTION cho tới khi có evidence ngoài repo và authority cần thiết. Không biến VERIFIED_OFFLINE thành learner/pilot PASS.

### RP-10 — Chỉ thực hiện sau khi đóng các code/test gaps

- Chọn phiên bản n8n hỗ trợ, topology kết nối canonical adapter, host/nguồn read-only, provider/model và budget với chủ repo. Không truy cập loopback của host khác qua cấu hình mặc định; có hướng dẫn container/host đúng topology.
- Import và chạy blueprint thật; lưu export workflow/version, execution ID, tool/store ACK, failure cases và provenance. Provider credentials/secret không commit.
- Pilot máy sạch: học viên làm theo walkthrough, ghi trợ giúp, lỗi, kết quả và khả năng giải thích boundaries; có trợ giúp thiết kế/code thì không gọi self-service PASS.
- Deployment drill trên target được cấp quyền: build/start/health/status/logs/stop, resource limits, backup/restore, restart/expiry/STOP, human-reviewed recovery. Không suy power-loss/24×7 proof từ smoke local vài giây.
- Ghi riêng operated evidence và business evidence; chưa chọn affiliate platform/channel hoặc live executor thì để phần đó OPEN, không chặn sửa offline đã được xác định.

## 5. Cách kiểm và giao việc cho từng PR

Các lệnh baseline dưới đây là lệnh đã có trong repo; các negative cases trong từng RP là công việc bổ sung. Chạy từ repo root; dùng Go theo `go.mod`, Python theo CI, quyền loopback cho tests liên quan. Không cần provider credential cho offline suite.

```bash
for module in contracts core lab/affiliate-bot lab/mission-runtime; do
  (cd "$module" && GOWORK=off go test -count=1 ./... && GOWORK=off go vet ./...) || exit 1
done
python3 -m unittest discover -s scripts/tests -v
for validator in scripts/validate*.py; do
  python3 "$validator" || exit 1
done
python3 scripts/audit_readiness.py
python3 scripts/smoke_br12d.py
python3 scripts/smoke_br13b.py
python3 scripts/smoke_br16a_offline.py
python3 scripts/smoke_br18b_backup_restore.py
git diff --check
```

Quickstart clone/cache trống và pinned HTTPS fetch cần mạng: `python3 scripts/smoke_quickstart.py`, `python3 scripts/smoke_br13c.py`. Chạy cho đợt nghiệm thu tổng và khi sửa phạm vi liên quan. Guard offline phải kiểm bằng transport fixture, không suy test reject PASS từ lỗi DNS/sandbox.

Mỗi PR implementation phải ghi:

- RP/R IDs và BR affected; owner/reviewer; base/head SHA; scope ngoài phạm vi.
- Case ID, fixture version/path, expected/actual; bằng chứng trước fix FAIL và sau fix PASS trên cùng implementation được learner/workflow dùng. Không yêu cầu merge test đỏ vào main.
- Changed schemas/APIs/stores và compatibility/migration/rollback; docs/evals/checkpoints liên quan.
- Local test kết quả riêng với GitHub Actions kết quả; run URL/head cho CI. Không suy CI PASS từ local PASS.
- Remaining implementation/test/external gaps; không đóng R chỉ vì thêm checklist, có file test hay có một smoke happy path.

## 6. Điều kiện dừng an toàn

Conflict với dữ liệu/người dùng, thay đổi schema không có migration rõ, không bảo đảm transaction, chưa có quyền live, chưa có host/provider/budget hoặc reviewer chưa chấp thuận recovery: dừng phần phụ thuộc và ghi blocker. Tiếp tục các phần độc lập đã được cấp phạm vi; không tự mở rộng sang live hoặc reset ledger.

Chỉ đề xuất công bố “đường offline liên tục” sau RP-01…RP-09 đạt nghiệm thu; chỉ đề xuất “người mới tự làm được” sau pilot RP-10 có bằng chứng. Không có cam kết production hoặc business outcome trong kế hoạch này.
