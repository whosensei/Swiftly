// +build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	godotenv.Load()
	
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal("Failed to connect:", err)
	}
	defer db.Close()

	migration := `
ALTER TABLE urls ADD COLUMN IF NOT EXISTS title TEXT;
ALTER TABLE urls ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE urls ADD COLUMN IF NOT EXISTS og_image TEXT;
ALTER TABLE urls ADD COLUMN IF NOT EXISTS password_hash TEXT;

CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_tags_user_id ON tags(user_id);

CREATE TABLE IF NOT EXISTS url_tags (
    url_id UUID REFERENCES urls(id) ON DELETE CASCADE,
    tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (url_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_url_tags_url_id ON url_tags(url_id);
CREATE INDEX IF NOT EXISTS idx_url_tags_tag_id ON url_tags(tag_id);
`

	_, err = db.Exec(migration)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}

	fmt.Println("✅ Migration completed successfully!")
}
