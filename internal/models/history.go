package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// HistoryItem represents a single entry on the company history timeline.
type HistoryItem struct {
	ID          int
	Year        string
	Title       string
	Subtitle    string
	Description string
	Quote       string
	SortOrder   int
}

// GetHistoryItems returns all history items ordered by sort_order.
func GetHistoryItems() ([]HistoryItem, error) {
	rows, err := db.DB.Query(
		`SELECT id, year, title, subtitle, description, quote, sort_order
		 FROM history_items ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetHistoryItems: %w", err)
	}
	defer rows.Close()
	var items []HistoryItem
	for rows.Next() {
		var h HistoryItem
		if err := rows.Scan(&h.ID, &h.Year, &h.Title, &h.Subtitle, &h.Description, &h.Quote, &h.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, h)
	}
	return items, rows.Err()
}

// CreateHistoryItem inserts a new history item.
func CreateHistoryItem(year, title, subtitle, description, quote string) error {
	_, err := db.DB.Exec(
		`INSERT INTO history_items (year, title, subtitle, description, quote, sort_order)
		 VALUES (?, ?, ?, ?, ?, (SELECT COALESCE(MAX(sort_order)+1, 0) FROM history_items))`,
		year, title, subtitle, description, quote,
	)
	return err
}

// UpdateHistoryItem updates an existing history item.
func UpdateHistoryItem(id int, year, title, subtitle, description, quote string) error {
	_, err := db.DB.Exec(
		`UPDATE history_items SET year=?, title=?, subtitle=?, description=?, quote=? WHERE id=?`,
		year, title, subtitle, description, quote, id,
	)
	return err
}

// DeleteHistoryItem removes a history item by ID.
func DeleteHistoryItem(id int) error {
	_, err := db.DB.Exec(`DELETE FROM history_items WHERE id=?`, id)
	return err
}
