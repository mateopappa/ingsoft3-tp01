# Flow — DevOps Evolution Roadmap

## 1. Overview & Strategy

**Flow** is intentionally designed to evolve through an 8-stage progressive DevOps journey.

Rather than implementing infrastructure tools prematurely, each stage introduces tools only when solving a concrete operational requirement.

```text
Stage 0: Initial Monolith (Go + PostgreSQL + HTML/CSS/JS, Native Execution)
   ↓
Stage 1 (TP2): Dockerfile, Multi-Stage Builds, Containerization & Compose
   ↓
Stage 2 (TP4): Continuous Integration Pipeline (GitHub Actions)
   ↓
Stage 3 (TP5): Comprehensive Automated Testing (8+ Backend, 4+ Frontend Tests)
   ↓
Stage 4 (TP6): Continuous Delivery & Multi-Environment Deployments (QA/Prod)
   ↓
Stage 5 (TP7): Immutable Container Releases & Registry Packaging (ghcr.io)
   ↓
Stage 6 (TP8): Infrastructure as Code (Terraform / OpenTofu)
   ↓
Stage 7 (TP9): DevSecOps & Observability (Scanning, Security Headers, slog)
   ↓
Final Integration: Complete End-to-End Delivery System
```

---

## 2. Detailed Stages Breakdown

### Stage 0 — Initial Application (Current Baseline)
* **Stack**: Go 1.22+, PostgreSQL 15+, Vanilla HTML/CSS/JS.
* **Execution**: 100% native on developer machine; **zero Docker dependency**.
* **Focus**: Solid domain business logic, data persistence, and clean monolithic architecture.

---

### Stage 1 — Containerization (TP2)
* **Goal**: Package application and database into reproducible container definitions.
* **Deliverables**:
  * Multi-stage `Dockerfile` (Go builder stage + minimal Alpine runtime image).
  * `docker-compose.yml` orchestrating `flow-app` and `postgres` with health checks and persistent volumes.
* **Benefit**: Eliminates *"works on my machine"* syndrome across team environments.

---

### Stage 2 — Continuous Integration Pipeline (TP4)
* **Goal**: Automate build and verification checks on every pull request.
* **Deliverables**:
  * `.github/workflows/ci.yml` running linter, unit tests, and compilation on PRs.
  * Branch protection rules on `main`.
* **Benefit**: Enforces code quality gates before code enters the main branch.

---

### Stage 3 — Automated Testing & Quality Gates (TP5)
* **Goal**: Expand test suite to guarantee regression prevention.
* **Deliverables**:
  * 8+ backend unit & integration tests covering domain rules, duration validation, and single-timer invariants.
  * 4+ frontend behavior tests.
  * Code coverage reporting.
* **Benefit**: High deployment confidence.

---

### Stage 4 — Continuous Delivery & Environments (TP6)
* **Goal**: Automated deployment to isolated staging and production environments.
* **Deliverables**:
  * Environment variable-driven configuration (`DATABASE_URL`, `APP_ENV`, `SESSION_SECRET`).
  * Zero-downtime deployment scripts and approval gates.
* **Benefit**: Predictable, risk-free release cycles.

---

### Stage 5 — Container Packaging & Releases (TP7)
* **Goal**: Publish immutable versioned images to a container registry.
* **Deliverables**:
  * Automated image builds and pushes to GitHub Container Registry (`ghcr.io`).
  * Semantic Versioning (`v1.0.0`) release tags.
* **Benefit**: Same verified artifact promoted across all environments.

---

### Stage 6 — Infrastructure as Code (TP8)
* **Goal**: Declare all cloud compute and database infrastructure in code.
* **Deliverables**:
  * Terraform / OpenTofu modules provisioning cloud hosting and managed PostgreSQL.
* **Benefit**: Reproducible infrastructure with zero manual click-ops.

---

### Stage 7 — DevSecOps & Observability (TP9)
* **Goal**: Production-grade security scanning and operational telemetry.
* **Deliverables**:
  * Automated vulnerability scanning with Trivy and `govulncheck`.
  * Structured JSON logging via `log/slog`.
  * Security headers and least-privilege container runtime.
* **Benefit**: Proactive security and visibility into application runtime health.

---

### Final Integration
The complete delivery chain will demonstrate:
```text
Source Code ──► Git ──► CI ──► Tests ──► Build ──► Package ──► Infra ──► Deploy ──► Security ──► Running App
```
