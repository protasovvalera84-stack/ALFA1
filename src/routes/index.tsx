import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { siteMeta } from "../lib/site-meta";

export const Route = createFileRoute("/")({
  component: HomePage,
});

// ── Shared data ────────────────────────────────────────────────────────────
const services = [
  {
    name: "Вооружённая охрана",
    desc: "Стационарные и мобильные посты с вооружёнными сотрудниками для объектов повышенного уровня безопасности.",
    icon: "🛡️",
  },
  {
    name: "Невооружённая охрана",
    desc: "Профессиональная охрана без огнестрельного оружия для офисов, магазинов и общественных мест.",
    icon: "👮",
  },
  {
    name: "Охрана бизнес-центров",
    desc: "Организация многоуровневой системы охраны деловых центров, контроль доступа и безопасности арендаторов.",
    icon: "🏢",
  },
  {
    name: "Охрана складов",
    desc: "Круглосуточная охрана складских комплексов с организацией пропускного режима и видеонаблюдения.",
    icon: "📦",
  },
  {
    name: "Охрана производств",
    desc: "Защита производственных объектов, промышленных предприятий и заводов с контролем въезда/выезда.",
    icon: "🏭",
  },
  {
    name: "Охрана строительных объектов",
    desc: "Охрана строительных площадок и объектов недвижимости на всех этапах строительства.",
    icon: "🏗️",
  },
  {
    name: "Охрана магазинов",
    desc: "Профессиональная охрана торговых объектов: предотвращение краж, контроль посетителей.",
    icon: "🛒",
  },
  {
    name: "Охрана мероприятий",
    desc: "Обеспечение безопасности на концертах, конференциях, корпоративах и спортивных мероприятиях.",
    icon: "🎪",
  },
  {
    name: "КПП и пропускной режим",
    desc: "Организация контрольно-пропускных пунктов, управление доступом персонала и посетителей.",
    icon: "🔑",
  },
  {
    name: "Сопровождение грузов",
    desc: "Вооружённое и невооружённое сопровождение ценных грузов по Санкт-Петербургу и России.",
    icon: "🚛",
  },
  {
    name: "Охрана коттеджей",
    desc: "Охрана частных домов, коттеджей и загородных объектов с установкой систем видеонаблюдения.",
    icon: "🏡",
  },
  {
    name: "Консультации по безопасности",
    desc: "Профессиональный аудит и разработка индивидуальных решений по безопасности вашего объекта.",
    icon: "💬",
  },
];

const faqItems = [
  {
    q: "Сколько стоят услуги охраны?",
    a: "Стоимость зависит от типа объекта, количества постов, вооружённости охраны и режима работы. Позвоните нам — подготовим коммерческое предложение в течение 24 часов.",
  },
  {
    q: "Есть ли у компании лицензия на вооружённую охрану?",
    a: "Да. Группа компаний «Альфа Юнит-1» имеет все необходимые лицензии, включая вооружённую охрану, в соответствии с Законом РФ №2487-1.",
  },
  {
    q: "Как быстро вы можете приступить к охране?",
    a: "В стандартных случаях — 1–3 рабочих дня после подписания договора. При срочной необходимости — в течение 24 часов.",
  },
  {
    q: "В каких регионах вы работаете?",
    a: "Основной регион — Санкт-Петербург и Ленинградская область. Также работаем по всему Северо-Западному федеральному округу.",
  },
  {
    q: "Можно ли заказать охрану на разовое мероприятие?",
    a: "Да, мы охраняем разовые мероприятия: корпоративы, конференции, концерты, спортивные события.",
  },
  {
    q: "Как устроены сотрудники охраны?",
    a: "Все официально трудоустроены, имеют удостоверение частного охранника, прошли квалификационный экзамен в Росгвардии, медкомиссию и психологическое тестирование.",
  },
];

