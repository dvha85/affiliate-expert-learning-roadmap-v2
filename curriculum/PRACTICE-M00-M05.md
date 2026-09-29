# Walkthrough thực hành M00–M05 trên cùng learner Bot

Chuỗi thực hành: M00 packet → input M01 → history M02 → action/outcome M03 → advisor M04 →
evaluation/proposal/review M05. Các lệnh dưới dùng **một Bot, một workspace và
cùng artifact lineage**, không thêm Mission hay PASS gate.

## 1. Cách dùng và ranh giới

| Mode | Mục đích | Giới hạn |
|---|---|---|
| synthetic rehearsal | luyện CLI, schema, failure và continuity bằng mẫu bên dưới | không phải evidence thật; không tạo Mission PASS, Reality PASS hoặc Operated PASS |
| learner evidence | nhập artifact thật do người học thu thập/thực hiện và được review | không thay bằng fixture hoặc tool success |

Thứ tự chính thức vẫn là M00 PASS → M01 PASS → M02 PASS → M03 → M04 → M05.
Rehearsal giúp học capability khi còn thiếu source/account; giữ gate tương ứng
mở. Không cập nhật PROGRESS.md từ fixture/CI hoặc kết quả chạy hộ. Chỉ ghi tiến độ
khi người học thực sự hoàn thành và bằng chứng được review.

M00 vẫn cần tối thiểu 3 public observations E1 thật và Human DecisionPacket
(action: null), không cần Go/provider key để PASS. Làm theo
[bài M00.1](M00/M00.1-affiliate-intelligence-objective.md),
[bài M00.2](M00/M00.2-evidence-uncertainty.md),
[M00.3](M00/M00.3-decision-approval-execution.md) và
[gói bằng chứng M00](../evidence/M00/M00-EVIDENCE-PACKET.md).
Giữ nguồn, observation_id/subject_id, observed_at, claim_kind, origin real/synthetic,
use context test/replay, giá trị/trạng thái, transformation và limitation.
Missing/pending/unknown không phải số 0.

## 2. Chuẩn bị một lần

Chọn **một** shell, bắt đầu tại repo root và giữ nguyên terminal/cwd tới cuối.
Go phải có trong PATH và đáp ứng go.mod; thiếu thì ghi BLOCKED_TOOLCHAIN.
PowerShell dùng Python qua lệnh python; macOS/Linux dùng python3.
Các đoạn shell là hai phương án tương đương; bằng chứng OS thực sự đã chạy nằm
trong [evidence](../docs/architecture/EVIDENCE-CURRICULUM-ALIGNMENT-20260923.md).

macOS/Linux:

<!-- run:sh:setup -->
```sh
set -eu
repo="$PWD"
work="$(mktemp -d)"
bot="$work/bot"
git status --short --branch
git rev-parse HEAD
go version
go -C "$repo/lab/affiliate-bot" build -o "$bot" ./cmd/bot
```

PowerShell (kể cả Windows PowerShell 5.1):

<!-- run:ps:setup -->
```powershell
$ErrorActionPreference = "Stop"
$repo = (Get-Location).Path
$work = Join-Path ([IO.Path]::GetTempPath()) ("affiliate-rehearsal-" + [guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $work | Out-Null
$bot = Join-Path $work "bot.exe"
git status --short --branch
git rev-parse HEAD
go version
if ($LASTEXITCODE -ne 0) { throw "BLOCKED_TOOLCHAIN: go" }
go -C (Join-Path $repo "lab/affiliate-bot") build -o $bot .\cmd\bot
if ($LASTEXITCODE -ne 0) { throw "Bot build failed" }
```

Ghi commit và các sửa đổi chưa commit bên cạnh output. work là thư mục tạm mới,
không nằm trong evidence/history cá nhân. Input, envelope, output và các store
phải là những tệp khác nhau. Giữ workspace để kiểm tra; chỉ xóa đúng thư mục này
sau khi đã lưu các ghi chú cần thiết.

### Input, output, failure và evidence cho từng chặng

