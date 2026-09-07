# BR-16b — walkthrough và pilot người mới

## Đường walkthrough

Người học bắt đầu từ `curriculum/BOOT/QUICKSTART.md`, sau đó chạy lần lượt
fixture smoke M00→M06. Mỗi chặng phải đọc input/output và tự giải thích một
failure case; không coi output synthetic là affiliate truth.

| Chặng | Người học tự làm | Fixture/automation đã có | Cần ghi nhận |
|---|---|---|---|
| Setup | clone, chọn working directory, chạy quickstart | smoke_quickstart | lỗi môi trường |
| History | import observation, list và replay | BR-13b smoke | record ID, retry/conflict |
| Action/outcome | validate link, thử pending/0 | BR-10a/b/c smoke | EffectRef, orphan ID |
| Advisor/review | đọc abstain và evidence limits | BR-11/12 smoke | unknowns, human review |
| Watcher | chạy fixture fetch/import, đổi nội dung | BR-13b/c smoke | NEW/duplicate, provenance |

## Mẫu bản ghi pilot

```text
pilot_id: <ẩn danh>
started_at_utc: <RFC3339>
completed_at_utc: <RFC3339>
repo_commit: <sha>
completed_steps: [..]
minutes_total: <number>
questions: [..]
assistance_given: [..]
failure_explanations: [..]
residual_gaps: [..]
learner_can_explain_input_output_limit: true|false
result: PASS | NEEDS_HELP | INCOMPLETE
```

Một pilot chỉ PASS khi người học tự hoàn thành đường walkthrough, giải thích
được input/output, một failure case và giới hạn authority. Hỗ trợ thiết kế/code
phải ghi rõ, không tính là self-service PASS. Chưa có pilot thật thì giữ
`INCOMPLETE`.
