# projectmanager

projectmanager is a full-stack project and task management platform with a Go API, PostgreSQL, and a React frontend. It supports authentication, project collaboration, task lifecycle management, and Docker-based local development.

## Overview

### Tech stack

- Go 1.22
- PostgreSQL 16
- `pgx` for database access
- `golang-migrate` with embedded SQL migrations
- JWT for auth
- `bcrypt` for password hashing
- Docker Compose for local setup

### What the backend does

- Registers and authenticates users
- Creates and lists projects
- Returns project detail with tasks
- Creates, updates, filters, and deletes tasks
- Enforces `401` vs `403` correctly
- Starts through Docker with automatic migrations

## Architecture Decisions

### Why this structure

- `cmd/api` contains HTTP bootstrap and handlers.
- `internal/auth`, `internal/projects`, `internal/tasks`, and `internal/users` hold domain logic.
- `internal/db` contains PostgreSQL connection management, nullable helpers, transactions, pagination scaffolding, and migration startup.
- `migrations` and `seed` keep schema and test data explicit and reviewable.

### Key choices

- Hand-written SQL instead of an ORM: this keeps queries visible, predictable, and easy to reason about during review.
- Standard library `http.ServeMux`: lightweight and clear for this service size.
- Embedded migrations: the API binary can apply SQL migrations automatically at container startup.
- Service layer for permissions: project and task authorization rules live in business logic instead of being spread through handlers.
- Explicit PATCH nullability handling for tasks: this avoids confusing “field omitted” with “field set to null”.

### Tradeoffs

- I chose clarity over abstraction, so there is some repetition across handlers and DTOs.
- Startup seeding is intentionally optimized for local/dev convenience and should be disabled in production unless explicitly desired.
- Pagination support is intentionally minimal (page/limit + next/prev controls) without advanced sorting UI controls.

## System Design

### HLD (High-Level Design)

```mermaid
flowchart LR
  U[Client Browser] --> FE[Frontend React App]
  FE --> API[Go REST API]
  API --> DB[(PostgreSQL)]
  API --> MIG[Embedded SQL Migrations]
```

### Data Flow

```mermaid
sequenceDiagram
  participant User
  participant Frontend
  participant API
  participant DB

  User->>Frontend: Login/Register
  Frontend->>API: POST /auth/login or /auth/register
  API->>DB: Validate / create user
  DB-->>API: User data
  API-->>Frontend: JWT / user response

  User->>Frontend: Open project and apply filters
  Frontend->>API: GET /projects/:id
  Frontend->>API: GET /projects/:id/tasks?status=&assignee=
  API->>DB: Permission + filtered task query
  DB-->>API: Project/tasks rows
  API-->>Frontend: JSON response
```

### LLD (Low-Level Design)

```mermaid
flowchart TD
  Router[Router + Middleware] --> Handlers[HTTP Handlers]
  Handlers --> Services[Domain Services]
  Services --> Repos[Repositories]
  Repos --> PG[(PostgreSQL)]
  Handlers --> Resp[Response/Error Helpers]
```

```mermaid
erDiagram
  USERS ||--o{ PROJECTS : owns
  PROJECTS ||--o{ TASKS : contains
  USERS ||--o{ TASKS : creates
  USERS ||--o{ TASKS : assigned

  USERS {
    uuid id PK
    string name
    string email
    string password
    timestamp created_at
  }

  PROJECTS {
    uuid id PK
    string name
    string description
    uuid owner_id FK
    timestamp created_at
  }

  TASKS {
    uuid id PK
    string title
    string description
    string status
    string priority
    uuid project_id FK
    uuid assignee_id FK
    uuid creator_id FK
    date due_date
    timestamp created_at
    timestamp updated_at
  }
```

## Running Locally

### Docker

From clone to browser (full stack):

```bash
git clone https://github.com/<your-username>/projectmanager.git
cd projectmanager
cp .env.example .env
docker compose up --build
```

