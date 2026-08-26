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

		// ── Clients (Section 7 — Наши клиенты) ───────────────────────────────
		`CREATE TABLE IF NOT EXISTS clients (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			initial    TEXT    NOT NULL DEFAULT '',
			name       TEXT    NOT NULL,
			type_label TEXT    NOT NULL DEFAULT '',
			description TEXT   NOT NULL DEFAULT '',
			tags       TEXT    NOT NULL DEFAULT '',
			sort_order INTEGER NOT NULL DEFAULT 0,
			active     INTEGER NOT NULL DEFAULT 1
		)`,

		// ── Team members (Section 8 — Команда) ───────────────────────────────
		`CREATE TABLE IF NOT EXISTS team_members (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			initial    TEXT    NOT NULL DEFAULT '',
			name       TEXT    NOT NULL,
			role       TEXT    NOT NULL DEFAULT '',
			department TEXT    NOT NULL DEFAULT '',
			description TEXT   NOT NULL DEFAULT '',
			tags       TEXT    NOT NULL DEFAULT '',
			sort_order INTEGER NOT NULL DEFAULT 0,
			active     INTEGER NOT NULL DEFAULT 1
		)`,

		// ── Process steps (Section 9 — Как мы работаем) ──────────────────────
		`CREATE TABLE IF NOT EXISTS process_steps (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			step_num    TEXT    NOT NULL DEFAULT '',
			title       TEXT    NOT NULL,
			description TEXT    NOT NULL DEFAULT '',
			sort_order  INTEGER NOT NULL DEFAULT 0,
			active      INTEGER NOT NULL DEFAULT 1
		)`,

		// ── FAQ (Section 10) ──────────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS faq_items (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			question   TEXT    NOT NULL,
			answer     TEXT    NOT NULL DEFAULT '',
			sort_order INTEGER NOT NULL DEFAULT 0,
			active     INTEGER NOT NULL DEFAULT 1
		)`,

		// ── Advantages (ПОЧЕМУ ВЫБИРАЮТ НАС) ─────────────────────────────────
		`CREATE TABLE IF NOT EXISTS advantages (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			title       TEXT    NOT NULL,
			description TEXT    NOT NULL DEFAULT '',
			sort_order  INTEGER NOT NULL DEFAULT 0,
			active      INTEGER NOT NULL DEFAULT 1
		)`,

		// ── History events (ИСТОРИЯ КОМПАНИИ) ────────────────────────────────
		`CREATE TABLE IF NOT EXISTS history_events (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			year_label  TEXT    NOT NULL,
			title       TEXT    NOT NULL,
			description TEXT    NOT NULL DEFAULT '',
			quote       TEXT    NOT NULL DEFAULT '',
			sort_order  INTEGER NOT NULL DEFAULT 0,
			active      INTEGER NOT NULL DEFAULT 1
		)`,

		// ── Licenses (ЛИЦЕНЗИИ И ДОКУМЕНТЫ) ──────────────────────────────────
		`CREATE TABLE IF NOT EXISTS licenses (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			type_label  TEXT    NOT NULL DEFAULT '',
			company     TEXT    NOT NULL,
			description TEXT    NOT NULL DEFAULT '',
			status_text TEXT    NOT NULL DEFAULT 'Действующая лицензия',
			sort_order  INTEGER NOT NULL DEFAULT 0,
			active      INTEGER NOT NULL DEFAULT 1
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
		// Run partial seeds for new tables added after initial deployment.
		return seedNewTables(db)
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
		"about_text":       "Группа компаний включает ЧОО «Альфа Юнит-1» и ЧОО «Альфа Безопасность» (Санкт-Петербург) — входим в Международную Ассоциацию ветеранов подразделения антитеррора «Альфа». Основа нашей идеологии: профессионализм, надёжность и особый уровень доверия. Работаем по Закону РФ от 11.03.1992 г. №2487-1.",
		"stats_years":      "23",
		"stats_objects":    "50+",
		"stats_staff":      "200+",
		"stats_licenses":   "2",
		"whatsapp_link":    "",
		"telegram_link":    "",
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
	if err := seedServices(db); err != nil {
		return err
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

	// Seed new content tables.
	if err := seedNewTables(db); err != nil {
		return err
	}

	log.Println("db: initial data seeded successfully")
	return nil
}

// seedNewTables seeds tables added after initial deployment (idempotent via INSERT OR IGNORE).
func seedNewTables(db *sql.DB) error {
	if err := seedClients(db); err != nil {
		return err
	}
	if err := seedTeam(db); err != nil {
		return err
	}
	if err := seedProcess(db); err != nil {
		return err
	}
	if err := seedFAQ(db); err != nil {
		return err
	}
	if err := seedAdvantages(db); err != nil {
		return err
	}
	if err := seedHistory(db); err != nil {
		return err
	}
	if err := seedLicenses(db); err != nil {
		return err
	}
	// Ensure new settings keys exist.
	for _, pair := range [][2]string{
		{"whatsapp_link", ""},
		{"telegram_link", ""},
	} {
		if _, err := db.Exec(`INSERT OR IGNORE INTO settings (key, value) VALUES (?, ?)`, pair[0], pair[1]); err != nil {
			return err
		}
	}
	return nil
}

func seedServices(db *sql.DB) error {
	services := []struct{ name, desc, icon string }{
		{
			"Вооружённая охрана",
			"Стационарные посты с вооружёнными сотрудниками для объектов повышенного уровня безопасности: склады, производства, ТРК.",
			`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.955 11.955 0 01.42 12c0 2.01.5 3.903 1.378 5.56A11.956 11.956 0 013.6 18M12 2.764A11.959 11.959 0 0120.402 6 11.955 11.955 0 0123.58 12a11.955 11.955 0 01-3.177 5.56A11.956 11.956 0 0112 21.236" /></svg>`,
		},
		{
			"Невооружённая охрана",
			"Профессиональные мобильные посты без огнестрельного оружия для офисов, магазинов, бизнес-центров и общественных мест.",
			`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" /></svg>`,
		},
		{
			"Охрана ТРК и бизнес-центров",
			"Многоуровневая система охраны торгово-развлекательных комплексов и бизнес-центров: КПП, видеонаблюдение, контроль арендаторов.",
			`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3.75h.008v.008h-.008v-.008zm0 3h.008v.008h-.008v-.008zm0 3h.008v.008h-.008v-.008z" /></svg>`,
		},
		{
			"Охрана логистических центров и складов",
			"Круглосуточная охрана складских и логистических комплексов с организацией пропускного режима транспорта и персонала.",
			`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M20.25 7.5l-.625 10.632a2.25 2.25 0 01-2.247 2.118H6.622a2.25 2.25 0 01-2.247-2.118L3.75 7.5M10 11.25h4M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z" /></svg>`,
		},
		{
			"Охрана производственных объектов",
			"Защита заводов и производственных предприятий: охрана периметра, контроль въезда/выезда, сохранность оборудования и сырья.",
			`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M11.42 15.17L17.25 21A2.652 2.652 0 0021 17.25l-5.877-5.877M11.42 15.17l2.496-3.03c.317-.384.74-.626 1.208-.766M11.42 15.17l-4.655 5.653a2.548 2.548 0 11-3.586-3.586l6.837-5.63m5.108-.233c.55-.164 1.163-.188 1.743-.14a4.5 4.5 0 004.486-6.336l-3.276 3.277a3.004 3.004 0 01-2.25-2.25l3.276-3.276a4.5 4.5 0 00-6.336 4.486c.091 1.076-.071 2.264-.904 2.95l-.102.085m-1.745 1.437L5.909 7.5H4.5L2.25 3.75l1.5-1.5L7.5 4.5v1.409l4.26 4.26m-1.745 1.437l1.745-1.437m6.615 8.206L15.75 15.75M4.867 19.125h.008v.008h-.008v-.008z" /></svg>`,
		},
		{
			"Охрана строительных площадок",
			"Охрана объектов строительства на всех этапах: защита материалов, оборудования, контроль персонала подрядчиков.",
			`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 12l8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25" /></svg>`,
		},
		{
			"Личная охрана и сопровождение VIP",
			"Профессиональная личная охрана первых лиц компаний, сопровождение на деловых встречах и в поездках по СПб и СЗФО.",
			`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M17.982 18.725A7.488 7.488 0 0012 15.75a7.488 7.488 0 00-5.982 2.975m11.963 0a9 9 0 10-11.963 0m11.963 0A8.966 8.966 0 0112 21a8.966 8.966 0 01-5.982-2.275M15 9.75a3 3 0 11-6 0 3 3 0 016 0z" /></svg>`,
		},
		{
			"Охрана массовых мероприятий",
			"Обеспечение безопасности на концертах, корпоративах, конференциях, спортивных и культурных мероприятиях.",
			`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M18 18.72a9.094 9.094 0 003.741-.479 3 3 0 00-4.682-2.72m.94 3.198l.001.031c0 .225-.012.447-.037.666A11.944 11.944 0 0112 21c-2.17 0-4.207-.576-5.963-1.584A6.062 6.062 0 016 18.719m12 0a5.971 5.971 0 00-.941-3.197m0 0A5.995 5.995 0 0012 12.75a5.995 5.995 0 00-5.058 2.772m0 0a3 3 0 00-4.681 2.72 8.986 8.986 0 003.74.477m.94-3.197a5.971 5.971 0 00-.94 3.197M15 6.75a3 3 0 11-6 0 3 3 0 016 0zm6 3a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0zm-13.5 0a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z" /></svg>`,
		},
		{
			"Охрана ТСЖ и коттеджей",
			"Охрана жилых объектов: товариществ собственников жилья, коттеджных посёлков, частных домов и резиденций.",
			`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 12l8.954-8.955c.44-.439 1.152-.439 1.591 0L21.75 12M4.5 9.75v10.125c0 .621.504 1.125 1.125 1.125H9.75v-4.875c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21h4.125c.621 0 1.125-.504 1.125-1.125V9.75M8.25 21h8.25" /></svg>`,
		},
		{
			"КПП и пропускной режим",
			"Организация контрольно-пропускных пунктов: управление доступом сотрудников, посетителей и транспортных средств.",
			`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M7.864 4.243A7.5 7.5 0 0119.5 10.5c0 2.92-.556 5.709-1.568 8.268M5.742 6.364A7.465 7.465 0 004.5 10.5a7.464 7.464 0 01-1.15 3.993m1.989 3.559A11.209 11.209 0 008.25 10.5a3.75 3.75 0 117.5 0c0 .527-.021 1.049-.064 1.565M12 10.5a14.94 14.94 0 01-3.6 9.75m6.633-4.596a18.666 18.666 0 01-2.485 5.33" /></svg>`,
		},
		{
			"Вооружённое сопровождение грузов",
			"Сопровождение ценных и опасных грузов вооружёнными охранниками на автомобиле сопровождения по СПб и России.",
			`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M8.25 18.75a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 01-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 01-3 0m3 0a1.5 1.5 0 00-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 00-3.213-9.193 2.056 2.056 0 00-1.58-.86H14.25M16.5 18.75h-2.25m0-11.177v-.958c0-.568-.422-1.048-.987-1.106a48.554 48.554 0 00-10.026 0 1.106 1.106 0 00-.987 1.106v7.635m12-6.677v6.677m0 4.5v-4.5m0 0h-12" /></svg>`,
		},
		{
			"Невооружённое сопровождение грузов",
			"Сопровождение грузов и транспортных средств без огнестрельного оружия — для стандартных коммерческих перевозок.",
			`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3.75v4.5m0-4.5h4.5m-4.5 0L9 9M3.75 20.25v-4.5m0 4.5h4.5m-4.5 0L9 15M20.25 3.75h-4.5m4.5 0v4.5m0-4.5L15 9m5.25 11.25h-4.5m4.5 0v-4.5m0 4.5L15 15" /></svg>`,
		},
		{
			"Охранная сигнализация",
			"Проектирование и монтаж стандартных охранных систем сигнализации для офисов, складов и производственных помещений.",
			`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0" /></svg>`,
		},
		{
			"Тревожная и пультовая сигнализация",
			"Установка тревожных кнопок с выводом сигнала на пульт централизованной охраны. Быстрое реагирование группы.",
			`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" /></svg>`,
		},
		{
			"Комбинированная сигнализация",
			"Интегрированные системы охранной и тревожной сигнализации с видеонаблюдением — комплексная защита объекта.",
			`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25m18 0A2.25 2.25 0 0018.75 3H5.25A2.25 2.25 0 003 5.25m18 0H3" /></svg>`,
		},
	}
	for i, s := range services {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO services (name, description, icon, sort_order) VALUES (?, ?, ?, ?)`,
			s.name, s.desc, s.icon, i,
		); err != nil {
			return err
		}
	}
	return nil
}

func seedClients(db *sql.DB) error {
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM clients`).Scan(&n)
	if n > 0 {
		return nil
	}
	clients := []struct{ initial, name, typeLabel, description, tags string }{
		{"С", "«Советская Звезда»", "Коммерческая недвижимость", "Охрана объектов коммерческой недвижимости в Адмиралтейском районе Санкт-Петербурга. Контроль доступа арендаторов и посетителей.", "Бизнес-центр,КПП"},
		{"Т", "АО «Трест № 68»", "Строительное предприятие", "Охрана строительного предприятия, включающего 4 производственных подразделения. Круглосуточный контроль периметра и въезда техники.", "Строительство,Периметр,24/7"},
		{"В", "«ВК Сервис»", "Инфраструктурный объект", "Охрана объектов компании, специализирующейся на очистке водоотведения и диагностике трубопроводов. Охрана оборудования и складов.", "Инфраструктура,Склад"},
		{"К", "Гостиница «Северная Корона»", "Строительная площадка", "Охрана строительной площадки гостиницы в Санкт-Петербурге. Контроль материалов, техники и персонала на всех этапах строительства.", "Стройплощадка,Материалы"},
		{"Х", "«Технология холода»", "Холодильные склады", "Охрана холодильных складских комплексов и арендуемых помещений. Контроль въезда транспорта, сохранность продукции.", "Склад,Холодильник,Въезд/выезд"},
		{"П", "Ведомственная парковка", "Тележная ул., д. 32", "Охрана ведомственной парковки: контроль въезда/выезда транспортных средств, видеонаблюдение, пропускной режим.", "Парковка,Видеонаблюдение"},
		{"А", "ГК «Алгоритм»", "Демонтажные работы", "Охрана объектов группы компаний, специализирующейся на демонтажных работах. Обеспечение безопасности строительных площадок.", "Демонтаж,Строительство"},
		{"R", "«Ренейссанс Констракшн»", "Международная строительная компания", "Охрана объектов международной строительной компании в Санкт-Петербурге. Многоуровневая система безопасности крупных стройплощадок.", "Стройплощадка,Международный"},
		{"М", "«Медиэстетик»", "Клиника эстетической медицины", "Охрана клиники эстетической медицины: обеспечение безопасности пациентов и персонала, контроль доступа в медицинское учреждение.", "Медицина,Клиника"},
		{"Э", "ООО «ЭнергоСвязьСтрой»", "Электросетевая инфраструктура", "Охрана объектов компании, проектирующей электросетевую инфраструктуру от 0,4 до 750 кВ. Защита проектной документации и оборудования.", "Энергетика,Офис,Оборудование"},
		{"В", "«Венчурный Капитал»", "Инвестиционная компания", "Физическая охрана офиса инвестиционной компании, обеспечение конфиденциальности переговоров и защита ценных активов.", "Офис,Конфиденциальность"},
	}
	for i, c := range clients {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO clients (initial, name, type_label, description, tags, sort_order) VALUES (?, ?, ?, ?, ?, ?)`,
			c.initial, c.name, c.typeLabel, c.description, c.tags, i,
		); err != nil {
			return err
		}
	}
	return nil
}

func seedTeam(db *sql.DB) error {
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM team_members`).Scan(&n)
	if n > 0 {
		return nil
	}
	members := []struct{ initial, name, role, department, description, tags string }{
		{"А", "Генеральный директор", "Генеральный директор", "ЧОО «Альфа Юнит-1»", "Ветеран силовых структур. Руководит группой компаний с момента основания в 2002 году. Свыше 25 лет опыта в сфере безопасности.", "Стратегия,Управление,Ветеран «Альфа»"},
		{"О", "Начальник охраны", "Начальник охраны", "Оперативный отдел", "Ветеран СОБР. Организация охранных мероприятий, тактическая подготовка личного состава. 20 лет боевого и оперативного опыта.", "Тактика,СОБР,Подготовка"},
		{"К", "Руководитель HR", "Руководитель HR", "Кадровая служба", "Отбор и аттестация сотрудников. Организация медицинских комиссий, квалификационных экзаменов в Росгвардии, физподготовки.", "Кадры,Аттестация"},
	}
	for i, m := range members {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO team_members (initial, name, role, department, description, tags, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			m.initial, m.name, m.role, m.department, m.description, m.tags, i,
		); err != nil {
			return err
		}
	}
	return nil
}

func seedProcess(db *sql.DB) error {
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM process_steps`).Scan(&n)
	if n > 0 {
		return nil
	}
	steps := []struct{ num, title, desc string }{
		{"01", "Заявка", "Позвоните по телефону или заполните форму — специалист свяжется с вами в течение 30 минут в рабочее время."},
		{"02", "Анализ объекта", "Выезд специалиста на объект. Оценка рисков, уязвимостей, составление технического задания на охрану."},
		{"03", "Подготовка решения", "Разработка персонального плана охраны: численность, режим работы, оборудование. Коммерческое предложение."},
		{"04", "Заключение договора", "Подписание договора на оказание охранных услуг. Все условия прозрачны, скрытых платежей нет."},
		{"05", "Организация охраны", "Расстановка сотрудников, монтаж сигнализации и оборудования, инструктаж персонала. Начало охраны объекта."},
		{"06", "Контроль качества", "Постоянный мониторинг работы охраны, ежемесячные отчёты клиенту, оперативное реагирование на замечания."},
	}
	for i, s := range steps {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO process_steps (step_num, title, description, sort_order) VALUES (?, ?, ?, ?)`,
			s.num, s.title, s.desc, i,
		); err != nil {
			return err
		}
	}
	return nil
}

