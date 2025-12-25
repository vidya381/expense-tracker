-- Add indexes for frequently queried columns to improve performance

-- Transactions table indexes
-- user_id is used in almost every transaction query
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);

-- date is used for filtering and sorting (e.g., monthly summaries, date ranges)
CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date);

-- category_id is used for joins and category-specific queries
CREATE INDEX IF NOT EXISTS idx_transactions_category_id ON transactions(category_id);

-- Composite index for common query pattern: user + date
CREATE INDEX IF NOT EXISTS idx_transactions_user_date ON transactions(user_id, date DESC);

-- Composite index for user + category (budget calculations)
CREATE INDEX IF NOT EXISTS idx_transactions_user_category ON transactions(user_id, category_id);

-- Categories table indexes
-- user_id for listing user's categories
CREATE INDEX IF NOT EXISTS idx_categories_user_id ON categories(user_id);

-- Composite index for user + type (filtering by income/expense)
CREATE INDEX IF NOT EXISTS idx_categories_user_type ON categories(user_id, type);

-- Budgets table indexes
-- user_id for listing user's budgets
CREATE INDEX IF NOT EXISTS idx_budgets_user_id ON budgets(user_id);

-- category_id for category-specific budget queries
CREATE INDEX IF NOT EXISTS idx_budgets_category_id ON budgets(category_id);

-- Composite index for user + category (checking existing budgets)
CREATE INDEX IF NOT EXISTS idx_budgets_user_category ON budgets(user_id, category_id);

-- Recurring transactions table indexes
-- user_id for listing user's recurring transactions
CREATE INDEX IF NOT EXISTS idx_recurring_user_id ON recurring_transactions(user_id);

-- start_date for processing recurring jobs
CREATE INDEX IF NOT EXISTS idx_recurring_start_date ON recurring_transactions(start_date);

-- last_occurrence for calculating next due date
CREATE INDEX IF NOT EXISTS idx_recurring_last_occurrence ON recurring_transactions(last_occurrence);
