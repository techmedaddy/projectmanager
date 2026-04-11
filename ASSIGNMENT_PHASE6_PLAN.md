# ASSIGNMENT PHASE 6 PLAN — Acceptance Checklist

Status legend:
- [ ] Not done
- [x] Done

## Scope targets (requested)

- [ ] Project visibility rule fix
- [ ] Zero-step seed flow
- [ ] Assignee UX improvement
- [ ] `/auth/me` client consistency
- [ ] `GET /projects/:id/stats`
- [ ] Pagination support
- [ ] QA + docs updates

---

## 1) Project visibility rule fix

### Acceptance criteria
- [ ] `GET /projects` includes projects where user is:
  - [ ] owner, OR
  - [ ] assignee of at least one task, OR
  - [ ] creator of at least one task
- [ ] Project detail access rule is consistent with list visibility policy.
- [ ] Integration tests added/updated for owner, assignee, creator, and forbidden cases.
- [ ] README/API docs reflect final visibility behavior.

### Evidence
- [ ] Query/service change committed
- [ ] Tests passing and recorded

---

## 2) Zero-step seed flow

### Acceptance criteria
- [ ] Fresh start works with one command:
  - [ ] `docker compose down -v`
  - [ ] `docker compose up --build`
- [ ] Test credentials work immediately after startup (without manual seed command).
- [ ] Seeding is idempotent (safe on repeated starts).
- [ ] Seed behavior is configurable (env toggle documented).
- [ ] Manual seed fallback remains documented.

### Evidence
- [ ] Startup logs show migration + seed behavior
- [ ] Login proof with seeded account

---

## 3) Assignee UX improvement

### Acceptance criteria
- [ ] Task modal uses human-friendly assignee selection (not raw UUID entry only).
- [ ] User can set assignee and clear assignee.
- [ ] Payload still sends `assignee_id` correctly:
  - [ ] UUID when selected
  - [ ] `null` when cleared
- [ ] Filter by assignee remains functional.
- [ ] Validation/error states remain visible and clear.

### Evidence
- [ ] UI screenshots or runtime test proof
- [ ] API request samples for set/clear assignee

---

## 4) `/auth/me` client consistency

### Acceptance criteria
- [ ] Login flow uses shared API client abstraction consistently.
- [ ] Error handling is unified (`ApiError` path) for `/auth/me` and other endpoints.
- [ ] No regression in auth persistence and protected-route redirects.

### Evidence
- [ ] Login success/failure smoke checks

---

## 5) `GET /projects/:id/stats`

### Acceptance criteria
- [ ] Endpoint implemented and routed.
- [ ] Authorization matches project access policy.
- [ ] Response includes:
  - [ ] counts by status
  - [ ] counts by assignee
- [ ] Error handling follows existing API conventions.
- [ ] README/API reference includes request/response examples.

### Evidence
- [ ] Endpoint test(s) added or updated
- [ ] Manual curl/sample response captured

---

## 6) Pagination support

### Acceptance criteria
- [ ] `GET /projects` supports `page` and `limit`.
- [ ] `GET /projects/:id/tasks` supports `page` and `limit` with filters.
- [ ] Defaults and max-limit guard are enforced.
- [ ] Response includes pagination metadata (documented contract).
- [ ] Frontend supports basic pagination controls without breaking filters.

### Evidence
- [ ] API tests for bounds/defaults
- [ ] Frontend smoke proof for next/prev with filters

---

## 7) QA + docs updates

### Acceptance criteria
- [ ] Backend tests pass.
- [ ] Frontend lint/build pass.
- [ ] Runtime smoke checks pass for core flows.
- [ ] README sections are up to date with final behavior.
- [ ] API reference includes new/changed endpoints and params.
- [ ] Final checklist maps each gap -> file change -> proof.

### Evidence
- [ ] Commands + outputs recorded in final report
- [ ] Final completion summary committed

---

## Execution tracking

- [ ] Phase 1 completed
- [ ] Phase 2 completed
- [ ] Phase 3 completed
- [ ] Phase 4 completed
- [ ] Phase 5 completed
- [ ] Phase 6 completed
- [ ] Phase 7 completed

## Notes
- Keep commits small and phase-scoped.
- Prefer assignment compliance over feature creep.
- Preserve existing behavior unless it conflicts with acceptance criteria.
