# Mission index — canonical execution spine

Learner order được quyết định bởi `CURRICULUM.md` và Mission spine này.

| Mission | Outcome | Authority | Status |
|---|---|---|---|
| O00 | Safe synthetic walkthrough, không PASS | no side effect | orientation |
| M00 | First Real Evidence Packet + Human DecisionPacket | human/read-only | ready |
| M01 | Smallest Deterministic Bot v0.1 | A0 deterministic | ready |
| M02 | Trustworthy History + Replay v0.2 | A0 deterministic | ready |
| M03 | First Tracked Human Action + Outcome context | human executes | planned |
| M04 | Grounded AI Advisor v0.4 | A1 advisory | planned |
| M05 | First Reviewed Improvement | A1 propose only | planned |
| M06 | Reliable Automatic Watcher | automatic read-only | planned |
| M07 | Read-only Evidence Agent | A2-RO | planned |
| M08 | Shadow ActionIntent + Policy | A3-shadow | planned |
| M09 | Durable Approval + Controlled Executor | approval-gated | planned |
| M10 | Governed Canary | bounded RISK0/RISK1 auto | planned |
| M11 | Production Closed Loop | governed production | planned |

```text
O00 → M00 → M01 → M02 → M03 → M04 → M05 → M06 → M07 → M08 → M09 → M10 → M11
```

`ready` nghĩa là lesson + starter + executable eval/runtime cần thiết đã được author và CI bảo vệ. Nó **không** có nghĩa learner đã PASS. Learner PASS phụ thuộc evidence thật và operated result theo Mission contract.

M02 `ready` hiện tại đã qua corrective two-stage gate: PASS khi còn `planned`, sau đó mới promote sang `ready` và bắt CI xác minh lại.
