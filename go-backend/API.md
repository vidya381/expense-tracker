# MySpendo API Documentation

Base URL: `http://localhost:8080` (configurable via `constants.DefaultServerPort`)

## Authentication

All endpoints except `/register` and `/login` require JWT authentication via the `Authorization` header:
```
Authorization: Bearer <token>
```

---

## 1. Authentication Endpoints

### 1.1 Register User
**POST** `/register`

Create a new user account.

**Request (form-data):**
```
username: string (3-50 chars, alphanumeric/underscore/hyphen only)
email: string (valid email format)
password: string (min 8 chars)
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "User registered successfully!"
}
```

**Response (409 Conflict):**
```json
{
  "success": false,
  "error": "This email is already registered."
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/register \
  -F "username=johndoe" \
  -F "email=john@example.com" \
  -F "password=password123"
```

---

### 1.2 Login User
**POST** `/login`

Authenticate and receive JWT token.

**Request (form-data):**
```
email: string
password: string
```

**Response (200 OK):**
```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response (401 Unauthorized):**
```json
{
  "success": false,
  "error": "Email or password is incorrect."
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/login \
  -F "email=john@example.com" \
  -F "password=password123"
```

---

## 2. Category Endpoints

### 2.1 Add Category
**POST** `/category/add`

**Authentication:** Required

**Request (form-data):**
```
name: string (max 100 chars)
type: string ("expense" or "income")
```

**Response (200 OK):**
```json
{
  "success": true,
  "category": {
    "id": 1,
    "name": "Groceries",
    "type": "expense"
  }
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/category/add \
  -H "Authorization: Bearer <token>" \
  -F "name=Groceries" \
  -F "type=expense"
```

---

### 2.2 List Categories
**GET** `/category/list`

**Authentication:** Required

**Response (200 OK):**
```json
{
  "success": true,
  "categories": [
    {
      "id": 1,
      "name": "Groceries",
      "type": "expense",
      "user_id": 1,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

**Example:**
```bash
curl -X GET http://localhost:8080/category/list \
  -H "Authorization: Bearer <token>"
```

---

## 3. Transaction Endpoints

### 3.1 Add Transaction
**POST** `/transaction/add`

**Authentication:** Required

**Request (form-data):**
```
category_id: integer (positive number)
amount: float (positive, max 2 decimal places)
description: string (optional)
date: string (format: YYYY-MM-DD)
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Transaction added successfully"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/transaction/add \
  -H "Authorization: Bearer <token>" \
  -F "category_id=1" \
  -F "amount=45.99" \
  -F "description=Weekly groceries" \
  -F "date=2024-01-15"
```

---

### 3.2 List Transactions
**GET** `/transaction/list`

**Authentication:** Required

**Response (200 OK):**
```json
{
  "success": true,
  "transactions": [
    {
      "id": 1,
      "user_id": 1,
      "category_id": 1,
      "category": "Groceries",
      "category_type": "expense",
      "amount": 45.99,
      "description": "Weekly groceries",
      "date": "2024-01-15",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

**Example:**
```bash
curl -X GET http://localhost:8080/transaction/list \
  -H "Authorization: Bearer <token>"
```

---

### 3.3 Update Transaction
**POST** `/transaction/update`

**Authentication:** Required

**Request (form-data):**
```
id: integer (transaction ID)
category_id: integer
amount: float
description: string
date: string (format: YYYY-MM-DD)
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Transaction updated successfully"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/transaction/update \
  -H "Authorization: Bearer <token>" \
  -F "id=1" \
  -F "category_id=1" \
  -F "amount=50.00" \
  -F "description=Updated groceries" \
  -F "date=2024-01-15"
```

---

### 3.4 Delete Transaction
**POST** `/transaction/delete`

**Authentication:** Required

**Request (form-data):**
```
id: integer (transaction ID)
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Transaction deleted"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/transaction/delete \
  -H "Authorization: Bearer <token>" \
  -F "id=1"
```

---

### 3.5 Search and Filter Transactions
**GET** `/transactions/search`

**Authentication:** Required

**Query Parameters:**
```
q: string (keyword search in description)
category_id: integer (filter by category)
from: string (date format: YYYY-MM-DD)
to: string (date format: YYYY-MM-DD)
min_amount: float
max_amount: float
sort: string (date_asc, date_desc, amount_asc, amount_desc)
limit: integer (default: 20, max: 1000)
offset: integer (default: 0)
```

**Response (200 OK):**
```json
{
  "success": true,
  "transactions": [...],
  "limit": 20,
  "offset": 0
}
```

**Example:**
```bash
curl -X GET "http://localhost:8080/transactions/search?category_id=1&from=2024-01-01&to=2024-01-31&sort=date_desc&limit=10&offset=0" \
  -H "Authorization: Bearer <token>"
```

---

## 4. Summary Endpoints

### 4.1 Overall Totals
**GET** `/summary/totals`

**Authentication:** Required

**Response (200 OK):**
```json
{
  "total_expenses": 1250.50,
  "total_income": 3000.00
}
```

**Example:**
```bash
curl -X GET http://localhost:8080/summary/totals \
  -H "Authorization: Bearer <token>"
```

---

### 4.2 Monthly Totals
**GET** `/summary/monthly`

**Authentication:** Required

**Response (200 OK):**
```json
[
  {
    "month": "2024-01-01T00:00:00Z",
    "total_expenses": 450.00,
    "total_income": 3000.00
  }
]
```

**Example:**
```bash
curl -X GET http://localhost:8080/summary/monthly \
  -H "Authorization: Bearer <token>"
```

---

### 4.3 Category Breakdown
**GET** `/summary/category`

**Authentication:** Required

**Query Parameters:**
```
from: string (date format: YYYY-MM-DD, optional)
to: string (date format: YYYY-MM-DD, optional)
```

**Response (200 OK):**
```json
[
  {
    "category_id": 1,
    "category_name": "Groceries",
    "total": 450.50
  }
]
```

**Example:**
```bash
curl -X GET "http://localhost:8080/summary/category?from=2024-01-01&to=2024-01-31" \
  -H "Authorization: Bearer <token>"
```

---

### 4.4 Group Totals by Period
**GET** `/summary/group`

**Authentication:** Required

**Query Parameters:**
```
by: string ("day", "week", "month", "year")
```

**Response (200 OK):**
```json
[
  {
    "period": "2024-01",
    "total_expenses": 450.00,
    "total_income": 3000.00
  }
]
```

**Example:**
```bash
curl -X GET "http://localhost:8080/summary/group?by=month" \
  -H "Authorization: Bearer <token>"
```

---

### 4.5 Category Summary for Month
**GET** `/summary/category/monthly`

**Authentication:** Required

**Query Parameters:**
```
year: integer (required)
month: integer (1-12, required)
```

**Response (200 OK):**
```json
[
  {
    "category_id": 1,
    "category_name": "Groceries",
    "total": 450.50
  }
]
```

**Example:**
```bash
curl -X GET "http://localhost:8080/summary/category/monthly?year=2024&month=1" \
  -H "Authorization: Bearer <token>"
```

---

## 5. Export Endpoint

### 5.1 Export Transactions
**GET** `/export`

**Authentication:** Required

**Query Parameters:**
```
format: string ("csv" or "json", default: "json")
```

**Response (200 OK - CSV):**
```csv
ID,CategoryID,Amount,Description,Date,CreatedAt
1,1,45.99,Weekly groceries,2024-01-15,2024-01-15T10:30:00Z
```

**Response (200 OK - JSON):**
```json
[
  {
    "id": 1,
    "user_id": 1,
    "category_id": 1,
    "category": "Groceries",
    "category_type": "expense",
    "amount": 45.99,
    "description": "Weekly groceries",
    "date": "2024-01-15",
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

**Example:**
```bash
curl -X GET "http://localhost:8080/export?format=csv" \
  -H "Authorization: Bearer <token>" \
  -o transactions.csv
```

---

## 6. Recurring Transaction Endpoints

### 6.1 Add Recurring Transaction
**POST** `/recurring/add`

**Authentication:** Required

**Request (form-data):**
```
category_id: integer (positive number)
amount: float (positive)
description: string (optional)
start_date: string (format: YYYY-MM-DD)
recurrence: string ("daily", "weekly", "monthly", "yearly")
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Recurring transaction added successfully"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/recurring/add \
  -H "Authorization: Bearer <token>" \
  -F "category_id=1" \
  -F "amount=100.00" \
  -F "description=Monthly rent" \
  -F "start_date=2024-01-01" \
  -F "recurrence=monthly"
```

---

### 6.2 List Recurring Transactions
**GET** `/recurring/list`

**Authentication:** Required

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "user_id": 1,
    "category_id": 1,
    "amount": 100.00,
    "description": "Monthly rent",
    "start_date": "2024-01-01",
    "recurrence": "monthly",
    "last_occurrence": "2024-01-01T00:00:00Z"
  }
]
```

**Example:**
```bash
curl -X GET http://localhost:8080/recurring/list \
  -H "Authorization: Bearer <token>"
```

---

### 6.3 Edit Recurring Transaction
**POST** `/recurring/edit`

**Authentication:** Required

**Request (form-data):**
```
id: integer (recurring transaction ID)
amount: float (positive)
description: string
start_date: string (format: YYYY-MM-DD)
recurrence: string ("daily", "weekly", "monthly", "yearly")
```

**Response (200 OK):**
```
Recurring transaction updated!
```

**Example:**
```bash
curl -X POST http://localhost:8080/recurring/edit \
  -H "Authorization: Bearer <token>" \
  -F "id=1" \
  -F "amount=120.00" \
  -F "description=Updated rent" \
  -F "start_date=2024-01-01" \
  -F "recurrence=monthly"
```

---

### 6.4 Delete Recurring Transaction
**POST** `/recurring/delete`

**Authentication:** Required

**Request (form-data):**
```
id: integer (recurring transaction ID)
```

**Response (200 OK):**
```
Recurring transaction deleted!
```

**Example:**
```bash
curl -X POST http://localhost:8080/recurring/delete \
  -H "Authorization: Bearer <token>" \
  -F "id=1"
```

---

## 7. Budget Endpoints

### 7.1 Add Budget
**POST** `/budget/add`

**Authentication:** Required

**Request (form-data):**
```
category_id: integer (0 for overall budget, >0 for category-specific)
amount: float (positive)
period: string ("monthly" or "yearly", default: "monthly")
alert_threshold: integer (0-100, default: 80)
```

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Budget added successfully"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/budget/add \
  -H "Authorization: Bearer <token>" \
  -F "category_id=1" \
  -F "amount=500.00" \
  -F "period=monthly" \
  -F "alert_threshold=80"
```

---

### 7.2 List Budgets
**GET** `/budget/list`

**Authentication:** Required

**Response (200 OK):**
```json
{
  "success": true,
  "budgets": [
    {
      "id": 1,
      "user_id": 1,
      "category_id": 1,
      "amount": 500.00,
      "period": "monthly",
      "alert_threshold": 80,
      "created_at": "2024-01-01T10:00:00Z"
    }
  ]
}
```

**Example:**
```bash
curl -X GET http://localhost:8080/budget/list \
  -H "Authorization: Bearer <token>"
```

---

### 7.3 Update Budget
**POST** `/budget/update`

**Authentication:** Required

**Request (form-data):**
```
id: integer (budget ID)
amount: float (positive)
alert_threshold: integer (0-100)
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Budget updated successfully"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/budget/update \
  -H "Authorization: Bearer <token>" \
  -F "id=1" \
  -F "amount=600.00" \
  -F "alert_threshold=85"
```

---

### 7.4 Delete Budget
**POST** `/budget/delete`

**Authentication:** Required

**Request (form-data):**
```
id: integer (budget ID)
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Budget deleted successfully"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/budget/delete \
  -H "Authorization: Bearer <token>" \
  -F "id=1"
```

---

### 7.5 Budget Alerts
**GET** `/budget/alerts`

**Authentication:** Required

Returns budgets where spending has exceeded the alert threshold.

**Response (200 OK):**
```json
[
  {
    "budget_id": 1,
    "category_id": 1,
    "category_name": "Groceries",
    "budget_amount": 500.00,
    "spent_amount": 450.00,
    "percentage_used": 90,
    "alert_threshold": 80,
    "period": "monthly"
  }
]
```

**Example:**
```bash
curl -X GET http://localhost:8080/budget/alerts \
  -H "Authorization: Bearer <token>"
```

---

## Error Responses

All endpoints may return the following error responses:

**400 Bad Request:**
```json
{
  "success": false,
  "error": "Validation error message"
}
```

**401 Unauthorized:**
```json
{
  "success": false,
  "error": "User not authenticated"
}
```

**405 Method Not Allowed:**
```json
{
  "success": false,
  "error": "Only POST allowed"
}
```

**500 Internal Server Error:**
```json
{
  "success": false,
  "error": "Internal server error message"
}
```

---

## Rate Limiting

Authentication endpoints (`/register` and `/login`) are rate-limited to prevent brute-force attacks:
- Default: 5 requests per minute per IP
- Burst: 5 requests (configurable via `constants.AuthRateLimitPerMinute` and `constants.AuthRateLimitBurst`)

---

## CORS Configuration

CORS is enabled with the following settings:
- Allowed Origins: Configurable via `CORS_ORIGIN` environment variable (default: `http://localhost:3000`)
- Allowed Methods: GET, POST, PUT, DELETE, OPTIONS
- Allowed Headers: Authorization, Content-Type
- Credentials: Allowed

---

## Database Connection Pool

Connection pool configuration:
- Max Open Connections: 25 (configurable via `constants.MaxOpenConnections`)
- Max Idle Connections: 10 (configurable via `constants.MaxIdleConnections`)
- Connection Max Lifetime: 5 minutes (configurable via `constants.ConnectionMaxLifetime`)
- Connection Max Idle Time: 2 minutes (configurable via `constants.ConnectionMaxIdleTime`)

---

## Background Jobs

### Recurring Transaction Processor

Runs every hour to process recurring transactions and create actual transactions based on schedules.

- Uses PostgreSQL advisory locks to prevent concurrent processing
- Handles daily, weekly, monthly, and yearly recurrences
- Automatically handles month-end edge cases (e.g., Jan 31 → Feb 28/29)
- Catches up on missed occurrences
