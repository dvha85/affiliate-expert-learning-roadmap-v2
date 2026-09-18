# Kế hoạch sửa sau review toàn repo tại ece6a32

<!-- readiness-as-of: 2026-09-18 -->
<!-- readiness-main-baseline: af2eb0d20d10d8f638480b8cdd8141c8b051ec1c -->

**Baseline sync after PR #435 (2026-09-18):** PR #435 was squash-merged into
`main` at `af2eb0d20d10d8f638480b8cdd8141c8b051ec1c`, from implementation head
`f398c6e5782803eed494fd9a5a609c4aec1320ca`. Curriculum CI run
`35359232122` and Mission Agent Path CI run `35359232042` both passed all 13
checks, including Windows runtime, learner race, full Go test/vet and the
backup/restore mutation smoke. The post-merge RP-07b snapshot records explicit
assertions for a restored `CLOSED` M11 cycle with matching execution,
outcome/evaluation/cycle lineage and a `NORMAL` post-ledger, plus the reviewed
UNKNOWN-to-STOP reconciliation boundary. This remains bounded
offline/fixture/read-only evidence; it does not claim prior-runtime
availability, live recovery, power-loss/filesystem-crash durability, atomic
multi-file, distributed/multi-host, provider/live execution, deployment,
business outcomes or clean-machine/target-host readiness. Record:
`docs/architecture/EVIDENCE-RP07B-M11-RESTORE-CHAIN-20260918.md`. Marker:
`Baseline sync after PR #435`.

**Baseline sync after PR #433 (2026-09-18):** PR #433 was squash-merged into
`main` at `c8fe14218ef2a22f6faf0cbd1c9a335c1073dffb`, from implementation head
`e406e405b43c954353c6fe58c21f096f28a5ea1f`. Hosted Curriculum CI run
`35350741433` and Mission Agent Path CI run `35350741436` both passed all 13
checks, including Windows runtime `105618004234`, learner race `105618003984`
and the new backup/restore regression. The post-merge RP-07a snapshot records
that backup/restore v3 preserves a valid M11 `ProductionRecoveryAdmission` and
fails closed when a checksum-valid registry is missing the immutable new-runtime
approval, without publishing a restore target. This remains bounded
offline/fixture/read-only evidence; it does not claim prior-runtime availability,
live recovery, power-loss/filesystem-crash durability, atomic multi-file,
distributed/multi-host, provider/live execution, deployment or business
outcomes. Record:
`docs/architecture/EVIDENCE-RP07A-M11-ADMISSION-RESTORE-20260918.md`. Marker:
`Baseline sync after PR #433`.

**Baseline sync after PR #429 (2026-09-18):** PR #429 đã squash-merge vào
`main` tại `4950b2fd1be80f3c0e58d9f270a44e275c3de0d5`, từ implementation head
`d079dfc3b2d229805b458ec95c92be0fd15b375f`. Hosted Curriculum CI run
`35340989562` và Mission Agent Path CI run `35340989656` đều PASS đủ 13 checks,
bao gồm Windows runtime, learner race, M07 adversarial/output contracts,
mission-runtime và n8n regression. Post-merge snapshot cập nhật RP-05b với
request identity adapter-derived, canonical JSON body digest và fail-closed cho
forged/legacy provenance; readiness vẫn `NOT_READY_FOR_PRODUCTION`, không thêm
provider/live execution, deployment, business outcome, distributed locking,
power-loss hay multi-file atomicity claim. Record:
`docs/architecture/EVIDENCE-PR429-POST-MERGE-20260918.md`. Marker:
`Baseline sync after PR #429`.

**Baseline sync after PR #431 (2026-09-18):** PR #431 đã squash-merge vào
`main` tại `ad5347bf1aae4c0770e0471b668c94bf8621b141`, từ implementation head
`321eb74b91e991ab319c784205d3f587ef0b01bc`. Hosted Curriculum CI run
`35344493952` và Mission Agent Path CI run `35344493969` đều PASS đủ 13 checks,
bao gồm Windows runtime, learner race và backup/restore mutation smoke. RP-06
giờ validate mọi M10 registry hiện hữu ngay cả khi mutable state không còn active
canary, vẫn restore được registry lịch sử hợp lệ và reject orphan registry trước
publish. Evidence vẫn bounded offline/fixture/read-only; không claim power-loss,
atomic multi-file, distributed/multi-host, provider, live execution, deployment
hay business outcome. Record:
`docs/architecture/EVIDENCE-PR431-POST-MERGE-20260918.md`. Marker:
`Baseline sync after PR #431`.

**Baseline sync after PR #425 (2026-09-18):** PR #425 đã squash-merge vào
`main` tại `b7cfe1480973254ad106481d22c142651af3732b`, từ implementation head
`d3127b99dda310be3620ff54ea0b3add0cd6d44b`. Hosted Curriculum CI run
`35329168897` và Mission Agent Path CI run `35329168855` đều PASS đủ 13 checks,
bao gồm Windows runtime, learner race, Go vet và n8n/mission regressions. Hậu-
merge snapshot chỉ rebinding RP-04 shared canonical context về merge baseline;
readiness vẫn `NOT_READY_FOR_PRODUCTION`, không thêm provider/live execution,
deployment, business outcome, distributed locking, power-loss hay multi-file
atomicity claim. Record:
`docs/architecture/EVIDENCE-PR425-POST-MERGE-20260918.md`. Marker:
`Baseline sync after PR #425`.

> Reconcile 18/09/2026: đây là tracker hiện tại của `main` tại baseline trên.
> Xem [kế hoạch pre-merge tại 737e85a](PRE-MERGE-REMEDIATION-737E85A.md) cho
> phát hiện PMR-01…07 ban đầu. Các ghi chú cũ chỉ có giá trị lịch sử; matrix và
> bảng gói dưới đây là nguồn trạng thái hiện hành.

**Baseline sync after PR #413 (2026-09-18):** PR #413 đã squash-merge vào
`main` tại `bdbffb51643f0922e0dcf616f92da26b98f37faf`, đưa bản cập nhật
post-merge evidence của PR #412 vào snapshot hiện hành. Hậu-merge có đủ 12/12
check-runs PASS; readiness audit vẫn `NOT_READY_FOR_PRODUCTION` và không có
runtime/provider/pilot/deployment claim mới. Đây là bookkeeping/evidence sync,
không đóng các blocker external hoặc mở rộng phạm vi crash/power-loss,
distributed locking và business outcome. Marker: `Baseline sync after PR #413`.

**Baseline sync after portability fix (2026-09-18):** commit
`3eb1a9670a7e5ff005427d1af7d49c549953d926` sửa stable append trên macOS để
canonicalize alias hệ thống `/var` → `/private/var` trước descriptor traversal,
trong khi vẫn từ chối symlink do caller kiểm soát. Regression parent-swap chạy
PASS với `TMPDIR` mặc định; learner race/vet, các smoke/readiness local và
Windows cross-compile cũng PASS. Đây là hardening pathname và test portability
có phạm vi hẹp; không đóng Windows native lock/path parity, crash/power-loss,
distributed locking, provider, live executor, business outcome, pilot hoặc
deployment. Record: `docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-3EB1A96-20260918.md`.

**Baseline sync after Windows native lock hardening (2026-09-18):** commit
`598bb21801d44d9f2a1bf5ed56fbcb522aa2c0e9` replaces the Windows cooperative
directory claim with an exclusive native file handle that the kernel releases
on process exit, and opens stable append targets with
`FILE_FLAG_OPEN_REPARSE_POINT`. GitHub CI completed 12/12 PASS; local learner
race/vet and Windows amd64 cross-compile also PASS. Windows runtime execution,
full ancestor-race parity, distributed locking, crash/power-loss, provider,
live executor, business outcome, pilot and deployment evidence remain open.
Record: `docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-598BB21-20260918.md`.

**Cập nhật RP-01 Windows runtime CI (2026-09-18):** PR #415 bổ sung job
`windows-runtime` trên `windows-latest` chạy thật `go test ./...`, `go vet ./...`
và nhóm regression managed lock/backup/restore. Các run đầu tiên đã bắt được
directory sync không tương thích Windows, stable reader chưa share-delete,
append bắt đầu ở offset 0 và kiểm tra incomplete outcome path bị bỏ qua; các
seam này đã được sửa trong các commit `822cd83`, `f043d3f` và `dad7d4c`.
Curriculum CI run `35302232344`, head `dad7d4c`, PASS toàn bộ; riêng job
`windows-runtime` PASS trong 2m48s. Đây là runtime filesystem một máy trên
hosted Windows, không đóng full ancestor-race parity, distributed locking,
crash/power-loss, provider, live executor, business outcome, pilot hoặc
deployment. RP-01 vẫn `PARTIAL`, overall vẫn `NOT_READY_FOR_PRODUCTION`.
Record: `docs/architecture/EVIDENCE-RP01-WINDOWS-RUNTIME-CI-20260918.md`.

**Cập nhật RP-01 Windows ancestor-race hardening (2026-09-18):** PR #416 tại
head `bbd8aec14e56ba0b484376014b2ab5ec775ce3ec` rà lại các Windows reader/
writer còn đi qua pathname trực tiếp và thêm native parent-chain pinning cho
backup/restore output, managed lock, stable reader/append và internal/store
JSONL reader. Existing ancestors được mở với `OPEN_REPARSE_POINT`, không share
delete; component còn thiếu được tạo và pin từng bước. Regression Windows thay
ancestor sau preflight bằng junction/reparse point cho các đường backup/restore,
stable reader/append và store reader đều fail closed, không đổi external tree.
Curriculum CI run `35306191800` PASS; `windows-runtime` job
`105478750919` PASS với `go test ./...`, `go vet ./...` và targeted
lock/backup/restore; learner race job `105478751048` PASS với
`go test -race ./...`. Mission Agent Path CI run `35306191794` cũng PASS cả
ba job. Vì Windows không có portable `openat`/`mkdirat` cho
traversal nhiều component trong boundary này, implementation chỉ claim
conservative local single-host boundary, không claim POSIX parity, distributed
locking, power-loss, multi-file atomicity, provider/live execution hoặc
business outcome. RP-01 vẫn `PARTIAL`, overall vẫn `NOT_READY_FOR_PRODUCTION`.
Record: `docs/architecture/EVIDENCE-RP01-WINDOWS-ANCESTOR-RACE-20260918.md`.

**Baseline sync after PR #417 (2026-09-18):** PR #417 đã squash-merge vào
`main` tại `c3d10f8`, từ implementation head `6e8ad1a`. PR có 13/13 checks
PASS, gồm Curriculum CI với Windows runtime và learner race, cùng Mission
Agent Path CI. Snapshot đã rebinding plan, readiness matrix và evidence graph
theo baseline RP-02 mới; record: `docs/architecture/EVIDENCE-PR417-POST-MERGE-20260918.md`.
Đây là post-merge bookkeeping cho bounded offline/fixture decoder, policy và
provenance evidence; hash migration, cross-store persistence, provider/live
execution và production readiness vẫn mở. Marker: `Baseline sync after PR #417`.

**Cập nhật RP-03 canonical M10 grant registry binding (2026-09-18):** trước
`m10-gate` hoặc `m10-reserve`, learner state phải resolve exact immutable
`CANARY_GRANT` từ M10 registry; grant hợp lệ về schema nhưng chỉ xuất hiện
trong mutable `mission-state.json` bị fail closed trước gate/ledger mutation.
Regression dùng real Bot fixture, đổi sang grant mới có hash hợp lệ nhưng không
đăng ký, rồi kiểm `m10-gate` và `m10-reserve` không tạo output, registry entry,
reservation hay usage-counter change. Worktree không có Go executable nên hosted
`go test -race ./...` vẫn là acceptance gate. Đây là local offline/read-only
lineage guard; không đóng multi-file crash/power-loss, distributed locking,
provider, live executor, business outcome, pilot hay deployment. Marker:
`Cập nhật RP-03 canonical M10 grant registry binding`.

**Baseline sync after PR #416 (2026-09-18):** PR #416 đã squash-merge vào
`main` tại `cefdb758f70ce36c08bc822ee658737149600bcd`, từ implementation head
`bbd8aec14e56ba0b484376014b2ab5ec775ce3ec`. Snapshot hiện tại đã đồng bộ plan,
readiness matrix và evidence graph về merge baseline này; CI evidence được giữ
theo các run đã review của PR gồm Windows runtime, targeted lock/backup/restore,
learner race và Mission Agent Path. RP-01 vẫn `PARTIAL`, overall vẫn
`NOT_READY_FOR_PRODUCTION`; record này không thêm provider, live-executor,
business-outcome, pilot, deployment, distributed-locking, power-loss hay
multi-file atomicity claim. Record:
`docs/architecture/EVIDENCE-PR416-POST-MERGE-20260918.md`. Marker:
`Baseline sync after PR #416`.

**Baseline sync after PR #419 (2026-09-18):** PR #419 đã squash-merge vào
`main` tại `2acfcfc5a80ad590ce06813940ce9bbff83fd839`, từ implementation head
`104bd186b0546dd4c3e7d8c6a3eebc44fe639ecc`. PR có 13/13 checks PASS: Curriculum
CI run `35321402705` gồm Windows runtime job `105524487971` và learner race job
`105524487897`, cùng Mission Agent Path CI run `35321402712` với 3/3 job PASS.
Snapshot đã rebinding plan, readiness matrix và evidence graph về RP-03
canonical M10 grant registry binding sau merge; RP-03 vẫn `PARTIAL`, overall
vẫn `NOT_READY_FOR_PRODUCTION`. Đây là bounded offline/fixture lineage
evidence, không thêm provider, live-executor, business-outcome, pilot,
deployment, distributed-locking, power-loss hay multi-file atomicity claim.
Record: `docs/architecture/EVIDENCE-RP03-M10-CANONICAL-GRANT-REGISTRY-20260918.md`.
Marker: `Baseline sync after PR #419`.

**Cập nhật RP-03 approval expiry và STOP/reserve boundary (2026-09-18):**
learner `m10-reserve` nay có regression trực tiếp cho cả bốn authority expiry:
before boundary phải reserve được, tại/sau boundary phải `REJECTED` và giữ
nguyên canonical runtime snapshot. Regression riêng persist durable M11 STOP
rồi chứng minh M10 reserve trả `STOPPED` không đổi state, counter, reservation
hay registry; synchronized offline smoke tiếp tục cover race thật giữa STOP và
reserve qua các Bot process. Worktree không có `go.exe`, nên hosted race CI là
acceptance gate. Đây là bounded local/offline/synthetic/read-only evidence;
không đóng multi-file crash/power-loss, distributed locking, provider,
live-executor, business outcome, pilot hay deployment. Marker:
`Cập nhật RP-03 approval expiry và STOP/reserve boundary`.

**Baseline sync after PR #421 (2026-09-18):** PR #421 đã squash-merge vào
`main` tại `12c50f3337991b533e5ecf363860df1e27bc2217`, từ implementation head
`3620b24cc3b51d6490473c40103f328c83bc8c20`. PR có 13/13 checks PASS: Curriculum
CI run `35323363318` gồm Windows runtime job `105530687600` và learner race job
`105530687837`, cùng Mission Agent Path CI run `35323363297` với 3/3 job PASS.
Snapshot đã rebinding plan, readiness matrix và evidence graph về regression
RP-03 expiry/STOP-reserve sau merge; RP-03 vẫn `PARTIAL`, overall vẫn
`NOT_READY_FOR_PRODUCTION`. Đây là bounded local/offline/synthetic/read-only
evidence, không thêm multi-file crash/power-loss, distributed locking,
provider, live-executor, business outcome, pilot hay deployment claim. Record:
`docs/architecture/EVIDENCE-RP03-EXPIRY-STOP-RESERVE-20260918.md`. Marker:
`Baseline sync after PR #421`.

**Cập nhật RP-03 exhausted-budget monotonicity (2026-09-18):** R08 nay có
regression trực tiếp sau khi một grant cap=1 đã được tiêu thụ bởi fresh Bot
process. Exact replay của M09 approval và exact re-import của M10 grant chỉ
trả acknowledgement, không reset usage counter. Các payload checksum-valid
nhưng tăng execution cap, đổi currency của cùng grant, hoặc đăng ký cost bound
khác currency đều bị reject; `mission-state.json`, usage/reservation state,
M10 artifact registry và trusted-cost-bound registry giữ byte-identical. Một
reservation mới vẫn nhận `BUDGET_DENIED`. Cùng chuỗi này được chạy lại trên
runtime mới sau backup/restore. Local worktree vẫn thiếu `go.exe`, nên hosted
Windows runtime/race CI là acceptance gate. Đây là bounded
local/offline/synthetic/read-only evidence; không claim multi-file
crash/power-loss, distributed locking, provider, live executor, business
outcome, pilot hay deployment. Marker:
`Cập nhật RP-03 exhausted-budget monotonicity`.

**Baseline sync after PR #423 (2026-09-18):** PR #423 đã squash-merge regression
RP-03 exhausted-budget monotonicity vào `main` tại
`35ebabc2f068cda0a4eef1dd479fc102d03a7a53`, từ implementation head
`425a78a884247756d8c2202997e5075bc04fc3d2`. PR có 13/13 checks PASS:
Curriculum CI run `35326131724` gồm Windows runtime job `105539484848` và
learner race job `105539484967`, cùng Mission Agent Path CI run `35326131948`
với 3/3 job PASS. Post-merge readiness audit và 118 audit unit tests PASS;
snapshot đã rebinding plan, readiness matrix và evidence graph về main.
RP-03 vẫn `PARTIAL`, overall vẫn `NOT_READY_FOR_PRODUCTION`. Đây là bounded
local/offline/synthetic/read-only evidence, không thêm multi-file
crash/power-loss, distributed locking, provider, live executor, business
outcome, pilot hay deployment claim. Record:
`docs/architecture/EVIDENCE-RP03-BUDGET-MONOTONICITY-20260918.md`. Marker:
`Baseline sync after PR #423`.

**Baseline sync after PR #403 (2026-09-17):** PR #403 đã squash-merge vào
`main` tại `8b44011465ea0c8afcfcc832b76f045da0811bd4`, đưa M07 raw-JSON
transport hardening vào learner adapter, blueprint và regression runner. Tất
cả 12 PR checks PASS, gồm Node 24/n8n 2.38.1 engine regression; local
re-verification được ghi tại
`docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-403-20260917.md`. Tracker
vẫn `NOT_READY_FOR_PRODUCTION`; không suy bằng chứng synthetic/read-only này
thành provider, live executor, business outcome, pilot, deployment,
distributed-locking hoặc power-loss/atomic multi-file proof.

**Baseline sync after PR #406 (2026-09-17):** PR #406 đã squash-merge vào
`main` tại `7eacb661d2083d97b42ea31207a3e677705d7fea`, đưa regression
process-kill cho direct M11 STOP vào đường learner Bot thật và đồng bộ
evidence/plan/matrix/audit. Hậu-merge local regression chạy bốn Go module
test/vet, 124 Python tests, các validator standalone, BR-12d, BR-13b, BR-16a,
BR-18b và readiness audit; tất cả phần có đủ input đều PASS, audit vẫn
`NOT_READY_FOR_PRODUCTION`. Node/n8n và các operated validator cần execution
artifact không chạy local. Record:
`docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-406-20260917.md`. Đây vẫn
là evidence fixture/read-only, không đóng provider, live executor, business
outcome, pilot, deployment, distributed locking hay power-loss/atomic
multi-file proof.

**Baseline sync after PR #408 (2026-09-17):** PR #408 đã squash-merge vào
`main` tại `c6debee4b1ef82626eb3e22469886b1f55b45e54`, đưa RP-01 registry
append parent guard vào baseline. Head trước merge `4360858b0d3c04620838fcd94cfd7df8f0f2e287`
đã PASS cả 12 remote PR checks trong runs `35202587298` và
`35202587341`, gồm Go race/vet, learner shards, offline smokes, mission
runtime và Node 24/n8n engine. Worktree này không có Go executable nên không
ghi local Go PASS; Python audit/structural checks được chạy riêng trong
evidence. Tracker vẫn `NOT_READY_FOR_PRODUCTION`; sync này không đóng
provider, live executor, business outcome, clean-machine pilot, deployment,
distributed locking hay power-loss/atomic multi-file proof. Record:
`docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-408-20260917.md`.

**Baseline sync after PR #410 (2026-09-17):** PR #410 đã squash-merge vào
`main` tại `0cf32660166908a4cb5ab6c45ec36f2a174e64be`, đưa hardening parent
guard cho inventory campaign/fixture/backup writers của RP-01 vào baseline.
Head PR `d03e4ef4ef2a399f553dfa83bc7d5cc04ea4c2e1` đã PASS 12/12 remote checks
trong Curriculum CI run `35210755354` và Mission Agent Path CI run
`35210755459`, gồm Go race/vet, learner shards, offline smokes, mission
runtime, quickstart và Node 24/n8n engine. Hậu kiểm Python/audit được ghi tại
`docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-410-20260917.md`; worktree
không có Go executable nên không ghi local Go PASS. Tracker vẫn
`NOT_READY_FOR_PRODUCTION`; sync này không đóng provider, live executor,
business outcome, clean-machine pilot, deployment, distributed locking hay
power-loss/atomic multi-file evidence.

**Cập nhật RP-01 advisor campaign writers openat guard (2026-09-17):** inventory
writer cục bộ còn mở đã được đưa về shared stable append boundary: campaign
manifest, reservation, result, offline fixture JSON và backup staging; các path
campaign report/result/canary dùng managed lock. Trên POSIX, helper pin parent
bằng directory descriptor rồi dùng `openat(O_NOFOLLOW|O_EXCL)` cho file mới;
fallback giữ `O_EXCL` nhưng không claim native parity. Regression parent-swap
chạy đường learner Bot thật cho cả năm writer và xác nhận external sentinel/
target không bị ghi. CI `go test -race ./...` là acceptance gate; worktree này
không có Go executable nên không ghi local Go PASS. Đây chỉ là hardening
pathname/writer offline trên một host, không đóng arbitrary writer cooperation,
Windows native lock, power-loss/atomic multi-file, distributed lock, provider,
live executor, business outcome, pilot hay deployment.

