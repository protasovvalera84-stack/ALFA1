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

// ── Dashboard ─────────────────────────────────────────────────────────────────

// AdminDashboard renders the admin dashboard.
func AdminDashboard(w http.ResponseWriter, r *http.Request) {
	settings, err := models.GetSettings()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	svcs, _ := models.GetServices()

	data := struct {
		adminData
		Settings     *models.SiteSettings
		ServiceCount int
	}{
		adminData:    baseAdmin("Dashboard — Альфа Юнит-1"),
		Settings:     settings,
		ServiceCount: len(svcs),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		log.Printf("admin: dashboard render: %v", err)
	}
}

// ── Auth ──────────────────────────────────────────────────────────────────────

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

// ── Site settings ─────────────────────────────────────────────────────────────

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
			"whatsapp_link":    r.FormValue("whatsapp_link"),
			"telegram_link":    r.FormValue("telegram_link"),
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

// ── Services ──────────────────────────────────────────────────────────────────

// AdminServices handles service listing, inline edit, add and delete via POST.
func AdminServices(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		action := r.FormValue("action")
		switch action {
		case "add":
			if err := models.AddService(
				r.FormValue("name"),
				r.FormValue("description"),
			); err != nil {
				log.Printf("admin: services: add: %v", err)
			}
		case "delete":
			if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
				_ = models.DeleteService(id)
			}
		default:
			// update existing
			id, err := strconv.Atoi(r.FormValue("id"))
			if err != nil {
				http.Error(w, "invalid id", http.StatusBadRequest)
				return
			}
			active := r.FormValue("active") == "1"
			if err := models.UpdateService(id, r.FormValue("name"), r.FormValue("description"), active); err != nil {
				log.Printf("admin: services: update %d: %v", id, err)
			}
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

// ── Contacts ──────────────────────────────────────────────────────────────────

// AdminContacts shows all contact form submissions.
func AdminContacts(w http.ResponseWriter, r *http.Request) {
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

// ── SEO ───────────────────────────────────────────────────────────────────────

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
		Pages   []models.SEOPage
		HomeSEO *models.SEOPage
		Robots  string
		Saved   bool
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

// ── Clients ───────────────────────────────────────────────────────────────────

// AdminClients manages client / protected-object cards.
func AdminClients(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		switch r.FormValue("action") {
		case "add":
			_ = models.AddClient(
				r.FormValue("initial"),
				r.FormValue("name"),
				r.FormValue("type_label"),
				r.FormValue("description"),
				r.FormValue("tags"),
			)
		case "delete":
			if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
				_ = models.DeleteClient(id)
			}
		default:
			if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
				_ = models.UpdateClient(id,
					r.FormValue("initial"),
					r.FormValue("name"),
					r.FormValue("type_label"),
					r.FormValue("description"),
					r.FormValue("tags"),
					r.FormValue("active") == "1",
				)
			}
		}
		http.Redirect(w, r, "/admin/clients?saved=1", http.StatusFound)
		return
	}

	items, err := models.GetAllClients()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		adminData
		Items []models.Client
		Saved bool
	}{
		adminData: baseAdmin("Наши клиенты"),
		Items:     items,
		Saved:     r.URL.Query().Get("saved") == "1",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "admin_clients.html", data); err != nil {
		log.Printf("admin: clients render: %v", err)
	}
}

// ── Team ──────────────────────────────────────────────────────────────────────

// AdminTeam manages team member cards.
func AdminTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		switch r.FormValue("action") {
		case "add":
			_ = models.AddTeamMember(
				r.FormValue("initial"),
				r.FormValue("name"),
				r.FormValue("role"),
				r.FormValue("department"),
				r.FormValue("description"),
				r.FormValue("tags"),
			)
		case "delete":
			if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
				_ = models.DeleteTeamMember(id)
			}
		default:
			if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
				_ = models.UpdateTeamMember(id,
					r.FormValue("initial"),
					r.FormValue("name"),
					r.FormValue("role"),
					r.FormValue("department"),
					r.FormValue("description"),
					r.FormValue("tags"),
					r.FormValue("active") == "1",
				)
			}
		}
		http.Redirect(w, r, "/admin/team?saved=1", http.StatusFound)
		return
	}

	items, err := models.GetAllTeamMembers()
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
	if err := AdminTemplates.ExecuteTemplate(w, "admin_team.html", data); err != nil {
		log.Printf("admin: team render: %v", err)
	}
}

// ── Process ───────────────────────────────────────────────────────────────────

// AdminProcess manages "how we work" process steps.
func AdminProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		switch r.FormValue("action") {
		case "add":
			_ = models.AddProcessStep(
				r.FormValue("step_num"),
				r.FormValue("title"),
				r.FormValue("description"),
			)
		case "delete":
			if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
				_ = models.DeleteProcessStep(id)
			}
		default:
			if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
				_ = models.UpdateProcessStep(id,
					r.FormValue("step_num"),
					r.FormValue("title"),
					r.FormValue("description"),
					r.FormValue("active") == "1",
				)
			}
		}
		http.Redirect(w, r, "/admin/process?saved=1", http.StatusFound)
		return
	}

	items, err := models.GetAllProcessSteps()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		adminData
		Items []models.ProcessStep
		Saved bool
	}{
		adminData: baseAdmin("Как мы работаем"),
		Items:     items,
		Saved:     r.URL.Query().Get("saved") == "1",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "admin_process.html", data); err != nil {
		log.Printf("admin: process render: %v", err)
	}
}

