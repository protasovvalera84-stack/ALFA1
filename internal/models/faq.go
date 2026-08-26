package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// FAQItem represents a single question-answer pair in the FAQ section.
type FAQItem struct {
	ID        int
	Question  string
	Answer    string
	SortOrder int
}

// GetFAQItems returns all FAQ items ordered by sort_order.
func GetFAQItems() ([]FAQItem, error) {
	rows, err := db.DB.Query(
		`SELECT id, question, answer, sort_order FROM faq_items ORDER BY sort_order ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetFAQItems: %w", err)
	}
	defer rows.Close()
	var items []FAQItem
	for rows.Next() {
		var f FAQItem
		if err := rows.Scan(&f.ID, &f.Question, &f.Answer, &f.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, f)
	}
	return items, rows.Err()
}

// CreateFAQItem inserts a new FAQ item.
func CreateFAQItem(question, answer string) error {
	_, err := db.DB.Exec(
		`INSERT INTO faq_items (question, answer, sort_order)
		 VALUES (?, ?, (SELECT COALESCE(MAX(sort_order)+1, 0) FROM faq_items))`,
		question, answer,
	)
	return err
}

// UpdateFAQItem updates an existing FAQ item.
func UpdateFAQItem(id int, question, answer string) error {
	_, err := db.DB.Exec(
		`UPDATE faq_items SET question=?, answer=? WHERE id=?`,
		question, answer, id,
	)
	return err
}

// DeleteFAQItem removes a FAQ item by ID.
func DeleteFAQItem(id int) error {
	_, err := db.DB.Exec(`DELETE FROM faq_items WHERE id=?`, id)
	return err
}