**Cập nhật M11 reverse-link ledger graph guard (2026-09-17):** canonical M11
decode reject `ReconciliationResolutionIDs` rỗng/trùng; graph validator yêu cầu
mỗi resolution ID mà ledger đã công bố phải resolve ngược về đúng execution,
lease/version/hash và một stopped head đã hoàn tất review. Khi evaluation đã có
trong inventory, `OutcomeID` của ledger link cũng phải khớp evaluation của cùng
execution. Regression core và loader learner Bot thật dựng các payload checksum
hợp lệ nhưng dangling/mismatched, rồi yêu cầu reject trước runtime use. Đây là
graph integrity offline/read-only; không đóng ledger durability, power-loss,
atomic multi-file, distributed locking, provider, live executor, business
outcome, pilot hay deployment.

**Cập nhật M11 fixture-evaluation evidence cardinality (2026-09-17):** vì
`ProductionOutcomeEvaluation` chỉ là đánh giá offline cho fixture, graph
validator nay yêu cầu `EvidenceIDs` có đúng một phần tử và phần tử đó phải là
`OutcomeID`; schema boundary đã chặn mảng rỗng, còn graph guard chặn payload
schema/checksum hợp lệ có nhiều ID hoặc trỏ sang evidence khác trước
`m11-resolve`/runtime loader. Regression core và learner loader thật tạo rồi
rewrite payload checksum-valid có evidence lệch và đều yêu cầu fail closed.
Phạm vi chỉ là liên kết graph local có thể nghiệm thu; không nâng
`NOT_READY_FOR_PRODUCTION`, không chứng minh live executor, provider, business
outcome, pilot, deployment, distributed locking, power-loss hay atomic
multi-file. Marker: `Cập nhật M11 fixture-evaluation evidence cardinality`.

**Cập nhật M10/M11 registry graph envelope integrity (2026-09-17):** public
`ValidateArtifactGraph` giờ canonicalize lại mọi `ArtifactEntry` bằng
`NewArtifactEntry` trước khi đưa vào lookup map, rồi kiểm exact
`artifact_id`/`content_hash`/bytes. Vì vậy một caller không thể giữ artifact
canonical nhưng thay metadata envelope để làm gate trỏ tới parent giả; learner
loader vẫn giữ lớp kiểm tra JSONL riêng trước graph. Core M10/M11 regression
truyền envelope checksum-valid-artifact nhưng ID/hash giả và yêu cầu reject,
đây là guard offline/read-only ở API graph. Không suy thành proof
power-loss/atomic multi-file, distributed locking, provider, live executor,
business outcome, pilot hay deployment; readiness vẫn
`NOT_READY_FOR_PRODUCTION`. Marker: `Cập nhật M10/M11 registry graph envelope integrity`.

**Cập nhật RP-08 registry graph envelope mutation proof (2026-09-17):** CI
thêm mutation trên disposable copy, lần lượt bỏ graph-only envelope guard của
M10 và M11 rồi chạy đúng core regression thật. Test phải fail với forged
`artifact_id`, `content_hash` hoặc canonical bytes; vì vậy regression không chỉ
được audit bằng source marker. Đây là bằng chứng offline/read-only cho test
được nối với runtime guard, không đóng power-loss, atomic multi-file,
distributed locking, provider, live executor, business outcome, pilot hay
deployment; readiness vẫn `NOT_READY_FOR_PRODUCTION`. Marker: `Cập nhật RP-08
registry graph envelope mutation proof`.

**Cập nhật M11 recovery-admission field integrity (2026-09-17):** canonical
decoder reject các field identity/lineage chỉ chứa whitespace trong
`PRODUCTION_RECOVERY_ADMISSION`, vì JSON Schema `minLength:1` chưa loại được
giá trị rỗng về nghĩa. Regression core thử từng field bắt buộc và yêu cầu reject
trước graph/runtime use; learner vẫn giữ kiểm stopped-runtime, human resolution
và new-lease riêng. Đây là input integrity offline/read-only, không đóng
cross-runtime atomicity, power-loss, distributed locking, provider, live
executor, business outcome, pilot hay deployment; readiness vẫn
`NOT_READY_FOR_PRODUCTION`. Marker: `Cập nhật M11 recovery-admission field integrity`.

**Baseline sync after PR #401 (2026-09-16):** PR #401 đã squash-merge vào
`main` tại `a0f8f19854276589dd9f858ef48869691420294`, bổ sung regression
process-kill cho M10 governed execution sau canonical execution-record append
và reservation binding. Matrix và evidence graph đã được chuyển theo đúng
head này; trạng thái vẫn `NOT_READY_FOR_PRODUCTION`. Bằng chứng là local
POSIX synthetic/read-only, không mở rộng thành proof power-loss, atomic
multi-file, provider, business outcome, pilot hoặc deployment.

**Baseline sync after PR #400 (2026-09-16):** PR #400 đã merge vào `main` tại
`65c70409830b0a789d877bf708a0daec6a2a9be6`, bổ sung process-kill regression
cho M10 canary/cost-bound sau canonical append. Bản đồng bộ này giữ nguyên
`NOT_READY_FOR_PRODUCTION`; local SIGKILL evidence không được suy thành proof
power-loss, atomic multi-file, provider, business outcome, pilot hoặc
deployment.

**Baseline sync after PR #398 (2026-09-16):** PR #398 đã merge vào `main` tại
`1c071ca539de7e77050f43804d91f36884921b58`, bổ sung regression process-kill
cho M11 sau khi ledger transition đã append và directory-sync. Re-run local
trên đúng head này giữ nguyên `NOT_READY_FOR_PRODUCTION`; bằng chứng vẫn chỉ
là fixture/read-only POSIX và không mở rộng thành proof power-loss, atomic
multi-file, provider, business outcome, pilot hoặc deployment. Record:
`docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-398-20260916.md`.

**Baseline sync after PR #397 (2026-09-16):** PR #397 đã merge vào `main` tại
`780bc6bf2ab0774d684f0eb3cc856e6531c32304`, bổ sung process-termination
regression cho M11 outcome journal sau outcome append. Bản đồng bộ này giữ
nguyên phạm vi `PARTIAL`/`NOT_READY_FOR_PRODUCTION`; nó không biến local
SIGKILL evidence thành proof power-loss, atomic multi-file, provider, business
outcome, pilot hoặc deployment.

**Baseline sync after PR #394 (2026-09-16):** PR #394 đã merge thành
`2a2b22b1452dd48a3c8e5e56b151b6933f502d55`, bổ sung mutation proof chạy graph
M11 thật để chứng minh forged authorization-lineage bị bắt. Việc cập nhật
baseline này chỉ sửa liên kết tracker; không nâng trạng thái readiness và
không thay thế bằng chứng CI từ GitHub.

**Baseline sync after PR #395 (2026-09-16):** PR #395 đã merge thành
`aa16839bd6c6d0e2d24ab3b6d7b9000165c720b6`, tách các deterministic smoke thành
ba job CI độc lập để giảm thời gian chờ mà vẫn giữ các boundary M00-M11.
Đây là thay đổi thời gian chạy/độ quan sát của CI; không nâng trạng thái
readiness và không thay thế bằng chứng provider, live executor, business
outcome, pilot hay deployment.

**Baseline sync after PR #392 (2026-09-16):** PR #392 đã merge thành
`74c1719ee6fc90014a9e6599ed0837160e8de6a5`, siết authorization M11 phải giữ
đúng lineage của gate, health snapshot và cost bound đã được gate đánh giá.
Hậu-merge local regression chạy lại bốn Go module, 114 Python tests, BR-16a,
BR-18b và readiness audit; tất cả PASS. Đây vẫn là bằng chứng
fixture/read-only và `NOT_READY_FOR_PRODUCTION`.

**Local regression after PR #392 (2026-09-16):** kết quả chi tiết được ghi tại
`docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-392-20260916.md`.

**Baseline sync after PR #390 (2026-09-16):** PR #390 đã merge thành
`b9890bf4c25ebbb74c09c96b253a6f25c38949e9`. PR chỉ bổ sung evidence và đồng bộ
plan/matrix/graph; không thay đổi runtime/CI implementation. Các kết quả local
ở record post-389 vẫn giữ nguyên phạm vi fixture/read-only và
`NOT_READY_FOR_PRODUCTION`.

- Mã: RR-2026-09-07; phiên bản kế hoạch: 2.
- Ngày lập kế hoạch: 08/09/2026; mã kế hoạch theo ngày review baseline.

**Post-merge regression after PR #387 (2026-09-16):** trên `main`
`58160fa1afee9652efc52fbdaddfc770ea922cc2`, bốn Go module test/vet, 113 Python
tests, BR-16a, BR-18b, readiness audit và 98 audit tests đều PASS; 10/10 PR
checks cũng PASS. Audit vẫn trả `NOT_READY_FOR_PRODUCTION` và resolve 70 scoped
claims. Record:
`docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-387-20260916.md`.

**Local regression after PR #389 (2026-09-16):** trên merged `main`
`fb201c24243dd7bd901849e2bb3cdab15d2b3595`, bốn Go module test/vet, 113 Python
tests, BR-16a, BR-18b, readiness audit, 98 audit tests và `git diff --check` đều
PASS. Disposable n8n `2.38.1` với Node `24.21.0` cũng PASS cho M06/M07 engine
và M06 Schedule Trigger: canonical record được append một lần, retry được
exact-deduplicate trước/sau restart, và adapter unavailable bị fail-closed.
Audit vẫn trả `NOT_READY_FOR_PRODUCTION` và resolve 71 scoped claims. Record:
`docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-389-20260916.md`. Đây vẫn là
fixture/read-only evidence; không đóng provider, live executor, business
outcome, clean-machine pilot, target-host deployment, distributed locking hay
power-loss/atomic multi-file durability.

**Re-run after PR #388 (2026-09-16):** trên merged `main`
`8baae92352d9165affffd45f6dfb85e7ebdfd7b9`, toàn bộ nhóm kiểm tra trên được
chạy lại và PASS; audit vẫn `NOT_READY_FOR_PRODUCTION` và resolve 70 scoped
claims. Đây là lần xác nhận local mới nhất, vẫn chỉ là fixture/read-only
evidence và không mở rộng các tuyên bố provider, live executor, business
outcome, pilot, deployment, distributed locking hay power-loss/atomic
multi-file durability.

Đây vẫn là fixture/read-only evidence; provider, business outcome, live
executor, clean-machine pilot, target deployment, distributed locking và
power-loss/atomic multi-file durability chưa được chứng minh.

**Post-merge regression after PR #385 (2026-09-16):** trên `main`
`4942ffaf0ef8c7a76ba9edafb6654670449d5521`, learner Bot test/vet, 113 Python
tests, BR-18b backup/restore và readiness audit đều PASS; cả 10 GitHub Actions
checks của PR cũng PASS. Audit vẫn trả `NOT_READY_FOR_PRODUCTION` và resolve 68
scoped claims. Record:
`docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-385-20260916.md`. Đây vẫn là
fixture/read-only evidence; provider, business outcome, live executor,
clean-machine pilot, target deployment, distributed locking và power-loss/
atomic multi-file durability chưa được chứng minh.

**Post-merge regression after PR #383 (2026-09-16):** trên `main`
`c910185c435cbbef576208047487ce1394ec9337`, learner Bot test/vet, 113 Python
regression tests, BR-18b backup/restore và readiness audit đều PASS; 10 GitHub
Actions checks của PR cũng PASS. Audit vẫn trả `NOT_READY_FOR_PRODUCTION` và
resolve 68 scoped claims. Record:
`docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-383-20260916.md`. Đây vẫn là
fixture/read-only evidence; provider, business outcome, live executor,
clean-machine pilot, target deployment, distributed locking và power-loss/
atomic multi-file durability chưa được chứng minh.

**M11 expiry authority no-mutation regression (2026-09-16):** regression trên
đường lệnh thật của learner Bot dùng fixture M08-M10 đầy đủ, tạo các checkpoint
backup/restore M11 và chạy fresh Bot binary tại đúng biên `lease.ExpiresAt` hoặc
`authorization.ExpiresAt`. Gate, authorization, reservation và fixture
execution đều bị từ chối; `mission-state.json`, M10/cost registries và
`m11-artifacts.jsonl` của runtime restore vẫn byte-identical sau từng lệnh.
Health registration không được coi là authority rejection vì đó là evidence
append, không phải quyết định production authority. Seam expiry/no-mutation
local này bổ sung cho lease rebind và activation-expiry, nhưng RP-07 vẫn
`PARTIAL`; trusted external time, live executor, provider, business outcome,
crash/power-loss, atomic multi-file, distributed/multi-host, pilot và deployment
vẫn còn mở.

**Post-merge regression after PR #381 (2026-09-16):** trên `main`
`8229ad3c8d3490a8780ea987c3a406a3f08cd96c`, learner Bot test/vet, 113 Python
regression tests, BR-18b backup/restore và readiness audit đều PASS. Audit vẫn
trả `NOT_READY_FOR_PRODUCTION` và resolve 67 scoped claims. Record:
`docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-381-20260916.md`. Đây vẫn là
fixture/read-only evidence; provider, business outcome, live executor,
clean-machine pilot, target deployment, distributed locking và
power-loss/atomic multi-file durability chưa được chứng minh.

**M11 recovery admission concurrent-writer proof (2026-09-16):** regression
`TestMissionM11RecoveryAdmissionSingleWriterAcrossBotProcesses` khởi chạy tám
learner Bot process thật cùng lúc trên một recovery handoff đã được human
review. Runtime gate của new runtime cho phép đúng một immutable
`PRODUCTION_RECOVERY_ADMISSION`; contender còn lại chỉ nhận `BUSY` hoặc
`EXACT_DUPLICATE`, registry sau đó có đúng một admission và process mới retry
vẫn trả exact duplicate. Đây là bằng chứng local fixture/read-only cho seam
concurrent-writer của RP-07, không đóng crash/power-loss, atomic multi-file,
distributed lock, live executor, provider, pilot hay deployment.

**Post-merge regression after PR #379 (2026-09-16):** trên `main`
`80a895ed69d339b7c66e69e70466aac52ded6783`, learner Bot test/vet, 113 Python
regression tests, BR-18b backup/restore và readiness audit đều PASS. Audit vẫn
trả `NOT_READY_FOR_PRODUCTION` và resolve 66 scoped claims. Record:
`docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-379-20260916.md`. Đây vẫn là
fixture/read-only evidence; provider, business outcome, live executor,
clean-machine pilot, target deployment, distributed locking và
power-loss/atomic multi-file durability chưa được chứng minh.

**Full local regression after PR #377 (2026-09-16):** trên `main`
`c11880b19be8b979f90f6dcc1d21cc8bd032aaa5`, bốn Go module test/vet, 113 Python
tests, toàn bộ static validators, BR-16a, BR-18b, disposable n8n M06/M07 engine
và Schedule Trigger đều PASS. Audit vẫn trả `NOT_READY_FOR_PRODUCTION`. Record:
`docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-377-20260916.md`. Đây vẫn là
fixture/read-only evidence; provider, business outcome, live executor,
clean-machine pilot, target deployment, distributed lock và power-loss proof
chưa được chứng minh.

**BR-19 local regression re-verification (2026-09-16):** đã chạy lại các Go
module, semantic validators, smoke BR-16a/BR-18b và disposable n8n engine /
Schedule Trigger trên local Darwin arm64 với n8n 2.38.1 và Node 24.21.0.
Kết quả PASS chỉ là fixture/read-only evidence; provider, business outcome,
clean-machine pilot, target deployment, power-loss và distributed locking vẫn
chưa có bằng chứng.

**M11 source canary grant cross-store binding (2026-09-16):** learner Bot nay
không chấp nhận một `PRODUCTION_LEASE` chỉ vì cặp lease/approval tự nhất quán.
Trong runtime đã có mission state, lease phải khớp ID/version/domain hash của
active canonical M10 `CANARY_GRANT`, và registry entry M10 phải resolve/decode
đúng trước khi append. Backup/restore áp dụng cùng liên kết này; BR-18b tạo
mutation checksum-valid với source grant hash không liên quan và xác nhận
`GRAPH_FAILED` trước khi publish target, còn positive path vẫn restore/replay.
Đây là seam local fixture/read-only có thể nghiệm thu, không đóng RP-07: live
executor, provider/business outcome, pilot, deployment, distributed lock và
power-loss/atomic multi-file durability vẫn mở.

**Full local regression after PR #376 (2026-09-16):** trên `main`
`5bca64f88c1050779ed37882267a1d65d6f5223a`, bốn Go module test/vet, 113 Python
tests, toàn bộ static validators, BR-16a, BR-18b, disposable n8n M06/M07 engine
và Schedule Trigger đều PASS. Audit vẫn trả `NOT_READY_FOR_PRODUCTION`. Record:
`docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-376-20260916.md`. Đây vẫn là
fixture/read-only evidence; không đóng provider, business outcome, live
executor, pilot, target deployment, distributed lock hoặc power-loss proof.

**Full local regression after PR #369/#370 (2026-09-16):** trên `main`
`3ebe32e3aa5157c1c89d34d4cd3b86a94b2f66f5`, chạy lại bốn Go module với
`GOWORK=off go test -count=1 ./...` và `go vet ./...`, 113 Python regression
tests, toàn bộ static validators, BR-12d, BR-13b, BR-16a, BR-18b, disposable
n8n M06/M07 engine và Schedule Trigger. Các lệnh đã chạy đều PASS; readiness
audit vẫn trả `NOT_READY_FOR_PRODUCTION`. Hai validator
`*_operated_execution.py` chỉ nhận execution JSON/store artifact nên không
được gọi độc lập; engine runner đã thực hiện đường end-to-end và PASS. Đây là
evidence fixture/read-only local; provider, business outcome, live executor,
clean-machine pilot, target deployment, multi-host và power-loss vẫn mở.
Record: `docs/architecture/EVIDENCE-LOCAL-REGRESSION-RERUN-20260916.md`.
**BR-18A runbook command smoke (2026-09-16):** chạy chuỗi lệnh runbook bằng
learner Bot build mới trong workspace tạm: `mission init`, history
capture/list/replay, watcher loopback health/log/stop, status sau STOP, rồi
backup v3/restore/replay và hai lệnh M11 bị chặn trước resolve. Kết quả
`APPENDED`, `replay=MATCH`, health `OK`, `RESTORED`, status giữ
`stop_reason=backup-drill` và `RUNBOOK BACKUP/RESTORE STOP DRILL PASS`.
Lần chạy đầu phát hiện ví dụ runbook đặt log/PID trong canonical runtime nên
strict backup inventory từ chối `watcher.log`; đã sửa tài liệu để process
log/PID nằm ngoài artifact store và chạy lại pass. Đây chỉ là local
fixture/documentation smoke cho profile tối thiểu, không đóng full M00–M11
backup graph, power-loss/atomic multi-file, deployment, provider, pilot hoặc
business outcome. Record: `docs/architecture/EVIDENCE-LOCAL-REGRESSION-RERUN-20260916.md`.

**Cập nhật backup/restore M00-M05 graph guards (2026-09-16):** BR-18b smoke
giờ tạo thật `actions.jsonl`, `outcomes.jsonl`, `evaluations.jsonl`,
`proposals.jsonl` và `reviews.jsonl` trước khi tạo manifest v3. Năm mutation
checksum-valid lần lượt orphan decision/action/outcome/evaluation/proposal;
`backup restore` phải trả `VERIFY_FAILED` và không publish target cho từng ca.
Đây là bằng chứng loader/graph upstream local được gọi trên runtime thật,
không phải atomic multi-file, power-loss, distributed lock, provider,
deployment, pilot hay business outcome. Marker: `Cập nhật backup/restore M00-M05 graph guards`.

**Cập nhật backup/restore output-parent ancestor guard (2026-09-16):** POSIX
backup/restore output creation dùng descriptor-pinned `openat(O_NOFOLLOW)` và
`mkdirat` sau preflight. Regression gọi command thật, thay ancestor bằng
symlink ngay trước traversal, yêu cầu cả backup và restore trả `TARGET_ERROR`
và không ghi vào cây external. macOS `/var` alias hệ thống được xử lý riêng;
Windows/fallback vẫn fail closed theo preflight. Đây chỉ là local path-race
hardening, không đóng distributed lock, power-loss/atomic multi-file,
provider, deployment, pilot hoặc business outcome. Marker: `Cập nhật backup/restore output-parent ancestor guard`.

- Baseline review gốc: `ece6a32619e5b9a05d0599b87f50023f38931cb9`; snapshot
  `main` hiện hành nằm trong metadata ở đầu file.
- Trạng thái: **CURRENT_MAIN_TRACKER** — có implementation/test offline đã
  merge; không criterion nào được coi production-ready hoặc pilot-complete.
- Cơ sở: [sổ phát hiện và bằng chứng baseline](evidence/REVIEW-ECE6A32.md).
- Liên kết kế hoạch gốc: [BR-2026-09](BEGINNER-READINESS-PLAN.md); trạng thái tích hợp BR: [readiness matrix](READINESS-MATRIX.json).
- Người lập kế hoạch: Codex theo yêu cầu chủ repo. Người triển khai/reviewer từng đợt: chưa phân công; điền khi nhận việc. Không mặc định chủ repo đã nghiệm thu kế hoạch hoặc cho phép live operation.

**Cập nhật shared M11 historical-chain validator (2026-09-16):** `m11-chain-check`
giữ strict decode tại mission boundary nhưng chuyển toàn bộ kiểm tra liên kết
closed-cycle/resolved-stop, EffectRef machine execution, scope, time, health,
ledger/replay/budget, STOP transition và review sang `core/m11`. Adapter
canonical closed-cycle cũng dùng `core/m11.ValidateClosedCycle`; regression âm
tính hiện tại chạy đúng implementation dùng bởi harness. Đây là bằng chứng
offline/read-only, không chứng minh provider, live executor, business outcome,
crash/power-loss, distributed lock hay production readiness.

