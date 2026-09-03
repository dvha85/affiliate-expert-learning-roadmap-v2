# Chính sách ngôn ngữ — tiếng Việt là ngôn ngữ chính cho người học

## Quy tắc

Nội dung hướng dẫn người học phải dùng **tiếng Việt làm ngôn ngữ chính**.

Thuật ngữ kỹ thuật tiếng Anh chỉ giữ nguyên khi thuộc một trong các nhóm sau:

- `code identifier` (định danh trong mã nguồn);
- `contract/schema field` (trường trong contract/schema);
- tên product/framework/tool;
- enum/state/key cần khớp chính xác với code;
- thuật ngữ ngành mà dịch hoàn toàn làm giảm độ chính xác.

Ở lần xuất hiện đầu trong phần giải thích cho người học, thuật ngữ tiếng Anh nên có nghĩa tiếng Việt đi kèm khi cần, ví dụ:

```text
Evidence (bằng chứng)
DecisionPacket (gói quyết định có cấu trúc)
ActionIntent (ý định hành động có cấu trúc)
Human Review (người kiểm tra/phê duyệt thủ công)
Replay (phát lại quyết định từ dữ liệu và phiên bản cũ)
```

Không dịch các identifier như `RANK_SCENARIO`, `GET_MORE_DATA`, `HUMAN_REVIEW`, `DecisionPacket`, `ActionIntent`, JSON keys hoặc code symbols nếu việc dịch làm mất liên kết với code/contract. Thay vào đó, giải thích nghĩa tiếng Việt ở prose xung quanh.

## Không chấp nhận

Trong learner-facing Markdown, không để các đoạn prose, heading, bảng mô tả hoặc flow giải thích hoàn toàn bằng tiếng Anh nếu chúng có thể diễn đạt rõ bằng tiếng Việt.

Ví dụ không nên viết:

```text
First Tracked Human Action
Reliable Automatic Watcher
Grounded AI Advisor
```

Nên viết:

```text
First Tracked Human Action (hành động thật đầu tiên do người thực hiện và được theo dõi)
Reliable Automatic Watcher (bộ theo dõi tự động chỉ đọc, có độ tin cậy)
Grounded AI Advisor (AI tư vấn dựa trên bằng chứng)
```

Tên Mission/Artifact tiếng Anh có thể được giữ để ổn định tham chiếu, nhưng phải đi kèm diễn giải tiếng Việt trong nội dung người học nhìn thấy.

## Phạm vi learner-facing

Chính sách này áp dụng tối thiểu cho:

- `README.md`, `CURRICULUM.md`, `ROADMAP.md`, `PROGRESS.md`;
- `curriculum/**/*.md`;
- `missions/**/*.md`;
- `starter-kits/**/*.md`;
- `templates/**/*.md` và evidence template dành cho learner.

Code, JSON schema, enum, command, file path và machine-readable metadata không bắt buộc dịch, nhưng prose giải thích chúng phải ưu tiên tiếng Việt.

## Mục tiêu

Người học phải hiểu được nghĩa, boundary (ranh giới) và mục đích mà không bị buộc phải đọc một đoạn tiếng Anh thuần túy chỉ để theo curriculum.
