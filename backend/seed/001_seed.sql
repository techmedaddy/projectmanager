BEGIN;

INSERT INTO users (
    id,
    name,
    email,
    password,
    created_at
) VALUES (
    '11111111-1111-1111-1111-111111111111',
    'Test User',
    'test@example.com',
    '$2b$12$k3NHj0yBtD2XgjBk/SThuOuHDAos.2HRewt2pMMW3BYk/sc1r9yxm',
    '2026-04-10T15:11:09Z'
);

INSERT INTO projects (
    id,
    name,
    description,
    owner_id,
    created_at
) VALUES (
    '22222222-2222-2222-2222-222222222222',
    'TaskFlow Launch',
    'Seed project used for local development and assignment review.',
    '11111111-1111-1111-1111-111111111111',
    '2026-04-10T15:11:09Z'
);

INSERT INTO tasks (
    id,
    title,
    description,
    status,
    priority,
    project_id,
    assignee_id,
    creator_id,
    due_date,
    created_at,
    updated_at
) VALUES
(
    '33333333-3333-3333-3333-333333333333',
    'Set up API foundation',
    'Create the initial backend scaffold and configuration loading.',
    'todo',
    'high',
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    '11111111-1111-1111-1111-111111111111',
    '2026-04-15',
    '2026-04-10T15:11:09Z',
    '2026-04-10T15:11:09Z'
),
(
    '44444444-4444-4444-4444-444444444444',
    'Implement authentication',
    'Add registration, login, JWT middleware, and password hashing.',
    'in_progress',
    'medium',
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    '11111111-1111-1111-1111-111111111111',
    '2026-04-16',
    '2026-04-10T15:11:09Z',
    '2026-04-10T15:11:09Z'
),
(
    '55555555-5555-5555-5555-555555555555',
    'Design responsive project board',
    'Create the initial project detail view and task presentation.',
    'done',
    'low',
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    '11111111-1111-1111-1111-111111111111',
    '2026-04-17',
    '2026-04-10T15:11:09Z',
    '2026-04-10T15:11:09Z'
);

COMMIT;
