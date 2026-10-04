# Local environment checklist

Use this checklist before running JobHub locally for the first time on a machine.

## Required tools

- Docker with Docker Compose v2.
- Go 1.25 or newer.
- Node.js 22.12 or newer with npm.
- A local `.env` created from `.env.example`.

## Safe defaults

- Local Docker PostgreSQL uses `sslmode=disable`.
- Supabase and other remote PostgreSQL connections use `sslmode=require`.
- Provider keys stay in `.env` and must not be copied into documentation,
  screenshots or committed files.
- `VITE_API_URL` should point at the local backend for ordinary development.

## Quick smoke check

1. Run `make local-up`.
2. Start the frontend with `npm run dev` from `frontend/`.
3. Open `http://localhost:5173`.
4. Check `http://localhost:8080/api/v1/health`.
5. Stop local services with `make local-down` when finished.

Use `make dev` only when the local POC database described in `README.md` is
available.
