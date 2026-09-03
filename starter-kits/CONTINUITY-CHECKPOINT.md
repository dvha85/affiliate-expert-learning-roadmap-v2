# Continuity Checkpoint — M03→M11

Checklist này áp dụng **bổ sung** cho checkpoint riêng của từng Mission từ M03 tới M11.

- [ ] Capability mới được gắn vào cùng learner Bot/workspace đang tiến hóa từ M01/M02, không chỉ chạy demo rời.
- [ ] Ghi `learner_bot_commit_or_version` của implementation đã vận hành.
- [ ] `previous_mission_artifact_refs` resolve được tới artifact thật từ Mission trước.
- [ ] Ghi `new_capability_entrypoint` và component chịu trách nhiệm.
- [ ] Ghi canonical contracts được dùng; adapter/vendor không tự tạo ontology thay thế.
- [ ] Chạy conformance/eval failure cases tương ứng và ghi kết quả.
- [ ] Phân biệt rõ canonical state với cache/orchestration/telemetry state.
- [ ] Reality/Operated evidence dùng cùng integration path, không thay bằng fixture/demo.
- [ ] Authority ceiling của Mission được test fail-closed.
- [ ] Known gap/limitation và rollback/recovery path được ghi rõ.

Nếu checklist này chưa đạt, Mission có thể PASS Capability nhưng **chưa đủ continuity cho Operated PASS**.

Chi tiết: `docs/architecture/LEARNER-BOT-CONTINUITY.md`.
