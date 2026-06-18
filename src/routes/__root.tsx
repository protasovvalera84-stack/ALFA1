import {
  HeadContent,
  Outlet,
  Scripts,
  ScrollRestoration,
  createRootRoute,
} from "@tanstack/react-router";
import "../styles.css";
import { siteMeta } from "../lib/site-meta";

const schemaLD = {
  "@context": "https://schema.org",
  "@type": "LocalBusiness",
  name: "Альфа Юнит-1",
  telephone: "+7-931-362-56-88",
  email: "admin@alfaunit1.ru",
  address: {
    "@type": "PostalAddress",
    streetAddress: "ул. Лифляндская, д. 3",
    addressLocality: "Санкт-Петербург",
    postalCode: "190020",
    addressCountry: "RU",
  },
  openingHours: "Mo-Fr 09:00-20:00",
  url: "https://alfaunit1.ru",
};

export const Route = createRootRoute({
  head() {
    return {
      meta: [
        { charSet: "UTF-8" },
        { name: "viewport", content: "width=device-width, initial-scale=1.0" },
        { title: siteMeta.title },
        { name: "description", content: siteMeta.description },
        { name: "robots", content: "index, follow" },
        { property: "og:type", content: "website" },
        { property: "og:title", content: siteMeta.title },
        { property: "og:description", content: siteMeta.description },
        { property: "og:locale", content: "ru_RU" },
        { name: "twitter:card", content: "summary_large_image" },
        { name: "twitter:title", content: "Альфа Юнит-1" },
        { name: "twitter:description", content: siteMeta.description },
      ],
      links: [
        { rel: "preconnect", href: "https://fonts.googleapis.com" },
        {
          rel: "preconnect",
          href: "https://fonts.gstatic.com",
          crossOrigin: "anonymous",
        },
        {
          rel: "stylesheet",
          href: "https://fonts.googleapis.com/css2?family=Bebas+Neue&family=Inter:wght@300;400;500;600;700&display=swap",
        },
      ],
      scripts: [
        {
          type: "application/ld+json",
          children: JSON.stringify(schemaLD),
        },
      ],
    };
  },
  component: RootComponent,
});

function RootComponent() {
  return (
    <html lang="ru" className="scroll-smooth">
      <head>
        <HeadContent />
      </head>
      <body
        className="antialiased overflow-x-hidden"
        style={{ backgroundColor: "#05070A", color: "#ffffff" }}
      >
        <div id="scroll-progress" />
        <Outlet />
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  );
}
