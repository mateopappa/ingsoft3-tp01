# Flow — REST API Specification

## 1. Overview & Conventions

* **Base URL**: `/api`
* **Content-Type**: `application/json; charset=utf-8`
* **Authentication**: Stateful session cookie named `session_id` (`HttpOnly`, `SameSite=Lax`)
* **Standard Error Schema**:
```json
{
  "error": "Error message explanation",
  "code": "ERROR_CODE"
}
```

---

## 2. Authentication Endpoints

### 2.1 Register
* **Path**: `POST /api/auth/register`
* **Auth**: Public
* **Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```
* **Response (201 Created)**:
```json
{
  "user": {
    "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "email": "user@example.com",
    "created_at": "2026-08-19T14:00:00Z"
  }
}
```
* **Errors**: `400 Bad Request` (invalid email or password < 8 chars), `409 Conflict` (email already registered).

---

### 2.2 Login
* **Path**: `POST /api/auth/login`
* **Auth**: Public
* **Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```
* **Response (200 OK)**:
  * Sets cookie: `Set-Cookie: session_id=...; Path=/; HttpOnly; SameSite=Lax`
```json
{
  "user": {
    "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "email": "user@example.com"
  }
}
```
* **Errors**: `401 Unauthorized` (invalid credentials).

---

### 2.3 Logout
* **Path**: `POST /api/auth/logout`
* **Auth**: Required
* **Response (200 OK)**: Clears `session_id` cookie.
```json
{
  "message": "Logged out successfully"
}
```

---

### 2.4 Get Current User Profile
* **Path**: `GET /api/auth/me`
* **Auth**: Required
* **Response (200 OK)**:
```json
{
  "user": {
    "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "email": "user@example.com"
  }
}
```

---

## 3. Categories Endpoints

### 3.1 List Categories
* **Path**: `GET /api/categories`
* **Auth**: Required
* **Response (200 OK)**:
```json
{
  "categories": [
    {
      "id": "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
      "name": "Programming",
      "color": "#10B981",
      "created_at": "2026-08-19T14:00:00Z"
    }
  ]
}
```

---

### 3.2 Create Category
* **Path**: `POST /api/categories`
* **Auth**: Required
* **Request Body**:
```json
{
  "name": "Design",
  "color": "#EC4899"
}
```
* **Response (201 Created)**: Category object.
* **Errors**: `409 Conflict` (duplicate category name for this user).

---

### 3.3 Update Category
* **Path**: `PUT /api/categories/{id}`
* **Auth**: Required
* **Request Body**:
```json
{
  "name": "Product Design",
  "color": "#F43F5E"
}
```
* **Response (200 OK)**: Updated category object.

---

### 3.4 Delete Category
* **Path**: `DELETE /api/categories/{id}`
* **Auth**: Required
* **Response (204 No Content)**
* **Errors**: `409 Conflict` (category is referenced by activities and cannot be deleted).

---

## 4. Activities Endpoints

### 4.1 List Activities
* **Path**: `GET /api/activities`
* **Auth**: Required
* **Query Parameters**:
  * `search` (string): Text query matching description or note.
  * `category_id` (UUID): Filter by category.
  * `from_date` (YYYY-MM-DD): Start date inclusive.
  * `to_date` (YYYY-MM-DD): End date inclusive.
  * `favorite` (boolean): `true` to filter favorites only.
* **Response (200 OK)**:
```json
{
  "activities": [
    {
      "id": "9d78408a-2c8b-4ef2-9f33-6b3a2a4b5c6d",
      "description": "Working on thesis",
      "category_id": "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
      "category_name": "Programming",
      "category_color": "#10B981",
      "duration_seconds": 4800,
      "activity_date": "2026-08-19",
      "note": "Feature engineering for delay model",
      "favorite": true,
      "created_at": "2026-08-19T12:00:00Z",
      "updated_at": "2026-08-19T12:00:00Z"
    }
  ]
}
```

---

### 4.2 Create Activity (Manual)
* **Path**: `POST /api/activities`
* **Auth**: Required
* **Request Body**:
```json
{
  "description": "Working on thesis",
  "category_id": "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
  "duration_seconds": 4800,
  "activity_date": "2026-08-19",
  "note": "Feature engineering for delay model",
  "favorite": true
}
```
* **Response (201 Created)**: Created activity object.
* **Errors**: `400 Bad Request` (duration <= 0, invalid category, or invalid date).

