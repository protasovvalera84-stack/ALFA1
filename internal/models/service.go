package models

import (
	"fmt"
	"html/template"

	"alfaunit1/internal/db"
)

// Service represents a single security service offered by the company.
type Service struct {
	ID          int
	Name        string
	Description string
	Icon        template.HTML // raw SVG, safe to render
	SortOrder   int
	Active      bool
}

// GetServices returns all active services ordered by sort_order.
func GetServices() ([]Service, error) {
	rows, err := db.DB.Query(
		`SELECT id, name, description, icon, sort_order, active
		 FROM services WHERE active = 1 ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetServices: %w", err)
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var s Service
		var active int
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Icon, &s.SortOrder, &active); err != nil {
			return nil, err
		}
		s.Active = active == 1
		services = append(services, s)
	}
	return services, rows.Err()
}

// GetAllServices returns all services (including inactive) for the admin panel.
func GetAllServices() ([]Service, error) {
	rows, err := db.DB.Query(
		`SELECT id, name, description, icon, sort_order, active
		 FROM services ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetAllServices: %w", err)
	}
	defer rows.Close()

	var services []Service
	for rows.Next() {
		var s Service
		var active int
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Icon, &s.SortOrder, &active); err != nil {
			return nil, err
		}
		s.Active = active == 1
		services = append(services, s)
	}
	return services, rows.Err()
}

// UpdateService updates an existing service record.
func UpdateService(id int, name, description string, active bool) error {
	activeInt := 0
	if active {
		activeInt = 1
	}
	_, err := db.DB.Exec(
		`UPDATE services SET name=?, description=?, active=? WHERE id=?`,
		name, description, activeInt, id,
	)
	return err
}