func seedFAQ(db *sql.DB) error {
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM faq_items`).Scan(&n)
	if n > 0 {
		return nil
	}
	items := []struct{ q, a string }{
		{"Сколько стоят услуги охраны?", "Стоимость зависит от типа объекта, количества постов, вооружённости охраны и режима работы. Позвоните нам или оставьте заявку — подготовим коммерческое предложение в течение 24 часов после осмотра объекта."},
		{"Есть ли у компании лицензия на вооружённую охрану?", "Да. Обе компании группы — ЧОО «Альфа Юнит-1» и ЧОО «Альфа Безопасность» — имеют все необходимые лицензии, включая вооружённую охрану, выданные в соответствии с Законом РФ от 11.03.1992 г. №2487-1 «О частной детективной и охранной деятельности»."},
		{"Как быстро можно приступить к охране объекта?", "В стандартных случаях охрана объекта организуется в течение 1–3 рабочих дней после подписания договора. При срочной необходимости возможен выход охранников в течение 24 часов."},
		{"В каких регионах вы работаете?", "Основной регион — Санкт-Петербург и Ленинградская область. Также работаем по всему Северо-Западному федеральному округу. Дополнительный офис расположен в Симферополе (ул. Карла Маркса, 14)."},
		{"Что входит в охранную сигнализацию?", "Мы устанавливаем охранную сигнализацию трёх типов: стандартная охранная, тревожная/пультовая (кнопка вызова группы реагирования) и комбинированная. Стоимость и тип подбираются индивидуально под объект."},
		{"Как трудоустроены сотрудники охраны?", "Все сотрудники официально трудоустроены, имеют удостоверение частного охранника и прошли квалификационный экзамен в Росгвардии. Обязательно: медицинская комиссия, проверка биографии, физическая подготовка и психологическое тестирование."},
	}
	for i, f := range items {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO faq_items (question, answer, sort_order) VALUES (?, ?, ?)`,
			f.q, f.a, i,
		); err != nil {
			return err
		}
	}
	return nil
}

