/**
 * Альфа Юнит-1 — main.js
 * Animations · Counters · FAQ · Contact Form · Navigation
 * 2026 standard — vanilla JS, zero framework dependencies
 */

"use strict";

/* ════════════════════════════════════════════════════
   1. NAVIGATION — Scroll shadow + mobile menu
════════════════════════════════════════════════════ */
(function initNav() {
  const navbar = document.getElementById("navbar");
  const burger = document.getElementById("burger");
  const mobileMenu = document.getElementById("mobile-menu");
  const mobileLinks = document.querySelectorAll(".mobile-nav-link");

  if (!navbar) return;

  // Add .scrolled class when page is scrolled
  function onScroll() {
    navbar.classList.toggle("scrolled", window.scrollY > 40);
  }
  window.addEventListener("scroll", onScroll, { passive: true });
  onScroll();

  // Burger toggle
  if (burger && mobileMenu) {
    burger.addEventListener("click", () => {
      const isOpen = !mobileMenu.classList.contains("hidden");
      mobileMenu.classList.toggle("hidden", isOpen);
      burger.classList.toggle("open", !isOpen);
      burger.setAttribute("aria-expanded", String(!isOpen));
    });

    // Close on mobile link click
    mobileLinks.forEach((link) => {
      link.addEventListener("click", () => {
        mobileMenu.classList.add("hidden");
        burger.classList.remove("open");
        burger.setAttribute("aria-expanded", "false");
      });
    });
  }

  // Highlight active nav link on scroll
  const sections = document.querySelectorAll("section[id]");
  const navLinks = document.querySelectorAll(".nav-link");

  function highlightNav() {
    const scrollY = window.scrollY + 120;
    sections.forEach((section) => {
      const top = section.offsetTop;
      const height = section.offsetHeight;
      const id = section.getAttribute("id");
      const link = document.querySelector(`.nav-link[href="#${id}"]`);
      if (link) {
        link.classList.toggle("active", scrollY >= top && scrollY < top + height);
      }
    });
  }
  window.addEventListener("scroll", highlightNav, { passive: true });
})();

/* ════════════════════════════════════════════════════
   2. AOS — Animate on Scroll
════════════════════════════════════════════════════ */
(function initAOS() {
  if (typeof AOS !== "undefined") {
    AOS.init({
      duration: 700,
      easing: "ease-out-cubic",
      once: true,
      offset: 80,
      delay: 0,
    });
  }
})();

