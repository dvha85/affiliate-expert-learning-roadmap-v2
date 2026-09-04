# Local AI Runtime — LM Studio và model local

**Last reviewed:** 2026-09-04  
**Status:** implementation/development guidance (hướng dẫn triển khai/phát triển), không phải curriculum authority (thẩm quyền chương trình học).

Tài liệu này ghi lại baseline (mốc cơ sở) cho việc dùng AI local trong Affiliate Intelligence Bot. Mục tiêu là tận dụng model local cho reasoning (suy luận), research support (hỗ trợ nghiên cứu), coding và proposal (đề xuất) mà không làm mờ deterministic boundary (ranh giới tất định), policy (chính sách) hoặc authority ceiling (trần quyền hạn) của Mission.

## 1. Kết luận ngắn

Với máy phát triển thuộc profile (cấu hình) **MacBook Air M2, 24 GB unified memory (bộ nhớ hợp nhất)**:

- dùng **LM Studio** làm local inference server (máy chủ suy luận local) là phù hợp cho học, development (phát triển) và prototype (nguyên mẫu);
- **Qwen3.8-27B MLX 4-bit** là candidate (ứng viên) hợp lý cho model reasoning chính ở local;
- bắt đầu với **8K context (ngữ cảnh)**, tăng lên **16K** khi cần; chỉ thử **32K** khi có use case và theo dõi memory pressure (áp lực bộ nhớ)/swap;
- không coi 64K–262K là cấu hình mặc định trên máy 24 GB chỉ vì model hỗ trợ context dài;
- `reasoning effort = medium` là baseline; `low` cho tác vụ đơn giản, `xhigh` chỉ khi tác vụ khó chứng minh được lợi ích;
- không dùng MacBook Air làm dedicated 24/7 inference server (máy chủ suy luận chuyên dụng 24/7) cho model 27B lâu dài.

Đây là **learner/development profile (cấu hình cho người học/phát triển)**, không phải production requirement (yêu cầu production).

## 2. Vai trò đúng của AI local

```text
Go
├── deterministic domain logic (logic nghiệp vụ tất định)
├── schema / validation (lược đồ / kiểm tra hợp lệ)
├── policy / authority checks (kiểm tra chính sách / quyền hạn)
├── canonical state (trạng thái chuẩn)
└── controlled execution (thực thi có kiểm soát)

Qwen local
├── reasoning (suy luận)
├── classification (phân loại)
├── summarization (tóm tắt)
├── evidence interpretation (diễn giải bằng chứng)
├── research synthesis (tổng hợp nghiên cứu)
├── content drafting (soạn nội dung)
└── proposal generation (tạo đề xuất)

n8n
├── orchestration (điều phối)
├── scheduling (lập lịch)
├── integrations (tích hợp)
└── notifications (thông báo)

MCP / explicit tools (công cụ tường minh)
├── web/search
├── browser
├── files
├── database
└── external APIs
```

Nguyên tắc:

```text
LLM output != truth
LLM proposal != authorization
Tool available != tool permitted
Tool call succeeded != result trusted
```

AI local không thay deterministic core (lõi tất định) và không tự mở authority (quyền hạn).

## 3. Baseline model cho MacBook Air M2 24 GB

### Primary local reasoning candidate (ứng viên suy luận local chính)

```text
Model: Qwen3.8-27B
Runtime: LM Studio
Format ưu tiên trên Apple Silicon: MLX
Quantization (lượng tử hóa): 4-bit
```

Lý do chọn profile này:

- 27B đủ mạnh để thử nghiệm reasoning, coding, tool use (gọi công cụ), vision (thị giác) và agentic workflow (luồng Agent nhiều bước);
- bản MLX phù hợp Apple Silicon;
- 4-bit tạo headroom (dư địa tài nguyên) tốt hơn 5/6/8-bit trên máy chỉ có 24 GB unified memory;
- model này là **candidate để đánh giá**, không trở thành dependency (phụ thuộc) bắt buộc của curriculum.

Không khóa kiến trúc Bot vào tên model cụ thể. Có thể thay Qwen bằng model khác nếu cùng eval (bộ đánh giá) chứng minh tốt hơn về chất lượng, tốc độ, RAM, tool calling hoặc chi phí.

