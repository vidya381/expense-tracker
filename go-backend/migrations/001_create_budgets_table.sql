-- Create budgets table
CREATE TABLE IF NOT EXISTS budgets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id INTEGER DEFAULT 0, -- 0 means overall budget, otherwise references categories(id)
    amount DECIMAL(10, 2) NOT NULL CHECK (amount >= 0),
    period VARCHAR(10) NOT NULL CHECK (period IN ('monthly', 'yearly')),
    alert_threshold INTEGER NOT NULL CHECK (alert_threshold >= 0 AND alert_threshold <= 100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Ensure unique budget per user/category/period combination
    UNIQUE(user_id, category_id, period)
);

-- Create index for faster queries
CREATE INDEX idx_budgets_user_id ON budgets(user_id);
CREATE INDEX idx_budgets_category_id ON budgets(category_id);

-- Note: We don't add a foreign key constraint for category_id because:
-- - When category_id = 0, it represents an "overall budget" (not a specific category)
-- - When category_id > 0, it should reference categories(id), but PostgreSQL doesn't support conditional foreign keys
-- The application layer (handlers/budget.go) handles validation of category_id
