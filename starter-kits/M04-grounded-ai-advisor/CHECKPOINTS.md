# M04 Checkpoints

- Dùng [walkthrough M00–M05](../../curriculum/PRACTICE-M00-M05.md) cho mock continuity; fixture không thay provider hoặc learner evidence.
- [ ] Advisor dùng evidence IDs resolve được.
- [ ] JSON gốc được kiểm trước khi unmarshal mất thông tin: đủ required fields, không field lạ/trùng; arrays không null, IDs không rỗng/trùng, reason không rỗng/khoảng trắng.
- [ ] Đã chạy `advisor-check` với file riêng; reason rỗng bị `INVALID_SCHEMA`, write request bị reject; không dùng exit 0 thay validation result.
- [ ] stale và future evidence gây abstain.
- [ ] hallucinated evidence bị reject.
- [ ] `write_tool_requested=false` có mặt trong output contract.
- [ ] model/provider/version được ghi nếu dùng LLM thật.
- [ ] model success không được coi là business truth/permission.
- [ ] Reality + Operated evidence đã lưu.
