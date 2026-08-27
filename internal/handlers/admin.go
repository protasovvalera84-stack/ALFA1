package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"alfaunit1/internal/middleware"
	"alfaunit1/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// AdminTemplates is the parsed admin template set, assigned at server startup.
var AdminTemplates *template.Template

// adminData is the base data struct passed to all admin templates.
type adminData struct {
	Title       string
	Flash       string
	UnreadCount int
}

func baseAdmin(title string) adminData {
	return adminData{
		Title:       title,
		UnreadCount: models.UnreadContactCount(),
	}
}

// AdminDashboard renders the admin dashboard.
func AdminDashboard(w http.ResponseWriter, r *http.Request) {
	settings, err := models.GetSettings()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		adminData
		Settings    *models.SiteSettings
		ServiceCount int
	}{
		adminData:    baseAdmin("Dashboard — Альфа Юнит-1"),
		Settings:     settings,
	}

	// Count active services.
	svcs, _ := models.GetServices()
	data.ServiceCount = len(svcs)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		log.Printf("admin: dashboard render: %v", err)
	}
}

// AdminLogin handles GET (show form) and POST (verify password) for admin login.
func AdminLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = AdminTemplates.ExecuteTemplate(w, "login.html", map[string]string{"Error": ""})
		return
	}

	password := strings.TrimSpace(r.FormValue("password"))
	if password == "" {
		renderLoginError(w, "Введите пароль.")
		return
	}

	hash := models.GetSetting("admin_password_hash")
	if hash == "" {
		renderLoginError(w, "Пароль администратора не задан. Установите ADMIN_PASSWORD в .env.")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		renderLoginError(w, "Неверный пароль.")
		return
	}

	if err := middleware.CreateSession(w); err != nil {
		log.Printf("admin: login: create session: %v", err)
		renderLoginError(w, "Ошибка создания сессии.")
		return
	}

	http.Redirect(w, r, "/admin/", http.StatusFound)
}

// AdminLogout destroys the session and redirects to login.
func AdminLogout(w http.ResponseWriter, r *http.Request) {
	middleware.DestroySession(w, r)
	http.Redirect(w, r, "/admin/login", http.StatusFound)
}

// AdminSettings handles GET (show settings form) and POST (save settings).
func AdminSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		m := map[string]string{
			"site_title":       r.FormValue("site_title"),
			"site_description": r.FormValue("site_description"),
			"company_name":     r.FormValue("company_name"),
			"phone1":           r.FormValue("phone1"),
			"phone2":           r.FormValue("phone2"),
			"phone_hr":         r.FormValue("phone_hr"),
			"email":            r.FormValue("email"),
			"address":          r.FormValue("address"),
			"address2":         r.FormValue("address2"),
			"working_hours":    r.FormValue("working_hours"),
			"hero_title":       r.FormValue("hero_title"),
			"hero_subtitle":    r.FormValue("hero_subtitle"),
			"about_text":       r.FormValue("about_text"),
			"stats_years":      r.FormValue("stats_years"),
			"stats_objects":    r.FormValue("stats_objects"),
			"stats_staff":      r.FormValue("stats_staff"),
			"domain":           r.FormValue("domain"),
		}

		// Handle password change.
		newPwd := strings.TrimSpace(r.FormValue("new_password"))
		if newPwd != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("admin: settings: hash password: %v", err)
			} else {
				m["admin_password_hash"] = string(hash)
			}
		}

		if err := models.SetSettings(m); err != nil {
			log.Printf("admin: settings: save: %v", err)
		}

		http.Redirect(w, r, "/admin/settings?saved=1", http.StatusFound)
		return
	}

	settings, err := models.GetSettings()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		adminData
		Settings *models.SiteSettings
		Saved    bool
	}{
		adminData: baseAdmin("Настройки сайта"),
		Settings:  settings,
		Saved:     r.URL.Query().Get("saved") == "1",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "settings.html", data); err != nil {
		log.Printf("admin: settings render: %v", err)
	}
}

