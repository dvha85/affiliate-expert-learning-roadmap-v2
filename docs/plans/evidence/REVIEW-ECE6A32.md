# Baseline review R01–R16 tại ece6a32

- Review thực hiện ngày 07/09/2026 trên `ece6a32619e5b9a05d0599b87f50023f38931cb9`, checkout sạch và khớp remote main khi review.
- Đây là bản ghi kết quả review trước PR kế hoạch; không phải kết quả test của code đã sửa. [Kế hoạch xử lý](../REVIEW-REMEDIATION-PLAN.md).
- Phạm vi: contracts/core/learner/harness, M06/M07 blueprint, store/backup/mission entrypoint, CI/readiness và walkthrough/runbook.
- CLI chạy binary build từ baseline. JavaScript được lấy nguyên văn từ blueprint và chạy bằng Node với biến đầu vào n8n được mô phỏng; không phải operated n8n execution. HTTP adapter được chạy bằng loopback; không gọi LLM/provider hoặc live executor.

Các source links dưới đây trỏ file trong repo; vị trí có thể đổi sau fix. Dùng baseline SHA để đối chiếu lịch sử. Script thử và raw logs của lượt review nằm ở môi trường tạm của reviewer, **chưa được commit thành regression suite**. Bảng này là record của người review; mỗi PR sửa phải thêm fixture/test tái hiện bền vững, không dựa vào đường dẫn máy cá nhân hoặc coi bảng là executable proof.

