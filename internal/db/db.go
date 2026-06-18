// Package db handles SQLite database initialization and migrations.
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // SQLite driver (pure Go, no CGO required)
)

// DB is the global database connection handle.
var DB *sql.DB

// Init opens (or creates) the SQLite database at the given path and runs all
// schema migrations. It must be called once at program startup.
func Init(dbPath string) error {
	// Ensure the directory for the database file exists.
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("db: create directory %q: %w", dir, err)
	}

	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return fmt.Errorf("db: open %q: %w", dbPath, err)
	}

	// Verify the connection.
	if err := db.Ping(); err != nil {
		return fmt.Errorf("db: ping: %w", err)
	}

	// Tune connection pool for a single-file SQLite database.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	DB = db

	log.Printf("db: connected to %s", dbPath)
	return migrate(db)
}

// migrate creates all required tables if they do not yet exist and seeds
// initial data on first run.
func migrate(db *sql.DB) error {
	stmts := []string{
		// ── Settings ──────────────────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL DEFAULT ''
		)`,

		// ── Services ──────────────────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS services (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT    NOT NULL,
			description TEXT    NOT NULL DEFAULT '',
			icon        TEXT    NOT NULL DEFAULT '',
			sort_order  INTEGER NOT NULL DEFAULT 0,
			active      INTEGER NOT NULL DEFAULT 1
		)`,

		// ── SEO per-page settings ─────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS seo_pages (
			slug        TEXT PRIMARY KEY,
			title       TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			og_image    TEXT NOT NULL DEFAULT '',
			schema_json TEXT NOT NULL DEFAULT ''
		)`,

		// ── Contact form submissions ───────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS contacts (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT    NOT NULL,
			phone      TEXT    NOT NULL,
			email      TEXT    NOT NULL DEFAULT '',
			message    TEXT    NOT NULL DEFAULT '',
			service    TEXT    NOT NULL DEFAULT '',
			created_at TEXT    NOT NULL DEFAULT (datetime('now')),
			read       INTEGER NOT NULL DEFAULT 0
		)`,

		// ── Admin sessions ────────────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS sessions (
			token      TEXT    PRIMARY KEY,
			expires_at TEXT    NOT NULL
		)`,
	}

	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("db: migrate: %w", err)
		}
	}

	if err := seedDefaults(db); err != nil {
		return fmt.Errorf("db: seed: %w", err)
	}

	return nil
}

// seedDefaults inserts the initial site content if the database is empty.
func seedDefaults(db *sql.DB) error {
	// Check whether settings have been seeded already.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM settings`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // already seeded
	}

	log.Println("db: seeding initial data...")

	defaults := map[string]string{
		"site_title":       "Альфа Юнит-1 — Охранная компания в Санкт-Петербурге",
		"site_description": "Лицензированная охранная компания в СПб. Вооружённая и невооружённая охрана объектов. Члены ассоциации ветеранов «Альфа». Звоните: +7 (931) 362-56-88",
		"company_name":     "Альфа Юнит-1",
		"phone1":           "+7 (931) 362-56-88",
		"phone2":           "+7 (921) 946-21-97",
		"phone_hr":         "+7 (921) 884-33-88",
		"email":            "admin@alfaunit1.ru",
		"address":          "190020, Санкт-Петербург, ул. Лифляндская, д. 3",
		"address2":         "Симферополь, ул. Карла Маркса, 14",
		"working_hours":    "Пн–Пт: 9:00–20:00",
		"founded_year":     "2002",
		"hero_title":       "Комплексная безопасность объектов любой сложности",
		"hero_subtitle":    "Вооружённая и невооружённая охрана. Санкт-Петербург и Северо-Запад России.",
		"about_text":       "Группа компаний «Альфа Юнит-1» — лицензированная охранная организация, действующая в Санкт-Петербурге с 2002 года. Мы являемся членами Международной ассоциации ветеранов подразделения «АЛЬФА» и предлагаем полный спектр охранных услуг для объектов любой сложности.",
		"stats_years":      "23",
		"stats_objects":    "50+",
		"stats_staff":      "200+",
		"stats_licenses":   "2",
		"robots_txt":       "User-agent: *\nAllow: /\nDisallow: /admin/\n\nSitemap: /sitemap.xml",
		// Admin credentials (password set via env on first boot)
		"admin_password_hash": "",
	}

	for k, v := range defaults {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO settings (key, value) VALUES (?, ?)`, k, v,
		); err != nil {
			return err
		}
	}

	// Seed initial services.
	services := []struct{ name, desc, icon string }{
		{"Вооружённая охрана", "Стационарные и мобильные посты с вооружёнными сотрудниками для объектов повышенного уровня безопасности.", `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.955 11.955 0 01.42 12c0 2.01.5 3.903 1.378 5.56A11.956 11.956 0 013.6 18M12 2.764A11.959 11.959 0 0120.402 6 11.955 11.955 0 0123.58 12a11.955 11.955 0 01-3.177 5.56A11.956 11.956 0 0112 21.236" /></svg>`},
		{"Невооружённая охрана", "Профессиональная охрана без огнестрельного оружия для офисов, магазинов и общественных мест.", `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" /></svg>`},
		{"Охрана бизнес-центров", "Организация многоуровневой системы охраны деловых центров, контроль доступа и безопасности арендаторов.", `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3.75h.008v.008h-.008v-.008zm0 3h.008v.008h-.008v-.008zm0 3h.008v.008h-.008v-.008z" /></svg>`},
		{"Охрана складов", "Круглосуточная охрана складских комплексов с организацией пропускного режима и видеонаблюдения.", `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M20.25 7.5l-.625 10.632a2.25 2.25 0 01-2.247 2.118H6.622a2.25 2.25 0 01-2.247-2.118L3.75 7.5M10 11.25h4M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z" /></svg>`},
		{"Охрана производств", "Защита производственных объектов, промышленных предприятий и заводов с контролем въезда/выезда.", `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M11.42 15.17L17.25 21A2.652 2.652 0 0021 17.25l-5.877-5.877M11.42 15.17l2.496-3.03c.317-.384.74-.626 1.208-.766M11.42 15.17l-4.655 5.653a2.548 2.548 0 11-3.586-3.586l6.837-5.63m5.108-.233c.55-.164 1.163-.188 1.743-.14a4.5 4.5 0 004.486-6.336l-3.276 3.277a3.004 3.004 0 01-2.25-2.25l3.276-3.276a4.5 4.5 0 00-6.336 4.486c.091 1.076-.071 2.264-.904 2.95l-.102.085m-1.745 1.437L5.909 7.5H4.5L2.25 3.75l1.5-1.5L7.5 4.5v1.409l4.26 4.26m-1.745 1.437l1.745-1.437m6.615 8.206L15.75 15.75M4.867 19.125h.008v.008h-.008v-.008z" /></svg>`},
		{"Охрана строительных объектов", "Охрана строительных площадок и объектов недвижимости на всех этапах строительства.", `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 12l8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25" /></svg>`},
		{"Охрана магазинов", "Профессиональная охрана торговых объектов: предотвращение краж, контроль посетителей и сотрудников.", `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 3h1.386c.51 0 .955.343 1.087.835l.383 1.437M7.5 14.25a3 3 0 00-3 3h15.75m-12.75-3h11.218c1.121-2.3 2.1-4.684 2.924-7.138a60.114 60.114 0 00-16.536-1.84M7.5 14.25L5.106 5.272M6 20.25a.75.75 0 11-1.5 0 .75.75 0 011.5 0zm12.75 0a.75.75 0 11-1.5 0 .75.75 0 011.5 0z" /></svg>`},
		{"Охрана мероприятий", "Обеспечение безопасности на концертах, конференциях, корпоративах и спортивных мероприятиях.", `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72m.94 3.198l.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 016 18.719m12 0a5.971 5.971 0 00-.941-3.197m0 0A5.995 5.995 0 0012 12.75a5.995 5.995 0 00-5.058 2.772m0 0a3 3 0 00-4.681 2.72 8.986 8.986 0 003.74.477m.94-3.197a5.971 5.971 0 00-.94 3.197M15 6.75a3 3 0 11-6 0 3 3 0 016 0zm6 3a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0zm-13.5 0a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z" /></svg>`},
		{"КПП и пропускной режим", "Организация контрольно-пропускных пунктов, управление доступом персонала и посетителей.", `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M7.864 4.243A7.5 7.5 0 0119.5 10.5c0 2.92-.556 5.709-1.568 8.268M5.742 6.364A7.465 7.465 0 004.5 10.5a7.464 7.464 0 01-1.15 3.993m1.989 3.559A11.209 11.209 0 008.25 10.5a3.75 3.75 0 117.5 0c0 .527-.021 1.049-.064 1.565M12 10.5a14.94 14.94 0 01-3.6 9.75m6.633-4.596a18.666 18.666 0 01-2.485 5.33" /></svg>`},
		{"Сопровождение грузов", "Вооружённое и невооружённое сопровождение ценных грузов по Санкт-Петербургу и России.", `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M8.25 18.75a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 01-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 00-3.213-9.193 2.056 2.056 0 00-1.58-.86H14.25M16.5 18.75h-2.25m0-11.177v-.958c0-.568-.422-1.048-.987-1.106a48.554 48.554 0 00-10.026 0 1.106 1.106 0 00-.987 1.106v7.635m12-6.677v6.677m0 4.5v-4.5m0 0h-12" /></svg>`},
		{"Охрана коттеджей", "Охрана частных домов, коттеджей и загородных объектов с установкой систем видеонаблюдения.", `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 12l8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25" /></svg>`},
		{"Консультации по безопасности", "Профессиональный аудит и разработка индивидуальных решений по безопасности вашего объекта.", `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M8.625 9.75a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H8.25m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0H12m4.125 0a.375.375 0 11-.75 0 .375.375 0 01.75 0zm0 0h-.375m-13.5 3.01c0 1.6 1.123 2.994 2.707 3.227 1.087.16 2.185.283 3.293.369V21l4.184-4.183a1.14 1.14 0 01.778-.332 48.294 48.294 0 005.83-.498c1.585-.233 2.708-1.626 2.708-3.228V6.741c0-1.602-1.123-2.995-2.707-3.228A48.394 48.394 0 0012 3c-2.392 0-4.744.175-7.043.513C3.373 3.746 2.25 5.14 2.25 6.741v6.018z" /></svg>`},
	}

	for i, s := range services {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO services (name, description, icon, sort_order) VALUES (?, ?, ?, ?)`,
			s.name, s.desc, s.icon, i,
		); err != nil {
			return err
		}
	}

	// Seed SEO for the home page.
	if _, err := db.Exec(`INSERT OR IGNORE INTO seo_pages (slug, title, description, schema_json)
		VALUES ('/', ?, ?,
		'{"@context":"https://schema.org","@type":"LocalBusiness","name":"Альфа Юнит-1","telephone":"+7-931-362-56-88","address":{"@type":"PostalAddress","streetAddress":"ул. Лифляндская, д. 3","addressLocality":"Санкт-Петербург","postalCode":"190020","addressCountry":"RU"},"openingHours":"Mo-Fr 09:00-20:00","url":"https://alfaunit1.ru"}')`,
		"Альфа Юнит-1 — Охранная компания в Санкт-Петербурге | С 2002 года",
		"Лицензированная охранная компания в СПб. Вооружённая и невооружённая охрана объектов. Члены ассоциации ветеранов «Альфа». Звоните: +7 (931) 362-56-88",
	); err != nil {
		return err
	}

	log.Println("db: initial data seeded successfully")
	return nil
}