`docker compose up --build` also works without creating `.env` because compose defaults are defined, but copying `.env.example` is recommended for explicit local config.

Services:

- Frontend: `http://localhost:3001`
- API: `http://localhost:8080`
- PostgreSQL: `localhost:5432`

Open the app in your browser at `http://localhost:3001`.

Backend health endpoint:

- `GET http://localhost:8080/health`

Health check:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{"status":"ok"}
```

### Local backend only

If you want to run the API outside Docker:

```bash
cp backend/.env.example backend/.env
cd backend
go run ./cmd/api
```

That assumes PostgreSQL is already running and `DATABASE_URL` points to it.

## Environment Variables

### Root `.env` for Docker Compose

Copy [`.env.example`](/home/techmedaddy/projects/projectmanager/.env.example) to `.env`.

| Variable | Purpose | Default |
|---|---|---|
| `POSTGRES_DB` | Compose database name | `projectmanager` |
| `POSTGRES_USER` | Compose database user | `postgres` |
| `POSTGRES_PASSWORD` | Compose database password | `postgres` |
| `POSTGRES_PORT` | Host port mapped to PostgreSQL | `5432` |
| `APP_PORT` | Host port mapped to the API | `8080` |
| `FRONTEND_PORT` | Host port mapped to frontend container | `3001` |
| `VITE_API_BASE_URL` | API base URL injected into frontend build | `http://localhost:8080` |
| `JWT_SECRET` | JWT signing secret | `change-this-secret-before-production` |
| `JWT_EXPIRY_HOURS` | Token lifetime in hours | `24` |
| `BCRYPT_COST` | Password hashing cost | `12` |
| `AUTO_SEED` | Auto-run embedded seed SQL at API startup (set `false` in production if you do not want startup seeding) | `true` |

### Backend env for local Go execution

Copy [`backend/.env.example`](/home/techmedaddy/projects/projectmanager/backend/.env.example) to `backend/.env`.

| Variable | Purpose |
|---|---|
| `APP_PORT` | HTTP server port |
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | JWT signing secret |
| `JWT_EXPIRY_HOURS` | Token lifetime in hours |
| `BCRYPT_COST` | Password hashing cost |
| `AUTO_SEED` | Auto-run embedded seed SQL at startup |
| `TEST_DATABASE_URL` | Dedicated database used by integration tests |

## Migration Strategy

Auto-run status: **Enabled**.

- Migrations are automatically executed by the API on startup.
- No manual migration command is required for normal Docker startup.

- Schema changes live in [`backend/migrations`](/home/techmedaddy/projects/projectmanager/backend/migrations).
- Migrations are plain SQL with both `up` and `down` files.
- The API embeds those SQL files into the binary and runs pending migrations automatically on startup.
- In Docker, `docker compose up --build` is enough to apply migrations.

Relevant files:

- [`backend/migrations/0001_init_schema.up.sql`](/home/techmedaddy/projects/projectmanager/backend/migrations/0001_init_schema.up.sql)
- [`backend/migrations/0001_init_schema.down.sql`](/home/techmedaddy/projects/projectmanager/backend/migrations/0001_init_schema.down.sql)
- [`backend/internal/db/migrate.go`](/home/techmedaddy/projects/projectmanager/backend/internal/db/migrate.go)

## Seed Data

Seed SQL lives in [`backend/seed/001_seed.sql`](/home/techmedaddy/projects/projectmanager/backend/seed/001_seed.sql).

By default (`AUTO_SEED=true`), the API now runs an idempotent startup seed check after migrations:
- if seed user data is missing, it applies embedded seed SQL
- if already present, it skips seeding safely (`already_seeded`)

For production-safe behavior, set `AUTO_SEED=false` in your environment.

## Test Credentials

Use these credentials immediately after a fresh compose startup (with `AUTO_SEED=true`):

- Email: `test@example.com`
- Password: `password123`

Manual fallback seed command (Docker):

