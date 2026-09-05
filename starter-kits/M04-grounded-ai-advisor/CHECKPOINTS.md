# M04 Checkpoints

- [ ] Advisor dùng evidence IDs resolve được.
- [ ] JSON gốc được kiểm trước khi unmarshal mất thông tin: đủ required fields, không field lạ/trùng; arrays không null, IDs không rỗng/trùng, reason không rỗng/khoảng trắng.
- [ ] Đã chạy `advisor-check` với file riêng; reason rỗng bị `INVALID_SCHEMA`, write request bị reject; không dùng exit 0 thay validation result.
- [ ] stale và future evidence gây abstain.
- [ ] hallucinated evidence bị reject.
- [ ] `write_tool_requested=false` có mặt trong output contract.
- [ ] model/provider/version được ghi nếu dùng LLM thật.
- [ ] model success không được coi là business truth/permission.
- [ ] Reality + Operated evidence đã lưu.
