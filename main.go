// Альфа Юнит-1 — корпоративный сайт охранной компании.
// Запуск: go run main.go  или  docker compose up -d
package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"

	"alfaunit1/internal/db"
	"alfaunit1/internal/handlers"
	"alfaunit1/internal/middleware"
	"alfaunit1/internal/models"
	"golang.org/x/crypto/bcrypt"
)

//go:embed templates static
var embeddedFiles embed.FS

func main() {
	// ── Configuration ─────────────────────────────────────────────────────────
	port := envOr("PORT", "8080")
	dbPath := envOr("DB_PATH", "./data/alfa1.db")

	// ── Database ──────────────────────────────────────────────────────────────
	if err := db.Init(dbPath); err != nil {
		log.Fatalf("database init: %v", err)
	}

	// Apply ADMIN_PASSWORD from environment on every startup.
	// This allows changing the password by editing .env and restarting.
	if pwd := os.Getenv("ADMIN_PASSWORD"); pwd != "" {
		// Only re-hash if the env password differs from the stored one
		// (avoids unnecessary bcrypt work on every restart).
		stored := models.GetSetting("admin_password_hash")
		needUpdate := stored == "" // first boot
		if !needUpdate {
			// If stored hash can't be verified against the current env
			// password, the password was changed in .env → update it.
			if err := bcrypt.CompareHashAndPassword([]byte(stored), []byte(pwd)); err != nil {
				needUpdate = true
			}
		}
		if needUpdate {
			hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
			if err != nil {
				log.Fatalf("bcrypt: %v", err)
			}
			if err := models.SetSetting("admin_password_hash", string(hash)); err != nil {
				log.Fatalf("set admin password: %v", err)
			}
			log.Println("main: admin password updated from ADMIN_PASSWORD env")
		}
	}

	// Set domain from environment if provided.
	if domain := os.Getenv("SITE_DOMAIN"); domain != "" {
		_ = models.SetSetting("domain", domain)
	}

	// ── Templates ─────────────────────────────────────────────────────────────
	funcMap := template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) }, //nolint:gosec
		"mul":      func(a, b int) int { return a * b },
		"add":      func(a, b int) int { return a + b },
	}

	// Public site templates.
	publicTmpl, err := template.New("").Funcs(funcMap).ParseFS(embeddedFiles, "templates/*.html")
	if err != nil {
		log.Fatalf("parse public templates: %v", err)
	}
	handlers.Templates = publicTmpl

	// Admin templates.
	adminTmpl, err := template.New("").Funcs(funcMap).ParseFS(embeddedFiles, "templates/admin/*.html")
	if err != nil {
		log.Fatalf("parse admin templates: %v", err)
	}
	handlers.AdminTemplates = adminTmpl

	// ── Static files ──────────────────────────────────────────────────────────
	staticFS, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Fatalf("static fs: %v", err)
	}
	staticServer := http.FileServer(http.FS(staticFS))

	// ── Router ────────────────────────────────────────────────────────────────
	mux := http.NewServeMux()

	// Static assets.
	mux.Handle("/static/", http.StripPrefix("/static/", staticServer))

	// SEO files.
	mux.HandleFunc("/robots.txt", handlers.RobotsTxt)
	mux.HandleFunc("/sitemap.xml", handlers.Sitemap)

	// Public site.
	mux.HandleFunc("/", handlers.Home)
	mux.HandleFunc("/contact", handlers.Contact)

	// Admin — unauthenticated routes.
	mux.HandleFunc("/admin/login", handlers.AdminLogin)
	mux.HandleFunc("/admin/logout", handlers.AdminLogout)

	// Admin — authenticated routes (wrapped with RequireAuth).
	adminRoutes := http.NewServeMux()
	adminRoutes.HandleFunc("/admin/", handlers.AdminDashboard)
	adminRoutes.HandleFunc("/admin/settings", handlers.AdminSettings)
	adminRoutes.HandleFunc("/admin/services", handlers.AdminServices)
	adminRoutes.HandleFunc("/admin/contacts", handlers.AdminContacts)
	adminRoutes.HandleFunc("/admin/seo", handlers.AdminSEO)

	mux.Handle("/admin/", middleware.RequireAuth(adminRoutes))

	// ── Security headers middleware ────────────────────────────────────────────
	handler := securityHeaders(mux)

	// ── Start ─────────────────────────────────────────────────────────────────
	addr := ":" + port
	log.Printf("main: Альфа Юнит-1 started on http://localhost%s", addr)
	log.Printf("main: admin panel → http://localhost%s/admin/", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// securityHeaders adds secure HTTP response headers.
func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME sniffing.
		w.Header().Set("X-Content-Type-Options", "nosniff")
		// Deny embedding in iframes.
		w.Header().Set("X-Frame-Options", "DENY")
		// Disable XSS auditor (legacy, but harmless).
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		// Referrer policy.
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Remove Server header.
		w.Header().Del("Server")

		// Minimal CSP – allow CDN resources used by the frontend.
		if !strings.HasPrefix(r.URL.Path, "/admin") {
			w.Header().Set("Content-Security-Policy",
				"default-src 'self'; "+
					"script-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com https://cdn.tailwindcss.com https://unpkg.com; "+
					"style-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com https://cdn.tailwindcss.com https://unpkg.com https://fonts.googleapis.com; "+
					"font-src 'self' https://fonts.gstatic.com; "+
					"img-src 'self' data: https:; "+
					"connect-src 'self'; "+
					"frame-src https://www.google.com https://yandex.ru;",
			)
		}

		next.ServeHTTP(w, r)
	})
}

// envOr returns the value of the environment variable key, or fallback if empty.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
