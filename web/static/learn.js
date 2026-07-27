"use strict";

(function () {
  const grid = document.getElementById("learn-grid");
  if (!grid || !window.QR_GUIDES) return;

  grid.innerHTML = window.QR_GUIDES.map((g) => {
    const badge = g.popular ? '<span class="pill">Popular</span>' : "";
    return `
      <a class="learn-card" href="/qr/${g.slug}">
        <div class="learn-ico">${g.icon}</div>
        <div class="learn-body">
          <h3>${g.label} ${badge}</h3>
          <p>${g.tagline}</p>
          <span class="learn-cta">Learn how to create →</span>
        </div>
      </a>`;
  }).join("");
})();