| Chặng | Input synthetic | Expected output | Failure thử | Evidence giữ lại |
|---|---|---|---|---|
| M00→M01 | packet-t1/t2 của repo | VALID; t1 GET_MORE_DATA, t2 RANK_SCENARIO | invalid profile, missing khác zero | envelope, input, prediction/output |
| M02 | input-t1/t2, cùng history.jsonl | APPENDED; EXACT_DUPLICATE; 2 replay=MATCH | source-ID conflict, projection tamper | record IDs, count/hash, replay |
| M03 | action.json, outcome.json bên dưới | APPENDED; list VALID | orphan, window sai | IDs, nguồn/window, rejection |
| M04 | advisor-config.json | envelope SUPPORTED; advisor_output.state HUMAN_REVIEW | dữ liệu cũ/tương lai/mồ côi | config, evidence IDs, output |
| M05 | evaluation-config/proposal/review.json | INCONCLUSIVE; auto_apply=false; review REQUEST_CHANGES | ID mồ côi, auto_apply, review time | IDs, list/replay, diff/rollback |

Mẫu JSON bên dưới chỉ là synthetic. Dùng editor lưu đúng tên file vào work,
UTF-8 không BOM; không sửa ID trong store để ép các lệnh chạy được.

## 3. Chuyển packet M00 → input M01

Hai input có sẵn:
[packet-t1.json](../examples/m00-import/packet-t1.json) và
[mẫu packet-t2.json](../examples/m00-import/packet-t2.json).
Profile m00-input/v1 chỉ chuyển price/commission_rate sang input scenario.
Importer không parse Markdown, fetch URL hay ghi history. Các claim khác phải
giữ trong packet/context, không nhét vào field khác. Metadata đầy đủ và cách map
evidence thật nằm trong [bridge BR-09](../examples/m00-import/README.md).

macOS/Linux:

<!-- run:sh:import -->
```sh
for n in 1 2; do
  "$bot" evidence import "$repo/examples/m00-import/packet-t$n.json" > "$work/envelope-t$n.json"
done
python3 - "$work" <<'PY'
import json, sys
from pathlib import Path
work = Path(sys.argv[1])
for n in (1, 2):
    e = json.loads((work / f"envelope-t{n}.json").read_text(encoding="utf-8-sig"))
    assert e["status"] == "VALID" and e["execution_permitted"] is False
    assert isinstance(e["artifact"], list)
    (work / f"input-t{n}.json").write_text(json.dumps(e["artifact"]), encoding="utf-8")
PY
```

PowerShell:

<!-- run:ps:import -->
```powershell
foreach ($n in 1, 2) {
    $raw = & $bot evidence import (Join-Path $repo "examples/m00-import/packet-t$n.json")
    if ($LASTEXITCODE -ne 0) { throw "Import failed at t$n" }
    [IO.File]::WriteAllText((Join-Path $work "envelope-t$n.json"), ($raw -join [Environment]::NewLine), [Text.UTF8Encoding]::new($false))
}
@'
import json, sys
from pathlib import Path
work = Path(sys.argv[1])
for n in (1, 2):
    e = json.loads((work / f"envelope-t{n}.json").read_text(encoding="utf-8-sig"))
    assert e["status"] == "VALID" and e["execution_permitted"] is False
    assert isinstance(e["artifact"], list)
    (work / f"input-t{n}.json").write_text(json.dumps(e["artifact"]), encoding="utf-8")
'@ | python - $work
if ($LASTEXITCODE -ne 0) { throw "Envelope validation failed" }
```

Chỉ dùng artifact sau khi envelope VALID và execution_permitted=false.
Không đưa toàn envelope cho Bot. Lưu raw envelope riêng để đối chiếu; Python
tránh PowerShell tự đổi timestamp/string provenance khi đọc rồi ghi lại JSON.

Dữ liệu thật: sau M00 gate, tự map packet thành profile v1 trong workspace riêng,
kiểm exact IDs và provenance. Chưa có commission thì ghi null/unknown đúng nguồn;
không chép commission của fixture. T1 giữ pending/null; t2 là giả định synthetic 0.08.

