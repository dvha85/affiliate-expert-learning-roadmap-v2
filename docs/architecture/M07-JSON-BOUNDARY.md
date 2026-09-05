# BR-03c.4 — M07 JSON boundary

IN_REVIEW: [PR #31](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/31), code/tests `b652b27`; Codex thực hiện, chủ repo chưa review. Local tests/vet ba module, 8 validators, 10 Python regression tests và diff check PASS. CI phải kiểm theo head PR riêng.

Phạm vi: conformance offline, không model/fetch/persist/n8n. `ToolRegistry` canonical là mảng không rỗng; ToolSpec chỉ là phần tử. AgentProposal không có schema canonical: decoder kiểm shape local state/answer/evidence_ids/tool_calls, không gọi nó là AdvisorOutput. Context là mảng ID do người cung cấp, không phải payload evidence đáng tin.

Từ repo root:

```text
cd lab/mission-runtime
go run ./cmd/demo m07-check testdata/m07-proposal.json testdata/m07-registry.json testdata/m07-evidence-ids.json
go test ./cmd/demo -run TestM07 -count=1 -v
```

Fixture synthetic; output gồm proposal, registry, result SUPPORTED và execution_authorized=false. Schema được kiểm trên toàn registry input/output. Proposal serialize được kiểm lại shape local. Thiếu/null, key trùng, field lạ, sai enum/type, evidence ID rỗng/trùng bị từ chối. Arrays rỗng được nhận cho proposal/context nhưng registry canonical không cho rỗng. Answer phải không chỉ whitespace; state nhận PROPOSE/HUMAN_REVIEW/ABSTAIN.

Registry trùng tên hoặc host trùng không phân biệt hoa thường → INVALID_REGISTRY. Host registry phải là hostname không kèm scheme/path/query/userinfo/port. Call chỉ GET/HEAD trên HTTPS, đúng tool và host, không userinfo/explicit port. Đây là profile hẹp của checker; allowlist trong file không tự trở thành trusted policy vận hành. Schema const read_only=true chặn write registry trước semantic.

Calls và reference IDs được kiểm ngay cả khi state ABSTAIN, tránh early return của typed helper bỏ qua write request. Sau đó giữ EvaluateAgentProposal để kiểm semantics hiện có. Không sửa global helper/typed eval. E02/E04 registry rỗng và E03 registry ghi có raw result INVALID_SCHEMA; expected typed cũ REJECT_TOOL/REJECT_UNGROUNDED vẫn giữ. Test riêng kiểm unknown evidence với registry hợp lệ; không coi schema reject thay tất cả semantic cases.

Lỗi đọc/schema/profile/semantic trả exit 1, không có success envelope; writer error được truyền về. ABSTAIN hợp lệ xuất result ABSTAIN. HUMAN_REVIEW của proposal vẫn được giữ trong output dù helper trả SUPPORTED; không hiểu SUPPORTED là người đã review. Không thay decision, không tự gắn evidence ID từ context để làm grounded.

## Bài thực hành

Copy ba fixture vào thư mục bài tập, thay path lệnh: đổi GET thành POST → REJECT_TOOL, kể cả đổi state sang ABSTAIN; đổi evidence ID sang missing → REJECT_UNGROUNDED; bỏ read_only hoặc đặt false → INVALID_SCHEMA. Khôi phục sửa đổi rồi chạy về PASS. Không sửa registry để hợp thức hóa write request. Không cập nhật PROGRESS từ fixture.

## Giới hạn

SUPPORTED chỉ kiểm membership ID và registry/call, không chứng minh claim có evidence hỗ trợ, freshness, provenance hoặc tool registry được phê duyệt. Không tải URL, nên chưa thực thi timeout/redirect enforcement trên network; các phần đó thuộc BR-15. Text injection trong eval cũ vẫn là dữ liệu, không được parse thành cấu hình/quyền. Không có đường nạp raw tool text vào CLI này, không claim live injection proof.

Input file chưa có size quota; chỉ dùng file do người vận hành chọn, chưa là public endpoint. Tests kiểm schema/shape, required/null/duplicate, registry duplicate, host/method/port/userinfo, ABSTAIN write, IDs, CLI output/read-only/writer failure và raw eval song song typed eval. BR-03 còn M08–M11 và các integration chưa hoàn tất.
