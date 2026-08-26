package models

import (
	"fmt"
	"strings"

	"alfaunit1/internal/db"
)

// Client represents a protected object / customer card on the site.
type Client struct {
	ID          int
	Initial     string
	Name        string
	TypeLabel   string
	Description string
	Tags        []string // stored as comma-separated in DB
	SortOrder   int
	Active      bool
}

func scanClient(id int, initial, name, typeLabel, description, tags string, sortOrder, active int) Client {
	c := Client{
		ID:          id,
		Initial:     initial,
		Name:        name,
		TypeLabel:   typeLabel,
		Description: description,
		SortOrder:   sortOrder,
		Active:      active == 1,
	}
	if tags != "" {
		c.Tags = strings.Split(tags, ",")
	}
	return c
}

// GetClients returns all active clients ordered by sort_order.
func GetClients() ([]Client, error) {
	rows, err := db.DB.Query(
		`SELECT id, initial, name, type_label, description, tags, sort_order, active
		 FROM clients WHERE active = 1 ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetClients: %w", err)
	}
	defer rows.Close()
	var out []Client
	for rows.Next() {
		var id, sortOrder, active int
		var initial, name, typeLabel, description, tags string
		if err := rows.Scan(&id, &initial, &name, &typeLabel, &description, &tags, &sortOrder, &active); err != nil {
			return nil, err
		}
		out = append(out, scanClient(id, initial, name, typeLabel, description, tags, sortOrder, active))
	}
	return out, rows.Err()
}

// GetAllClients returns all clients (including inactive) for the admin panel.
func GetAllClients() ([]Client, error) {
	rows, err := db.DB.Query(
		`SELECT id, initial, name, type_label, description, tags, sort_order, active
		 FROM clients ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetAllClients: %w", err)
	}
	defer rows.Close()
	var out []Client
	for rows.Next() {
		var id, sortOrder, active int
		var initial, name, typeLabel, description, tags string
		if err := rows.Scan(&id, &initial, &name, &typeLabel, &description, &tags, &sortOrder, &active); err != nil {
			return nil, err
		}
		out = append(out, scanClient(id, initial, name, typeLabel, description, tags, sortOrder, active))
	}
	return out, rows.Err()
}

// AddClient inserts a new client record.
func AddClient(initial, name, typeLabel, description, tags string) error {
	_, err := db.DB.Exec(
		`INSERT INTO clients (initial, name, type_label, description, tags, sort_order)
		 VALUES (?, ?, ?, ?, ?, (SELECT COALESCE(MAX(sort_order)+1, 0) FROM clients))`,
		initial, name, typeLabel, description, tags,
	)
	return err
}

// UpdateClient updates an existing client record.
func UpdateClient(id int, initial, name, typeLabel, description, tags string, active bool) error {
	a := 0
	if active {
		a = 1
	}
	_, err := db.DB.Exec(
		`UPDATE clients SET initial=?, name=?, type_label=?, description=?, tags=?, active=? WHERE id=?`,
		initial, name, typeLabel, description, tags, a, id,
	)
	return err
}

// DeleteClient removes a client record.
func DeleteClient(id int) error {
	_, err := db.DB.Exec(`DELETE FROM clients WHERE id=?`, id)
	return err
}