// AdminServices handles service listing and inline edit via POST.
func AdminServices(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		idStr := r.FormValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}

		active := r.FormValue("active") == "1"
		if err := models.UpdateService(id,
			r.FormValue("name"),
			r.FormValue("description"),
			active,
		); err != nil {
			log.Printf("admin: services: update %d: %v", id, err)
		}

		http.Redirect(w, r, "/admin/services?saved=1", http.StatusFound)
		return
	}

	services, err := models.GetAllServices()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		adminData
		Services []models.Service
		Saved    bool
	}{
		adminData: baseAdmin("Управление услугами"),
		Services:  services,
		Saved:     r.URL.Query().Get("saved") == "1",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "services.html", data); err != nil {
		log.Printf("admin: services render: %v", err)
	}
}

// AdminContacts shows all contact form submissions.
func AdminContacts(w http.ResponseWriter, r *http.Request) {
	// Mark as read if id provided.
	if idStr := r.URL.Query().Get("read"); idStr != "" {
		if id, err := strconv.Atoi(idStr); err == nil {
			_ = models.MarkContactRead(id)
		}
	}

	contacts, err := models.GetContacts()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		adminData
		Contacts []models.Contact
	}{
		adminData: baseAdmin("Заявки"),
		Contacts:  contacts,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "contacts.html", data); err != nil {
		log.Printf("admin: contacts render: %v", err)
	}
}

// AdminSEO handles the SEO management page.
func AdminSEO(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		slug := r.FormValue("slug")
		if slug == "" {
			slug = "/"
		}

		page := &models.SEOPage{
			Slug:        slug,
			Title:       r.FormValue("title"),
			Description: r.FormValue("description"),
			OGImage:     r.FormValue("og_image"),
			SchemaJSON:  r.FormValue("schema_json"),
		}

		if err := models.UpsertSEOPage(page); err != nil {
			log.Printf("admin: seo: upsert: %v", err)
		}

		// Update robots.txt if submitted.
		if robots := r.FormValue("robots_txt"); robots != "" {
			_ = models.SetSetting("robots_txt", robots)
		}

		http.Redirect(w, r, "/admin/seo?saved=1", http.StatusFound)
		return
	}

	seoPages, _ := models.GetAllSEOPages()
	homeSEO, _ := models.GetSEOPage("/")
	robots := models.GetSetting("robots_txt")

	data := struct {
		adminData
		Pages    []models.SEOPage
		HomeSEO  *models.SEOPage
		Robots   string
		Saved    bool
	}{
		adminData: baseAdmin("SEO-инструменты"),
		Pages:     seoPages,
		HomeSEO:   homeSEO,
		Robots:    robots,
		Saved:     r.URL.Query().Get("saved") == "1",
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "seo.html", data); err != nil {
		log.Printf("admin: seo render: %v", err)
	}
}

func renderLoginError(w http.ResponseWriter, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = AdminTemplates.ExecuteTemplate(w, "login.html", map[string]string{"Error": errMsg})
}

// ── Advantages (Преимущества) ──────────────────────────────────────────────

// AdminAdvantages handles GET (list) and POST (save/delete) for advantages.
func AdminAdvantages(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if r.FormValue("_action") == "delete" {
			id, _ := strconv.Atoi(r.FormValue("id"))
			_ = models.DeleteAdvantage(id)
			http.Redirect(w, r, "/admin/advantages?saved=1", http.StatusFound)
			return
		}
		id, _ := strconv.Atoi(r.FormValue("id"))
		active := r.FormValue("active") == "1"
		a := &models.Advantage{
			ID:          id,
			Title:       r.FormValue("title"),
			Description: r.FormValue("description"),
			Icon:        r.FormValue("icon"),
			SortOrder:   mustAtoi(r.FormValue("sort_order")),
			Active:      active,
		}
		if err := models.SaveAdvantage(a); err != nil {
			log.Printf("admin: advantages: save: %v", err)
		}
		http.Redirect(w, r, "/admin/advantages?saved=1", http.StatusFound)
		return
	}
	items, err := models.GetAdvantages()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := struct {
		adminData
		Items []models.Advantage
		Saved bool
	}{
		adminData: baseAdmin("Преимущества"),
		Items:     items,
		Saved:     r.URL.Query().Get("saved") == "1",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "advantages.html", data); err != nil {
		log.Printf("admin: advantages render: %v", err)
	}
}