| ID / ưu tiên | Source baseline | Ca kiểm và kết quả đã quan sát |
|---|---|---|
| R01 / P1 | [core M07](../../../core/m07/m07.go), [blueprint M07](../../../lab/n8n/M07-readonly-evidence-agent.blueprint.json) | `answer` và claim text bảo đảm lợi nhuận một triệu USD/ngày, không rủi ro; field/value và ID của snapshot synthetic đúng: CLI VALID, JS boundary nhận |
| R02 / P1 | [blueprint M06](../../../lab/n8n/M06-readonly-watcher.blueprint.json) | Full ASCII → APPENDED/MATCH; thiếu price → APPENDED, ACK=true, replay DRIFT; tên A&B → INVALID_HISTORY/input_hash mismatch; giá âm bị schema từ chối đúng |
| R03 / P1 | [M07 context](../../../lab/affiliate-bot/cmd/bot/m07.go), [mission](../../../lab/affiliate-bot/cmd/bot/mission_command.go), [blueprint M07](../../../lab/n8n/M07-readonly-evidence-agent.blueprint.json) | ID tự dựng `br09-product-a-t2#price` được JS nhận nhưng CLI ABSTAIN; ID thật `br09-price-t2` M07 VALID nhưng M08 REJECTED/not linked |
| R04 / P1 | [watcher HTTP](../../../lab/affiliate-bot/cmd/bot/watcher.go), [mission](../../../lab/affiliate-bot/cmd/bot/mission_command.go) | Record thiếu price có DRIFT: M07 context HISTORY_ERROR; cùng record HTTP GET 200, M08 APPENDED |
| R05 / P1 | [blueprint M07](../../../lab/n8n/M07-readonly-evidence-agent.blueprint.json) | GET example.com/a preflight ALLOW_READ_ONLY nhưng boundary throw TOOL_CALLS_MUST_BE_RECORDED_BY_PREFLIGHT; đọc graph không có bước đăng ký response thành canonical evidence |
| R06 / P1 | [mission](../../../lab/affiliate-bot/cmd/bot/mission_command.go), [harness M08](../../../lab/mission-runtime/cmd/demo/m08_boundary.go) | Agent thiếu proposal_ref và parameters=null: learner APPENDED/ALLOW, harness INVALID_SCHEMA; created_at 2090/policy now 2026: learner ALLOW, harness WAIT/INTENT_NOT_YET_VALID |
| R07 / P2 | [mission JSON decoder](../../../lab/affiliate-bot/cmd/bot/mission_command.go) | Raw parameters.offer_id=9007199254740993 trở thành 9007199254740992 trong intent được APPENDED |
| R08 / P1 | [mission bind/canary](../../../lab/affiliate-bot/cmd/bot/mission_command.go) | Dùng hết grant 1 lần/10 minor; tăng cap cùng grant/approval được ACK và reserve lại; bind i1→i2→i1 rồi dùng lại approval/grant cũng RESERVED. Không sửa trực tiếp state |
| R09 / P1 | [mission reserve](../../../lab/affiliate-bot/cmd/bot/mission_command.go) | Import approval còn hạn, tạo grant, chờ expiry thật rồi process mới reserve: RESERVED |
| R10 / P1 | [mission state](../../../lab/affiliate-bot/cmd/bot/mission_command.go) | 24 process reserve cap=1: lượt thứ ba có 2 RESERVED nhưng state cuối executions_used=1. HTTP append đồng thời thử riêng không tái hiện duplicate; không báo đó là lỗi đã chứng minh |
| R11 / P1 | [mission output writer](../../../lab/affiliate-bot/cmd/bot/mission_command.go) | Trên copy tạm, HISTORY=OUT của m08-intent: APPENDED, history bị thay bằng intent; replay lỗi corrupt. Không làm hỏng store của người dùng |
| R12 / P1 | [backup inventory](../../../lab/affiliate-bot/cmd/bot/backup_command.go) | advisor fixture-run tạo nested bundle trong source; BACKED_UP/RESTORED nhưng bundle không có trong đích |
| R13 / P1 | [restore validation](../../../lab/affiliate-bot/cmd/bot/backup_command.go) | Tạo artifact bằng Bot rồi mô phỏng orphan action ở source; backup create sinh checksum đúng; restore RESTORED nhưng action list STORE_ERROR/decision không resolve. Không sửa manifest để vượt checksum |
| R14 / P1 | [BR-16a](../../../scripts/smoke_br16a_offline.py), [learner mission](../../../lab/affiliate-bot/cmd/bot/mission_command.go) | Đọc chain: proposals/reviews không dùng; bỏ M04, phần proposal/review M05, watcher M06; tự tạo human intent thay vì dùng proposal M07; learner M11 chỉ STOP/status |
| R15 / P2 | [Curriculum CI](../../../.github/workflows/curriculum-ci.yml), [Mission CI](../../../.github/workflows/mission-agent-path-ci.yml), [M06 cases](../../../scripts/validate_n8n_m06_cases.py) | Hai workflow không gọi BR-16a/BR-18b smoke; M06 cases tự test canonical Python/dictionary ACK; M07 adversarial đã gọi CLI nhưng chưa gọi JS boundary |
| R16 / P2 | [BR plan](../BEGINNER-READINESS-PLAN.md), [matrix](../READINESS-MATRIX.json), [audit](../../../scripts/audit_readiness.py) | Baseline plan ghi BR-15…18 IMPLEMENTED_OFFLINE, BR-19 IMPLEMENTED; matrix các mục tương ứng PARTIAL; audit vẫn PASS. PR kế hoạch đính chính phần mô tả, chưa sửa cơ chế audit |

## Tests baseline và giới hạn

Kết quả review đã chạy: test không dùng cached result và vet cả bốn Go module PASS; 10 Python regression tests PASS; 13 validators và readiness audit PASS; quickstart clone/cache trống và intentional FAIL/fix PASS; smoke BR-12d, BR-13b, BR-13c, BR-16a, BR-18b PASS trong phạm vi hiện có. Các ca đối chứng ở bảng vẫn tái hiện lỗi dù suite hiện hành xanh.

Một số lần đầu bị sandbox chặn loopback/cache/network; đã chạy lại với phạm vi/quyền phù hợp. Không coi lỗi môi trường đó là bug repo. Không suy từ test process restart thành durability khi mất điện. Không suy từ local PASS thành GitHub Actions PASS ở head PR sau này.

Learner vẫn xuất `execution_permitted=false`; lỗi budget/approval là lỗi guard/persistence offline, không phải bằng chứng có chi tiêu hay thao tác live trái phép. M00–M05, quickstart, fixture watcher và giới hạn synthetic/business evidence vẫn có giá trị. M06 ACK, M07 CLI validation, init không overwrite state và backup checksum/replay là cải thiện thật nhưng chưa đủ đóng các phát hiện trên.
