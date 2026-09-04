# Tiến độ — trạng thái người học

Repo v2 bắt đầu theo trục Mission mới. File này chỉ ghi **tiến độ học thực tế trong repo v2**; kết quả từ repo lịch sử được tách riêng để tránh nhầm với bài đã học trong chương trình mới.

## Trạng thái học v2 hiện tại

- Trạng thái: **IN PROGRESS (đang học)**
- Mission hiện tại: **M00 — First Real Evidence Packet (gói bằng chứng thật đầu tiên)**
- Bài đã hoàn thành trong repo v2: **BOOT.0 — PASS; BOOT.1 — PASS; O00.1 — ORIENTATION COMPLETED (hoàn thành định hướng); M00.1 — PASS**
- Bài tiếp theo: **M00.2 — Evidence, uncertainty và missing data (bằng chứng, bất định và dữ liệu còn thiếu)**
- Tài liệu tham khảo BOOT: **`curriculum/BOOT/BOOT-REFERENCE.md`**
- Trạng thái Reality (thực tế) của M00: **chưa PASS — chưa có đủ real public observations (quan sát công khai thật) theo yêu cầu M00**

## Evidence học tập cho BOOT

### BOOT.0 — PASS

Learner đã:

- xác định được local environment (môi trường local);
- chạy được Git;
- chạy được Go `1.27.0` trên `darwin/arm64`;
- chạy `go test ./...` thành công trong `lab/affiliate-bot`;
- giải thích được vì sao API key/token/credential là secret (bí mật) và không được commit vào Git repo.

BOOT.0 PASS chỉ chứng minh environment/tooling basics (nền tảng môi trường/công cụ), không tạo E1 evidence hoặc Mission PASS.

### BOOT.1 — PASS

Learner đã:

- chạy baseline bằng `go run ./cmd/bot` và quan sát deterministic runtime output (output runtime tất định);
- phân biệt được output ranking theo công thức hiện tại với business truth (sự thật kinh doanh);
- tạo intentional regression (hồi quy lỗi có chủ đích) ở tie-break (quy tắc phá hòa);
- quan sát `TestSameInputSameOutput` và eval case phát hiện `[b a]` thay vì `[a b]`;
- phân loại đúng đây là test assertion failure (lỗi assertion của test), không phải environment hay compile failure;
- sửa tối thiểu để đưa behavior về đúng expectation;
- giải thích được `go run` chứng minh Bot chạy/runtime behavior hiện tại, còn `go test ./...` PASS chỉ chứng minh code thỏa các expectation được test hiện tại kiểm tra; không chứng minh sản phẩm đứng #1 là sản phẩm Affiliate tốt nhất.

BOOT.1 PASS chỉ xác nhận tooling/test semantics capability (năng lực công cụ/ngữ nghĩa kiểm thử); không tạo E1, M00 PASS hoặc quyền automation.

## O00.1 — ORIENTATION COMPLETED

Learner đã chạy:

```bash
cd lab/mission-runtime
go run ./cmd/demo O00
```

và quan sát được chuỗi artifact (đối tượng dữ liệu):

```text
Observation
→ DecisionPacket
→ AdvisorOutput
→ ActionIntent
→ PolicyDecision
→ ApprovalReview
→ ExecutionRecord
→ OutcomeRecord
→ EvaluationRecord
```

Learner đã giải thích được các boundary (ranh giới) chính:

- `Observation (quan sát)` là dữ liệu/sự kiện được ghi nhận, **không phải recommendation (khuyến nghị)**;
- `DecisionPacket (gói quyết định)` là quyết định/kết luận dựa trên evidence (bằng chứng);
- `ActionIntent = PROPOSAL_ONLY (chỉ là đề xuất)` nghĩa là Bot **chưa có execution permission (quyền thực thi)**;
- `ExecutionRecord = DRY_RUN_ONLY (chỉ chạy mô phỏng)` nghĩa là execution (thực thi) chỉ được mô phỏng và không tạo hành động thật; dry-run/live execution là trục khác với synthetic/real evidence;
- `external_side_effects=false (không có tác động bên ngoài)` nghĩa là runtime (chương trình khi chạy) **đã thực sự chạy**, nhưng không publish (đăng), send (gửi), spend (chi tiền), mutate account (sửa tài khoản) hoặc gọi write API (API ghi) ra hệ thống bên ngoài.

O00.1 chỉ hoàn thành orientation (định hướng). Nó dùng synthetic data (dữ liệu mô phỏng), không tạo E1 evidence, Reality PASS, Mission PASS hoặc quyền automation.

## M00.1 — PASS

### Context (bối cảnh) xuyên suốt M00 đã chốt

