"use strict";

// Shared site chrome (top nav + footer) injected into every page. Pages include
// a <div id="site-nav"></div> and <div id="site-footer"></div> placeholder.

(function () {
  const NAV_LINKS = [
    { href: "/qr/business", label: "Business Cards" },
    { href: "/qr/wifi", label: "WiFi QR" },
    { href: "/learn", label: "Learn QR" },
    { href: "/faq", label: "FAQ" },
    { href: "/my-links", label: "My Links" },
  ];

  function currentPath() {
    return location.pathname.replace(/\/+$/, "") || "/";
  }

  function isActive(href) {
    const path = currentPath();
    if (href === "/") return path === "/";
    return path === href || path.startsWith(href + "/");
  }

  function renderNav() {
    const mount = document.getElementById("site-nav");
    if (!mount) return;
    const links = NAV_LINKS.map(
      (l) =>
        `<a class="nav-link${isActive(l.href) ? " active" : ""}" href="${l.href}">${l.label}</a>`
    ).join("");
    mount.outerHTML = `
      <nav class="brand-bar">
        <a class="brand" href="/"><img src="/logo.png" alt="Qrunex" class="brand-logo" /></a>
        <button class="nav-toggle" id="nav-toggle" aria-label="Menu" aria-expanded="false">☰</button>
        <div class="nav-links" id="nav-links">${links}</div>
      </nav>`;

    const toggle = document.getElementById("nav-toggle");
    const linksEl = document.getElementById("nav-links");
    if (toggle && linksEl) {
      toggle.addEventListener("click", () => {
        const open = linksEl.classList.toggle("open");
        toggle.setAttribute("aria-expanded", String(open));
      });
    }
  }

  function renderFooter() {
    const mount = document.getElementById("site-footer");
    if (!mount) return;
    const year = new Date().getFullYear();
    mount.outerHTML = `
      <footer class="footer">
        <div class="footer-links">
          <a href="/">Home</a>
          <a href="/learn">Learn QR</a>
          <a href="/faq">FAQ</a>
          <a href="/my-links">My Links</a>
        </div>
        <span>Qrunex · Free QR Code Generator &amp; URL Shortener · &copy; ${year}</span>
      </footer>`;
  }

  document.addEventListener("DOMContentLoaded", () => {
    renderNav();
    renderFooter();
  });
})();