// ── Main component ─────────────────────────────────────────────────────────
function HomePage() {
  const [menuOpen, setMenuOpen] = useState(false);
  const [scrolled, setScrolled] = useState(false);
  const [openFaq, setOpenFaq] = useState<number | null>(null);
  const [formState, setFormState] = useState<"idle" | "loading" | "success" | "error">("idle");
  const [formMsg, setFormMsg] = useState("");
  const formRef = useRef<HTMLFormElement>(null);

  // Navbar scroll + progress bar
  useEffect(() => {
    const bar = document.getElementById("scroll-progress");
    const onScroll = () => {
      setScrolled(window.scrollY > 40);
      if (bar) {
        const pct = (window.scrollY / (document.body.scrollHeight - window.innerHeight)) * 100;
        bar.style.width = Math.min(pct, 100) + "%";
      }
    };
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  async function handleContactSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!formRef.current) return;
    const fd = new FormData(formRef.current);
    const payload = {
      name: (fd.get("name") as string) || "",
      phone: (fd.get("phone") as string) || "",
      email: (fd.get("email") as string) || "",
      message: (fd.get("message") as string) || "",
      service: (fd.get("service") as string) || "",
    };
    if (!payload.name || !payload.phone) {
      setFormState("error");
      setFormMsg("Укажите имя и телефон.");
      return;
    }
    setFormState("loading");
    try {
      const res = await fetch("/contact", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      const json = (await res.json()) as { message?: string; error?: string };
      if (!res.ok) throw new Error(json.error || `HTTP ${res.status}`);
      setFormState("success");
      setFormMsg(json.message || "Заявка принята! Мы свяжемся с вами.");
      formRef.current.reset();
    } catch (err: unknown) {
      setFormState("error");
      setFormMsg(err instanceof Error ? err.message : "Ошибка. Позвоните нам напрямую.");
    }
  }

  return (
    <div className="min-h-screen" style={{ fontFamily: "Inter, sans-serif" }}>
      {/* ── NAVIGATION ── */}
      <nav
        className={`fixed top-0 left-0 right-0 z-50 transition-all duration-500 ${
          scrolled ? "bg-[#05070A]/92 backdrop-blur-xl border-b border-white/6 shadow-2xl" : ""
        }`}
      >
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16 lg:h-20">
            {/* Logo */}
            <a href="/" className="flex items-center gap-3 group">
              <div className="w-10 h-10 rounded-lg bg-[#D4AF37] flex items-center justify-center shrink-0 group-hover:bg-[#E8C84A] transition-colors">
                <svg width="22" height="22" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                  <path
                    d="M12 2L3 7V12C3 17.25 6.96 22.18 12 23.5C17.04 22.18 21 17.25 21 12V7L12 2Z"
                    fill="#05070A"
                  />
                  <path
                    d="M9 12L11 14L15 10"
                    stroke="#D4AF37"
                    strokeWidth="1.5"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
              </div>
              <div className="leading-none">
                <div
                  className="text-xl text-white tracking-wider"
                  style={{ fontFamily: "Bebas Neue, sans-serif" }}
                >
                  АЛЬФА ЮНИТ-1
                </div>
                <div className="text-xs text-white/40 tracking-widest uppercase">
                  Охранная компания
                </div>
              </div>
            </a>

            {/* Desktop links */}
            <div className="hidden lg:flex items-center gap-8">
              {["#about", "#services", "#history", "#contacts"].map((href, i) => (
                <a
                  key={href}
                  href={href}
                  className="text-sm text-white/70 hover:text-[#D4AF37] transition-colors tracking-wide uppercase"
                >
                  {["О компании", "Услуги", "История", "Контакты"][i]}
                </a>
              ))}
            </div>

            {/* CTA */}
            <div className="flex items-center gap-4">
              <a
                href={`tel:${siteMeta.phone1}`}
                className="hidden sm:flex items-center gap-2 text-[#D4AF37] text-sm font-medium hover:text-[#E8C84A] transition-colors"
              >
                📞 {siteMeta.phone1}
              </a>
              <a
                href="#contacts"
                className="hidden md:block text-sm px-5 py-2 rounded-lg font-semibold bg-[#D4AF37] text-[#05070A] hover:bg-[#E8C84A] transition-colors"
              >
                Консультация
              </a>
              <button
                onClick={() => setMenuOpen(!menuOpen)}
                className="lg:hidden w-10 h-10 flex flex-col justify-center items-center gap-1.5"
                aria-label="Меню"
                aria-expanded={menuOpen}
              >
                {[0, 1, 2].map((i) => (
                  <span
                    key={i}
                    className={`block h-0.5 bg-white transition-all duration-300 ${
                      menuOpen
                        ? i === 0
                          ? "w-6 translate-y-2 rotate-45"
                          : i === 1
                            ? "w-6 opacity-0"
                            : "w-6 -translate-y-2 -rotate-45"
                        : i === 2
                          ? "w-4"
                          : "w-6"
                    }`}
                  />
                ))}
              </button>
            </div>
          </div>
        </div>

        {/* Mobile menu */}
        {menuOpen && (
          <div className="lg:hidden glass border-t border-white/10 px-4 py-6 flex flex-col gap-4">
            {["#about", "#services", "#history", "#contacts"].map((href, i) => (
              <a
                key={href}
                href={href}
                onClick={() => setMenuOpen(false)}
                className="text-white/80 hover:text-[#D4AF37] py-2 transition-colors"
              >
                {["О компании", "Услуги", "История", "Контакты"][i]}
              </a>
            ))}
            <a href={`tel:${siteMeta.phone1}`} className="text-[#D4AF37] font-semibold py-2">
              {siteMeta.phone1}
            </a>
            <a
              href="#contacts"
              onClick={() => setMenuOpen(false)}
              className="bg-[#D4AF37] text-[#05070A] text-center py-3 rounded-lg font-bold"
            >
              Получить консультацию
            </a>
          </div>
        )}
      </nav>

      {/* ── HERO ── */}
      <section
        id="hero"
        className="relative min-h-screen flex items-center justify-center overflow-hidden"
      >
        <div className="absolute inset-0 bg-gradient-to-br from-[#05070A] via-[#080d14] to-[#0B111B]" />
        <div className="hero-grid absolute inset-0 opacity-20" />
        <div className="orb-gold absolute top-1/4 left-1/4 w-96 h-96 bg-[#D4AF37]/5 rounded-full blur-3xl" />
        <div className="orb-blue absolute bottom-1/4 right-1/4 w-80 h-80 bg-[#2D5FFF]/8 rounded-full blur-3xl" />

        <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-24 pb-16">
          <div className="max-w-4xl">
            <div className="inline-flex items-center gap-2 glass border border-[#D4AF37]/30 rounded-full px-4 py-1.5 mb-8">
              <span className="w-2 h-2 rounded-full bg-[#D4AF37] animate-pulse" />
              <span className="text-[#D4AF37] text-xs font-semibold tracking-[.2em] uppercase">
                Ассоциация ветеранов «Альфа» · С 2002 года
              </span>
            </div>

            <h1
              className="text-5xl sm:text-6xl lg:text-7xl xl:text-8xl leading-none tracking-wide mb-6 text-white"
              style={{ fontFamily: "Bebas Neue, sans-serif" }}
            >
              Комплексная безопасность объектов любой сложности
            </h1>

            <p className="text-lg sm:text-xl text-white/60 max-w-2xl mb-10 leading-relaxed">
              Вооружённая и невооружённая охрана. Санкт-Петербург и Северо-Запад России.
            </p>

            <div className="flex flex-col sm:flex-row gap-4 mb-16">
              <a
                href="#contacts"
                className="inline-flex items-center justify-center gap-2 px-8 py-4 rounded-xl font-semibold text-base bg-[#D4AF37] text-[#05070A] hover:bg-[#E8C84A] hover:-translate-y-0.5 transition-all shadow-[0_8px_24px_rgba(212,175,55,.35)]"
              >
                📞 Получить консультацию
              </a>
              <a
                href="#services"
                className="inline-flex items-center justify-center gap-2 px-8 py-4 rounded-xl font-semibold text-base glass border border-white/20 hover:border-white/40 hover:-translate-y-0.5 transition-all"
              >
                Наши услуги ↓
              </a>
            </div>

            <div className="flex flex-wrap gap-6">
              {["Лицензированная деятельность", "Работаем с 2002 года", "СПб и Северо-Запад"].map(
                (text) => (
                  <div key={text} className="flex items-center gap-2">
                    <span className="text-[#D4AF37]">✓</span>
                    <span className="text-sm text-white/60">{text}</span>
                  </div>
                ),
              )}
            </div>
          </div>
        </div>

        <div className="absolute bottom-8 left-1/2 -translate-x-1/2 flex flex-col items-center gap-2 bounce-slow">
          <span className="text-white/30 text-xs tracking-widest uppercase">Scroll</span>
          <div className="w-px h-12 bg-gradient-to-b from-white/30 to-transparent" />
        </div>
      </section>

      {/* ── ABOUT ── */}
      <section
        id="about"
        className="py-24 lg:py-32 relative overflow-hidden"
        style={{ backgroundColor: "#0B111B" }}
      >
        <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-[#D4AF37]/30 to-transparent" />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <SectionLabel>О нас</SectionLabel>
            <h2
              className="text-4xl sm:text-5xl lg:text-6xl tracking-wide mb-6"
              style={{ fontFamily: "Bebas Neue, sans-serif" }}
            >
              О КОМПАНИИ
            </h2>
            <p className="text-white/60 text-lg max-w-3xl mx-auto leading-relaxed">
              Группа компаний «Альфа Юнит-1» — лицензированная охранная организация, действующая в
              Санкт-Петербурге с 2002 года. Члены Международной ассоциации ветеранов подразделения
              «АЛЬФА».
            </p>
          </div>

          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 lg:gap-6 mb-20">
            {[
              { value: "23", label: "года на рынке" },
              { value: "50+", label: "объектов под охраной" },
              { value: "200+", label: "сотрудников" },
              { value: "2", label: "лицензии ЧОО" },
            ].map((stat) => (
              <div
                key={stat.label}
                className="glass border border-white/10 rounded-2xl p-6 lg:p-8 text-center hover:border-[#D4AF37]/40 transition-all duration-300"
              >
                <div
                  className="text-4xl lg:text-5xl text-[#D4AF37] mb-2"
                  style={{ fontFamily: "Bebas Neue, sans-serif" }}
                >
                  {stat.value}
                </div>
                <div className="text-white/50 text-sm uppercase tracking-wider">{stat.label}</div>
              </div>
            ))}
          </div>

          <div className="grid md:grid-cols-3 gap-6">
            {[
              {
                icon: "🏅",
                title: "Профессионализм",
                desc: "Все сотрудники проходят медицинскую комиссию, квалификационные экзамены и регулярные тренировки.",
              },
              {
                icon: "🔒",
                title: "Надёжность",
                desc: "Члены Международной ассоциации ветеранов подразделения «АЛЬФА» — гарантия высочайшего доверия.",
              },
              {
                icon: "👤",
                title: "Индивидуальный подход",
                desc: "Разрабатываем персональные решения с учётом особенностей каждого объекта и бюджета клиента.",
              },
            ].map((v) => (
              <ValueCard key={v.title} icon={v.icon} title={v.title} desc={v.desc} />
            ))}
          </div>
        </div>
        <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/10 to-transparent" />
      </section>

      {/* ── SERVICES ── */}
      <section id="services" className="py-24 lg:py-32" style={{ backgroundColor: "#05070A" }}>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <SectionLabel>Что мы предлагаем</SectionLabel>
            <h2
              className="text-4xl sm:text-5xl lg:text-6xl tracking-wide mb-6"
              style={{ fontFamily: "Bebas Neue, sans-serif" }}
            >
              НАШИ УСЛУГИ
            </h2>
            <p className="text-white/60 text-lg max-w-2xl mx-auto">
              Комплексные решения в области охраны объектов, грузов и мероприятий в
              Санкт-Петербурге.
            </p>
          </div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 lg:gap-5">
            {services.map((s) => (
              <a
                key={s.name}
                href="#contacts"
                className="glass border border-white/10 rounded-2xl p-6 group hover:border-[#D4AF37]/40 hover:-translate-y-1 transition-all duration-300 cursor-pointer no-underline"
              >
                <div className="w-12 h-12 rounded-xl bg-[#D4AF37]/10 border border-[#D4AF37]/20 flex items-center justify-center mb-4 text-2xl group-hover:scale-110 transition-transform duration-300">
                  {s.icon}
                </div>
                <h3 className="font-semibold text-base mb-2 group-hover:text-[#D4AF37] transition-colors text-white">
                  {s.name}
                </h3>
                <p className="text-white/45 text-sm leading-relaxed">{s.desc}</p>
              </a>
            ))}
          </div>

          <div className="mt-14 text-center">
            <a
              href="#contacts"
              className="inline-flex items-center gap-2 px-8 py-4 rounded-xl font-semibold bg-[#D4AF37] text-[#05070A] hover:bg-[#E8C84A] transition-colors"
            >
              Рассчитать стоимость охраны →
            </a>
          </div>
        </div>
      </section>

      {/* ── WHY US ── */}
      <section
        id="why-us"
        className="py-24 lg:py-32 relative overflow-hidden"
        style={{ backgroundColor: "#0B111B" }}
      >
        <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-[#D4AF37]/20 to-transparent" />
        <div
          className="absolute inset-0 opacity-[.03]"
          style={{
            backgroundImage: "radial-gradient(circle, #D4AF37 1px, transparent 1px)",
            backgroundSize: "40px 40px",
          }}
        />
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 relative">
          <div className="text-center mb-16">
            <SectionLabel>Наши преимущества</SectionLabel>
            <h2
              className="text-4xl sm:text-5xl lg:text-6xl tracking-wide"
              style={{ fontFamily: "Bebas Neue, sans-serif" }}
            >
              ПОЧЕМУ ВЫБИРАЮТ НАС
            </h2>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6 lg:gap-8">
            {[
              {
                icon: "📋",
                title: "Лицензированная деятельность",
                desc: "Имеем все необходимые лицензии для осуществления частной охранной деятельности для обоих юридических лиц группы.",
              },
              {
                icon: "⭐",
                title: "Ветераны подразделения «Альфа»",
                desc: "Костяк компании — профессионалы с боевым опытом, члены Международной ассоциации ветеранов «АЛЬФА».",
              },
              {
                icon: "🎯",
                title: "Современные средства охраны",
                desc: "Системы видеонаблюдения, тревожные кнопки, GPS-мониторинг транспорта — передовые технологии.",
              },
              {
                icon: "📝",
                title: "Индивидуальный подход",
                desc: "Разрабатываем план охраны конкретно под ваш объект: анализируем риски, предлагаем оптимальное решение.",
              },
              {
                icon: "🔐",
                title: "Полная конфиденциальность",
                desc: "Строго соблюдаем режим коммерческой тайны. Информация об объектах и клиентах не разглашается.",
              },
              {
                icon: "🕐",
                title: "Круглосуточная поддержка",
                desc: "Оперативный центр работает 24/7. Минимальное время реагирования на тревожный сигнал по СПб.",
              },
            ].map((a) => (
              <div
                key={a.title}
                className="flex items-start gap-4 group p-4 rounded-2xl hover:glass hover:border hover:border-white/10 transition-all"
              >
                <div className="w-12 h-12 rounded-xl bg-[#D4AF37]/10 border border-[#D4AF37]/30 flex items-center justify-center shrink-0 text-2xl group-hover:bg-[#D4AF37]/20 transition-colors">
                  {a.icon}
                </div>
                <div>
                  <h3 className="font-semibold text-lg mb-2 group-hover:text-[#D4AF37] transition-colors">
                    {a.title}
                  </h3>
                  <p className="text-white/50 text-sm leading-relaxed">{a.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
        <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/10 to-transparent" />
      </section>

      {/* ── HISTORY ── */}
      <section
        id="history"
        className="py-24 lg:py-32 relative overflow-hidden"
        style={{ backgroundColor: "#0B111B" }}
      >
        <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-[#D4AF37]/20 to-transparent" />
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 relative">
          <div className="text-center mb-16">
            <SectionLabel>Наш путь</SectionLabel>
            <h2
              className="text-4xl sm:text-5xl lg:text-6xl tracking-wide mb-6"
              style={{ fontFamily: "Bebas Neue, sans-serif" }}
            >
              ИСТОРИЯ КОМПАНИИ
            </h2>
          </div>

          <div className="relative space-y-8">
            <div className="hidden md:block absolute left-8 top-0 bottom-0 w-px bg-gradient-to-b from-[#D4AF37]/40 via-[#D4AF37]/20 to-transparent" />

            {[
              {
                year: "1980",
                title: "Олимпийская группа захвата",
                desc: "Создана специальная группа захвата для обеспечения безопасности Олимпийских игр в Москве. Одно из лучших подразделений МВД СССР.",
              },
              {
                year: "1985–95",
                title: "Рота оперативного реагирования (РОР)",
                desc: "Специализация: задержание вооружённых и особо опасных преступников живыми, без применения оружия. Десятки успешных операций.",
              },
              {
                year: "1990-е",
                title: "Формирование СОБР",
                desc: "На базе РОР создан Резерв, впоследствии преобразованный в Специальный отряд быстрого реагирования (СОБР).",
              },
              {
                year: "2002",
                title: "Основание ЧОО «Альфа Юнит-1»",
                desc: "Ветераны элитных силовых структур создали частное охранное предприятие в Санкт-Петербурге, объединив боевой опыт с коммерческой гибкостью.",
              },
              {
                year: "Сегодня",
                title: "Группа компаний — лидер СЗФО",
                desc: "ЧОО «Альфа Юнит-1» + ЧОО «Альфа Безопасность». Более 50 объектов, 200+ сотрудников. Офисы в СПб и Симферополе.",
                highlight: true,
              },
            ].map((event) => (
              <div key={event.year} className="flex gap-8 md:pl-20 relative">
                <div className="hidden md:flex absolute left-5 w-6 h-6 rounded-full bg-[#D4AF37] border-2 border-[#0B111B] items-center justify-center -translate-x-1/2 top-3">
                  <div className="w-2 h-2 rounded-full bg-[#0B111B]" />
                </div>
                <div
                  className={`glass border rounded-2xl p-6 flex-1 ${event.highlight ? "border-[#D4AF37]/30" : "border-white/10"}`}
                >
                  <div
                    className={`inline-block rounded-xl px-4 py-2 mb-4 ${event.highlight ? "bg-[#D4AF37] text-[#05070A]" : "glass border border-[#D4AF37]/30"}`}
                  >
                    <span
                      className={`text-2xl ${event.highlight ? "text-[#05070A]" : "text-[#D4AF37]"}`}
                      style={{ fontFamily: "Bebas Neue, sans-serif" }}
                    >
                      {event.year}
                    </span>
                  </div>
                  <h3 className="font-bold text-xl mb-3">{event.title}</h3>
                  <p className="text-white/55 leading-relaxed">{event.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
        <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/10 to-transparent" />
      </section>

      {/* ── HOW WE WORK ── */}
      <section id="process" className="py-24 lg:py-32" style={{ backgroundColor: "#05070A" }}>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <SectionLabel>Процесс</SectionLabel>
            <h2
              className="text-4xl sm:text-5xl lg:text-6xl tracking-wide"
              style={{ fontFamily: "Bebas Neue, sans-serif" }}
            >
              КАК МЫ РАБОТАЕМ
            </h2>
          </div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-6 lg:gap-8">
            {[
              {
                n: "01",
                title: "Заявка",
                desc: "Позвоните или заполните форму — специалист свяжется с вами в течение 30 минут в рабочее время.",
              },
              {
                n: "02",
                title: "Анализ объекта",
                desc: "Выезд специалиста на объект. Оценка рисков, анализ уязвимостей, составление технического задания.",
              },
              {
                n: "03",
                title: "Решение",
                desc: "Персональный план охраны: численность, режим, оборудование. Коммерческое предложение.",
              },
              {
                n: "04",
                title: "Договор",
                desc: "Заключение договора оказания охранных услуг. Все условия прозрачны, скрытых платежей нет.",
              },
              {
                n: "05",
                title: "Организация охраны",
                desc: "Расстановка сотрудников, монтаж оборудования, инструктаж персонала. Начало охраны объекта.",
              },
              {
                n: "06",
                title: "Контроль качества",
                desc: "Постоянный мониторинг, ежемесячные отчёты клиенту, оперативное реагирование на замечания.",
              },
            ].map((step) => (
              <div key={step.n} className="group">
                <div className="flex items-start gap-5">
                  <div className="w-16 h-16 rounded-2xl glass border border-[#D4AF37]/40 flex items-center justify-center shrink-0 group-hover:bg-[#D4AF37]/10 transition-colors">
                    <span
                      className="text-2xl text-[#D4AF37]"
                      style={{ fontFamily: "Bebas Neue, sans-serif" }}
                    >
                      {step.n}
                    </span>
                  </div>
                  <div className="mt-2">
                    <h3 className="font-bold text-lg mb-2 group-hover:text-[#D4AF37] transition-colors">
                      {step.title}
                    </h3>
                    <p className="text-white/50 text-sm leading-relaxed">{step.desc}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── FAQ ── */}
      <section
        id="faq"
        className="py-24 lg:py-32 relative overflow-hidden"
        style={{ backgroundColor: "#0B111B" }}
      >
        <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-[#D4AF37]/20 to-transparent" />
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <SectionLabel>Вопросы и ответы</SectionLabel>
            <h2
              className="text-4xl sm:text-5xl lg:text-6xl tracking-wide"
              style={{ fontFamily: "Bebas Neue, sans-serif" }}
            >
              ЧАСТО ЗАДАЮТ
            </h2>
          </div>

          <div className="space-y-3">
            {faqItems.map((item, i) => (
              <div
                key={i}
                className={`glass border rounded-2xl overflow-hidden transition-all duration-200 ${openFaq === i ? "border-[#D4AF37]/30" : "border-white/10 hover:border-white/25"}`}
              >
                <button
                  className="w-full flex items-center justify-between gap-4 p-6 text-left"
                  onClick={() => setOpenFaq(openFaq === i ? null : i)}
                  aria-expanded={openFaq === i}
                >
                  <span
                    className={`font-semibold text-base transition-colors ${openFaq === i ? "text-[#D4AF37]" : ""}`}
                  >
                    {item.q}
                  </span>
                  <span
                    className={`text-[#D4AF37] text-xl shrink-0 transition-transform duration-300 ${openFaq === i ? "rotate-45" : ""}`}
                  >
                    +
                  </span>
                </button>
                {openFaq === i && (
                  <div className="faq-answer px-6 pb-6">
                    <p className="text-white/55 leading-relaxed">{item.a}</p>
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
        <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/10 to-transparent" />
      </section>

      {/* ── CONTACTS ── */}
      <section id="contacts" className="py-24 lg:py-32" style={{ backgroundColor: "#05070A" }}>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <SectionLabel>Свяжитесь с нами</SectionLabel>
            <h2
              className="text-4xl sm:text-5xl lg:text-6xl tracking-wide mb-6"
              style={{ fontFamily: "Bebas Neue, sans-serif" }}
            >
              КОНТАКТЫ
            </h2>
            <p className="text-white/60 text-lg max-w-2xl mx-auto">
              Работаем Пн–Пт: 9:00–20:00. По срочным вопросам — звоните в любое время.
            </p>
          </div>

          <div className="grid lg:grid-cols-2 gap-8 lg:gap-12">
            {/* Info */}
            <div>
              <div className="space-y-4 mb-8">
                {[
                  {
                    icon: "📞",
                    label: "Телефоны",
                    value: `${siteMeta.phone1} · ${siteMeta.phone2}`,
                    href: `tel:${siteMeta.phone1}`,
                  },
                  {
                    icon: "✉️",
                    label: "Email",
                    value: siteMeta.email,
                    href: `mailto:${siteMeta.email}`,
                  },
                  { icon: "📍", label: "Адрес", value: siteMeta.address, href: "#" },
                  { icon: "🕐", label: "Режим работы", value: "Пн–Пт: 9:00–20:00", href: "#" },
                ].map((item) => (
                  <a
                    key={item.label}
                    href={item.href}
                    className="flex items-center gap-5 glass border border-white/10 rounded-2xl p-5 hover:border-[#D4AF37]/30 transition-all duration-300 no-underline"
                  >
                    <div className="w-12 h-12 rounded-xl bg-[#D4AF37]/10 border border-[#D4AF37]/20 flex items-center justify-center shrink-0 text-xl">
                      {item.icon}
                    </div>
                    <div>
                      <div className="text-xs text-white/40 uppercase tracking-wider mb-1">
                        {item.label}
                      </div>
                      <div className="font-semibold text-white">{item.value}</div>
                    </div>
                  </a>
                ))}
              </div>

              <div className="flex gap-3 mb-6">
                <a
                  href="https://wa.me/79313625688"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex-1 flex items-center justify-center gap-2 glass border border-white/10 rounded-xl py-3 hover:border-[#25D366]/40 hover:text-[#25D366] transition-all text-sm font-semibold no-underline"
                >
                  💬 WhatsApp
                </a>
                <a
                  href="https://t.me/alfaunit1"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex-1 flex items-center justify-center gap-2 glass border border-white/10 rounded-xl py-3 hover:border-[#2AABEE]/40 hover:text-[#2AABEE] transition-all text-sm font-semibold no-underline"
                >
                  ✈️ Telegram
                </a>
              </div>

              {/* Yandex Map */}
              <div className="rounded-2xl overflow-hidden border border-white/10 h-52">
                <iframe
                  src="https://yandex.ru/map-widget/v1/?ll=30.288936%2C59.910634&z=15&pt=30.288936,59.910634,pm2rdm"
                  width="100%"
                  height="100%"
                  title="Карта — офис Альфа Юнит-1"
                  loading="lazy"
                  className="border-0"
                />
              </div>
            </div>

            {/* Form */}
            <div>
              <div className="glass border border-white/10 rounded-2xl p-7 lg:p-9">
                <h3 className="font-bold text-2xl mb-2">Оставить заявку</h3>
                <p className="text-white/50 text-sm mb-7">
                  Заполните форму — перезвоним в течение 30 минут.
                </p>

                <form ref={formRef} onSubmit={handleContactSubmit} noValidate>
                  <div className="mb-5">
                    <label className="block text-sm text-white/55 mb-1.5">Услуга</label>
                    <select
                      name="service"
                      className="w-full glass border border-white/15 rounded-xl px-4 py-3 text-white/80 text-sm focus:border-[#D4AF37] outline-none"
                      style={{ backgroundColor: "rgba(255,255,255,.06)" }}
                    >
                      <option value="">— Выберите услугу —</option>
                      {services.map((s) => (
                        <option key={s.name} value={s.name}>
                          {s.name}
                        </option>
                      ))}
                      <option value="Другое">Другое / Консультация</option>
                    </select>
                  </div>
                  <div className="mb-5">
                    <label className="block text-sm text-white/55 mb-1.5">Имя *</label>
                    <input
                      name="name"
                      type="text"
                      className="w-full glass border border-white/15 rounded-xl px-4 py-3 text-white text-sm focus:border-[#D4AF37] outline-none placeholder:text-white/25"
                      placeholder="Иван Петров"
                      required
                    />
                  </div>
                  <div className="mb-5">
                    <label className="block text-sm text-white/55 mb-1.5">Телефон *</label>
                    <input
                      name="phone"
                      type="tel"
                      className="w-full glass border border-white/15 rounded-xl px-4 py-3 text-white text-sm focus:border-[#D4AF37] outline-none placeholder:text-white/25"
                      placeholder="+7 (___) ___-__-__"
                      required
                    />
                  </div>
                  <div className="mb-7">
                    <label className="block text-sm text-white/55 mb-1.5">Описание объекта</label>
                    <textarea
                      name="message"
                      rows={3}
                      className="w-full glass border border-white/15 rounded-xl px-4 py-3 text-white text-sm focus:border-[#D4AF37] outline-none placeholder:text-white/25 resize-none"
                      placeholder="Тип объекта, площадь, пожелания..."
                    />
                  </div>

                  <button
                    type="submit"
                    disabled={formState === "loading"}
                    className="w-full py-4 rounded-xl font-bold bg-[#D4AF37] text-[#05070A] hover:bg-[#E8C84A] disabled:opacity-70 transition-colors"
                  >
                    {formState === "loading" ? "Отправка..." : "Отправить заявку"}
                  </button>

                  {formState === "success" && (
                    <div className="mt-4 p-4 rounded-xl bg-green-500/10 border border-green-500/30 text-green-400 text-sm text-center">
                      {formMsg}
                    </div>
                  )}
                  {formState === "error" && (
                    <div className="mt-4 p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400 text-sm text-center">
                      {formMsg}
                    </div>
                  )}
                  <p className="text-white/30 text-xs mt-4 text-center">
                    Нажимая кнопку, вы соглашаетесь с политикой конфиденциальности.
                  </p>
                </form>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── FOOTER ── */}
      <footer className="border-t border-white/10" style={{ backgroundColor: "#0B111B" }}>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-14">
          <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-10">
            <div className="lg:col-span-2">
              <a href="/" className="flex items-center gap-3 mb-5 no-underline">
                <div className="w-10 h-10 rounded-lg bg-[#D4AF37] flex items-center justify-center shrink-0">
                  <svg width="22" height="22" viewBox="0 0 24 24" fill="none">
                    <path
                      d="M12 2L3 7V12C3 17.25 6.96 22.18 12 23.5C17.04 22.18 21 17.25 21 12V7L12 2Z"
                      fill="#05070A"
                    />
                    <path
                      d="M9 12L11 14L15 10"
                      stroke="#D4AF37"
                      strokeWidth="1.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                </div>
                <div>
                  <div
                    className="font-display text-xl text-white tracking-wider"
                    style={{ fontFamily: "Bebas Neue, sans-serif" }}
                  >
                    АЛЬФА ЮНИТ-1
                  </div>
                  <div className="text-xs text-white/40 tracking-widest uppercase">
                    Охранная компания
                  </div>
                </div>
              </a>
              <p className="text-white/45 text-sm leading-relaxed mb-6 max-w-sm">
                Лицензированная охранная организация. Члены Международной ассоциации ветеранов
                подразделения «АЛЬФА». Работаем с 2002 года.
              </p>
              <a
                href={`tel:${siteMeta.phone1}`}
                className="text-[#D4AF37] font-semibold hover:text-[#E8C84A] transition-colors text-sm no-underline"
              >
                {siteMeta.phone1}
              </a>
            </div>

            <div>
              <h4 className="font-semibold text-sm uppercase tracking-wider text-white/80 mb-5">
                Услуги
              </h4>
              <ul className="space-y-2.5">
                {services.slice(0, 6).map((s) => (
                  <li key={s.name}>
                    <a
                      href="#services"
                      className="text-white/45 text-sm hover:text-[#D4AF37] transition-colors no-underline"
                    >
                      {s.name}
                    </a>
                  </li>
                ))}
              </ul>
            </div>

            <div>
              <h4 className="font-semibold text-sm uppercase tracking-wider text-white/80 mb-5">
                Компания
              </h4>
              <ul className="space-y-2.5 text-sm text-white/45">
                {[
                  ["#about", "О компании"],
                  ["#history", "История"],
                  ["#contacts", "Контакты"],
                ].map(([href, label]) => (
                  <li key={href}>
                    <a href={href} className="hover:text-[#D4AF37] transition-colors no-underline">
                      {label}
                    </a>
                  </li>
                ))}
              </ul>
              <div className="mt-8">
                <h4 className="font-semibold text-sm uppercase tracking-wider text-white/80 mb-3">
                  Адрес
                </h4>
                <p className="text-white/45 text-sm leading-relaxed">{siteMeta.address}</p>
                <p className="text-white/30 text-xs mt-2">Пн–Пт: 9:00–20:00</p>
              </div>
            </div>
          </div>
        </div>
        <div className="border-t border-white/10">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-5 flex flex-col sm:flex-row items-center justify-between gap-3">
            <p className="text-white/30 text-xs">
              © {new Date().getFullYear()} ЧОО «Альфа Юнит-1». Все права защищены.
            </p>
            <div className="flex gap-5">
              <a
                href="#"
                className="text-white/30 text-xs hover:text-white/60 transition-colors no-underline"
              >
                Политика конфиденциальности
              </a>
              <a
                href="#"
                className="text-white/30 text-xs hover:text-white/60 transition-colors no-underline"
              >
                Реквизиты
              </a>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}

// ── Shared small components ────────────────────────────────────────────────
function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <div className="inline-flex items-center gap-2 mb-4">
      <div className="w-8 h-px bg-[#D4AF37]" />
      <span className="text-[#D4AF37] text-xs font-semibold tracking-[.3em] uppercase">
        {children}
      </span>
      <div className="w-8 h-px bg-[#D4AF37]" />
    </div>
  );
}

function ValueCard({ icon, title, desc }: { icon: string; title: string; desc: string }) {
  return (
    <div className="glass border border-white/10 rounded-2xl p-8 group hover:border-[#D4AF37]/30 transition-all duration-300">
      <div className="w-12 h-12 rounded-xl bg-[#D4AF37]/10 border border-[#D4AF37]/20 flex items-center justify-center mb-5 text-2xl group-hover:bg-[#D4AF37]/20 transition-colors">
        {icon}
      </div>
      <h3 className="text-xl font-semibold mb-3 group-hover:text-[#D4AF37] transition-colors">
        {title}
      </h3>
      <p className="text-white/50 leading-relaxed text-sm">{desc}</p>
    </div>
  );
}
