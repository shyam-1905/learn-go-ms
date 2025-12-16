-- Migration: Create receipts table
-- This SQL script creates the receipts table with all required fields
-- Run this script against your PostgreSQL database before starting the service

-- Create the receipts table
CREATE TABLE IF NOT EXISTS receipts (
    -- UUID primary key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- User ID from auth-service (UUID, no foreign key since different DB)
    user_id UUID NOT NULL,
    
    -- Expense ID from expense-service (UUID, nullable - links to expense)
    expense_id UUID NULL,
    
    -- File information
    file_name VARCHAR(255) NOT NULL,
    file_key VARCHAR(500) NOT NULL,  -- S3 object key
    file_url VARCHAR(1000) NOT NULL, -- S3 presigned URL or public URL
    file_size BIGINT NOT NULL,       -- File size in bytes
    mime_type VARCHAR(100) NOT NULL, -- e.g., "image/jpeg", "image/png", "application/pdf"
    
    -- Optional receipt metadata (can be extracted via OCR later)
    merchant_name VARCHAR(255) NULL,
    receipt_date DATE NULL,
    total_amount DECIMAL(10, 2) NULL,
    
    -- Timestamps for auditing
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Soft delete support
    deleted_at TIMESTAMP NULL
);

-- Create indexes for better query performance
-- Indexes speed up lookups and filtering

-- Index on user_id (most common query - get user's receipts)
CREATE INDEX IF NOT EXISTS idx_receipts_user_id ON receipts(user_id) WHERE deleted_at IS NULL;

-- Index on expense_id (for fetching receipts by expense)
CREATE INDEX IF NOT EXISTS idx_receipts_expense_id ON receipts(expense_id) WHERE deleted_at IS NULL;

-- Composite index on user_id and expense_id (for user's receipts for a specific expense)
CREATE INDEX IF NOT EXISTS idx_receipts_user_expense ON receipts(user_id, expense_id) WHERE deleted_at IS NULL;

-- Index on created_at (for sorting by date)
CREATE INDEX IF NOT EXISTS idx_receipts_created_at ON receipts(created_at) WHERE deleted_at IS NULL;

-- Add a comment to the table (documentation)
COMMENT ON TABLE receipts IS 'Stores user receipt records with S3 file references and optional expense links';
