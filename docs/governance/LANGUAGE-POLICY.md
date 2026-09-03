# Language Policy — tiếng Việt là ngôn ngữ learner-facing chính

## Rule

Nội dung hướng dẫn người học dùng tiếng Việt làm ngôn ngữ chính.

Thuật ngữ kỹ thuật tiếng Anh được giữ khi là:

- code identifier;
- contract/schema field;
- product/framework name;
- thuật ngữ ngành mà dịch hoàn toàn làm giảm độ chính xác.

Lần xuất hiện đầu trong learner-facing prose nên có diễn giải tiếng Việt khi cần, ví dụ:

```text
Evidence (bằng chứng)
DecisionPacket (gói quyết định có cấu trúc)
ActionIntent (ý định hành động có cấu trúc)
Human Review (người kiểm tra)
```

Không dịch identifier như `RANK_SCENARIO`, `GET_MORE_DATA`, `HUMAN_REVIEW`, `DecisionPacket`, `ActionIntent`, JSON keys hoặc code symbols nếu việc dịch làm mất liên kết với code/contract.

## Goal

Learner phải hiểu nghĩa và boundary, không bị buộc ghi nhớ tiếng Anh thuần túy để học curriculum.
