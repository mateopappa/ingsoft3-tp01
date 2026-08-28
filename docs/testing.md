# Flow — Testing Strategy & Test Specifications

## 1. Testing Strategy

The application architecture is structured to support comprehensive, deterministic testing for both domain business rules (unit tests) and persistence/transaction guarantees (integration tests).

---

## 2. Test Specifications

### 2.1 Backend Tests

| Test # | Test Case | Target Service / Component | Scenario & Expected Outcome |
|:---|:---|:---|:---|
| **BT-1** | Valid activity accepted | `internal/activities/service.go` | Valid activity with `duration_seconds > 0`, valid category, and valid date is saved successfully. |
| **BT-2** | Zero/negative duration rejected | `internal/activities/service.go` | Submitting `duration_seconds = 0` or `-120` returns validation error and rejects creation. |
| **BT-3** | Invalid category rejected | `internal/activities/service.go` | Referencing a non-existent category ID or a category owned by another user returns an error. |
| **BT-4** | User cannot access another user's activity | `internal/activities/service.go` | User A querying, updating, or deleting User B's activity ID receives 404 / unauthorized error. |
| **BT-5** | Category deletion rejected when activities reference it | `internal/categories/service.go` | Attempting to delete a category that has 1+ linked activities is rejected with HTTP 409 Conflict. |
| **BT-6** | Second active timer rejected | `internal/timer/service.go` | Attempting to start a timer while another is already active returns HTTP 409 Conflict with running timer details. |
| **BT-7** | Timer stop calculates correct duration | `internal/timer/service.go` | Stopping a timer computes exact difference: `duration = stopped_at - started_at` and creates activity. |
| **BT-8** | Dashboard calculates correct totals | `internal/dashboard/service.go` | Computes today's total, weekly total, category percentage distribution, and 7-day daily breakdown accurately from persisted activities. |

---

### 2.2 Frontend Behavior Tests

| Test # | Behavior | Target Frontend Module | Scenario & Expected Outcome |
|:---|:---|:---|:---|
| **FT-1** | Invalid activity form cannot submit | `web/static/js/app.js` | Form prevents submission when description or category is empty; shows field validation feedback. |
| **FT-2** | Timer controls reflect timer state | `web/static/js/app.js` | Displays `[ Start Timer ]` when idle; switches to live ticking counter and `[ Stop Timer ]` when active. |
| **FT-3** | Dashboard changes after activity creation | `web/static/js/app.js` | Creating or stopping an activity immediately triggers a dashboard metrics refresh. |
| **FT-4** | Filters update displayed activities | `web/static/js/app.js` | Changing category dropdown, typing in search box, or toggling favorite updates the activity list reactively. |

---

## 3. Running Tests

```bash
# Run all unit tests
make test
# Or:
go test -v -race ./...

# Run unit tests in a specific package
go test -v ./internal/activities/...
go test -v ./internal/timer/...
go test -v ./internal/dashboard/...
```
