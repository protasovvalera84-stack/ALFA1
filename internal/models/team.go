package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// TeamMember represents a person in the "Our Team" section.
type TeamMember struct {
	ID          int
	Letter      string
	Name        string
	Role        string
	Department  string
	Description string
	Tags        string
	SortOrder   int
}

// GetTeamMembers returns all team members ordered by sort_order.
func GetTeamMembers() ([]TeamMember, error) {
	rows, err := db.DB.Query(
		`SELECT id, letter, name, role, department, description, tags, sort_order
		 FROM team_members ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetTeamMembers: %w", err)
	}
	defer rows.Close()
	var items []TeamMember
	for rows.Next() {
		var t TeamMember
		if err := rows.Scan(&t.ID, &t.Letter, &t.Name, &t.Role, &t.Department, &t.Description, &t.Tags, &t.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

// CreateTeamMember inserts a new team member record.
func CreateTeamMember(letter, name, role, department, description, tags string) error {
	_, err := db.DB.Exec(
		`INSERT INTO team_members (letter, name, role, department, description, tags, sort_order)
		 VALUES (?, ?, ?, ?, ?, ?, (SELECT COALESCE(MAX(sort_order)+1, 0) FROM team_members))`,
		letter, name, role, department, description, tags,
	)
	return err
}

// UpdateTeamMember updates an existing team member record.
func UpdateTeamMember(id int, letter, name, role, department, description, tags string) error {
	_, err := db.DB.Exec(
		`UPDATE team_members SET letter=?, name=?, role=?, department=?, description=?, tags=? WHERE id=?`,
		letter, name, role, department, description, tags, id,
	)
	return err
}

// DeleteTeamMember removes a team member by ID.
func DeleteTeamMember(id int) error {
	_, err := db.DB.Exec(`DELETE FROM team_members WHERE id=?`, id)
	return err
}