**Cập nhật learner schema identity (2026-09-16):** learner Bot không còn khai
báo bản sao cục bộ của `Intent`, `PolicyDecision` và `ApprovalRecord`. Các
field tương ứng trong mission state giờ là alias trực tiếp của
`core/m08.Intent`, `core/m08.PolicyDecision` và `core/m09.ApprovalRecord`, còn
decoder/validator dùng chung tiếp tục là điểm kiểm tra canonical. Regression
`TestLearnerMissionStateUsesCanonicalM08M09Types` sẽ fail nếu một local struct
được reintroduce. Đây chỉ đóng một seam schema-drift offline; authorization/
execution migration rộng hơn, crash/power-loss, provider, live executor,
business outcome và multi-host proof vẫn mở.

Cập nhật sau review PR #94: bổ sung dependency RP-05 cho RP-06, tách nghiệm thu
restore M00–M10 khỏi phần mở rộng M11 bắt buộc ở RP-07a, và giao rõ việc tạo/lưu/
kiểm chuỗi cost-bound, gate, authorization, execution và EffectRef. Các phạm vi
còn lại được ghi cụ thể `PARTIAL`/`OPEN`, không suy từ test PASS sang nghiệm thu
toàn bộ R01–R16.

**Cập nhật committed recovery-journal cleanup acknowledgement (2026-09-15):**
M10 canary, cost-bound và execution-record, cùng direct `m11-stop`, đều đi
qua một cleanup boundary sau khi transition canonical đã hiện hữu. Nếu `Remove`
hoặc parent-directory sync không được ACK, CLI trả non-success
`PUBLISHED_RECOVERY_REQUIRED` kèm transition xác định thay vì `STORE_ERROR` hay
`RECOVERY_REQUIRED` ngụ ý chưa có gì được ghi. Regression gọi CLI thật, inject
lỗi chỉ sau `Remove`, xác nhận journal name đã biến mất và exact retry không tạo
thêm canary/bound/execution hay đổi STOP. Đây chỉ là acknowledgement local,
không chứng minh power-loss, atomic multi-file, executor, business outcome hay
multi-host safety.

**Cập nhật M11 post-remove journal cleanup acknowledgement (2026-09-15):**
M11 outcome, FAILED execution và UNKNOWN→STOP đều có regression fault sau
`Remove` nhưng trước parent sync. Cả ba đường CLI phải trả non-success
`PUBLISHED_RECOVERY_REQUIRED` cùng transition canonical đã thấy; journal path
đã absent và exact retry không thêm outcome/execution hoặc thay STOP. Đây vẫn
chỉ là test seam local sau `Remove`, không chứng minh durability khi mất điện,
atomic transaction đa file, executor, business outcome hay multi-host.

**Cập nhật M10 canary/cost journal path guards (2026-09-15):** trên
Linux/macOS, từng canary và cost-bound recovery journal bị thay bằng FIFO phải
bị prelude/status/reader thật chặn trước đọc. Cùng test stable reader mở rộng
đến execution, canary, cost-bound và M11 journal: thay symlink cùng bytes sau
open đều bị reject. Đây là proof local path identity cho ba M10 replay plan,
không bao phủ Windows FIFO, crash/power-loss, atomic multi-file hay multi-host.

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
| R10 / P1 | Concurrent reserve vượt cap | RP-03 | Barrier đồng bộ 24 process, cap=1: đúng một M10 hoặc M11 reservation, ledger khớp ACK; crash injection không làm mở budget |
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
| RP-01 | Bảo vệ đường dẫn và file đầu vào | RP-00 | S | PARTIAL — M08/M10/M11 output paths reject alias/overwrite and preserve canonical state before portable output; campaign/fixture/backup writer inventory is covered by the shared append guard, while broader platform and non-cooperating-writer boundaries remain open |
| RP-02 | Shared M08 decoder/policy, exact-number/hash contract | RP-01 | M | PARTIAL — learner and harness share M08 decoding/policy plus the strict M09 approval boundary; broader authorization/execution conformance and migration remain open |
| RP-03 | Shared M09/M10 guard, cost-bound/gate/authorization/execution, ledger và STOP | RP-02 | L; chia 03a/03b | PARTIAL — one shared learner-Bot fixture chain now exercises EC-01…EC-05 with byte-level no-mutation rejects, restore and restart; multi-file crash/power-loss, distributed locking and business-execution proof remain open |
| RP-04 | Canonical M06 builder và resolver M07/M08/HTTP | RP-01; tích hợp M08 sau RP-02 | M | PARTIAL — shared fixture builder/resolver, n8n node-chain và real Schedule Trigger regressions exist; governed selected-source profile and deployment-operated run remain open |
| RP-05 | M07 grounded output và tool-result lifecycle | RP-04 | L; chia 05a/05b | PARTIAL — adapter-owned trace/proposal persistence, strict untrusted-model JSON decoding and n8n stub path exist; selected-source/provider operated evidence remains open |
| RP-06 | Snapshot/restore và graph M00–M10, gồm proposal M07 và execution chain | RP-03, RP-04, RP-05 | M | PARTIAL — v3 typed inventory, graph validation and cross-process gate exist; M11 fixture outcome/ledger links are now checked both ways, while broader semantic orphan and crash/host proof remain open |
| RP-07 | M11 lifecycle + mở rộng restore (07a), rồi full chain/walkthrough (07b) | 07a sau RP-02…RP-06; 07b sau gate lifecycle/restore của 07a | L; chia 07a/07b | PARTIAL — learner lifecycle, UNKNOWN→STOP/reconciliation, admission, shared M00–M11 smoke and restore exist; each M11 gate follows activation and retains its exact budget snapshot, each authorization has a prior ALLOW gate/health snapshot and must reserve against that unchanged ledger, authorization lineage now must retain the exact gate-evaluated lease/health/cost artifact IDs, hashes and cost minor, each attempt resolves its prior normal reservation ledger and historical authorization lifetime, each offline evaluation cites exactly its fixture outcome, each M11 chain closes with exact lease/correlation lineage, an UNKNOWN attempt may have only one post-attempt human `NOT_PERFORMED` reconciliation resolution, and fresh-process exact-expiry gate/authorization/reservation/execution rejects are no-mutation checked after restore; multi-file crash seams remain open |
| RP-08 | CI parity/mutation/cross-process coverage | Bắt đầu cùng RP-01; đóng sau RP-07 | M, xuyên các PR | PARTIAL — required offline smokes, disposable M06/M07 n8n engine regressions including M06 Schedule Trigger admission, and one M11 canonical gate-ID mutation proof run in CI; mutation breadth and operated parity remain open |
| RP-09 | Readiness audit có dữ liệu/evidence, chốt offline acceptance | RP-06, RP-07, RP-08 | M | PARTIAL — matrix/graph/plan/CI audit is structured and public entrypoints retain scoped NOT_READY boundary; remote CI and external evidence remain outside local audit |
| RP-10 | n8n operated run, pilot máy sạch, deployment drill | RP-09 và lựa chọn môi trường/quyền cần thiết | M/L | OPEN — requires selected environment, authority and independently recorded operated evidence |

Luồng ưu tiên: RP-01 → RP-02 → RP-03; RP-04 có thể làm song song trên file độc lập. RP-06 chỉ merge sau RP-03/RP-04/RP-05 để kiểm proposal đã persist và execution chain thật. RP-06 nghiệm thu inventory M00–M10; RP-07a bổ sung artifact M11 và phải mở rộng manifest/loader/restore tests trong cùng gói, rồi RP-07b mới nghiệm thu toàn chuỗi. Không thêm dependency RP-07 ngược vào RP-06 gây vòng lặp. RP-08 đưa test vào từng PR, không đợi cuối dự án mới bật gate. Không đặt ngày production trước khi chốt điều kiện RP-10.

### Runtime gap đã nghiệm thu cục bộ — RP-07 canonical M11 identity integrity

Phạm vi đã chuyển công thức ID deterministic của M11 từ learner vào
`core/m11`, rồi bắt canonical graph tự tính lại ID sau khi resolve parents.
Trước đó, learner đã tạo gate/authorization/execution theo công thức cố định,
nhưng backup/harness caller trực tiếp vẫn có thể nhận các ID schema-valid giả
nếu thay đồng bộ liên kết downstream. Đây là boundary offline; nó không cấp
authority, không thay ledger runtime, và không chứng minh executor, power-loss
hay multi-host transaction.

- `core/m11` owns gate ID từ lease/intent/health/cost/ledger/time, authorization
  ID từ gate/executor và execution ID từ authorization.
- `core/m11.ValidateArtifactGraph` resolve parents rồi require cả ba ID khớp
  computation canonical; learner reuse chính helpers này.
- Regression dựng graph M11 hợp lệ rồi thử forge từng gate, authorization và
  execution ID khi các link còn lại đã được cập nhật nhất quán.
- Không gọi executor/provider và không suy identifier guard thành proof ledger,
  multi-file transaction, power-loss hoặc multi-host.

**Cập nhật M11 authorization exact gate lineage (2026-09-16):**
`core/m11.ValidateArtifactGraph` giờ yêu cầu authorization khớp chính xác với
gate mà nó viện dẫn về lease ID/version/hash, health snapshot ID/hash và cost
bound ID/hash/minor. Regression tạo health snapshot và cost bound riêng nhưng
đều checksum-valid, rồi thử chuyển authorization sang hai artifact đó; graph
phải reject trước khi authorization có thể được dùng. Đây là bằng chứng graph
offline/read-only cho lineage của authorization, không phải proof ledger,
executor/provider, business outcome, crash/power-loss, atomic multi-file,
distributed locking, pilot hoặc deployment.

**Cập nhật M11 correlation lineage (2026-09-17):**
`core/m11` coi `ProductionLease.CorrelationID` là correlation root của M11.
Graph registry không nhận cost bound checksum-valid nhưng thuộc correlation
khác, cũng không nhận authorization hoặc execution/cycle record mang
correlation không khớp với lease. Shared historical-chain validator áp dụng
cùng invariant cho cả `resolved_stop` và `closed_cycle`; learner registry
loader và mission-runtime `m11-chain-check` đều chạy đúng các guard này trên
fixture thật. Đây là bằng chứng offline/read-only cho lineage, không phải proof
provider, live executor, business outcome, crash/power-loss, atomic multi-file,
distributed locking, pilot hoặc deployment.

**Cập nhật RP-08 M11 authorization lineage mutation proof (2026-09-16):**
CI có mutation runner tạo bản sao disposable của `core/m11`, tháo toàn bộ
enforcement nối authorization với đúng artifact mà gate đã đánh giá, rồi chạy
graph regression thật. Regression phải fail bằng forged-lineage assertion; nếu
implementation bị làm yếu mà test vẫn xanh thì job fail. Đây là mutation proof
offline/read-only cho một block validator, không phải proof ledger,
executor/provider, business outcome, crash/power-loss, atomic multi-file,
distributed locking, pilot hoặc deployment.

**Cập nhật implementation RP-07 (journal path guard):** M11 chỉ đọc recovery
journal là regular file ngay trong runtime. Symlink hoặc special file trả
`RECOVERY_REQUIRED` trước resolver/writer, không đọc target ngoài runtime và
không tạo registry entry. Regression khởi động binary Bot mới với cả FAILED,
UNKNOWN→STOP và outcome journal: `status` chỉ trả `RECOVERY_REQUIRED`, còn
writer có lock mới replay exact transition; UNKNOWN vẫn giữ STOP. Đây là
path-integrity/restart guard offline, không phải proof power-loss, atomic
multi-file commit hay multi-host recovery.

**Cập nhật M11 special-file journal guard (2026-09-12):** trên Linux/macOS,
regression tạo FIFO cho từng journal FAILED, UNKNOWN và outcome. Recovery
prelude, `mission status` và reader đều phải trả fail-closed trước bất kỳ read
FIFO nào. Điều này hoàn tất bằng chứng special-file của local path boundary;
Windows giữ ngoài test do API FIFO khác, còn crash/power-loss/multi-host vẫn
không được suy ra.

**Cập nhật M11 runtime-loader orphan guard (2026-09-12):** registry loader
được regression bằng một `PRODUCTION_LEASE_APPROVAL` schema-valid, integrity
valid nhưng không có immutable lease cha. `loadM11ArtifactRegistry` phải gọi
canonical graph validator và reject orphan trước khi runtime có thể dùng entry.
Case này kiểm đường learner/restore thực, không chỉ gọi core validator trực
tiếp; vẫn không bao phủ mọi tổ hợp orphan, crash/power-loss hay multi-host.

**Cập nhật M10 execution-journal path guard (2026-09-11):** M10 dùng cùng
`Lstat` regular-file boundary trước khi status/resolver đọc pending journal,
và trước writer/backup recovery parse replay record. Regression tạo symlink tới
file bên ngoài runtime rồi khởi động Bot binary mới: status, `m10-resolve`,
writer và backup đều trả `RECOVERY_REQUIRED`; mutable state/registry và target
bên ngoài giữ nguyên. Đây là path-integrity/restart guard cho M10 journal,
không phải transaction đa-file, proof power-loss hay multi-host recovery.

**Cập nhật M10 special-file journal guard (2026-09-12):** trên Linux/macOS,
regression tạo FIFO ở execution journal và xác nhận recovery prelude, status
và reader fail-closed trước read. Windows FIFO semantics và proof
crash/power-loss/multi-host vẫn nằm ngoài phạm vi.

**Cập nhật canonical runtime-store path guard (2026-09-12):** trước parser hay
graph validation, learner chỉ đọc `mission-state.json`, `STOP`, M10 artifact
registry/cost-bound registry/fixture outcomes và M11 artifact registry/fixture
outcomes qua `Lstat → open/fstat → Lstat`. Regression thay từng path bằng symlink đến bytes ngoài runtime vẫn
hợp lệ (state dùng đúng bytes đã initialize) và nhận reject trước khi dùng state
hoặc authority artifact; external target không đổi. CI mutation đồng thời bỏ
các guard trong checkout tạm và đòi mọi assertion runtime-store thật fail. Đây
chỉ là guard path local cho các store nêu tên, không bao phủ mọi JSONL runtime,
TOCTOU ngoài primitive này, crash/power-loss, multi-host hay executor.

**Cập nhật portable-input stable reader (2026-09-13):** reader có giới hạn
cho file campaign cục bộ và input CLI (fixture/config/capture) nay dùng cùng
ranh giới `Lstat → open/fstat → đọc descriptor hai lần → Lstat` với kiểm giới
hạn kích thước trước/trong/sau khi đọc. Regression gọi đúng
`watcher accesstrade-shopee-campaign-import`, thay sanitized capture sau open
bằng symlink tới file ngoài có cùng bytes, và đòi `INPUT_ERROR` trước khi có
history ACK; external file giữ nguyên. Đây là hardening local cho một shared
portable reader, chưa bao phủ mọi entrypoint portable, writer không hợp tác,
crash/power-loss, multi-host hay operated evidence.

**Cập nhật M08-M11 portable-input stable reader (2026-09-13):** mọi artifact
được command mission nhận theo pathname — M08 request/policy/intent, M09
approval, M10 grant/cost-bound/gate/authorization/outcome, M11 register/
outcome/recovery-admission và cost-bound legacy compatibility — nay đi qua
reader bounded có identity ổn định trước khi chúng được decode hoặc bind runtime.
Regression gọi `m10-reserve-authorization` thật, thay authorization sau open
bằng symlink ngoài có cùng bytes, và phải `INPUT_ERROR` không thay ledger;
mutation CI thay shared reader bằng `os.ReadFile` và đòi assertion thực fail.
Đây là guard pathname local, không chứng minh atomic transaction đa-file,
crash/power-loss, host phân tán, executor hay business outcome.

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

**Cập nhật M10 resolver read gate (2026-09-11):** `m10-resolve` giữ runtime
gate local trước khi kiểm journal hoặc registry. Writer đang hoạt động trả
`BUSY`, thay vì resolver đọc artifact giữa lifecycle transition. Đây chỉ là
exclusion local cho read path, không thay transaction đa-file, distributed lock
hay proof recovery sau power-loss.

**Cập nhật M10 reservation barrier (2026-09-11):** shared smoke giờ khởi tạo
24 subprocess Bot, xác nhận mọi wrapper đã sẵn sàng rồi mới publish một release
file chung. Với đúng một slot cap/cost còn lại, chỉ một `RESERVED` được commit;
retry của cả 23 loser đều phải `BUDGET_DENIED` và usage giữ nguyên. Đây chứng
minh local cross-process cap=1 cho command/runtime lock hiện hữu, không chứng
minh distributed/multi-host lock hoặc atomic recovery sau crash.

**Cập nhật STOP/reserve race (2026-09-11):** shared smoke clone cùng runtime
còn một slot, cho `m10-reserve` và `m11-stop` qua một subprocess barrier. Chỉ
operation lấy local runtime gate trước có thể hoàn tất; sau khi STOP durable,
mọi reservation mới bị `STOPPED` và mutable state không đổi. Đây không suy ra
ordering liên-host, transaction đa-file hay crash/power-loss recovery.

**Cập nhật EC-01…EC-05 learner acceptance (2026-09-11):** cùng một shared
workspace và Bot binary giờ ghi rõ năm ca execution chain trong
`smoke_br16a_offline.py`. EC-01 sinh M09 approval, M10 grant/cost/gate/
authorization/reservation, deterministic failed-before-dispatch fixture,
EffectRef outcome, rồi backup/restore và resolve lại bằng process mới. EC-02
chặn cost bound missing, unregistered/tampered và expired trước reservation;
EC-03 chặn gate/authorization chưa register và execution chưa reserve; EC-04
chặn outcome trỏ reservation, approval hoặc effect kind sai. Mỗi reject EC-02
đến EC-04 so byte canonical state/registry/outcome trước-sau. EC-05 retry
đúng UNKNOWN là exact duplicate, giữ reservation/reconciliation/STOP và vẫn
được kiểm sau restore/restart. Đây là fixture no-side-effect, không phải
executor, business outcome, transaction đa-file, power-loss hay multi-host
proof; RP-03 vẫn `PARTIAL`.

**Cập nhật cost overflow boundary (2026-09-11):** canonical M10 gate được
regression tại ranh giới `int64`: ledger ở `MaxInt64 - 1` với bound 2 phải
`CANARY_COST_BUDGET_EXHAUSTED`, không wrap thành capacity hay execution
authority. Đây là chứng cứ arithmetic cho core gate; không thay evidence về
transaction đa-file hoặc recovery sau crash.

**Cập nhật M07 output-path preflight (2026-09-11):** M07 tool-result/proposal
giờ reject same path, hardlink và symlink output alias trước khi đọc bất kỳ
history/model/registry/tool input nào. Regression gọi CLI implementation thật
và kiểm input bytes không đổi. Đây chỉ mở rộng inventory output của M07; các
writer khác vẫn phải được audit riêng trước khi đổi RP-01 khỏi `PARTIAL`.

**Cập nhật advisor fixture output-parent guard (2026-09-15):** `bot advisor
fixture-run` kiểm `Lstat` output parent do caller chọn trước `MkdirTemp`, rồi
kiểm lại ngay trước khi tạo bundle; parent phải tồn tại, là directory và không
phải symlink. Regression gọi đúng CLI với symlink đến thư mục ngoài ở cả hai
boundary, nhận `PATH_ERROR` và kiểm thư mục ngoài vẫn rỗng.
Đây chỉ là guard trực tiếp của output parent cho bundle fixture read-only;
ancestor-path substitution, crash/power-loss, multi-host, provider và business
outcome vẫn nằm ngoài scope. Inventory các writer khác vẫn mở, nên RP-01 giữ
`PARTIAL`.

**Cập nhật advisor fixture failed-staging cleanup (2026-09-15):** sau khi tạo
staging directory, `bot advisor fixture-run` giờ xóa staging nếu build/evaluate,
ghi bundle hoặc sync bị lỗi; chỉ giữ bundle sau khi toàn bộ nội dung đã được sync
và parent sync thành công. Regression gọi đúng CLI, inject lỗi build, kiểm path
bundle được báo đã biến mất, parent không còn staging và
`execution_permitted=false`. Đây là cleanup local có giới hạn; crash/power-loss,
multi-host, provider và business outcome vẫn ngoài scope, nên RP-01 vẫn
`PARTIAL`.

**Cập nhật ACCESSTRADE pending import recovery (2026-09-15):** importer outcome
read-only đã có pending receipt journal để chặn mixed snapshot, nhưng chưa có
entrypoint hoàn tất replay. `bot outcome accesstrade-recover HISTORY ACTIONS
OUTCOMES REPORT.csv MANIFEST.json` nay chỉ mở journal strict, re-derive đúng
CSV/manifest local và yêu cầu receipt/hash khớp byte identity đã ghi. Nó chỉ
append phần còn thiếu khi tập outcome là rỗng hoặc toàn bộ exact; changed input
hay partial set bị từ chối mà giữ nguyên journal/store. Sau khi receipt graph
được reload hợp lệ mới xóa journal; uncertainty sau khi journal visible hoặc
cleanup parent-sync trả `PUBLISHED_RECOVERY_REQUIRED`, không mời tạo snapshot
mới. Regression gọi đúng CLI, bao phủ empty, outcome-visible/receipt-missing,
changed report và partial outcomes. Fault seam chỉ trong test cũng chứng minh
khi receipt/outcome đã visible nhưng journal removal mất ACK trước parent sync,
import trả `PUBLISHED_RECOVERY_REQUIRED`; retry exact là `EXACT_DUPLICATE`,
không tạo snapshot thứ hai. Đây là recovery local cho report đã được cung cấp,
không fetch/đăng nhập ACCESSTRADE, không xác nhận business outcome, atomicity
power-loss hay multi-host transaction; RP-01/RP-06 vẫn `PARTIAL`.

**Cập nhật campaign result visible acknowledgement (2026-09-15):** controlled
advisor canary đã reserve trước request, nên sau response không thể coi lỗi
persist result chung chung là lý do gửi provider lần nữa. Result writer nay
canonical-reload exact immutable result sau lỗi file-sync/close/parent-sync;
nếu đã visible, runner và CLI trả non-success
`PUBLISHED_RECOVERY_REQUIRED`, CLI bàn giao result đã resolve với
`execution_permitted=false`. Canary gọi lại là `REVIEW_REQUIRED` và không gửi
request thứ hai. Regression dùng loopback và test-only fault seam, không dùng
credit. Đây là acknowledgement boundary local cho một result file, không phải
billing reconciliation, provider operation, crash/power-loss, transaction
nhiều file, business outcome hay multi-host proof; RP-03 vẫn `PARTIAL`.

