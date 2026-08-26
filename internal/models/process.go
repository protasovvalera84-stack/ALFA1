package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// ProcessStep represents a Как мы работаем step.
type ProcessStep struct {
	ID          int
	StepNum     string
	Title       string
	Description string
	SortOrder   int
}

// GetProcessSteps returns all steps ordered by sort_order.
func GetProcessSteps() ([]ProcessStep, error) {
	rows, err := db.DB.Query(`SELECT id, step_num, title, description, sort_order FROM process_steps ORDER BY sort_order`)
	if err != nil {
		return nil, fmt.Errorf("models: GetProcessSteps: %w", err)
	}
	defer rows.Close()
	var out []ProcessStep
	for rows.Next() {
		var s ProcessStep
		if err := rows.Scan(&s.ID, &s.StepNum, &s.Title, &s.Description, &s.SortOrder); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// AddProcessStep inserts a new step.
func AddProcessStep(stepNum, title, description string) error {
	_, err := db.DB.Exec(`INSERT INTO process_steps (step_num, title, description, sort_order) VALUES (?, ?, ?, (SELECT COALESCE(MAX(sort_order),0)+1 FROM process_steps))`, stepNum, title, description)
	return err
}

// UpdateProcessStep updates a step.
func UpdateProcessStep(id int, stepNum, title, description string) error {
	_, err := db.DB.Exec(`UPDATE process_steps SET step_num=?, title=?, description=? WHERE id=?`, stepNum, title, description, id)
	return err
}

// DeleteProcessStep removes a step by ID.
func DeleteProcessStep(id int) error {
	_, err := db.DB.Exec(`DELETE FROM process_steps WHERE id=?`, id)
	return err
}
