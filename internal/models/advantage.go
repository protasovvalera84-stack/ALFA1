package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// Advantage represents a single "why choose us" card.
type Advantage struct {
	ID          int
	Title       string
	Description string
	SortOrder   int
}

// GetAdvantages returns all advantages ordered by sort_order.
func GetAdvantages() ([]Advantage, error) {
	rows, err := db.DB.Query(
		`SELECT id, title, description, sort_order FROM advantages ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetAdvantages: %w", err)
	}
	defer rows.Close()
	var items []Advantage
	for rows.Next() {
		var a Advantage
		if err := rows.Scan(&a.ID, &a.Title, &a.Description, &a.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

// CreateAdvantage inserts a new advantage.
func CreateAdvantage(title, description string) error {
	_, err := db.DB.Exec(
		`INSERT INTO advantages (title, description, sort_order)
		 VALUES (?, ?, (SELECT COALESCE(MAX(sort_order)+1, 0) FROM advantages))`,
		title, description,
	)
	return err
}

// UpdateAdvantage updates an existing advantage.
func UpdateAdvantage(id int, title, description string) error {
	_, err := db.DB.Exec(
		`UPDATE advantages SET title=?, description=? WHERE id=?`,
		title, description, id,
	)
	return err
}

// DeleteAdvantage removes an advantage by ID.
func DeleteAdvantage(id int) error {
	_, err := db.DB.Exec(`DELETE FROM advantages WHERE id=?`, id)
	return err
}
