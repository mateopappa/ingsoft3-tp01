# Flow — Architecture Decision Records (ADRs)

This document records the foundational architectural decisions for **Flow**.

---

## ADR-001: Go Backend

### Context
We need a robust, fast, low-overhead backend language that compiles to a single binary with zero external runtime dependencies, providing standard library networking and database capabilities.

### Decision
Use **Go 1.22+** with the standard library (`net/http`, `database/sql`, `log/slog`, `crypto/rand`) and standard PostgreSQL driver (`github.com/lib/pq`).

### Alternatives Considered
* **Node.js / Express**: Heavy runtime, large dependency trees, and higher memory footprint.
* **Python / FastAPI**: Higher memory usage, slower cold starts.
* **Java / Spring Boot**: High operational complexity for a small application.

### Consequences
* **Positive**: Fast compile times, tiny runtime footprint (< 20MB RAM), standard library excellence, and straightforward local execution without Docker.
* **Negative**: Requires explicit error handling and model mapping boilerplate.

---

## ADR-002: PostgreSQL Database

### Context
The application requires strict ACID transactions, strong referential integrity, and consistent date/time arithmetic.

### Decision
Use **PostgreSQL 15+** from the beginning, configured strictly through the `DATABASE_URL` environment variable.

### Alternatives Considered
* **SQLite**: Lacks rich network concurrency and diverges from production deployment workflows.
* **MySQL**: Viable, but PostgreSQL provides superior interval arithmetic and UUID support.

### Consequences
* **Positive**: Enforces database invariants (`UNIQUE`, `ON DELETE RESTRICT`) and guarantees transactional consistency for timer stop operations.
* **Negative**: Requires running a local PostgreSQL instance during development.

---

## ADR-003: Monolithic Architecture

### Context
We need an architecture that minimizes operational overhead, is trivial to run locally, and serves as an understandable foundation for DevOps coursework.

### Decision
Build a single **Modular Monolith** where backend services, database migrations, and web frontend assets reside within a single codebase and compile to a single executable.

### Alternatives Considered
* **Microservices**: Unnecessary network latency, complex distributed transactions, and multiple repositories.
* **Serverless Functions**: Cold starts and difficult local testing.

### Consequences
* **Positive**: Single process, single `go run` command to start the entire system, zero inter-service network failure modes.
* **Negative**: All domains share the same compute process.

---

## ADR-004: Vanilla HTML/CSS/JS Frontend

### Context
We need a clean, responsive, fast UI without introducing massive npm toolchains, complex build steps, or client-side framework bloat.

### Decision
Build the user interface using **semantic HTML5, modern CSS3 (Flexbox/Grid), and vanilla ES6+ JavaScript**, served directly by the Go application.

### Alternatives Considered
* **React / Next.js / Vue**: Requires Node.js build pipelines, large bundles, hydration mismatches, and CORS configuration.
* **Server-Only Templates (no JS)**: Would prevent real-time live timer ticking in the browser.

### Consequences
* **Positive**: Zero npm build step, instant page loads (< 200ms), zero client dependency vulnerabilities, and 100% embeddable directly inside the Go binary.
* **Negative**: DOM updates are handled via standard vanilla JavaScript.

---

## ADR-005: REST API

### Context
The frontend needs to interact with backend business logic for timer actions, CRUD operations, authentication, and dynamic dashboard metrics.

### Decision
Implement a clean **RESTful JSON API** under the `/api` namespace, adhering to standard HTTP methods (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`) and status codes.

### Alternatives Considered
* **GraphQL**: Unnecessary schema complexity and query parsing overhead for a small application.
* **gRPC-Web**: Requires protocol buffer compilation and browser proxying.

### Consequences
* **Positive**: Simple, testable with `curl` or standard Go `net/http/httptest`, clear contract between frontend and backend.
* **Negative**: Multiple API calls required if not batched.

---

## ADR-006: Server-Side Persistent Timer

### Context
The timer must survive browser tab closure, device reboots, and network disconnects without losing tracked time.

### Decision
Persist the timer's start timestamp in PostgreSQL (`active_timers.started_at`). The server computes elapsed duration dynamically (`NOW() - started_at`). The browser is strictly a display client.

### Alternatives Considered
* **Browser LocalStorage Timer**: Fragile; loses tracking if browser storage is cleared or device changes.
* **WebSocket Ticking Server**: Unnecessary network overhead; time difference math is exact and stateless.

### Consequences
* **Positive**: 100% resilient to browser closure; zero timer drift across machines.
* **Negative**: Requires an initial API call on page load to fetch the active timer state.

---

## ADR-007: One Active Timer per User

### Context
A user must not be allowed to run multiple timers concurrently, as this corrupts time tracking semantics.

### Decision
Enforce the invariant at the database level via a **`UNIQUE (user_id)` constraint on `active_timers`** in combination with application service checks. Attempting to start a second timer returns an **HTTP 409 Conflict** error with the active timer details.

### Alternatives Considered
* **Application-Only Check**: Vulnerable to race conditions under concurrent requests.
* **Auto-Stop Previous Timer**: Dangerous; could silently truncate work sessions without user confirmation.

### Consequences
* **Positive**: Invariant is impossible to violate under any concurrency scenario.
* **Negative**: Client must handle HTTP 409 and prompt the user to stop or discard the active timer.
