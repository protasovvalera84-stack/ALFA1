// Package models — data access for the 7 content sections:
// Advantages, History, Licenses, Clients, Team, Workflow, FAQ.
package models

import (
	"fmt"

	"alfaunit1/internal/db"
)

// ── Advantage (Преимущество) ──────────────────────────────────────────────

// Advantage represents a single company advantage/benefit shown on the homepage.
type Advantage struct {
	ID          int
	Title       string
	Description string
	Icon        string
	SortOrder   int
	Active      bool
}

// GetAdvantages returns all advantages ordered by sort_order.
func GetAdvantages() ([]Advantage, error) {
	rows, err := db.DB.Query(
		`SELECT id, title, description, icon, sort_order, active
		 FROM advantages ORDER BY sort_order, id`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetAdvantages: %w", err)
	}
	defer rows.Close()
	var items []Advantage
	for rows.Next() {
		var a Advantage
		var active int
		if err := rows.Scan(&a.ID, &a.Title, &a.Description, &a.Icon, &a.SortOrder, &active); err != nil {
			return nil, err
		}
		a.Active = active == 1
		items = append(items, a)
	}
	return items, rows.Err()
}

// GetAdvantage returns a single advantage by ID.
func GetAdvantage(id int) (*Advantage, error) {
	a := &Advantage{}
	var active int
	err := db.DB.QueryRow(
		`SELECT id, title, description, icon, sort_order, active FROM advantages WHERE id = ?`, id,
	).Scan(&a.ID, &a.Title, &a.Description, &a.Icon, &a.SortOrder, &active)
	if err != nil {
		return nil, err
	}
	a.Active = active == 1
	return a, nil
}

// SaveAdvantage inserts or updates an advantage.
func SaveAdvantage(a *Advantage) error {
	active := 0
	if a.Active {
		active = 1
	}
	if a.ID == 0 {
		_, err := db.DB.Exec(
			`INSERT INTO advantages (title, description, icon, sort_order, active) VALUES (?, ?, ?, ?, ?)`,
			a.Title, a.Description, a.Icon, a.SortOrder, active,
		)
		return err
	}
	_, err := db.DB.Exec(
		`UPDATE advantages SET title=?, description=?, icon=?, sort_order=?, active=? WHERE id=?`,
		a.Title, a.Description, a.Icon, a.SortOrder, active, a.ID,
	)
	return err
}

// DeleteAdvantage removes an advantage by ID.
func DeleteAdvantage(id int) error {
	_, err := db.DB.Exec(`DELETE FROM advantages WHERE id = ?`, id)
	return err
}

// ── HistoryEvent (История) ────────────────────────────────────────────────

// HistoryEvent represents a single entry in the company timeline.
type HistoryEvent struct {
	ID          int
	Year        string
	Title       string
	Description string
	SortOrder   int
}

// GetHistoryEvents returns all history events ordered by sort_order.
func GetHistoryEvents() ([]HistoryEvent, error) {
	rows, err := db.DB.Query(
		`SELECT id, year, title, description, sort_order
		 FROM history_events ORDER BY sort_order, id`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetHistoryEvents: %w", err)
	}
	defer rows.Close()
	var items []HistoryEvent
	for rows.Next() {
		var h HistoryEvent
		if err := rows.Scan(&h.ID, &h.Year, &h.Title, &h.Description, &h.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, h)
	}
	return items, rows.Err()
}

// GetHistoryEvent returns a single history event by ID.
func GetHistoryEvent(id int) (*HistoryEvent, error) {
	h := &HistoryEvent{}
	err := db.DB.QueryRow(
		`SELECT id, year, title, description, sort_order FROM history_events WHERE id = ?`, id,
	).Scan(&h.ID, &h.Year, &h.Title, &h.Description, &h.SortOrder)
	if err != nil {
		return nil, err
	}
	return h, nil
}