```bash
cat backend/seed/001_seed.sql | docker compose exec -T postgres psql -U postgres -d projectmanager
```

If you are running PostgreSQL locally outside Docker:

```bash
psql "$DATABASE_URL" -f backend/seed/001_seed.sql
```

## Frontend Behavior

### Pages and routes

- `/login`
  - Authenticates user and stores JWT in localStorage.
  - Immediately fetches `GET /auth/me` after login to hydrate navbar user state.
- `/register`
  - Registers a new user and redirects to login.
- `/projects` (protected)
  - Lists accessible projects.
  - Supports project creation through modal form.
- `/projects/:id` (protected)
  - Shows project header and task board.
  - Includes task create/edit modal.
  - Includes status + assignee filter controls with clear/reset action.

### Auth persistence and protected routes

- JWT is persisted in browser localStorage for session continuity.
- On app boot, auth context calls `GET /auth/me` when token exists.
- Protected routes redirect unauthenticated users to `/login`.
- Logout clears token and user state.

### Task filters behavior

- Status filter options: `all`, `todo`, `in_progress`, `done`.
- Assignee filter options: `all` + assignee IDs found in project tasks.
- Changing filters re-queries tasks from API (not just client-side filtering).
- Empty filtered result shows an explicit empty state with clear-filters CTA.

### Task modal behavior

- Create and edit use the same modal.
- Fields: `title`, `description`, `status`, `priority`, `due_date`, `assignee`.
- Assignee now uses a dropdown sourced from `GET /projects/:id/assignees`.
- Assignee supports set and clear:
  - selecting a user sends that UUID in `assignee_id`
  - selecting **Unassigned** sends `"assignee_id": null`
- Nullable date clear sends `"due_date": null`.
- API field errors are mapped to inline form validation.

## Frontend API Usage Notes

- Base URL is configurable via `VITE_API_BASE_URL` (default `http://localhost:8080`).
- Task filter requests use backend query params exactly as required:
  - `GET /projects/:id/tasks?status=todo|in_progress|done`
  - `GET /projects/:id/tasks?assignee=<user-uuid>`
  - or both together in one request.
- Auth header handling:
  - frontend automatically adds `Authorization: Bearer <token>` when token exists.
- Error handling strategy:
  - API errors are parsed into `ApiError` with `status`, `message`, and optional `fields`.
  - field errors are mapped to form validation messages.
  - non-field errors are shown via toast notifications.
  - list/detail fetch failures render visible retryable error states.

## API Reference

All responses use `Content-Type: application/json`.

Non-auth endpoints require:

```http
Authorization: Bearer <token>
```

### Error format

Validation errors:

```json
{
  "error": "validation failed",
  "fields": {
    "email": "is required"
  }
}
```

Other common errors:

```json
{"error":"unauthenticated"}
{"error":"forbidden"}
{"error":"not found"}
{"error":"internal server error"}
```

### Auth

#### `POST /auth/register`

Request:

```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "password123"
}
```

Response `201 Created`:

```json
{
  "user": {
    "id": "c3a4c8f1-9c46-47d8-b2a1-f7418e3d95d1",
    "name": "Jane Doe",
    "email": "jane@example.com",
    "created_at": "2026-04-10T15:11:09Z"
  }
}
```

#### `POST /auth/login`

Request:

```json
{
  "email": "jane@example.com",
  "password": "password123"
}
```

Response `200 OK`:

```json
{
  "access_token": "<jwt>"
}
```

JWT claims include:

- `user_id`
- `email`
- `exp` with a 24-hour lifetime

#### `GET /auth/me`

Response `200 OK`:

```json
{
  "user": {
    "id": "c3a4c8f1-9c46-47d8-b2a1-f7418e3d95d1",
    "name": "Jane Doe",
    "email": "jane@example.com"
  }
}
```

### Projects

#### Pagination response contract

