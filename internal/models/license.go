package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// License represents a license or certificate card.
type License struct {
	ID          int
	Title       string
	Subtitle    string
	Description string
	Badge       string
	SortOrder   int
}

// GetLicenses returns all licenses ordered by sort_order.
func GetLicenses() ([]License, error) {
	rows, err := db.DB.Query(
		`SELECT id, title, subtitle, description, badge, sort_order
		 FROM licenses ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetLicenses: %w", err)
	}
	defer rows.Close()
	var items []License
	for rows.Next() {
		var l License
		if err := rows.Scan(&l.ID, &l.Title, &l.Subtitle, &l.Description, &l.Badge, &l.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, l)
	}
	return items, rows.Err()
}

// CreateLicense inserts a new license record.
func CreateLicense(title, subtitle, description, badge string) error {
	_, err := db.DB.Exec(
		`INSERT INTO licenses (title, subtitle, description, badge, sort_order)
		 VALUES (?, ?, ?, ?, (SELECT COALESCE(MAX(sort_order)+1, 0) FROM licenses))`,
		title, subtitle, description, badge,
	)
	return err
}

// UpdateLicense updates an existing license record.
func UpdateLicense(id int, title, subtitle, description, badge string) error {
	_, err := db.DB.Exec(
		`UPDATE licenses SET title=?, subtitle=?, description=?, badge=? WHERE id=?`,
		title, subtitle, description, badge, id,
	)
	return err
}

// DeleteLicense removes a license by ID.
func DeleteLicense(id int) error {
	_, err := db.DB.Exec(`DELETE FROM licenses WHERE id=?`, id)
	return err
}
