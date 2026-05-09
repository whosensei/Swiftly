-- 002_features.sql
-- New features: Tags, Metadata, Password-protected links, Custom aliases

-- Add metadata columns to urls table
ALTER TABLE urls ADD COLUMN IF NOT EXISTS title TEXT;
ALTER TABLE urls ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE urls ADD COLUMN IF NOT EXISTS og_image TEXT;

-- Add password protection
ALTER TABLE urls ADD COLUMN IF NOT EXISTS password_hash TEXT;

-- Tags table
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_tags_user_id ON tags(user_id);

-- URL-Tags junction table (many-to-many)
CREATE TABLE IF NOT EXISTS url_tags (
    url_id UUID REFERENCES urls(id) ON DELETE CASCADE,
    tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (url_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_url_tags_url_id ON url_tags(url_id);
CREATE INDEX IF NOT EXISTS idx_url_tags_tag_id ON url_tags(tag_id);