## 4. M01 — dự đoán rồi chạy

Giữ cwd ở repo root, dùng binary đã build:

<!-- run:sh:m01 -->
```sh
"$bot" "$work/input-t1.json"
"$bot" "$work/input-t2.json"
```

<!-- run:ps:m01 -->
```powershell
& $bot (Join-Path $work "input-t1.json")
& $bot (Join-Path $work "input-t2.json")
```

Expected t1 GET_MORE_DATA; t2 RANK_SCENARIO, score=8.00 và
formula_version=commission-per-order/v1. Lưu prediction/observed/limitation.
Đây là price × commission_rate, chưa chứng minh sản phẩm tốt nhất hay lợi nhuận.

Failure: tạo **bản sao input** thiếu price/commission hoặc đổi origin; dự đoán
state/reason, chạy lại và giữ file/output lỗi riêng. Thử observed zero riêng
với missing. Nếu cần pointer/null/map/JSON và test-first, dùng
[Go/JSON practice](BOOT/GO-JSON-PRACTICE.md); không phát sinh gate mới.

[bài M01.1](M01/M01.1-deterministic-contract.md) →
[M01.4](M01/M01.4-failure-first-operated-proof.md) và
[mẫu bằng chứng M01](../starter-kits/M01-deterministic-bot/M01-OPERATED-EVIDENCE-TEMPLATE.md)
quy định evidence thật. Không bắt hoàn tất M02 history để PASS M01.

## 5. M02 — capture, restart, replay và DecisionPacket

Đường học chính chỉ chuyển M02 sau M01 PASS. Rehearsal tiếp tục hai input trên.
Argument capture là HISTORY INPUT RECORD_ID **as_of ingested_at**;
observed_at nằm trong từng observation, không phải argument thay as_of.

<!-- run:sh:m02 -->
```sh
"$bot" history capture "$work/history.jsonl" "$work/input-t1.json" d11 2026-09-01T01:00:00Z 2026-09-01T02:00:00Z
"$bot" history capture "$work/history.jsonl" "$work/input-t2.json" d12 2026-09-02T01:00:00Z 2026-09-02T02:00:00Z
"$bot" history capture "$work/history.jsonl" "$work/input-t2.json" d12 2026-09-02T01:00:00Z 2026-09-02T02:00:00Z
"$bot" history list "$work/history.jsonl"
"$bot" history replay "$work/history.jsonl"
"$bot" history decision "$work/history.jsonl" d12 "$repo/lab/affiliate-bot/data/m02-decision-context.json"
```

<!-- run:ps:m02 -->
```powershell
& $bot history capture "$work/history.jsonl" "$work/input-t1.json" d11 2026-09-01T01:00:00Z 2026-09-01T02:00:00Z
& $bot history capture "$work/history.jsonl" "$work/input-t2.json" d12 2026-09-02T01:00:00Z 2026-09-02T02:00:00Z
& $bot history capture "$work/history.jsonl" "$work/input-t2.json" d12 2026-09-02T01:00:00Z 2026-09-02T02:00:00Z
& $bot history list "$work/history.jsonl"
& $bot history replay "$work/history.jsonl"
& $bot history decision "$work/history.jsonl" d12 "$repo/lab/affiliate-bot/data/m02-decision-context.json"
```

Expected: APPENDED hai lần, EXACT_DUPLICATE lần ba; history vẫn hai record,
replay=MATCH hai lần. Mỗi invocation là process mới nên list/replay không dựa
vào bộ nhớ lần chạy trước. DecisionPacket d12 resolve br09-product-a-t2 và
các source field IDs qua provenance; action=null. Context file trên là synthetic,
không dùng nguyên cho packet thật.

