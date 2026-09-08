# M07 — JSON grounding boundary (learner CLI)

Trạng thái: **partial offline implementation**. Đây là contract đang được
`core/m07` và learner Bot dùng; không phải bằng chứng một model provider hoặc
n8n workflow đã vận hành. `execution_permitted` luôn `false`.

M07 resolve một `HistoryRecord` canonical/replay `MATCH`, sau đó tạo context
với các evidence ID thật của record và từng source field. Model không được
nhận một danh sách ID để boundary tự gắn vào câu trả lời.

## Contract đầu ra

Đầu ra model là JSON có `state`, `answer`, `claims`, `evidence_ids`,
`tool_calls`, `authority` và `write_permission`.

- Chỉ `HUMAN_REVIEW` hoặc `ABSTAIN`; authority là `A2-RO` và
  `write_permission` phải là literal `false`.
- Với `HUMAN_REVIEW`, mọi claim phải trỏ field/value bằng đúng evidence ID đã
  resolve. `evidence_ids` phải đúng bằng union evidence được claim cite.
- `claim.text` và `answer` phải là render deterministic của các claim đã kiểm.
  Prose tự do của model không được gắn nhãn grounded.
- `ABSTAIN` không được chứa claim hoặc citation.
- `tool_calls` trong final model output bị từ chối. Tool trace không được lấy
  từ lời tự khai của model.

Ví dụ kiểm context và output không có tool evidence:

```text
bot m07 context HISTORY.jsonl RECORD_ID
bot m07 validate HISTORY.jsonl RECORD_ID MODEL_OUTPUT.json REGISTRY.json
```

## Tool-result evidence

Adapter phải preflight request bằng registry **trước** khi request được thực
hiện. Khi nhận response, adapter ghi JSON `ToolResult` (record ID, request,
status 2xx, thời điểm, `redirected=false`, body JSON data), rồi learner Bot
đăng ký artifact bất biến:

```text
bot m07 register-tool-result HISTORY.jsonl RECORD_ID REGISTRY.json TOOL_RESULT.json REGISTERED.json
bot m07 validate HISTORY.jsonl RECORD_ID MODEL_OUTPUT.json REGISTRY.json REGISTERED.json
```

Registration kiểm host/method/scheme/port/userinfo/query theo registry, thời
gian, redirect và status. Trace ID và evidence ID được hash từ response do
adapter ghi; khi validate, hash và binding `record_id` được tính lại. Body chỉ
được expose như `tool_result.body`, claim kind `unknown` và limitation
untrusted; M07 không tự phân loại nó thành product fact.

Các fail-closed case gồm trace ID/body/request/record ID bị sửa, write request,
host/port sai, redirect, status không thành công, claim ID/value sai, prose bịa
và quyền ghi. Retry byte-identical của registration trả `EXACT_DUPLICATE`; file
output khác nội dung không bị ghi đè.

## Persist validated AgentProposal

Chỉ output `HUMAN_REVIEW` đã qua cùng boundary mới có thể được persist:

```text
bot m07 register-proposal HISTORY.jsonl RECORD_ID MODEL_OUTPUT.json REGISTRY.json PROPOSAL.json [REGISTERED_TOOL_RESULT.json]
```

Artifact giữ raw model output, canonical digest/proposal ID và `record_id`.
Proposal phải có `proposed_action` gồm `action_type`, `target` và parameters
JSON object. Khi resolve, digest được tính lại và raw output được validate lại
với canonical context; output `ABSTAIN`, output đã bị sửa hoặc record khác đều
bị từ chối. Artifact này vẫn proposal-only, không phải approval hoặc execution
authority.

M08 agent path dùng thêm proposal artifact: `m08-intent HISTORY REQUEST
M07_PROPOSAL OUT`, và `m08-policy HISTORY INTENT POLICY M07_PROPOSAL OUT`.
`proposal_ref`, action type, target, exact parameters và evidence của intent
phải bind với proposal khi tạo intent; policy resolve lại artifact sau đó. Đường
human cũ không nhận proposal artifact.

## Kiểm offline

Từ root repo:

```text
(cd core && GOWORK=off go test ./m07)
(cd lab/affiliate-bot && GOWORK=off go test ./cmd/bot -run TestM07 -count=1)
python3 scripts/validate_n8n_m07_adversarial.py
python3 scripts/validate_n8n_m07_output_cases.py
```

Hai script cuối chạy learner CLI thật; chúng không thực thi JavaScript node của
blueprint. Harness `demo m07-check` là fixture legacy/conformance cục bộ, không
phải authoritative proof cho path learner hoặc workflow.

## Giới hạn còn mở

Blueprint n8n gọi các endpoint loopback chung theo thứ tự `preflight → GET
full response/no redirect → register tool result → context → model → validate
→ register proposal`. Nó không còn có code node tự quyết định registry hoặc
grounding. Tuy nhiên repo chưa import/chạy blueprint trên một n8n instance,
chưa có transport seam kiểm response-size/private-address policy hoặc parity
execution thật với CLI. Vì vậy BR-15/RP-05 vẫn **chưa hoàn tất**; không dùng
tài liệu này để tuyên bố model/n8n/provider đã grounded hoặc operated.