**Cập nhật backup/restore missing-parent guard (2026-09-14):** trước khi
`backup create` hoặc `backup restore` gọi `MkdirAll` cho output parent do caller
chọn, shared preflight lùi từng component chưa tồn tại tới entry đã có và reject
symlink/non-directory. Regression đi đúng CLI bằng đường dẫn `symlink/missing`
cho cả hai lệnh, nhận `TARGET_ERROR` và kiểm thư mục ngoài vẫn rỗng. Đây không
phải guard cho ancestor có sẵn hay concurrent substitution sau preflight; cũng
không là proof crash/power-loss, multi-host, provider hay business outcome.

**Cập nhật immutable artifact missing-parent guard (2026-09-14):** `M07`,
`M08`, `M10` và `M11` portable artifact publisher dùng cùng preflight trước
`MkdirAll`. Regression gọi `writeNewJSON` qua đường `symlink/missing/artifact`
và kiểm không có thư mục/file nào xuất hiện ở external target. Đây chỉ khép
missing component của parent path; ancestor có sẵn, concurrent substitution,
crash/power-loss, multi-host, provider và business outcome vẫn ngoài scope.

**Cập nhật runtime missing-parent guard (2026-09-14):** trước `mission init`
và mọi mission writer tạo runtime mới, shared preflight reject
`symlink/missing/runtime` trước `MkdirAll`. Regression thực nhận `STORE_ERROR`
và external target rỗng. Pre-existing ancestor alias, concurrent substitution,
crash/power-loss, multi-host, provider và business outcome vẫn ngoài scope.

**Cập nhật M07 history read gate (2026-09-11):** CLI M07 và adapter M07 giờ
giữ cùng local history runtime gate với watcher trước khi resolve canonical
record. Regression giữ writer gate thật rồi xác nhận cả `m07 context` và
`/v1/m07/context` trả `BUSY`, không phát hành context từ history đang thay đổi.
Đây chỉ là exclusion trên filesystem một host; distributed/multi-host locking
và proof crash/power-loss vẫn mở.

**Cập nhật M08 history read gate (2026-09-11):** `m08-intent` và nhánh agent
của `m08-policy` giờ giữ local history runtime gate trước khi resolve canonical
record/proposal. Regression giữ writer gate thật rồi xác nhận learner command
trả `BUSY`, không tạo intent từ history đang thay đổi. Đây vẫn chỉ là exclusion
trên filesystem một host; distributed/multi-host locking và proof
crash/power-loss vẫn mở.

**Cập nhật derived-store read gate (2026-09-11):** list của M03 action, M04
outcome, M05 evaluation/proposal/review giờ giữ cùng local history runtime gate
với import/watcher. Một regression gọi đúng năm learner list implementation
trong khi giữ writer gate thật; tất cả phải trả `BUSY`, thay vì lộ snapshot
history/action/outcome/evaluation pha trộn. Đây chỉ là exclusion trên
filesystem một host; distributed/multi-host locking và proof crash/power-loss
vẫn mở.

**Cập nhật canonical HTTP history gate (2026-09-11):** `/v1/history` giờ giữ
local history runtime gate trước khi resolve record, còn M06/fetch/HTTP handoff
append và replayed ACK chạy chung một critical section. Regression giữ writer
gate thật: HTTP read phải trả `BUSY` và append+resolve không được hoàn tất. Đây
chỉ là exclusion trên filesystem một host; distributed/multi-host locking và
proof crash/power-loss vẫn mở.

**Cập nhật history-consumer gates (2026-09-11):** `history` list/replay/decision,
`advisor mock` và list receipt ACCESSTRADE giờ giữ local history runtime gate
trước khi load/resolve chuỗi evidence. Regression giữ writer gate thật rồi gọi
toàn bộ command: từng path phải fail closed (`BUSY` hoặc lỗi busy cho CLI
history), không trả derived context trong lúc write. Đây chỉ là exclusion trên
filesystem một host; distributed/multi-host locking và proof crash/power-loss
vẫn mở.

**Cập nhật atomic immutable artifact publish (2026-09-11):** mọi portable JSON
artifact qua `writeNewJSON` giờ write+fsync vào temporary sibling rồi publish
bằng hard-link không overwrite. Regression inject lỗi trước publish: target và
temporary không được còn lại; retry tạo artifact hoàn chỉnh và exact retry vẫn
`EXACT_DUPLICATE`. Đây không chứng minh power-loss/filesystem durability hay
multi-host atomicity, nên các phạm vi đó vẫn `PARTIAL`.

**Cập nhật immutable artifact publish recovery status (2026-09-14):** sau khi
`Link` đã làm artifact bất biến visible nhưng trước khi sync directory cha,
`writeNewJSON` phân biệt lỗi đó với lỗi chưa publish bằng typed uncertainty.
M08 intent/policy CLI, M07 CLI/HTTP adapter và M11 `recovery-export` trả
`PUBLISHED_RECOVERY_REQUIRED` kèm artifact visible, không mô tả là `CONFLICT`
hoặc `PERSISTENCE_ERROR` retryable. Handoff M11 vẫn non-authorizing. Regression
gọi publisher, M08 CLI, M07 adapter và reviewed M11 handoff thật: artifact/tool
trace vẫn tồn tại, còn exact retry chỉ ACK bytes y hệt. Đây là báo cáo
fail-closed về trạng thái syscall local; không chứng minh durability sau
power-loss, transaction đa file hay filesystem đa host.

**Cập nhật M07 visible artifact recovery disclosure (2026-09-15):** M07 đã
trả `PUBLISHED_RECOVERY_REQUIRED` khi trace/proposal bất biến visible nhưng
directory sync chưa được xác nhận, nhưng response trước đây không mang identity
đã validate để caller đối chiếu retry. CLI và HTTP adapter nay kèm trace hoặc
proposal deterministic chỉ trong response non-success đó;
`execution_permitted=false` và workflow không có ACK nên không được cite/đi
tiếp. Regression chạy cả CLI tool/proposal lẫn HTTP tool/proposal, inject fault
sau publish, kiểm sidecar visible, identity khớp và exact retry mới
`ACK`/`EXACT_DUPLICATE`. Đây là recovery disclosure local, không là provider,
power-loss, transaction đa file hay multi-host proof.

**Cập nhật JSONL acknowledgement boundary (2026-09-11):** canonical và derived
JSONL append giờ chỉ thành công sau `fsync` file và parent directory. Regression
inject lỗi sau write nhưng trước sync: call phải trả error, không ACK thành
công; append tiếp theo giữ line framing hợp lệ. Đây không là mô phỏng
power-loss/filesystem crash hoặc multi-host transaction, nên vẫn `PARTIAL`.

**Cập nhật history handoff visible append uncertainty (2026-09-14):** nếu
canonical history append báo lỗi sau khi exact `HistoryRecord` đã replay được,
`AppendHistory` trả `PUBLISHED_RECOVERY_REQUIRED` thay vì ngụ ý append có thể
retry an toàn. M06 watcher/HTTP handoff canonical-reload record đó nhưng giữ
`canonical_history_ack=false`, trả non-zero/HTTP 500 và không cho consumer coi
handoff là ACK; chỉ retry exact mới trả `EXACT_DUPLICATE` với ACK. Regression
chạy `watcher history-handoff` thật, inject ACK loss sau append, rồi kiểm record
canonical và retry. Đây chỉ là boundary acknowledgement một file; không chứng
minh fsync/power-loss, transaction đa file, provider/selected-source run hay
multi-host safety.

**Cập nhật M06 adapter visible append uncertainty (2026-09-15):** workflow
M06 trước đây biến acknowledgement loss sau append thành `HANDOFF_ERROR`, làm
n8n không phân biệt được record đã canonical nhưng chưa ACK với một append chưa
tồn tại. Adapter nay trả HTTP 500 `PUBLISHED_RECOVERY_REQUIRED`, kèm record đã
resolve và `canonical_history_persisted=true` nhưng `canonical_history_ack=false`.
Regression gọi handler thật với fault sau append, rồi xác nhận exact retry trả
`EXACT_DUPLICATE` cùng ACK. Đây là proof local một JSONL file; không là n8n hay
provider operated run, fsync/power-loss, transaction đa file hoặc multi-host
safety.

**Cập nhật M06 fixture-import visible append uncertainty (2026-09-15):** CLI
fixture-import có cùng append/replay path, nhưng từng trả `HANDOFF_ERROR` dù
record đã canonical sau ACK loss. Lệnh nay giữ status
`PUBLISHED_RECOVERY_REQUIRED` và artifact `persisted=true`; exact retry duy
nhất mới trả `EXACT_DUPLICATE`. Regression dùng CLI thật và fault sau append để
kiểm record vẫn replay được rồi retry ACK. Đây là boundary một JSONL local,
không là provider/n8n operated run, fsync/power-loss, transaction đa file hay
multi-host safety.

**Cập nhật M06 pinned-fetch visible append uncertainty (2026-09-15):** CLI
`fetch-fixture` gọi fixed pinned synthetic source nhưng cũng từng che ACK loss
sau append dưới `HANDOFF_ERROR`. Lệnh nay trả
`PUBLISHED_RECOVERY_REQUIRED`, giữ artifact đã resolve là `persisted=true`, và
chỉ exact retry trả `EXACT_DUPLICATE`. Regression gọi command thật; fetch seam
chỉ thay request mạng trong test, không mở URL/configuration mới, rồi inject lỗi
sau append. Đây là boundary JSONL local; không là external fetch/provider
operation, fsync/power-loss, transaction đa file hay multi-host safety.

**Cập nhật canonical JSONL framing guard (2026-09-14):** reader JSONL dùng
chung giờ fail-closed nếu file không rỗng kết thúc thiếu LF. Vì `AppendLine`
chỉ ACK một record đã có LF sau sync, final line không có LF được xem là append
có thể bị gián đoạn, không phải JSON hợp lệ để history/M03 action/M04 outcome,
M05 evaluation/proposal/review hoặc receipt ACCESSTRADE tiếp tục dùng. M10
artifact/cost-bound registry, M11 artifact registry, M10/M11 fixture outcome
loader và backup receipt requirement ACCESSTRADE vốn split một stable byte
snapshot cũng gọi cùng rule trước parse. Regression chạy real action/outcome
command với store bị cắt LF, đòi `STORE_ERROR` và bytes không đổi; shared store
regression cùng direct M10/M11 registry/outcome và backup-receipt loader
coverage chặn framing trước semantic decode. Đây là integrity guard filesystem
local, không chứng minh recovery sau crash/power-loss, atomic multi-file hay
transaction multi-host; RP-03 vẫn `PARTIAL`.

**Cập nhật derived JSONL visible append uncertainty (2026-09-14):** M03 action,
M04 outcome và M05 evaluation/proposal/review không còn trả `STORE_ERROR` mơ hồ
khi shared JSONL writer báo lỗi sau lúc exact record đã có thể được canonical
loader đọc lại. Mỗi command re-read đúng derived store; chỉ khi record bằng đúng
payload mới trả `PUBLISHED_RECOVERY_REQUIRED` (non-zero, kèm artifact) và retry
exact trả `EXACT_DUPLICATE`. Regression gọi từng command thật, inject ACK loss
sau append, rồi kiểm loader/retry thật. Đây chỉ là acknowledgement boundary của
một JSONL file local; không chứng minh `fsync`/power-loss, transaction nhiều
file, business outcome hay multi-host safety.

**Cập nhật atomic backup publish (2026-09-12):** `backup create` và `restore`
copy/verify toàn bộ snapshot trong sibling staging, rồi publish qua rename tới
target rỗng dưới target gate. Mỗi copied file mới phải `fsync`; mọi directory edge
từ file tới staging root cũng sync trước graph verify/publish. Regression inject
lỗi sau write, sau file-sync, trước và sau directory-sync: staging bị dọn,
target không bị chiếm và retry publish/restore thành công. Reader snapshot cũng kiểm
`Lstat → open/fstat → Lstat` của regular file; regression swap `mission-state`
thành symlink sau inventory phải reject trước copy/publish, không đọc hay đổi
external target. Đây không chứng minh crash/power-loss durability sau rename hay
transaction multi-host, nên vẫn `PARTIAL`.

**Cập nhật backup target-absence guard (2026-09-15):** `backup create` giờ
đòi target chưa tồn tại, giống contract publish staging của `restore`; không
còn nhận rồi tự xoá một directory rỗng do caller tạo. Sau khi nhận target gate
và ngay trước rename, bất cứ pathname target nào xuất hiện đều trả
`TARGET_NOT_EMPTY`, giữ nguyên directory đó và dọn staging. Regression gọi CLI
thật cho cả target rỗng có sẵn và target xuất hiện trong lúc copy. Đây ngăn một
lần publish lỗi làm thay đổi target caller-owned mà chưa tạo backup; nó không
chứng minh crash/power-loss, atomic transaction nhiều file hoặc multi-host.

**Cập nhật immutable artifact parent recheck (2026-09-15):** trước hard-link
publish, publisher kiểm lại chính parent directory đã dùng để tạo temporary
artifact. Parent mất hoặc thành symlink ở seam `before_publish` bị reject trước
khi `Link` resolve pathname; regression thay parent bằng symlink tới directory
ngoài và xác nhận nó vẫn rỗng. Đây chỉ thu hẹp một cửa sổ local đã kiểm tại
boundary publish; không chứng minh thay tên không hợp tác sau recheck,
crash/power-loss, atomic transaction nhiều file hay multi-host.

**Cập nhật publish recovery status (2026-09-14):** sau khi final `rename`
thành công, staging không còn riêng tư. Nếu sync directory cha sau đó lỗi,
`backup create` hoặc `restore` trả `PUBLISHED_RECOVERY_REQUIRED` với manifest
thay vì giả vờ target chưa publish hoặc mời retry ghi đè. Regression inject lỗi
đúng giữa rename và parent sync cho cả backup/restore: target đã visible vẫn
verify/load được, còn lần gọi lại cùng target bị `TARGET_NOT_EMPTY`. Thành công
chỉ trả `BACKED_UP`/`RESTORED` sau parent sync. Đây là báo cáo fail-closed của
trạng thái syscall cục bộ, không chứng minh durability qua crash/power-loss,
không tự sửa target unsynced và không là transaction đa host.

**Cập nhật mission-state publish recovery status (2026-09-14):** `writeJSONAtomic`
phân biệt lỗi trước rename với lỗi parent-directory sync sau rename của
`mission-state.json`. Lỗi sau rename trả typed error, và entrypoint mission
map thành `PUBLISHED_RECOVERY_REQUIRED` thay vì `STORE_ERROR` retryable.
Regression M10 cap=1 inject seam ngay sau rename: reservation đã visible trong
canonical state, retry cùng ID là `EXACT_DUPLICATE`, ID mới vẫn bị từ chối và
Bot process mới replay `VALID`. Đây chỉ là trạng thái an toàn cho mission-state
local. Cùng primitive cũng giữ `STOP` sau lỗi ở parent sync của lần ghi state
hoặc marker: nếu state đã rename nhưng marker chưa tồn tại, state canonical vẫn
chặn Bot mới trước khi đọc input; nếu marker đã rename, cả hai đều visible.
Journals/store khác,
crash/power-loss và transaction nhiều file vẫn còn phạm vi mở.

**Cập nhật sticky STOP reason (2026-09-14):** khi `mission-state.json`
đã stopped, `m11-stop` fail-closed với `STOPPED` và không ghi lại state hoặc
marker dù caller gửi reason khác. Regression lưu bytes hai artifact sau STOP
đầu tiên, thử overwrite reason rồi yêu cầu cả hai byte-identical. Đây chỉ khóa
local reason và không giải quyết recovery multi-file/power-loss.

**Cập nhật sticky STOP mutation proof (2026-09-14):** CI tạo checkout Bot
disposable, bỏ đúng existing-STOP guard rồi chạy lại byte-identity regression.
Job chỉ pass khi regression thực sự fail tại assertion dedicated; vì vậy test
không thể xanh nếu guard bị bỏ vô tình. Đây là mutation proof local, không mở
rộng guarantee crash/power-loss hay locking đa host.

**Cập nhật artifact-registry sync boundary (2026-09-11):** M10/M11 registry
append giờ sync parent directory trước khi return success. M11 journal recovery
regression mở rộng fault từ trước/sau write sang sau file và directory sync;
pending journal vẫn là boundary fail-closed và recovery exact. Đây không chứng
minh power-loss hoặc multi-host transaction, nên vẫn `PARTIAL`.

**Cập nhật M10 partial-commit restart guard (2026-09-11):** regression inject
fault sau khi M10 execution record đã vào registry nhưng trước khi reservation
state commit, sau đó khởi động Bot binary mới. Process mới trả
`RECOVERY_REQUIRED` cho status và resolver, nên không lộ graph registry/state
dở dang trước locked recovery. Backup vẫn là con đường recovery có kiểm; ca này
không chứng minh kill/power-loss tại filesystem boundary, transaction đa-file
hay recovery multi-host.

**Cập nhật M10 execution-journal publish boundary (2026-09-14):** một seam
khác inject lỗi sau khi `m10-execution-journal.json` đã rename-visible nhưng
trước parent-directory sync. Command trả `PUBLISHED_RECOVERY_REQUIRED`, không
tạo portable execution output và Bot process mới trả `RECOVERY_REQUIRED` cho
`status` lẫn `m10-resolve`. Chỉ exact retry dưới local writer lock mới replay
journal, đăng ký/bind đúng một execution rồi mới ghi portable output. Đây là
kiểm local cho một syscall boundary; không chứng minh atomicity đa-file,
kill/power-loss hoặc phối hợp đa host.

**Cập nhật M11 outcome-journal publish boundary (2026-09-14):** `m11-outcome`
trước đây trả `REJECTED` nếu `m11-outcome-journal.json` đã rename-visible nhưng
directory sync chưa được xác nhận. Emit nay nhận mọi `atomicPublishUncertain`
và trả `PUBLISHED_RECOVERY_REQUIRED`; riêng command trả outcome/post-ledger
xác định từ journal, không tuyên bố outcome đã vào mọi store. Regression chạy
CLI thật, khởi động Bot mới để chặn `status`/resolver, rồi exact retry có lock
replay một outcome/ledger. Các journal FAILED và UNKNOWN→STOP cần nghiệm thu
riêng; không suy atomicity đa-file, power-loss hoặc multi-host từ seam này.

**Cập nhật M11 execution-journal publish boundary (2026-09-14):** cùng CLI
regression giờ cover riêng `m11-record-failed` và `m11-record-unknown` khi
journal của chúng đã rename-visible nhưng parent sync lỗi. Mỗi lệnh trả
`PUBLISHED_RECOVERY_REQUIRED` với execution + ledger dự kiến; Bot mới chặn
status/resolver, sau đó exact retry có lock replay một execution. UNKNOWN chỉ
chứng minh stopped-ledger/STOP recovery local; không chứng minh external effect
được xác minh, atomicity đa-file, power-loss hoặc multi-host.

**Cập nhật M11 execution-journal cleanup uncertainty (2026-09-14):** Sau khi
FAILED execution/ledger, hoặc UNKNOWN execution/stopped-ledger/durable STOP,
đã canonical, lỗi dọn journal không còn bị báo `REJECTED`. Cả hai command trả
`PUBLISHED_RECOVERY_REQUIRED` cùng transition xác định; journal còn lại giữ Bot
mới ở `RECOVERY_REQUIRED`, còn exact retry chỉ dọn journal và trả
`EXACT_DUPLICATE`. Regression chạy hai command CLI thật với seam trước remove.
Đây vẫn chỉ là acknowledgement boundary local, không chứng minh fsync,
power-loss, transaction đa-file, external effect hay multi-host.

**Cập nhật direct M11 STOP journal recovery (2026-09-14):** `m11-stop` ghi
journal strict chứa reason trước `mission-state.json` và marker `STOP`. Fault
sau rename ở journal, state hoặc marker giữ journal; Bot mới ở read path trả
`RECOVERY_REQUIRED`, còn writer có lock hoặc `backup create` replay đúng
state/marker/reason rồi mới giữ `STOPPED`. Không còn chấp nhận state-only STOP thiếu marker vô thời hạn.
Đây là replay local cho đúng hai file, không chứng minh atomic transaction,
kill/power-loss durability hay coordination đa host.

**Cập nhật M10 cost-bound two-store journal (2026-09-13):** đăng ký một
trusted cost bound nay ghi journal bất biến trước khi ghi cả immutable M10
artifact registry và compact cost-bound index. Lỗi injected trước artifact,
sau artifact, sau index, hoặc sau khi journal visible nhưng trước directory
sync giữ journal; `status` và `m10-resolve` của Bot mới trả
`RECOVERY_REQUIRED` trước khi lộ transition một nửa. Seam visible-journal trả
`PUBLISHED_RECOVERY_REQUIRED`, rồi chỉ writer có local lock hoặc `backup create`
replay đúng bound theo scope/time quan sát; retry trả `EXACT_DUPLICATE`. Đây là
recovery bounded cho đúng hai files M10, không phải transaction toàn runtime,
mô phỏng power-loss hay guarantee multi-host.

**Cập nhật M10 canary registry/state journal (2026-09-13):** canary grant
giờ journal trước immutable `CANARY_GRANT` registry và mutable `mission-state`
binding/counters. Fault trước artifact, sau artifact, sau state, hoặc sau khi
journal visible nhưng trước directory sync giữ journal; Bot mới không cho
`status`/`m10-resolve` đọc half-commit. Seam visible-journal trả
`PUBLISHED_RECOVERY_REQUIRED`, còn writer có lock và `backup create` replay đúng
grant rồi retry `ACK`. Đây chỉ là recovery local của grant registry/state,
không chứng minh transaction toàn runtime, power-loss, multi-host locking,
executor hay production authority.

