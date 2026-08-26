package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// FAQ represents a frequently asked question.
type FAQ struct {
	ID        int
	Question  string
	Answer    string
	SortOrder int
	Active    bool
}

// GetFAQs returns all active FAQs ordered by sort_order.
func GetFAQs() ([]FAQ, error) {
	rows, err := db.DB.Query(`SELECT id, question, answer, sort_order, active FROM faqs WHERE active=1 ORDER BY sort_order`)
	if err != nil {
		return nil, fmt.Errorf("models: GetFAQs: %w", err)
	}
	defer rows.Close()
	return scanFAQs(rows)
}

// GetAllFAQs returns all FAQs (including inactive) for admin.
func GetAllFAQs() ([]FAQ, error) {
	rows, err := db.DB.Query(`SELECT id, question, answer, sort_order, active FROM faqs ORDER BY sort_order`)
	if err != nil {
		return nil, fmt.Errorf("models: GetAllFAQs: %w", err)
	}
	defer rows.Close()
	return scanFAQs(rows)
}

func scanFAQs(rows interface{ Next() bool; Scan(...any) error; Err() error }) ([]FAQ, error) {
	var out []FAQ
	for rows.Next() {
		var f FAQ
		var active int
		if err := rows.Scan(&f.ID, &f.Question, &f.Answer, &f.SortOrder, &active); err != nil {
			return nil, err
		}
		f.Active = active == 1
		out = append(out, f)
	}
	return out, rows.Err()
}

// AddFAQ inserts a new FAQ.
func AddFAQ(question, answer string) error {
	_, err := db.DB.Exec(`INSERT INTO faqs (question, answer, sort_order) VALUES (?, ?, (SELECT COALESCE(MAX(sort_order),0)+1 FROM faqs))`, question, answer)
	return err
}

// UpdateFAQ updates a FAQ.
func UpdateFAQ(id int, question, answer string, active bool) error {
	a := 0
	if active {
		a = 1
	}
	_, err := db.DB.Exec(`UPDATE faqs SET question=?, answer=?, active=? WHERE id=?`, question, answer, a, id)
	return err
}

// DeleteFAQ removes a FAQ by ID.
func DeleteFAQ(id int) error {
	_, err := db.DB.Exec(`DELETE FROM faqs WHERE id=?`, id)
	return err
}
