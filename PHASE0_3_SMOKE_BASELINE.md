# Phase 0.3 — Smoke Baseline Report

Date: 2026-04-11 (IST)
Branch: `fix/assignment-gap-closure`

## Environment

Commands run:

```bash
docker compose down -v
docker compose up -d --build
docker compose ps
docker compose logs --tail=200 api
```

Observed:
- `postgres`, `api`, and `frontend` containers started successfully.
- API healthcheck passed.
- API startup logs show migrations applied.

---

## A) Seed login flow baseline

Validation command used: API login check against seeded credentials.

Result:

```text
SEED_LOGIN_STATUS 401 {'error': 'unauthenticated'}
```

Baseline conclusion:
- On a fresh DB after `down -v` + `up --build`, seed credentials are **not** available by default.
- Current behavior is **not zero-step seed login**.

---

## B) Projects visibility baseline

Scenario tested:
1. User A creates project.
2. User A creates task assigned to User B (B gets access).
3. User B creates another task in same project, but assigned to A.
4. User A deletes the first bootstrap task.
5. User B lists projects again.

Key outputs:

```text
B_SEES_PROJECT_BEFORE 200 True count 1
B_CREATES_TASK 201 ...
A_DELETES_BOOTSTRAP 204
B_SEES_PROJECT_AFTER 200 False count 0
```

Baseline conclusion:
- Current visibility rule is effectively owner OR current assignee.
- User who is task **creator-only** (no longer assignee) loses project visibility.
- This confirms the earlier spec-gap risk.

---

## C) Filters + task board baseline

UI smoke run (Playwright) confirmed:
- Project detail filter by status triggers filtered request.
- Empty filtered state is visible.
- Clear filters restores task board list.

Key outputs:

```text
FILTER_EMPTY_STATE true
TASK_VISIBLE_AFTER_CLEAR true
TASK_FILTER_REQUEST_COUNT 2
REQ http://localhost:8080/projects/<id>/tasks
REQ http://localhost:8080/projects/<id>/tasks?status=done
```

Additional API filter checks:

```text
TASK_FILTER_ALL 200 3
TASK_FILTER_STATUS_DONE 200 1
TASK_FILTER_ASSIGNEE_A 200 2
```

Baseline conclusion:
- Task board and filtering are functioning.
- Status and assignee filtering endpoints behave correctly at API level.

---

## Phase 0.3 status

- [x] App runs in current baseline.
- [x] Seed login flow baseline captured.
- [x] Projects visibility baseline captured.
- [x] Filters + task board baseline captured.