**Cập nhật atomic JSONL write-path guard (2026-09-13):** ACCESSTRADE outcome/
receipt và M10 cost-bound index đọc JSONL cũ qua stable regular-file reader,
reject target symlink hoặc name replacement sau read. Với target chưa tồn tại,
temporary snapshot publish qua non-overwriting hard-link thay vì rename đè một
tên vừa xuất hiện. Regression giữ external bytes không đổi cho symlink ban đầu
và symlink thay sau read. Đây chỉ là local path/publish boundary; TOCTOU rộng,
power-loss, multi-host và dữ liệu ACCESSTRADE thật vẫn ngoài phạm vi.

**Cập nhật immutable artifact publisher path guard (2026-09-13):** retry của
`writeNewJSON` nay đọc output tồn tại qua shared stable regular-file reader,
không còn dùng `os.ReadFile` sau một `Lstat`. Publisher cũng từ chối immediate
parent là symlink trước khi tạo temporary file. Regression dùng đúng publisher, chặn cả
symlink có bytes giống artifact mong đợi và replacement sang external symlink
sau khi descriptor đã mở; external bytes không đổi. Đây là local output-path
guard cho immutable artifacts, không phải snapshot atomic trước uncooperative
writer, transaction đa-file, power-loss hay guarantee multi-host.

**Cập nhật mutable runtime-root path guard (2026-09-13):** trước mọi learner
mission command có thể mutate state, runtime root phải là direct non-symlink
directory trước khi lock, recovery journal hay `mission-state` có thể được tạo.
Regression chạy CLI `mission init` với runtime symlink và kiểm cả external
`mission-state` lẫn `.mission.lock` không được tạo. Đây chỉ là guard local cho
mutable command root; portable inputs, ancestor-path TOCTOU, crash/power-loss
và multi-host vẫn ngoài phạm vi.

**Cập nhật runtime-root read guard (2026-09-13):** các reader `status`,
`m10-resolve`, `m11-resolve`, `m11-recovery-export` và old source của
`m11-recovery-admit` nay reject direct state-root symlink trước khi lấy runtime
gate hay đọc journal/registry. Regression tạo runtime hợp lệ bên ngoài, gọi
`status` bằng alias symlink, và nhận `STATE_ERROR` thay vì lộ state. Đây vẫn là
boundary filesystem local; portable input, ancestor-path TOCTOU, crash/power-loss
và multi-host vẫn ngoài phạm vi.

**Cập nhật shared M09 approval boundary (2026-09-11):** `core/m09` hiện owns
strict `approval-record` decode (schema, unknown/duplicate field rejection)
và validation không-authorizing của proposal-only intent, policy state, human
one-time approval, link và timeline. Learner dùng boundary đó cả khi append
approval lẫn khi reload/revalidate authority; harness M09 dùng cùng decoder và
validator trước authorization. Regression chứng minh duplicate-key approval
không ghi state, và persisted policy bị đổi sang `DENY` không thể reserve sau
reload.

**Cập nhật shared M09 authorization/execution decoders (2026-09-16):**
`core/m09` hiện owns strict decoding cho cả `APPROVED_LIVE`
`execution-authorization` và M09 `execution-record`, bao gồm unknown/duplicate
field rejection, profile check và schema constraints. Harness `m09-check` và
state restore dùng các decoder này qua compatibility aliases thay vì tự định
nghĩa lại struct/parser. Đây chỉ là contract/conformance reuse; chain links,
migration, executor thật và multi-file crash proof vẫn mở, nên RP-02 vẫn
`PARTIAL`.

**Cập nhật shared M09 historical chain validator (2026-09-16):** `core/m09`
hiện kiểm tra non-authorizing toàn bộ chuỗi intent → policy → approval →
`APPROVED_LIVE` authorization → execution record, gồm exact links, profile,
timeline, authorization expiry và điều kiện `PERFORMED`. Harness `m09-check`
chỉ còn decode file rồi gọi validator chung; regression có chain hợp lệ,
broken link và performed-after-expiry reject. Đây chưa phải executor thật,
migration của mọi runtime path hay multi-file crash/power-loss proof.

**Cập nhật shared M10 artifact decoders (2026-09-16):** `core/m10` hiện owns
decoder cho canary grant, trusted cost-bound, gate, authorization và
execution-record mà mission harness gọi trực tiếp. Decoder execution-record
schema-only được tách khỏi validator terminal no-side-effect, vì vậy record
`SUCCEEDED/PERFORMED` hợp lệ không bị loại nhầm trong khi canary profile,
unknown/duplicate field, hash, time và limit vẫn được kiểm. Harness chỉ giữ
ledger/approval compatibility boundary và các chain/time checks riêng. Đây là
contract/conformance reuse trong fixture runtime; executor thật, business
outcome, migration và crash/power-loss vẫn mở.

**Cập nhật M08 policy semantic binding (2026-09-11):** `core/m08` nay kiểm
decision/risk/timeline có thể xác minh chỉ từ immutable intent và policy;
learner `bind` strict-decode policy trước khi ghi state, còn M09/harness dùng
cùng validator khi revalidate approval. Một policy schema-valid nhưng
`ALLOW/RISK2` bị từ chối trước khi tạo state. Đây không chứng minh policy
context provenance hoặc thay executor-side reevaluation, nên RP-02 vẫn
`PARTIAL`.

**Cập nhật core M11 recovery-admission cardinality (2026-09-11):** canonical
graph nay sở hữu một admission cho mỗi new lease và mỗi cặp prior runtime/human
resolution. Regression bắt cả hai conflict trên graph hợp lệ, nên generic
registry không lệch learner/backup boundary. Đây không chứng minh recovery
atomic, power-loss hay multi-host, nên RP-07 vẫn `PARTIAL`.

**Cập nhật recovery admission approval guard (2026-09-15):** canonical M11
graph không còn coi `RECOVERY_ADMISSION` là lease draft. Nó phải resolve exact
`PRODUCTION_LEASE_APPROVAL` của new lease, và thời điểm human review của
admission phải sau review của approval. Core regression dựng registry integrity
hợp lệ chỉ gồm lease + admission để chứng minh reject; regression loader Bot
đọc đúng file JSONL đó và cũng reject trước khi runtime dùng artifact. Điều này
không chứng minh source lineage giữa hai runtime, atomic recovery, power-loss
hay multi-host safety, nên RP-07 vẫn `PARTIAL`.

**Cập nhật core M10 registry graph (2026-09-11):** M10 immutable graph được
canonicalize vào core và learner registry dùng trực tiếp implementation này.
Core regression reject authorization orphan trên một graph còn schema-valid.
Ledger/reservation state, executor, power-loss và multi-host proof vẫn ngoài
scope, nên RP-03 vẫn `PARTIAL`.

**Cập nhật core M10 duplicate registry entry (2026-09-11):** canonical graph
reject duplicate envelope cùng artifact kind/ID trước khi tạo lookup map, nên
mọi caller dùng core không thể che entry bất biến bằng overwrite map. Đây chỉ
là graph-integrity guard offline; ledger transaction, executor, power-loss và
multi-host proof vẫn mở.

**Cập nhật core M10 execution cardinality (2026-09-11):** canonical graph chỉ
chấp nhận một terminal no-side-effect record cho mỗi governed authorization;
record ID khác không thể bypass lineage này. Đây chỉ là graph-integrity guard
offline; reservation ledger, executor, power-loss và multi-host proof vẫn mở.

**Cập nhật core M10 identifier integrity (2026-09-11):** sau khi resolve
grant/cost, canonical graph tự tính `gate_id` từ immutable parent links và
ledger snapshot; sau khi resolve gate, graph tự tính `authorization_id` từ
gate/intent/executor/time/idempotency. Regression thử sửa một gate ID và nối
authorization theo ID giả, cũng như sửa trực tiếp authorization ID, trong khi
mọi field/schema/liên kết khác vẫn hợp lệ; cả hai bị chặn trước khi graph được
chấp nhận. Đây chỉ là deterministic immutable graph guard offline; reservation
ledger, executor, multi-file crash/power-loss và multi-host proof vẫn mở.

**Cập nhật core M10 execution identity (2026-09-12):** sau khi resolve exact
immutable authorization, canonical graph tự tính `execution_id` từ
authorization, attempt time, terminal status và reason. Regression thay riêng
ID trong một terminal record còn schema-valid và còn tất cả link đúng; graph
chặn record đó. Đây chỉ là deterministic immutable graph guard offline;
reservation ledger, executor, multi-file crash/power-loss và multi-host proof
vẫn mở.

**Cập nhật core M11 identity integrity (2026-09-12):** `core/m11` nay owns
deterministic gate/authorization/execution ID builders mà learner previously
đã dùng riêng. Sau khi resolve lease, health, cost và ledger, graph tự tính
`gate_id`; rồi tự tính `authorization_id` từ gate/executor và `execution_id`
từ authorization. Learner gọi đúng helpers core; fixture journal dùng identity
canonical thay ID tự đặt. Regression forge từng ID trong graph còn
schema-valid/với link downstream đã rewrite và cả ba bị chặn. Đây chỉ là
immutable graph guard offline; ledger transaction, executor, multi-file
crash/power-loss và multi-host proof vẫn mở.

**Cập nhật M11 authorization time identity (2026-09-14):** production
`AuthorizationID` nay digest `gate_id`, `executor_id` và `authorized_at`.
Trước đó cùng gate/executor nhưng thời điểm authorization khác nhau có ID như
nhau trong khi payload/expiry khác, khiến registry append-only chỉ có thể từ
chối artifact sau như collision. Core graph và learner lifecycle dùng một
builder time-bound; regression core và compiled-Bot backup smoke tạo hai
authorization cùng gate/executor khác thời điểm, yêu cầu ID khác nhau và chặn
reservation còn lại sau khi ledger đã advance; mutation CI bỏ chính guard
canonical phải làm regression fail. Điều này không cấp thêm reservation hay
execution authority, và không là proof clock trust, ledger transaction,
power-loss, multi-host hay business outcome; RP-07 vẫn `PARTIAL`.

**Cập nhật M10/M11 registry publish uncertainty (2026-09-14):** sau khi một
line immutable registry đã file-sync và tên file còn đúng nhưng parent-directory
sync không xác nhận được, adapter trả typed publish uncertainty thay vì lỗi
append thường. `m10-gate` và `m10-authorize` bàn giao
`PUBLISHED_RECOVERY_REQUIRED` kèm artifact đã resolve được và không ghi
portable output; M11 registry cùng CLI `m11-activate`, `m11-ledger-init`,
M11 gate, authorization, reservation, offline `m11-evaluate`,
`m11-close-cycle`, reviewed `m11-reconcile` và non-authorizing
`m11-recovery-admit` bàn giao record/ledger/evaluation/cycle, transition
reconciliation hoặc admission đã visible để resolver hoặc canonical ledger-head
kiểm tra trước exact retry. Regression dùng seam sau file sync, trước parent
sync, và kiểm các đường M10, M11 registry, activation, ledger-init, gate,
authorization, reservation, evaluation, cycle, reconciliation và cross-runtime
admission. Đây là
disclosure/recovery local có giới hạn, chưa chứng minh durability qua power
loss, transaction nhiều file, bao phủ mọi M11
lifecycle envelope, multi-host, execution hay business outcome; RP-03/RP-07
vẫn `PARTIAL`.

**Cập nhật M10 cost-bound JSONL boundary (2026-09-11):** `m10-cost-register`
canonicalize JSON đã decode trước khi append, nên input pretty-printed không
thể tách thành nhiều line registry; append dùng chung file+directory sync trước
ACK. Regression đưa input pretty JSON thật, kiểm chỉ thêm một line và resolver
load lại đúng hash/ID. Đây không chứng minh power-loss hoặc multi-host
transaction, nên vẫn `PARTIAL`.

**Cập nhật M06 missing-field projection (2026-09-11):** canonical synthetic
offer builder giờ có regression cho `price:null` và `commission_rate` vắng mặt:
cả hai projected field phải `value:null`, `state:missing`, `claim_kind:unknown`.
Không được suy diễn giá/hoa hồng từ field khác. Đây chỉ là profile fixture
canonical; selected-source parser/profile vẫn là phạm vi riêng và `PARTIAL`.

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
- **RP-03 M10 canonical output disclosure (2026-09-14):** `m10-gate` và
  `m10-authorize` vẫn append/validate canonical registry trước portable output
  để output path không là commit boundary. Nếu output immutable đã tồn tại với
  bytes khác, CLI trả non-zero
  `CANONICAL_ARTIFACT_REGISTERED_OUTPUT_UNAVAILABLE` kèm chính gate hoặc
  authorization durable; caller có thể dùng ID đó với `m10-resolve` hoặc xuất
  exact artifact sang path sạch. Regression gọi Bot thật cho cả hai loại,
  resolve ID trả về và chứng minh bytes của output cũ không đổi; retry
  authorization sang path sạch thành công. Lỗi sync parent sau publish vẫn là
  `PUBLISHED_RECOVERY_REQUIRED`. Đây chỉ là disclosure/recovery envelope local,
  không chứng minh atomicity nhiều file, power-loss, executor hay business
  outcome; RP-03 vẫn `PARTIAL`.
- **RP-03 M10 gate evaluation identity (2026-09-14):** `GateID` canonical
  nay hash cả `evaluated_at`. Trước đó hai đánh giá hợp lệ khác thời điểm nhưng
  cùng grant/cost/intent/policy/ledger có cùng ID và bytes khác nhau, nên
  registry append-only chỉ có thể từ chối lần đánh giá sau như conflict.
  Core graph recompute cùng time-bound identity; regression core và learner Bot
  tạo hai gate cùng snapshot cách nhau một giây, yêu cầu ID khác nhau và cả hai
  resolve trong registry/graph. Không suy điều này thành clock trust, execution authority, transaction
  nhiều file, power-loss hay business outcome; RP-03 vẫn `PARTIAL`.
- **RP-03 M10 registry foundation:** state directory nay có registry append-only
  `m10-artifacts.jsonl`; core canonicalize/hash envelope và learner chỉ ACK
  `CanaryGrant`, trusted cost-bound, gate, authorization hoặc cancellation
  record sau khi registry validate toàn bộ link grant → cost/gate → authorization
  → record. `m10-resolve` chỉ đọc artifact canonical theo kind/ID/(tùy chọn)
  content hash. Gate/authorization hợp schema nhưng không có entry chính xác bị
  smoke BR-16a từ chối. Registry chưa có reservation-to-attempt link, execution
  result/outcome/EffectRef, migration/restore graph hoặc fault injection đầy
  đủ, vì vậy không đóng RP-03/RP-06/RP-07.
- **RP-04 foundation (historical, superseded by the M06/M07 updates below):**
  thay đổi sau mốc head ở trên đã thêm learner resolver read-only dùng chung
  cho M07 context và M08 intent: record phải resolve đúng một lần từ history
  và replay `MATCH`. M06 watcher handoff và HTTP GET cũng resolve lại record
  từ store sau append trước khi ACK/return artifact. Các giới hạn cũ về core
  builder, n8n adapter và proposal resolver đã được các cập nhật tiếp theo
  triển khai; RP-04 vẫn PARTIAL chỉ vì generic source/deployment evidence và
  coverage rộng hơn còn mở.
- **RP-04 M06 shared fixture path:** `core/m06` nay owns strict
  `br13-offer-fixture/v1` decoding, timestamp/identity/provenance normalization
  và M00 packet construction. Learner CLI local/pinned fetch cùng n8n endpoint
  `/v1/m06/fixture-import` dùng profile này; endpoint append rồi resolve/replay
  canonical record trước ACK. Blueprint không còn GET/parse/hash/build history
  bằng JavaScript. Unit test cover adapter APPENDED/EXACT_DUPLICATE/replay và
  reject không mutate history. Đây chỉ là fixed synthetic profile; generic
  source profile, n8n operated execution và full shared HistoryRecord type vẫn
  còn mở, nên RP-04 chưa đóng.
- **Selected source ACCESSTRADE Shopee Smartlink (2026-09-13):** theo lựa
  chọn rõ ràng của chủ repo, `core/m06` có profile riêng
  `accesstrade-shopee-campaign-capture/v1`. Profile fixed host/path campaign
  Shopee Smartlink, chỉ nhận `GET`/200/no-redirect và sanitized capture gồm
  title, merchant, category, status/period label, time, correlation ID và hash
  provenance. Nó reject extra/raw HTML, credential, report, URL thay thế và
  claim số commission/EPC/CVR. Builder vẫn dùng M00/History canonical chung;
  mọi capture input được gắn `evidence_kind:synthetic` vì runtime không thể
  độc lập chứng minh transcript là page thật;
  price/commission được ghi `unknown`/`missing`, nên result là
  `GET_MORE_DATA` thay vì ranking. CLI/loopback endpoint và blueprint n8n
  inactive/manual đều handoff capture tới adapter; test thực reject claim
  commission bịa qua M07. Đây là implementation/test offline của contract
  selected-source, không phải fetch authenticated page, n8n operated run,
  approval campaign, Smartlink, report/outcome hay execution authority. Xem
  `docs/architecture/BR-13-ACCESSTRADE-SHOPEE-SELECTED-SOURCE.md`.
- **RP-04 canonical fingerprint + M08 field links:** M06 dùng canonical JSON
  fingerprint cho body có cấu trúc (key order không tạo observation mới), còn
  raw byte hash giữ riêng cho HTTPS fixture pinning. CLI/HTTP adapter retry
  cùng body reordered là `EXACT_DUPLICATE`; price/commission thiếu tạo field
  `missing` và HistoryRecord vẫn replay `MATCH`. M08 resolve cùng canonical
  record rồi dùng đúng field IDs mà M07 context công bố; ID tự dựng và record
  `DRIFT` đều bị reject. Regression chạy implementation learner thật, không
  dùng parser Python thay thế. Generic source profile và operated deployment
  evidence vẫn mở; fixture và selected-source sanitized paths đã có n8n engine
  regression nên không được mô tả lại là “chưa có n8n execution”.
- **RP-05 foundation:** core/learner M07 nay kiểm output thực: chỉ
  `HUMAN_REVIEW`/`ABSTAIN`, claim/evidence/value và `answer`/`claim.text` phải
  là render deterministic; prose tự do, ID dư/giả, quyền ghi và `tool_calls`
  tự khai bị reject. `m07 register-tool-result` preflight/validate response,
  ghi artifact immutable có trace hash rồi `m07 validate` resolve lại hash và
  record binding trước khi body `unknown` được cite. Test learner trực tiếp,
  regression M07 và smoke shared workspace tạo/cite trace rồi reject trace giả.
  `m07 register-proposal` persist raw validated output/proposed action với
  canonical digest và record binding; M08 agent intent/policy resolve lại
  artifact và reject target/parameters đổi ngoài proposal. Canonical
  tool-evidence store/transport seam và blueprint n8n adapter path đã có; RP-05
  vẫn PARTIAL vì provider/source operated evidence, transport ngoài fixture và
  broader parity chưa có, không vì thiếu wiring cơ bản.
- **RP-05 HTTP adapter foundation:** watcher có endpoint loopback
  `/v1/m07/register-tool-result`, `/v1/m07/validate` và
  `/v1/m07/register-proposal`. Chúng resolve history canonical, lưu artifact
  trace/proposal bất biến dưới store dẫn xuất từ history và ACK sau validation;
  validate chỉ resolve tool trace bằng ID đã persist. Test handler cover ACK,
  validate, proposal persistence và trace ID giả. Blueprint n8n đã chuyển flow
  sang preflight/full-response/no-redirect/register/context/validate/proposal
  endpoints và static validator kiểm wiring. Disposable n8n engine đã chạy
  policy reject, registered trace/context, grounding và proposal persistence
  với loopback model stub; điều đó vẫn không là deployment/provider operated
  evidence.
- **RP-05 adapter-owned transport:** watcher thêm `fetch-and-register`; n8n
  không còn gọi remote HTTP trực tiếp. Adapter validate registry trước fetch,
  resolve DNS trước request và mỗi dial, reject non-public/mixed IP, proxy,
  redirect và response vượt 256 KiB; timeout đến từ registry. Unit test cover
  private/CGNAT/mixed DNS và body quá cỡ. Disposable n8n import/run và
  sink-failure fail-closed đã có; controlled public-source integration vẫn bị
  loại khỏi phạm vi offline nên không đóng RP-05.
- **RP-05 blueprint persistence boundary:** blueprint M07 dùng adapter URL và
  registry review cố định, không lấy hai policy boundary này từ event input.
  Trace tool, context canonical, grounding `HUMAN_REVIEW` có draft và proposal
  persistence đều phải nhận ACK ở node riêng trước khi node sau chạy. Learner
  HTTP regression kiểm proposal ACK và xác nhận output không grounded không
  thể sửa artifact proposal đã persist. Đây vẫn là fixture/offline evidence:
  có n8n engine và credential loopback disposable, nhưng chưa có provider
  credential hoặc provider integration được vận hành.
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

**Cập nhật RP-02 shared M08 policy context decoder (2026-09-18):**
`core/m08` hiện sở hữu full policy-context decoder và semantic validator.
Mission-runtime dùng decoder này; learner giữ input contract rút gọn nhưng gọi
cùng validator sau khi bind decision/evidence/proposal IDs. Regression bao phủ
missing/null/duplicate/unknown/case-variant fields, RFC3339 time, duplicate
IDs, unknown risk và blank idempotency. Local worktree không có Go executable,
nên CI hosted vẫn là acceptance gate; RP-02/R06/R07 chưa đóng do còn hash
version/migration, provenance/conformance và external execution blockers.
Record: `docs/architecture/EVIDENCE-RP02-SHARED-M08-POLICY-CONTEXT-20260918.md`.

Hosted verification is complete: PR #417 passed all 13 checks in Curriculum CI
run `35314331386` and Mission Agent Path CI run `35314331378`, including the
learner race, Windows runtime and mission-runtime test/vet jobs. This records
offline/fixture acceptance for the shared decoder seam; RP-02/R06/R07 remain
PARTIAL for shared conformance, hash-version/migration, provenance/persistence
and external execution requirements.