/* ════════════════════════════════════════════════════
   3. GSAP — Hero + ScrollTrigger animations
════════════════════════════════════════════════════ */
(function initGSAP() {
  if (typeof gsap === "undefined" || typeof ScrollTrigger === "undefined") return;

  gsap.registerPlugin(ScrollTrigger);

  // Hero entrance (overrides CSS fallback)
  const heroTl = gsap.timeline({ defaults: { ease: "power3.out" } });

  if (document.querySelector(".hero-badge")) {
    heroTl
      .from(".hero-badge", { y: -20, opacity: 0, duration: 0.7, delay: 0.2 })
      .from(".hero-title", { y: 40, opacity: 0, duration: 0.8 }, "-=0.4")
      .from(".hero-subtitle", { y: 30, opacity: 0, duration: 0.7 }, "-=0.5")
      .from(".hero-cta", { y: 25, opacity: 0, duration: 0.6 }, "-=0.4")
      .from(".hero-trust", { y: 20, opacity: 0, duration: 0.5 }, "-=0.4")
      .from(".trust-item", { x: -15, opacity: 0, duration: 0.4, stagger: 0.1 }, "-=0.3");
  }

  // Parallax on hero background orbs
  gsap.to(".hero-bg .rounded-full:first-of-type", {
    yPercent: -30,
    ease: "none",
    scrollTrigger: { trigger: "#hero", start: "top top", end: "bottom top", scrub: 1 },
  });

  // Section headings — slide in from left
  gsap.utils.toArray(".section-header").forEach((el) => {
    gsap.from(el, {
      x: -40,
      opacity: 0,
      duration: 0.8,
      ease: "power3.out",
      scrollTrigger: { trigger: el, start: "top 85%", toggleActions: "play none none none" },
    });
  });

  // Service cards stagger on scroll
  const serviceGrid = document.querySelector(
    ".grid.sm\\:grid-cols-2.lg\\:grid-cols-3.xl\\:grid-cols-4",
  );
  if (serviceGrid) {
    gsap.from(serviceGrid.querySelectorAll(".service-card"), {
      y: 40,
      opacity: 0,
      duration: 0.5,
      stagger: 0.06,
      ease: "power2.out",
      scrollTrigger: {
        trigger: serviceGrid,
        start: "top 80%",
        toggleActions: "play none none none",
      },
    });
  }

  // Stats cards bounce in
  gsap.utils.toArray(".stat-card").forEach((card, i) => {
    gsap.from(card, {
      scale: 0.85,
      opacity: 0,
      duration: 0.6,
      delay: i * 0.08,
      ease: "back.out(1.5)",
      scrollTrigger: { trigger: card, start: "top 88%", toggleActions: "play none none none" },
    });
  });

  // Timeline items fade in alternating sides
  gsap.utils.toArray(".timeline-item").forEach((item, i) => {
    gsap.from(item, {
      x: i % 2 === 0 ? -50 : 50,
      opacity: 0,
      duration: 0.75,
      ease: "power3.out",
      scrollTrigger: { trigger: item, start: "top 85%", toggleActions: "play none none none" },
    });
  });

  // Process steps stagger
  gsap.utils.toArray(".process-step").forEach((step, i) => {
    gsap.from(step, {
      y: 30,
      opacity: 0,
      duration: 0.5,
      delay: i * 0.1,
      ease: "power2.out",
      scrollTrigger: { trigger: step, start: "top 88%", toggleActions: "play none none none" },
    });
  });

  // Contact info items slide in
  gsap.utils.toArray(".contact-info-item").forEach((item, i) => {
    gsap.from(item, {
      x: -30,
      opacity: 0,
      duration: 0.45,
      delay: i * 0.08,
      ease: "power2.out",
      scrollTrigger: { trigger: item, start: "top 90%", toggleActions: "play none none none" },
    });
  });
})();

/* ════════════════════════════════════════════════════
   4. COUNTERS — Animate numbers when in viewport
════════════════════════════════════════════════════ */
(function initCounters() {
  const counters = document.querySelectorAll(".counter[data-target]");
  if (!counters.length) return;

  function animateCounter(el) {
    const rawTarget = el.dataset.target;
    // Strip non-numeric chars (e.g. "50+", "200+") — keep suffix
    const numPart = parseFloat(rawTarget);
    const suffix = rawTarget.replace(/[\d.]/g, ""); // e.g. "+"

    if (isNaN(numPart)) return;

    const duration = 2000; // ms
    const startTime = performance.now();

    function update(now) {
      const elapsed = now - startTime;
      const progress = Math.min(elapsed / duration, 1);
      // Ease out cubic
      const eased = 1 - Math.pow(1 - progress, 3);
      const current = Math.round(eased * numPart);

      el.textContent = current + suffix;

      if (progress < 1) {
        requestAnimationFrame(update);
      }
    }

    requestAnimationFrame(update);
  }

  // Observe counters — fire once when they enter the viewport
  const observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (entry.isIntersecting) {
          animateCounter(entry.target);
          observer.unobserve(entry.target);
        }
      });
    },
    { threshold: 0.4 },
  );

  counters.forEach((counter) => observer.observe(counter));
})();

