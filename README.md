# Expense Tracker

A full-stack personal finance management application for tracking expenses, managing budgets, and analyzing spending patterns.

## Features

- **Transaction Management** - Add, edit, delete, and search transactions with filtering options
- **Budget Management** - Set monthly or yearly budgets with customizable alert thresholds
- **Recurring Transactions** - Automate regular income/expenses (daily, weekly, monthly, yearly)
- **Financial Analytics** - View spending breakdowns by category, time period, and trends
- **Category Management** - Create custom categories for organizing transactions
- **Data Export** - Export transaction data in CSV or JSON format

## Tech Stack

### Frontend
- Next.js 15 with React 19
- TypeScript
- Tailwind CSS 4
- Recharts for data visualization

### Backend
- Go 1.24
- PostgreSQL
- JWT authentication
- RESTful API

## Getting Started

### Prerequisites
- Node.js 18+
- Go 1.24+
- PostgreSQL

### Installation

1. Clone the repository
```bash
git clone https://github.com/vidya381/expense-tracker.git
cd expense-tracker
```

2. Set up the backend
```bash
cd go-backend
go mod download
# Configure your database in .env
go run .
```

3. Set up the frontend
```bash
cd nextjs-frontend
npm install
# Configure API_URL in .env.local
npm run dev
```

## License

MIT
