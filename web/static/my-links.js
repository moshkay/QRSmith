"use strict";

(function () {
  const KEY = "qrunex_links";

  const listEl = document.getElementById("ml-list");
  const statsEl = document.getElementById("ml-stats");
  const searchEl = document.getElementById("ml-search");
  const sortEl = document.getElementById("ml-sort");

  let links = load();
  let query = "";

  function load() {
    try {
      const raw = JSON.parse(localStorage.getItem(KEY) || "[]");
      return Array.isArray(raw) ? raw : [];
    } catch (_) {
      return [];
    }
  }
  function save() {
    localStorage.setItem(KEY, JSON.stringify(links));
  }

  function view() {
    let items = links.slice();
    if (query) {
      items = items.filter(
        (l) =>
          (l.url || "").toLowerCase().includes(query) ||
          (l.shortUrl || "").toLowerCase().includes(query) ||
          (l.code || "").toLowerCase().includes(query)
      );
    }
    const sort = sortEl.value;
    items.sort((a, b) => {
      if (sort === "clicks") return (b.clicks || 0) - (a.clicks || 0);
      const ta = new Date(a.createdAt || 0).getTime();
      const tb = new Date(b.createdAt || 0).getTime();
      return sort === "old" ? ta - tb : tb - ta;
    });
    return items;
  }

  function render() {
    const items = view();
    if (!links.length) {
      listEl.innerHTML = `
        <div class="ml-empty">
          <div class="ml-empty-ico">🔗</div>
          <h3>No links yet</h3>
          <p class="muted">Shorten a URL and it'll show up here with click tracking.</p>
          <a class="cta compact" href="/?tab=short">+ Create your first link</a>
        </div>`;
      statsEl.hidden = true;
      return;
    }

    listEl.innerHTML = items.length
      ? items.map(card).join("")
      : `<p class="faq-empty">No links match your search.</p>`;

    // Aggregate stats (over all links, not just filtered)
    const total = links.length;
    const clicks = links.reduce((s, l) => s + (l.clicks || 0), 0);
    const avg = total ? Math.round(clicks / total) : 0;
    statsEl.hidden = false;
    statsEl.innerHTML = `
      <div class="ml-stat"><b class="orange">${total}</b><span>Total Links</span></div>
      <div class="ml-stat"><b class="green">${clicks}</b><span>Total Clicks</span></div>
      <div class="ml-stat"><b class="orange">${avg}</b><span>Avg. Clicks per Link</span></div>`;

    bindCards();
  }

  function card(l) {
    const created = l.createdAt ? new Date(l.createdAt).toLocaleDateString() : "—";
    const clicks = l.clicks || 0;
    return `
      <div class="ml-card" data-code="${escapeAttr(l.code)}">
        <div class="ml-card-top">
          <a class="ml-short" href="${escapeAttr(l.shortUrl)}" target="_blank" rel="noopener">${escapeHtml(l.shortUrl)}</a>
          <span class="ml-clicks">${clicks} click${clicks === 1 ? "" : "s"}</span>
          <div class="ml-card-actions">
            <button class="icon-btn" data-act="copy" title="Copy">⧉</button>
            <button class="icon-btn" data-act="qr" title="Make QR code">▦</button>
            <a class="icon-btn" href="${escapeAttr(l.shortUrl)}" target="_blank" rel="noopener" title="Open">↗</a>
            <button class="icon-btn danger" data-act="delete" title="Delete">🗑</button>
          </div>
        </div>
        <div class="ml-orig"><span>Original URL:</span> ${escapeHtml(l.url)}</div>
        <div class="ml-meta">Created: ${created} · Code: ${escapeHtml(l.code)}</div>
      </div>`;
  }

  function bindCards() {
    listEl.querySelectorAll(".ml-card").forEach((cardEl) => {
      const code = cardEl.dataset.code;
      const link = links.find((l) => l.code === code);
      cardEl.querySelectorAll("[data-act]").forEach((btn) => {
        btn.addEventListener("click", (e) => {
          e.preventDefault();
          const act = btn.dataset.act;
          if (act === "copy") copy(link.shortUrl, btn);
          else if (act === "qr") location.href = "/?type=url&prefill=" + encodeURIComponent(link.shortUrl);
          else if (act === "delete") remove(code);
        });
      });
    });
  }

  function copy(text, btn) {
    if (navigator.clipboard) navigator.clipboard.writeText(text);
    const orig = btn.textContent;
    btn.textContent = "✓";
    setTimeout(() => (btn.textContent = orig), 1000);
  }

  function remove(code) {
    links = links.filter((l) => l.code !== code);
    save();
    render();
  }

  async function refreshClicks() {
    await Promise.all(
      links.map(async (l) => {
        try {
          const res = await fetch("/api/v1/short-links/" + encodeURIComponent(l.code));
          const out = await res.json();
          if (out.status) l.clicks = out.entity.clicks;
        } catch (_) {
          /* keep cached */
        }
      })
    );
    save();
    render();
  }

  searchEl.addEventListener("input", () => {
    query = searchEl.value.trim().toLowerCase();
    render();
  });
  sortEl.addEventListener("change", render);
  document.getElementById("ml-refresh").addEventListener("click", refreshClicks);

  function escapeHtml(s) {
    return String(s == null ? "" : s).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
  }
  function escapeAttr(s) {
    return escapeHtml(s);
  }

  render();
  if (links.length) refreshClicks();
})();