**Cập nhật RP-02 shared M08 conformance table (2026-09-18):**
`core/m08/conformance.go` now owns one scenario table with expected decision,
risk, reason and non-authorizing flags. Core, mission-runtime and learner
regressions consume that table; learner cases that require an independently
trusted missing decision/evidence registry are marked as an explicit parity
boundary because the learner policy-input contract derives those links from
the bound intent. Curriculum CI run `35315240076` and Mission Agent Path CI
run `35315240089` passed all 13 checks for this change. RP-02/R06/R07 remain
PARTIAL pending hash-version/migration, persistence/provenance and external
execution evidence.

**Cập nhật RP-02 intent hash version boundary (2026-09-18):** `core/m08`
formalizes the existing `sha256:` prefix as V1. Decode, policy evaluation and
the mission harness reject an unknown prefix with `UNSUPPORTED_HASH_VERSION`
and never reseal the old intent in place. Hosted Curriculum CI run
`35316020312` and Mission Agent Path CI run `35316020309` passed all 13 checks.
This is version recognition/fail-closed evidence only; migration, approval
review, exact restart persistence and provenance remain open, so RP-02/R06/R07
stay PARTIAL.

**Cập nhật RP-02 exact-number restart conformance (2026-09-18):** learner và
mission-runtime đều ghi/đọc lại intent có `9007199254740993`, giữ
`json.Number` và xác nhận hash không đổi sau reload. Curriculum CI run
`35316418613` và Mission Agent Path CI run `35316418619` đã PASS toàn bộ 13
checks. Đây là local fixture/restart evidence bounded; crash/power-loss,
multi-process persistence and distributed storage remain open, nên R07 và
RP-02 vẫn PARTIAL.

**Cập nhật RP-02 persisted M07 proposal provenance (2026-09-18):** learner
resolver revalidate digest, grounding và canonical `record_id` khi bind
`proposal_ref`; regression reject proposal có digest hợp lệ nhưng provenance
khác và không ghi M08 intent. Curriculum CI run `35317820062` và Mission Agent
Path CI run `35317820048` đã PASS toàn bộ 13 checks. Đây là bounded local
resolver/read-only evidence; cross-store transactionality, distributed
persistence and live execution remain open, nên RP-02/R06/R07 vẫn `PARTIAL`.

Marker: `Cập nhật RP-02 shared M08 policy context decoder`.

**Nghiệm thu:** R06/R07 vẫn `PARTIAL`; output vẫn đúng schema và
proposal-only trong boundary offline hiện tại. **Migration:** báo version/hash
không hỗ trợ; migration có lệnh riêng, backup và human review; không rewrite
approval cũ để khớp hash mới.

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

**Cập nhật RP-04 shared canonical evidence context (2026-09-18):**
`core/canonical` nay tạo envelope `canonical-evidence-context/v1` dùng chung
cho M07/M08 từ canonical history observations. Aggregate IDs phải khớp đúng
recorded observations; field IDs/provenance được lấy lại từ raw M00 projection,
reject malformed/forged/duplicate/collision IDs và giữ nguyên missing/null.
Learner M07 không còn tự dựng field evidence riêng. Đây vẫn là bounded
offline/fixture/read-only evidence; resolver chỉ nhận record resolve đúng một
lần và replay `MATCH`, còn provider/live execution, deployment, distributed
locking, power-loss và multi-file atomicity vẫn mở. Marker:
`RP-04 shared canonical context`.

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

**Cập nhật shared M07 harness boundary (2026-09-16):** mission-runtime không
còn giữ parser/policy M07 độc lập. `ToolSpec`, tool request và Agent output
dùng type/decoder của `core/m07`; registry, strict output contract,
claim/value grounding và read-only tool policy đều chạy qua implementation
chung. Eval/testdata được nâng lên output `HUMAN_REVIEW` có claims/evidence/
authority/write boundary; tool request chưa có adapter trace vẫn bị chặn. Test
negative chạy qua harness thật và bắt unknown tool, write/host/port/userinfo,
forged known-ID value, duplicate/null/trailing fields và allowlist mơ hồ.
Đây là local fixture evidence, không đóng provider/live executor/business
outcome/pilot/deployment hoặc multi-host/power-loss.

**Cập nhật shared M10 historical-chain validator (2026-09-16):**
`m10-chain-check` giờ strict-decode tại mission boundary rồi chuyển toàn bộ
kiểm tra lineage intent/policy → grant/approval → cost/ledger → gate →
authorization → execution sang `core/m10.ValidateHistoricalChain`. Core sở hữu
link, profile/scope, thứ tự thời gian, snapshot ledger, duplicate/block và
budget/overflow checks; harness không còn giữ bản sao cross-artifact validator.
Các regression chain hiện tại vì thế chạy đúng implementation dùng chung. Đây
chỉ là historical canary fixture audit, không cấp quyền, không gọi executor và
không chứng minh provider, business outcome, crash/power-loss hay multi-host.

**Cập nhật RP-05a proposal metadata (2026-09-18):** `RegisteredAgentProposal`
giờ persist và resolver kiểm tra lại `validation_result`, `validation_version`,
`authority_ceiling` cùng provenance của canonical context: `record_id`,
`decision_id` và evidence IDs đã được model cite. Learner, watcher và restore
đều truyền decision identity vào cùng resolver; proposal vẫn chỉ là
`HUMAN_REVIEW` proposal, không cấp approval hay execution authority. Đây vẫn là
bounded offline/fixture evidence; provider/model operated evidence và n8n
end-to-end vẫn mở. Marker: `RP-05a proposal metadata`.

**Baseline sync after PR #427 (2026-09-18):** PR #427 đã squash-merge vào
`main` tại `dcda7894ad72227c588de0f7077dc18cdab92fea` từ implementation head
`0f9f3aaee9b0e41afd31eaa448df0c59bd7b0504`. Final head đạt đủ 13 hosted
checks, gồm Windows runtime, learner race, Go vet, backup mutation và n8n
regression. Evidence post-merge được ghi tại
`docs/architecture/EVIDENCE-PR427-POST-MERGE-20260918.md`; RP-05 vẫn
`PARTIAL`, provider/model operated evidence và live execution vẫn mở. Marker:
`Baseline sync after PR #427`.

**Cập nhật RP-05b tool trace provenance (2026-09-18):** registered tool trace
giờ lưu riêng `request_id` do adapter tính từ canonical record/tool request và
`content_digest` tính từ canonical JSON body (ổn định qua serialize/restore). Nếu caller gửi metadata forged,
registration/restore từ chối; evidence body gắn `SubjectID` với request ID,
trong khi `TraceID` vẫn bao phủ toàn bộ result. Trace cũ thiếu metadata vì vậy
fail closed thay vì được tự động nâng cấp. Đây vẫn là bounded offline/fixture
evidence; không claim provider, live execution hay external readiness.
Marker: `Cập nhật RP-05b tool trace provenance`.

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

**Cập nhật recovery-admission source gate (2026-09-11):** sau khi command đã
giữ gate runtime mới, `m11-recovery-admit` giữ thêm runtime gate của **runtime
cũ** xuyên suốt handoff/registry validation. Old writer đang chạy trả fail-closed
trước khi admission đọc handoff input hoặc ghi artifact mới; regression giữ gate
cũ rồi xác nhận new registry không đổi. Đây là local cross-runtime exclusion,
không thay distributed lock, multi-host coordination hay power-loss recovery.

**Cập nhật M11 read gate (2026-09-11):** `m11-recovery-export`, `m11-resolve`
và `status` giữ runtime gate của nguồn xuyên suốt journal preflight và lifecycle
read. Writer local đang hoạt động trả `BUSY`: recovery export không tạo handoff,
resolver/status không lộ snapshot đua tranh. Đây chỉ là exclusion trên local
filesystem; không chứng minh read transaction đa-host, distributed lock hay
crash/power-loss recovery.

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

**Cập nhật M11 ledger outcome continuity (2026-09-11):** canonical append-only
ledger history phải giữ mọi prior `outcome_links`, successful idempotency key và
reconciliation resolution ID. Core regression append post-outcome ledger rồi xóa
link outcome ở entry tiếp theo, và graph reject. Đây giữ audit lineage của state
hiện có; không biến journal thành transaction đa-file hay chứng minh outcome
business bên ngoài fixture.

**Cập nhật M11 artifact cardinality (2026-09-11):** canonical graph reject mọi
duplicate `(artifact_kind, artifact_id)` trước khi decode/link resolution. Core
regression thêm lại activation y hệt và bị reject, nên map overwrite không thể
ẩn duplicate input khi validator được dùng ngoài learner registry. Đây là
integrity guard offline, không thay thế locking hay migration/version policy.

**Cập nhật M11 ledger outcome execution link (2026-09-11):** mỗi outcome link
trong canonical ledger phải resolve execution thật cùng lease lineage và không
được có `observed_at` trước attempt. Core regression gắn outcome vào execution
không tồn tại rồi reject, nên forged link không thể giải phóng pending budget.
Đây là fixture graph validation, không xác thực business outcome external.

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

**Cập nhật M10 24-process reservation barrier (2026-09-14):** regression mới
khởi tạo một fixture authority thực còn hiệu lực, cap execution/cost bằng 1,
rồi dùng barrier test-owned để giải phóng đồng thời 24 process chạy binary Bot
đã build riêng vào `m10-reserve-authorization`. Chính binary vận hành đọc wall
clock thông thường; barrier chỉ nằm trong test wrapper, không phải override
runtime hay authority. Đúng một lệnh trả `RESERVED`; mọi process khác trả
`BUSY`, `BUDGET_DENIED` hoặc `REJECTED` với lý do authorization đã có
reservation. Sau đó loader kiểm `ExecutionsUsed=1`,
`CostUsedMinor=1` và đúng một governed reservation, rồi một process Bot mới
chạy `status` để replay state đã persist. Điều này thay claim cũ sai rằng smoke
BR-16a đã có test 24 process. Nó chứng minh lock cục bộ/cap accounting, không
chứng minh distributed lock, kill/power-loss hoặc transaction đa-file.

**Cập nhật M10 24-process fixture-outcome barrier (2026-09-14):** sau một
reservation M10 đã được ghi và một execution `CANCELLED`/`NOT_PERFORMED` đã
canonicalize, barrier test-owned giải phóng 24 binary Bot riêng cùng ghi các
outcome ID khác nhau nhưng cùng `MACHINE_EXECUTION` effect. Đúng một lệnh trả
`APPENDED`; phần còn lại chỉ `BUSY` hoặc `CONFLICT`. Loader đọc đúng một outcome
trong JSONL, Bot process mới replay trạng thái, rồi backup v3/restore sang
runtime trống vẫn giữ đúng một outcome đó. Đây chứng minh local single-writer
cardinality trên command/store/restore path; không chứng minh ingestion business
outcome, crash/power-loss, transaction đa-file hay multi-host.

**Cập nhật M10 visible fixture-outcome append uncertainty (2026-09-14):** nếu
shared JSONL writer trả lỗi sau khi exact outcome line đã hiện diện, command
`m10-outcome` re-read bằng canonical loader. Khi line đó hợp lệ và đúng payload,
nó trả `PUBLISHED_RECOVERY_REQUIRED` thay vì che trạng thái này dưới
`STORE_ERROR`; exact retry sau đó trả `EXACT_DUPLICATE`, không thể append outcome
cạnh tranh cho cùng execution. Regression gọi command thật với seam chỉ nằm
trong test, rồi xác minh store replay được. Đây là acknowledgement/retry boundary
của một fixture JSONL file; không chứng minh fsync/power-loss, atomic transaction
nhiều file, business outcome hay multi-host.

**Cập nhật M11 visible fixture-outcome append uncertainty (2026-09-14):** M11
đã giữ recovery journal trước ledger/outcome transition, nhưng lỗi ACK sau khi
exact JSONL outcome line xuất hiện trước đây vẫn bị command báo `REJECTED`.
`appendM11FixtureOutcome` nay canonical-reload outcome store; khi record đúng
payload đã hiện diện, `m11-outcome` trả `PUBLISHED_RECOVERY_REQUIRED` (non-zero)
kèm outcome và post-ledger, còn journal giữ runtime ở `RECOVERY_REQUIRED`.
Bot process mới không được status/resolve cho tới khi retry exact dưới runtime
gate hoàn tất journal và trả `EXACT_DUPLICATE`; regression kiểm đúng một outcome
trước/sau retry. Đây là acknowledgement/recovery boundary local cho một JSONL
line và journal đã visible, không chứng minh fsync/power-loss, transaction nhiều
file, live executor, business outcome hoặc multi-host.

**Cập nhật M11 fixture-outcome execution cardinality (2026-09-15):** writer
và backup graph đã cấm hai outcome cùng `MACHINE_EXECUTION`, nhưng loader JSONL
chỉ từng cấm duplicate `outcome_id`. Loader nay giữ thêm `effect_id` đã thấy và
fail closed khi hai outcome hợp lệ khác nhau trỏ tới cùng execution. Regression
tạo đúng hai JSONL records hợp lệ cùng execution trên fixture runtime và chứng
minh loader reject trước evaluate hoặc recovery. Đây là fixture-store cardinality
local; không là business outcome, atomic recovery, power-loss hay multi-host
proof.

**Cập nhật M11 recovery-handoff strict decoder (2026-09-15):**
`m11-recovery-admit` nhận handoff qua portable input, nhưng trước đây dùng
`json.Unmarshal` vào map nên duplicate key có thể bị collapse trước khi so với
lineage canonical của runtime đã STOP. Đường này nay dùng decoder strict chung
trước khi admission có thể đọc/đăng ký artifact. Regression gọi CLI recovery
admission thật với duplicate `profile`, yêu cầu `REJECTED` và registry runtime
mới giữ byte-identical. Đây chỉ là hardening JSON cục bộ; không thay thế human
review, live execution, crash/power-loss durability hay multi-host recovery.

**Cập nhật M06 history-handoff strict decoder (2026-09-15):** M06 handoff
nhận `HistoryRecord` qua CLI và `/v1/history/append`, nhưng parse thường có thể
collapse duplicate key trước khi hash được recompute và record được append. Hai
entrypoint nay cùng strict-decode raw bytes trước canonical append. Regression
gửi duplicate `record_id` qua cả CLI và HTTP, yêu cầu `INVALID_SCHEMA` và không
được tạo history. Đây là boundary input local; không là n8n/provider operated
proof, crash/power-loss durability, transaction đa-file hay multi-host safety.

**Cập nhật M11 outcome-journal cleanup uncertainty (2026-09-14):** Sau khi
ledger head và exact fixture outcome đã canonical, lỗi khi dọn
`m11-outcome-journal.json` không còn được trả là `REJECTED`. Runtime báo
`PUBLISHED_RECOVERY_REQUIRED`, giữ journal nếu cleanup chưa bắt đầu, và Bot mới
chặn `status` cho đến exact retry dưới runtime lock. Retry chỉ hoàn tất cleanup
và trả `EXACT_DUPLICATE`; regression dùng seam ngay trước remove, xác minh một
outcome và post-ledger duy nhất cả trước lẫn sau restart/retry. Đây là một
acknowledgement boundary local; không chứng minh fsync/power-loss, transaction
đa-file, live executor, business outcome hoặc multi-host.

**Cập nhật M11 24-process reservation barrier (2026-09-14):** shared smoke
M00–M11 clone đúng runtime đã có M11 lease/approval/activation/health/cost và
ALLOW gate, rồi đăng ký 24 authorization riêng cùng trỏ vào pre-ledger cap=1.
Barrier test-owned chỉ release sau khi cả 24 Bot binary sẵn sàng. Đúng một
`m11-reserve-authorization` trả `APPENDED`; các contender còn lại chỉ có thể
trả `BUSY` hoặc `REJECTED` vì runtime gate/gate snapshot đã stale. Exact retry
của winner là `EXACT_DUPLICATE`, tất cả loser bị reject sau commit, và M11
ledger head phải giữ đúng một pending execution/counter bằng 1. Đây là chứng
cứ local cross-process cho lifecycle M11 thật, không chứng minh locking đa
host, kill/power-loss hay transaction nhiều file.

**Cập nhật M11 24-process outcome barrier (2026-09-14):** shared smoke cũng
clone runtime ngay sau một FAILED fixture execution rồi giải phóng đồng thời 24
Bot binary cùng mang một outcome immutable và cùng predecessor ledger. Đúng một
`m11-outcome` được `APPENDED`; các contender khác chỉ có `BUSY` hoặc
`EXACT_DUPLICATE`. Sau barrier, tất cả exact retry đều trả `EXACT_DUPLICATE`,
outcome JSONL có đúng một record và ledger head có đúng một `outcome_link`,
`pending_outcomes=0`. Đây chứng minh command path serialize cặp outcome-store /
post-ledger trong điều kiện local cross-process bình thường; không chứng minh
atomicity khi crash/power-loss, transaction đa-host hay durable multi-file
commit.

**Cập nhật M10 reservation commit-fault cap continuity (2026-09-14):** test
in-process inject lỗi sau temporary-file `fsync` và ngay trước `writeJSONAtomic`
rename của một governed
reservation cap=1. Lệnh trả `STORE_ERROR`, state canonical giữ byte-identical
và không còn temporary state file. Sau khi bỏ seam, binary Bot mới chạy bằng
wall clock bình thường reserve được đúng một lần từ cap chưa tiêu; attempt thứ
hai bị `REJECTED`, counters và reservation persisted vẫn khớp ACK này. Đây là
failure seam local tại một rename boundary, không phải kill/power-loss proof,
không cover mọi write/sync syscall, transaction đa-file hay distributed lock.

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

**Cập nhật M11 evaluation→closed-cycle cardinality (2026-09-16):** restore
graph giờ đòi mỗi `PRODUCTION_OUTCOME_EVALUATION` đã persist phải resolve tới
đúng một `PRODUCTION_CYCLE` đóng và đã kiểm tra lineage. BR-18b tạo backup thật,
xóa riêng cycle rồi cập nhật checksum/manifest; restore trả `GRAPH_FAILED` và
không publish target. Snapshot recovery-only không có evaluation vẫn hợp lệ.
Đây là reverse-cardinality guard local; crash/power-loss, atomic multi-file,
distributed/multi-host, provider, deployment, pilot và business outcome vẫn
mở.

**Cập nhật M11 restore authorization/execution lineage (2026-09-16):** BR-18b
giờ tạo bản sao checksum-valid của backup bằng learner Bot thật rồi làm lệch
`production_health_snapshot_hash` trong `PRODUCTION_EXECUTION_AUTHORIZATION`
hoặc `PRODUCTION_EXECUTION_RECORD`, trong khi artifact health và các parent
khác vẫn còn nguyên. Cả hai bản sao phải trả `GRAPH_FAILED` và không publish
restore target; positive restore/replay vẫn chạy trong cùng smoke. Đây là
cross-artifact lineage evidence của local restore path, không đóng provider,
live executor, business outcome, pilot, deployment, distributed lock hoặc
crash/power-loss/atomic multi-file proof.

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

**Cập nhật n8n CI cache và path gate (2026-09-14):** Job bắt buộc
`n8n-engine-regression` vẫn xuất hiện ở mọi PR, nhưng chỉ bootstrap runtime và
chạy hai regression engine khi diff chạm blueprint M06/M07, adapter/reader M06–
M07, runner hoặc chính workflow. Push vào `main` luôn chạy full. `node_modules`
của n8n `2.38.1`/Node 24 được cache theo OS/Node/version để cache hit không
phải resolve npm lại. Job skip chỉ nói thay đổi không liên quan, không phải
evidence n8n; thay đổi có liên quan và `main` vẫn cần engine thật.

**Cập nhật deterministic CI sharding và Go cache (2026-09-15):** coverage
trước đây nối tiếp trong một `deterministic-runtime` được giữ nguyên nhưng chia
thành sáu required jobs chạy song song: core Go/CLI, hai shard
top-level test learner Bot, `learner-bot-race`, `deterministic-quickstart`, và
`deterministic-smokes-and-mutations`. Helper shard hỏi chính `go test -list`,
gán từng tên `Test*` bằng SHA-256 ổn định vào đúng một shard không rỗng, rồi gọi
`go test` thực; test mới tự được bao phủ thay vì phụ thuộc một regex thủ công.
Tách `go test -race ./...` có build cache key riêng ra khỏi core job để kết quả
test thường/CLI trả về sớm, nhưng race vẫn là check bắt buộc trên mọi PR và push
`main`. Mỗi job Go khai báo đủ bốn `go.sum` cho setup-go cache, nên cache chỉ tái
sử dụng build và module hợp lệ theo toàn dependency set. Không shard nào bị skip
theo thay đổi: mọi PR và push `main` vẫn chạy toàn bộ Go/race, quickstart,
M00–M11 smoke và mutation proof. Mục tiêu là giảm wall-clock CI, không giảm
coverage hoặc suy cache hit thành runtime evidence.

**Cập nhật parallel smoke groups (2026-09-16):** nhóm
`deterministic-smokes-and-mutations` đã được tách thành ba required jobs chạy
song song: smoke nền tảng M00–M05, smoke M06–M07/BR-16a và backup/mutation/audit
M11. Tổng cộng workflow có tám job coverage bắt buộc; mỗi nhóm vẫn giữ setup
Go/Python, toàn bộ lệnh cũ và failure semantics. Audit kiểm tra đủ cả ba job để
không thể rút ngắn thời gian bằng cách bỏ một nhóm. Đây chỉ là tối ưu
wall-clock CI, không làm thay đổi phạm vi readiness hay biến cache hit thành
bằng chứng runtime.

**Cập nhật deterministic-runtime single-build (2026-09-16):** job giữ nguyên
toàn bộ M01/M02 CLI smoke nhưng build learner Bot một lần vào
`$RUNNER_TEMP/learner-bot`, sau đó tái sử dụng binary cho các lệnh baseline,
capture, list, replay và decision. Cách này loại compile lặp của nhiều `go run`
trong cùng job; không thay đổi test coverage, không bỏ required check và không
được coi là bằng chứng production/runtime ngoài CI.

