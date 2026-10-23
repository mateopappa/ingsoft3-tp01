# Flow — Testing Strategy & Test Specifications

## 1. Testing Strategy

Unit tests target **service-layer business rules** with injected doubles (no Postgres, no HTTP).
Frontend unit tests target **pure modules** under `frontend/src/lib/` (no DOM). Coverage is measured
in CI inside Docker test stages; thresholds fail the same required jobs from TP4 (`build-backend`,
`build-frontend`).

```bash
# Backend unit tests
make test
# Backend coverage gate (70% statements on *service.go)
make test-coverage-backend

# Frontend (vitest + thresholds)
cd frontend && npm ci && npm run test:coverage
# or:
make test-coverage-frontend
```

---

## 2. Test Specifications

### 2.1 Backend Tests (samples)

| Test | Rule | Notes |
|:---|:---|:---|
| `TestCreateActivity_DurationValidation` | duration > 0 | Parameterized table |
| `TestCreateActivity_EmptyDescriptionRejected` | description required | Error case |
| `TestCreateActivity_InvalidCategoryRejected_AndRepoNotCalled` | category ownership | **Mock**: store must not be called |
| `TestDeleteCategory_InUseRejected` | category protection | Parameterized |
| `TestStartTimer_SecondActiveRejected` | one active timer | **Mock** interaction |
| `TestRegister_PasswordTooShort` | password ≥ 8 | Parameterized |
| `TestGetMetrics_CategoryPercentages` | dashboard math | Stub metrics store |
| `TestFormatDuration` | display formatting | Parameterized |

### 2.2 Frontend Tests (`src/lib`)

| Module | Techniques |
|:---|:---|
| `format.test.js` | Parameterized `it.each` |
| `validate.test.js` | Error cases (email, password, duration) |
| `api.test.js` | **Mock** `fetch` via `vi.fn()` |
| `filters.test.js` | Parameterized query building |

---

## 3. Coverage scope

- **Backend in count:** `internal/*/service.go` only (script filters the cover profile).
- **Frontend in count:** `src/lib/**/*.js` (`vitest.config.js` `include`).
- **Out of count:** `main`, handlers, repositories, models, migrations, DOM controller.
