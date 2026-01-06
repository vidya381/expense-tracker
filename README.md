# 💰 Expense Tracker

A production-ready, full-stack personal finance management application for tracking income, expenses, budgets, and analyzing spending patterns with comprehensive analytics.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/go-1.24-blue.svg)
![Next.js](https://img.shields.io/badge/next.js-15-black.svg)
![React](https://img.shields.io/badge/react-19-blue.svg)

---

## 📋 Table of Contents

- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [Project Structure](#-project-structure)
- [Getting Started](#-getting-started)
- [Environment Variables](#-environment-variables)
- [API Documentation](#-api-documentation)
- [Deployment](#-deployment)
- [Security Features](#-security-features)
- [Documentation](#-documentation)
- [Contributing](#-contributing)
- [License](#-license)

---

## ✨ Features

### Core Functionality

- **🔐 User Authentication**
  - Secure JWT-based authentication (72-hour token expiry)
  - Password hashing with bcrypt
  - Protected routes and user session management

- **💸 Transaction Management**
  - Add, edit, and delete income/expense transactions
  - Advanced search with filters (category, date range, amount, keyword)
  - Sorting and pagination support
  - Transaction history with detailed records

- **📊 Budget Management**
  - Set budgets per category or overall spending
  - Monthly and yearly budget periods
  - Customizable alert thresholds (percentage-based)
  - Real-time budget tracking and notifications

- **🔄 Recurring Transactions**
  - Automated recurring transactions (daily, weekly, monthly, yearly)
  - Background job processes transactions every hour
  - Handles edge cases (month-end dates, leap years)
  - Catch-up for missed occurrences

- **📈 Financial Analytics & Reports**
  - Total income and expense summaries
  - Monthly historical breakdowns
  - Category-wise spending analysis
  - Time-period grouping (day/week/month/year)
  - Interactive charts and visualizations

- **🏷️ Category Management**
  - Create custom categories for income and expenses
  - Organize transactions by purpose
  - Category-based reporting

- **📤 Data Export**
  - Export transactions to CSV format
  - Export transactions to JSON format
  - Backup and data portability

---

## 🛠️ Tech Stack

### Frontend
- **Next.js 15** - React framework with App Router
- **React 19** - UI library
- **TypeScript 5** - Type safety
- **Tailwind CSS 4** - Utility-first CSS framework
- **Recharts 3.1** - Interactive charts and graphs
- **date-fns 4.1** - Date manipulation
- **React Context API** - State management

### Backend
- **Go 1.24** - Backend programming language
- **PostgreSQL** - Primary database with connection pooling
- **pgx/v5** - PostgreSQL driver for Go
- **JWT** - Token-based authentication (`golang-jwt/jwt/v5`)
- **bcrypt** - Password hashing (`golang.org/x/crypto/bcrypt`)
- **CORS** - Cross-origin resource sharing (`github.com/rs/cors`)
- **Rate Limiting** - IP-based rate limiting (`golang.org/x/time/rate`)

### Security & Middleware
- HTTPS enforcement with HSTS
- Security headers (XSS, clickjacking protection)
- Rate limiting (5 req/min for auth, 100 req/min for API)
- Input validation and sanitization
- SQL injection prevention (prepared statements)
- CORS configuration

### Infrastructure
- **Render.com** - Backend hosting (recommended)
- **Vercel** - Frontend hosting (recommended)
- **Supabase** - Managed PostgreSQL database (recommended)

---

## 📁 Project Structure

```
expense-tracker/
├── go-backend/                 # Backend API server
│   ├── main.go                # Entry point, routes, HTTP server
│   ├── db.go                  # Database connection utilities
│   ├── migrate.go             # Database migration runner
│   ├── constants/             # Application constants
│   ├── models/                # Data models (User, Transaction, etc.)
│   ├── handlers/              # HTTP request handlers
│   │   ├── auth.go           # Authentication handlers
│   │   ├── transaction.go    # Transaction CRUD
│   │   ├── category.go       # Category management
│   │   ├── budget.go         # Budget tracking
│   │   ├── recurring.go      # Recurring transactions
│   │   └── summary.go        # Analytics and reports
│   ├── middleware/            # HTTP middleware
│   │   ├── auth.go           # JWT authentication
│   │   ├── ratelimit.go      # Rate limiting
│   │   └── security.go       # Security headers
│   ├── jobs/                  # Background jobs
│   │   └── recurring.go      # Recurring transaction processor
│   ├── utils/                 # Helper functions
│   │   ├── validation.go     # Input validation
│   │   ├── sanitize.go       # Input sanitization
│   │   ├── response.go       # HTTP response helpers
│   │   └── db.go             # Database utilities
│   ├── migrations/            # SQL migration files
│   ├── go.mod                 # Go dependencies
│   └── .env.example          # Environment variables template
│
├── nextjs-frontend/           # Frontend application
│   └── src/
│       ├── app/              # Next.js app router pages
│       │   ├── page.tsx      # Home page
│       │   ├── login/        # Login page
│       │   ├── register/     # Registration page
│       │   ├── dashboard/    # Main dashboard
│       │   ├── transactions/ # Transaction list
│       │   ├── transaction/  # Add/edit transaction
│       │   ├── budgets/      # Budget management
│       │   └── recurring/    # Recurring transactions
│       ├── components/       # React components
│       │   ├── TransactionForm.tsx
│       │   └── CategoryManager.tsx
│       ├── context/          # React context
│       │   └── AuthContext.tsx
│       ├── lib/              # Utilities
│       │   └── api.ts        # API integration
│       └── styles/           # Global styles
│
├── README.md                  # This file
├── PROJECT_DESCRIPTION.md     # Comprehensive project documentation
├── DEPLOYMENT.md              # Deployment guide
└── LICENSE                    # MIT License
```

---

## 🚀 Getting Started

### Prerequisites

Before you begin, ensure you have the following installed:

- **Node.js** 18+ ([Download](https://nodejs.org/))
- **Go** 1.24+ ([Download](https://golang.org/dl/))
- **PostgreSQL** 14+ ([Download](https://www.postgresql.org/download/))
- **Git** ([Download](https://git-scm.com/downloads))

### Installation Steps

#### 1. Clone the Repository

```bash
git clone https://github.com/vidya381/expense-tracker.git
cd expense-tracker
```

#### 2. Set Up PostgreSQL Database

Create a new database for the application:

```bash
# Login to PostgreSQL
psql -U postgres

# Create database
CREATE DATABASE expense_tracker;

# Create user (optional)
CREATE USER expense_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE expense_tracker TO expense_user;

# Exit
\q
```

#### 3. Set Up the Backend

```bash
# Navigate to backend directory
cd go-backend

# Install Go dependencies
go mod download

# Copy environment variables template
cp .env.example .env

# Edit .env with your configuration
# Required: DATABASE_URL or DB_* variables, JWT_SECRET
nano .env  # or use your preferred editor

# Run database migrations
go run migrate.go

# Start the backend server
go run main.go db.go

# Server will start on http://localhost:8080
```

#### 4. Set Up the Frontend

Open a new terminal window:

```bash
# Navigate to frontend directory
cd nextjs-frontend

# Install npm dependencies
npm install

# Copy environment variables template
cp .env.example .env.local

# Edit .env.local with your configuration
# Required: NEXT_PUBLIC_API_URL
nano .env.local  # or use your preferred editor

# Start the development server
npm run dev

# Frontend will start on http://localhost:3000
```

#### 5. Access the Application

Open your browser and navigate to:

```
http://localhost:3000
```

**Default behavior:**
- Create a new account via the registration page
- Login with your credentials
- Start tracking your expenses!

---

## 🔧 Environment Variables

### Backend (.env)

Create a `.env` file in the `go-backend` directory:

```bash
# Database Configuration (Choose Option 1 OR 2)

# Option 1: Connection String (Recommended)
DATABASE_URL=postgresql://username:password@localhost:5432/expense_tracker?sslmode=disable

# Option 2: Individual Parameters
# DB_HOST=localhost
# DB_PORT=5432
# DB_USER=postgres
# DB_PASSWORD=your_password
# DB_NAME=expense_tracker
# DB_SSLMODE=disable

# Authentication (REQUIRED)
JWT_SECRET=your_strong_random_secret_at_least_32_characters

# CORS Configuration
CORS_ORIGIN=http://localhost:3000

# Security (Production)
# ENFORCE_HTTPS=true

# Server Configuration
PORT=8080
```

**Important:**
- Generate a strong `JWT_SECRET` using: `openssl rand -hex 32`
- Never commit `.env` files to version control
- Use `sslmode=require` in production

### Frontend (.env.local)

Create a `.env.local` file in the `nextjs-frontend` directory:

```bash
# API Base URL
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## 📚 API Documentation

### Authentication Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/register` | Create new user account | No |
| POST | `/login` | Login and get JWT token | No |

### Transaction Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/transaction/add` | Create new transaction | Yes |
| GET | `/transaction/list` | List all transactions | Yes |
| POST | `/transaction/update` | Update transaction | Yes |
| POST | `/transaction/delete` | Delete transaction | Yes |
| GET | `/transactions/search` | Advanced search with filters | Yes |

### Category Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/category/add` | Create new category | Yes |
| GET | `/category/list` | List all categories | Yes |

### Budget Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/budget/add` | Create budget | Yes |
| GET | `/budget/list` | List all budgets | Yes |
| POST | `/budget/update` | Update budget | Yes |
| POST | `/budget/delete` | Delete budget | Yes |
| GET | `/budget/alerts` | Get budget alerts | Yes |

### Recurring Transaction Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/recurring/add` | Create recurring rule | Yes |
| GET | `/recurring/list` | List recurring rules | Yes |
| POST | `/recurring/edit` | Update recurring rule | Yes |
| POST | `/recurring/delete` | Delete recurring rule | Yes |

### Summary & Analytics Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/summary/totals` | Overall income/expense totals | Yes |
| GET | `/summary/monthly` | Monthly breakdown | Yes |
| GET | `/summary/current-month` | Current month with projections | Yes |
| GET | `/summary/category` | Category-wise spending | Yes |
| GET | `/summary/group` | Time-period grouping | Yes |
| GET | `/summary/category/monthly` | Category spending per month | Yes |

### Export Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/export?format=csv\|json` | Export transactions | Yes |

**Authentication:**
All protected endpoints require a JWT token in the Authorization header:
```
Authorization: Bearer <your_jwt_token>
```

For detailed API documentation with request/response examples, see [`go-backend/API.md`](go-backend/API.md) and [`PROJECT_DESCRIPTION.md`](PROJECT_DESCRIPTION.md).

---

## 🚢 Deployment

### Quick Deploy

**Recommended Stack:**
- **Backend:** Render.com (Free tier available)
- **Frontend:** Vercel (Free tier available)
- **Database:** Supabase (Free tier available)

### Production Configuration

**Backend Environment Variables:**
```bash
DATABASE_URL=postgresql://user:pass@host:6543/db?sslmode=require&prefer_simple_protocol=true
JWT_SECRET=production_strong_random_secret_32_plus_characters
CORS_ORIGIN=https://your-frontend-domain.vercel.app
ENFORCE_HTTPS=true
```

**Frontend Environment Variables:**
```bash
NEXT_PUBLIC_API_URL=https://your-backend.onrender.com
```

For detailed deployment instructions, see [**DEPLOYMENT.md**](DEPLOYMENT.md).

---

## 🔒 Security Features

This application implements multiple layers of security:

### Authentication & Authorization
- ✅ JWT-based authentication with 72-hour token expiry
- ✅ Bcrypt password hashing (cost factor 10)
- ✅ Protected routes with token validation

### Security Middleware
- ✅ HTTPS enforcement with HSTS headers
- ✅ Security headers (X-Frame-Options, X-XSS-Protection, CSP)
- ✅ IP-based rate limiting (5 req/min for auth, 100 req/min for API)
- ✅ CORS configuration with allowed origins

### Input Protection
- ✅ Input validation (username, email, amount, date formats)
- ✅ Input sanitization (XSS prevention)
- ✅ SQL injection prevention (prepared statements)
- ✅ Maximum input length limits

### Database Security
- ✅ Prepared statements for all queries
- ✅ Foreign key constraints with cascade delete
- ✅ Row-level security (user_id filtering)
- ✅ Connection pooling with timeouts
- ✅ SSL/TLS support for production

### API Security
- ✅ Token-based authentication
- ✅ Request validation
- ✅ Error handling without information leakage
- ✅ Audit logging

---

## 📖 Documentation

Comprehensive documentation is available:

- **[PROJECT_DESCRIPTION.md](PROJECT_DESCRIPTION.md)** - Complete project documentation with architecture, features, API details, and database schema
- **[DEPLOYMENT.md](DEPLOYMENT.md)** - Detailed deployment guide for Render, Vercel, and Supabase
- **[go-backend/API.md](go-backend/API.md)** - API endpoint documentation (if exists)
- **[Testing.txt](Testing.txt)** - Manual testing scenarios and results

---

## 🧪 Testing

### Run Backend Tests

```bash
cd go-backend

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./utils -v
```

### Run Frontend Tests

```bash
cd nextjs-frontend

# Run tests (if configured)
npm test

# Run with coverage
npm test -- --coverage
```

---

## 🛠️ Development

### Backend Development

```bash
cd go-backend

# Install dependencies
go mod download

# Run with hot reload (using air)
go install github.com/cosmtrek/air@latest
air

# Or run normally
go run main.go db.go

# Format code
go fmt ./...

# Run linter
go vet ./...
```

### Frontend Development

```bash
cd nextjs-frontend

# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Start production server
npm start

# Lint code
npm run lint
```

### Database Migrations

```bash
cd go-backend

# Run migrations
go run migrate.go

# Add new migration
# Create a new file in migrations/ folder
# Example: migrations/003_add_new_table.sql
```

---

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. **Commit your changes**
   ```bash
   git commit -m "Add amazing feature"
   ```
4. **Push to the branch**
   ```bash
   git push origin feature/amazing-feature
   ```
5. **Open a Pull Request**

### Contribution Guidelines

- Write clear, descriptive commit messages
- Add tests for new features
- Update documentation as needed
- Follow existing code style and conventions
- Ensure all tests pass before submitting PR

---

## 📊 Project Stats

- **Backend:** ~1,349 lines (main.go) + ~1,045 lines (handlers)
- **API Endpoints:** 25+ RESTful endpoints
- **Database Tables:** 5 core tables with 11 indexes
- **Frontend Pages:** 8 main pages
- **Security Layers:** 5 middleware components
- **Test Files:** 4 test suites

---

## 🗺️ Roadmap

Future enhancements being considered:

- [ ] Multi-currency support
- [ ] Receipt photo uploads
- [ ] Bill splitting functionality
- [ ] Financial goal tracking
- [ ] Investment portfolio tracking
- [ ] Tax report generation
- [ ] Mobile app (React Native)
- [ ] Email notifications for budget alerts
- [ ] Import from bank CSV
- [ ] OAuth social login
- [ ] Multi-user households

---

## 📝 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

## 👤 Author

**Vidyasagar**

- GitHub: [@vidya381](https://github.com/vidya381)
- Project Repository: [expense-tracker](https://github.com/vidya381/expense-tracker)

---

## 🙏 Acknowledgments

- Built with [Go](https://golang.org/)
- Frontend powered by [Next.js](https://nextjs.org/) and [React](https://react.dev/)
- Charts by [Recharts](https://recharts.org/)
- Database: [PostgreSQL](https://www.postgresql.org/)
- Styling: [Tailwind CSS](https://tailwindcss.com/)

---

## 📞 Support

If you encounter any issues or have questions:

1. Check the [PROJECT_DESCRIPTION.md](PROJECT_DESCRIPTION.md) for detailed documentation
2. Review the [DEPLOYMENT.md](DEPLOYMENT.md) for deployment issues
3. Open an issue on [GitHub Issues](https://github.com/vidya381/expense-tracker/issues)

---

**⭐ If you find this project helpful, please consider giving it a star on GitHub!**

---

*Last Updated: January 2026*