List endpoints return:
- `items`: current page data
- `meta.page`: current page number
- `meta.limit`: page size
- `meta.total`: total matching rows across all pages

#### `GET /projects`

Returns projects the current user can access (project owner OR task creator OR task assignee).

Supports pagination:
- `page` (default `1`)
- `limit` (default `20`)

Example:

```http
GET /projects?page=2&limit=6
```

Response `200 OK`:

```json
{
  "items": [
    {
      "id": "22222222-2222-2222-2222-222222222222",
      "name": "projectmanager Launch",
      "description": "Seed project used for local development and review.",
      "owner_id": "11111111-1111-1111-1111-111111111111",
      "created_at": "2026-04-10T15:11:09Z"
    }
  ],
  "meta": {
    "page": 2,
    "limit": 6,
    "total": 12
  }
}
```

#### `POST /projects`

Request:

```json
{
  "name": "New Project",
  "description": "API-delivered project"
}
```

Response `201 Created`:

```json
{
  "id": "9a7b4aa8-a048-4b1d-9bc0-6ec4ea9513c6",
  "name": "New Project",
  "description": "API-delivered project",
  "owner_id": "11111111-1111-1111-1111-111111111111",
  "created_at": "2026-04-10T16:00:00Z"
}
```

#### `GET /projects/:id`

Returns project detail plus its tasks. Access policy matches list visibility (owner OR task creator OR task assignee).

Response `200 OK`:

```json
{
  "project": {
    "id": "22222222-2222-2222-2222-222222222222",
    "name": "projectmanager Launch",
    "description": "Seed project used for local development and review.",
    "owner_id": "11111111-1111-1111-1111-111111111111",
    "created_at": "2026-04-10T15:11:09Z"
  },
  "tasks": [
    {
      "id": "33333333-3333-3333-3333-333333333333",
      "title": "Set up API foundation",
      "description": "Create the initial backend scaffold and configuration loading.",
      "status": "todo",
      "priority": "high",
      "project_id": "22222222-2222-2222-2222-222222222222",
      "assignee_id": "11111111-1111-1111-1111-111111111111",
      "creator_id": "11111111-1111-1111-1111-111111111111",
      "due_date": "2026-04-15",
      "created_at": "2026-04-10T15:11:09Z",
      "updated_at": "2026-04-10T15:11:09Z"
    }
  ]
}
```

#### `PATCH /projects/:id`

Owner only.

Request:

```json
{
  "name": "Renamed Project",
  "description": "Updated description"
}
```

Response `200 OK`: same shape as `POST /projects`.

#### `DELETE /projects/:id`

Owner only.

Response `204 No Content`

#### `GET /projects/:id/stats`

Returns aggregate stats for the project:
- task counts by status
- task counts grouped by assignee (`"unassigned"` bucket included)

Response `200 OK`:

```json
{
  "by_status": {
    "todo": 4,
    "in_progress": 2,
    "done": 1
  },
  "by_assignee": {
    "11111111-1111-1111-1111-111111111111": 3,
    "unassigned": 4
  }
}
```

### Tasks

#### `GET /projects/:id/tasks`

Supports filters + pagination:

- `status=todo|in_progress|done`
- `assignee=<user-uuid>`
- `page=<n>`
- `limit=<n>`

Example:

```http
GET /projects/22222222-2222-2222-2222-222222222222/tasks?status=todo&assignee=11111111-1111-1111-1111-111111111111&page=1&limit=5
```

#### Pagination query examples

Projects:

```http
GET /projects?page=1&limit=6
GET /projects?page=2&limit=6
```

Tasks without filters:

```http
GET /projects/22222222-2222-2222-2222-222222222222/tasks?page=1&limit=10
GET /projects/22222222-2222-2222-2222-222222222222/tasks?page=2&limit=10
```

Tasks with filters preserved while paginating:

