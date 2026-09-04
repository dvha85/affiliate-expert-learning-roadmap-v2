# Deployment 24/7 — VPS và always-on runtime (runtime luôn hoạt động)

**Status:** implementation/deployment guidance (hướng dẫn triển khai), không phải curriculum authority (thẩm quyền chương trình học).

## 1. Nguyên tắc chính

```text
24/7 availability (khả dụng 24/7)
!=
24/7 authority (quyền hành động 24/7)
```

Bot có thể luôn hoạt động để quan sát, thu thập, đánh giá và cảnh báo nhưng quyền thực thi vẫn bị giới hạn bởi Mission, policy (chính sách), approval (phê duyệt), audit (kiểm toán) và kill switch (công tắc dừng).

VPS/server là **deployment concern (vấn đề triển khai)**, không phải learning prerequisite (điều kiện tiên quyết để học). Không mua/cấu hình VPS chỉ để bắt đầu BOOT, M00 hoặc các Mission chưa cần always-on runtime.

## 2. Khi nào cần VPS?

| Giai đoạn | Runtime mặc định | VPS/server |
|---|---|---|
| BOOT → M05 | local-first (ưu tiên chạy trên máy cá nhân) | chưa cần mặc định |
| M06 | automatic read-only watcher (bộ theo dõi tự động chỉ đọc) | bắt đầu có ý nghĩa nếu cần chạy liên tục khi máy cá nhân tắt |
| M07–M08 | read-only Agent + shadow ActionIntent (Agent chỉ đọc + ý định hành động chạy bóng) | phù hợp cho runtime luôn hoạt động, vẫn không tự mở quyền ghi |
| M09 | approval-gated executor (bộ thực thi qua cổng phê duyệt) | nên có môi trường server ổn định nếu chạy thật |
| M10–M11 | governed canary + production closed loop (canary có quản trị + vòng kín production) | cần production-like runtime (môi trường gần/đúng production), health/recovery/audit rõ |

VPS không phải lựa chọn duy nhất. Managed hosting/PaaS/cloud service (dịch vụ hosting/PaaS/cloud được quản lý) có thể thay thế nếu đáp ứng cùng boundary (ranh giới), durability (độ bền), security (bảo mật), audit và cost control (kiểm soát chi phí).

## 3. AI có cần cài trên VPS không?

Không mặc định.

Nếu dùng OpenAI/Anthropic/Gemini hoặc AI API tương đương:

```text
VPS/server
→ chạy Go Bot + orchestration + state + policy + workers

External AI API
→ chạy model inference (suy luận mô hình)
```

Không self-host LLM (tự lưu trữ mô hình ngôn ngữ lớn) chỉ vì Bot chạy 24/7. Self-host model chỉ được cân nhắc khi có bottleneck/use case (điểm nghẽn/trường hợp sử dụng) đo được về privacy (riêng tư), latency (độ trễ), cost (chi phí), availability (khả dụng) hoặc model control (kiểm soát mô hình), và phải tính riêng CPU/RAM/GPU, model lifecycle (vòng đời mô hình) và security burden (gánh nặng bảo mật).

## 4. Baseline triển khai đơn giản

Một baseline (mốc cơ sở) phù hợp cho giai đoạn đầu của always-on runtime:

```text
VPS hoặc server nhỏ
├── Go Bot / worker
├── n8n
├── PostgreSQL hoặc canonical state store (kho trạng thái chuẩn)
├── reverse proxy + TLS
├── health check + restart policy
├── logs/telemetry
└── backup

External
├── AI API
├── Affiliate/public APIs
├── notification channel (kênh thông báo)
└── off-host backup/object storage khi cần
```

Docker Compose có thể là deployment baseline (mốc triển khai) đơn giản trước khi có bằng chứng rằng Kubernetes hoặc orchestration phức tạp hơn là cần thiết.

Một VPS khoảng `2 vCPU / 2–4 GB RAM` có thể dùng làm **starting heuristic (ước lượng khởi đầu)** cho tải nhỏ của Go Bot + n8n + database, nhưng **không phải requirement (yêu cầu)**. Chọn cấu hình bằng đo CPU/RAM, queue depth (độ sâu hàng đợi), latency, database load (tải cơ sở dữ liệu) và headroom (dư địa tài nguyên) của workload thật.

## 5. Không phải mọi thành phần đều chạy liên tục

`24/7 Bot` không có nghĩa một Agent luôn thức và gọi model liên tục.

Ví dụ:

```text
watcher (bộ theo dõi)         → mỗi 5–30 phút khi use case cần
analytics/evaluation          → theo giờ/ngày hoặc theo event
AI/Agent                      → chỉ gọi khi trigger đủ điều kiện
executor                      → chỉ chạy khi có authorization phù hợp
human review                  → exception/high-risk (ngoại lệ/rủi ro cao)
```

Cadence (nhịp chạy) phải được chọn theo tốc độ dữ liệu thay đổi, latency requirement (yêu cầu độ trễ), API limit (giới hạn API), chi phí và rủi ro; không dùng polling nhanh chỉ để gọi là “24/7”.

## 6. Canonical ownership (quyền sở hữu trạng thái chuẩn)

```text
n8n workflow state (trạng thái workflow n8n)
!=
canonical business state (trạng thái nghiệp vụ chuẩn)
```

n8n có thể orchestrate (điều phối), nhưng evidence/history/decision/action/outcome/audit quan trọng phải tuân theo canonical contracts (hợp đồng dữ liệu chuẩn) của repo. Không để một workflow engine trở thành nguồn sự thật duy nhất chỉ vì nó đang chạy trên VPS.

## 7. Minimum production hygiene (vệ sinh production tối thiểu)

Khi chuyển từ local sang always-on server, tối thiểu phải có:

- secret/environment handling (quản lý bí mật/biến môi trường); không commit credential;
- TLS và network exposure (phạm vi mở mạng) tối thiểu;
- backup + restore test (sao lưu + kiểm thử khôi phục) cho canonical state;
- health check (kiểm tra sức khỏe) và restart policy;
- resource limit (giới hạn tài nguyên) và disk monitoring (theo dõi ổ đĩa);
- correlation/audit (liên kết/kiểm toán) cho run quan trọng;
- cost visibility/alert (hiển thị/cảnh báo chi phí), đặc biệt với AI/API;
- fail-safe/fail-closed (an toàn khi lỗi/đóng khi lỗi) ở boundary có hậu quả;
- kill switch và recovery path (đường phục hồi) phù hợp authority của Mission.

Production deployment không tự tạo Reality/Operated PASS. Mission vẫn phải chứng minh đúng evidence, linkage (liên kết), authority ceiling (trần quyền hạn) và failure cases (ca lỗi) mà contract yêu cầu.

## 8. Kiến trúc đích

```text
Scheduler / Watcher 24/7
        ↓
Read-only Data Collection (thu thập chỉ đọc)
        ↓
Evidence / Canonical State
        ↓
Deterministic Decision Engine (bộ quyết định tất định)
        ↓
AI / Agent reasoning khi cần
        ↓
Deterministic Policy + Risk Gate
        ↓
Approval hoặc Delegated Authority (ủy quyền giới hạn)
        ↓
Controlled Executor (bộ thực thi có kiểm soát)
        ↓
Outcome Tracking (theo dõi kết quả)
        ↓
Evaluation (đánh giá)
        ↓
Reviewed Improvement (cải tiến đã review)
        ↺
```

Mục tiêu cuối không phải “Agent tự do chạy 24/7”, mà là **hệ thống luôn hoạt động với authority bị giới hạn, quan sát được, có thể dừng và phục hồi được**.