**Merged-main verification (2026-09-16):** sau khi merge PR #344, local
re-run theo đúng single-build sequence đã PASS internal tests, contracts
test/vet, learner vet, M01/M02 smoke và readiness audit. Evidence được ghi tại
`docs/architecture/EVIDENCE-DETERMINISTIC-RUNTIME-SINGLE-BUILD-20260916.md`.
Đây là kiểm chứng fixture/local cho đường CI; không đoán wall-clock GitHub
runner và không nâng phạm vi sang provider, business outcome, pilot hay
deployment.

**Cập nhật mutation proof M11 (2026-09-12):** job `deterministic-runtime`
chạy `scripts/mutate_m11_identity_guard.py`. Script copy riêng `core` và
`contracts`, rồi lần lượt bỏ đúng từng guard canonical `GateID`,
`AuthorizationID`, `ExecutionID` trong copy và chạy lại chính
`core/m11.TestArtifactGraphAcceptsExactProductionLifecycleLinks`. Mỗi mutation
chỉ được tính PASS khi test thật FAIL ở assertion forged-ID tương ứng; anchor
không rõ, test PASS, hoặc lỗi không liên quan đều làm script FAIL. Đây là
mutation proof hẹp cho ba ID immutable M11, không bao phủ ledger,
backup/restore, blueprint, crash/power-loss, multi-host hay operated n8n.

**Cập nhật mutation proof M10 (2026-09-12):** cùng job chạy
`scripts/mutate_m10_identity_guard.py`. Script tạo copy riêng `core` và
`contracts`, lần lượt bỏ các guard canonical gate, authorization và terminal
execution ID M10, rồi gọi đúng `TestCanaryAuthorizationBindsGateWithoutExecuting`
hoặc `TestCancelledCanaryExecutionRecordHasNoSideEffect` theo ca tương ứng.
Mỗi copy chỉ PASS nếu assertion forged-ID đúng ca thất bại; lỗi compile hoặc lỗi
không liên quan fail closed. Proof này chỉ kiểm immutable graph ID M10, không
chứng minh ledger mutable, crash/power-loss, multi-host, executor hay outcome
business thật.

**Cập nhật backup source-identity mutation (2026-09-12):** cùng job chạy
`scripts/mutate_backup_source_guard.py`. Script tạo một copy riêng của learner
Bot, rồi đồng thời bỏ ba bước `Lstat → open/fstat → Lstat` trong
`readStableRegularFile`. Hai regression thực tế đổi `mission-state.json` sau
inventory khi backup, hoặc sau `verifyBackup` khi restore, thành symlink đến một
file ngoài có **cùng bytes**; bản mutation chỉ PASS nếu cả hai test thật fail tại
assertion source-symlink-swap riêng. Restore còn đối chiếu metadata của bytes
vừa đọc với manifest đã verify. Vì bytes và metadata file vẫn khớp inventory,
đây chứng minh guard identity path thay vì chỉ digest mismatch. Đây là proof hẹp
cho guard local này; không chứng minh mọi race, crash/power-loss, filesystem,
multi-host hay backup operated.

**Cập nhật stable recovery-journal reader (2026-09-12):** M10 execution và
M11 FAILED/UNKNOWN/outcome recovery parser nay đọc cùng
`readStableRegularFile` với backup/restore (`Lstat → open/fstat → Lstat`).
Regression thực thay journal đang mở bằng symlink đến file ngoài có cùng bytes;
parser phải fail trước khi decode replay plan. CI tạo Bot copy riêng, thay cả
hai reader bằng `os.ReadFile`, và chỉ PASS mutation nếu assertion swap thật
thất bại. Đây là guard path identity cục bộ, không chứng minh crash/power-loss,
filesystem phân tán, multi-host hay các runtime store khác.

**Cập nhật M07 tool-result stable reader (2026-09-12):** adapter chỉ resolve
`tool_result_id` qua path sidecar do chính adapter tính từ trace ID. Reader nay
dùng `readStableRegularFile`; regression thay file sau open bằng symlink ngoài
cùng bytes và phải reject trước grounding. CI mutation thay reader bằng
`os.ReadFile` rồi đòi assertion thật thất bại. Portable proposal input, backup
validation breadth, crash/power-loss, multi-host và selected-source evidence
vẫn ngoài phạm vi.

**Cập nhật M07 CLI portable-input stable reader (2026-09-13):** registry
policy, model output và registered tool result đi vào command M07 qua pathname
do caller cung cấp đều đọc bounded với identity ổn định trước strict decode,
grounding hay persistence. Regression gọi `register-tool-result` thật, thay
tool result sau open bằng symlink ngoài cùng bytes, và phải trả
`TOOL_RESULT_ERROR` trước khi ghi evidence artifact. Mutation CI thay wrapper
M07 bằng `os.ReadFile` trong checkout tạm và đòi assertion regression thật
fail. Đây chỉ là guard pathname local; không chứng minh provider, source live,
atomic transaction đa-file, crash/power-loss hay multi-host.

**Cập nhật general CLI portable-input stable reader (2026-09-13):** bỏ các
`os.ReadFile` trực tiếp còn lại trong `cmd/bot` cho action/outcome imports,
decision context, advisor config, M06 history handoff và fixture observations.
Report/manifest ACCESSTRADE fixture và M11 recovery handoff cũng giữ stable
identity trước decode; report vẫn giới hạn 16 MiB và manifest/general input là
16 MiB. Receipt derivation đọc canonical outcomes qua stable reader. Regression
gọi M03 `action record` thật, thay JSON action sau open bằng symlink ngoài cùng
bytes, và phải `IO_ERROR` trước append. Mutation CI thay shared wrapper bằng
`os.ReadFile` trong checkout tạm và đòi assertion thực fail. Đây là local
pathname hardening, không chứng minh capture/report thật, provider, executor,
atomic transaction đa-file, crash/power-loss hay multi-host.

**Cập nhật internal learner portable-input stable reader (2026-09-13):**
mọi đường đọc file production còn lại dưới `lab/affiliate-bot/internal` dùng
`store.ReadPortableInput`: app action/outcome, evidence import và learning
report. Reader giữ identity regular-file từ `Lstat` qua `open`/`fstat`, hai lần
đọc cùng descriptor có giới hạn 16 MiB và kiểm pathname cuối trước khi trả bytes
cho strict decode. Regression store thay file sau khi descriptor đã mở bằng
symlink ngoài cùng bytes và phải reject; mutation CI hạ wrapper về `os.ReadFile`
trong checkout tạm và chỉ PASS khi regression thật fail tại assertion đó. Nhờ
vậy không còn `os.ReadFile` production trực tiếp trong learner module. Đây vẫn
chỉ là guard pathname local: không chứng minh crash/power-loss, writer race rồi
restore bytes, multi-host hay evidence ACCESSTRADE/provider vận hành.

**Cập nhật M07 backup-sidecar stable reader (2026-09-12):** trước khi backup
được chấp nhận, graph validator đọc cả tool-result và proposal sidecar bằng
stable reader, không chỉ tin `WalkDir` trước đó. Regression đổi canonical
tool-result sau open thành symlink cùng bytes; validation reject. Mutation CI
bỏ riêng tool-result guard và đòi assertion thật fail. Đây không chứng minh mọi
sidecar mutation, crash/power-loss, multi-host hay backup operated.

**Cập nhật M08 M07-proposal stable reader (2026-09-12):** M08 chỉ bind agent
intent với proposal immutable của M07 sau khi resolver giữ stable regular-file
identity. Regression gọi learner CLI thật, đổi proposal sau open thành symlink
ngoài cùng bytes và yêu cầu M08 reject trước khi ghi intent. Mutation CI bỏ
riêng reader đó và đòi assertion thật fail. Đây chỉ là guard identity cục bộ;
portable input khác, crash/power-loss, multi-host và authority/executor vẫn
ngoài phạm vi.

**Cập nhật audit CI mutation wiring (2026-09-13):** readiness audit coi các
lệnh mutation M10/M11, backup source identity, recovery-journal, M07 tool-result,
M07 CLI portable input, general CLI portable input, internal learner portable
input, backup-sidecar và M08
M07-proposal path identity
là required regression của `curriculum-ci.yml`, đồng thời đòi script hiện diện;
negative fixture xoá lệnh M10, backup hoặc recovery-journal khỏi workflow phải
làm audit fail. Audit này chỉ xác minh text
workflow đang khai báo lệnh, không thay bằng chứng GitHub Actions ở head hay mở
rộng mutation coverage ra ngoài các guard đã nêu.

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

**Re-run n8n engine cục bộ (2026-09-14):** workflow fixture M06/M07 được chạy
lại qua n8n `2.38.1` trên macOS arm64, Node `24.21.0`, SQLite disposable và
loopback adapters/model stub. Node 22 bị n8n hiện hành từ chối vì package yêu
cầu `>=24.0.0`; native `isolated-vm` phải được build/rebuild bằng đúng Node 24
trước import. CI đã pin Node 24. Đây là reproduction có thể tái lập của engine
fixture path, không thay release admission `UNVERIFIED`, không chứng minh
provider, selected-source operated run, topology deploy hay business outcome.
Runner Schedule Trigger thực cũng PASS trong cùng runtime: append một lần,
`EXACT_DUPLICATE` trước/sau restart n8n/adapter và adapter failure chặn
ACK/report. Không suy fixture schedule này thành vận hành deployment.

**Mở rộng M06 engine CI (2026-09-09):** runner còn chạy key-order retry,
same-correlation content conflict, unsupported source và changed event. Hai ca
reject phải dừng ở adapter trước ACK/report và không được thay đổi bytes history;
changed event mới phải persist record riêng và replay `MATCH`. Phạm vi vẫn là
fixture synthetic; selected-source profile không được suy ra là đã nghiệm thu.

**M06 selected-source operated run (2026-09-15):** phiên vận hành đã quan sát
campaign Shopee Smartlink trên ACCESSTRADE ở chế độ read-only, lọc thành
sanitized capture theo profile cố định và tính SHA-256 ảnh trang giữ riêng.
Learner Bot thật import capture vào history mới với `APPENDED`,
`GET_MORE_DATA`, `execution_permitted=false`, không fetch/network/link và
`history replay=MATCH`. Cùng capture được đưa qua bản copy inactive của
blueprint M06 trong n8n local `2.38.1`, gọi adapter loopback thật; execution
`success`, `APPENDED`, canonical history ACK/persisted đều `true`, sau đó replay
vẫn `MATCH`. Bản ghi chi tiết nằm tại
`docs/architecture/EVIDENCE-M06-ACCESSTRADE-SHOPEE-OPERATED-20260915.md`.
Runtime vẫn gắn `evidence_kind=synthetic`, thiếu `price`/`commission_rate`, và
không suy ra business outcome. Đây là local operated evidence, chưa phải
independent review/deployment evidence; approval campaign, provider diversity,
Smartlink, live executor, payout, pilot máy sạch và recovery host thật vẫn mở,
nên BR-13/14 và overall readiness giữ `PARTIAL`/`NOT_READY_FOR_PRODUCTION`.

**BR-18b local recovery drill (2026-09-15):** chạy
`python3 scripts/smoke_br18b_backup_restore.py` trên `main`
`7be0ccc59aa05d4068ade3193ace0d3c26295aa5`. Learner Bot thật đã build graph
M10/M11 trong workspace tạm, tạo backup/restore bằng manifest v3, đọc lại bằng
process mới, kiểm replay, budget/lease-window, các link evaluation/cycle và
durable STOP; kết quả `BR-18b PASS`. Bằng chứng chi tiết nằm tại
`docs/architecture/EVIDENCE-BR18B-LOCAL-RECOVERY-DRILL-20260915.md`.
Đây chỉ là local fixture/runtime evidence: không đóng deployment recovery trên
target host, power-loss/multi-host/24/7 hoặc business/provider evidence, nên
BR-18b vẫn `PARTIAL`, RP-10 vẫn `OPEN` và readiness giữ
`NOT_READY_FOR_PRODUCTION`.

**BR-16b assisted fresh-workspace verification (2026-09-15):** chạy
`smoke_quickstart.py` trên clone cô lập với cache rỗng và
`smoke_br16a_offline.py` trên workspace mới
`/tmp/br16b-assisted-pilot.zcjhwp`. Quickstart đã thực hiện intentional
FAIL/fix; chuỗi M00–M11 trả `BR-16a PASS`, replay `MATCH`, EC-01…EC-05 PASS,
`stop=true`, `stop_reason=RECONCILIATION_REQUIRED` và
`recovery_admission_execution_permitted=false`. Bản ghi chi tiết nằm tại
`docs/architecture/EVIDENCE-BR16B-ASSISTED-FRESH-WORKSPACE-20260915.md`.
Đây là automation/maintainer-assisted verification, không phải clean-machine
self-service pilot PASS; pilot độc lập, target deployment và external outcome
vẫn mở, nên BR-16a/RP-10 và overall readiness không được nâng trạng thái.

**M06 selected-source engine CI (2026-09-13):** cùng runner disposable import
blueprint M06 ACCESSTRADE Shopee Smartlink ở dạng inactive/manual, thay trigger
chỉ trong bản copy test và truyền metadata fixture đã làm sạch. Nó phải append
một observation, trả `EXACT_DUPLICATE` cho retry, chặn field `commission_rate`
ngoài allowlist tại adapter trước ACK/report mà không đổi history, append khi
page fingerprint đổi, rồi replay mọi record `MATCH`. Không có request tới
ACCESSTRADE, capture thật, credential, Smartlink, price/commission fact hay
business outcome; đây chỉ là engine evidence cho contract offline.

**M06/M07 disposable n8n re-verification (2026-09-14, `main` `d0d02c0`):**
với n8n `2.38.1` và Node `24.21.0` local, hai runner disposable
`run_n8n_engine_regression.py` và `run_n8n_m06_schedule_regression.py` đều
PASS lại qua adapter loopback riêng. M06 Schedule Trigger append đúng một
record, exact-deduplicate trước/sau restart n8n+adapter, và dừng trước
ACK/report khi adapter unavailable; engine M06/M07 chỉ persist output grounded
fixture canonical và reject tool/model invalid. Reverify này không dùng workflow
hoặc credential local đã lưu, không gọi ACCESSTRADE/Cockpit/provider, không là
selected-source operated capture, topology deployment hay business evidence.

**M06 Schedule Trigger CI (2026-09-12):** runner riêng khởi động `n8n start`
trên SQLite disposable và active một copy M06 có Schedule Trigger thật cadence
một giây. Nó đòi execution `mode=trigger` `APPENDED`, rồi `EXACT_DUPLICATE`
cùng record trước và sau restart n8n/adapter, report ACK/read-only, record history
duy nhất và replay `MATCH`; adapter loopback không khả dụng phải tạo lỗi trước
ACK/report không ghi history. Điều này đóng schedule admission/idempotency
offline, không là deployment hay selected-source operated evidence.

**Mở rộng M07 model-stub CI (2026-09-09):** cùng job import credential disposable
chỉ trỏ endpoint OpenAI-compatible loopback, thay network fetch ở bản copy test
bằng fixture `register-tool-result`, rồi giữ Agent/context/grounding/proposal
nodes của blueprint. Stub phải nhận canonical context và tool evidence; output
được validate, persist, rồi revalidate sau restart adapter. Điều này không đưa
secret/provider vào CI và không là bằng chứng received-redirect hoặc provider
diversity.

**M07 selected-source engine grounding CI (2026-09-13):** sau khi metadata
Shopee Smartlink fixture đã làm sạch được M06 n8n append, cùng runner truyền
record đó vào Agent M07 thật với model OpenAI-compatible loopback. Case hợp lệ
chỉ restate `price:null` cùng evidence ID canonical và phải persist draft
read-only. Case model bịa `commission_rate:0.9` dưới evidence ID commission
canonical phải dừng tại grounding adapter trước proposal/report, không được
tạo artifact mới. Không gọi ACCESSTRADE, không dùng provider credential thật,
không biến metadata campaign thành fact commission hay business outcome.

**M07 malformed-output engine CI (2026-09-13):** model stub loopback trả prose
không phải JSON qua Agent n8n thật. Workflow phải dừng tại grounding adapter
trước proposal/report và không được tạo artifact; đây là decode fail-closed
offline, không phải provider evidence.

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

**Cập nhật review-finding mapping (2026-09-11):** matrix giờ có mapping có cấu
trúc R01–R16 → RP chủ trì → BR liên quan cùng scope/status. Audit yêu cầu đủ
16 ID duy nhất, package/BR hợp lệ và marker R tương ứng trong plan; negative
fixture bỏ một mapping phải fail. Đây làm R16 kiểm được liên kết review, không
xác nhận remote CI, operated evidence hay đóng các finding `PARTIAL`.

**Cập nhật public readiness boundary (2026-09-12):** audit còn đọc hai entrypoint
`README.md` và `curriculum/README.md` khi matrix giữ `NOT_READY_FOR_PRODUCTION`.
Cả hai phải giữ marker này; curriculum không được tự gọi toàn bộ đường học là
`learner-operable`. Negative fixtures xóa marker hoặc đưa lại claim đó phải fail
trên chính audit. Điều này chỉ chặn overclaim public-facing, không chứng minh
pilot, selected source, provider hay deployment evidence.

**Cập nhật M07 strict output decoder (2026-09-12):** `ValidateAgentOutput`
giờ strict-decode toàn bộ JSON model trước khi kiểm required fields và semantic
grounding. Vì vậy duplicate key ở mọi depth, key viết sai hoa/thường, field
không thuộc contract và JSON có trailing value bị reject trước khi output có
thể trở thành proposal. Regression core đưa payload thực vào boundary, còn
mutation proof thay strict decoder bằng `json.Unmarshal` và chỉ PASS khi case
duplicate/unknown thực sự làm test fail. CI yêu cầu proof này; đây chỉ là
contract offline của output model, không phải bằng chứng selected source,
provider, execution hay business outcome.

**Cập nhật M07 strict registry decoder (2026-09-12):** learner Bot giờ
strict-decode registry policy trước `ValidateRegistry`; duplicate key,
case-variant/unknown field và trailing JSON đều bị chặn trước khi host/method/
redirect policy được dùng cho tool result hoặc proposal. Regression gọi đúng
registry loader của CLI và mutation proof thay decoder bằng `json.Unmarshal`;
case policy độc hại phải làm regression fail. Đây là hardening cho policy
offline, không phải bằng chứng về selected source, provider, execution hay
business outcome.

**Cập nhật M10/M11 stable registry append (2026-09-12):** append-only
registries không chỉ đọc qua stable regular-file guard: đường append giữ cùng
regular-file identity từ `Lstat` qua open và kiểm lại sau sync, trước ACK. Hai
regression gọi trực tiếp M10 và M11 registry, thay registry đã kiểm bằng
symlink cùng byte trỏ ra file ngoài ngay trước append, rồi xác nhận append bị
reject và file ngoài không đổi. Đây chỉ là guard filesystem cục bộ; không là
transaction đa-file, recovery power-loss, multi-host lock hay proof executor.

**Cập nhật canonical history stable append (2026-09-13):** writer JSONL của
learner không còn mở append trực tiếp vào một pathname. Với target có sẵn, nó
giữ identity của regular file từ `Lstat` qua open/fstat, sync/close và kiểm lại
trước ACK; với target chưa tồn tại, nó dùng `O_EXCL` để không theo tên xuất hiện
sau check. Regression thay history đã kiểm bằng symlink cùng byte trỏ tới file
ngoài, và tạo symlink sau absent check; cả hai đều reject và file ngoài giữ
nguyên. Đây chưa bảo vệ toàn bộ read path, không chứng minh crash/power-loss hay
multi-host transaction.

**Cập nhật canonical history stable reader (2026-09-13):** reader JSONL dùng
cùng boundary đầu vào: `Lstat` regular file rồi open/fstat cùng inode trước khi
trả handle cho parser. Regression thay history đã kiểm bằng symlink cùng byte
trỏ tới file ngoài ngay trước open và reader reject.

**Cập nhật canonical history close verification (2026-09-13):** reader giữ
inode đã open và kiểm lại pathname trên `Close`; mọi loader canonical history,
action, outcome, evaluation, improvement và ACCESSTRADE receipt đều propagate
`Close` error sau parse thay vì trả artifact đã đọc. Regression đổi pathname
sau open thành symlink cùng byte: bytes của descriptor vẫn đọc được nhưng
`Close` fail và loader trả lỗi. Đây chưa phát hiện mutation in-place của inode
đang mở; phạm vi đó được kiểm riêng ở cập nhật kế tiếp. Crash/power-loss hoặc
multi-host transaction vẫn nằm ngoài boundary này.

**Cập nhật canonical history content verification (2026-09-13):** khi parser
đã đọc tới EOF, reader giữ SHA-256 của chính descriptor đã mở và đọc lại chính
descriptor đó trên `Close`. Nếu nội dung bị rewrite in-place (kể cả cùng kích
thước và cùng inode) sau khi parser nhận bytes, `Close` fail nên loader không
ACK artifact đã parse. Regression gọi JSONL reader thật, đọc snapshot ban đầu,
rewrite pathname qua cùng inode và xác nhận `Close` reject. Đây là kiểm nhất
quán cục bộ của một snapshot đã consume, không phải snapshot atomic trước writer
không hợp tác có thể race/restore bytes; crash/power-loss và multi-host
transaction vẫn mở.

**Cập nhật canonical runtime-store content verification (2026-09-13):** shared
`readStableRegularFile` của mission state/STOP, M10/M11 authority stores,
backup/restore, M07 sidecar và recovery journal nay reread chính descriptor sau
lần đọc hoàn chỉnh đầu tiên rồi so SHA-256 trước ACK. Regression gọi đúng
primitive, rewrite nội dung cùng kích thước trên cùng inode giữa hai lần đọc và
nhận reject. Điều này không là snapshot atomic trước writer không hợp tác có
thể race/restore bytes, cũng không chứng minh crash/power-loss, transaction đa
file hay multi-host recovery.

**Cập nhật backup-manifest stable reader (2026-09-13):** manifest là artifact
chọn inventory và digest cho restore, nên `verifyBackup` đọc nó bằng shared
stable reader trước khi duyệt layout. Regression tạo backup hợp lệ, thay
`manifest.json` bằng symlink cùng byte ra ngoài, rồi xác nhận restore trả
`VERIFY_FAILED`, không publish target và không đổi file ngoài. Đây không thay
thế proof snapshot atomic, crash/power-loss hoặc recovery multi-host.