// SaveHistoryEvent inserts or updates a history event.
func SaveHistoryEvent(h *HistoryEvent) error {
	if h.ID == 0 {
		_, err := db.DB.Exec(
			`INSERT INTO history_events (year, title, description, sort_order) VALUES (?, ?, ?, ?)`,
			h.Year, h.Title, h.Description, h.SortOrder,
		)
		return err
	}
	_, err := db.DB.Exec(
		`UPDATE history_events SET year=?, title=?, description=?, sort_order=? WHERE id=?`,
		h.Year, h.Title, h.Description, h.SortOrder, h.ID,
	)
	return err
}

// DeleteHistoryEvent removes a history event by ID.
func DeleteHistoryEvent(id int) error {
	_, err := db.DB.Exec(`DELETE FROM history_events WHERE id = ?`, id)
	return err
}

// ── License (Лицензия) ────────────────────────────────────────────────────

// License represents a company license or certificate.
type License struct {
	ID        int
	Name      string
	Number    string
	Issuer    string
	IssuedAt  string
	SortOrder int
}

// GetLicenses returns all licenses ordered by sort_order.
func GetLicenses() ([]License, error) {
	rows, err := db.DB.Query(
		`SELECT id, name, number, issuer, issued_at, sort_order
		 FROM licenses ORDER BY sort_order, id`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetLicenses: %w", err)
	}
	defer rows.Close()
	var items []License
	for rows.Next() {
		var l License
		if err := rows.Scan(&l.ID, &l.Name, &l.Number, &l.Issuer, &l.IssuedAt, &l.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, l)
	}
	return items, rows.Err()
}

// GetLicense returns a single license by ID.
func GetLicense(id int) (*License, error) {
	l := &License{}
	err := db.DB.QueryRow(
		`SELECT id, name, number, issuer, issued_at, sort_order FROM licenses WHERE id = ?`, id,
	).Scan(&l.ID, &l.Name, &l.Number, &l.Issuer, &l.IssuedAt, &l.SortOrder)
	if err != nil {
		return nil, err
	}
	return l, nil
}

// SaveLicense inserts or updates a license.
func SaveLicense(l *License) error {
	if l.ID == 0 {
		_, err := db.DB.Exec(
			`INSERT INTO licenses (name, number, issuer, issued_at, sort_order) VALUES (?, ?, ?, ?, ?)`,
			l.Name, l.Number, l.Issuer, l.IssuedAt, l.SortOrder,
		)
		return err
	}
	_, err := db.DB.Exec(
		`UPDATE licenses SET name=?, number=?, issuer=?, issued_at=?, sort_order=? WHERE id=?`,
		l.Name, l.Number, l.Issuer, l.IssuedAt, l.SortOrder, l.ID,
	)
	return err
}

// DeleteLicense removes a license by ID.
func DeleteLicense(id int) error {
	_, err := db.DB.Exec(`DELETE FROM licenses WHERE id = ?`, id)
	return err
}

// ── Client (Клиент) ───────────────────────────────────────────────────────

// Client represents a protected client/object.
type Client struct {
	ID          int
	Name        string
	Sector      string
	Description string
	SortOrder   int
}