// ── FAQ ───────────────────────────────────────────────────────────────────────

// AdminFAQ manages FAQ items.
func AdminFAQ(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		switch r.FormValue("action") {
		case "add":
			_ = models.AddFAQItem(r.FormValue("question"), r.FormValue("answer"))
		case "delete":
			if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
				_ = models.DeleteFAQItem(id)
			}
		default:
			if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
				_ = models.UpdateFAQItem(id,
					r.FormValue("question"),
					r.FormValue("answer"),
					r.FormValue("active") == "1",
				)
			}
		}
		http.Redirect(w, r, "/admin/faq?saved=1", http.StatusFound)
		return
	}

	items, err := models.GetAllFAQItems()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		adminData
		Items []models.FAQItem
		Saved bool
	}{
		adminData: baseAdmin("Вопросы и ответы"),
		Items:     items,
		Saved:     r.URL.Query().Get("saved") == "1",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "admin_faq.html", data); err != nil {
		log.Printf("admin: faq render: %v", err)
	}
}

// ── Advantages ────────────────────────────────────────────────────────────────

// AdminAdvantages manages "why choose us" advantage cards.
func AdminAdvantages(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		switch r.FormValue("action") {
		case "add":
			_ = models.AddAdvantage(r.FormValue("title"), r.FormValue("description"))
		case "delete":
			if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
				_ = models.DeleteAdvantage(id)
			}
		default:
			if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
				_ = models.UpdateAdvantage(id,
					r.FormValue("title"),
					r.FormValue("description"),
					r.FormValue("active") == "1",
				)
			}
		}
		http.Redirect(w, r, "/admin/advantages?saved=1", http.StatusFound)
		return
	}

	items, err := models.GetAllAdvantages()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		adminData
		Items []models.Advantage
		Saved bool
	}{
		adminData: baseAdmin("Почему выбирают нас"),
		Items:     items,
		Saved:     r.URL.Query().Get("saved") == "1",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "admin_advantages.html", data); err != nil {
		log.Printf("admin: advantages render: %v", err)
	}
}

// ── History ───────────────────────────────────────────────────────────────────

// AdminHistory manages company history timeline events.
func AdminHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		switch r.FormValue("action") {
		case "add":
			_ = models.AddHistoryEvent(
				r.FormValue("year_label"),
				r.FormValue("title"),
				r.FormValue("description"),
				r.FormValue("quote"),
			)
		case "delete":
			if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
				_ = models.DeleteHistoryEvent(id)
			}
		default:
			if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
				_ = models.UpdateHistoryEvent(id,
					r.FormValue("year_label"),
					r.FormValue("title"),
					r.FormValue("description"),
					r.FormValue("quote"),
					r.FormValue("active") == "1",
				)
			}
		}
		http.Redirect(w, r, "/admin/history?saved=1", http.StatusFound)
		return
	}

	items, err := models.GetAllHistoryEvents()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		adminData
		Items []models.HistoryEvent
		Saved bool
	}{
		adminData: baseAdmin("История компании"),
		Items:     items,
		Saved:     r.URL.Query().Get("saved") == "1",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "admin_history.html", data); err != nil {
		log.Printf("admin: history render: %v", err)
	}
}

// ── Licenses ──────────────────────────────────────────────────────────────────

// AdminLicenses manages license / certificate cards.
func AdminLicenses(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		switch r.FormValue("action") {
		case "add":
			_ = models.AddLicense(
				r.FormValue("type_label"),
				r.FormValue("company"),
				r.FormValue("description"),
				r.FormValue("status_text"),
			)
		case "delete":
			if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
				_ = models.DeleteLicense(id)
			}
		default:
			if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
				_ = models.UpdateLicense(id,
					r.FormValue("type_label"),
					r.FormValue("company"),
					r.FormValue("description"),
					r.FormValue("status_text"),
					r.FormValue("active") == "1",
				)
			}
		}
		http.Redirect(w, r, "/admin/licenses?saved=1", http.StatusFound)
		return
	}

	items, err := models.GetAllLicenses()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		adminData
		Items []models.License
		Saved bool
	}{
		adminData: baseAdmin("Лицензии и документы"),
		Items:     items,
		Saved:     r.URL.Query().Get("saved") == "1",
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := AdminTemplates.ExecuteTemplate(w, "admin_licenses.html", data); err != nil {
		log.Printf("admin: licenses render: %v", err)
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func renderLoginError(w http.ResponseWriter, errMsg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = AdminTemplates.ExecuteTemplate(w, "login.html", map[string]string{"Error": errMsg})
}