```text
Audience (nhóm mục tiêu):
Freelancer / remote worker tại Việt Nam.

Problem (vấn đề):
Làm việc lâu bằng laptop/MacBook trên bàn nhỏ; cần một giải pháp gọn,
rẻ và phù hợp để nâng laptop khi làm việc.

Offer/category (sản phẩm/danh mục):
Giá đỡ laptop nhôm gấp gọn, điều chỉnh độ cao, phân khúc giá rẻ.

Market Question (câu hỏi thị trường):
Trong các giá đỡ laptop nhôm gấp gọn giá rẻ có thể quan sát công khai,
offer nào đáng để tiếp tục nghiên cứu như một cơ hội Affiliate cho
freelancer/remote worker Việt Nam?
```

Decision (quyết định) ở M00 chỉ hỗ trợ các trạng thái kiểu:

```text
CONTINUE_RESEARCH (tiếp tục nghiên cứu)
GET_MORE_DATA (cần thêm dữ liệu)
```

không phải:

```text
BUY
PUBLISH
RECOMMEND
EXECUTE
```

### Learner notes (ghi chú người học)

Learner đã hiểu và giải thích được:

1. `price × commission_rate (giá × tỷ lệ hoa hồng)` chỉ là **weak scenario (kịch bản yếu)** vì chưa phản ánh đầy đủ demand (nhu cầu), conversion rate (tỷ lệ chuyển đổi), valid-order rate (tỷ lệ đơn hợp lệ), chi phí, rủi ro, competition (cạnh tranh), product–audience fit (độ phù hợp sản phẩm–nhóm mục tiêu) và các yếu tố quyết định khác.
2. `order (đơn hàng) != valid order (đơn hợp lệ) != payment (tiền thực nhận)` vì order có thể bị hủy, hoàn, không đủ điều kiện hoặc không được attribution (ghi nhận nguồn Affiliate); valid order là phần đơn được xác nhận hợp lệ; payment còn phụ thuộc trạng thái hoa hồng, điều chỉnh và lịch/điều kiện thanh toán của nền tảng.
3. Một offer có số bán hoặc commission (hoa hồng) cao nhất vẫn chưa đủ để gọi là **best affiliate opportunity (cơ hội Affiliate tốt nhất)** nếu nó không phù hợp với audience (nhóm mục tiêu) hoặc có conversion, rủi ro, competition, seller quality (chất lượng người bán) hay expected value (giá trị kỳ vọng) kém hơn.

Mental model (mô hình tư duy) đã chốt:

```text
highest commission rate (tỷ lệ hoa hồng cao nhất)
!= best affiliate opportunity (cơ hội Affiliate tốt nhất)

local metric (chỉ số cục bộ)
!= final business outcome (kết quả kinh doanh cuối)

DATA/EVIDENCE (dữ liệu/bằng chứng)
> OPINION (ý kiến)

EXPECTED VALUE (giá trị kỳ vọng)
> COMMISSION RATE (tỷ lệ hoa hồng) khi đánh giá cơ hội tổng thể
```

M00.1 PASS là knowledge/decision framing PASS (PASS về hiểu biết và định khung quyết định). Nó **chưa tạo Reality PASS cho M00**, vì real public observations (quan sát công khai thật) và provenance (nguồn gốc/truy vết bằng chứng) sẽ được làm tiếp ở M00.2.

## Kết quả từ repo lịch sử

- `BOOT.1 — Chạy, sửa và kiểm thử Bot`: learner từng có **historical PASS credit (kết quả PASS lịch sử có thể được công nhận)**, nhưng hiện learner đã **thực sự học lại và PASS BOOT.1 trong repo v2**, nên không cần dùng historical credit cho trạng thái hiện tại.
- Historical credit không tạo E1 evidence (bằng chứng E1), M00 Reality PASS hoặc quyền automation (tự động hóa).

## Đường học hiện tại

```text
BOOT.0 PASS
→ BOOT.1 PASS
→ O00.1 ORIENTATION COMPLETED
→ M00.1 PASS
→ M00.2
→ M00.3
→ M00 PASS
→ M01 ...
```

`BOOT.0` và `BOOT.1` là onboarding/tooling (làm quen môi trường/công cụ), không phải Mission PASS gate. `O00.1` là orientation (định hướng) bằng dữ liệu synthetic (mô phỏng), không tạo Mission PASS. Mission đầu tiên tạo Reality evidence (bằng chứng thực tế) là M00.

## Quy tắc cập nhật tiến độ

- Chỉ đánh dấu một bài trong repo v2 là đã hoàn thành khi learner thực sự học/thực hành bài đó hoặc chủ động áp dụng historical credit khi curriculum cho phép.
- Progress (tiến độ) của Mission chỉ được cập nhật khi có evidence (bằng chứng) tương ứng.
- CI xanh, fixture PASS, sample/synthetic data (dữ liệu mẫu/mô phỏng) hoặc trạng thái authoring `ready` không tự tạo learner PASS hay Reality PASS.
