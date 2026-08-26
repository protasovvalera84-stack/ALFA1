package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// FAQItem represents a question/answer pair in the FAQ section.
type FAQItem struct {
	ID        int
	Question  string
	Answer    string
	SortOrder int
	Active    bool
}

// GetFAQItems returns all active FAQ items.
func GetFAQItems() ([]FAQItem, error) {
	rows, err := db.DB.Query(
		`SELECT id, question, answer, sort_order, active
		 FROM faq_items WHERE active = 1 ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetFAQItems: %w", err)
	}
	defer rows.Close()
	return scanFAQItems(rows)
}

// GetAllFAQItems returns all FAQ items for the admin panel.
func GetAllFAQItems() ([]FAQItem, error) {
	rows, err := db.DB.Query(
		`SELECT id, question, answer, sort_order, active
		 FROM faq_items ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetAllFAQItems: %w", err)
	}
	defer rows.Close()
	return scanFAQItems(rows)
}

func scanFAQItems(rows interface{ Next() bool; Scan(...any) error; Err() error }) ([]FAQItem, error) {
	var out []FAQItem
	for rows.Next() {
		var f FAQItem
		var active int
		if err := rows.Scan(&f.ID, &f.Question, &f.Answer, &f.SortOrder, &active); err != nil {
			return nil, err
		}
		f.Active = active == 1
		out = append(out, f)
	}
	return out, rows.Err()
}

// AddFAQItem inserts a new FAQ item.
func AddFAQItem(question, answer string) error {
	_, err := db.DB.Exec(
		`INSERT INTO faq_items (question, answer, sort_order)
		 VALUES (?, ?, (SELECT COALESCE(MAX(sort_order)+1, 0) FROM faq_items))`,
		question, answer,
	)
	return err
}

// UpdateFAQItem updates an existing FAQ item.
func UpdateFAQItem(id int, question, answer string, active bool) error {
	a := 0
	if active {
		a = 1
	}
	_, err := db.DB.Exec(
		`UPDATE faq_items SET question=?, answer=?, active=? WHERE id=?`,
		question, answer, a, id,
	)
	return err
}

// DeleteFAQItem removes a FAQ item.
func DeleteFAQItem(id int) error {
	_, err := db.DB.Exec(`DELETE FROM faq_items WHERE id=?`, id)
	return err
}
