# Flow — Local Development Guide (No Docker Required)

## 1. Initial Development Philosophy

**Flow is built to run directly on the host operating system without any containerization dependencies.**

Docker, Compose, and Kubernetes are explicitly **NOT** required for Stage 0 development. The system architecture is straightforward:

```text
Browser ──► Go Application (localhost:8080) ──► PostgreSQL (localhost:5432)
```

---

## 2. Prerequisites

* **Go**: `1.22+` / `1.24+` (`go version`)
* **PostgreSQL**: `15+` / `16+` (`psql --version`)
* **GNU Make**: (`make --version`)

---

## 3. Step-by-Step Local Setup

### Step 1: Clone the Repository
```bash
git clone https://github.com/mateopappa/ingsoft3-tp01.git
cd ingsoft3-tp01
```

### Step 2: Create Local PostgreSQL Database
Using `createdb` or `psql`:
```bash
createdb flow_db
# Or via psql:
# psql -U postgres -c "CREATE DATABASE flow_db;"
```

### Step 3: Configure Environment Variables
Copy the `.env.example` file:
```bash
cp .env.example .env
```

Edit `.env` to match your local PostgreSQL credentials:
```env
APP_ENV=development
APP_PORT=8080
DATABASE_URL=postgres://localhost:5432/flow_db?sslmode=disable
SESSION_SECRET=local_development_secret_key_1234567890abcdef
```

### Step 4: Run Database Migrations
Migrations execute automatically on server startup by default. You can also run them explicitly via Make:
```bash
make migrate
```

### Step 5: Start the Flow Application
```bash
make run
# Equivalent to:
# go run cmd/server/main.go
```

### Step 6: Open in Browser
Navigate to **`http://localhost:8080`** in your browser.

---

## 4. Standard Development Commands

| Task | Make Command | Underlying Go Command |
|:---|:---|:---|
| **Run server** | `make run` | `go run cmd/server/main.go` |
| **Build binary** | `make build` | `go build -o bin/flow cmd/server/main.go` |
| **Run tests** | `make test` | `go test -v -race ./...` |
| **Run migrations** | `make migrate` | `go run cmd/server/main.go -migrate` |
| **Format code** | `make fmt` | `go fmt ./...` |
| **Clean artifacts** | `make clean` | `rm -rf bin/` |

---

## 5. Troubleshooting

### "connection refused" / "cannot connect to database"
* Verify PostgreSQL is running on your machine:
  * macOS (Homebrew): `brew services start postgresql@15`
  * Linux (systemd): `sudo systemctl start postgresql`
* Verify `DATABASE_URL` in `.env` matches your username, port, and database name.

### Port 8080 already in use
* Change `APP_PORT=8081` in your `.env` file.