// ── History (История) ─────────────────────────────────────────────────────

// AdminHistory handles GET (list) and POST (save/delete) for history events.
func AdminHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if r.FormValue("_action") == "delete" {
			id, _ := strconv.Atoi(r.FormValue("id"))
			_ = models.DeleteHistoryEvent(id)
			http.Redirect(w, r, "/admin/history?saved=1", http.StatusFound)
			return
		}
		id, _ := strconv.Atoi(r.FormValue("id"))
		h := &models.HistoryEvent{
			ID:          id,
			Year:        r.FormValue("year"),
			Title:       r.FormValue("title"),
			Description: r.FormValue("description"),
			SortOrder:   mustAtoi(r.FormValue("sort_order")),
		}
		if err := models.SaveHistoryEvent(h); err != nil {
			log.Printf("admin: history: save: %v", err)
		}
		http.Redirect(w, r, "/admin/history?saved=1", http.StatusFound)
		return
	}
	items, err := models.GetHistoryEvents()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := struct {
		adminData
		Items []models.HistoryEvent
		Saved bool
	}{
		adminData: baseAdmin("История"),
		Items:     items,
		Saved:     r.URL.Query().Get("saved") == "1",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "history.html", data); err != nil {
		log.Printf("admin: history render: %v", err)
	}
}

// ── Licenses (Лицензии) ───────────────────────────────────────────────────

// AdminLicenses handles GET (list) and POST (save/delete) for licenses.
func AdminLicenses(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if r.FormValue("_action") == "delete" {
			id, _ := strconv.Atoi(r.FormValue("id"))
			_ = models.DeleteLicense(id)
			http.Redirect(w, r, "/admin/licenses?saved=1", http.StatusFound)
			return
		}
		id, _ := strconv.Atoi(r.FormValue("id"))
		l := &models.License{
			ID:        id,
			Name:      r.FormValue("name"),
			Number:    r.FormValue("number"),
			Issuer:    r.FormValue("issuer"),
			IssuedAt:  r.FormValue("issued_at"),
			SortOrder: mustAtoi(r.FormValue("sort_order")),
		}
		if err := models.SaveLicense(l); err != nil {
			log.Printf("admin: licenses: save: %v", err)
		}
		http.Redirect(w, r, "/admin/licenses?saved=1", http.StatusFound)
		return
	}
	items, err := models.GetLicenses()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := struct {
		adminData
		Items []models.License
		Saved bool
	}{
		adminData: baseAdmin("Лицензии"),
		Items:     items,
		Saved:     r.URL.Query().Get("saved") == "1",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "licenses.html", data); err != nil {
		log.Printf("admin: licenses render: %v", err)
	}
}

// ── Clients (Клиенты) ─────────────────────────────────────────────────────

// AdminClients handles GET (list) and POST (save/delete) for clients.
func AdminClients(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if r.FormValue("_action") == "delete" {
			id, _ := strconv.Atoi(r.FormValue("id"))
			_ = models.DeleteClient(id)
			http.Redirect(w, r, "/admin/clients?saved=1", http.StatusFound)
			return
		}
		id, _ := strconv.Atoi(r.FormValue("id"))
		c := &models.Client{
			ID:          id,
			Name:        r.FormValue("name"),
			Sector:      r.FormValue("sector"),
			Description: r.FormValue("description"),
			SortOrder:   mustAtoi(r.FormValue("sort_order")),
		}
		if err := models.SaveClient(c); err != nil {
			log.Printf("admin: clients: save: %v", err)
		}
		http.Redirect(w, r, "/admin/clients?saved=1", http.StatusFound)
		return
	}
	items, err := models.GetClients()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := struct {
		adminData
		Items []models.Client
		Saved bool
	}{
		adminData: baseAdmin("Клиенты"),
		Items:     items,
		Saved:     r.URL.Query().Get("saved") == "1",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "clients.html", data); err != nil {
		log.Printf("admin: clients render: %v", err)
	}
}

// ── Team (Команда) ────────────────────────────────────────────────────────

