review this project against case.md

11:44 AM


Worked for 8m 49s



The project implements the core flow in case.md, but I found two defects and one delivery documentation gap.
1. [P2] Rate limiting happens too early. In [consumer.go (line 48)](/Users/kaanaray/Desktop/projects/case-studies/insider_one/backend/internal/delivery/consumer.go:48), consumers acquire a token before claiming and loading jobs from Postgres. Database delays can hold already permitted sends, then release them together alongside newly permitted sends. This can exceed the required 100 requests per second per channel. Enforce pacing immediately before the provider request and test actual request timestamps under database delays.
2. [P2] Saving can silently discard newer edits. [review-dialog.vue (line 220)](/Users/kaanaray/Desktop/projects/case-studies/insider_one/frontend/src/components/campaign/review-dialog.vue:220) snapshots the form, awaits the save, then resets all fields from the response. Inputs remain editable during that request. If the marketer continues typing, those newer edits disappear when the response arrives. Disable editing while saving or preserve edits made after submission.
3. [P3] Default mock behavior needs clearer disclosure. Default startup uses a mock that returns the same Istanbul segment and 30% discount content regardless of the brief ([mockdefault.go (line 13)](/Users/kaanaray/Desktop/projects/case-studies/insider_one/backend/cmd/api/mockdefault.go:13)). Mock development is explicitly allowed, but the [README (line 5)](/Users/kaanaray/Desktop/projects/case-studies/insider_one/README.md:5) does not explain this limitation or how to enable Claude. Document both so reviewers can assess the main AI requirement.
The main requirements are present:
Requirement	Assessment
Structured LLM output and backend validation	Implemented, with bounded repair attempts
Safe segment queries	Allowlist validation and parameterized SQL
Channel consent	Checked during planning and again before sending
Personalization and character limits	Checked after rendering
Human editing and approval	Implemented; save issue above
Async delivery, retries, duplicate prevention	Implemented; uncertain outcomes stop automatic retries
Live status, metrics, health endpoints	Implemented
Seed data, mock provider, session exports	Present


Verification passed: fresh Go tests with PostgreSQL and Redis, 221 frontend tests, lint, Compose configuration validation, and the frontend build with polling enabled after a file-watcher limit error. Existing services reported healthy and returned dashboard data.

I did not verify real Claude calls, a fresh Compose installation, or the complete browser flow. No source files were changed.

   
12:21 PM







re review 







1:16 PM






Worked for 1m 30s



The three previous findings are fixed in 280a142. I found two regressions in the delivery change, both reproduced with isolated tests in a temporary copy.
1. [P1] Consent revoked during the rate-limit wait is ignored.
   [consumer.go (line 64)](/Users/kaanaray/Desktop/projects/case-studies/insider_one/backend/internal/delivery/consumer.go:64) loads the recipient and checks consent before waiting for permission to send. If consent is revoked during that wait, the cached message still goes to the provider. My reproduction revoked SMS consent during the wait; the provider received one request. Revalidate consent near dispatch while preserving request pacing.
2. [P2] Recovery can expire a live job that is waiting to send.
   Jobs enter sending before the limiter wait, but [worker.go (line 250)](/Users/kaanaray/Desktop/projects/case-studies/insider_one/backend/internal/delivery/worker.go:250) expires them after SEND_TIMEOUT + 30s, including time spent waiting. A long 429 pause or a supported low-rate/high-consumer configuration can exceed that threshold. The consumer then sends anyway, while Finish cannot update the expired row. My reproduction produced one provider request with the job permanently marked unknown. Distinguish waiting jobs from active requests, or maintain and verify ownership before dispatch.
The previous fixes check out:
- Pacing now happens after database loading, with a regression test.
- Save responses preserve edits made during the request.
- README explains canned mock responses and enabling Claude.
Fresh project tests pass: all Go packages and 222 frontend tests. Lint passes. The two additional reproduction tests fail as described above. The working tree remains unchanged.