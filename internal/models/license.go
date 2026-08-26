package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// License represents a Лицензии и документы card.
type License struct {
	ID         int
	TypeLabel  string
	Title      string
	Description string
	BadgeLabel string
	SortOrder  int
	Active     bool
}

// GetLicenses returns all active licenses ordered by sort_order.
func GetLicenses() ([]License, error) {
	rows, err := db.DB.Query(`SELECT id, type_label, title, description, badge_label, sort_order, active FROM licenses WHERE active=1 ORDER BY sort_order`)
	if err != nil {
		return nil, fmt.Errorf("models: GetLicenses: %w", err)
	}
	defer rows.Close()
	return scanLicenses(rows)
}

// GetAllLicenses returns all licenses (including inactive) for admin.
func GetAllLicenses() ([]License, error) {
	rows, err := db.DB.Query(`SELECT id, type_label, title, description, badge_label, sort_order, active FROM licenses ORDER BY sort_order`)
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
		if err := rows.Scan(&l.ID, &l.TypeLabel, &l.Title, &l.Description, &l.BadgeLabel, &l.SortOrder, &active); err != nil {
			return nil, err
		}
		l.Active = active == 1
		out = append(out, l)
	}
	return out, rows.Err()
}

// AddLicense inserts a new license.
func AddLicense(typeLabel, title, description, badgeLabel string) error {
	_, err := db.DB.Exec(`INSERT INTO licenses (type_label, title, description, badge_label, sort_order) VALUES (?, ?, ?, ?, (SELECT COALESCE(MAX(sort_order),0)+1 FROM licenses))`, typeLabel, title, description, badgeLabel)
	return err
}

// UpdateLicense updates a license.
func UpdateLicense(id int, typeLabel, title, description, badgeLabel string, active bool) error {
	a := 0
	if active {
		a = 1
	}
	_, err := db.DB.Exec(`UPDATE licenses SET type_label=?, title=?, description=?, badge_label=?, active=? WHERE id=?`, typeLabel, title, description, badgeLabel, a, id)
	return err
}

// DeleteLicense removes a license by ID.
func DeleteLicense(id int) error {
	_, err := db.DB.Exec(`DELETE FROM licenses WHERE id=?`, id)
	return err
}
