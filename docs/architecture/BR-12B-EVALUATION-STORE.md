# BR-12b — Evaluation liên kết store của bot

Baseline `46f3d66`; dùng `core/m05` sau #69. Phạm vi lab một writer, không API/model hoặc execution, không proposal/review store trong lát cắt này.

## Lệnh

Tại `lab/affiliate-bot`, chạy với ba store BR-10 có sẵn trong cùng workspace. Thay các tên đường dẫn bằng file thực của bạn; không dùng ledger DeepSeek làm store evaluation.

```sh
go run ./cmd/bot evaluation create history.jsonl actions.jsonl outcomes.jsonl evaluations.jsonl evaluation-config.json
go run ./cmd/bot evaluation list history.jsonl actions.jsonl outcomes.jsonl evaluations.jsonl
```

Config ví dụ cho bundle synthetic do `advisor fixture-run` tạo (không cần key):

```json
{
  "evaluation_id": "br12-evaluation-1",
  "decision_id": "br11-decision",
  "effect_ref": {"effect_kind": "HUMAN_ACTION", "effect_id": "br11-action"},
  "outcome_ids": ["br11-outcome"],
  "evaluated_at": "2026-09-06T00:00:00Z"
}
```

Lưu config cạnh history/actions/outcomes trong bundle, chuyển vào thư mục bundle trước khi chạy các lệnh trên. Không dùng thời gian ví dụ cho dữ liệu thực. Không nhận result/evidence/limitations tùy ý từ config. Input bounded 1 MiB, regular file, không symlink cuối; unknown/duplicate fields bị từ chối.

Expected create: status APPENDED, artifact là EvaluationRecord, result INCONCLUSIVE, evidence_ids và outcome_ids bằng danh sách outcome đã chọn (sắp xếp ổn định), execution_permitted=false, auto_apply=false. Lần tạo lại cùng ID/nội dung trả EXACT_DUPLICATE không ghi; khác nội dung trả CONFLICT. List trả VALID và mảng record, đọc lại từ filesystem, không phụ thuộc bộ nhớ lần chạy trước.

## Quy tắc và quyền sở hữu

- Load history/action/outcome bằng validator hiện có; decision phải resolve đúng một lần và replay MATCH. Action phải thuộc decision; mỗi outcome phải tồn tại đúng một lần và EffectRef trùng action.
- Evaluated_at không trước bất kỳ observed_at được chọn. Outcome PENDING có thể được ghi nhận khi cửa sổ còn mở nhưng không là số 0/absence. Các giới hạn đo của outcome vẫn do M03 kiểm.
- Producer `br12-evaluation/v1` luôn INCONCLUSIVE vì chưa có protocol/tiêu chí hiệu quả được review. Kể cả zero hoặc OBSERVED cũng không tự suy ra SUPPORTED/NOT_SUPPORTED. Không cộng snapshot, không biến test PASS thành business evidence. Đây là baseline ghi nhận thiếu cơ sở kết luận, chưa phải thuật toán đo hiệu quả.
- Evaluation đi qua raw M05 schema rồi semantic validator. Đọc store kiểm lại toàn bộ liên kết và tái dựng producer v1; outcome/evidence/limitations/result bị sửa lệch sẽ bị chặn. ID trùng, record hỏng/quá lớn hoặc JSONL thiếu newline cuối đều bị chặn trước append; không sửa dữ liệu cũ tự động.
- File evaluation riêng; chặn path trùng hoặc hardlink/symlink alias với upstream/config. Upstream không bị ghi. Cùng ID tham chiếu không phải content-addressed snapshot: chưa chống thay thế upstream hợp lệ bằng nội dung khác nhưng giữ nguyên ID, local rollback hoặc writer độc hại.
- Dùng JSONL adapter hiện hữu: một writer trong workspace tin cậy, không transaction đa file, không khóa liên tiến trình/fsync cam kết chịu mất điện. File mới quyền 0600. Không thay ownership contract của BR-08, không dùng CLI này với writer đồng thời.

Các lỗi trả nonzero, không artifact: USAGE_ERROR (exit 2), PATH_ERROR, HISTORY_ERROR, ACTION_STORE_ERROR, OUTCOME_STORE_ERROR, STORE_ERROR, CONFIG_ERROR, EVALUATION_ERROR, CONFLICT (exit 1). Không xóa/reset store để vượt lỗi. Kiểm config và upstream; cần giữ nguyên file lỗi để review.

## Kiểm thử và phần còn lại

Regression: PENDING/zero vẫn INCONCLUSIVE; create/list/restart/duplicate/conflict; upstream bytes không đổi; decision/action/outcome orphan, machine effect, thời gian sớm, duplicate ID/key, result injection, null time, store tampering, mất outcome, path alias và framing thiếu newline.

BR-12c còn proposal và human review có version/benefit/risk/rollback, auto_apply=false. BR-12d còn walkthrough M00→M05 và thay đổi nhỏ có regression FAIL/PASS/rollback. BR-12 cha chưa DONE.
