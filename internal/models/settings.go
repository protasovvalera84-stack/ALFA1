// Package models provides data access layer functions for the site's content.
package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// SiteSettings holds all key-value settings loaded from the database.
type SiteSettings struct {
	SiteTitle       string
	SiteDescription string
	CompanyName     string
	Phone1          string
	Phone2          string
	PhoneHR         string
	Email           string
	Address         string
	Address2        string
	WorkingHours    string
	FoundedYear     string
	HeroTitle       string
	HeroSubtitle    string
	AboutText       string
	StatsYears      string
	StatsObjects    string
	StatsStaff      string
	StatsLicenses   string
	RobotsTxt       string
	Domain          string
}

// GetSettings loads all settings from the database and returns a filled struct.
func GetSettings() (*SiteSettings, error) {
	rows, err := db.DB.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return nil, fmt.Errorf("models: GetSettings: %w", err)
	}
	defer rows.Close()

	m := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		m[k] = v
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &SiteSettings{
		SiteTitle:       m["site_title"],
		SiteDescription: m["site_description"],
		CompanyName:     m["company_name"],
		Phone1:          m["phone1"],
		Phone2:          m["phone2"],
		PhoneHR:         m["phone_hr"],
		Email:           m["email"],
		Address:         m["address"],
		Address2:        m["address2"],
		WorkingHours:    m["working_hours"],
		FoundedYear:     m["founded_year"],
		HeroTitle:       m["hero_title"],
		HeroSubtitle:    m["hero_subtitle"],
		AboutText:       m["about_text"],
		StatsYears:      m["stats_years"],
		StatsObjects:    m["stats_objects"],
		StatsStaff:      m["stats_staff"],
		StatsLicenses:   m["stats_licenses"],
		RobotsTxt:       m["robots_txt"],
		Domain:          m["domain"],
	}, nil
}

// GetSetting returns a single setting value by key.
func GetSetting(key string) string {
	var value string
	_ = db.DB.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	return value
}

// SetSetting upserts a single setting key/value pair.
func SetSetting(key, value string) error {
	_, err := db.DB.Exec(
		`INSERT INTO settings (key, value) VALUES (?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}

// SetSettings bulk-upserts a map of key/value pairs.
func SetSettings(m map[string]string) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare(`INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for k, v := range m {
		if _, err := stmt.Exec(k, v); err != nil {
			return err
		}
	}
	return tx.Commit()
}
