# BR-16a — offline end-to-end continuity smoke

Smoke này chạy lại các bài smoke đã nghiệm thu của learner Bot theo một chuỗi
liên tục: quickstart, action/outcome, evaluation/review và watcher history. Nó
kiểm tra fixture, retry/replay, wrong ID/sink failure ở các lát cắt tương ứng;
không gọi affiliate program, không tạo external side effect và không thay thế
pilot người mới.

```text
python scripts/smoke_br16a_offline.py
```

Kết quả `PASS` chỉ là lab/fixture continuity evidence. BR-16 vẫn cần walkthrough
với người mới, ghi thời gian/câu hỏi/trợ giúp và residual gaps trước khi coi là
beginner MVP đạt.
