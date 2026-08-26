package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// ProcessStep represents a single step in the "How We Work" section.
type ProcessStep struct {
	ID          int
	StepNum     string
	Title       string
	Description string
	SortOrder   int
}

// GetProcessSteps returns all process steps ordered by sort_order.
func GetProcessSteps() ([]ProcessStep, error) {
	rows, err := db.DB.Query(
		`SELECT id, step_num, title, description, sort_order
		 FROM process_steps ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetProcessSteps: %w", err)
	}
	defer rows.Close()
	var items []ProcessStep
	for rows.Next() {
		var p ProcessStep
		if err := rows.Scan(&p.ID, &p.StepNum, &p.Title, &p.Description, &p.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, p)
	}
	return items, rows.Err()
}

// CreateProcessStep inserts a new process step.
func CreateProcessStep(stepNum, title, description string) error {
	_, err := db.DB.Exec(
		`INSERT INTO process_steps (step_num, title, description, sort_order)
		 VALUES (?, ?, ?, (SELECT COALESCE(MAX(sort_order)+1, 0) FROM process_steps))`,
		stepNum, title, description,
	)
	return err
}

// UpdateProcessStep updates an existing process step.
func UpdateProcessStep(id int, stepNum, title, description string) error {
	_, err := db.DB.Exec(
		`UPDATE process_steps SET step_num=?, title=?, description=? WHERE id=?`,
		stepNum, title, description, id,
	)
	return err
}

// DeleteProcessStep removes a process step by ID.
func DeleteProcessStep(id int) error {
	_, err := db.DB.Exec(`DELETE FROM process_steps WHERE id=?`, id)
	return err
}
