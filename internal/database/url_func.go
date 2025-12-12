package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github/whosensei/shortenn/internal/model"
	"log"
	"time"
)

func Redirect(db *sql.DB, shorturl string) (string, string, sql.NullTime) {
	var u string
	var url_id string
	var expires_at sql.NullTime
	query := `SELECT long_url,id,expires_at FROM urls WHERE short_code = $1`
	if err := db.QueryRow(query, shorturl).Scan(&u, &url_id, &expires_at); err != nil {
		fmt.Println("Failed to fetch", err)
	}
	return u, url_id, expires_at
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
		log.Fatal("failed to find uuid")
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

func Add_authenticated_url(db *sql.DB, short_code string, long_url string, userID string, ip_address string) error {
	query := `INSERT INTO urls (short_code, long_url, user_id, ip_address) VALUES ($1,$2,$3,$4)`
	_, err := db.Exec(query, short_code, long_url, userID, ip_address)
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
			(SELECT COUNT(*) FROM clicks c WHERE c.url_id = urls.id) AS clicks
		FROM urls
		WHERE anonymous_token=$1
		  AND user_id is NULL
		  AND (expires_at is NULL OR expires_at > NOW())
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, anonymous_token)
	if err != nil {
		log.Fatal("Failed to get urls")
	}

	for rows.Next() {
		var data model.URL
		var dbClicks int64
		rows.Scan(&data.Id, &data.Short_code, &data.Long_url, &data.Created_At, &data.Expires_at, &dbClicks)
		data.Clicks = dbClicks
		u = append(u, data)
	}
	return u, nil

}

func Get_auth_urls(db *sql.DB, user_id string) ([]model.URL, error) {

	query := `
		SELECT
			id,
			short_code,
			long_url,
			created_at,
			(SELECT COUNT(*) FROM clicks c WHERE c.url_id = urls.id) AS clicks
		FROM urls
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.Query(query, user_id)
	if err != nil {
		log.Fatal("Failed to get urls")
	}
	var u []model.URL

	for rows.Next() {
		var data model.URL
		var dbClicks int64
		rows.Scan(&data.Id, &data.Short_code, &data.Long_url, &data.Created_At, &dbClicks)
		data.Clicks = dbClicks
		u = append(u, data)
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

// to get url_id from short_code
// func Get_url_id_from_shortcode(db *sql.DB, short_code string) (string, error) {
// 	var url_id string
// 	query := `SELECT id FROM urls WHERE short_code=$1`
// 	err := db.QueryRow(query, short_code).Scan(&url_id)
// 	if err != nil {
// 		return "", err
// 	}
// 	return url_id, nil
// }

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
