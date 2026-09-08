# Kế hoạch sửa sau review toàn repo tại ece6a32

- Mã: RR-2026-09-07; phiên bản kế hoạch: 2.
- Ngày lập kế hoạch: 08/09/2026; mã kế hoạch theo ngày review baseline.
- Baseline: `ece6a32619e5b9a05d0599b87f50023f38931cb9`.
- Trạng thái: **PROPOSED** — PR này chỉ đề xuất kế hoạch và đính chính trạng thái; chưa triển khai các sửa lỗi R01–R16.
- Cơ sở: [sổ phát hiện và bằng chứng baseline](evidence/REVIEW-ECE6A32.md).
- Liên kết kế hoạch gốc: [BR-2026-09](BEGINNER-READINESS-PLAN.md); trạng thái tích hợp BR: [readiness matrix](READINESS-MATRIX.json).
- Người lập kế hoạch: Codex theo yêu cầu chủ repo. Người triển khai/reviewer từng đợt: chưa phân công; điền khi nhận việc. Không mặc định chủ repo đã nghiệm thu kế hoạch hoặc cho phép live operation.

Cập nhật sau review PR #94: bổ sung dependency RP-05 cho RP-06, tách nghiệm thu
restore M00–M10 khỏi phần mở rộng M11 bắt buộc ở RP-07a, và giao rõ việc tạo/lưu/
kiểm chuỗi cost-bound, gate, authorization, execution và EffectRef. Đây vẫn là
sửa kế hoạch; các implementation gaps R01–R16 chưa được đóng.

## 1. Mục tiêu và giới hạn

Sửa những đường chạy đã tái hiện sai; thống nhất canonical implementation giữa learner Bot, harness và n8n; chứng minh một chuỗi artifact M00–M11 bằng runtime thật trong môi trường offline. Sau đó mới tổ chức pilot và kiểm chứng n8n/host/provider được chọn.

PR kế hoạch này không sửa Go/Python/blueprint/schema, không đổi required checks/branch protection, không chạy live executor, không tạo tài khoản, không mua host hoặc dùng API trả phí. Các PR triển khai tiếp theo cần yêu cầu thực hiện riêng; việc merge kế hoạch không tự khởi động chúng.

Không đổi `PROGRESS.md` hoặc công nhận Mission PASS cho học viên. `execution_permitted=false`, evidence synthetic và trạng thái `NOT_READY_FOR_PRODUCTION` vẫn được giữ. Hoàn tất offline không chứng minh business outcome, human pilot hoặc production readiness.

Các từ trạng thái trong tài liệu:

- `PROPOSED`: kế hoạch đang xin review; chưa phải implementation.
- Công việc RP: `TODO → IN_PROGRESS → IN_REVIEW → VERIFIED_OFFLINE`; `BLOCKED` phải kèm blocker cụ thể.
- R01–R16 mặc định OPEN; chỉ đóng khi có diff triển khai, regression trước/sau và reviewer xác nhận đúng baseline/head.
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

Tên package, endpoint và test file mới bên dưới là **đề xuất**, không phải implementation refs đã tồn tại. PR triển khai chốt versioned API/contract và cập nhật curriculum/mission/starter/checkpoint/eval bị ảnh hưởng cùng lúc.

## 3. Bao phủ 16 phát hiện

Tất cả hàng dưới đây còn OPEN. “Kiểm chứng bắt buộc” là test phải viết/chạy khi sửa, không phải test đã PASS.

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

RP-00 là PR kế hoạch hiện tại; các mã RP khác chưa phải số PR GitHub đã tạo. Chỉ RP-00 đang xin review. Mỗi RP khi bắt đầu phải thêm người làm, reviewer, URL PR, baseline/head và evidence vào mục tương ứng.

