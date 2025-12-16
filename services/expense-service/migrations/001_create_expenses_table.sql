-- Migration: Create expenses table
-- This SQL script creates the expenses table with all required fields
-- Run this script against your PostgreSQL database before starting the service

-- Create the expenses table
CREATE TABLE IF NOT EXISTS expenses (
    -- UUID primary key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- User ID from auth-service (UUID, no foreign key since different DB)
    user_id UUID NOT NULL,
    
    -- Amount as DECIMAL for precise currency handling
    -- DECIMAL(10,2) allows up to 99,999,999.99
    amount DECIMAL(10, 2) NOT NULL,
    
    -- Description of the expense
    description VARCHAR(500) NOT NULL,
    
    -- Category (e.g., Food, Transport, Entertainment)
    category VARCHAR(50) NOT NULL,
    
    -- Date when the expense occurred
    expense_date DATE NOT NULL,
    
    -- Timestamps for auditing
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Soft delete support
    deleted_at TIMESTAMP NULL
);

-- Create indexes for better query performance
-- Indexes speed up lookups and filtering

-- Index on user_id (most common query - get user's expenses)
CREATE INDEX IF NOT EXISTS idx_expenses_user_id ON expenses(user_id) WHERE deleted_at IS NULL;

-- Index on category (for category filtering)
CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category) WHERE deleted_at IS NULL;

-- Index on expense_date (for date range queries)
CREATE INDEX IF NOT EXISTS idx_expenses_date ON expenses(expense_date) WHERE deleted_at IS NULL;

-- Composite index on user_id and expense_date (for user's expenses sorted by date)
CREATE INDEX IF NOT EXISTS idx_expenses_user_date ON expenses(user_id, expense_date) WHERE deleted_at IS NULL;

-- Add a comment to the table (documentation)
COMMENT ON TABLE expenses IS 'Stores user expense records with category and date information';
