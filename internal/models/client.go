package models

import (
	"fmt"
	"strings"

	"alfaunit1/internal/db"
)

// Client represents a Наши клиенты card.
type Client struct {
	ID         int
	Letter     string
	ColorClass string // "gold" or "blue"
	Name       string
	TypeLabel  string
	Description string
	Tags       []string // stored as comma-separated in DB
	SortOrder  int
	Active     bool
}

// TagsRaw returns the comma-joined tags string (for form values).
func (c Client) TagsRaw() string {
	return strings.Join(c.Tags, ",")
}

// GetClients returns all active clients ordered by sort_order.
func GetClients() ([]Client, error) {
	rows, err := db.DB.Query(`SELECT id, letter, color_class, name, type_label, description, tags, sort_order, active FROM clients WHERE active=1 ORDER BY sort_order`)
	if err != nil {
		return nil, fmt.Errorf("models: GetClients: %w", err)
	}
	defer rows.Close()
	return scanClients(rows)
}

// GetAllClients returns all clients (including inactive) for admin.
func GetAllClients() ([]Client, error) {
	rows, err := db.DB.Query(`SELECT id, letter, color_class, name, type_label, description, tags, sort_order, active FROM clients ORDER BY sort_order`)
	if err != nil {
		return nil, fmt.Errorf("models: GetAllClients: %w", err)
	}
	defer rows.Close()
	return scanClients(rows)
}

func scanClients(rows interface{ Next() bool; Scan(...any) error; Err() error }) ([]Client, error) {
	var out []Client
	for rows.Next() {
		var c Client
		var active int
		var tagsRaw string
		if err := rows.Scan(&c.ID, &c.Letter, &c.ColorClass, &c.Name, &c.TypeLabel, &c.Description, &tagsRaw, &c.SortOrder, &active); err != nil {
			return nil, err
		}
		c.Active = active == 1
		if tagsRaw != "" {
			for _, t := range strings.Split(tagsRaw, ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					c.Tags = append(c.Tags, t)
				}
			}
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// AddClient inserts a new client.
func AddClient(letter, colorClass, name, typeLabel, description, tags string) error {
	_, err := db.DB.Exec(`INSERT INTO clients (letter, color_class, name, type_label, description, tags, sort_order) VALUES (?, ?, ?, ?, ?, ?, (SELECT COALESCE(MAX(sort_order),0)+1 FROM clients))`, letter, colorClass, name, typeLabel, description, tags)
	return err
}

// UpdateClient updates a client.
func UpdateClient(id int, letter, colorClass, name, typeLabel, description, tags string, active bool) error {
	a := 0
	if active {
		a = 1
	}
	_, err := db.DB.Exec(`UPDATE clients SET letter=?, color_class=?, name=?, type_label=?, description=?, tags=?, active=? WHERE id=?`, letter, colorClass, name, typeLabel, description, tags, a, id)
	return err
}

// DeleteClient removes a client by ID.
func DeleteClient(id int) error {
	_, err := db.DB.Exec(`DELETE FROM clients WHERE id=?`, id)
	return err
}
