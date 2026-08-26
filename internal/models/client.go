package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// Client represents a protected-object card in the "Our Clients" section.
type Client struct {
	ID          int
	Letter      string
	Name        string
	Sector      string
	Description string
	Tags        string
	SortOrder   int
}

// GetClients returns all clients ordered by sort_order.
func GetClients() ([]Client, error) {
	rows, err := db.DB.Query(
		`SELECT id, letter, name, sector, description, tags, sort_order
		 FROM clients ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetClients: %w", err)
	}
	defer rows.Close()
	var items []Client
	for rows.Next() {
		var c Client
		if err := rows.Scan(&c.ID, &c.Letter, &c.Name, &c.Sector, &c.Description, &c.Tags, &c.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

// CreateClient inserts a new client record.
func CreateClient(letter, name, sector, description, tags string) error {
	_, err := db.DB.Exec(
		`INSERT INTO clients (letter, name, sector, description, tags, sort_order)
		 VALUES (?, ?, ?, ?, ?, (SELECT COALESCE(MAX(sort_order)+1, 0) FROM clients))`,
		letter, name, sector, description, tags,
	)
	return err
}

// UpdateClient updates an existing client record.
func UpdateClient(id int, letter, name, sector, description, tags string) error {
	_, err := db.DB.Exec(
		`UPDATE clients SET letter=?, name=?, sector=?, description=?, tags=? WHERE id=?`,
		letter, name, sector, description, tags, id,
	)
	return err
}

// DeleteClient removes a client by ID.
func DeleteClient(id int) error {
	_, err := db.DB.Exec(`DELETE FROM clients WHERE id=?`, id)
	return err
}
