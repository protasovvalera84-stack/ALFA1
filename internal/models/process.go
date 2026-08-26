package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// ProcessStep represents a step in "how we work" section.
type ProcessStep struct {
	ID          int
	StepNum     string
	Title       string
	Description string
	SortOrder   int
	Active      bool
}

// GetProcessSteps returns all active process steps.
func GetProcessSteps() ([]ProcessStep, error) {
	rows, err := db.DB.Query(
		`SELECT id, step_num, title, description, sort_order, active
		 FROM process_steps WHERE active = 1 ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetProcessSteps: %w", err)
	}
	defer rows.Close()
	return scanProcessSteps(rows)
}

// GetAllProcessSteps returns all steps for the admin panel.
func GetAllProcessSteps() ([]ProcessStep, error) {
	rows, err := db.DB.Query(
		`SELECT id, step_num, title, description, sort_order, active
		 FROM process_steps ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetAllProcessSteps: %w", err)
	}
	defer rows.Close()
	return scanProcessSteps(rows)
}

func scanProcessSteps(rows interface{ Next() bool; Scan(...any) error; Err() error }) ([]ProcessStep, error) {
	var out []ProcessStep
	for rows.Next() {
		var s ProcessStep
		var active int
		if err := rows.Scan(&s.ID, &s.StepNum, &s.Title, &s.Description, &s.SortOrder, &active); err != nil {
			return nil, err
		}
		s.Active = active == 1
		out = append(out, s)
	}
	return out, rows.Err()
}

// AddProcessStep inserts a new process step.
func AddProcessStep(stepNum, title, description string) error {
	_, err := db.DB.Exec(
		`INSERT INTO process_steps (step_num, title, description, sort_order)
		 VALUES (?, ?, ?, (SELECT COALESCE(MAX(sort_order)+1, 0) FROM process_steps))`,
		stepNum, title, description,
	)
	return err
}

// UpdateProcessStep updates an existing step.
func UpdateProcessStep(id int, stepNum, title, description string, active bool) error {
	a := 0
	if active {
		a = 1
	}
	_, err := db.DB.Exec(
		`UPDATE process_steps SET step_num=?, title=?, description=?, active=? WHERE id=?`,
		stepNum, title, description, a, id,
	)
	return err
}

// DeleteProcessStep removes a step.
func DeleteProcessStep(id int) error {
	_, err := db.DB.Exec(`DELETE FROM process_steps WHERE id=?`, id)
	return err
}