func seedAdvantages(db *sql.DB) error {
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM advantages`).Scan(&n)
	if n > 0 {
		return nil
	}
	items := []struct{ title, desc string }{
		{"Лицензированная деятельность", "Имеем все необходимые лицензии на осуществление частной охранной деятельности для обоих юридических лиц группы."},
		{"Ветераны подразделения «Альфа»", "Костяк компании — профессионалы с боевым опытом, члены Международной ассоциации ветеранов «АЛЬФА»."},
		{"Современные средства охраны", "Используем передовые технологии: системы видеонаблюдения, тревожные кнопки, GPS-мониторинг транспорта."},
		{"Индивидуальный подход", "Разрабатываем план охраны конкретно под ваш объект: анализируем риски, предлагаем оптимальное решение."},
		{"Полная конфиденциальность", "Строго соблюдаем режим коммерческой тайны. Информация об объектах и клиентах не разглашается третьим лицам."},
		{"Круглосуточная поддержка", "Оперативный центр работает 24/7. Время реагирования на тревожный сигнал — минимальное по СПб."},
	}
	for i, a := range items {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO advantages (title, description, sort_order) VALUES (?, ?, ?)`,
			a.title, a.desc, i,
		); err != nil {
			return err
		}
	}
	return nil
}

func seedHistory(db *sql.DB) error {
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM history_events`).Scan(&n)
	if n > 0 {
		return nil
	}
	events := []struct{ year, title, desc, quote string }{
		{"1980", "Олимпийская группа захвата", "Создана специальная группа захвата для обеспечения безопасности в период проведения Олимпийских игр в Москве. Одно из самых подготовленных и профессиональных подразделений МВД СССР — образец для всех последующих структур компании.", "Десятки блестяще проведённых операций — задержание вооружённых и особо опасных преступников живыми, без применения оружия"},
		{"1985–1995", "Рота оперативного реагирования (РОР)", "Преемник олимпийской группы — Рота оперативного реагирования. Специализация: задержание вооружённых и особо опасных преступников. Десятки успешных операций по всему Северо-Западу.", ""},
		{"1990-е", "Формирование Резерва и СОБР", "На базе РОР сформирован Резерв — подразделение для выполнения задач наивысшей сложности. Впоследствии преобразован в самостоятельный Специальный отряд быстрого реагирования (СОБР).", ""},
		{"2002", "Основание ЧОО «Альфа Юнит-1»", "Ветераны элитных силовых структур создали частное охранное предприятие в Санкт-Петербурге. С первых дней компания объединила высочайший профессионализм силовых структур с гибкостью коммерческой организации.", ""},
		{"Сегодня", "Группа компаний — лидер СЗФО", "Два юридических лица (ЧОО «Альфа Юнит-1» и ЧОО «Альфа Безопасность»), более 50+ объектов под охраной, 200+ сотрудников. Офисы в Санкт-Петербурге и Симферополе.", ""},
	}
	for i, e := range events {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO history_events (year_label, title, description, quote, sort_order) VALUES (?, ?, ?, ?, ?)`,
			e.year, e.title, e.desc, e.quote, i,
		); err != nil {
			return err
		}
	}
	return nil
}

