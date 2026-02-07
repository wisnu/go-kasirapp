# Kasir API

REST API for product and category management, backed by PostgreSQL.

## Run the server

```bash
go run .
```

## Environment

Create a `.env` file at the project root (or export the variables) to configure the app:

```bash
APP_NAME=kasir-app
APP_ENV=development
APP_PORT=8080

DB_DRIVER=postgres
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=
DB_NAME=kasirapp
DB_SSLMODE=disable
```

---

## Testing

Two test suites are available. Both support a **unit mode** (default, no network) and an **integration mode** that hits a live deployed service.

### 1. Go tests (`go test`)

Unit mode uses `sqlmock` + `httptest` — no server or database needed:

```bash
go test -v ./...
```

Integration mode makes real HTTP calls to a running service. Set `BASE_URL` before running:

```bash
BASE_URL=http://localhost:8080 go test -v ./...
```

### 2. Curl tests (`tests/test_curl.sh`)

Requires a running server. Pass the target URL as the first argument (defaults to `http://localhost:8080`):

```bash
# against localhost
bash tests/test_curl.sh

# against a deployed service
bash tests/test_curl.sh https://your-deployed-api.example.com
```

The script prints a PASS/FAIL summary at the end and exits with status 1 if any assertion fails.
