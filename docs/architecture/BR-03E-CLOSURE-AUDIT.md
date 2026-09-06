# BR-03e — Audit điều kiện đóng BR-03

Ngày 06/09/2026; baseline `e0d3558552fc1968e2a69e55be8080e604170350` (main sau #42). Người audit: Codex; chờ chủ repo review. Đây là audit code/test local, không phải chứng nhận vận hành. PR này chỉ tài liệu, không sửa runtime hoặc schema.

## Kết luận

BR-03 **chưa đủ điều kiện DONE**. Có đường kiểm cho cả 29 file schema, nhưng có lỗi timestamp tái hiện được và chưa có ma trận nghiệm thu output thực theo nhánh. Không cần biến provenance, chống rollback hay crash consistency thành yêu cầu vô hạn của BR-03: tách backlog riêng với giới hạn rõ, không tuyên bố sẵn sàng live.

Các phần đã merge: d.1 M09 #39 `d009d1f`; d.2 M10 #40 `1c974ca`; d.3 M11 #41 `e4616a2`; d.4 thời gian #42 `e0d3558`. Tài liệu IN_REVIEW trong các PR cũ là ảnh chụp lịch sử, không phải trạng thái mới nhất.

## Ma trận 29 schema và phạm vi bằng chứng

Tên trong bảng là basename, đều có hậu tố `.schema.json`. Boundary source và tests cùng tên `_test.go` nằm trong [cmd/demo](../../lab/mission-runtime/cmd/demo/). Có decoder không đồng nghĩa mọi caller typed đã được bảo vệ.

| Nhóm / schema | Boundary và bằng chứng | Phạm vi/giới hạn |
|---|---|---|
| M02: observation, history-record, decision-packet | [history_schema.go](../../lab/affiliate-bot/cmd/bot/history_schema.go), history.go/tests, decision_packet.go/tests | Capture/load/append và adapter; history ranked:null legacy được xử lý có chủ đích, không rewrite |
| M03: action-record, outcome-record, effect-ref | [m03_boundary.go](../../lab/mission-runtime/cmd/demo/m03_boundary.go), CheckM03Pair, output marshal+schema | EffectRef qua `$ref` của outcome/evaluation; không phải endpoint standalone; ActionID alias nội bộ không là canonical input |
| M04: advisor-output | [advisor_boundary.go](../../lab/mission-runtime/cmd/demo/advisor_boundary.go), advisor_boundary_test.go | Validator chuyên biệt và schema drift test; không chứng minh nội dung evidence đúng |
| M05: evaluation-record, improvement-proposal, review-record | [m05_boundary.go](../../lab/mission-runtime/cmd/demo/m05_boundary.go), CheckM05Chain | Raw collection + chain và serialize output; không apply proposal |
| M06: observation (dùng lại) | [m06_boundary.go](../../lab/mission-runtime/cmd/demo/m06_boundary.go), ValidateM06Observation | Normalizer/file offline; chưa là source connector/store live |
| M07: tool-registry | [m07_boundary.go](../../lab/mission-runtime/cmd/demo/m07_boundary.go), DecodeM07Registry | Canonical toàn registry; AgentProposal là local shape, không giả làm AdvisorOutput |
| M08: action-intent, policy-decision | [m08_boundary.go](../../lab/mission-runtime/cmd/demo/m08_boundary.go) | Raw file/semantic policy; không cấp quyền effect |
| M09: approval-record, execution-authorization, execution-record | [m09_boundary.go](../../lab/mission-runtime/cmd/demo/m09_boundary.go), [m09_persistence.go](../../lab/mission-runtime/cmd/demo/m09_persistence.go) | APPROVED_LIVE profile, state envelope và artifact có mặt được kiểm; typed trust context không được xác thực bằng schema |
| M10: canary-grant, canary-grant-approval, trusted-cost-bound, canary-ledger, canary-gate-decision | [m10_boundary.go](../../lab/mission-runtime/cmd/demo/m10_boundary.go), m10_chain.go, m10_persistence.go/tests | Thêm execution authorization/record profile GOVERNED_CANARY dùng schema chung; reader/writer ledger gắn exact grant |
| M11: production-lease, production-lease-approval, production-health-snapshot, production-ledger, production-gate-decision, production-activation-record, production-reconciliation-resolution, production-cycle-record | [m11_boundary.go](../../lab/mission-runtime/cmd/demo/m11_boundary.go), m11_chain.go, m11_persistence.go/tests | Dùng lại cost/authorization/execution; closed_cycle và resolved_stop không cấp execute/resume; ledger/activation reader+writer gắn lease |

Đếm unique: 3 + 3 + 1 + 3 + 1 + 2 + 3 + 5 + 8 = 29 (observation M06 và shared execution/cost không đếm lại). Package [contracts](../../contracts/go.mod) pin jsonschema/v6 v6.0.2. Phân biệt canonical artifact, local envelope/context và summary; không yêu cầu mọi struct nội bộ có schema canonical.

## Bắt buộc trước khi đóng BR-03

### E-01 — P1: bảo toàn timestamp khi serialize (cần PR sửa riêng)

Bằng chứng: m10.go `normalizedCanaryLedger`/`AuthorizeCanary` và m11.go `normalizedProductionLedger`/`AuthorizeProduction` dùng `Format(time.RFC3339)` tại các dòng 270/398 và 360/573 của baseline, trong khi parser nhận nanosecond. d.4 chống overflow nhưng chưa sửa precision.

Probe local tạm đã chạy rồi gỡ, không đưa vào PR tài liệu: tạo baseM11; đặt Health.ObservedAt=`2026-09-03T07:55:00.000000001Z`; refreshHealthTrust; gọi AuthorizeProduction tại Now=`2026-09-03T08:00:00Z`. Kết quả thực:

```text
status=AUTHORIZED
authorized_at=2026-09-03T08:00:00Z
expires_at=2026-09-03T08:00:00Z
```

Expiry trước format còn 1ns nhưng sau format bằng authorized_at. DecodeM11Artifact("authorization") yêu cầu expires_at > authorized_at: output runtime tự mâu thuẫn với boundary semantic. TestMissionHealth không được suy là phủ trường hợp này: TestM11HealthExactExpiry hiện không assert authorization trả về ở nhánh còn 1ns. Đây không phải bằng chứng thực thi trái phép: executor còn guard độc lập; lỗi đã xác nhận là artifact/expiry không nhất quán.

Window reset tương tự ghi bỏ phần lẻ; theo code có thể đưa window start lùi dưới một giây và làm lệch lần reset kế tiếp. Đây là suy luận từ code, chưa chạy probe executor nhiều chu kỳ trong audit này.

Nghiệm thu: chọn bảo toàn RFC3339Nano hoặc từ chối precision không hỗ trợ một cách tường minh; không làm tròn nới quyền. Regression M10/M11 gồm 1ns trước/đúng/sau expiry, reset window nhiều chu kỳ, serialize→decode boundary→executor/restart; authorization được cấp luôn có expiry thực sự sau authorized_at và không vượt bất kỳ expiry nguồn. Không thay hash artifact cũ ngầm.

### E-02 — P2: ma trận output runtime theo nhánh (thiếu bằng chứng, chưa kết luận có lỗi)

TestM10RawArtifacts/m10ArtifactRaw và TestM11RawSchema/m11Raw dùng authorization/gate thật nhưng execution fixture được dựng CANCELLED/NOT_PERFORMED. Chain tests bổ sung đường thực thi, song không thể suy các bộ đó phủ mọi trạng thái output chỉ vì suite PASS.

Nghiệm thu PR riêng: liệt kê return branches có artifact khác zero cho gate/authorization/execution/cycle M09–M11; nối từng nhánh với test serialize artifact thật và ValidateRaw/decoder đúng profile. Bao gồm success, cancellation/failure, unknown effect/reconciliation, STOP/WAIT/DENY có artifact; zero-value sentinel không xuất thành canonical success artifact. Ghi rõ branch không áp dụng và lý do. Giữ tests bad raw required/null/extra/duplicate/enum/time và persistence hiện hữu. Không cần tạo adapter live để hoàn thành ma trận.

### E-03 — P2: quyết định nghiệm thu và dọn trạng thái (tài liệu)

Sau E-01/E-02: cập nhật ma trận với test cụ thể, đánh dấu checklist BR-03 đã được chứng minh; chủ repo review giới hạn và chấp thuận đóng. Không dùng bảng cũ ghi “M09 reader còn mở” làm backlog mới. Audit này mới đưa tiêu chí, chưa tự đánh dấu DONE.

## Tách hardening, không dùng để kéo dài BR-03

| Backlog đề xuất | Vì sao riêng / tiêu chí trước triển khai thật |
|---|---|
| H-01 provenance/trusted clock, approval/lease/health registry | Hash và expected object caller không là xác thực. Cần threat model, nguồn trust, thu hồi và tests giả mạo; không coi CONSISTENT_UNVERIFIED là quyền |
| H-02 rollback và migration legacy | Marker local không chống xóa cả ledger+marker/activation hoặc khôi phục snapshot cũ. Cần quyết định migration, backup, monotonic revision/CAS và test rollback; không xóa marker để retry |
| H-03 crash/concurrency durability | Temp+rename/lock hiện hữu không là transaction đa file. Cần fault injection, recovery sau crash và fsync policy trước claim durable production |
| H-04 source/store integration và operated evidence | Thuộc tích hợp/pilot, không schema conformance. BR-06b vẫn chờ chương trình/kênh; fixture không thay bằng chứng vận hành |

Các mã H là đề xuất backlog trong audit, chưa phải triển khai hoặc yêu cầu mua/chọn dịch vụ. Tách phạm vi không có nghĩa những rủi ro này đã được giải quyết.

## Kiểm chứng và giới hạn audit

Đọc boundary M03–M11, các reader/writer M09–M11, chain checks và tests liên quan; đối chiếu mục nghiệm thu BR-03 trong kế hoạch. Probe E-01 trên baseline như trên. Chạy Go tests/vet ba module contracts, lab/affiliate-bot, lab/mission-runtime; 8 scripts validate và 10 Python regression. Kết quả suite không thay chứng minh branch coverage.

Thứ tự đề xuất: review audit → E-01 sửa precision → E-02 bổ sung ma trận/test → E-03 review đóng BR-03 có giới hạn. Không tự sửa code trong audit hoặc mở rộng authority của bot.