func seedLicenses(db *sql.DB) error {
	var n int
	_ = db.QueryRow(`SELECT COUNT(*) FROM licenses`).Scan(&n)
	if n > 0 {
		return nil
	}
	items := []struct{ typeLabel, company, desc, status string }{
		{"Лицензия ЧОО", "«Альфа Юнит-1»", "Лицензия на осуществление частной охранной деятельности ЧОО «Альфа Юнит-1». Санкт-Петербург и СЗФО.", "Действующая лицензия"},
		{"Лицензия ЧОО", "«Альфа Безопасность»", "Лицензия на осуществление частной охранной деятельности ЧОО «Альфа Безопасность». Санкт-Петербург и СЗФО.", "Действующая лицензия"},
		{"Свидетельство", "Ассоциация «Альфа»", "Свидетельство о членстве в Международной ассоциации ветеранов подразделения «АЛЬФА» — знак высокого профессионализма.", "Действующее свидетельство"},
	}
	for i, l := range items {
		if _, err := db.Exec(
			`INSERT OR IGNORE INTO licenses (type_label, company, description, status_text, sort_order) VALUES (?, ?, ?, ?, ?)`,
			l.typeLabel, l.company, l.desc, l.status, i,
		); err != nil {
			return err
		}
	}
	return nil
}
