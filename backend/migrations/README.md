# Migrations

This directory contains numbered SQL migrations.

Current schema choices:
- PostgreSQL enum types for `task_status` and `task_priority`
- UUID primary keys generated with `pgcrypto`
- `tasks.creator_id` added as a practical extension so task delete authorization can distinguish task creators from project owners
- `tasks.updated_at` maintained with a trigger
