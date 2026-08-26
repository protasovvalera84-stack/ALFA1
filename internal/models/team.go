package models

import (
	"fmt"
	"strings"

	"alfaunit1/internal/db"
)

// TeamMember represents a Наша команда card.
type TeamMember struct {
	ID          int
	Letter      string
	ColorClass  string // "gold" or "blue"
	Title       string
	Department  string
	Description string
	Tags        []string // stored as comma-separated in DB
	SortOrder   int
	Active      bool
}

// TagsRaw returns the comma-joined tags string (for form values).
func (m TeamMember) TagsRaw() string {
	return strings.Join(m.Tags, ",")
}

// GetTeamMembers returns all active team members ordered by sort_order.
func GetTeamMembers() ([]TeamMember, error) {
	rows, err := db.DB.Query(`SELECT id, letter, color_class, title, department, description, tags, sort_order, active FROM team_members WHERE active=1 ORDER BY sort_order`)
	if err != nil {
		return nil, fmt.Errorf("models: GetTeamMembers: %w", err)
	}
	defer rows.Close()
	return scanTeamMembers(rows)
}

// GetAllTeamMembers returns all members (including inactive) for admin.
func GetAllTeamMembers() ([]TeamMember, error) {
	rows, err := db.DB.Query(`SELECT id, letter, color_class, title, department, description, tags, sort_order, active FROM team_members ORDER BY sort_order`)
	if err != nil {
		return nil, fmt.Errorf("models: GetAllTeamMembers: %w", err)
	}
	defer rows.Close()
	return scanTeamMembers(rows)
}

func scanTeamMembers(rows interface{ Next() bool; Scan(...any) error; Err() error }) ([]TeamMember, error) {
	var out []TeamMember
	for rows.Next() {
		var m TeamMember
		var active int
		var tagsRaw string
		if err := rows.Scan(&m.ID, &m.Letter, &m.ColorClass, &m.Title, &m.Department, &m.Description, &tagsRaw, &m.SortOrder, &active); err != nil {
			return nil, err
		}
		m.Active = active == 1
		if tagsRaw != "" {
			for _, t := range strings.Split(tagsRaw, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					m.Tags = append(m.Tags, t)
				}
			}
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// AddTeamMember inserts a new team member.
func AddTeamMember(letter, colorClass, title, department, description, tags string) error {
	_, err := db.DB.Exec(`INSERT INTO team_members (letter, color_class, title, department, description, tags, sort_order) VALUES (?, ?, ?, ?, ?, ?, (SELECT COALESCE(MAX(sort_order),0)+1 FROM team_members))`, letter, colorClass, title, department, description, tags)
	return err
}

// UpdateTeamMember updates a team member.
func UpdateTeamMember(id int, letter, colorClass, title, department, description, tags string, active bool) error {
	a := 0
	if active {
		a = 1
	}
	_, err := db.DB.Exec(`UPDATE team_members SET letter=?, color_class=?, title=?, department=?, description=?, tags=?, active=? WHERE id=?`, letter, colorClass, title, department, description, tags, a, id)
	return err
}

// DeleteTeamMember removes a team member by ID.
func DeleteTeamMember(id int) error {
	_, err := db.DB.Exec(`DELETE FROM team_members WHERE id=?`, id)
	return err
}
