# TaskFlow

## Running With Docker

```bash
cp .env.example .env
docker compose up --build
```

The API will be available at `http://localhost:8080`, PostgreSQL at `localhost:5432`, and database migrations run automatically when the API container starts.

## Local Backend Env

For running the backend outside Docker, copy `backend/.env.example` to `backend/.env` and adjust `DATABASE_URL` as needed.