```http
GET /projects/22222222-2222-2222-2222-222222222222/tasks?status=done&assignee=11111111-1111-1111-1111-111111111111&page=1&limit=5
GET /projects/22222222-2222-2222-2222-222222222222/tasks?status=done&assignee=11111111-1111-1111-1111-111111111111&page=2&limit=5
```

Response `200 OK`:

```json
{
  "items": [
    {
      "id": "33333333-3333-3333-3333-333333333333",
      "title": "Set up API foundation",
      "description": "Create the initial backend scaffold and configuration loading.",
      "status": "todo",
      "priority": "high",
      "project_id": "22222222-2222-2222-2222-222222222222",
      "assignee_id": "11111111-1111-1111-1111-111111111111",
      "creator_id": "11111111-1111-1111-1111-111111111111",
      "due_date": "2026-04-15",
      "created_at": "2026-04-10T15:11:09Z",
      "updated_at": "2026-04-10T15:11:09Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 5,
    "total": 7
  }
}
```

#### `GET /projects/:id/assignees`

Returns users eligible for assignment in the project context.

Response `200 OK`:

```json
{
  "users": [
    {
      "id": "11111111-1111-1111-1111-111111111111",
      "name": "Test User",
      "email": "test@example.com"
    }
  ]
}
```

#### `POST /projects/:id/tasks`

Request:

```json
{
  "title": "Integration Task",
  "description": "Created through the API",
  "status": "in_progress",
  "priority": "high",
  "assignee_id": "11111111-1111-1111-1111-111111111111",
  "due_date": "2026-04-25"
}
```

Response `201 Created`:

```json
{
  "id": "0fbf9d77-9490-4f53-b1d0-c2966a7c6492",
  "title": "Integration Task",
  "description": "Created through the API",
  "status": "in_progress",
  "priority": "high",
  "project_id": "22222222-2222-2222-2222-222222222222",
  "assignee_id": "11111111-1111-1111-1111-111111111111",
  "creator_id": "11111111-1111-1111-1111-111111111111",
  "due_date": "2026-04-25",
  "created_at": "2026-04-10T16:10:00Z",
  "updated_at": "2026-04-10T16:10:00Z"
}
```

#### `PATCH /tasks/:id`

Allowed fields:

- `title`
- `description`
- `status`
- `priority`
- `assignee_id`
- `due_date`

Nullable fields can be cleared by sending `null`.

Request:

```json
{
  "status": "done",
  "assignee_id": null,
  "due_date": null
}
```

Response `200 OK`: same shape as `POST /projects/:id/tasks`.

#### `DELETE /tasks/:id`

Allowed for the project owner or the task creator.

Response `204 No Content`

## Authorization Rules

Project visibility rules:

- project owners can list, read, update, and delete their own projects
- non-owners can list/read a project when they are involved in at least one task in that project (task creator or task assignee)
- non-owners cannot update or delete projects

Task rules:

- listing and creating tasks requires project visibility
- updating a task requires project visibility plus task-level permission checks
- deleting a task is allowed only for the project owner or the task creator

This keeps `401` reserved for missing/invalid authentication, while `403` is used for authenticated users attempting actions they are not allowed to perform.

## Testing

Integration tests live in [integration_test.go](/home/techmedaddy/projects/projectmanager/backend/cmd/api/integration_test.go).

Current coverage includes:

- register/login flow
- protected route rejects missing token
- create task in project
- delete task authorization

Run tests with a dedicated database:

```bash
cd backend
TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/projectmanager_test?sslmode=disable go test ./cmd/api -v
```

The test harness:

- applies migrations before tests
- truncates tables between tests
- reseeds known data before each test

## What I’d Do With More Time

- Add richer assignee UX (searchable dropdown + avatars).
- Add page-size controls and explicit page-number jump UI.
- Add OpenAPI and a checked-in API collection (Bruno/Postman) for faster API review.
- Add CI pipeline for backend tests + frontend lint/build + end-to-end smoke tests in containers.
- Improve auditability/observability (structured request correlation for auth + permission denials).
