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
