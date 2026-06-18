package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// SEOPage holds the per-page SEO configuration.
type SEOPage struct {
	Slug        string
	Title       string
	Description string
	OGImage     string
	SchemaJSON  string
}

// GetSEOPage returns the SEO settings for a given URL slug.
// If no record is found, a default struct is returned.
func GetSEOPage(slug string) (*SEOPage, error) {
	var p SEOPage
	err := db.DB.QueryRow(
		`SELECT slug, title, description, og_image, schema_json FROM seo_pages WHERE slug = ?`, slug,
	).Scan(&p.Slug, &p.Title, &p.Description, &p.OGImage, &p.SchemaJSON)
	if err != nil {
		// Return empty defaults rather than an error for missing pages.
		return &SEOPage{Slug: slug}, nil
	}
	return &p, nil
}

// GetAllSEOPages returns all configured SEO pages.
func GetAllSEOPages() ([]SEOPage, error) {
	rows, err := db.DB.Query(
		`SELECT slug, title, description, og_image, schema_json FROM seo_pages ORDER BY slug`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetAllSEOPages: %w", err)
	}
	defer rows.Close()

	var pages []SEOPage
	for rows.Next() {
		var p SEOPage
		if err := rows.Scan(&p.Slug, &p.Title, &p.Description, &p.OGImage, &p.SchemaJSON); err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}
	return pages, rows.Err()
}

// UpsertSEOPage creates or updates a SEO page record.
func UpsertSEOPage(p *SEOPage) error {
	_, err := db.DB.Exec(
		`INSERT INTO seo_pages (slug, title, description, og_image, schema_json) VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(slug) DO UPDATE SET
		   title       = excluded.title,
		   description = excluded.description,
		   og_image    = excluded.og_image,
		   schema_json = excluded.schema_json`,
		p.Slug, p.Title, p.Description, p.OGImage, p.SchemaJSON,
	)
	return err
}