## 4. Context baseline

Khuyến nghị bắt đầu:

| Context | Cách dùng |
|---|---|
| `8K` | baseline mặc định cho chat, coding và task Agent ngắn |
| `16K` | dùng thường xuyên khi cần nhiều evidence/tài liệu hơn |
| `32K` | chỉ bật cho task có lợi ích rõ và phải theo dõi memory pressure/swap |
| `64K+` | không mặc định trên máy 24 GB |
| `262K` | capability (khả năng) của model, không phải target runtime (mục tiêu chạy) của laptop này |

Quy tắc kiến trúc:

```text
Long context available
!=
put everything into context
```

Ưu tiên:

```text
Search / RAG / database query
        ↓
lọc evidence liên quan
        ↓
context nhỏ, có provenance (nguồn gốc)
        ↓
Qwen reasoning
```

Không đưa toàn bộ repo, toàn bộ history và toàn bộ dữ liệu thị trường vào mỗi prompt nếu retrieval (truy xuất) có thể lấy đúng phần cần dùng.

## 5. Reasoning baseline

Khuyến nghị:

```text
simple extraction / formatting / classification → low hoặc model nhỏ
normal analysis / coding / DecisionPacket      → medium
hard research / difficult multi-step reasoning → xhigh khi thật sự cần
```

Reasoning level cao hơn không tự đồng nghĩa với output tốt hơn. Mọi thay đổi phải được đo bằng eval set (bộ đánh giá), latency (độ trễ), resource use (mức dùng tài nguyên) và failure rate (tỷ lệ lỗi).

Không lưu chain-of-thought/reasoning nội bộ làm canonical evidence (bằng chứng chuẩn). Canonical record chỉ lưu input cần thiết, source/provenance, structured output (đầu ra có cấu trúc), decision rationale (lý do quyết định) ở mức phù hợp và audit metadata (siêu dữ liệu kiểm toán).

## 6. Internet và tool use

Model local không tự có dữ liệu Internet mới nhất chỉ vì LM Studio đang online.

Đường đúng:

```text
Qwen
  ↓ yêu cầu tool
Tool Registry / MCP / n8n
  ↓
Search / API / Browser
  ↓
raw result (kết quả thô, chưa tin cậy)
  ↓
validation + provenance
  ↓
Observation
  ↓
DecisionPacket
```

Mọi Internet access (truy cập Internet) phải qua tool boundary (ranh giới công cụ) rõ ràng với allowlist (danh sách cho phép), timeout (thời gian chờ), credential isolation (cách ly thông tin xác thực), audit và authority ceiling theo Mission.

## 7. Structured output trước execution

Không dùng pattern (mẫu):

```text
LLM nói gì
→ làm luôn
```

Dùng:

```text
LLM
→ structured output
→ schema validation
→ deterministic checks
→ policy/risk gate
→ DecisionPacket / ActionIntent
→ approval hoặc delegated authority phù hợp
→ executor
```

Trong giai đoạn hiện tại, local AI phù hợp nhất cho `Observation`, evidence interpretation và proposal. Nó không thay đổi ý nghĩa của:

```text
ActionIntent = PROPOSAL_ONLY
ExecutionRecord = DRY_RUN_ONLY
external_side_effects = false
```

## 8. Không bắt model 27B làm mọi việc

Khi workload (khối lượng công việc) tăng, có thể dùng task router (bộ định tuyến tác vụ):

```text
Task
  ↓
Router
  ├── model nhỏ / deterministic code
  │     ├── extraction
  │     ├── formatting
  │     ├── simple classification
  │     └── cheap summarization
  │
  └── Qwen3.8-27B
        ├── difficult reasoning
        ├── research synthesis
        ├── coding/review khó
        ├── evidence conflict analysis
        └── DecisionPacket proposal
```

Chỉ thêm router/multi-model khi đo được bottleneck. Không tạo complexity (độ phức tạp) chỉ để có nhiều Agent/model.

## 9. Quản lý RAM và nhiệt trên MacBook Air

MacBook Air dùng passive cooling (tản nhiệt thụ động), vì vậy sustained inference (suy luận kéo dài) có thể giảm hiệu năng khi máy nóng.