---

### 4.3 Get Activity by ID
* **Path**: `GET /api/activities/{id}`
* **Auth**: Required
* **Response (200 OK)**: Activity object.

---

### 4.4 Update Activity
* **Path**: `PUT /api/activities/{id}`
* **Auth**: Required
* **Request Body**: Same as Create Activity.
* **Response (200 OK)**: Updated activity object.

---

### 4.5 Delete Activity
* **Path**: `DELETE /api/activities/{id}`
* **Auth**: Required
* **Response (204 No Content)**

---

### 4.6 Toggle Favorite
* **Path**: `PATCH /api/activities/{id}/favorite`
* **Auth**: Required
* **Request Body**:
```json
{
  "favorite": true
}
```
* **Response (200 OK)**: `{ "id": "...", "favorite": true }`

---

## 5. Persistent Timer Endpoints

### 5.1 Get Active Timer State
* **Path**: `GET /api/timer`
* **Auth**: Required
* **Response (200 OK - Active)**:
```json
{
  "active_timer": {
    "id": "e8a94b30-1c2d-4e5f-9a0b-1c2d3e4f5a6b",
    "description": "Programming — Go",
    "category_id": "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
    "category_name": "Programming",
    "category_color": "#10B981",
    "started_at": "2026-08-19T13:00:00Z",
    "elapsed_seconds": 2537
  }
}
```
* **Response (200 OK - Idle)**: `{ "active_timer": null }`

---

### 5.2 Start Timer
* **Path**: `POST /api/timer/start`
* **Auth**: Required
* **Request Body**:
```json
{
  "description": "Programming — Go",
  "category_id": "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d"
}
```
* **Response (201 Created)**: Active timer object.
* **Errors**: `409 Conflict` (user already has an active timer).

---

### 5.3 Stop Timer & Create Activity
* **Path**: `POST /api/timer/stop`
* **Auth**: Required
* **Request Body** (optional):
```json
{
  "note": "Optional note for session",
  "favorite": false
}
```
* **Response (200 OK)**: Newly generated Activity object.
* **Errors**: `404 Not Found` (no active timer found).

---

### 5.4 Discard Timer
* **Path**: `DELETE /api/timer/discard`
* **Auth**: Required
* **Response (204 No Content)**: Deletes active timer without creating activity.

---

## 6. Dashboard Analytics Endpoint

### 6.1 Get Dashboard Data
* **Path**: `GET /api/dashboard`
* **Auth**: Required
* **Query Parameters**:
  * `tz_offset` (integer, optional): Client timezone offset in minutes.
* **Response (200 OK)**:
```json
{
  "today_seconds": 20520,
  "today_formatted": "5h 42m",
  "week_seconds": 98100,
  "week_formatted": "27h 15m",
  "category_distribution": [
    {
      "category_id": "a1b2c3d4-e5f6-7a8b-9c0d-1e2f3a4b5c6d",
      "name": "Programming",
      "color": "#10B981",
      "seconds": 41202,
      "percentage": 42
    },
    {
      "category_id": "b2c3d4e5-f6a7-8b9c-0d1e-2f3a4b5c6d7e",
      "name": "Work",
      "color": "#3B82F6",
      "seconds": 30411,
      "percentage": 31
    }
  ],
  "daily_activity": [
    { "day": "Mon", "date": "2026-08-17", "seconds": 18000 },
    { "day": "Tue", "date": "2026-08-18", "seconds": 25200 },
    { "day": "Wed", "date": "2026-08-19", "seconds": 20520 },
    { "day": "Thu", "date": "2026-08-20", "seconds": 0 },
    { "day": "Fri", "date": "2026-08-21", "seconds": 0 },
    { "day": "Sat", "date": "2026-08-22", "seconds": 0 },
    { "day": "Sun", "date": "2026-08-23", "seconds": 0 }
  ]
}
```

---

## 7. System Health Endpoint

### 7.1 Health Check
* **Path**: `GET /api/health`
* **Auth**: Public
* **Response (200 OK)**:
```json
{
  "status": "healthy",
  "database": "connected"
}
```
