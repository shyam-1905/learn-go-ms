-- Migration: Create users table
-- This SQL script creates the users table with all required fields
-- Run this script against your PostgreSQL database before starting the service

-- Create the users table
CREATE TABLE IF NOT EXISTS users (
    -- UUID primary key (PostgreSQL's gen_random_uuid() generates UUIDs)
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Email must be unique and not null
    -- VARCHAR(255) allows up to 255 characters
    email VARCHAR(255) UNIQUE NOT NULL,
    
    -- Password hash (bcrypt hashes are 60 characters, but we use 255 for safety)
    password_hash VARCHAR(255) NOT NULL,
    
    -- User's display name
    name VARCHAR(255) NOT NULL,
    
    -- Timestamps for auditing
    -- TIMESTAMP stores date and time
    -- DEFAULT NOW() sets the current timestamp when a row is inserted
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- Soft delete support
    -- NULL means the user is active
    -- If set to a timestamp, the user is considered deleted
    deleted_at TIMESTAMP NULL
);

-- Create indexes for better query performance
-- Indexes speed up lookups (like finding a user by email)

-- Index on email (for login lookups)
-- WHERE deleted_at IS NULL means we only index active users
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL;

-- Index on created_at (for sorting/filtering by registration date)
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- Add a comment to the table (documentation)
COMMENT ON TABLE users IS 'Stores user accounts with authentication information';