/* ════════════════════════════════════════════════════
   5. FAQ — Accordion
════════════════════════════════════════════════════ */
(function initFAQ() {
  const container = document.getElementById("faq-container");
  if (!container) return;

  container.addEventListener("click", (e) => {
    const btn = e.target.closest(".faq-question");
    if (!btn) return;

    const item = btn.closest(".faq-item");
    const answer = item.querySelector(".faq-answer");
    const isOpen = !answer.classList.contains("hidden");

    // Close all open items
    container.querySelectorAll(".faq-item.open").forEach((openItem) => {
      if (openItem !== item) {
        openItem.classList.remove("open");
        openItem.querySelector(".faq-answer").classList.add("hidden");
        openItem.querySelector(".faq-question").setAttribute("aria-expanded", "false");
      }
    });

    // Toggle current item
    item.classList.toggle("open", !isOpen);
    answer.classList.toggle("hidden", isOpen);
    btn.setAttribute("aria-expanded", String(!isOpen));
  });
})();

/* ════════════════════════════════════════════════════
   6. CONTACT FORM — AJAX submission
════════════════════════════════════════════════════ */
(function initContactForm() {
  const form = document.getElementById("contact-form");
  if (!form) return;

  const submitBtn = document.getElementById("form-submit");
  const submitText = document.getElementById("form-submit-text");
  const spinner = document.getElementById("form-spinner");
  const successEl = document.getElementById("form-success");
  const errorEl = document.getElementById("form-error");

  function setLoading(loading) {
    submitBtn.disabled = loading;
    submitText.textContent = loading ? "Отправка..." : "Отправить заявку";
    spinner.classList.toggle("hidden", !loading);
  }

  function showMessage(el, msg) {
    el.textContent = msg;
    el.classList.remove("hidden");
    setTimeout(() => el.classList.add("hidden"), 8000);
  }

  function validateForm(data) {
    const name = data.name.trim();
    const phone = data.phone.trim();

    if (!name) {
      return "Укажите ваше имя.";
    }
    if (!phone || phone.replace(/\D/g, "").length < 10) {
      return "Укажите корректный номер телефона.";
    }
    return null;
  }

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    successEl.classList.add("hidden");
    errorEl.classList.add("hidden");

    // Remove field error states
    form.querySelectorAll(".form-input.error").forEach((f) => f.classList.remove("error"));

    const data = {
      name: form.querySelector('[name="name"]')?.value.trim() || "",
      phone: form.querySelector('[name="phone"]')?.value.trim() || "",
      email: form.querySelector('[name="email"]')?.value.trim() || "",
      message: form.querySelector('[name="message"]')?.value.trim() || "",
      service: form.querySelector('[name="service"]')?.value || "",
    };

    const validationError = validateForm(data);
    if (validationError) {
      showMessage(errorEl, validationError);
      if (!data.name) form.querySelector('[name="name"]')?.classList.add("error");
      if (!data.phone) form.querySelector('[name="phone"]')?.classList.add("error");
      return;
    }

    setLoading(true);

    try {
      const res = await fetch("/contact", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });

      const json = await res.json();

      if (!res.ok) {
        throw new Error(json.error || `HTTP ${res.status}`);
      }

      // Success
      showMessage(successEl, json.message || "Заявка принята! Мы свяжемся с вами.");
      form.reset();

      // Pulse effect on button
      submitBtn.classList.add("ring-2", "ring-gold", "ring-offset-2", "ring-offset-bg-secondary");
      setTimeout(
        () =>
          submitBtn.classList.remove(
            "ring-2",
            "ring-gold",
            "ring-offset-2",
            "ring-offset-bg-secondary",
          ),
        2000,
      );
    } catch (err) {
      showMessage(errorEl, err.message || "Ошибка отправки. Позвоните нам напрямую.");
    } finally {
      setLoading(false);
    }
  });

  // Phone formatting helper
  const phoneInput = form.querySelector('[name="phone"]');
  if (phoneInput) {
    phoneInput.addEventListener("input", (e) => {
      let val = e.target.value.replace(/\D/g, "");
      if (val.startsWith("7") || val.startsWith("8")) {
        val = "+7" + val.slice(1);
      } else if (val.length > 0 && !val.startsWith("+")) {
        val = "+7" + val;
      }
      e.target.value = val.slice(0, 12);
    });
  }
})();

