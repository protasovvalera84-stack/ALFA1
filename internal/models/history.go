package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// HistoryEvent represents a timeline event in the company history section.
type HistoryEvent struct {
	ID          int
	YearLabel   string
	Title       string
	Description string
	Quote       string
	SortOrder   int
	Active      bool
}

// GetHistoryEvents returns all active history events.
func GetHistoryEvents() ([]HistoryEvent, error) {
	rows, err := db.DB.Query(
		`SELECT id, year_label, title, description, quote, sort_order, active
		 FROM history_events WHERE active = 1 ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetHistoryEvents: %w", err)
	}
	defer rows.Close()
	return scanHistoryEvents(rows)
}

// GetAllHistoryEvents returns all events for the admin panel.
func GetAllHistoryEvents() ([]HistoryEvent, error) {
	rows, err := db.DB.Query(
		`SELECT id, year_label, title, description, quote, sort_order, active
		 FROM history_events ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetAllHistoryEvents: %w", err)
	}
	defer rows.Close()
	return scanHistoryEvents(rows)
}

func scanHistoryEvents(rows interface{ Next() bool; Scan(...any) error; Err() error }) ([]HistoryEvent, error) {
	var out []HistoryEvent
	for rows.Next() {
		var e HistoryEvent
		var active int
		if err := rows.Scan(&e.ID, &e.YearLabel, &e.Title, &e.Description, &e.Quote, &e.SortOrder, &active); err != nil {
			return nil, err
		}
		e.Active = active == 1
		out = append(out, e)
	}
	return out, rows.Err()
}

// AddHistoryEvent inserts a new history event.
func AddHistoryEvent(yearLabel, title, description, quote string) error {
	_, err := db.DB.Exec(
		`INSERT INTO history_events (year_label, title, description, quote, sort_order)
		 VALUES (?, ?, ?, ?, (SELECT COALESCE(MAX(sort_order)+1, 0) FROM history_events))`,
		yearLabel, title, description, quote,
	)
	return err
}

// UpdateHistoryEvent updates an existing history event.
func UpdateHistoryEvent(id int, yearLabel, title, description, quote string, active bool) error {
	a := 0
	if active {
		a = 1
	}
	_, err := db.DB.Exec(
		`UPDATE history_events SET year_label=?, title=?, description=?, quote=?, active=? WHERE id=?`,
		yearLabel, title, description, quote, a, id,
	)
	return err
}

// DeleteHistoryEvent removes a history event.
func DeleteHistoryEvent(id int) error {
	_, err := db.DB.Exec(`DELETE FROM history_events WHERE id=?`, id)
	return err
}
