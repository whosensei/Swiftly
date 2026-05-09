package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github/whosensei/shortenn/internal/model"
	"log"
	"time"
)

func Redirect(db *sql.DB, shorturl string) (model.RedirectInfo, sql.NullTime) {
	var info model.RedirectInfo
	var expires_at sql.NullTime
	var passwordHash sql.NullString

	query := `SELECT long_url, id, expires_at, password_hash FROM urls WHERE short_code = $1`
	if err := db.QueryRow(query, shorturl).Scan(&info.LongURL, &info.UrlID, &expires_at, &passwordHash); err != nil {
		log.Printf("Failed to fetch redirect info: %v", err)
		return info, expires_at
	}

	if passwordHash.Valid && passwordHash.String != "" {
		info.HasPassword = true
		info.PasswordHash = passwordHash.String
	}

	return info, expires_at
}

func Add_anon_url(db *sql.DB, short_code string, long_url string, anonymous_token string, ip_address string, expires_at time.Time) error {
	query := `INSERT INTO urls (short_code,long_url,anonymous_token,ip_address,expires_at) VALUES ($1,$2,$3,$4,$5)`
	_, err := db.Exec(query, short_code, long_url, anonymous_token, ip_address, expires_at)
	return err
}

func Find_uuid_from_UserID(db *sql.DB, userID string) string {

	var u string
	query := `SELECT id FROM users WHERE auth_user_id = $1`
	if err := db.QueryRow(query, userID).Scan(&u); err != nil {
		log.Printf("Failed to find uuid: %v", err)
	}

	return u

}

func EnsureUserExists(db *sql.DB, authUserID string, email string, name string) (string, error) {
	var userUUID string

	query := `
		INSERT INTO users (auth_user_id, email, name, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		ON CONFLICT (auth_user_id)
		DO UPDATE SET
			email = EXCLUDED.email,
			name = EXCLUDED.name,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`

	err := db.QueryRow(query, authUserID, email, name).Scan(&userUUID)
	if err != nil {
		return "", fmt.Errorf("failed to ensure user exists: %w", err)
	}

	return userUUID, nil
}

func Add_authenticated_url(db *sql.DB, short_code string, long_url string, userID string, ip_address string, expiresAt *time.Time, passwordHash string, title string, description string, ogImage string) error {
	query := `INSERT INTO urls (short_code, long_url, user_id, ip_address, expires_at, password_hash, title, description, og_image) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`

	var expiry interface{}
	if expiresAt != nil {
		expiry = *expiresAt
	}

	var pwHash interface{}
	if passwordHash != "" {
		pwHash = passwordHash
	}

	var t, d, o interface{}
	if title != "" {
		t = title
	}
	if description != "" {
		d = description
	}
	if ogImage != "" {
		o = ogImage
	}

	_, err := db.Exec(query, short_code, long_url, userID, ip_address, expiry, pwHash, t, d, o)
	return err
}

