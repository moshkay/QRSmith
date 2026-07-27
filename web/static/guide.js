"use strict";

(function () {
  const guides = window.QR_GUIDES || [];
  const bySlug = window.QR_GUIDE_BY_SLUG || {};

  // Slug from /qr/{slug}
  const parts = location.pathname.replace(/\/+$/, "").split("/");
  const slug = parts[parts.length - 1];
  const guide = bySlug[slug];

  if (!guide) {
    document.getElementById("guide-main").innerHTML = `
      <div class="card">
        <h2>Guide not found</h2>
        <p class="muted">We don't have a guide for "${escapeHtml(slug)}".</p>
        <a class="cta" href="/learn">Browse all guides</a>
      </div>`;
    return;
  }

  document.title = `${guide.label} QR Codes — Qrunex`;

  // Breadcrumbs
  document.getElementById("crumbs").innerHTML =
    `<a href="/">Home</a> <span>›</span> <a href="/learn">QR Instructions</a> <span>›</span> <b>${guide.label} QR Codes</b>`;

  // Sidebar list of all types
  document.getElementById("guide-side-list").innerHTML = guides
    .map(
      (g) =>
        `<a class="guide-side-link${g.slug === slug ? " active" : ""}" href="/qr/${g.slug}">
           <span class="gs-ico">${g.icon}</span><span>${g.label}</span>
         </a>`
    )
    .join("");

  // Main content
  const benefits = guide.benefits
    ? `<div class="guide-block">
         <h3>Benefits</h3>
         <ul class="check-list">${guide.benefits.map((b) => `<li>${escapeHtml(b)}</li>`).join("")}</ul>
       </div>`
    : "";

  document.getElementById("guide-main").innerHTML = `
    <section class="card guide-header">
      <div class="guide-header-ico">${guide.icon}</div>
      <h1>${guide.label} QR Codes</h1>
      <p>${escapeHtml(guide.headerDesc)}</p>
    </section>

    <section class="card">
      <div class="guide-cols">
        <div>
          <h2>What are ${guide.label} QR Codes?</h2>
          <p class="muted">${escapeHtml(guide.whatIs)}</p>

          ${benefits}

          <div class="guide-block">
            <h3>Common uses</h3>
            <ul class="check-list">${guide.commonUses.map((u) => `<li>${escapeHtml(u)}</li>`).join("")}</ul>
          </div>

          <a class="cta" href="/?type=${guide.slug}">▦ Create a ${guide.label} QR Code</a>
        </div>

        <div class="guide-preview">
          <p class="section-label">Live Preview</p>
          <div class="preview-frame">
            <img id="guide-qr" alt="${guide.label} QR sample" />
            <div class="ph" id="guide-ph"><span class="ph-ico">▦</span><span>Loading sample…</span></div>
          </div>
          <p class="hint center">Sample — create your own with your details.</p>
        </div>
      </div>
    </section>`;

  renderSample(guide.sample);

  async function renderSample(sample) {
    try {
      const res = await fetch("/api/v1/qr-codes", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          type: sample.type,
          data: sample.data,
          format: "svg",
          size: 320,
          foreground: "#0F172A",
          background: "#FFFFFF",
        }),
      });
      const out = await res.json();
      if (out.status) {
        const img = document.getElementById("guide-qr");
        img.src = out.entity.image;
        img.classList.add("ready");
        document.getElementById("guide-ph").classList.add("hidden");
      } else {
        document.getElementById("guide-ph").querySelector("span:last-child").textContent =
          "Preview unavailable";
      }
    } catch (_) {
      const ph = document.getElementById("guide-ph");
      if (ph) ph.querySelector("span:last-child").textContent = "Preview unavailable";
    }
  }

  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
  }
})();