Failure: trên bản sao packet t2, đổi price nhưng giữ source field ID, import rồi
capture record mới phải CONFLICT; sửa projection sau import phải bị từ chối,
history bytes/hash không đổi. Tạo observation mới cần ID/thời gian riêng.
Không xóa/repair history lỗi để tiếp tục. Regression đầy đủ từ repo root:
python scripts/smoke_br09.py (macOS dùng python3).

Xem [M02.4](M02/M02.4-restart-query-operated-proof.md) và
[mẫu bằng chứng M02](../starter-kits/M02-history-replay/M02-OPERATED-EVIDENCE-TEMPLATE.md).
Replay MATCH kiểm deterministic integrity, không phải business truth; giữ
DRIFT, INTEGRITY_ERROR, UNREPLAYABLE riêng biệt.

## 6. M03 — nối action và outcome trên history đó

Chỉ chuyển Mission chính thức sau M02 PASS. Lưu hai mẫu bên dưới trong work.
Decision d12 đã được tạo ở bước trước; các store mới vẫn do learner Bot sở hữu.

<!-- rehearsal:action.json -->
```json
{"action_id":"a12","decision_id":"d12","action_type":"synthetic","target":"fixture:none","performed_by":"human","performed_at":"2026-09-04T00:00:00Z","measurement_window_end":"2026-09-05T00:00:00Z","compliance_reviewed":true}
```

<!-- rehearsal:outcome.json -->
```json
{"outcome_id":"o12","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"a12"},"observed_at":"2026-09-05T00:00:00Z","status":"PENDING","metrics":{},"source_ref":"fixture:br12d"}
```

<!-- run:sh:m03 -->
```sh
"$bot" action record "$work/history.jsonl" "$work/actions.jsonl" "$work/action.json"
"$bot" action list "$work/history.jsonl" "$work/actions.jsonl"
"$bot" outcome import "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/outcome.json"
"$bot" outcome list "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl"
```

<!-- run:ps:m03 -->
```powershell
& $bot action record "$work/history.jsonl" "$work/actions.jsonl" "$work/action.json"
& $bot action list "$work/history.jsonl" "$work/actions.jsonl"
& $bot outcome import "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/outcome.json"
& $bot outcome list "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl"
```

Expected record/import APPENDED, list VALID; chạy lại đúng input trả
EXACT_DUPLICATE. Lưu d12→a12→o12, path/store owner, input/output và upstream
hash trước/sau. Thử bản sao action có decision_id không tồn tại: DECISION_ERROR.
Outcome effect_id mồ côi: ORPHAN_ACTION; outcome trước action bị từ chối.

PENDING với metrics={} nghĩa chưa chốt, khác số 0 đã đo. NO_OBSERVED_OUTCOME
chỉ hợp lệ tại/sau measurement_window_end khi thực sự đã đo. Báo cáo đến muộn
dùng outcome_id mới, cùng effect_ref, giữ snapshot cũ; không cộng trùng các snapshot.
Dữ liệu thật chỉ ghi ActionRecord **sau khi người có quyền thực sự làm action**
và đã review disclosure/terms/privacy. Fixture compliance_reviewed=true chỉ là
dữ liệu kiểm thử, không phải xác nhận compliance/E2/E3 thật.

Đối chiếu [BR-10a](../docs/architecture/BR-10A-ACTION-STORE.md),
[BR-10b](../docs/architecture/BR-10B-OUTCOME-STORE.md) và
[bộ khởi đầu M03](../starter-kits/M03-tracked-human-action/README.md).

## 7. M04 — mock trước provider

Lưu config vào work; as_of/max_age_hours áp dụng đúng timeline synthetic:

<!-- Mẫu nhập --> <!-- rehearsal:advisor-config.json -->
```json
{"decision_id":"d12","question":"Thiếu bằng chứng gì?","as_of":"2026-09-06T00:00:00Z","max_age_hours":8760}
```

<!-- run:sh:m04 -->
```sh
"$bot" advisor mock "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/advisor-config.json"
```

<!-- run:ps:m04 -->
```powershell
& $bot advisor mock "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/advisor-config.json"
```

