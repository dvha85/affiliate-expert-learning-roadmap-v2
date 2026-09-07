# BR-12d — Walkthrough M00→M05 và regression/rollback

Baseline `9dd9070` (#72). Phạm vi nghiệm thu: lab/schema, một workspace một writer; không chứng minh chất lượng tư vấn, hiệu quả affiliate hoặc pilot người mới. Codex thực hiện và review kỹ thuật theo yêu cầu chủ repo. ReviewRecord sinh bởi smoke được gắn rõ synthetic, không thay bằng chứng người thật review.

Bàn giao qua [PR #73](https://github.com/dvha85/affiliate-expert-learning-roadmap-v2/pull/73), implementation `7cf10b3`; xem trạng thái merge/CI tại PR. Smoke local, learner tests/race/vet và 8 validators PASS. Review không có lỗi chặn trong phạm vi lab; nghiệm thu DONE trong kế hoạch áp dụng khi PR này được merge, không đóng các giới hạn bên dưới.

## Chạy lại

Từ repo root, có Git, Python 3 và Go theo learner go.mod:

```sh
python3 scripts/smoke_br12d.py
```

Có thể đặt GO_BIN theo đường dẫn Go. Script build binary và tạo hai thư mục tạm, không chạm PROGRESS, evidence cá nhân, ledger DeepSeek hoặc key. Mỗi lệnh bot chạy bằng process mới; hết phiên script dọn đúng thư mục tạm nó tạo. Không gọi model hoặc platform. CI chạy cùng script trong deterministic-runtime.

## Walkthrough và ma trận

| Bước | Artifact/kiểm tra | Expected |
|---|---|---|
| M00 import | examples/m00-import/packet-t2.json → observations | Nguồn synthetic, giữ product và price/commission provenance |
| M01/M02 | history capture → DecisionPacket d12 → replay | evidence_ids=[br09-product-a-t2], action=null, MATCH |
| M03 | Action a12 → Outcome o12 | HUMAN_ACTION/a12, PENDING, metrics={} |
| M04 | Mock advisor đọc cùng history/actions/outcomes | HUMAN_REVIEW, đủ 6 ID gồm hai source fields; không write |
| M05 evaluation | e12 → d12/a12/o12 | INCONCLUSIVE; evidence_ids=[o12], không coi pending là zero |
| M05 proposal | p12 → e12 | version/benefit/risk/rollback đầy đủ, auto_apply=false |
| M05 review | r12 → p12 | REQUEST_CHANGES, reason ghi synthetic; không actual human approval |
| Restart | list evaluation/proposal/review và replay bằng process mới | Artifact giống nội dung đã tạo; upstream bytes không đổi |
| Negative | Review duplicate; evaluation upstream rỗng | EXACT_DUPLICATE không ghi; PROPOSAL_STORE_ERROR không artifact |

Tests BR-12b/c bổ sung orphan/time/duplicate/conflict/schema/framing/path guards; smoke này không thay chúng. Review chỉ nối proposal/evaluation, không biến evaluation thành kết luận có hiệu quả.

## Thay đổi nhỏ: nhãn thao tác diagnostic

Trước sửa, cả `proposal import` và `proposal list` đều có command="proposal" (tương tự review). Điều này làm mất thông tin thao tác trong output. Sau sửa command ghi rõ thao tác import/list, kể cả lỗi cú pháp thiếu file. Không thay canonical EvaluationRecord/ImprovementProposal/ReviewRecord hay quyền thực thi. Consumer so sánh command cũ phải cập nhật; đây là rủi ro tương thích có chủ đích.

Regression `TestImprovementCommandIdentifiesOperation` được chạy trước sửa và thật sự FAIL: `operation label: got proposal, want proposal import`. Sau sửa PASS. Test chỉ gọi trường hợp thiếu tham số, không mở store; smoke trên còn kiểm nhãn các lệnh thành công/error đã đi qua filesystem.

Diff được script in ra gồm thêm biến command với import/list và đổi duy nhất giá trị command trong envelope. Review kỹ thuật: whitelist chỉ hai thao tác đã có, không đưa input arbitrary vào nhãn; auto_apply/execution_permitted vẫn false. Không có hook sửa code hoặc executor trong proposal/review.

Diễn tập rollback: sao chép chỉ source/schema/module files được Git theo dõi vào workspace tạm (cộng file regression đã biết), phục hồi nguyên file improvement_store.go trước sửa. Git blob SHA1 phải bằng `a3363b51a073a83a4b6006834eaec2d99930069d`, là file tại baseline `9dd9070`. Không yêu cầu shallow CI có lịch sử Git baseline; hash được xác minh từ bytes. Không rollback repo người dùng hoặc toàn bộ version hệ thống. Giữ regression mới để quan sát behavior cũ thất bại, rồi phục hồi file sửa và kiểm PASS.

Bằng chứng chạy local:

```text
M00->M05 PASS: exact IDs; fresh-process list/replay; INCONCLUSIVE; synthetic review; no execution
regression_exit_codes: before=1, after=0, rollback=1, restore=0
rollback_blob: a3363b51a073a83a4b6006834eaec2d99930069d
business_effectiveness: NOT_MEASURED
human_review_fixture_only: true
execution_permitted: false
BR-12d PASS: synthetic continuity and isolated FAIL/PASS/rollback; no business or human-pilot proof
```

Nếu file implementation đổi sau này, rehearsal sẽ fail khi hash không khớp: cần review/cập nhật bài thực hành, không xóa assertion để làm PASS. Log script in diff và mã thoát để lưu từ CI; tài liệu này là bản tóm tắt kết quả, không business OutcomeRecord.

## Nghiệm thu BR-12 trong lab

BR-12a #69 chia sẻ M05; BR-12b #70/#71 tạo EvaluationRecord linked store; BR-12c #72 nhập proposal/human review; BR-12d khép walkthrough cùng workspace và bài sửa nhỏ có FAIL/PASS/rollback. Đủ checklist triển khai lab sau khi PR BR-12d được review/merge; không yêu cầu API hoặc action thật để làm đẹp bằng chứng.

Giữ giới hạn: một writer trusted local filesystem, khai báo human/source/time không được xác thực, không transaction/CAS/fsync recovery, không content-pin upstream, không chọn review hiệu lực để execute, producer evaluation luôn INCONCLUSIVE khi thiếu protocol. Chưa đo hiệu quả cải tiến, chưa human pilot, chưa campaign thật. Những phần này không được đánh DONE theo nghiệm thu lab; BR-06b/BR-10d và BR-16 vẫn giữ phạm vi riêng.

Bước triển khai kế tiếp theo kế hoạch là BR-13 watcher M06; không mở quyền gọi nguồn/platform thật khi chưa được cấp quyền và chốt nguồn.
