package models

import (
	"fmt"
	"strings"

	"alfaunit1/internal/db"
)

// TeamMember represents a team member card on the site.
type TeamMember struct {
	ID          int
	Initial     string
	Name        string
	Role        string
	Department  string
	Description string
	Tags        []string
	SortOrder   int
	Active      bool
}

func scanTeamMember(id int, initial, name, role, department, description, tags string, sortOrder, active int) TeamMember {
	m := TeamMember{
		ID:          id,
		Initial:     initial,
		Name:        name,
		Role:        role,
		Department:  department,
		Description: description,
		SortOrder:   sortOrder,
		Active:      active == 1,
	}
	if tags != "" {
		m.Tags = strings.Split(tags, ",")
	}
	return m
}

// GetTeamMembers returns all active team members.
func GetTeamMembers() ([]TeamMember, error) {
	rows, err := db.DB.Query(
		`SELECT id, initial, name, role, department, description, tags, sort_order, active
		 FROM team_members WHERE active = 1 ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetTeamMembers: %w", err)
	}
	defer rows.Close()
	var out []TeamMember
	for rows.Next() {
		var id, sortOrder, active int
		var initial, name, role, department, description, tags string
		if err := rows.Scan(&id, &initial, &name, &role, &department, &description, &tags, &sortOrder, &active); err != nil {
			return nil, err
		}
		out = append(out, scanTeamMember(id, initial, name, role, department, description, tags, sortOrder, active))
	}
	return out, rows.Err()
}

// GetAllTeamMembers returns all team members for the admin panel.
func GetAllTeamMembers() ([]TeamMember, error) {
	rows, err := db.DB.Query(
		`SELECT id, initial, name, role, department, description, tags, sort_order, active
		 FROM team_members ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetAllTeamMembers: %w", err)
	}
	defer rows.Close()
	var out []TeamMember
	for rows.Next() {
		var id, sortOrder, active int
		var initial, name, role, department, description, tags string
		if err := rows.Scan(&id, &initial, &name, &role, &department, &description, &tags, &sortOrder, &active); err != nil {
			return nil, err
		}
		out = append(out, scanTeamMember(id, initial, name, role, department, description, tags, sortOrder, active))
	}
	return out, rows.Err()
}

// AddTeamMember inserts a new team member.
func AddTeamMember(initial, name, role, department, description, tags string) error {
	_, err := db.DB.Exec(
		`INSERT INTO team_members (initial, name, role, department, description, tags, sort_order)
		 VALUES (?, ?, ?, ?, ?, ?, (SELECT COALESCE(MAX(sort_order)+1, 0) FROM team_members))`,
		initial, name, role, department, description, tags,
	)
	return err
}

// UpdateTeamMember updates an existing team member.
func UpdateTeamMember(id int, initial, name, role, department, description, tags string, active bool) error {
	a := 0
	if active {
		a = 1
	}
	_, err := db.DB.Exec(
		`UPDATE team_members SET initial=?, name=?, role=?, department=?, description=?, tags=?, active=? WHERE id=?`,
		initial, name, role, department, description, tags, a, id,
	)
	return err
}

// DeleteTeamMember removes a team member.
func DeleteTeamMember(id int) error {
	_, err := db.DB.Exec(`DELETE FROM team_members WHERE id=?`, id)
	return err
}
