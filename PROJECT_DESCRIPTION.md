# Expense Tracker - Comprehensive Project Description

## Table of Contents
- [Overview](#overview)
- [Project Architecture](#project-architecture)
- [Features Implemented](#features-implemented)
- [Technologies & Frameworks](#technologies--frameworks)
- [Database Schema](#database-schema)
- [API Endpoints](#api-endpoints)
- [Authentication & Security](#authentication--security)
- [Frontend Implementation](#frontend-implementation)
- [Configuration](#configuration)
- [Background Jobs](#background-jobs)
- [Testing](#testing)
- [Deployment](#deployment)
- [Project Statistics](#project-statistics)

---

## Overview

This is a **full-stack personal finance management application** built with Go and Next.js. It enables users to track income and expenses, manage budgets, set up recurring transactions, and visualize their financial data through comprehensive analytics and reporting features.

### Project Structure
```
expense-tracker/
├── go-backend/              # Go REST API server
│   ├── main.go             # Entry point, HTTP server, routes
│   ├── db.go               # Database connection utilities
│   ├── migrate.go          # Database migration runner
│   ├── constants/          # Application constants
│   ├── models/             # Data models
│   ├── handlers/           # Business logic
│   ├── middleware/         # Auth, rate limiting, security
│   ├── jobs/               # Background jobs
│   ├── migrations/         # SQL schema files
│   └── utils/              # Helper functions
├── nextjs-frontend/         # Next.js 15 React frontend
│   └── src/
│       ├── app/            # Next.js app router pages
│       ├── components/     # Reusable React components
│       ├── context/        # React context (AuthContext)
│       └── lib/            # API utilities
├── README.md
├── DEPLOYMENT.md
└── LICENSE (MIT)
```

---

## Project Architecture

### Backend Architecture (Go)

**Clean Architecture Pattern:**
- **Models Layer:** Data structures and business entities
- **Handlers Layer:** HTTP request handlers and business logic
- **Middleware Layer:** Cross-cutting concerns (auth, rate limiting, security)
- **Utils Layer:** Shared utilities (validation, sanitization, logging)
- **Jobs Layer:** Background processing tasks

**Key Design Patterns:**
- RESTful API design
- Context-based operations with timeouts
- Dependency injection for database connections
- Structured logging with slog
- Error handling with consistent response formats

### Frontend Architecture (Next.js)

**Modern React Patterns:**
- Next.js 15 App Router for file-based routing
- React Context API for global state (authentication)
- Custom hooks for API integration
- Component composition for reusability
- TypeScript for type safety

**Page Structure:**
- Authentication flow (login/register)
- Protected dashboard routes
- Transaction management interfaces
- Budget and recurring transaction management
- Analytics and reporting views

---

## Features Implemented

### 1. User Authentication & Authorization

**Registration:**
- Username validation (3-50 chars, alphanumeric + underscore/hyphen)
- Email format validation
- Password requirements (minimum 8 characters)
- Bcrypt password hashing
- Automatic account creation with timestamps

**Login:**
- JWT token generation with 72-hour expiry
- Token contains user_id claim
- Bearer token authentication for protected routes
- Secure token storage in localStorage (frontend)
- Automatic logout on 401 responses

**Security:**
- Password never stored in plain text
- Tokens expire after 3 days
- Per-user data isolation
- Session management on frontend

### 2. Transaction Management

**Core Operations:**
- **Add Transaction:** Create income or expense entries
- **Edit Transaction:** Update existing transaction details
- **Delete Transaction:** Remove transactions (cascade safe)
- **List Transactions:** View all user transactions with details

**Advanced Features:**
- **Search & Filter:**
  - Filter by category ID
  - Date range filtering (from/to dates)
  - Amount range filtering (min/max)
  - Keyword search in descriptions
  - Sort by date or amount (ascending/descending)
  - Pagination support (limit/offset)
  - Maximum 1000 records per page

**Data Fields:**
- Amount (decimal, positive, max 1 billion)
- Description (max 500 characters)
- Date (YYYY-MM-DD format)
- Category assignment
- Automatic timestamps

### 3. Category Management

**Features:**
- Create custom categories per user
- Separate categories for income and expense
- Category name validation (max 100 chars)
- Unique constraint per user
- List all user categories with filtering

**Use Cases:**
- Organize transactions by purpose
- Track spending patterns by category
- Enable category-based budgeting
- Generate category-wise reports

### 4. Recurring Transactions

**Recurrence Patterns:**
- **Daily:** Every day from start date
- **Weekly:** Every 7 days
- **Monthly:** Same day each month
- **Yearly:** Same date each year

**Automation:**
- Background job runs every hour
- Automatically creates transactions for due dates
- Updates `last_occurrence` timestamp
- Catches up on multiple missed occurrences
- PostgreSQL advisory locks prevent duplicate processing

**Edge Case Handling:**
- **Month-end dates:** Jan 31 → Feb 28/29 (uses last day of month)
- **Leap years:** Yearly recurrence on Feb 29 → Feb 28 on non-leap years
- **Multiple missed dates:** Processes all past-due occurrences in sequence
- **Infinite loop prevention:** Maximum 3650 iterations per rule
- **Date normalization:** All dates normalized to UTC midnight

**Operations:**
- Add recurring transaction rule
- List all recurring rules
- Edit existing rule
- Delete rule (stops future occurrences)

### 5. Budget Management

**Budget Types:**
- **Category Budget:** Limit spending for specific category
- **Overall Budget:** Total spending limit (category_id = 0)

**Budget Periods:**
- **Monthly:** Budget resets each month
- **Yearly:** Budget resets each year

**Alert System:**
- Configurable alert threshold (0-100%)
- Budget alerts API endpoint
- Returns budgets exceeding threshold
- Includes current spending amount and percentage

**Operations:**
- Create budget with amount and threshold
- List all budgets with current spending
- Update budget amount or threshold
- Delete budget

**Budget Tracking:**
- Real-time spending calculation
- Percentage used vs. budget amount
- Alert when threshold exceeded
- Historical budget data retained

### 6. Financial Analytics & Reports

**Summary Endpoints:**

**1. Overall Totals (`/summary/totals`)**
- Total income across all time
- Total expenses across all time
- Net balance calculation

**2. Monthly Summary (`/summary/monthly`)**
- Income and expenses per month
- Historical data for trend analysis
- Net income per month
- Ordered chronologically

**3. Current Month Summary (`/summary/current-month`)**
- Current month income and expenses
- Projected recurring expenses
- Shows upcoming automated transactions
- Helps predict month-end balance

**4. Category Breakdown (`/summary/category`)**
- Spending per category
- Date range filtering (from/to)
- Income vs. expense separation
- Total amount per category

**5. Time-Period Grouping (`/summary/group`)**
- Group by: day, week, month, or year
- Aggregated income and expenses
- Time series data for charts
- Flexible date range

**6. Category Monthly Analysis (`/summary/category/monthly`)**
- Category spending for specific month
- Year and month parameters
- Detailed breakdown per category
- Helps identify spending patterns

### 7. Data Export

**Formats:**
- **CSV:** Comma-separated values for Excel/Sheets
- **JSON:** Structured data for programmatic use

**Export Fields:**
- Transaction ID
- Amount
- Description
- Date
- Category name
- Category type (income/expense)
- Created timestamp

**Use Cases:**
- Backup transaction data
- Import into other tools
- Data analysis in spreadsheets
- Tax preparation

---

## Technologies & Frameworks

### Backend Stack

#### Core Technologies
| Technology | Version | Purpose |
|------------|---------|---------|
| **Go** | 1.24 | Backend language |
| **PostgreSQL** | Latest | Primary database |
| **pgx/v5** | 5.x | PostgreSQL driver |

#### Key Libraries

**Authentication & Security:**
```go
github.com/golang-jwt/jwt/v5           // JWT token generation/validation
golang.org/x/crypto/bcrypt             // Password hashing
```

**HTTP & Middleware:**
```go
github.com/rs/cors                     // CORS middleware
golang.org/x/time/rate                 // Rate limiting
```

**Utilities:**
```go
github.com/joho/godotenv               // Environment variables
encoding/json                          // JSON encoding/decoding
log/slog                               // Structured logging
```

**Database Features:**
- Connection pooling (25 max open, 10 idle)
- SSL/TLS support for production
- Context-based timeouts (10 seconds default)
- Prepared statements for all queries
- Advisory locks for job coordination

#### Architecture Patterns
- RESTful API design
- Clean architecture (separation of concerns)
- Middleware pipeline pattern
- Repository pattern for data access
- Structured logging with context

### Frontend Stack

#### Core Framework
| Technology | Version | Purpose |
|------------|---------|---------|
| **Next.js** | 15.x | React framework with App Router |
| **React** | 19.x | UI library |
| **TypeScript** | 5.x | Type safety |

#### Styling & UI
```json
"tailwindcss": "^4.0.0"              // Utility-first CSS
"react-icons": "^5.5.0"              // Icon components
"@fontsource-variable/geist-sans"    // Google fonts
"@fontsource-variable/geist-mono"    // Monospace font
```

#### Data Visualization
```json
"recharts": "^3.1.0"                 // Charts and graphs
"victory-vendor": "^37.3.2"          // D3.js for visualizations
```

#### Utilities
```json
"date-fns": "^4.1.0"                 // Date manipulation
```

#### Development Tools
```json
"eslint": "^9"                       // Code linting
"typescript": "^5"                   // Type checking
```

#### State Management
- React Context API (AuthContext)
- React hooks (useState, useEffect, useCallback)
- localStorage for token persistence

#### API Integration
- Custom `apiFetch()` function with JWT
- `useApi()` hook for authenticated calls
- Automatic token injection
- Error handling with auto-logout on 401

---

## Database Schema

### Tables Overview

#### 1. users
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
```

**Purpose:** Store user accounts with authentication credentials

**Fields:**
- `id`: Auto-incrementing primary key
- `username`: Unique username (3-50 chars)
- `email`: Unique email address
- `password`: Bcrypt hashed password (never plain text)
- `created_at`: Account creation timestamp

#### 2. categories
```sql
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(10) NOT NULL CHECK (type IN ('income', 'expense')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, name, type)
);
CREATE INDEX idx_categories_user ON categories(user_id);
```

**Purpose:** User-defined transaction categories

**Fields:**
- `id`: Auto-incrementing primary key
- `user_id`: Foreign key to users (cascades on delete)
- `name`: Category name (max 100 chars)
- `type`: Either 'income' or 'expense'
- `created_at`: Creation timestamp

**Constraints:**
- Unique combination of user_id, name, and type
- Prevents duplicate category names per user per type

#### 3. transactions
```sql
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    description TEXT,
    date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_transactions_user ON transactions(user_id);
CREATE INDEX idx_transactions_category ON transactions(category_id);
CREATE INDEX idx_transactions_date ON transactions(date);
```

**Purpose:** Store all financial transactions

**Fields:**
- `id`: Auto-incrementing primary key
- `user_id`: Foreign key to users (cascades on delete)
- `category_id`: Foreign key to categories (cascades on delete)
- `amount`: Decimal amount (positive, up to 99,999,999.99)
- `description`: Optional text description (max 500 chars validated in app)
- `date`: Transaction date (YYYY-MM-DD)
- `created_at`: Record creation timestamp

**Indexes:**
- User ID for fast user-based queries
- Category ID for category-based filtering
- Date for date-range queries and sorting

#### 4. recurring_transactions
```sql
CREATE TABLE recurring_transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    description TEXT,
    start_date DATE NOT NULL,
    recurrence VARCHAR(10) NOT NULL CHECK (recurrence IN ('daily', 'weekly', 'monthly', 'yearly')),
    last_occurrence TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_recurring_user ON recurring_transactions(user_id);
CREATE INDEX idx_recurring_category ON recurring_transactions(category_id);
```

**Purpose:** Define rules for automated recurring transactions

**Fields:**
- `id`: Auto-incrementing primary key
- `user_id`: Foreign key to users (cascades on delete)
- `category_id`: Foreign key to categories (cascades on delete)
- `amount`: Decimal amount for each occurrence
- `description`: Description for generated transactions
- `start_date`: First occurrence date
- `recurrence`: Frequency (daily/weekly/monthly/yearly)
- `last_occurrence`: Timestamp of last processed occurrence
- `created_at`: Rule creation timestamp

**Processing:**
- Background job checks every hour
- Creates actual transactions based on rules
- Updates last_occurrence after processing

#### 5. budgets
```sql
CREATE TABLE budgets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INTEGER DEFAULT 0,
    amount DECIMAL(10, 2) NOT NULL CHECK (amount > 0),
    period VARCHAR(10) NOT NULL CHECK (period IN ('monthly', 'yearly')),
    alert_threshold INTEGER DEFAULT 80 CHECK (alert_threshold >= 0 AND alert_threshold <= 100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, category_id, period)
);
CREATE INDEX idx_budgets_user ON budgets(user_id);
CREATE INDEX idx_budgets_category ON budgets(category_id);
```

**Purpose:** Track spending limits and alerts

**Fields:**
- `id`: Auto-incrementing primary key
- `user_id`: Foreign key to users (cascades on delete)
- `category_id`: Category to budget (0 = overall budget)
- `amount`: Budget limit amount
- `period`: Time period (monthly/yearly)
- `alert_threshold`: Alert percentage (0-100)
- `created_at`: Budget creation timestamp

**Constraints:**
- Unique combination of user_id, category_id, and period
- Prevents multiple budgets for same category/period

**Special Cases:**
- `category_id = 0`: Overall spending budget (all categories)
- Alert triggers when spending exceeds threshold percentage

### Database Configuration

**Connection Pooling:**
- Max open connections: 25
- Max idle connections: 10
- Connection max lifetime: 5 minutes
- Connection max idle time: 2 minutes

**Timeouts:**
- Default query timeout: 10 seconds
- Context-based cancellation support

**Security:**
- SSL/TLS support (configurable via sslmode)
- Prepared statements prevent SQL injection
- Row-level security via user_id filtering

---

## API Endpoints

### Public Endpoints (No Authentication Required)

#### POST /register
**Purpose:** Create a new user account

**Request Body:**
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "securepass123"
}
```

**Response (201):**
```json
{
  "message": "User registered successfully"
}
```

**Validation:**
- Username: 3-50 chars, alphanumeric + underscore/hyphen
- Email: Valid email format
- Password: Minimum 8 characters

---

#### POST /login
**Purpose:** Authenticate user and receive JWT token

**Request Body:**
```json
{
  "username": "john_doe",
  "password": "securepass123"
}
```

**Response (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Login successful"
}
```

**Token Details:**
- Expires in 72 hours
- Contains user_id claim
- Use in Authorization header: `Bearer <token>`

---

### Protected Endpoints (Require JWT Token)

**Authentication Header:**
```
Authorization: Bearer <jwt_token>
```

### Category Endpoints

#### POST /category/add
**Purpose:** Create a new category

**Request Body:**
```json
{
  "name": "Groceries",
  "type": "expense"
}
```

**Response (201):**
```json
{
  "id": 1,
  "name": "Groceries",
  "type": "expense",
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

#### GET /category/list
**Purpose:** List all user categories

**Query Parameters:**
- `type` (optional): Filter by "income" or "expense"

**Response (200):**
```json
{
  "categories": [
    {
      "id": 1,
      "name": "Groceries",
      "type": "expense",
      "created_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "name": "Salary",
      "type": "income",
      "created_at": "2024-01-15T10:35:00Z"
    }
  ]
}
```

---

### Transaction Endpoints

#### POST /transaction/add
**Purpose:** Create a new transaction

**Request Body:**
```json
{
  "category_id": 1,
  "amount": 150.50,
  "description": "Weekly grocery shopping",
  "date": "2024-01-15"
}
```

**Response (201):**
```json
{
  "id": 1,
  "message": "Transaction added successfully"
}
```

---

#### GET /transaction/list
**Purpose:** List all user transactions

**Response (200):**
```json
{
  "transactions": [
    {
      "id": 1,
      "category_id": 1,
      "category_name": "Groceries",
      "category_type": "expense",
      "amount": 150.50,
      "description": "Weekly grocery shopping",
      "date": "2024-01-15",
      "created_at": "2024-01-15T14:20:00Z"
    }
  ]
}
```

---

#### POST /transaction/update
**Purpose:** Update existing transaction

**Request Body:**
```json
{
  "id": 1,
  "category_id": 1,
  "amount": 175.00,
  "description": "Updated grocery amount",
  "date": "2024-01-15"
}
```

**Response (200):**
```json
{
  "message": "Transaction updated successfully"
}
```

---

#### POST /transaction/delete
**Purpose:** Delete a transaction

**Request Body:**
```json
{
  "id": 1
}
```

**Response (200):**
```json
{
  "message": "Transaction deleted successfully"
}
```

---

#### GET /transactions/search
**Purpose:** Advanced transaction search with filters and pagination

**Query Parameters:**
- `category_id` (optional): Filter by category
- `from` (optional): Start date (YYYY-MM-DD)
- `to` (optional): End date (YYYY-MM-DD)
- `min_amount` (optional): Minimum amount
- `max_amount` (optional): Maximum amount
- `keyword` (optional): Search in description
- `sort_by` (optional): "date" or "amount" (default: "date")
- `sort_order` (optional): "asc" or "desc" (default: "desc")
- `limit` (optional): Max results (default: 100, max: 1000)
- `offset` (optional): Pagination offset (default: 0)

**Example:**
```
GET /transactions/search?from=2024-01-01&to=2024-01-31&category_id=1&sort_by=amount&sort_order=desc&limit=50
```

**Response (200):**
```json
{
  "transactions": [...],
  "total": 150,
  "limit": 50,
  "offset": 0
}
```

---

### Recurring Transaction Endpoints

#### POST /recurring/add
**Purpose:** Create a recurring transaction rule

**Request Body:**
```json
{
  "category_id": 1,
  "amount": 50.00,
  "description": "Monthly subscription",
  "start_date": "2024-01-01",
  "recurrence": "monthly"
}
```

**Response (201):**
```json
{
  "id": 1,
  "message": "Recurring transaction added successfully"
}
```

**Recurrence Options:** `daily`, `weekly`, `monthly`, `yearly`

---

#### GET /recurring/list
**Purpose:** List all recurring transaction rules

**Response (200):**
```json
{
  "recurring_transactions": [
    {
      "id": 1,
      "category_id": 1,
      "category_name": "Subscriptions",
      "amount": 50.00,
      "description": "Monthly subscription",
      "start_date": "2024-01-01",
      "recurrence": "monthly",
      "last_occurrence": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T10:00:00Z"
    }
  ]
}
```

---

#### POST /recurring/edit
**Purpose:** Update recurring transaction rule

**Request Body:**
```json
{
  "id": 1,
  "category_id": 1,
  "amount": 55.00,
  "description": "Updated subscription",
  "start_date": "2024-01-01",
  "recurrence": "monthly"
}
```

**Response (200):**
```json
{
  "message": "Recurring transaction updated successfully"
}
```

---

#### POST /recurring/delete
**Purpose:** Delete recurring transaction rule

**Request Body:**
```json
{
  "id": 1
}
```

**Response (200):**
```json
{
  "message": "Recurring transaction deleted successfully"
}
```

---

### Budget Endpoints

#### POST /budget/add
**Purpose:** Create a budget

**Request Body:**
```json
{
  "category_id": 1,
  "amount": 500.00,
  "period": "monthly",
  "alert_threshold": 80
}
```

**Response (201):**
```json
{
  "id": 1,
  "message": "Budget added successfully"
}
```

**Notes:**
- `category_id = 0` for overall budget
- `alert_threshold` is percentage (0-100)

---

#### GET /budget/list
**Purpose:** List all budgets with current spending

**Response (200):**
```json
{
  "budgets": [
    {
      "id": 1,
      "category_id": 1,
      "category_name": "Groceries",
      "amount": 500.00,
      "period": "monthly",
      "alert_threshold": 80,
      "current_spending": 350.00,
      "percentage_used": 70.0,
      "created_at": "2024-01-01T10:00:00Z"
    }
  ]
}
```

---

#### POST /budget/update
**Purpose:** Update budget amount or threshold

**Request Body:**
```json
{
  "id": 1,
  "amount": 600.00,
  "alert_threshold": 85
}
```

**Response (200):**
```json
{
  "message": "Budget updated successfully"
}
```

---

#### POST /budget/delete
**Purpose:** Delete a budget

**Request Body:**
```json
{
  "id": 1
}
```

**Response (200):**
```json
{
  "message": "Budget deleted successfully"
}
```

---

#### GET /budget/alerts
**Purpose:** Get budgets exceeding alert threshold

**Response (200):**
```json
{
  "alerts": [
    {
      "id": 1,
      "category_id": 1,
      "category_name": "Groceries",
      "amount": 500.00,
      "period": "monthly",
      "alert_threshold": 80,
      "current_spending": 450.00,
      "percentage_used": 90.0
    }
  ]
}
```

---

### Summary & Analytics Endpoints

#### GET /summary/totals
**Purpose:** Get overall income and expense totals

**Response (200):**
```json
{
  "total_income": 5000.00,
  "total_expenses": 3500.00,
  "net_balance": 1500.00
}
```

---

#### GET /summary/monthly
**Purpose:** Get monthly breakdown of income and expenses

**Response (200):**
```json
{
  "monthly_summary": [
    {
      "month": "2024-01",
      "total_income": 4000.00,
      "total_expenses": 2500.00,
      "net_income": 1500.00
    },
    {
      "month": "2024-02",
      "total_income": 4200.00,
      "total_expenses": 2800.00,
      "net_income": 1400.00
    }
  ]
}
```

---

#### GET /summary/current-month
**Purpose:** Get current month summary with recurring expense projections

**Response (200):**
```json
{
  "month": "2024-01",
  "total_income": 4000.00,
  "total_expenses": 2500.00,
  "projected_recurring_expenses": 300.00,
  "net_income": 1500.00
}
```

---

#### GET /summary/category
**Purpose:** Get spending breakdown by category

**Query Parameters:**
- `from` (optional): Start date (YYYY-MM-DD)
- `to` (optional): End date (YYYY-MM-DD)

**Example:**
```
GET /summary/category?from=2024-01-01&to=2024-01-31
```

**Response (200):**
```json
{
  "category_summary": [
    {
      "category_id": 1,
      "category_name": "Groceries",
      "category_type": "expense",
      "total_amount": 600.00
    },
    {
      "category_id": 2,
      "category_name": "Salary",
      "category_type": "income",
      "total_amount": 4000.00
    }
  ]
}
```

---

#### GET /summary/group
**Purpose:** Group transactions by time period

**Query Parameters:**
- `by` (required): "day", "week", "month", or "year"

**Example:**
```
GET /summary/group?by=month
```

**Response (200):**
```json
{
  "grouped_summary": [
    {
      "period": "2024-01",
      "total_income": 4000.00,
      "total_expenses": 2500.00
    },
    {
      "period": "2024-02",
      "total_income": 4200.00,
      "total_expenses": 2800.00
    }
  ]
}
```

---

#### GET /summary/category/monthly
**Purpose:** Get category spending for specific month

**Query Parameters:**
- `year` (required): Year (YYYY)
- `month` (required): Month (1-12)

**Example:**
```
GET /summary/category/monthly?year=2024&month=1
```

**Response (200):**
```json
{
  "year": 2024,
  "month": 1,
  "categories": [
    {
      "category_id": 1,
      "category_name": "Groceries",
      "category_type": "expense",
      "total_amount": 600.00
    }
  ]
}
```

---

### Export Endpoints

#### GET /export
**Purpose:** Export transactions to CSV or JSON

**Query Parameters:**
- `format` (required): "csv" or "json"

**Example:**
```
GET /export?format=csv
```

**Response (200):**

**CSV Format:**
```csv
ID,Amount,Description,Date,Category,Type,Created At
1,150.50,Groceries,2024-01-15,Food,expense,2024-01-15T10:00:00Z
```

**JSON Format:**
```json
{
  "transactions": [
    {
      "id": 1,
      "amount": 150.50,
      "description": "Groceries",
      "date": "2024-01-15",
      "category_name": "Food",
      "category_type": "expense",
      "created_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

---

## Authentication & Security

### Authentication Flow

#### 1. User Registration
```
User → POST /register → Server
         ↓
    Validate Input
         ↓
    Hash Password (bcrypt)
         ↓
    Store in Database
         ↓
    Return Success
```

#### 2. User Login
```
User → POST /login → Server
         ↓
    Validate Credentials
         ↓
    Verify Password (bcrypt)
         ↓
    Generate JWT Token (72h expiry)
         ↓
    Return Token
```

#### 3. Protected Request
```
User → Request + JWT → Server
         ↓
    Verify JWT Signature
         ↓
    Extract user_id
         ↓
    Process Request (user-scoped)
         ↓
    Return Response
```

### Security Middleware Stack

**Request Flow:**
```
Client Request
    ↓
1. HTTPS Enforcement (RequireHTTPS)
    ↓
2. CORS Headers (github.com/rs/cors)
    ↓
3. Security Headers (SecurityHeaders)
    ↓
4. Rate Limiting (RateLimitMiddleware)
    ↓
5. Authentication (AuthMiddleware) [Protected routes only]
    ↓
Handler
```

---

### 1. HTTPS Enforcement

**Middleware:** `RequireHTTPS`

**Features:**
- Checks if request is HTTP in production
- Examines X-Forwarded-Proto header (proxy support)
- Redirects HTTP → HTTPS (301 Moved Permanently)
- Adds HSTS header with 1-year max-age
- Disabled in development (allows localhost testing)

**Configuration:**
- Set `ENFORCE_HTTPS=true` in production
- Automatically detects proxy scenarios

**Headers Set:**
```
Strict-Transport-Security: max-age=31536000; includeSubDomains
```

---

### 2. CORS Configuration

**Library:** `github.com/rs/cors`

**Configuration:**
```go
AllowedOrigins: []string{
    "http://localhost:3000",
    "https://app.example.com"
}
AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
AllowedHeaders: []string{"Authorization", "Content-Type"}
AllowCredentials: true
```

**Features:**
- Multiple origin support (comma-separated in env)
- Credentials (cookies, auth headers) allowed
- Preflight request handling (OPTIONS)
- Secure default configuration

**Configuration:**
```bash
CORS_ORIGIN=http://localhost:3000,https://app.example.com
```

---

### 3. Security Headers

**Middleware:** `SecurityHeaders`

**Headers Applied:**

| Header | Value | Protection |
|--------|-------|------------|
| **X-Frame-Options** | `DENY` | Prevents clickjacking attacks |
| **X-Content-Type-Options** | `nosniff` | Prevents MIME sniffing |
| **X-XSS-Protection** | `1; mode=block` | Enables browser XSS filter |
| **Referrer-Policy** | `strict-origin-when-cross-origin` | Controls referrer information |
| **Permissions-Policy** | `geolocation=(), microphone=(), camera=()` | Restricts browser features |
| **Content-Security-Policy** | `default-src 'self'` | Restricts resource loading |

**CSP Configuration:**
- Default: `default-src 'self'`
- Customizable via `CSP_HEADER` environment variable
- Prevents XSS by restricting script sources

---

### 4. Rate Limiting

**Middleware:** `RateLimitMiddleware`

**Algorithm:** Token Bucket

**Limits:**

| Endpoint Type | Requests/Min | Burst |
|---------------|--------------|-------|
| **Auth endpoints** (/register, /login) | 5 | 5 |
| **API endpoints** (all others) | 100 | 20 |

**Features:**
- IP-based tracking
- Separate limiters per IP address
- Automatic cleanup of old limiters (every 10 minutes)
- Supports X-Forwarded-For header (proxy/CDN scenarios)
- Returns 429 Too Many Requests when exceeded

**Headers Set:**
```
Retry-After: 60
```

**Memory Management:**
- Limiters removed after 10 minutes of inactivity
- Prevents memory leaks from abandoned IPs

---

### 5. Input Validation

**Validation Rules:**

#### Username
- Length: 3-50 characters
- Pattern: `^[a-zA-Z0-9_-]+$`
- Allowed: alphanumeric, underscore, hyphen
- Unique: per account

#### Email
- Format: Valid email address
- Pattern: Contains '@' and domain
- Unique: per account

#### Password
- Minimum length: 8 characters
- No maximum (hashed with bcrypt)
- Stored as bcrypt hash (cost factor 10)

#### Amount
- Type: Decimal(10, 2)
- Minimum: Greater than 0
- Maximum: 1,000,000,000 (1 billion)
- Precision: 2 decimal places

#### Date
- Format: YYYY-MM-DD
- Validation: Must be valid date
- Time zone: Normalized to UTC

#### Category Name
- Maximum length: 100 characters
- Sanitized: Trimmed whitespace
- Unique: per user per type

#### Description
- Maximum length: 500 characters
- Sanitized: Trimmed whitespace
- Optional: Can be empty

#### Pagination
- Limit: Max 1000 records per page
- Offset: Non-negative integer
- Default limit: 100

---

### 6. Input Sanitization

**Sanitization Functions:**

**Text Sanitization:**
```go
func SanitizeString(s string) string {
    // Trim whitespace
    s = strings.TrimSpace(s)
    // Remove null bytes
    s = strings.ReplaceAll(s, "\x00", "")
    return s
}
```

**Applied To:**
- Category names
- Descriptions
- Usernames
- Emails

**XSS Prevention:**
- All user input sanitized before storage
- JSON encoding escapes special characters
- Database prepared statements prevent injection

---

### 7. Database Security

**SQL Injection Prevention:**
```go
// ❌ NEVER USED - Vulnerable
query := "SELECT * FROM users WHERE id = " + userInput

// ✅ ALWAYS USED - Safe
query := "SELECT * FROM users WHERE id = $1"
db.QueryContext(ctx, query, userID)
```

**All queries use:**
- Prepared statements with placeholders ($1, $2, etc.)
- Parameter binding
- No string concatenation

**Row-Level Security:**
```sql
-- All queries filter by user_id
SELECT * FROM transactions WHERE user_id = $1
```

**Cascade Protection:**
- Foreign keys with ON DELETE CASCADE
- Prevents orphaned records
- Maintains referential integrity

---

### 8. Password Security

**Hashing:**
- Algorithm: bcrypt
- Cost factor: 10 (2^10 iterations)
- Salt: Automatically generated per password
- Rainbow table resistant

**Process:**
```go
// Registration
hash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)
// Store hash in database

// Login
err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
if err == nil {
    // Password correct
}
```

**Best Practices:**
- Never log passwords
- Never return passwords in API responses
- Never store plain text passwords
- Minimum 8 character requirement

---

### 9. JWT Token Security

**Token Structure:**
```json
{
  "alg": "HS256",
  "typ": "JWT"
}
{
  "user_id": 123,
  "exp": 1735689600
}
```

**Security Features:**
- Signed with HMAC-SHA256
- Secret key from environment variable
- 72-hour expiry
- Contains only user_id (minimal info)

**Token Validation:**
```go
// Parse and validate
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    return []byte(jwtSecret), nil
})

// Verify signature and expiry
if !token.Valid {
    return ErrInvalidToken
}
```

**Best Practices:**
- Use strong JWT_SECRET (32+ characters)
- Rotate secrets periodically
- Store tokens securely on client (httpOnly cookies or secure storage)
- Never share JWT_SECRET

---

### 10. Connection Security

**Database TLS:**
```bash
# Development
DB_SSLMODE=disable

# Production
DB_SSLMODE=require
```

**HTTPS Only:**
- All production traffic over HTTPS
- HSTS header enforces HTTPS
- Certificates from Let's Encrypt or cloud provider

---

### Security Configuration Checklist

**Environment Variables (Production):**
```bash
✅ ENFORCE_HTTPS=true
✅ JWT_SECRET=<strong_random_secret_32+_chars>
✅ DB_SSLMODE=require
✅ CORS_ORIGIN=https://your-exact-frontend-url.com
✅ CSP_HEADER=default-src 'self'
```

**Best Practices:**
- Use environment variables for secrets
- Never commit .env files to git
- Rotate JWT_SECRET periodically
- Monitor rate limit violations
- Review security headers regularly
- Keep dependencies updated

---

## Frontend Implementation

### Pages & Routes

#### Authentication Pages

**1. /login**
- Login form with username/password
- JWT token storage in localStorage
- Redirect to dashboard on success
- Error handling for invalid credentials

**2. /register**
- Registration form with username/email/password
- Input validation
- Redirect to login on success
- Error display for validation failures

---

#### Protected Pages (Require Authentication)

**3. /dashboard**
- Main analytics dashboard
- Charts and graphs (Recharts)
- Summary cards (total income, expenses, net)
- Monthly trends visualization
- Category breakdown charts
- Recent transactions list

**4. /transactions**
- Transaction list with pagination
- Advanced search and filter form
- Sort by date/amount
- Edit/delete actions
- Add transaction button

**5. /transaction/add**
- Transaction creation form
- Category selection dropdown
- Amount input with validation
- Date picker (date-fns)
- Description textarea
- Submit and cancel actions

**6. /transaction/edit/[id]**
- Pre-filled transaction form
- Dynamic route with transaction ID
- Update existing transaction
- Validation and error handling

**7. /budgets**
- Budget list with progress bars
- Current spending vs. budget amount
- Alert indicators (exceeding threshold)
- Add/edit/delete budget actions
- Category filter

**8. /recurring**
- Recurring transaction rule list
- Add recurring rule form
- Edit/delete existing rules
- Recurrence frequency selector
- Next occurrence display

---

### Key Components

#### TransactionForm.tsx (23,514 bytes)
**Purpose:** Reusable form for adding/editing transactions

**Features:**
- Category dropdown (filtered by type)
- Amount input with number validation
- Date picker integration
- Description textarea (500 char limit)
- Form validation before submit
- Loading states during API calls
- Error message display

**Props:**
```typescript
interface TransactionFormProps {
  mode: 'add' | 'edit';
  transactionId?: number;
  initialData?: Transaction;
  onSuccess?: () => void;
}
```

**State Management:**
- Local state for form fields
- Validation errors
- Loading state
- Category list from API

---

#### CategoryManager.tsx (4,522 bytes)
**Purpose:** Manage user categories

**Features:**
- Add new category form
- Category type selector (income/expense)
- Category list display
- Delete category action
- Inline editing
- Category name validation

**Functionality:**
- Fetches categories from API
- Creates new categories
- Deletes categories (with confirmation)
- Filters by type

---

### State Management

#### AuthContext
**File:** `src/context/AuthContext.tsx`

**Purpose:** Global authentication state

**State:**
```typescript
interface AuthContextType {
  token: string | null;
  setToken: (token: string | null) => void;
  isAuthenticated: boolean;
  logout: () => void;
}
```

**Features:**
- Token persistence in localStorage
- Auto-restore token on page reload
- Logout function (clears token + budget alerts)
- Protected route wrapper

**Usage:**
```typescript
const { token, setToken, logout } = useAuth();
```

---

### API Integration

#### lib/api.ts
**Purpose:** Centralized API utilities

**Core Functions:**

**1. apiFetch()**
```typescript
async function apiFetch(
  endpoint: string,
  options?: RequestInit
): Promise<Response>
```
- Adds API base URL
- Injects JWT token from localStorage
- Sets Content-Type header
- Error handling
- Auto-logout on 401

**2. useApi() Hook**
```typescript
function useApi() {
  const { logout } = useAuth();

  const request = useCallback(async (
    endpoint: string,
    options?: RequestInit
  ) => {
    const response = await apiFetch(endpoint, options);
    if (response.status === 401) {
      logout();
      throw new Error('Unauthorized');
    }
    return response;
  }, [logout]);

  return { request };
}
```

**Specific API Functions:**

```typescript
// Transactions
async function fetchTransactions(filters?: TransactionFilters)
async function addTransaction(data: TransactionData)
async function updateTransaction(id: number, data: TransactionData)
async function deleteTransaction(id: number)

// Categories
async function fetchCategories(type?: 'income' | 'expense')
async function addCategory(name: string, type: string)

// Budgets
async function fetchBudgets()
async function addBudget(data: BudgetData)
async function updateBudget(id: number, data: BudgetData)

// Summaries
async function fetchMonthlySummary()
async function fetchCategorySummary(from?: string, to?: string)
```

---

### Styling & Design

#### Tailwind CSS Configuration
**File:** `tailwind.config.ts`

**Custom Theme:**
- Custom color palette
- Extended spacing scale
- Custom typography
- Responsive breakpoints
- Dark mode support (class-based)

**Usage:**
```tsx
<div className="bg-blue-500 hover:bg-blue-600 rounded-lg p-4 shadow-md">
  <h2 className="text-xl font-bold text-white">Total Income</h2>
  <p className="text-3xl font-semibold mt-2">$5,000.00</p>
</div>
```

---

#### Typography
- **Primary Font:** Geist Sans (variable)
- **Monospace Font:** Geist Mono (variable)
- Optimized for readability
- Variable fonts for performance

---

### Data Visualization

#### Recharts Integration

**Chart Types Used:**

**1. Line Chart**
- Monthly income/expense trends
- Time series data
- Multiple lines for comparison
- Tooltips with formatted values

**2. Bar Chart**
- Category spending comparison
- Monthly totals
- Stacked bars for income vs. expense

**3. Pie Chart**
- Category distribution
- Percentage breakdown
- Interactive legend

**4. Area Chart**
- Cumulative trends
- Net income over time
- Gradient fills

**Example:**
```tsx
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip } from 'recharts';

<LineChart data={monthlyData}>
  <CartesianGrid strokeDasharray="3 3" />
  <XAxis dataKey="month" />
  <YAxis />
  <Tooltip />
  <Line type="monotone" dataKey="income" stroke="#22c55e" />
  <Line type="monotone" dataKey="expense" stroke="#ef4444" />
</LineChart>
```

---

### User Experience Features

**1. Loading States**
- Spinner components during API calls
- Skeleton screens for data loading
- Disabled buttons during submission

**2. Error Handling**
- Toast notifications for errors
- Inline validation messages
- Error boundaries for component errors

**3. Form Validation**
- Real-time validation
- Error messages below fields
- Submit button disabled until valid

**4. Responsive Design**
- Mobile-first approach
- Breakpoints for tablet and desktop
- Hamburger menu on mobile
- Touch-friendly buttons

**5. Accessibility**
- Semantic HTML elements
- ARIA labels where needed
- Keyboard navigation support
- Focus visible states

---

### Performance Optimizations

**1. React Best Practices**
- useCallback for function memoization
- useMemo for expensive calculations
- Lazy loading for routes
- Code splitting with dynamic imports

**2. API Optimization**
- Request debouncing for search
- Pagination for large datasets
- Caching with React Query (if implemented)

**3. Asset Optimization**
- Next.js Image component
- Font optimization with next/font
- Static generation where possible

---

## Configuration

### Backend Configuration (.env)

```bash
# ============================================
# Database Configuration (Choose Option 1 OR 2)
# ============================================

# Option 1: Connection String (Recommended for Supabase, Render)
DATABASE_URL=postgresql://username:password@host:port/database?sslmode=require

# Option 2: Individual Parameters
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=expense_tracker
DB_SSLMODE=disable

# ============================================
# Authentication
# ============================================
JWT_SECRET=your_strong_random_secret_at_least_32_characters

# ============================================
# CORS Configuration
# ============================================
# Single origin
CORS_ORIGIN=http://localhost:3000

# Multiple origins (comma-separated)
# CORS_ORIGIN=http://localhost:3000,https://app.example.com

# ============================================
# Security (Production)
# ============================================
ENFORCE_HTTPS=true

# Optional: Custom Content Security Policy
# CSP_HEADER=default-src 'self'; script-src 'self' 'unsafe-inline'

# ============================================
# Server Configuration
# ============================================
PORT=8080
```

---

### Frontend Configuration (.env.local)

```bash
# API Base URL
NEXT_PUBLIC_API_URL=http://localhost:8080

# Production
# NEXT_PUBLIC_API_URL=https://api.example.com
```

---

### Constants Configuration

**File:** `/go-backend/constants/constants.go`

```go
package constants

import "time"

// Database Connection Pool
const (
    MaxOpenConns    = 25
    MaxIdleConns    = 10
    ConnMaxLifetime = 5 * time.Minute
    ConnMaxIdleTime = 2 * time.Minute
    DefaultDBTimeout = 10 * time.Second
)

// Validation Limits
const (
    MinPasswordLength = 8
    MaxAmount         = 1000000000 // 1 billion
    MaxCategoryNameLength = 100
    MaxDescriptionLength  = 500
    MaxPaginationLimit    = 1000
)

// Rate Limiting
const (
    AuthRateLimit = 5    // requests per minute
    AuthBurstLimit = 5
    APIRateLimit = 100   // requests per minute
    APIBurstLimit = 20
)

// Jobs
const (
    RecurringJobInterval = 1 * time.Hour
    MaxRecurringIterations = 3650 // ~10 years daily
)

// JWT
const (
    JWTExpiry = 72 * time.Hour // 3 days
)
```

---

### Production Configuration Checklist

#### Backend
- [ ] Set strong `JWT_SECRET` (32+ characters, random)
- [ ] Enable HTTPS enforcement (`ENFORCE_HTTPS=true`)
- [ ] Use SSL for database (`DB_SSLMODE=require`)
- [ ] Configure exact `CORS_ORIGIN` (no wildcards)
- [ ] Use connection pooling (port 6543 for Supabase)
- [ ] Set `prefer_simple_protocol=true` in DATABASE_URL
- [ ] Review rate limits for your traffic
- [ ] Configure CSP header if needed

#### Frontend
- [ ] Set production `NEXT_PUBLIC_API_URL`
- [ ] Enable HTTPS for all requests
- [ ] Configure error tracking (Sentry, etc.)
- [ ] Set up analytics if needed
- [ ] Optimize images and assets
- [ ] Test on multiple devices/browsers

#### Database
- [ ] Enable SSL/TLS connections
- [ ] Set up automated backups
- [ ] Configure connection pooling
- [ ] Review and optimize indexes
- [ ] Set up monitoring

---

## Background Jobs

### Recurring Transaction Processor

**File:** `/go-backend/jobs/recurring.go`

#### Overview
Automated background job that processes recurring transaction rules and creates actual transactions based on schedules.

#### Execution Schedule
- **Frequency:** Every 1 hour
- **Implementation:** Goroutine with ticker
- **Lifecycle:** Starts with server, stops on shutdown

#### Concurrency Control

**PostgreSQL Advisory Locks:**
```go
// Try to acquire lock
_, err := db.Exec("SELECT pg_try_advisory_lock(123456789)")
if err != nil {
    // Lock already held, skip this run
    return
}

// Always release lock when done
defer db.Exec("SELECT pg_advisory_unlock(123456789)")
```

**Benefits:**
- Prevents duplicate processing in multi-instance deployments
- Safe for horizontal scaling
- No external coordination needed

---

#### Processing Logic

**1. Fetch All Rules**
```sql
SELECT id, user_id, category_id, amount, description,
       start_date, recurrence, last_occurrence
FROM recurring_transactions
```

**2. Calculate Next Due Date**
```go
func calculateNextOccurrence(startDate, lastOccurrence time.Time, recurrence string) time.Time {
    switch recurrence {
    case "daily":
        return lastOccurrence.AddDate(0, 0, 1)
    case "weekly":
        return lastOccurrence.AddDate(0, 0, 7)
    case "monthly":
        return lastOccurrence.AddDate(0, 1, 0)
    case "yearly":
        return lastOccurrence.AddDate(1, 0, 0)
    }
}
```

**3. Process Overdue Occurrences**
```go
for nextDate <= now && iterations < MaxIterations {
    // Create transaction
    createTransaction(userID, categoryID, amount, description, nextDate)

    // Calculate next occurrence
    nextDate = calculateNextOccurrence(startDate, nextDate, recurrence)
    iterations++
}
```

**4. Update Last Occurrence**
```sql
UPDATE recurring_transactions
SET last_occurrence = $1
WHERE id = $2
```

---

#### Edge Case Handling

**1. Month-End Dates**
```
Rule: Monthly on 31st
- Jan 31 ✓
- Feb 28 (or 29 in leap year) ✓
- Mar 31 ✓
- Apr 30 ✓
```

**Solution:** Use last day of month when target day doesn't exist

```go
func normalizeMonthEnd(date time.Time) time.Time {
    year, month, day := date.Date()
    lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
    if day > lastDay {
        day = lastDay
    }
    return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
```

**2. Leap Year Handling**
```
Rule: Yearly on Feb 29
- 2024: Feb 29 ✓ (leap year)
- 2025: Feb 28 ✓ (non-leap year)
- 2026: Feb 28 ✓ (non-leap year)
- 2028: Feb 29 ✓ (leap year)
```

**3. Multiple Missed Occurrences**
```
Scenario: Server down for 1 week
Rule: Daily transaction

Result: Creates 7 transactions on next run
- Catches up all missed dates
- Maintains accurate history
```

**4. Infinite Loop Prevention**
```go
const MaxIterations = 3650 // ~10 years of daily

for iterations < MaxIterations {
    // Process occurrence
    iterations++
}

if iterations >= MaxIterations {
    log.Error("Max iterations reached", "rule_id", ruleID)
}
```

---

#### Date Normalization

**All dates normalized to UTC midnight:**
```go
func normalizeDate(t time.Time) time.Time {
    return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
```

**Prevents:**
- Time zone issues
- Daylight saving time bugs
- Hour/minute discrepancies

---

#### Error Handling

**Transaction-Level:**
- Individual rule failures don't stop processing
- Errors logged with rule ID
- Other rules continue processing

**Lock-Level:**
- Failed lock acquisition → skip run
- Lock automatically released on connection drop
- Next run picks up work

**Database-Level:**
- Connection errors logged
- Retry on next scheduled run
- No data corruption

---

#### Graceful Shutdown

```go
func StartRecurringJob(db *sql.DB, quit chan bool) {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            processRecurringTransactions(db)
        case <-quit:
            log.Info("Stopping recurring job")
            return
        }
    }
}
```

**Features:**
- Responds to shutdown signal
- Completes in-progress processing
- Releases database lock
- Clean exit

---

#### Monitoring & Logging

**Logged Events:**
- Job start/stop
- Rules processed count
- Transactions created count
- Errors with rule details
- Lock acquisition failures
- Processing duration

**Example Logs:**
```
INFO: Recurring job started
INFO: Processing 15 recurring rules
INFO: Created 23 transactions from recurring rules
INFO: Recurring job completed in 1.2s
ERROR: Failed to process rule_id=5: invalid category
```

---

#### Performance Considerations

**Database Queries:**
- Batch fetch all rules (1 query)
- Individual inserts for transactions (prepared statements)
- Single update per rule after processing

**Scaling:**
- Advisory locks enable multi-instance deployment
- Each instance tries to acquire lock
- Only one processes at a time
- Safe for horizontal scaling

**Resource Usage:**
- Runs in separate goroutine
- Doesn't block main server
- Connection pool shared with API
- Minimal CPU usage (hourly)

---

## Testing

### Test Files

| File | Purpose | Lines |
|------|---------|-------|
| `utils/db_test.go` | Database utility tests | TBD |
| `utils/response_test.go` | Response helper tests | TBD |
| `utils/sanitize_test.go` | Input sanitization tests | TBD |
| `utils/validation_test.go` | Validation logic tests | TBD |

---

### Running Tests

**Backend (Go):**
```bash
cd go-backend

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./utils

# Run specific test
go test ./utils -run TestValidateUsername
```

**Frontend (Next.js):**
```bash
cd nextjs-frontend

# Run tests (if configured)
npm test

# Run with coverage
npm test -- --coverage
```

---

### Test Coverage Areas

**1. Validation Functions**
- Username format validation
- Email format validation
- Password length validation
- Amount validation (positive, max limit)
- Date format validation
- Category name length
- Description length

**2. Sanitization Functions**
- String trimming
- Null byte removal
- XSS prevention
- SQL injection prevention

**3. Database Utilities**
- Connection establishment
- Query execution
- Transaction handling
- Error scenarios

**4. Response Helpers**
- JSON encoding
- Error responses
- Success responses
- Status codes

---

### Manual Testing

**Testing.txt:** Contains manual test scenarios and results

**Test Scenarios:**
- User registration and login
- Transaction CRUD operations
- Category management
- Recurring transaction processing
- Budget tracking and alerts
- Summary endpoint accuracy
- Export functionality
- Rate limiting behavior
- Authentication edge cases

---

## Deployment

### Recommended Infrastructure

| Component | Platform | Reason |
|-----------|----------|--------|
| **Backend** | Render.com | Easy Go deployment, free tier |
| **Frontend** | Vercel | Optimized for Next.js, free tier |
| **Database** | Supabase | Managed PostgreSQL, free tier |

---

### Backend Deployment (Render.com)

**1. Create Web Service**
- Repository: Connect your GitHub repo
- Branch: `main`
- Root Directory: `go-backend`
- Runtime: Go

**2. Build Command**
```bash
go build -o app .
```

**3. Start Command**
```bash
./app
```

**4. Environment Variables**
```bash
DATABASE_URL=postgresql://user:pass@host:6543/db?sslmode=require&prefer_simple_protocol=true
JWT_SECRET=your_production_secret_32_plus_characters
CORS_ORIGIN=https://your-frontend-domain.vercel.app
ENFORCE_HTTPS=true
PORT=8080
```

**5. Health Check Endpoint**
- Path: `/health` (if implemented)
- Or use any public endpoint like `/login`

---

### Frontend Deployment (Vercel)

**1. Import Project**
- Connect GitHub repository
- Framework Preset: Next.js
- Root Directory: `nextjs-frontend`

**2. Build Settings**
- Build Command: `npm run build` (auto-detected)
- Output Directory: `.next` (auto-detected)
- Install Command: `npm install` (auto-detected)

**3. Environment Variables**
```bash
NEXT_PUBLIC_API_URL=https://your-backend.onrender.com
```

**4. Deploy**
- Automatic deployments on git push
- Production: `main` branch
- Preview: Pull requests

---

### Database Setup (Supabase)

**1. Create Project**
- Sign up at supabase.com
- Create new project
- Choose region closest to users
- Note the database password

**2. Get Connection Details**
- Go to Project Settings → Database
- Connection String (Pooling): Port 6543
- Copy the connection string
- Add `prefer_simple_protocol=true` parameter

**3. Run Migrations**
```bash
# Local machine with psql
psql "postgresql://user:pass@host:6543/db?sslmode=require" < migrations/001_initial_schema.sql

# Or use Supabase SQL Editor
# Copy/paste migration files into editor
```

**4. Verify Schema**
```sql
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public';
```

---

### Environment-Specific Configuration

#### Development
```bash
# .env (backend)
DATABASE_URL=postgresql://postgres:password@localhost:5432/expense_tracker?sslmode=disable
JWT_SECRET=dev_secret_not_secure
CORS_ORIGIN=http://localhost:3000
ENFORCE_HTTPS=false
PORT=8080

# .env.local (frontend)
NEXT_PUBLIC_API_URL=http://localhost:8080
```

#### Production
```bash
# Render Environment Variables
DATABASE_URL=postgresql://user:pass@host:6543/db?sslmode=require&prefer_simple_protocol=true
JWT_SECRET=production_strong_random_secret_32_plus_characters
CORS_ORIGIN=https://expense-tracker.vercel.app
ENFORCE_HTTPS=true

# Vercel Environment Variables
NEXT_PUBLIC_API_URL=https://expense-tracker-api.onrender.com
```

---

### Deployment Checklist

#### Pre-Deployment
- [ ] All tests passing
- [ ] Environment variables documented
- [ ] Database migrations ready
- [ ] CORS origins configured correctly
- [ ] Strong JWT secret generated
- [ ] SSL/TLS enabled for database
- [ ] Rate limits appropriate for traffic
- [ ] Error logging configured

#### Post-Deployment
- [ ] Health check endpoint responding
- [ ] Database connection successful
- [ ] Frontend can reach backend API
- [ ] CORS working (no console errors)
- [ ] Authentication flow working
- [ ] Test transaction creation
- [ ] Test recurring job running
- [ ] Monitor logs for errors
- [ ] Performance testing
- [ ] Security headers present

---

### Monitoring & Maintenance

**Backend Monitoring:**
- Render.com logs and metrics
- Database connection pool usage
- API response times
- Error rates
- Rate limit violations

**Frontend Monitoring:**
- Vercel analytics
- Core Web Vitals
- Client-side errors
- API call failures

**Database Monitoring:**
- Supabase dashboard
- Query performance
- Connection count
- Storage usage
- Backup status

---

### Scaling Considerations

**Horizontal Scaling:**
- Multiple Render instances supported
- Advisory locks prevent job conflicts
- Stateless API design
- Connection pooling handles concurrency

**Vertical Scaling:**
- Adjust connection pool size
- Increase Render instance resources
- Optimize database queries
- Add indexes as needed

**Performance Optimization:**
- Database query optimization
- Add caching layer (Redis)
- CDN for frontend assets
- Database read replicas if needed

---

## Project Statistics

### Codebase Metrics

| Component | Metric | Value |
|-----------|--------|-------|
| **Backend** | Main server | ~1,349 lines |
| **Backend** | Handlers | ~1,045 lines |
| **Backend** | Total Go files | 20+ files |
| **Frontend** | TransactionForm | 23,514 bytes |
| **Frontend** | CategoryManager | 4,522 bytes |
| **Database** | Tables | 5 core tables |
| **Database** | Indexes | 11 indexes |
| **API** | Endpoints | 25+ endpoints |
| **API** | Public endpoints | 2 |
| **API** | Protected endpoints | 23+ |

---

### Features Summary

| Category | Count |
|----------|-------|
| **Core Features** | 7 major features |
| **Authentication** | JWT-based, bcrypt hashing |
| **Middleware** | 5 middleware layers |
| **Background Jobs** | 1 recurring processor |
| **Security Headers** | 6 headers |
| **Rate Limits** | 2 tiers (auth & API) |
| **Export Formats** | 2 (CSV, JSON) |
| **Chart Types** | 4 (Line, Bar, Pie, Area) |
| **Frontend Pages** | 8 pages |
| **React Components** | 2 major + misc |

---

### Technology Stack Summary

**Backend:**
- Go 1.24
- PostgreSQL
- 6 Go libraries (JWT, bcrypt, CORS, rate, godotenv, pgx)

**Frontend:**
- Next.js 15
- React 19
- TypeScript 5
- Tailwind CSS 4
- Recharts 3.1
- 5 npm packages (date-fns, react-icons, etc.)

**Infrastructure:**
- Render (backend hosting)
- Vercel (frontend hosting)
- Supabase (database)

---

### Development Timeline Estimate

This project represents approximately:
- **Planning & Design:** 5-10 hours
- **Backend Development:** 30-40 hours
- **Frontend Development:** 25-35 hours
- **Testing & Debugging:** 15-20 hours
- **Documentation:** 10-15 hours
- **Deployment Setup:** 5-10 hours

**Total:** ~90-130 hours of development work

---

### License

**MIT License**

This project is open source and available under the MIT License. See LICENSE file for details.

---

### Documentation Files

- **README.md** - Quick start guide
- **DEPLOYMENT.md** - Comprehensive deployment instructions
- **API.md** - API endpoint documentation
- **PROJECT_DESCRIPTION.md** - This file (complete project overview)
- **Testing.txt** - Manual testing notes

---

### Future Enhancement Ideas

**Potential Features:**
1. Multi-currency support
2. Receipt photo uploads
3. Bill splitting functionality
4. Financial goal tracking
5. Investment portfolio tracking
6. Tax report generation
7. Mobile app (React Native)
8. Email notifications for budget alerts
9. Scheduled report emails
10. Import from bank CSV
11. Multi-user households
12. Shared budgets
13. Custom report builder
14. API webhooks
15. OAuth social login

**Technical Improvements:**
1. GraphQL API option
2. WebSocket for real-time updates
3. Redis caching layer
4. Full-text search (Elasticsearch)
5. Database sharding
6. Read replicas
7. Microservices architecture
8. Event-driven architecture
9. Message queue (RabbitMQ/Kafka)
10. Kubernetes deployment

---

## Conclusion

This expense tracker is a **production-ready, full-stack financial management application** demonstrating modern web development practices with a focus on:

✅ **Security** - JWT auth, bcrypt, HTTPS, rate limiting, input validation
✅ **Scalability** - Connection pooling, advisory locks, stateless design
✅ **Maintainability** - Clean architecture, structured logging, comprehensive docs
✅ **User Experience** - Responsive design, charts, real-time feedback
✅ **Reliability** - Error handling, graceful shutdown, data integrity
✅ **Performance** - Optimized queries, indexes, background jobs

The project serves as an excellent reference for building secure, scalable Go + Next.js applications with PostgreSQL.

---

**Project Repository:** [Add your GitHub URL here]
**Live Demo:** [Add your deployment URL here]
**Contact:** [Add your contact information here]

---

*Last Updated: 2026-01-05*