- Matrix mở rộng tiêu chí theo từng gap và phân loại `implementation_gaps`, `test_gaps`, `external_evidence_gaps`; implementation/test/evidence refs có scope/version/commit và trạng thái rõ. Migrate version của schema/audit cùng lúc.
- Tự sinh hoặc kiểm bảng BR từ matrix. Audit phát hiện thiếu R01–R16 mapping, ref hỏng, thiếu evidence của claim đã đóng, status mâu thuẫn và prose đang tuyên bố cao hơn trạng thái được chấp nhận.
- Có negative fixtures của chính audit: empty refs, stale/incorrect commit, implemented nhưng thiếu regression, plan cao hơn matrix, toàn offline nhưng claim production. Không cố định một câu output rồi gọi là readiness computation.
- Reviewer đối chiếu các PR, rerun suite/cases theo baseline/head, ghi remaining gaps cụ thể. Current planning PR chỉ sửa thông tin sai; R16 chưa đóng cho tới khi audit mới bắt được chúng.

**Nghiệm thu:** R16 đóng ở mức cơ chế kiểm chứng; giữ overall NOT_READY_FOR_PRODUCTION cho tới khi có evidence ngoài repo và authority cần thiết. Không biến VERIFIED_OFFLINE thành learner/pilot PASS.

**BR-16a/BR-18b local runtime re-run (2026-09-16):** trên `main` tại
`2511efd`, chạy lại `smoke_br16a_offline.py` và
`smoke_br18b_backup_restore.py` bằng learner Bot thật. Cả hai PASS: chuỗi
M00→M11 dùng chung runtime/artifact, backup v3 được tạo và restore vào target
mới, graph được replay, và post-restore budget/lease-window cùng durable STOP
được kiểm. Cùng đợt chạy `audit_readiness.py`, 88 negative audit tests,
`go test -count=1 ./...` và `go vet ./...` cho cả bốn module đều PASS. Đây là
evidence local fixture/read-only; không đóng BR-16a/BR-18b ở mức provider,
business outcome, clean-machine pilot, deployment, multi-host hoặc
crash/power-loss. Overall vẫn `NOT_READY_FOR_PRODUCTION`. Record:
`docs/architecture/EVIDENCE-LOCAL-RUNTIME-RERUN-20260916-2.md`.

**Full local regression re-run (2026-09-16):** trên `main` tại
`6103eb8`, chạy lại bốn Go module (test/vet), 104 Python regression tests,
các validator M06/M07, BR-16a, BR-18b và `git diff --check`; tất cả đều PASS.
Disposable n8n `2.38.1` với Node `24.21.0` cũng PASS cho M06/M07 engine và
Schedule Trigger: append đúng một lần, retry exact-deduplicate trước/sau
restart, và adapter unavailable bị fail-closed. Đây là record local/offline
được tạo từ implementation thật; không đóng các gap provider, live executor,
business outcome, clean-machine pilot, target-host deployment, multi-host hay
crash/power-loss. Record:
`docs/architecture/EVIDENCE-LOCAL-FULL-REGRESSION-20260916.md`.

**Full local regression re-run on current main (2026-09-16):** tại `main`
`c56cf95`, chạy lại bốn Go module (test/vet), 104 Python regression tests,
validator M06/M07, BR-16a, BR-18b, disposable n8n engine và M06 Schedule
Trigger; tất cả PASS. Schedule Trigger append đúng một lần, exact-deduplicate
trước/sau restart n8n + adapter và fail-closed khi adapter unavailable. Audit
vẫn trả `NOT_READY_FOR_PRODUCTION`. Đây chỉ là bằng chứng local fixture/read-only;
không đóng provider, live executor, business outcome, clean-machine pilot,
target-host deployment, multi-host hoặc crash/power-loss. Record:
`docs/architecture/EVIDENCE-LOCAL-FULL-REGRESSION-20260916-C56CF95.md`.

**Cập nhật local backup/restore process-exit lock (2026-09-16):** regression
chạy learner Bot thật trong process con, buộc process thoát khi output còn ở
staging. POSIX advisory lock tự nhả khi process kết thúc; backup/restore retry
được và target dở dang không xuất hiện. Đây chỉ là một seam local trước
publish, không chứng minh power-loss, atomic multi-file, Windows native lock,
multi-host hay dọn staging mồ côi; readiness vẫn `NOT_READY_FOR_PRODUCTION`.

**Cập nhật managed lock path guard (2026-09-16):** POSIX lock publisher dùng
`O_NOFOLLOW`; regression thay lock pathname bằng symlink tới file ngoài và xác
nhận acquisition bị từ chối, bytes external vẫn nguyên vẹn. Đây là guard
pathname local cho runtime/backup/restore lock, không thay thế ancestor-path
TOCTOU, Windows native lock, multi-host hoặc power-loss proof.

**Deterministic runtime re-run trên current main (2026-09-16):** chạy lại đúng
single-build sequence của `curriculum-ci.yml`: learner internal tests, canonical
schema test/vet, build learner Bot một lần, M01 baseline, M02 capture/list/
replay/decision và readiness audit. Tất cả PASS trên `main` `2e86636`; audit
vẫn trả `NOT_READY_FOR_PRODUCTION`. Run local warm-cache khoảng 0.8 giây không
phải benchmark GitHub runner. Record:
`docs/architecture/EVIDENCE-DETERMINISTIC-RUNTIME-RERUN-20260916-2E86636.md`.
Đây chỉ là fixture/offline/read-only verification; không có provider, business
outcome, live executor, clean-machine pilot, target-host deployment,
multi-host hoặc power-loss evidence.

**Cập nhật backup/restore SIGKILL regression (2026-09-16):** backup/restore
process-boundary test giờ có cả nhánh `os.Exit` và nhánh kill process thật;
nhánh sau kiểm child kết thúc do `SIGKILL`, target staging chưa từng xuất hiện,
và retry bằng learner Bot thật tạo/restore được snapshot. Đây vẫn chỉ là
POSIX local advisory-lock seam; process kill không phải power-loss proof và
không chứng minh atomic multi-file, Windows native lock, multi-host hay dọn
staging mồ côi. Record:
`docs/architecture/EVIDENCE-BACKUP-SIGKILL-20260916.md`.

**Cập nhật managed lock ancestor preflight (2026-09-16):** POSIX managed lock
kiểm tra các ancestor hiện hữu của lock pathname trước khi mở lock, nên một
lock nằm dưới thư mục cha là symlink bị từ chối trước khi chạm file ngoài.
Regression dùng lock path thật dưới symlink parent, xác nhận external sentinel
không đổi và external lock không được tạo. Đây là bounded local preflight;
thay thế ancestor đồng thời vẫn cần directory-handle/openat hardening, và
Windows native lock, power-loss, multi-host vẫn ngoài phạm vi. Record:
`docs/architecture/EVIDENCE-MANAGED-LOCK-ANCESTOR-20260916.md`.

**Cập nhật managed lock directory-handle/openat hardening (2026-09-16):**
POSIX lock acquisition giờ mở từng directory component bằng descriptor với
`O_DIRECTORY|O_NOFOLLOW`, rồi mở entry cuối bằng `openat` trên parent descriptor.
Regression thay parent thật bằng symlink ngay sau preflight; acquisition bị
từ chối, external sentinel không đổi và external lock không được tạo. Đây là
hardening cho local ancestor replacement seam trên các POSIX target hỗ trợ
`openat`, không phải distributed lock, Windows native lock, power-loss,
multi-file atomicity hay multi-host proof. Record:
`docs/architecture/EVIDENCE-MANAGED-LOCK-OPENAT-20260916.md`.

**Cập nhật M11 process-kill journal recovery (2026-09-16):** learner Bot dùng
managed lock cho `.mission.lock` trên POSIX, nên SIGKILL sau khi replay journal
đã publish không để lại cooperative marker chặn writer kế tiếp; backup inventory
bỏ qua control file regular này. Regression chạy child Bot thật cho cả FAILED và
UNKNOWN→STOP: fresh read-only process trả `RECOVERY_REQUIRED`, writer lock sau đó
replay đúng journal và giữ nguyên FAILED ledger hoặc durable STOP. Đây chỉ là
bounded local process-termination seam, không phải power-loss, transaction đa
file, Windows native lock hay multi-host proof. Record:
`docs/architecture/EVIDENCE-M11-PROCESS-KILL-JOURNAL-20260916.md`.

**Cập nhật M11 outcome process-kill recovery (2026-09-16):** bổ sung các
regression riêng trên đường `m11-outcome`: child Bot thật nhận `SIGKILL` ở
hai điểm cắt, sau khi journal đã publish nhưng lần lượt trước ledger append và
sau outcome append trước cleanup. Process đọc mới trả `RECOVERY_REQUIRED`,
locked writer replay/confirm đúng transition, retry trả `EXACT_DUPLICATE` và
store chỉ còn một outcome. Đây là bounded local synthetic/read-only
process-termination evidence, không phải power-loss, atomic multi-file,
provider, business outcome, pilot hay deployment proof.
Record: `docs/architecture/EVIDENCE-M11-OUTCOME-PROCESS-KILL-20260916.md`.

**Cập nhật M11 process-kill sau ledger append (2026-09-16):** regression mới
chạy child Bot thật cho cả `FAILED` và `UNKNOWN→STOP`, gửi `SIGKILL` sau khi
ledger transition đã append và directory sync đã hoàn tất nhưng trước cleanup
journal. Process đọc mới vẫn trả `RECOVERY_REQUIRED`; writer có lock nhận diện
đúng execution/ledger đã hiện hữu, hoàn tất cleanup đúng một lần, giữ nguyên
`FAILED` ledger hoặc durable STOP, và không nhân đôi artifact. Đây là bounded
local POSIX process-termination evidence, không phải power-loss, atomic
multi-file, Windows native-lock, distributed/multi-host, provider, business
outcome, pilot hoặc deployment proof.
Record: `docs/architecture/EVIDENCE-M11-PROCESS-KILL-AFTER-LEDGER-20260916.md`.

**Cập nhật backup/restore orphan staging recovery (2026-09-16):** backup và
restore giờ đặt hash của target vào prefix staging. Sau khi giữ managed target
lock, lần retry sau process termination chỉ dọn các thư mục staging private
đúng target đó; symlink, file, staging của target khác và target visible không
bị theo. Regression chạy child Bot thật cho cả backup và restore với `SIGKILL`,
xác nhận target chưa publish, staging target-owned còn lại sau kill và biến mất
sau exact retry. Đây là bounded local POSIX cleanup seam, không phải proof
power-loss, transaction đa-file, filesystem crash, Windows native lock,
multi-host, provider, business outcome, pilot hay deployment. Record:
`docs/architecture/EVIDENCE-BACKUP-ORPHAN-STAGING-20260916.md`.

**Cập nhật backup/restore post-publish process-kill boundary (2026-09-16):**
regression chạy child Bot thật và kill ngay sau `os.Rename`, trước
parent-directory acknowledgement. Target backup visible phải qua
`verifyBackup`, runtime restore visible phải qua `loadMissionState`, và exact
retry cùng target phải trả `TARGET_NOT_EMPTY` để không ghi đè snapshot đã
publish. Đây là bounded local rename/process-kill safety seam, không chứng minh
parent sync qua power-loss, transaction đa-file, filesystem crash, Windows
native lock, multi-host, provider, business outcome, pilot hay deployment.
Record: `docs/architecture/EVIDENCE-BACKUP-POST-PUBLISH-PROCESS-KILL-20260916.md`.

**Cập nhật M10 journal process-kill recovery (2026-09-16):** learner Bot
test-binary chạy riêng hai transition canary và trusted cost-bound, kill thật
child process bằng `SIGKILL` ngay sau khi recovery journal đã được sync và
publish, trước canonical side effect đầu tiên. Process mới đọc `status` phải
trả `RECOVERY_REQUIRED`; writer có lock sau đó replay đúng một lần, xóa
journal và trả ACK/EXACT_DUPLICATE, đồng thời cost-bound resolve được từ M10
registry. Đây là bounded local POSIX process-termination evidence, chưa phải
power-loss, transaction đa-file toàn runtime, multi-host, provider,
business-outcome, pilot hoặc deployment proof. Record:
`docs/architecture/EVIDENCE-M10-JOURNAL-PROCESS-KILL-20260916.md`.

**Cập nhật M10 execution journal process-kill recovery (2026-09-16):** learner
Bot test-binary chuẩn bị governed authorization + reservation rồi kill thật
child `m10-record-failed` bằng `SIGKILL` ngay sau khi execution recovery
journal được sync và publish, trước canonical execution registry/state binding.
Process mới đọc `status` phải trả `RECOVERY_REQUIRED`, không thấy portable
output hay execution registry entry; writer có lock sau đó replay đúng một lần,
xóa journal, publish output và bind reservation → execution. Đây là bounded
local POSIX process-termination evidence, chưa phải power-loss, transaction
đa-file toàn runtime, multi-host, provider, business-outcome, pilot hoặc
deployment proof. Record:
`docs/architecture/EVIDENCE-M10-EXECUTION-PROCESS-KILL-20260916.md`.

**Cập nhật M10 process-kill sau canonical append (2026-09-16):** regression
chạy child Bot thật và gửi `SIGKILL` sau các cạnh canonical append tiếp theo
của hai transition M10: canary sau immutable artifact và sau state binding,
trusted cost-bound sau immutable artifact và sau compact index. Process đọc
mới vẫn trả `RECOVERY_REQUIRED`; writer có lock nhận diện phần đã hiện hữu,
hoàn tất phần còn lại, xóa journal và giữ đúng một artifact/index entry.
Đây là bounded local POSIX process-termination evidence, không phải
power-loss, atomic multi-file, Windows native-lock, distributed/multi-host,
provider, business outcome, pilot hoặc deployment proof.
Record: `docs/architecture/EVIDENCE-M10-CANONICAL-APPEND-PROCESS-KILL-20260916.md`.

**Cập nhật M10 execution process-kill sau canonical append (2026-09-16):**
regression `TestMissionM10ExecutionProcessKillAfterCanonicalAppendRequiresLockedReplay`
chạy child Bot thật cho hai cạnh của governed execution journal: sau khi
execution record đã append vào registry và sau khi reservation đã bind
execution ID vào mission state. Cả hai child đều bị `SIGKILL` trước cleanup;
process đọc mới trả `RECOVERY_REQUIRED`, không tạo portable output, còn locked
retry nhận diện phần đã hiện hữu, hoàn tất phần còn lại và để đúng một
execution record cùng một reservation-to-execution binding. Đây là bounded
local POSIX process-termination evidence, không phải proof power-loss,
atomic multi-file, Windows native-lock, distributed/multi-host, provider,
business outcome, pilot hoặc deployment. Record:
`docs/architecture/EVIDENCE-M10-EXECUTION-CANONICAL-APPEND-PROCESS-KILL-20260916.md`.

**Cập nhật backup/restore M11 graph guards (2026-09-16):**
`smoke_br18b_backup_restore.py` tạo backup bằng learner Bot thật, nhân đôi
nguyên `PRODUCTION_EXECUTION_RECORD` trong `m11-artifacts.jsonl` rồi cập nhật
checksum/size manifest để mutation vượt qua lớp integrity; restore trả
`VERIFY_FAILED` và không publish target. Cùng run tạo các bản sao checksum-valid
thiếu từng artifact M11 bắt buộc từ lease/approval/health/cost/gate tới
authorization/execution/evaluation; learner Bot trả `VERIFY_FAILED` hoặc
`GRAPH_FAILED` và không publish target. `PRODUCTION_CYCLE` hiện là hậu kiểm tùy
chọn trong snapshot lịch sử, nên không được gắn required trong test này. Đây là
kiểm cardinality/identity/required-link của canonical M11 registry trên đường
backup/restore. Phạm vi vẫn là local
fixture/read-only, không phải atomic multi-file crash/power-loss,
multi-host/provider/business-outcome/pilot/deployment proof. Record:
`docs/architecture/EVIDENCE-BACKUP-M11-DUPLICATE-ARTIFACT-20260916.md`.

**Cập nhật local full regression sau PR #360 (2026-09-16):** trên `main`
`930a040c80d4295c7dac60980b96030d9b6439eb`, maintainer đã chạy lại bốn Go
module test/vet, 113 Python regression tests, toàn bộ static semantic
validators, `smoke_br16a_offline.py`, `smoke_br18b_backup_restore.py` và
`audit_readiness.py`. Các lệnh đã chạy đều PASS; audit vẫn trả
`NOT_READY_FOR_PRODUCTION`. Hai validator `*_operated_execution.py` không chạy
độc lập vì cần execution JSON/store artifact; record đầy đủ nằm tại
`docs/architecture/EVIDENCE-LOCAL-REGRESSION-POST-360-20260916.md`. Đây là
fixture/read-only local evidence, không phải provider, business outcome,
clean-machine pilot, target deployment, distributed-lock hoặc power-loss proof.

### RP-10 — Chỉ thực hiện sau khi đóng các code/test gaps

- Chọn phiên bản n8n hỗ trợ, topology kết nối canonical adapter, host/nguồn read-only, provider/model và budget với chủ repo. Không truy cập loopback của host khác qua cấu hình mặc định; có hướng dẫn container/host đúng topology.
- Import và chạy blueprint thật; lưu export workflow/version, execution ID, tool/store ACK, failure cases và provenance. Provider credentials/secret không commit.
- Pilot máy sạch: học viên làm theo walkthrough, ghi trợ giúp, lỗi, kết quả và khả năng giải thích boundaries; có trợ giúp thiết kế/code thì không gọi self-service PASS.
- Deployment drill trên target được cấp quyền: build/start/health/status/logs/stop, resource limits, backup/restore, restart/expiry/STOP, human-reviewed recovery. Không suy power-loss/24×7 proof từ smoke local vài giây.
- Ghi riêng operated evidence và business evidence; chưa chọn affiliate platform/channel hoặc live executor thì để phần đó OPEN, không chặn sửa offline đã được xác định.

**Post-merge PR #412 acceptance evidence (2026-09-17):** PR #412 đã được merge
vào `main` tại `e197f854a9f396bbe1136004b792b2699dc647e6`, từ head
`f3e5a213e5b142d23961443c3af9097ae3e5456e`. Hai workflow hậu-merge
`35241923424` (Curriculum CI) và `35241923430` (Mission Agent Path CI) đều
hoàn tất thành công với 12/12 check-runs PASS, bao gồm learner race, mutation,
backup/restore và Node 24/n8n. Local Windows chạy lại 132 Python audit tests,
readiness audit, parse JSON và `git diff --check` đều PASS; không có `go.exe`
hoặc n8n CLI nên không claim Go/n8n local. Đây chỉ là regression evidence sau
merge, không phải provider, live executor, business outcome, clean-machine
pilot, target deployment, distributed locking, power-loss hay atomic multi-file
proof; readiness vẫn `NOT_READY_FOR_PRODUCTION`. Marker: `Post-merge PR #412 acceptance evidence`.

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

**Cập nhật M07 n8n raw JSON transport (2026-09-17):** nhánh regression
model-stub không còn JSON.parse tool result trước khi gửi vào
/v1/m07/register-tool-result; nó truyền tool_result_text nguyên dạng tới
adapter. Blueprint thêm node Require Raw Model JSON Text, chỉ cho phép output
model dạng chuỗi và gửi model_output_text nguyên dạng tới validate/proposal.
Ca fixture thật với 9007199254740993 kiểm tool sidecar, model output,
proposal bytes, restart và revalidation; input raw/object đồng thời bị reject.
Static validator và readiness audit giữ các cạnh này ở trạng thái có thể kiểm
tra. Worktree hiện không có Node/n8n để chạy engine, nên chưa ghi PASS local;
CI Node 24 vẫn là bằng chứng cần thiết. Đây là hardening synthetic/read-only,
không phải provider diversity, selected-source, business outcome, live
executor, pilot hay deployment evidence.

**Cập nhật M11 direct STOP process-kill recovery (2026-09-17):** regression
`TestM11ManualStopProcessKillRequiresLockedRecovery` chạy learner Bot
test-binary thật và gửi POSIX `SIGKILL` sau lần lượt các atomic rename của
`m11-manual-stop-journal.json`, `mission-state.json` và `STOP`. Mỗi ca giữ
journal làm recovery authority; process đọc mới trả `RECOVERY_REQUIRED`, còn
writer có lock replay đúng journal, giữ nguyên STOP state/marker, xóa journal
đúng một lần và không cho lần stop khác đổi reason. Đây là bằng chứng local
synthetic/read-only cho direct STOP replay, không phải power-loss, atomic
multi-file, distributed-lock, live-executor, provider, business-outcome, pilot
hay deployment proof. Marker: `Cập nhật M11 direct STOP process-kill recovery`.

**Cập nhật RP-01 registry append parent openat guard (2026-09-17):** M10 và
M11 append-only registry không còn mở pathname trực tiếp ở lần tạo entry đầu
tiên. Trên POSIX, helper mở toàn bộ parent bằng descriptor với
`O_DIRECTORY|O_NOFOLLOW`, gọi hook regression sau khi parent đã được pin,
kiểm parent name vẫn cùng inode rồi mới `openat` entry cuối với
`O_NOFOLLOW|O_EXCL` khi cần. Vì vậy parent bị đổi sang symlink trong khoảng
preflight → open bị từ chối trước khi registry rơi vào external tree; fallback
Windows/OS khác giữ preflight + `O_EXCL` và không claim parity native. Hai
regression M10/M11 gọi đúng register path và kiểm external sentinel/registry
không bị ghi. Worktree hiện thiếu Go executable nên chưa có local Go test/vet;
remote CI `go test -race ./...` trên head trước merge đã PASS trong 12 PR
checks. Đây là hardening pathname offline có phạm vi hẹp, không phải proof
writer không hợp tác ngoài
seam, power-loss, atomic multi-file, distributed locking, provider, business
outcome, pilot hoặc deployment. Marker: `Cập nhật RP-01 registry append parent openat guard`.