// AdminTeam handles GET (list) and POST (save/delete) for team members.
func AdminTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if r.FormValue("_action") == "delete" {
			id, _ := strconv.Atoi(r.FormValue("id"))
			_ = models.DeleteTeamMember(id)
			http.Redirect(w, r, "/admin/team?saved=1", http.StatusFound)
			return
		}
		id, _ := strconv.Atoi(r.FormValue("id"))
		active := r.FormValue("active") == "1"
		m := &models.TeamMember{
			ID:        id,
			Name:      r.FormValue("name"),
			Position:  r.FormValue("position"),
			Bio:       r.FormValue("bio"),
			PhotoURL:  r.FormValue("photo_url"),
			SortOrder: mustAtoi(r.FormValue("sort_order")),
			Active:    active,
		}
		if err := models.SaveTeamMember(m); err != nil {
			log.Printf("admin: team: save: %v", err)
		}
		http.Redirect(w, r, "/admin/team?saved=1", http.StatusFound)
		return
	}
	items, err := models.GetTeamMembers()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := struct {
		adminData
		Items []models.TeamMember
		Saved bool
	}{
		adminData: baseAdmin("Команда"),
		Items:     items,
		Saved:     r.URL.Query().Get("saved") == "1",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "team.html", data); err != nil {
		log.Printf("admin: team render: %v", err)
	}
}

// ── Workflow (Как мы работаем) ────────────────────────────────────────────

// AdminWorkflow handles GET (list) and POST (save/delete) for workflow steps.
func AdminWorkflow(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if r.FormValue("_action") == "delete" {
			id, _ := strconv.Atoi(r.FormValue("id"))
			_ = models.DeleteWorkflowStep(id)
			http.Redirect(w, r, "/admin/workflow?saved=1", http.StatusFound)
			return
		}
		id, _ := strconv.Atoi(r.FormValue("id"))
		s := &models.WorkflowStep{
			ID:          id,
			StepNumber:  mustAtoi(r.FormValue("step_number")),
			Title:       r.FormValue("title"),
			Description: r.FormValue("description"),
			SortOrder:   mustAtoi(r.FormValue("sort_order")),
		}
		if err := models.SaveWorkflowStep(s); err != nil {
			log.Printf("admin: workflow: save: %v", err)
		}
		http.Redirect(w, r, "/admin/workflow?saved=1", http.StatusFound)
		return
	}
	items, err := models.GetWorkflowSteps()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := struct {
		adminData
		Items []models.WorkflowStep
		Saved bool
	}{
		adminData: baseAdmin("Как мы работаем"),
		Items:     items,
		Saved:     r.URL.Query().Get("saved") == "1",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "workflow.html", data); err != nil {
		log.Printf("admin: workflow render: %v", err)
	}
}

// ── FAQ ───────────────────────────────────────────────────────────────────

// AdminFAQ handles GET (list) and POST (save/delete) for FAQ items.
func AdminFAQ(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if r.FormValue("_action") == "delete" {
			id, _ := strconv.Atoi(r.FormValue("id"))
			_ = models.DeleteFAQItem(id)
			http.Redirect(w, r, "/admin/faq?saved=1", http.StatusFound)
			return
		}
		id, _ := strconv.Atoi(r.FormValue("id"))
		active := r.FormValue("active") == "1"
		f := &models.FAQItem{
			ID:        id,
			Question:  r.FormValue("question"),
			Answer:    r.FormValue("answer"),
			SortOrder: mustAtoi(r.FormValue("sort_order")),
			Active:    active,
		}
		if err := models.SaveFAQItem(f); err != nil {
			log.Printf("admin: faq: save: %v", err)
		}
		http.Redirect(w, r, "/admin/faq?saved=1", http.StatusFound)
		return
	}
	items, err := models.GetFAQItems()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	data := struct {
		adminData
		Items []models.FAQItem
		Saved bool
	}{
		adminData: baseAdmin("FAQ"),
		Items:     items,
		Saved:     r.URL.Query().Get("saved") == "1",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "faq.html", data); err != nil {
		log.Printf("admin: faq render: %v", err)
	}
}

// mustAtoi converts a string to int, returning 0 on failure.
func mustAtoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