Với 24 GB unified memory:

- tránh load nhiều model lớn song song;
- không mở context rất dài theo mặc định;
- theo dõi macOS Memory Pressure và swap khi thử 16K/32K;
- khi benchmark, ghi cả tokens/s, time-to-first-token (thời gian tới token đầu), RAM/swap và nhiệt/throttling (giảm xung do nhiệt);
- đóng bớt workload nặng như nhiều browser tab/container khi benchmark model;
- model load thành công không đồng nghĩa cấu hình đó vận hành ổn định lâu dài.

## 10. Local-first bây giờ, split runtime sau này

### Giai đoạn học/phát triển

```text
MacBook Air M2 24 GB
├── Go Bot
├── n8n
├── LM Studio
│   └── Qwen3.8-27B MLX 4-bit
├── MCP / tools
└── local database nhẹ
```

Phù hợp cho:
- BOOT/Mission learning;
- development;
- DRY_RUN_ONLY;
- test/eval;
- read-only Agent prototype;
- controlled research.

### Khi cần runtime 24/7

Không mặc định giữ toàn bộ stack trên MacBook Air.

```text
VPS / always-on server
├── Go Bot / workers
├── n8n
├── canonical state/database
├── policy / audit / telemetry
└── scheduler

AI inference
├── external API
├── dedicated local inference host
└── hoặc hybrid route (định tuyến lai)
```

Chỉ self-host model 24/7 khi có bằng chứng về privacy, cost, latency, availability hoặc model control đủ để bù hardware/ops burden (gánh nặng phần cứng/vận hành).

## 11. Adoption gate cho Local AI

Local AI được coi là adopted (được chấp nhận chính thức) cho một use case khi:

```text
same canonical contracts
+ same authority ceiling
+ eval quality đạt ngưỡng
+ tool calling đủ tin cậy cho use case
+ resource use phù hợp máy/runtime
+ failure/fallback path rõ
+ provenance/audit không giảm
+ measured benefit > added complexity
```

Không adopt chỉ vì model chạy được trên laptop hoặc không mất phí API trực tiếp.

## 12. Những điều cần đo trước khi nâng cấp phần cứng

Chưa mua máy mới chỉ để chạy model lớn hơn. Trước tiên đo:

- bao nhiêu request/ngày thực sự cần 27B;
- tokens/request và context thực dùng;
- tokens/s chấp nhận được;
- RAM/swap peak (đỉnh);
- thermal throttling khi chạy dài;
- tỷ lệ task model nhỏ có thể xử lý;
- chất lượng local model so với external API trên cùng eval set;
- tổng chi phí nếu chuyển inference sang cloud/API hoặc dedicated server.

Chỉ nâng cấp khi bottleneck đã đo được ảnh hưởng tới learning velocity (tốc độ học), development velocity (tốc độ phát triển) hoặc production SLO (mục tiêu mức dịch vụ).

## 13. Official references (nguồn chính thức)

- LM Studio local server: https://lmstudio.ai/docs/developer/core/server
- LM Studio OpenAI-compatible endpoints: https://lmstudio.ai/docs/developer/openai-compat
- LM Studio tool use: https://lmstudio.ai/docs/developer/openai-compat/tools
- LM Studio app/MLX support: https://lmstudio.ai/docs/app
- LM Studio Qwen3.8-27B profile: https://lmstudio.ai/models/qwen/qwen3.8-27b
- Qwen3.8-27B official model card: https://huggingface.co/Qwen/Qwen3.8-27B

## 14. Decision snapshot

```text
NOW
MacBook Air M2 / 24 GB
→ LM Studio
→ Qwen3.8-27B MLX 4-bit
→ 8K context baseline
→ 16K when needed
→ medium reasoning baseline
→ local-first development / prototype

NOT YET
→ 64K/262K default context
→ multiple large models in parallel
→ laptop as 27B production inference server 24/7
→ LLM directly owns policy/execution
→ hardware upgrade without measured bottleneck

LATER, IF EVIDENCE JUSTIFIES
→ small-model router
→ dedicated inference host
→ hybrid local + API routing
→ measured 24/7 self-hosting
```
