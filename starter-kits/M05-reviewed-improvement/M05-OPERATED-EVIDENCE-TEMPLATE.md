# M05 Operated Evidence (bằng chứng vận hành)

- Kiểm JSON m05-check: input refs, exit code, output/rejection và case broken link; ghi riêng khỏi human review/operated proof:

- EvaluationRecord / evaluation_id (bản ghi đánh giá / mã đánh giá):
- decision_id / outcome_ids (mã quyết định / các kết quả):
- EvaluationRecord.effect_ref (tham chiếu hành động được đánh giá): `effect_kind=HUMAN_ACTION`; điền `effect_id` khớp `ActionRecord.action_id` và cùng `effect_ref` với các OutcomeRecord được tham chiếu:
- evaluation result + limitations (kết quả đánh giá + giới hạn):
- ImprovementProposal / proposal_id (đề xuất cải tiến / mã đề xuất):
- current_version → proposed_version (phiên bản hiện tại → đề xuất):
- expected benefit + risks (lợi ích kỳ vọng + rủi ro):
- rollback (cách quay lui):
- ReviewRecord / review_id (bản ghi review / mã review):
- human review decision + reason (quyết định review của người + lý do):
- next measurement (phép đo tiếp theo):
