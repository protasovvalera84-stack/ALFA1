package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// Contact represents a submitted contact form inquiry.
type Contact struct {
	ID        int
	Name      string
	Phone     string
	Email     string
	Message   string
	Service   string
	CreatedAt string
	Read      bool
}

// SaveContact inserts a new contact form submission into the database.
func SaveContact(name, phone, email, message, service string) error {
	_, err := db.DB.Exec(
		`INSERT INTO contacts (name, phone, email, message, service) VALUES (?, ?, ?, ?, ?)`,
		name, phone, email, message, service,
	)
	if err != nil {
		return fmt.Errorf("models: SaveContact: %w", err)
	}
	return nil
}

// GetContacts returns all contact submissions, newest first.
func GetContacts() ([]Contact, error) {
	rows, err := db.DB.Query(
		`SELECT id, name, phone, email, message, service, created_at, read
		 FROM contacts ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetContacts: %w", err)
	}
	defer rows.Close()

	var contacts []Contact
	for rows.Next() {
		var c Contact
		var read int
		if err := rows.Scan(&c.ID, &c.Name, &c.Phone, &c.Email, &c.Message, &c.Service, &c.CreatedAt, &read); err != nil {
			return nil, err
		}
		c.Read = read == 1
		contacts = append(contacts, c)
	}
	return contacts, rows.Err()
}

// MarkContactRead marks a contact submission as read.
func MarkContactRead(id int) error {
	_, err := db.DB.Exec(`UPDATE contacts SET read = 1 WHERE id = ?`, id)
	return err
}

// UnreadContactCount returns the count of unread contact submissions.
func UnreadContactCount() int {
	var count int
	_ = db.DB.QueryRow(`SELECT COUNT(*) FROM contacts WHERE read = 0`).Scan(&count)
	return count
}