// CheckShortCodeExists checks if a short code is already taken
func CheckShortCodeExists(db *sql.DB, shortCode string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM urls WHERE short_code = $1`
	err := db.QueryRow(query, shortCode).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateURL updates the destination URL for a short code
func UpdateURL(db *sql.DB, shortCode string, newLongURL string, userID string) error {
	query := `UPDATE urls SET long_url = $1, updated_at = NOW() WHERE short_code = $2 AND user_id = $3`
	result, err := db.Exec(query, newLongURL, shortCode, userID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("url not found or not owned by user")
	}
	return nil
}

// FullUpdateURL updates all editable fields of a URL
func FullUpdateURL(db *sql.DB, shortCode string, userID string, longURL *string, newSlug *string, passwordHash *string, removePassword bool, expiresAt *time.Time, removeExpiry bool) error {
	// Build dynamic SET clause
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if newSlug != nil && *newSlug != "" && *newSlug != shortCode {
		setClauses = append(setClauses, fmt.Sprintf("short_code = $%d", argIdx))
		args = append(args, *newSlug)
		argIdx++
	}

	if longURL != nil && *longURL != "" {
		setClauses = append(setClauses, fmt.Sprintf("long_url = $%d", argIdx))
		args = append(args, *longURL)
		argIdx++
	}

	if removePassword {
		setClauses = append(setClauses, fmt.Sprintf("password_hash = $%d", argIdx))
		args = append(args, nil)
		argIdx++
	} else if passwordHash != nil && *passwordHash != "" {
		setClauses = append(setClauses, fmt.Sprintf("password_hash = $%d", argIdx))
		args = append(args, *passwordHash)
		argIdx++
	}

	if removeExpiry {
		setClauses = append(setClauses, fmt.Sprintf("expires_at = $%d", argIdx))
		args = append(args, nil)
		argIdx++
	} else if expiresAt != nil {
		setClauses = append(setClauses, fmt.Sprintf("expires_at = $%d", argIdx))
		args = append(args, *expiresAt)
		argIdx++
	}

	query := fmt.Sprintf("UPDATE urls SET %s WHERE short_code = $%d AND user_id = $%d",
		joinStrings(setClauses, ", "), argIdx, argIdx+1)
	args = append(args, shortCode, userID)

	result, err := db.Exec(query, args...)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("url not found or not owned by user")
	}
	return nil
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// ClearTagsForURL removes all tag associations for a URL
func ClearTagsForURL(db *sql.DB, urlID string) error {
	_, err := db.Exec(`DELETE FROM url_tags WHERE url_id = $1`, urlID)
	return err
}

// UpdateURLMetadata updates the metadata for a URL
func UpdateURLMetadata(db *sql.DB, shortCode string, title string, description string, ogImage string) error {
	query := `UPDATE urls SET title = $1, description = $2, og_image = $3, updated_at = NOW() WHERE short_code = $4`
	_, err := db.Exec(query, title, description, ogImage, shortCode)
	return err
}

func Get_anon_urls(db *sql.DB, anonymous_token string) ([]model.URL, error) {

	var u []model.URL

	query := `
		SELECT
			id,
			short_code,
			long_url,
			created_at,
			expires_at,
			COALESCE(title, '') as title,
			COALESCE(description, '') as description,
			COALESCE(og_image, '') as og_image,
			(password_hash IS NOT NULL AND password_hash != '') as has_password,
			(SELECT COUNT(*) FROM clicks c WHERE c.url_id = urls.id) AS clicks
		FROM urls
		WHERE anonymous_token=$1
		  AND user_id is NULL
		  AND (expires_at is NULL OR expires_at > NOW())
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, anonymous_token)
	if err != nil {
		log.Printf("Failed to get anonymous urls: %v", err)
		return nil, fmt.Errorf("failed to get urls: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var data model.URL
		var dbClicks int64
		if err := rows.Scan(&data.Id, &data.Short_code, &data.Long_url, &data.Created_At, &data.Expires_at, &data.Title, &data.Description, &data.Og_image, &data.Has_password, &dbClicks); err != nil {
			log.Printf("Failed to scan anonymous url row: %v", err)
			continue
		}
		data.Clicks = dbClicks
		u = append(u, data)
	}
	return u, nil

}

func Get_auth_urls(db *sql.DB, user_id string) ([]model.URL, error) {

	query := `
		SELECT
			urls.id,
			short_code,
			long_url,
			urls.created_at,
			COALESCE(title, '') as title,
			COALESCE(description, '') as description,
			COALESCE(og_image, '') as og_image,
			(password_hash IS NOT NULL AND password_hash != '') as has_password,
			expires_at,
			(SELECT COUNT(*) FROM clicks c WHERE c.url_id = urls.id) AS clicks
		FROM urls
		WHERE user_id = $1
		ORDER BY urls.created_at DESC
	`

	rows, err := db.Query(query, user_id)
	if err != nil {
		log.Printf("Failed to get authenticated urls: %v", err)
		return nil, fmt.Errorf("failed to get urls: %w", err)
	}
	defer rows.Close()

	var u []model.URL

	for rows.Next() {
		var data model.URL
		var dbClicks int64
		var expiresAt sql.NullTime
		if err := rows.Scan(&data.Id, &data.Short_code, &data.Long_url, &data.Created_At, &data.Title, &data.Description, &data.Og_image, &data.Has_password, &expiresAt, &dbClicks); err != nil {
			log.Printf("Failed to scan authenticated url row: %v", err)
			continue
		}
		if expiresAt.Valid {
			data.Expires_at = expiresAt.Time
		}
		data.Clicks = dbClicks
		u = append(u, data)
	}

	// Load tags for each URL
	for i := range u {
		tags, err := GetTagsForURL(db, u[i].Id)
		if err == nil {
			u[i].Tags = tags
		}
	}

	return u, nil
}

func Delete_url(db *sql.DB, short_code string) error {

	query := `DELETE FROM urls WHERE short_code = $1`
	_, err := db.Exec(query, short_code)
	if err != nil {
		log.Println("Failed to delete")
		return err
	}
	return nil

}