Kết quả mong đợi: envelope status=SUPPORTED, artifact.advisor_output.state=HUMAN_REVIEW,
execution_permitted=false, write_tool_requested=false. Các evidence IDs phải
resolve được d12, br09-product-a-t2, br09-commission_rate-t2, br09-price-t2, a12, o12.
Lưu config/as_of/max_age, stdout và upstream hashes; mock không ghi upstream.

Failure: config decision_id mồ côi → CONTEXT_ERROR; max_age_hours=1 → stale
rejection, as_of trước observation → future rejection. Dùng
python scripts/smoke_br11a.py từ repo root để đối chiếu các case model bịa ID/write
qua provider test-double; mock CLI không tự cho chèn model output bằng đổi question.
Xem hướng dẫn [BR-11a](../docs/architecture/BR-11A-MOCK-ADVISOR.md).
Provider thật cần cấu hình/quyền/privacy và evidence riêng; model success không
thành Decision/Approval/Execution hay trusted evidence.

## 8. M05 — đánh giá, đề xuất và review

Lưu ba mẫu vào work. Evaluation chỉ chọn o12 của action a12 thuộc decision d12;
producer hiện hành trả INCONCLUSIVE. Chưa có protocol để suy hiệu quả kinh doanh.

<!-- Mẫu nhập --> <!-- rehearsal:evaluation-config.json -->
```json
{"evaluation_id":"e12","decision_id":"d12","effect_ref":{"effect_kind":"HUMAN_ACTION","effect_id":"a12"},"outcome_ids":["o12"],"evaluated_at":"2026-09-06T00:00:00Z"}
```

<!-- rehearsal:proposal.json -->
```json
{"proposal_id":"p12","evaluation_ids":["e12"],"current_version":"diagnostic/v1","proposed_version":"diagnostic/v2","change_summary":"SYNTHETIC: làm rõ thông báo lỗi input","expected_benefit":"Giúp đọc lỗi, chưa đo lợi ích","risks":["Consumer diagnostic cần review"],"rollback":"Khôi phục diff thông báo trong bản sao lab, giữ nguyên store","auto_apply":false}
```

<!-- rehearsal:review.json -->
```json
{"review_id":"r12","proposal_id":"p12","reviewed_by":"human","reviewed_at":"2026-09-07T00:00:00Z","decision":"REQUEST_CHANGES","reason":"SYNTHETIC fixture: cần regression và rollback; không phải review người thật"}
```

<!-- run:sh:m05 -->
```sh
"$bot" evaluation create "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/evaluations.jsonl" "$work/evaluation-config.json"
"$bot" evaluation list "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/evaluations.jsonl"
"$bot" proposal import "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/evaluations.jsonl" "$work/proposals.jsonl" "$work/proposal.json"
"$bot" proposal list "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/evaluations.jsonl" "$work/proposals.jsonl"
"$bot" review import "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/evaluations.jsonl" "$work/proposals.jsonl" "$work/reviews.jsonl" "$work/review.json"
"$bot" review list "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/evaluations.jsonl" "$work/proposals.jsonl" "$work/reviews.jsonl"
"$bot" history replay "$work/history.jsonl"
```

<!-- run:ps:m05 -->
```powershell
& $bot evaluation create "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/evaluations.jsonl" "$work/evaluation-config.json"
& $bot evaluation list "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/evaluations.jsonl"
& $bot proposal import "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/evaluations.jsonl" "$work/proposals.jsonl" "$work/proposal.json"
& $bot proposal list "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/evaluations.jsonl" "$work/proposals.jsonl"
& $bot review import "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/evaluations.jsonl" "$work/proposals.jsonl" "$work/reviews.jsonl" "$work/review.json"
& $bot review list "$work/history.jsonl" "$work/actions.jsonl" "$work/outcomes.jsonl" "$work/evaluations.jsonl" "$work/proposals.jsonl" "$work/reviews.jsonl"
& $bot history replay "$work/history.jsonl"
```

