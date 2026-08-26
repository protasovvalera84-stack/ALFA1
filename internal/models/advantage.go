package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// Advantage represents a "Почему выбирают нас" card.
type Advantage struct {
	ID          int
	Title       string
	Description string
	SortOrder   int
	Active      bool
}

// GetAdvantages returns all active advantages ordered by sort_order.
func GetAdvantages() ([]Advantage, error) {
	rows, err := db.DB.Query(`SELECT id, title, description, sort_order, active FROM advantages WHERE active=1 ORDER BY sort_order`)
	if err != nil {
		return nil, fmt.Errorf("models: GetAdvantages: %w", err)
	}
	defer rows.Close()
	return scanAdvantages(rows)
}

// GetAllAdvantages returns all advantages (including inactive) for admin.
func GetAllAdvantages() ([]Advantage, error) {
	rows, err := db.DB.Query(`SELECT id, title, description, sort_order, active FROM advantages ORDER BY sort_order`)
	if err != nil {
		return nil, fmt.Errorf("models: GetAllAdvantages: %w", err)
	}
	defer rows.Close()
	return scanAdvantages(rows)
}

func scanAdvantages(rows interface{ Next() bool; Scan(...any) error; Err() error }) ([]Advantage, error) {
	var out []Advantage
	for rows.Next() {
		var a Advantage
		var active int
		if err := rows.Scan(&a.ID, &a.Title, &a.Description, &a.SortOrder, &active); err != nil {
			return nil, err
		}
		a.Active = active == 1
		out = append(out, a)
	}
	return out, rows.Err()
}

// AddAdvantage inserts a new advantage.
func AddAdvantage(title, description string) error {
	_, err := db.DB.Exec(`INSERT INTO advantages (title, description, sort_order) VALUES (?, ?, (SELECT COALESCE(MAX(sort_order),0)+1 FROM advantages))`, title, description)
	return err
}

// UpdateAdvantage updates an existing advantage.
func UpdateAdvantage(id int, title, description string, active bool) error {
	a := 0
	if active {
		a = 1
	}
	_, err := db.DB.Exec(`UPDATE advantages SET title=?, description=?, active=? WHERE id=?`, title, description, a, id)
	return err
}

// DeleteAdvantage removes an advantage by ID.
func DeleteAdvantage(id int) error {
	_, err := db.DB.Exec(`DELETE FROM advantages WHERE id=?`, id)
	return err
}
