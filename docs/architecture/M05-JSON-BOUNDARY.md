# BR-03c.2 — Boundary JSON M05

Phạm vi: kiểm EvaluationRecord → ImprovementProposal → ReviewRecord cùng một ActionRecord và các OutcomeRecord M03, qua lệnh chỉ đọc. Đây là conformance harness, không phải store/learner integration BR-08/BR-12 và không đóng BR-03 tổng thể.

Trạng thái IN_REVIEW: [PR #29](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/29), code/tests `e02c915`; Codex thực hiện, chủ repo chưa review. Base main sau #28, chưa merge PR M05.

## Lệnh và input

Từ repo root:

```text
cd lab/mission-runtime
go run ./cmd/demo m05-check testdata/m03-action.json testdata/m05-outcomes.json testdata/m05-evaluations.json testdata/m05-proposal.json testdata/m05-review.json
go test ./cmd/demo -run 'TestM05' -count=1 -v
```

Thứ tự 5 file: action object, outcomes array không rỗng, evaluations array không rỗng, proposal object, review object. Mỗi phần tử array có canonical schema riêng; array file/envelope không phải artifact schema mới. Không hỗ trợ nhiều action trong một lần gọi. File do người vận hành chọn, chưa có quota kích thước cho endpoint công khai.

Fixture hoàn toàn synthetic, dùng lại `syn-a/syn-d/syn-o` của bài kiểm M03. `compliance_reviewed:true` và `reviewed_by:human` chỉ mô phỏng dữ liệu; không chứng minh có người thực sự review. Output phải giữ evaluation INCONCLUSIVE, review REQUEST_CHANGES, `result:VALID`, `execution_authorized:false`, `auto_apply:false`.

## Schema trước semantic

Tất cả 5 input được kiểm schema/strict decode trước semantic. Dùng package contracts đã pin từ BR-03b; không thêm schema hoặc dependency mới. Required/null/type/enum/date-time, key trùng, field lạ, case alias, nested EffectRef và uniqueness theo canonical schema được cưỡng chế. Không nhận alias `action_id` nội bộ của EvaluationRecord/M11.

`DecodeM05Evaluation/Proposal/Review` trả VALID chỉ về cấu trúc. `CheckM05Chain` mới kiểm chuỗi:

- Giữ ValidateHumanActionRecord, ValidateActionOutcomeLink, ValidateEvaluationRecord, EvaluateImprovementProposal, ValidateProposalEvaluationLink và ValidateReviewRecord. Zero trước window end vẫn bị MEASUREMENT_WINDOW_OPEN; PENDING không bị ép thành kết luận cả kỳ.
- Mọi outcome/evaluation cung cấp đều được kiểm, kể cả không được proposal tham chiếu. ID outcome/evaluation trùng bị DUPLICATE_ID trước khi map có thể ghi đè. Proposal lặp evaluation ID cũng bị DUPLICATE_ID ở semantic boundary (schema hiện không có uniqueItems cho trường đó).
- Evaluation phải cùng decision/effect với action và resolve từng outcome ID. Proposal resolve evaluation IDs; review resolve proposal ID. Các ID không được tự tạo hoặc sửa để qua kiểm.
- Kiểm timeline bổ sung ở boundary M05: evaluated_at không trước observed_at của outcome được tham chiếu; reviewed_at không trước evaluated_at của evaluation trong proposal. So sánh instant, nhận đúng bằng và timezone tương đương. State lần lượt EVALUATION_BEFORE_OUTCOME/REVIEW_BEFORE_EVALUATION. Không sửa typed validators dùng chung M11 trong đợt này.
- Proposal vẫn cần version khác nhau, change/benefit/rollback không rỗng. Blank evaluation ID/review ID/reason bị từ chối. Schema const chặn auto_apply=true và reviewed_by=agent bằng INVALID_SCHEMA; test typed cũ giữ REJECT_AUTO_APPLY/BROKEN_LINK.

## Output và field tùy chọn

Envelope chứa action/outcomes/evaluations/proposal/review, result và execution_authorized=false. Từng artifact serialize được kiểm canonical schema trước xuất. Lỗi input/schema/semantic trả stderr + exit 1, stdout không có success envelope; lỗi writer không bị bỏ qua. VALID nghĩa là kiểm record thành công, **không phải review APPROVE**: cả REJECT và REQUEST_CHANGES cũng là ReviewRecord hợp lệ, decision được giữ nguyên.

EvaluationRecord hỗ trợ `notes` tùy chọn bằng `*string`: giữ được chuỗi rỗng có chủ đích và nội dung, không bỏ field schema cho phép. Proposal `risks` không bắt buộc trong schema: khi nil/rỗng, serializer bỏ field thay vì xuất null sai schema; list có nội dung giữ nguyên. Canonical schema không đổi. Đây là thay đổi JSON tag của shared type; regression M03–M11 vẫn phải PASS. Không tự điền risk giả hoặc coi thiếu risks là đủ review nội dung.

## Bài lỗi có chủ đích

Copy các file vào thư mục bài tập riêng, thay path trong lệnh rồi:

1. Đổi evaluation.outcome_ids sang ID không có: BROKEN_LINK.
2. Bỏ proposal.auto_apply hoặc đổi thành true: INVALID_SCHEMA, không tự thêm false.
3. Đổi review.proposal_id sang ID khác: BROKEN_LINK; đổi reviewed_by=agent: INVALID_SCHEMA.
4. Đổi reviewed_at về trước evaluation: REVIEW_BEFORE_EVALUATION.
5. Khôi phục từng sửa đổi và chạy lại; giải thích vì sao REQUEST_CHANGES vẫn cho result VALID nhưng không được áp dụng thay đổi.

Lệnh không ghi file, merge code, apply proposal, fetch dữ liệu hoặc execute. Bản gốc input không bị chỉnh để chạy test. Không ghi PROGRESS chỉ vì fixture PASS.

## Giới hạn và evidence

Không resolve DecisionPacket hoặc evidence_ids với canonical store, không xác thực danh tính human, nội dung review, rollback thực hiện được, nguồn báo cáo, attribution hoặc causal support. evaluated_at/reviewed_at do file cung cấp; schema/timeline không chứng minh timestamp đáng tin. Proposal không có timestamp nên không thể kiểm thời điểm tạo proposal. VALID không thay Reality/Operated proof.

Test red/green: wrapper json.Unmarshal ban đầu không chặn required/null, key lạ/trùng, auto-apply hoặc machine reviewer; TestM05RawSchema FAIL. Sau schema-first PASS. Có tests cả chuỗi, ID trùng, window, timeline/offset, optional notes/risks, output schema, file không đổi, writer failure và fixture thực. Raw eval chạy cùng cases.json với expected riêng chỉ cho hai const violations trong raw-expectations.json; typed eval cũ giữ nguyên.

BR-03c còn các boundary khác (M06–M11 và phần output/tích hợp chưa nối); riêng M05 này không hoàn thiện BR-12 pipeline hay live proof.

Kiểm local: tests/vet ba module PASS, 8 Python validators và 10 Python regressions PASS, git diff --check PASS. Smoke lệnh thật với fixture xuất VALID, INCONCLUSIVE, REQUEST_CHANGES và execution_authorized=false. CI phải kiểm theo head PR riêng.
