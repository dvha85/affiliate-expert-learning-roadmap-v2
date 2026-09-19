# Contributing

Keep changes bounded to the mission and contract being changed. Preserve the
read-only and approval boundaries documented in `CURRICULUM.md` and keep the
repository status `NOT_READY_FOR_PRODUCTION` unless the evidence and external
gates in the readiness matrix are actually satisfied.

Before opening a pull request:

1. Run the Python regression suite: `python -m unittest discover -s scripts/tests -v`.
2. Run the applicable validators and Go tests listed in the changed workflow.
3. Run `git diff --check` and do not commit generated runtime state, secrets, or
   private evidence.
4. Update the relevant plan, evidence graph, and readiness matrix when a
   readiness claim, baseline, or external dependency changes.

Every functional change should include a regression test for the existing
contract and document any local tooling limitation (for example, Go or n8n not
being installed).