// GetClients returns all clients ordered by sort_order.
func GetClients() ([]Client, error) {
	rows, err := db.DB.Query(
		`SELECT id, name, sector, description, sort_order
		 FROM clients ORDER BY sort_order, id`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetClients: %w", err)
	}
	defer rows.Close()
	var items []Client
	for rows.Next() {
		var c Client
		if err := rows.Scan(&c.ID, &c.Name, &c.Sector, &c.Description, &c.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}

// GetClient returns a single client by ID.
func GetClient(id int) (*Client, error) {
	c := &Client{}
	err := db.DB.QueryRow(
		`SELECT id, name, sector, description, sort_order FROM clients WHERE id = ?`, id,
	).Scan(&c.ID, &c.Name, &c.Sector, &c.Description, &c.SortOrder)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// SaveClient inserts or updates a client.
func SaveClient(c *Client) error {
	if c.ID == 0 {
		_, err := db.DB.Exec(
			`INSERT INTO clients (name, sector, description, sort_order) VALUES (?, ?, ?, ?)`,
			c.Name, c.Sector, c.Description, c.SortOrder,
		)
		return err
	}
	_, err := db.DB.Exec(
		`UPDATE clients SET name=?, sector=?, description=?, sort_order=? WHERE id=?`,
		c.Name, c.Sector, c.Description, c.SortOrder, c.ID,
	)
	return err
}

// DeleteClient removes a client by ID.
func DeleteClient(id int) error {
	_, err := db.DB.Exec(`DELETE FROM clients WHERE id = ?`, id)
	return err
}

// ── TeamMember (Команда) ──────────────────────────────────────────────────

// TeamMember represents a company team member.
type TeamMember struct {
	ID        int
	Name      string
	Position  string
	Bio       string
	PhotoURL  string
	SortOrder int
	Active    bool
}

// GetTeamMembers returns all team members ordered by sort_order.
func GetTeamMembers() ([]TeamMember, error) {
	rows, err := db.DB.Query(
		`SELECT id, name, position, bio, photo_url, sort_order, active
		 FROM team_members ORDER BY sort_order, id`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetTeamMembers: %w", err)
	}
	defer rows.Close()
	var items []TeamMember
	for rows.Next() {
		var m TeamMember
		var active int
		if err := rows.Scan(&m.ID, &m.Name, &m.Position, &m.Bio, &m.PhotoURL, &m.SortOrder, &active); err != nil {
			return nil, err
		}
		m.Active = active == 1
		items = append(items, m)
	}
	return items, rows.Err()
}

// GetTeamMember returns a single team member by ID.
func GetTeamMember(id int) (*TeamMember, error) {
	m := &TeamMember{}
	var active int
	err := db.DB.QueryRow(
		`SELECT id, name, position, bio, photo_url, sort_order, active FROM team_members WHERE id = ?`, id,
	).Scan(&m.ID, &m.Name, &m.Position, &m.Bio, &m.PhotoURL, &m.SortOrder, &active)
	if err != nil {
		return nil, err
	}
	m.Active = active == 1
	return m, nil
}

// SaveTeamMember inserts or updates a team member.
func SaveTeamMember(m *TeamMember) error {
	active := 0
	if m.Active {
		active = 1
	}
	if m.ID == 0 {
		_, err := db.DB.Exec(
			`INSERT INTO team_members (name, position, bio, photo_url, sort_order, active) VALUES (?, ?, ?, ?, ?, ?)`,
			m.Name, m.Position, m.Bio, m.PhotoURL, m.SortOrder, active,
		)
		return err
	}
	_, err := db.DB.Exec(
		`UPDATE team_members SET name=?, position=?, bio=?, photo_url=?, sort_order=?, active=? WHERE id=?`,
		m.Name, m.Position, m.Bio, m.PhotoURL, m.SortOrder, active, m.ID,
	)
	return err
}

// DeleteTeamMember removes a team member by ID.
func DeleteTeamMember(id int) error {
	_, err := db.DB.Exec(`DELETE FROM team_members WHERE id = ?`, id)
	return err
}

// ── WorkflowStep (Как мы работаем) ───────────────────────────────────────

// WorkflowStep represents a single step in the company's workflow.
type WorkflowStep struct {
	ID          int
	StepNumber  int
	Title       string
	Description string
	SortOrder   int
}

// GetWorkflowSteps returns all workflow steps ordered by sort_order.
func GetWorkflowSteps() ([]WorkflowStep, error) {
	rows, err := db.DB.Query(
		`SELECT id, step_number, title, description, sort_order
		 FROM workflow_steps ORDER BY sort_order, step_number, id`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetWorkflowSteps: %w", err)
	}
	defer rows.Close()
	var items []WorkflowStep
	for rows.Next() {
		var s WorkflowStep
		if err := rows.Scan(&s.ID, &s.StepNumber, &s.Title, &s.Description, &s.SortOrder); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

// GetWorkflowStep returns a single workflow step by ID.
func GetWorkflowStep(id int) (*WorkflowStep, error) {
	s := &WorkflowStep{}
	err := db.DB.QueryRow(
		`SELECT id, step_number, title, description, sort_order FROM workflow_steps WHERE id = ?`, id,
	).Scan(&s.ID, &s.StepNumber, &s.Title, &s.Description, &s.SortOrder)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// SaveWorkflowStep inserts or updates a workflow step.
func SaveWorkflowStep(s *WorkflowStep) error {
	if s.ID == 0 {
		_, err := db.DB.Exec(
			`INSERT INTO workflow_steps (step_number, title, description, sort_order) VALUES (?, ?, ?, ?)`,
			s.StepNumber, s.Title, s.Description, s.SortOrder,
		)
		return err
	}
	_, err := db.DB.Exec(
		`UPDATE workflow_steps SET step_number=?, title=?, description=?, sort_order=? WHERE id=?`,
		s.StepNumber, s.Title, s.Description, s.SortOrder, s.ID,
	)
	return err
}

// DeleteWorkflowStep removes a workflow step by ID.
func DeleteWorkflowStep(id int) error {
	_, err := db.DB.Exec(`DELETE FROM workflow_steps WHERE id = ?`, id)
	return err
}

// ── FAQItem (FAQ) ─────────────────────────────────────────────────────────

// FAQItem represents a frequently asked question and its answer.
type FAQItem struct {
	ID        int
	Question  string
	Answer    string
	SortOrder int
	Active    bool
}

// GetFAQItems returns all FAQ items ordered by sort_order.
func GetFAQItems() ([]FAQItem, error) {
	rows, err := db.DB.Query(
		`SELECT id, question, answer, sort_order, active
		 FROM faq_items ORDER BY sort_order, id`,
	)
	if err != nil {
		return nil, fmt.Errorf("models: GetFAQItems: %w", err)
	}
	defer rows.Close()
	var items []FAQItem
	for rows.Next() {
		var f FAQItem
		var active int
		if err := rows.Scan(&f.ID, &f.Question, &f.Answer, &f.SortOrder, &active); err != nil {
			return nil, err
		}
		f.Active = active == 1
		items = append(items, f)
	}
	return items, rows.Err()
}

// GetFAQItem returns a single FAQ item by ID.
func GetFAQItem(id int) (*FAQItem, error) {
	f := &FAQItem{}
	var active int
	err := db.DB.QueryRow(
		`SELECT id, question, answer, sort_order, active FROM faq_items WHERE id = ?`, id,
	).Scan(&f.ID, &f.Question, &f.Answer, &f.SortOrder, &active)
	if err != nil {
		return nil, err
	}
	f.Active = active == 1
	return f, nil
}

// SaveFAQItem inserts or updates a FAQ item.
func SaveFAQItem(f *FAQItem) error {
	active := 0
	if f.Active {
		active = 1
	}
	if f.ID == 0 {
		_, err := db.DB.Exec(
			`INSERT INTO faq_items (question, answer, sort_order, active) VALUES (?, ?, ?, ?)`,
			f.Question, f.Answer, f.SortOrder, active,
		)
		return err
	}
	_, err := db.DB.Exec(
		`UPDATE faq_items SET question=?, answer=?, sort_order=?, active=? WHERE id=?`,
		f.Question, f.Answer, f.SortOrder, active, f.ID,
	)
	return err
}

// DeleteFAQItem removes a FAQ item by ID.
func DeleteFAQItem(id int) error {
	_, err := db.DB.Exec(`DELETE FROM faq_items WHERE id = ?`, id)
	return err
}