Expected: mỗi create/import APPENDED, list VALID; cùng input EXACT_DUPLICATE;
Kết quả: e12.result=INCONCLUSIVE, p12.auto_apply=false, r12.decision=REQUEST_CHANGES.
APPENDED chỉ là đã lưu. Review thật phải do người review thực sự thực hiện;
APPROVE_FOR_MANUAL_CHANGE cũng không auto-apply/deploy/execution.

Failure: đổi evaluation_ids của bản sao proposal thành ID mồ côi, hoặc
auto_apply=true; command phải nonzero, không artifact, bytes các store không đổi.
Chọn outcome snapshots theo cùng effect/kỳ đo, không cộng bản cập nhật.
Lưu e12→p12→r12, input/output, hash/store owner và lý do còn INCONCLUSIVE.

Bài thay đổi nhỏ **do learner tự viết**, trong bản sao source riêng:

1. Chọn một thông báo stderr cho JSON input sai ở cmd/bot/main.go; dự đoán thông
   báo giúp sửa lỗi thế nào. Giữ nguyên schema, exit code, stdout và authority.
2. Tự thêm TestLearnerDiagnostic trong cmd/bot; chạy
   go test ./cmd/bot -run '^TestLearnerDiagnostic$' -count=1 để thấy FAIL đúng lý do.
3. Sửa thông báo tối thiểu; chạy test mới rồi go test ./... và go vet ./....
4. Lưu diff/version và PASS. Revert đúng diff đó trong bản sao: test mới phải FAIL;
   khôi phục diff: PASS. Giữ output và explain-back, không sửa expected để làm xanh.

Xem hướng dẫn [BR-12b](../docs/architecture/BR-12B-EVALUATION-STORE.md),
và hướng dẫn [BR-12c](../docs/architecture/BR-12C-PROPOSAL-REVIEW-STORE.md).
python scripts/smoke_br12d.py tự chạy chain/rollback mẫu trong temp; kết quả đó
không chứng minh learner đã tự viết bài thay đổi nhỏ.

## 9. Ghi evidence và tiếp tục học

[Continuity Checkpoint](../starter-kits/CONTINUITY-CHECKPOINT.md) cần
ghi rõ learner_bot_commit_or_version, previous IDs, new_capability_entrypoint,
thành phần sở hữu canonical store, expected/observed, failure, authority ceiling, limitation
và rollback. Ghi note rehearsal trong work; evidence learner thật lưu tại
`workspace/learner/M01…M05`, theo template của starter tương ứng. Thư mục
`workspace/` được `.gitignore` bảo vệ; thư mục `learner/` ở repo root không được
ignore. Không dùng `git add -f` để đưa evidence riêng tư vào PR.

Khi chủ repo tự nghiệm thu cách dùng tài liệu, dùng phiếu CA-T18 ở mục 13 của
[evidence triển khai](../docs/architecture/EVIDENCE-CURRICULUM-ALIGNMENT-20260923.md),
ghi bản riêng tại `workspace/learner/CA-T18.md`; không lấy smoke tự động làm kết quả tự làm.

| Chặng | Evidence thật còn phải có | Khi thiếu |
|---|---|---|
| M00 | 3+ E1 observations và Human DecisionPacket | giữ M00 chưa Reality PASS |
| M01 | input M00, prediction, failure/explain-back | giữ M01 chưa PASS |
| M02 | observation t1/t2, IDs, restart/replay/integrity | giữ M02 chưa PASS |
| M03 | action người làm, quyền/compliance, nguồn đo E2/E3 | dừng ở rehearsal |
| M04 | grounded/stale/abstain và evidence advisor thực tế | không lấy mock làm provider proof |
| M05 | evaluation→proposal→review, diff/rollback E4 | không dùng review fixture thay người |

Dữ liệu thật dùng ID/timestamp/source riêng và context đã review; giữ subject
ổn định, không chép lịch sử synthetic vào store thật. lab/mission-runtime chỉ là
conformance oracle, không phải Bot thứ hai. Không tự sửa PROGRESS.md hoặc reset
bài đã PASS sau lần đồng bộ chương trình này.
