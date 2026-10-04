# Local troubleshooting

Use this page for common local development failures before reaching for broader
debugging.

## Backend port already in use

Do not run the Docker backend and manual Go backend at the same time. Stop one
process before starting the other.

## Frontend cannot reach the API

- Check `VITE_API_URL`.
- Check the backend health endpoint.
- Check `FRONTEND_ORIGIN` in `.env`.
- Restart Vite after changing environment variables.

## Local database looks empty

- `make local-up` starts an empty local database.
- `make dev` expects the local POC database described in `README.md`.
- Jooble import requires `JOOBLE_API_KEY` and remains request-capped.

## Remote migration refused

Confirm `APP_ENV=staging`, `sslmode=require` and a direct or session connection
on port `5432`.