| Gói | Phạm vi | Phụ thuộc trước khi merge | Quy mô | Trạng thái |
|---|---|---|---|---|
| RP-00 | Lưu kế hoạch/baseline, hạ tuyên bố quá mức | — | S | MERGED (`main` `12a088a`) |
| RP-01 | Bảo vệ đường dẫn và file đầu vào | RP-00 | S | IN PROGRESS — implementation trên `codex/rp-01-path-safety`, chưa merge |
| RP-02 | Shared M08 decoder/policy, exact-number/hash contract | RP-01 | M | IN PROGRESS — implementation trên `codex/rp-01-path-safety`, chưa merge |
| RP-03 | Shared M09/M10 guard, cost-bound/gate/authorization/execution, ledger và STOP | RP-02 | L; chia 03a/03b | IN PROGRESS — chỉ foundation/registry/reservation, chưa đủ graph canonical |
| RP-04 | Canonical M06 builder và resolver M07/M08/HTTP | RP-01; tích hợp M08 sau RP-02 | M | IN PROGRESS — learner history resolver foundation, chưa tích hợp n8n/M06 |
| RP-05 | M07 grounded output và tool-result lifecycle | RP-04 | L; chia 05a/05b | IN PROGRESS — core/learner contract và artifact trace, n8n adapter/proposal store chưa có |
| RP-06 | Snapshot/restore và graph M00–M10, gồm proposal M07 và execution chain | RP-03, RP-04, RP-05 | M | TODO |
| RP-07 | M11 lifecycle + mở rộng restore (07a), rồi full chain/walkthrough (07b) | 07a sau RP-02…RP-06; 07b sau gate lifecycle/restore của 07a | L; chia 07a/07b | TODO |
| RP-08 | CI parity/mutation/cross-process coverage | Bắt đầu cùng RP-01; đóng sau RP-07 | M, xuyên các PR | TODO |
| RP-09 | Readiness audit có dữ liệu/evidence, chốt offline acceptance | RP-06, RP-07, RP-08 | M | TODO |
| RP-10 | n8n operated run, pilot máy sạch, deployment drill | RP-09 và lựa chọn môi trường/quyền cần thiết | M/L | TODO |

Luồng ưu tiên: RP-01 → RP-02 → RP-03; RP-04 có thể làm song song trên file độc lập. RP-06 chỉ merge sau RP-03/RP-04/RP-05 để kiểm proposal đã persist và execution chain thật. RP-06 nghiệm thu inventory M00–M10; RP-07a bổ sung artifact M11 và phải mở rộng manifest/loader/restore tests trong cùng gói, rồi RP-07b mới nghiệm thu toàn chuỗi. Không thêm dependency RP-07 ngược vào RP-06 gây vòng lặp. RP-08 đưa test vào từng PR, không đợi cuối dự án mới bật gate. Không đặt ngày production trước khi chốt điều kiện RP-10.

### Cập nhật triển khai — 2026-09-08

Các thay đổi dưới đây nằm trên branch `codex/rp-01-path-safety` (head
`df16783` khi ghi mục này), chưa có PR/merge nên **không thay đổi trạng thái
nghiệm thu trên `main`**.

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

**Cập nhật thực hiện (2026-09-08, phạm vi hẹp):** learner backup đã lên
`affiliate-bot-backup/v2` và ghi inventory bắt buộc. Runtime có canary phải có
M10 registry; runtime có `FAILED` execution phải có fixture outcome store.
Restore gọi canonical loader để kiểm reservation → execution và FAILED
execution → `MACHINE_EXECUTION` EffectRef trước khi trả `RESTORED`. Smoke tạo
governed M10 chain bằng Bot, kiểm manifest thiếu artifact, checksum-hợp-lệ
nhưng orphan outcome, replay/resolve sau process mới, budget và STOP. Đây chỉ
là một lát cắt RP-06; các hạng mục còn lại bên dưới vẫn mở.

- Định nghĩa inventory M00–M10: history, human action/outcome, evaluation, improvement/review, AgentProposal đã persist từ RP-05, intent/policy/per-action approval, canary grant/grant approval, TrustedCostBound, gate decision, ExecutionAuthorization, ExecutionRecord, machine outcome/EffectRef, reservation/pre-post ledger/consumed markers/STOP, và nested advisor bundle. Các artifact từ RP-03/RP-05 phải được tạo qua runtime trong test snapshot, không dựng file placeholder. Khai báo store ngoài runtime root; không tự gom secret hoặc gọi toàn bộ home là backup.
- Manifest dùng relative paths chuẩn hóa và checksum/size/type/version. Backup đệ quy theo inventory; layout chưa hỗ trợ phải reject rõ thay vì skip thư mục. Reject symlink/path traversal và đích backup nằm trong source gây self-inclusion.
- Quiesce writer/lock snapshot hoặc dùng snapshot transaction. Copy bytes nhất quán với ledger/state/STOP; chỉ publish manifest sau snapshot hoàn tất.
- Restore vào staging trống, verify manifest rồi load từng artifact bằng canonical decoder và resolver; replay history, kiểm full graph trong inventory M00–M10, gồm proposal→intent và cost-bound→gate→authorization→execution→outcome/EffectRef, usage, consumed approval, expiry và STOP. Không emit RESTORED trước các gate.
- Test bundle từ `advisor fixture-run`, AgentProposal từ RP-05, execution chain EC-01 từ RP-03; orphan từng loại artifact, corrupted payload có checksum đúng do backup create, expired approval, missing ledger, nonempty target, interrupted copy và concurrent writer. Sau restore chạy lại EC-02…EC-05 bằng process mới, kiểm list/replay/status/reserve/resolve, không chỉ so JSON field. Hết hạn tại thời điểm restore không làm mất historical artifact hợp lệ: vẫn đọc/replay được, nhưng operation mới phải bị chặn; không gia hạn hoặc reissue quyền.
- Profile manifest phải nêu rõ phạm vi/version. Trước phần mở rộng RP-07a, gặp artifact M11 chưa được hỗ trợ thì reject tường minh; không skip hoặc báo restore toàn M00–M11. Không tạo lease/activation placeholder để đạt test.