func Verify_anon_url_ownership(db *sql.DB, short_code string, anonymous_token string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM urls WHERE short_code=$1 AND anonymous_token=$2`
	err := db.QueryRow(query, short_code, anonymous_token).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func Verify_auth_url_ownership(db *sql.DB, short_code string, user_id string) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM urls WHERE short_code=$1 AND user_id=$2`
	err := db.QueryRow(query, short_code, user_id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ======= Tag Functions =======

// GetOrCreateTag returns existing tag ID or creates a new one
func GetOrCreateTag(db *sql.DB, userID string, tagName string) (string, error) {
	var tagID string
	query := `
		INSERT INTO tags (user_id, name)
		VALUES ($1, $2)
		ON CONFLICT (user_id, name)
		DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`
	err := db.QueryRow(query, userID, tagName).Scan(&tagID)
	if err != nil {
		return "", fmt.Errorf("failed to get or create tag: %w", err)
	}
	return tagID, nil
}

// LinkTagToURL creates a url_tag association
func LinkTagToURL(db *sql.DB, urlID string, tagID string) error {
	query := `INSERT INTO url_tags (url_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := db.Exec(query, urlID, tagID)
	return err
}

// GetTagsForURL returns all tags for a URL
func GetTagsForURL(db *sql.DB, urlID string) ([]model.Tag, error) {
	query := `
		SELECT t.id, t.name
		FROM tags t
		JOIN url_tags ut ON t.id = ut.tag_id
		WHERE ut.url_id = $1
		ORDER BY t.name
	`
	rows, err := db.Query(query, urlID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var tag model.Tag
		if err := rows.Scan(&tag.Id, &tag.Name); err != nil {
			continue
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// GetUserTags returns all tags for a user
func GetUserTags(db *sql.DB, userID string) ([]model.Tag, error) {
	query := `SELECT id, name FROM tags WHERE user_id = $1 ORDER BY name`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var tag model.Tag
		if err := rows.Scan(&tag.Id, &tag.Name); err != nil {
			continue
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

// DeleteTag deletes a tag and all its associations
func DeleteTag(db *sql.DB, tagID string, userID string) error {
	query := `DELETE FROM tags WHERE id = $1 AND user_id = $2`
	_, err := db.Exec(query, tagID, userID)
	return err
}

// GetURLIDByShortCode returns the url id for a short code
func GetURLIDByShortCode(db *sql.DB, shortCode string) (string, error) {
	var urlID string
	query := `SELECT id FROM urls WHERE short_code = $1`
	err := db.QueryRow(query, shortCode).Scan(&urlID)
	if err != nil {
		return "", err
	}
	return urlID, nil
}

// CleanupExpiredURLs removes expired URLs
func CleanupExpiredURLs(db *sql.DB) (int64, error) {
	result, err := db.Exec(`DELETE FROM urls WHERE expires_at IS NOT NULL AND expires_at < NOW()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func GetAnalyticsBreakdownByShortCode(db *sql.DB, shortCode string) (model.AnalyticsBreakdown, error) {
	// One query that returns grouped counts for each dimension for a short_code.
	// Includes time-series data for the last 7 days.
	query := `
WITH target_url AS (
  SELECT id FROM urls WHERE short_code = $1
),
date_series AS (
  SELECT generate_series(
    CURRENT_DATE - INTERVAL '6 days',
    CURRENT_DATE,
    INTERVAL '1 day'
  )::date AS date
)
SELECT
  COALESCE((
    SELECT jsonb_agg(jsonb_build_object('value', country, 'count', cnt) ORDER BY cnt DESC)
    FROM (
      SELECT COALESCE(NULLIF(country, ''), 'Unknown') AS country, COUNT(*)::bigint AS cnt
      FROM clicks
      WHERE url_id = (SELECT id FROM target_url)
      GROUP BY 1
    ) t
  ), '[]'::jsonb) AS countries,
  COALESCE((
    SELECT jsonb_agg(jsonb_build_object('value', city, 'count', cnt) ORDER BY cnt DESC)
    FROM (
      SELECT COALESCE(NULLIF(city, ''), 'Unknown') AS city, COUNT(*)::bigint AS cnt
      FROM clicks
      WHERE url_id = (SELECT id FROM target_url)
      GROUP BY 1
    ) t
  ), '[]'::jsonb) AS cities,
  COALESCE((
    SELECT jsonb_agg(jsonb_build_object('value', ref, 'count', cnt) ORDER BY cnt DESC)
    FROM (
      SELECT COALESCE(NULLIF(referer, ''), 'Direct') AS ref, COUNT(*)::bigint AS cnt
      FROM clicks
      WHERE url_id = (SELECT id FROM target_url)
      GROUP BY 1
    ) t
  ), '[]'::jsonb) AS referrers,
  COALESCE((
    SELECT jsonb_agg(jsonb_build_object('value', device, 'count', cnt) ORDER BY cnt DESC)
    FROM (
      SELECT COALESCE(NULLIF(device_type, ''), 'Unknown') AS device, COUNT(*)::bigint AS cnt
      FROM clicks
      WHERE url_id = (SELECT id FROM target_url)
      GROUP BY 1
    ) t
  ), '[]'::jsonb) AS devices,
  COALESCE((
    SELECT jsonb_agg(jsonb_build_object('value', browser, 'count', cnt) ORDER BY cnt DESC)
    FROM (
      SELECT COALESCE(NULLIF(browser, ''), 'Unknown') AS browser, COUNT(*)::bigint AS cnt
      FROM clicks
      WHERE url_id = (SELECT id FROM target_url)
      GROUP BY 1
    ) t
  ), '[]'::jsonb) AS browsers,
  COALESCE((
    SELECT jsonb_agg(jsonb_build_object('value', os, 'count', cnt) ORDER BY cnt DESC)
    FROM (
      SELECT COALESCE(NULLIF(os, ''), 'Unknown') AS os, COUNT(*)::bigint AS cnt
      FROM clicks
      WHERE url_id = (SELECT id FROM target_url)
      GROUP BY 1
    ) t
  ), '[]'::jsonb) AS os,
  COALESCE((
    SELECT jsonb_agg(jsonb_build_object('date', to_char(ds.date, 'YYYY-MM-DD'), 'clicks', COALESCE(c.cnt, 0)) ORDER BY ds.date)
    FROM date_series ds
    LEFT JOIN (
      SELECT clicked_at::date AS click_date, COUNT(*)::bigint AS cnt
      FROM clicks
      WHERE url_id = (SELECT id FROM target_url)
        AND clicked_at >= CURRENT_DATE - INTERVAL '6 days'
      GROUP BY 1
    ) c ON ds.date = c.click_date
  ), '[]'::jsonb) AS timeseries,
  (SELECT COUNT(*)::bigint FROM clicks WHERE url_id = (SELECT id FROM target_url)) AS total_clicks,
  (SELECT MAX(clicked_at) FROM clicks WHERE url_id = (SELECT id FROM target_url)) AS last_clicked_at
`

	var (
		countriesJSON  []byte
		citiesJSON     []byte
		referrersJSON  []byte
		devicesJSON    []byte
		browsersJSON   []byte
		osJSON         []byte
		timeseriesJSON []byte
		totalClicks    int64
		lastClickedAt  sql.NullTime
	)

	if err := db.QueryRow(query, shortCode).Scan(
		&countriesJSON,
		&citiesJSON,
		&referrersJSON,
		&devicesJSON,
		&browsersJSON,
		&osJSON,
		&timeseriesJSON,
		&totalClicks,
		&lastClickedAt,
	); err != nil {
		return model.AnalyticsBreakdown{}, err
	}

	var out model.AnalyticsBreakdown
	if err := json.Unmarshal(countriesJSON, &out.Countries); err != nil {
		return model.AnalyticsBreakdown{}, err
	}
	if err := json.Unmarshal(citiesJSON, &out.Cities); err != nil {
		return model.AnalyticsBreakdown{}, err
	}
	if err := json.Unmarshal(referrersJSON, &out.Referrers); err != nil {
		return model.AnalyticsBreakdown{}, err
	}
	if err := json.Unmarshal(devicesJSON, &out.Devices); err != nil {
		return model.AnalyticsBreakdown{}, err
	}
	if err := json.Unmarshal(browsersJSON, &out.Browsers); err != nil {
		return model.AnalyticsBreakdown{}, err
	}
	if err := json.Unmarshal(osJSON, &out.OS); err != nil {
		return model.AnalyticsBreakdown{}, err
	}
	if err := json.Unmarshal(timeseriesJSON, &out.TimeSeries); err != nil {
		return model.AnalyticsBreakdown{}, err
	}

	out.TotalClicks = totalClicks
	if lastClickedAt.Valid {
		t := lastClickedAt.Time
		out.LastClickedAt = &t
	}

	return out, nil
}
