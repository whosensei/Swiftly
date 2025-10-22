CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table (minimal, synced from Better Auth)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    auth_user_id VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255),
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_users_auth_id ON users(auth_user_id);

-- URLs table
CREATE TABLE urls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    short_code VARCHAR(12) UNIQUE NOT NULL,
    long_url TEXT NOT NULL,
    
    -- User tracking
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    anonymous_token VARCHAR(255),
    
    -- Metadata
    ip_address VARCHAR(45),
    user_agent TEXT,
    
    -- Timestamps & TTL
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    
    -- Analytics (cached from Redis, periodically synced)
    clicks INTEGER DEFAULT 0,
    last_clicked_at TIMESTAMP
);

CREATE INDEX idx_urls_short_code ON urls(short_code);
CREATE INDEX idx_urls_user_id ON urls(user_id);
CREATE INDEX idx_urls_anonymous_token ON urls(anonymous_token);
CREATE INDEX idx_urls_expires_at ON urls(expires_at);
CREATE INDEX idx_urls_created_at ON urls(created_at DESC);

--Click events table (detailed analytics)
CREATE TABLE clicks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    url_id UUID REFERENCES urls(id) ON DELETE CASCADE,
    
    -- When & Where
    clicked_at TIMESTAMP DEFAULT NOW(),
    ip_address VARCHAR(45),
    country VARCHAR(2),
    city VARCHAR(100),
    
    -- How
    user_agent TEXT,
    referer TEXT,
    device_type VARCHAR(50),
    browser VARCHAR(50),
    os VARCHAR(50)
);

CREATE INDEX idx_clicks_url_id ON clicks(url_id);
CREATE INDEX idx_clicks_clicked_at ON clicks(clicked_at DESC);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for urls table
CREATE TRIGGER update_urls_updated_at BEFORE UPDATE ON urls
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Trigger for users table
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