**Nghiệm thu RP-06:** snapshot/restore M00–M10 PASS và giữ budget/STOP, proposal/authorization/execution/EffectRef resolve đúng. R12/R13 được ghi evidence cho phạm vi này nhưng **chưa đóng toàn phạm vi M00–M11**; chỉ đóng sau gate mở rộng restore RP-07a dưới đây. **Rollback:** giữ source/backup nguyên trạng; staging lỗi không trở thành active runtime; người vận hành quyết định cleanup/recovery, không tự ghi đè đích cũ.

### RP-07a/07b — M11 lifecycle và chain learner đầy đủ

**07a:** tách/reuse M11 lease activation, health gate, ledger/reconciliation, STOP, reviewed recovery từ harness; thêm learner entrypoint và store links. Offline executor stub không được gọi là live execution. Persist lifecycle artifact và version, kiểm time/authority/unknown usage/cycle closure; không chỉ thêm STOP/status alias.

RP-07a sở hữu graph M11: source canary/promotion review → lease + lease approval → activation; intent/policy + health snapshot + TrustedCostBound + pre-ledger → ProductionGateDecision → reservation/ExecutionAuthorization → ExecutionRecord → outcome có `MACHINE_EXECUTION` EffectRef → evaluation/cycle/post-ledger; trường hợp UNKNOWN/STOP có reconciliation resolution và reviewed recovery riêng. Tái sử dụng [M11 chain checker](../../lab/mission-runtime/cmd/demo/m11_chain.go), [production gate](../../contracts/production-gate-decision.schema.json) và canonical schemas; không bỏ gate/authorization/execution chỉ vì chạy offline.

**Gate restore M11 bắt buộc trước khi đóng RP-07a:**

- Mở rộng inventory/manifest version, decoder, graph resolver và snapshot/restore của RP-06 trong cùng gói RP-07a. Bao phủ lease/lease approval/promotion refs, activation, health, cost bound, gate, authorization, execution, pre/post/STOP ledgers, outcome/evaluation/cycle và reconciliation/recovery artifacts; giữ nguyên các artifact M00–M10.
- Sinh artifact bằng learner + executor stub, snapshot, restore vào đích trống rồi process mới replay/resolve đầy đủ. Kiểm `closed_cycle` và `resolved_stop` theo harness; STOP vẫn chặn, expiry không mở quyền mới, usage/consumed markers không reset. Mở rộng EC-01…EC-05 cho production profile.
- Negative cases: thiếu hoặc sai lease/activation/health/cost/gate/authorization/execution/EffectRef/cycle/resolution; dữ liệu bị sửa nhưng checksum backup hợp lệ; phiên bản manifest không hỗ trợ. Reject graph hỏng, không publish runtime sẵn dùng. Lịch sử đã hết hạn vẫn phục hồi được ở chế độ không cấp quyền; recovery chưa review phải bị chặn.
- Chỉ ghi đóng R12/R13 toàn phạm vi khi cả evidence RP-06 và gate restore M11 này PASS trên head tương thích. RP-07b và RP-09 không được nghiệm thu full chain/readiness nếu gate này chưa đạt. Dependency là RP-06 → RP-07a → RP-07b, không có vòng lặp.

**07b:** thay smoke BR-16a bằng một workspace chung và cùng evidence/decision lineage:

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
- Test vận hành HTTP bằng loopback; policy transport test không gọi internet/provider. Pinned HTTPS smoke hiện có giữ profile nguồn đã ghim, phân biệt lỗi network với guard reject.
- Cross-process tests có barrier/fault hook và timeout hữu hạn; không trông chờ xác suất race hoặc sleep dài. `go test -race` bổ sung, không thay test nhiều process.
- Mutation proof trong checkout tạm: bỏ expiry gate, bỏ lock, cho overwrite input, tự thêm ID hoặc skip nested bundle phải làm đúng test/job fail; restore checkout tạm sau test, không sửa worktree người dùng.
- Job logs ghi commit, case IDs, expected/actual, exit status và artifact refs đã loại secret. Test ref tồn tại không đủ: phải có run/head evidence.

**Nghiệm thu:** R15 đóng; không tắt check để làm CI xanh. Thay branch protection/required-check setting cần phê duyệt quản trị riêng, không ngầm nằm trong sửa workflow.

### RP-09 — Audit readiness và nghiệm thu offline

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
