# Flow — Product Requirements Specification

## 1. Product Vision

**Flow** is a minimalist, production-oriented personal activity and time tracking web application.

The application allows an individual to track how they spend their time across work, study, programming, training, and personal activities, and immediately understand where their time is going through a concise dashboard.

### Core Philosophy
* **Simplicity over Complexity**: No projects, tasks, subtasks, goals, calendars, notifications, social feeds, AI features, integrations, or gamification.
* **Useful in Real Life**: Provides instant one-click time tracking, robust manual logging, flexible category management, and clean weekly analytics.
* **Engineering Foundation**: Built as a clean, testable monolithic Go + PostgreSQL application designed to serve as the baseline for a semester-long DevOps curriculum.

---

## 2. Primary Product Features

### 2.1 Authentication & User Isolation
* **User Registration**: Register with email and password.
* **Login & Logout**: Secure session-based authentication using HTTP-only, SameSite cookies.
* **Multi-User Data Isolation**: Every entity belongs strictly to the authenticated user. No user can view, edit, or delete another user's categories, activities, or active timers.

### 2.2 Category Management
* **User-Managed Categories**: Users can create, rename, delete, and list categories (e.g. *Work*, *Study*, *Programming*, *Training*, *Personal*).
* **Duplicate Prevention**: A user cannot create duplicate category names.
* **Referential Protection**: A category that is referenced by one or more activities cannot be deleted. Deletion is rejected with a clear error explanation.

### 2.3 Persistent Time Tracking & Single-Timer Invariant
* **Live Timer**: Start a timer with a description and category.
* **Persistent State**: Timer state is persisted on the server in PostgreSQL (`active_timers`). The timer survives browser tab closure, browser restarts, and device switching.
* **Server-Side Truth**: The server calculates elapsed duration dynamically as `NOW() - started_at`. JavaScript in the browser is never the source of truth for duration.
* **Single Active Timer Rule**: A user may have at most **one** active timer at any time. Starting a second timer while one is running is rejected by both application logic and database unique constraints.
* **Atomic Stop & Conversion**: Stopping the timer calculates elapsed seconds, creates an activity record, and removes the active timer within an atomic database transaction.

### 2.4 Manual Activity Creation & History
* **Manual Time Entry**: Direct creation of completed activities with description, category, duration (stored as integer seconds), and activity date (`YYYY-MM-DD`).
* **Optional Fields**: Support for detailed notes and marking activities as favorites.
* **Activity History View**: Chronological view with real-time keyword search (matching description and notes), category filtering, date range filtering, and favorite filtering.
* **Activity Operations**: Edit duration, description, category, date, notes, and favorite status, or permanently delete an activity.

### 2.5 Dashboard & Weekly Analytics
* **Quick Timer Widget**: Prominent component on the dashboard to quickly start or stop an activity timer in seconds.
* **Today Total**: Aggregated duration of all activities completed today (e.g. `5h 42m`).
* **This Week Total**: Aggregated duration for the current week from Monday through Sunday (e.g. `27h 15m`).
* **Time by Category**: Visual percentage and total hour breakdown across categories for the current week.
* **Activity by Day**: Daily bar visualization displaying tracked time for each day of the current week (Mon–Sun).

---

## 3. Explicit Business Rules

| Rule ID | Rule Name | Specification |
|:---|:---|:---|
| **Rule 1** | Valid Duration | An activity must have `duration_seconds > 0`. Zero and negative durations are strictly rejected. |
| **Rule 2** | Valid Category | An activity must reference an existing category owned by the same authenticated user. |
| **Rule 3** | User Ownership | All operations strictly enforce tenant isolation (`WHERE user_id = $1`). Accessing another user's records is forbidden. |
| **Rule 4** | Category Protection | A category referenced by one or more activities cannot be deleted. The system rejects deletion with HTTP 409 Conflict. |
| **Rule 5** | Single Active Timer | A user cannot have more than one active timer simultaneously. Starting a second timer is rejected with HTTP 409 Conflict. |
| **Rule 6** | Server Duration Calculation | Stopping a timer calculates duration strictly from server timestamps: `duration = stopped_at - started_at`. |
| **Rule 7** | Dashboard Aggregations | Dashboard totals are computed from persisted activities in PostgreSQL for the user's current date and week. |
| **Rule 8** | Valid Activity Date | Activities must have a valid date (`YYYY-MM-DD`) and cannot be recorded with invalid date formats. |

---

## 4. Frontend Behaviors

| Behavior ID | Description |
|:---|:---|
| **Behavior 1** | The activity form validates client-side and prevents submission when required data (description, category, duration) is missing or invalid. |
| **Behavior 2** | The timer controls reactively reflect actual server state: shows `[ Start Timer ]` when idle, and active elapsed counter with `[ Stop Timer ]` when running. |
| **Behavior 3** | After an activity is created, edited, or stopped, the dashboard totals (Today, This Week, Category Distribution, Daily Chart) update immediately. |
| **Behavior 4** | Search keyword input, category filter dropdown, date filter, and favorite toggle dynamically filter the activity list without full page reloads. |

---

## 5. Scope & Non-Goals

### Explicit Non-Goals (Out of Scope)
* ❌ Docker / Kubernetes dependencies in Stage 0 (local Go + PostgreSQL only).
* ❌ Heavy frontend frameworks (React, Next.js, Vue, Angular, Svelte).
* ❌ Projects, subtasks, task lists, goals, milestones, or priorities.
* ❌ Native mobile apps (iOS / Android) or desktop apps (Electron).
* ❌ Calendars, scheduling, recurring reminders, or browser notifications.
* ❌ AI, machine learning, time forecasting, or productivity scoring.
* ❌ External API integrations, third-party sync (Google Calendar, Toggl, Notion).
* ❌ Social features, team workspaces, or collaboration.
* ❌ Payment gateways, subscriptions, or billing logic.
