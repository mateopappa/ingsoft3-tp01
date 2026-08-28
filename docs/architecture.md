# Flow — Monolithic Architecture Specification

## 1. System Overview

**Flow** is designed as a **Modular Monolith** in **Go** backed by **PostgreSQL**.

The architecture emphasizes simplicity, high readability, standard library usage, and direct native execution on the developer workstation without requiring Docker in initial development.

```text
                 Browser Client
          (Vanilla HTML5 / CSS3 / ES6+ JS)
                         │
                         │ HTTP / JSON & HTML
                         ▼
        ┌───────────────────────────────────┐
        │            Go Monolith            │
        │                                   │
        │  ├── HTTP Router (net/http)       │
        │  ├── Security & Auth Middleware   │
        │  ├── Auth & Session Service       │
        │  ├── Category Service             │
        │  ├── Activity Service             │
        │  ├── Persistent Timer Service     │
        │  ├── Dashboard Analytics Service  │
        │  └── Embedded Web Assets          │
        └─────────────────┬─────────────────┘
                          │
                          │ SQL (database/sql + pq driver)
                          ▼
        ┌───────────────────────────────────┐
        │          PostgreSQL 15+           │
        │ (users, categories, activities,   │
        │  active_timers, sessions)         │
        └───────────────────────────────────┘
```

---

## 2. Layered Package Structure

The repository follows standard, idiomatic Go package layout:

```text
flow/
├── cmd/
│   └── server/
│       └── main.go                 # Server entrypoint, configuration & dependency injection
├── internal/
│   ├── config/
│   │   └── config.go               # Environment configuration loader
│   ├── database/
│   │   ├── db.go                   # PostgreSQL connection pool setup
│   │   └── migrate.go              # Embedded SQL migration runner
│   ├── models/
│   │   ├── user.go                 # User & Session domain models
│   │   ├── category.go             # Category domain model
│   │   ├── activity.go             # Activity domain model & filter DTOs
│   │   ├── timer.go                # ActiveTimer domain model
│   │   └── dashboard.go            # Dashboard analytics DTOs
│   ├── auth/
│   │   ├── handler.go              # Auth HTTP endpoints (register, login, logout, me)
│   │   ├── service.go              # Password hashing & session logic
│   │   └── repository.go           # User & session PostgreSQL queries
│   ├── categories/
│   │   ├── handler.go              # Category HTTP endpoints
│   │   ├── service.go              # Category validation & referential checks
│   │   └── repository.go           # Category PostgreSQL queries
│   ├── activities/
│   │   ├── handler.go              # Activity CRUD & search/filter endpoints
│   │   ├── service.go              # Activity business rules & validation
│   │   └── repository.go           # Dynamic search/filter SQL queries
│   ├── timer/
│   │   ├── handler.go              # Timer endpoints (get, start, stop, discard)
│   │   ├── service.go              # Single timer invariant & atomic stop transaction
│   │   └── repository.go           # Active timer PostgreSQL queries
│   ├── dashboard/
│   │   ├── handler.go              # Dashboard analytics endpoint
│   │   ├── service.go              # Aggregation math & metric transformations
│   │   └── repository.go           # Analytical SQL queries
│   └── http/
│       ├── router.go               # ServeMux routing & route mapping
│       ├── response.go             # Standardized JSON response helpers
│       └── middleware/
│           ├── auth.go             # Session validation & User context injection
│           ├── logging.go          # Request logging via log/slog
│           ├── recovery.go         # Panic recovery middleware
│           └── security.go         # Security headers middleware
├── web/
│   ├── templates/
│   │   └── index.html              # Core application HTML interface
│   └── static/
│       ├── css/
│       │   └── style.css           # Clean modern SaaS styling
│       └── js/
│           └── app.js              # Vanilla JS application controller & live timer
├── migrations/
│   ├── 001_create_users.sql
│   ├── 002_create_categories.sql
│   ├── 003_create_activities.sql
│   ├── 004_create_active_timers.sql
│   └── 005_create_sessions.sql
├── tests/
│   ├── unit/                       # Fast domain unit tests
│   └── integration/                # Database repository integration tests
├── docs/                           # Documentation suite
├── .env.example                    # Environment variable template
├── .gitignore
├── go.mod
├── go.sum
├── Makefile                        # Native development automation
└── README.md
```

---

## 3. Data Flow & Transaction Guarantees

### 3.1 Persistent Timer Lifecycle

```mermaid
sequenceDiagram
    autonumber
    actor User as Browser
    participant Handler as Timer Handler
    participant Svc as Timer Service
    participant Repo as Timer Repository
    participant DB as PostgreSQL

    Note over User,DB: Start Persistent Timer
    User->>Handler: POST /api/timer/start { description, category_id }
    Handler->>Svc: StartTimer(ctx, userID, description, categoryID)
    Svc->>Repo: CheckActiveTimer(ctx, userID)
    Repo->>DB: SELECT * FROM active_timers WHERE user_id = $1
    alt Timer already active
        DB-->>Repo: Existing Timer Record
        Repo-->>Svc: ErrTimerAlreadyActive
        Svc-->>Handler: Error Conflict
        Handler-->>User: 409 Conflict { error: "Timer already active" }
    else No active timer
        Svc->>Repo: InsertActiveTimer(ctx, timer)
        Repo->>DB: INSERT INTO active_timers (user_id, category_id, description, started_at) VALUES (...)
        DB-->>Repo: Saved
        Repo-->>Svc: ActiveTimer
        Svc-->>Handler: ActiveTimer
        Handler-->>User: 201 Created { active_timer }
    end

    Note over User,DB: Stop Timer & Create Activity (Atomic Transaction)
    User->>Handler: POST /api/timer/stop { note, favorite }
    Handler->>Svc: StopTimer(ctx, userID, note, favorite)
    Svc->>DB: BEGIN TRANSACTION
    Svc->>DB: SELECT * FROM active_timers WHERE user_id = $1 FOR UPDATE
    Note over Svc: duration = NOW() - started_at
    Svc->>DB: INSERT INTO activities (user_id, category_id, description, duration_seconds, activity_date, note, favorite) VALUES (...)
    Svc->>DB: DELETE FROM active_timers WHERE user_id = $1
    Svc->>DB: COMMIT TRANSACTION
    Svc-->>Handler: Created Activity Record
    Handler-->>User: 200 OK { activity }
```

### 3.2 Category Referential Protection
* Deleting a category that has logged activities is rejected before deletion occurs.
* The system queries `COUNT(*) FROM activities WHERE category_id = $1 AND user_id = $2`.
* If count > 0, an HTTP 409 Conflict error is returned with a clear user message.
