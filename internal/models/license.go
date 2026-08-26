package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// License represents a license/certificate card on the site.
type License struct {
	ID          int
	TypeLabel   string
	Company     string
	Description string
	StatusText  string
	SortOrder   int
	Active      bool
}

// GetLicenses returns all active licenses.
func GetLicenses() ([]License, error) {
	rows, err := db.DB.Query(
		`SELECT id, type_label, company, description, status_text, sort_order, active
		 FROM licenses WHERE active = 1 ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetLicenses: %w", err)
	}
	defer rows.Close()
	return scanLicenses(rows)
}

// GetAllLicenses returns all licenses for the admin panel.
func GetAllLicenses() ([]License, error) {
	rows, err := db.DB.Query(
		`SELECT id, type_label, company, description, status_text, sort_order, active
		 FROM licenses ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetAllLicenses: %w", err)
	}
	defer rows.Close()
	return scanLicenses(rows)
}

func scanLicenses(rows interface{ Next() bool; Scan(...any) error; Err() error }) ([]License, error) {
	var out []License
	for rows.Next() {
		var l License
		var active int
		if err := rows.Scan(&l.ID, &l.TypeLabel, &l.Company, &l.Description, &l.StatusText, &l.SortOrder, &active); err != nil {
			return nil, err
		}
		l.Active = active == 1
		out = append(out, l)
	}
	return out, rows.Err()
}

// AddLicense inserts a new license.
func AddLicense(typeLabel, company, description, statusText string) error {
	_, err := db.DB.Exec(
		`INSERT INTO licenses (type_label, company, description, status_text, sort_order)
		 VALUES (?, ?, ?, ?, (SELECT COALESCE(MAX(sort_order)+1, 0) FROM licenses))`,
		typeLabel, company, description, statusText,
	)
	return err
}

// UpdateLicense updates an existing license.
func UpdateLicense(id int, typeLabel, company, description, statusText string, active bool) error {
	a := 0
	if active {
		a = 1
	}
	_, err := db.DB.Exec(
		`UPDATE licenses SET type_label=?, company=?, description=?, status_text=?, active=? WHERE id=?`,
		typeLabel, company, description, statusText, a, id,
	)
	return err
}

// DeleteLicense removes a license.
func DeleteLicense(id int) error {
	_, err := db.DB.Exec(`DELETE FROM licenses WHERE id=?`, id)
	return err
}
