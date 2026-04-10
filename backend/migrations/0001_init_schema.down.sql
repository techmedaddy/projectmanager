DROP TRIGGER IF EXISTS trg_tasks_set_updated_at ON tasks;
DROP FUNCTION IF EXISTS set_tasks_updated_at();

DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS task_priority;
DROP TYPE IF EXISTS task_status;

DROP EXTENSION IF EXISTS pgcrypto;