/* ════════════════════════════════════════════════════
   7. SMOOTH SCROLL — anchor links fallback
   (html { scroll-behavior: smooth } handles most cases)
════════════════════════════════════════════════════ */
(function initSmoothScroll() {
  document.querySelectorAll('a[href^="#"]').forEach((link) => {
    link.addEventListener("click", (e) => {
      const id = link.getAttribute("href").slice(1);
      if (!id) return;
      const target = document.getElementById(id);
      if (!target) return;
      e.preventDefault();
      const offset = 80; // navbar height
      const top = target.getBoundingClientRect().top + window.scrollY - offset;
      window.scrollTo({ top, behavior: "smooth" });
    });
  });
})();

/* ════════════════════════════════════════════════════
   8. SERVICE CARD — "Подробнее" tooltip scroll
════════════════════════════════════════════════════ */
(function initServiceCards() {
  document.querySelectorAll(".service-card").forEach((card) => {
    card.addEventListener("click", () => {
      const contactsSection = document.getElementById("contacts");
      if (contactsSection) {
        const top = contactsSection.getBoundingClientRect().top + window.scrollY - 80;
        window.scrollTo({ top, behavior: "smooth" });
      }
    });
  });
})();

/* ════════════════════════════════════════════════════
   9. SCROLL PROGRESS — thin gold line at top
════════════════════════════════════════════════════ */
(function initScrollProgress() {
  const bar = document.createElement("div");
  bar.style.cssText = [
    "position:fixed",
    "top:0",
    "left:0",
    "width:0",
    "height:2px",
    "background:linear-gradient(90deg,#D4AF37,#E8C84A)",
    "z-index:9999",
    "transition:width 0.1s linear",
    "pointer-events:none",
  ].join(";");
  document.body.prepend(bar);

  window.addEventListener(
    "scroll",
    () => {
      const pct = (window.scrollY / (document.body.scrollHeight - window.innerHeight)) * 100;
      bar.style.width = Math.min(pct, 100) + "%";
    },
    { passive: true },
  );
})();

/* ════════════════════════════════════════════════════
   10. BACK TO TOP — appears after scrolling 600px
════════════════════════════════════════════════════ */
(function initBackToTop() {
  const btn = document.createElement("button");
  btn.innerHTML = `<svg width="20" height="20" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 15.75l7.5-7.5 7.5 7.5"/></svg>`;
  btn.setAttribute("aria-label", "Вернуться наверх");
  btn.style.cssText = [
    "position:fixed",
    "bottom:1.5rem",
    "right:1.5rem",
    "width:2.75rem",
    "height:2.75rem",
    "background:rgba(212,175,55,0.15)",
    "backdrop-filter:blur(12px)",
    "border:1px solid rgba(212,175,55,0.35)",
    "border-radius:0.75rem",
    "color:#D4AF37",
    "display:flex",
    "align-items:center",
    "justify-content:center",
    "cursor:pointer",
    "z-index:998",
    "opacity:0",
    "transform:translateY(10px)",
    "transition:opacity 0.3s ease,transform 0.3s ease,background 0.2s ease",
    "pointer-events:none",
  ].join(";");

  document.body.appendChild(btn);

  window.addEventListener(
    "scroll",
    () => {
      const visible = window.scrollY > 600;
      btn.style.opacity = visible ? "1" : "0";
      btn.style.transform = visible ? "translateY(0)" : "translateY(10px)";
      btn.style.pointerEvents = visible ? "auto" : "none";
    },
    { passive: true },
  );

  btn.addEventListener("click", () => window.scrollTo({ top: 0, behavior: "smooth" }));
  btn.addEventListener("mouseenter", () => {
    btn.style.background = "rgba(212,175,55,0.25)";
  });
  btn.addEventListener("mouseleave", () => {
    btn.style.background = "rgba(212,175,55,0.15)";
  });
})();
