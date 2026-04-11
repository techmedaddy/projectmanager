# Final Done Matrix — Suggestion → Commit/File/Test Proof

Date: 2026-04-11 (IST)
Branch: `master`

This matrix maps each remediation suggestion to implementation commits, concrete files, and validation proof artifacts.

---

## 1) Visibility rule alignment (owner OR involved in task)

### Suggestion
- Align project visibility and project/task access checks so users can access projects when they are owner OR task creator OR task assignee.

### Commits
- `e86af19` docs(backend): clarify project visibility intent in comments
- `b3de623` fix(projects): include task creators in accessible projects query
- `f2ae29c` fix(authz): align project/task access with creator-or-assignee involvement
- `c627433` test(api): add project visibility integration coverage

### Files
- `backend/internal/projects/repository.go`
- `backend/internal/projects/service.go`
- `backend/internal/tasks/service.go`
- `backend/cmd/api/integration_test.go`

### Test/Proof
- Integration tests include:
  - assignee visibility
  - creator-only visibility
  - outsider denial
- Test pass run captured in Phase 7.1 execution output.

---

## 2) Zero-step startup seed flow

### Suggestion
- Fresh `docker compose up --build` should provide immediate test login without manual seed commands.

### Commits
- `a105574` feat(config): add AUTO_SEED toggle with default true
- `d8b2752` feat(seed): run idempotent auto-seed after migrations
- `49e7f3e` docs: document AUTO_SEED compose integration and production default

### Files
- `backend/internal/config/config.go`
- `backend/internal/db/seed.go`
- `backend/seed/embed.go`
- `backend/cmd/api/main.go`
- `.env.example`
- `backend/.env.example`
- `docker-compose.yml`
- `README.md`

### Test/Proof
- Fresh stack login validated earlier with seeded credentials.
- Seed behavior documented in README + env toggles.

---

## 3) Assignee UX improvement (dropdown instead of raw UUID)

### Suggestion
- Improve task modal assignee UX with selectable options and clear-unassign behavior.

### Commits
- `154ae38` feat(api): add project assignees endpoint for task assignee options
- `3a4e17b` feat(frontend): switch task assignee input to project assignee dropdown
- `26c0000` feat(frontend): show no-assignee-options hint while preserving unassigned flow
- `3390af1` feat(frontend): improve assignee inline validation and API field error mapping

### Files
- `backend/cmd/api/users.go`
- `backend/cmd/api/router.go`
- `backend/internal/users/{repository.go,service.go,dto.go}`
- `frontend/src/components/tasks/TaskModal.tsx`
- `frontend/src/types/index.ts`

### Test/Proof
- Runtime interaction checks during final review show assignee controls visible and working.
- README API + behavior updated for `GET /projects/:id/assignees` and dropdown flow.

---

## 4) `/auth/me` client consistency

### Suggestion
- Use shared API client consistently for auth hydration and unify error handling.

### Commits
- `1d7240f` refactor(frontend): use fetchApi for login me-profile retrieval
- `a7b0063` refactor(frontend): centralize auth/me failure handling with ApiError path

### Files
- `frontend/src/pages/Login.tsx`
- `frontend/src/contexts/AuthContext.tsx`

### Test/Proof
- Wrong creds / valid creds / invalid-token behavior previously verified in phase logs.

---

## 5) Bonus stats endpoint (`GET /projects/:id/stats`)

### Suggestion
- Add project-level stats endpoint (counts by status + assignee) and surface in UI.

### Commits
- `dd19dcd` feat(api): add project stats endpoint with status and assignee aggregates
- `b13a296` feat(frontend): add project stats widget for status totals and top assignees

### Files
- `backend/internal/projects/{repository.go,service.go,dto.go}`
- `backend/cmd/api/{projects.go,router.go}`
- `frontend/src/pages/ProjectDetail.tsx`
- `frontend/src/types/index.ts`

### Test/Proof
- Final review confirms stats section visible on project detail (375 + 1280).

---

## 6) Pagination backend contract + frontend integration

### Suggestion
- Add page/limit support and metadata; wire minimal next/prev controls; preserve filters while paginating tasks.

### Commits
- `357bd06` feat(api): add page/limit pagination with meta on projects and tasks lists
- `6b706b1` feat(api): move pagination to SQL with stable ordering and total counts
- `3abf369` feat(api): standardize paginated list responses to items+meta
- `ab1766f` feat(frontend): add minimal pagination controls for projects and tasks
- `d832bf8` docs: add phase 6.5 pagination validation matrix results

### Files
- Backend:
  - `backend/cmd/api/{pagination.go,projects.go,tasks.go}`
  - `backend/internal/projects/{repository.go,dto.go}`
  - `backend/internal/tasks/{repository.go,dto.go}`
- Frontend:
  - `frontend/src/pages/{Projects.tsx,ProjectDetail.tsx}`
  - `frontend/src/components/tasks/TaskBoard.tsx`
  - `frontend/src/types/index.ts`

### Test/Proof
- `PHASE6_5_VALIDATION.md` — 15/15 API pagination/filter boundary cases passed.
- `PHASE7_4_FINAL_REVIEW.md` — next/prev + filter/pagination UX verified at 375 and 1280.

---

## 7) QA, docs, and submission hardening

### Suggestion
- Complete QA pass, API docs refresh, final closure evidence.

### Commits
- `6c42492` test(backend): fix integration migration glob patterns
- `d9f2c04` docs(readme): align startup, visibility, assignee ux, stats and pagination
- `0098545` docs(api): expand reference with pagination contract and query examples
- `d539fd3` docs: add phase 7.4 final review validation report

### Files
- `backend/cmd/api/integration_test.go`
- `README.md`
- `PHASE6_5_VALIDATION.md`
- `PHASE7_4_FINAL_REVIEW.md`
- `FINAL_DONE_MATRIX.md` (this file)

### Test/Proof
- Backend integration tests pass after migration glob fix.
- Frontend `npm run lint && npm run build` pass.
- Final responsive + console-error pass documented.

---

## Final status

All requested suggestions are implemented and mapped to commit/file/test proof.
