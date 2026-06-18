// Package handlers contains all HTTP request handlers for the site.
package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"alfaunit1/internal/models"
)

// Templates is the parsed template set, assigned once during server startup.
var Templates *template.Template

// HomeData holds all data required to render the main page template.
type HomeData struct {
	Settings *models.SiteSettings
	Services []models.Service
	SEO      *models.SEOPage
	Year     int
}

// Home renders the main single-page site.
func Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	settings, err := models.GetSettings()
	if err != nil {
		log.Printf("handlers: Home: get settings: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	services, err := models.GetServices()
	if err != nil {
		log.Printf("handlers: Home: get services: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	seo, _ := models.GetSEOPage("/")

	data := HomeData{
		Settings: settings,
		Services: services,
		SEO:      seo,
		Year:     time.Now().Year(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := Templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("handlers: Home: render: %v", err)
	}
}

// contactRequest is the JSON body expected from the contact form.
type contactRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Message string `json:"message"`
	Service string `json:"service"`
}

// rateLimiter tracks per-IP submission counts to prevent form spam.
var (
	rateMu    sync.Mutex
	rateStore = make(map[string][]time.Time)
)

func isRateLimited(ip string) bool {
	rateMu.Lock()
	defer rateMu.Unlock()

	now := time.Now()
	window := now.Add(-time.Minute)

	// Filter to last minute.
	var recent []time.Time
	for _, t := range rateStore[ip] {
		if t.After(window) {
			recent = append(recent, t)
		}
	}
	rateStore[ip] = append(recent, now)

	return len(recent) >= 5 // max 5 per minute
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return strings.SplitN(xff, ",", 2)[0]
	}
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}
	// RemoteAddr is "host:port".
	if idx := strings.LastIndex(r.RemoteAddr, ":"); idx >= 0 {
		return r.RemoteAddr[:idx]
	}
	return r.RemoteAddr
}

// Contact handles POST requests from the contact form (JSON body).
func Contact(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if isRateLimited(clientIP(r)) {
		jsonError(w, "Слишком много запросов. Попробуйте позже.", http.StatusTooManyRequests)
		return
	}

	var req contactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "Неверный формат данных.", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Phone = strings.TrimSpace(req.Phone)
	req.Email = strings.TrimSpace(req.Email)
	req.Message = strings.TrimSpace(req.Message)

	if req.Name == "" || req.Phone == "" {
		jsonError(w, "Укажите имя и телефон.", http.StatusBadRequest)
		return
	}

	if err := models.SaveContact(req.Name, req.Phone, req.Email, req.Message, req.Service); err != nil {
		log.Printf("handlers: Contact: save: %v", err)
		jsonError(w, "Ошибка сохранения. Позвоните нам напрямую.", http.StatusInternalServerError)
		return
	}

	log.Printf("handlers: Contact: new inquiry from %s (%s)", req.Name, req.Phone)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "Ваша заявка принята! Мы свяжемся с вами в ближайшее время.",
	})
}

// Sitemap generates and serves a dynamic sitemap.xml.
func Sitemap(w http.ResponseWriter, r *http.Request) {
	domain := models.GetSetting("domain")
	if domain == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		domain = fmt.Sprintf("%s://%s", scheme, r.Host)
	}

	pages := []struct{ Loc, LastMod, Priority string }{
		{domain + "/", time.Now().Format("2006-01-02"), "1.0"},
		{domain + "/#services", time.Now().Format("2006-01-02"), "0.9"},
		{domain + "/#about", time.Now().Format("2006-01-02"), "0.8"},
		{domain + "/#contacts", time.Now().Format("2006-01-02"), "0.8"},
	}

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	b.WriteString("\n<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\"\n")
	b.WriteString("        xmlns:xhtml=\"http://www.w3.org/1999/xhtml\">\n")

	for _, p := range pages {
		b.WriteString("  <url>\n")
		b.WriteString(fmt.Sprintf("    <loc>%s</loc>\n", p.Loc))
		b.WriteString(fmt.Sprintf("    <lastmod>%s</lastmod>\n", p.LastMod))
		b.WriteString(fmt.Sprintf("    <changefreq>monthly</changefreq>\n"))
		b.WriteString(fmt.Sprintf("    <priority>%s</priority>\n", p.Priority))
		b.WriteString("  </url>\n")
	}
	b.WriteString("</urlset>\n")

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	fmt.Fprint(w, b.String())
}

// RobotsTxt serves the robots.txt file using content stored in the database.
func RobotsTxt(w http.ResponseWriter, r *http.Request) {
	content := models.GetSetting("robots_txt")
	if content == "" {
		content = "User-agent: *\nAllow: /\nDisallow: /admin/\n\nSitemap: /sitemap.xml\n"
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, content)
}

// jsonError writes a JSON error response.
func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
