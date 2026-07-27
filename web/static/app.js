"use strict";

const $ = (id) => document.getElementById(id);

// ---------------------------------------------------------------------------
// Content type configuration (drives the type grid, dynamic fields, guides).
// ---------------------------------------------------------------------------

const ICONS = {
  url: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M10 13a5 5 0 0 0 7 0l3-3a5 5 0 0 0-7-7l-1 1"/><path d="M14 11a5 5 0 0 0-7 0l-3 3a5 5 0 0 0 7 7l1-1"/></svg>',
  text: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><rect x="4" y="3" width="16" height="18" rx="2"/><path d="M8 8h8M8 12h8M8 16h5"/></svg>',
  wifi: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M5 12.5a10 10 0 0 1 14 0"/><path d="M8.5 16a5 5 0 0 1 7 0"/><circle cx="12" cy="19" r="1"/></svg>',
  business: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><rect x="3" y="7" width="18" height="13" rx="2"/><path d="M8 7V5a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>',
  contact: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/></svg>',
  email: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><rect x="3" y="5" width="18" height="14" rx="2"/><path d="m3 7 9 6 9-6"/></svg>',
  phone: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M5 4h4l2 5-3 2a11 11 0 0 0 5 5l2-3 5 2v4a2 2 0 0 1-2 2A16 16 0 0 1 3 6a2 2 0 0 1 2-2Z"/></svg>',
  sms: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M21 15a2 2 0 0 1-2 2H8l-4 4V5a2 2 0 0 1 2-2h13a2 2 0 0 1 2 2Z"/></svg>',
  location: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor"><path d="M12 21s7-6 7-11a7 7 0 1 0-14 0c0 5 7 11 7 11Z"/><circle cx="12" cy="10" r="2.5"/></svg>',
};

const TYPES = [
  {
    id: "url", label: "URL", icon: ICONS.url,
    fields: [{ name: "url", label: "URL Content", type: "url", placeholder: "https://example.com" }],
    guide: ["Paste the full web address (we add https:// if you omit it).", "Optionally tick “Shorten the URL first” for a smaller, trackable code.", "Download PNG for screens or SVG for print."],
  },
  {
    id: "text", label: "Text", icon: ICONS.text,
    fields: [{ name: "text", label: "Text Content", type: "textarea", placeholder: "Any text you want to encode…" }],
    guide: ["Enter any plain text — scanners will simply display it.", "Keep it short; long text needs a denser, harder-to-scan code."],
  },
  {
    id: "wifi", label: "WiFi", icon: ICONS.wifi,
    fields: [
      { name: "ssid", label: "Network name (SSID)", type: "text", placeholder: "MyNetwork" },
      { name: "password", label: "Password", type: "text", placeholder: "••••••••" },
      { name: "auth", label: "Security", type: "select", options: [["WPA", "WPA/WPA2"], ["WEP", "WEP"], ["nopass", "None (open)"]] },
      { name: "hidden", label: "Hidden network", type: "checkbox" },
    ],
    guide: ["Enter the exact network name (case-sensitive).", "Choose the security type your router uses.", "Scanning the code joins the network automatically on most phones."],
  },
  {
    id: "business", label: "Business", icon: ICONS.business,
    fields: [
      { name: "company", label: "Company", type: "text", placeholder: "Dojah Inc." },
      { name: "name", label: "Contact name", type: "text", placeholder: "Ada Lovelace" },
      { name: "title", label: "Job title", type: "text", placeholder: "CTO" },
      { name: "phone", label: "Phone", type: "tel", placeholder: "+1 555 123 4567" },
      { name: "email", label: "Email", type: "email", placeholder: "hello@dojah.io" },
      { name: "website", label: "Website", type: "url", placeholder: "https://dojah.io" },
      { name: "address", label: "Address", type: "text", placeholder: "1 Market St, San Francisco" },
    ],
    guide: ["Fill company details for a scannable business card (vCard).", "Scanning saves the contact straight to the phone’s address book."],
  },
  {
    id: "contact", label: "Contact", icon: ICONS.contact,
    fields: [
      { name: "firstName", label: "First name", type: "text", placeholder: "Ada" },
      { name: "lastName", label: "Last name", type: "text", placeholder: "Lovelace" },
      { name: "phone", label: "Phone", type: "tel", placeholder: "+1 555 123 4567" },
      { name: "email", label: "Email", type: "email", placeholder: "ada@example.com" },
      { name: "organization", label: "Organization", type: "text", placeholder: "Analytical Engines" },
      { name: "title", label: "Title", type: "text", placeholder: "Mathematician" },
      { name: "website", label: "Website", type: "url", placeholder: "https://example.com" },
    ],
    guide: ["Provide at least a first or last name.", "Generates a vCard so phones can save the contact in one tap."],
  },
  {
    id: "email", label: "Email", icon: ICONS.email,
    fields: [
      { name: "to", label: "To", type: "email", placeholder: "hello@example.com" },
      { name: "subject", label: "Subject", type: "text", placeholder: "Hello!" },
      { name: "body", label: "Message", type: "textarea", placeholder: "Your message…" },
    ],
    guide: ["Scanning opens a pre-filled email to the recipient.", "Subject and message are optional."],
  },
  {
    id: "phone", label: "Phone", icon: ICONS.phone,
    fields: [{ name: "number", label: "Phone number", type: "tel", placeholder: "+1 555 123 4567" }],
    guide: ["Scanning prompts the phone to dial this number."],
  },
  {
    id: "sms", label: "SMS", icon: ICONS.sms,
    fields: [
      { name: "number", label: "Phone number", type: "tel", placeholder: "+1 555 123 4567" },
      { name: "message", label: "Message", type: "textarea", placeholder: "Pre-filled text message…" },
    ],
    guide: ["Scanning opens a new text to the number with your message pre-filled."],
  },
  {
    id: "location", label: "Location", icon: ICONS.location,
    fields: [
      { name: "latitude", label: "Latitude", type: "text", placeholder: "37.7749" },
      { name: "longitude", label: "Longitude", type: "text", placeholder: "-122.4194" },
    ],
    guide: ["Enter decimal coordinates.", "Scanning opens the point in the phone’s maps app."],
  },
];

// ---------------------------------------------------------------------------
// State
// ---------------------------------------------------------------------------

let currentType = TYPES[0];
let logoBase64 = "";
let currentDataURL = "";
let currentFormat = "png";
const shortenCache = {}; // longURL -> shortURL

// ---------------------------------------------------------------------------
// Type grid + dynamic fields
// ---------------------------------------------------------------------------

function renderTypeGrid() {
  const grid = $("type-grid");
  grid.innerHTML = "";
  for (const t of TYPES) {
    const btn = document.createElement("button");
    btn.type = "button";
    btn.className = "type-btn" + (t.id === currentType.id ? " active" : "");
    btn.innerHTML = `${t.icon}<span>${t.label}</span>`;
    btn.addEventListener("click", () => selectType(t));
    grid.appendChild(btn);
  }
}

function selectType(t) {
  currentType = t;
  renderTypeGrid();
  renderFields();
  renderGuide();
  $("shorten-row").style.display = t.id === "url" ? "flex" : "none";
  render();
}

function renderFields() {
  const wrap = $("fields");
  wrap.innerHTML = "";
  for (const f of currentType.fields) {
    const field = document.createElement("div");
    field.className = "field";

    if (f.type === "checkbox") {
      const lbl = document.createElement("label");
      lbl.className = "toggle";
      lbl.innerHTML = `<input type="checkbox" data-name="${f.name}" /> <span>${f.label}</span>`;
      field.appendChild(lbl);
    } else {
      const lbl = document.createElement("label");
      lbl.textContent = f.label;
      field.appendChild(lbl);

      let input;
      if (f.type === "textarea") {
        input = document.createElement("textarea");
      } else if (f.type === "select") {
        input = document.createElement("select");
        for (const [val, text] of f.options) {
          const opt = document.createElement("option");
          opt.value = val; opt.textContent = text;
          input.appendChild(opt);
        }
      } else {
        input = document.createElement("input");
        input.type = f.type;
      }
      input.dataset.name = f.name;
      if (f.placeholder) input.placeholder = f.placeholder;
      field.appendChild(input);
    }
    wrap.appendChild(field);
  }

  wrap.querySelectorAll("[data-name]").forEach((el) => {
    const ev = el.type === "checkbox" || el.tagName === "SELECT" ? "change" : "input";
    el.addEventListener(ev, render);
  });
}

function collectData() {
  const data = {};
  $("fields").querySelectorAll("[data-name]").forEach((el) => {
    data[el.dataset.name] = el.type === "checkbox" ? String(el.checked) : el.value;
  });
  return data;
}

function hasContent(data) {
  return Object.values(data).some((v) => v && v !== "false" && v.trim && v.trim() !== "");
}

// ---------------------------------------------------------------------------
// Style / request building
// ---------------------------------------------------------------------------

function buildStyle() {
  const style = {
    foreground: $("fg-hex").value,
    background: $("bg-hex").value,
    transparent: $("transparent").checked,
    shape: $("shape").value,
    errorCorrection: $("ec").value,
    size: parseInt($("size").value, 10),
  };
  if ($("gradient-on").checked) {
    style.gradient = {
      angle: parseFloat($("grad-angle").value) || 0,
      stops: [
        { offset: 0, color: $("grad-a").value },
        { offset: 1, color: $("grad-b").value },
      ],
    };
  }
  if (logoBase64) style.logoBase64 = logoBase64;
  return style;
}

let renderTimer = null;
function render() {
  clearTimeout(renderTimer);
  renderTimer = setTimeout(doRender, 300);
}

let inFlight = null;
async function doRender() {
  const data = collectData();
  if (!hasContent(data)) {
    setStatus("Enter content to see preview.", false);
    resetPreview();
    return;
  }

  const body = { format: currentFormat, ...buildStyle() };

  // URL + "shorten first": resolve (and cache) a short link, then encode it.
  if (currentType.id === "url" && $("shorten-first").checked && data.url.trim()) {
    const shortUrl = await ensureShortURL(data.url.trim());
    if (shortUrl) body.content = shortUrl;
    else { body.type = "url"; body.data = data; }
  } else {
    body.type = currentType.id;
    body.data = data;
  }

  setStatus("Generating…", false);
  if (inFlight) inFlight.abort();
  const controller = new AbortController();
  inFlight = controller;

  try {
    const res = await fetch("/api/v1/qr-codes", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
      signal: controller.signal,
    });
    const out = await res.json();
    if (!out.status) {
      setStatus(out.error ? out.error.message : "Could not generate QR code.", true);
      $("generate").disabled = true;
      return;
    }
    currentDataURL = out.entity.image;
    const img = $("preview");
    img.src = currentDataURL;
    img.classList.add("ready");
    $("ph").classList.add("hidden");
    $("generate").disabled = false;
    setStatus(`Ready · ${out.entity.format.toUpperCase()} · ${out.entity.width}px`, false);
  } catch (err) {
    if (err.name === "AbortError") return;
    setStatus("Network error. Please try again.", true);
    $("generate").disabled = true;
  } finally {
    if (inFlight === controller) inFlight = null;
  }
}

function resetPreview() {
  currentDataURL = "";
  $("preview").classList.remove("ready");
  $("ph").classList.remove("hidden");
  $("generate").disabled = true;
}

function setStatus(msg, isError) {
  const s = $("status");
  s.textContent = msg;
  s.classList.toggle("error", !!isError);
}

async function ensureShortURL(longUrl) {
  if (shortenCache[longUrl]) return shortenCache[longUrl];
  try {
    const res = await fetch("/api/v1/short-links", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ url: longUrl }),
    });
    const out = await res.json();
    if (out.status) {
      shortenCache[longUrl] = out.entity.shortUrl;
      saveLink(out.entity);
      return out.entity.shortUrl;
    }
  } catch (_) { /* fall through to raw url */ }
  return "";
}

// Persist created short links to this browser so the My Links page can list them.
function saveLink(entity) {
  try {
    const KEY = "qrunex_links";
    const list = JSON.parse(localStorage.getItem(KEY) || "[]");
    if (!list.some((l) => l.code === entity.code)) {
      list.push({
        code: entity.code,
        url: entity.url,
        shortUrl: entity.shortUrl,
        clicks: entity.clicks || 0,
        createdAt: entity.createdAt,
      });
      localStorage.setItem(KEY, JSON.stringify(list));
    }
  } catch (_) { /* localStorage unavailable */ }
}

// ---------------------------------------------------------------------------
// Download
// ---------------------------------------------------------------------------

function download() {
  if (!currentDataURL) return;
  const a = document.createElement("a");
  a.href = currentDataURL;
  a.download = "qrcode." + currentFormat;
  document.body.appendChild(a);
  a.click();
  a.remove();
}

// ---------------------------------------------------------------------------
// Customize: colors, gradient, presets, logo
// ---------------------------------------------------------------------------

function bindColorPair(picker, hexInput) {
  picker.addEventListener("input", () => { hexInput.value = picker.value.toUpperCase(); clearPreset(); render(); });
  hexInput.addEventListener("input", () => {
    const v = hexInput.value.trim();
    if (/^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$/.test(v)) {
      picker.value = v.startsWith("#") ? v : "#" + v;
      clearPreset(); render();
    }
  });
}

async function loadPresets() {
  try {
    const res = await fetch("/api/v1/qr-presets");
    const out = await res.json();
    if (!out.status) return;
    const box = $("presets");
    for (const p of out.entity) {
      const chip = document.createElement("button");
      chip.type = "button";
      chip.className = "preset-chip";
      chip.dataset.id = p.id;
      chip.title = p.description || p.name;
      const sw = document.createElement("span");
      sw.className = "swatch";
      sw.style.background = p.gradient
        ? `linear-gradient(135deg, ${p.gradient.stops[0].color}, ${p.gradient.stops[p.gradient.stops.length - 1].color})`
        : p.foreground;
      chip.appendChild(sw);
      chip.appendChild(document.createTextNode(p.name));
      chip.addEventListener("click", () => applyPreset(p, chip));
      box.appendChild(chip);
    }
  } catch (_) { /* optional */ }
}

function applyPreset(p, chip) {
  document.querySelectorAll(".preset-chip").forEach((c) => c.classList.remove("active"));
  chip.classList.add("active");
  setColor($("fg"), $("fg-hex"), p.foreground);
  setColor($("bg"), $("bg-hex"), p.background);
  $("shape").value = p.shape || "square";
  $("gradient-on").checked = !!p.gradient;
  if (p.gradient) {
    $("grad-a").value = p.gradient.stops[0].color;
    $("grad-b").value = p.gradient.stops[p.gradient.stops.length - 1].color;
    $("grad-angle").value = p.gradient.angle;
  }
  syncGradient();
  render();
}

function setColor(picker, hexInput, value) { picker.value = value; hexInput.value = value.toUpperCase(); }
function clearPreset() { document.querySelectorAll(".preset-chip").forEach((c) => c.classList.remove("active")); }
function syncGradient() { $("gradient-panel").hidden = !$("gradient-on").checked; }

function onLogoChange() {
  const file = $("logo").files[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = () => { logoBase64 = reader.result; $("logo-clear").hidden = false; render(); };
  reader.readAsDataURL(file);
}
function clearLogo() { logoBase64 = ""; $("logo").value = ""; $("logo-clear").hidden = true; render(); }

// ---------------------------------------------------------------------------
// Tabs
// ---------------------------------------------------------------------------

function switchTab(which) {
  const qr = which === "qr";
  $("tab-qr").classList.toggle("active", qr);
  $("tab-short").classList.toggle("active", !qr);
  $("tab-qr").setAttribute("aria-selected", String(qr));
  $("tab-short").setAttribute("aria-selected", String(!qr));
  $("panel-qr").hidden = !qr;
  $("panel-short").hidden = qr;
}

// ---------------------------------------------------------------------------
// URL Shortener tab
// ---------------------------------------------------------------------------

let lastShort = null;

async function shortenURL() {
  const url = $("long-url").value.trim();
  const hint = $("short-hint");
  if (!url) { hint.textContent = "Enter a URL to shorten."; hint.className = "hint error"; return; }
  hint.textContent = "Shortening…"; hint.className = "hint";
  try {
    const res = await fetch("/api/v1/short-links", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ url }),
    });
    const out = await res.json();
    if (!out.status) { hint.textContent = out.error ? out.error.message : "Failed."; hint.className = "hint error"; return; }
    lastShort = out.entity;
    saveLink(out.entity);
    $("short-url").value = out.entity.shortUrl;
    $("short-clicks").textContent = out.entity.clicks + " clicks";
    $("short-result").hidden = false;
    hint.textContent = "";
  } catch (_) {
    hint.textContent = "Network error."; hint.className = "hint error";
  }
}

async function refreshClicks() {
  if (!lastShort) return;
  try {
    const res = await fetch("/api/v1/short-links/" + encodeURIComponent(lastShort.code));
    const out = await res.json();
    if (out.status) $("short-clicks").textContent = out.entity.clicks + " clicks";
  } catch (_) { /* ignore */ }
}

function copyShort() {
  const v = $("short-url").value;
  if (v && navigator.clipboard) navigator.clipboard.writeText(v);
  const btn = $("copy-short");
  const orig = btn.textContent;
  btn.textContent = "Copied!";
  setTimeout(() => (btn.textContent = orig), 1200);
}

function qrFromShort() {
  if (!lastShort) return;
  switchTab("qr");
  selectType(TYPES[0]);
  const input = $("fields").querySelector('[data-name="url"]');
  if (input) { input.value = lastShort.shortUrl; render(); }
}

// ---------------------------------------------------------------------------
// Guides
// ---------------------------------------------------------------------------

function renderGuide() {
  const el = $("guide-detail");
  const steps = currentType.guide.map((s) => `<li>${s}</li>`).join("");
  el.innerHTML = `<b>${currentType.label} codes</b><ol>${steps}</ol>`;
}

// ---------------------------------------------------------------------------
// Wire up
// ---------------------------------------------------------------------------

function wire() {
  // Deep-link support: /?type=wifi selects a content type on load.
  const params = new URLSearchParams(location.search);
  const typeParam = params.get("type");
  if (typeParam) {
    const found = TYPES.find((t) => t.id === typeParam);
    if (found) currentType = found;
  }

  renderTypeGrid();
  renderFields();
  renderGuide();
  $("shorten-row").style.display = currentType.id === "url" ? "flex" : "none";

  bindColorPair($("fg"), $("fg-hex"));
  bindColorPair($("bg"), $("bg-hex"));

  $("customize-toggle").addEventListener("click", () => {
    const c = $("customize");
    c.hidden = !c.hidden;
    $("customize-toggle").setAttribute("aria-expanded", String(!c.hidden));
  });

  $("shape").addEventListener("change", () => { clearPreset(); render(); });
  $("ec").addEventListener("change", render);
  $("transparent").addEventListener("change", render);
  $("gradient-on").addEventListener("change", () => { clearPreset(); syncGradient(); render(); });
  [$("grad-a"), $("grad-b"), $("grad-angle")].forEach((el) => el.addEventListener("input", () => { clearPreset(); render(); }));

  $("size").addEventListener("input", () => { $("size-val").textContent = $("size").value; render(); });

  $("logo").addEventListener("change", onLogoChange);
  $("logo-clear").addEventListener("click", clearLogo);
  $("shorten-first").addEventListener("change", render);

  document.querySelectorAll('input[name="fmt"]').forEach((r) =>
    r.addEventListener("change", () => { currentFormat = r.value; render(); }));

  $("generate").addEventListener("click", download);

  $("tab-qr").addEventListener("click", () => switchTab("qr"));
  $("tab-short").addEventListener("click", () => switchTab("short"));

  $("shorten-btn").addEventListener("click", shortenURL);
  $("copy-short").addEventListener("click", copyShort);
  $("refresh-clicks").addEventListener("click", refreshClicks);
  $("qr-from-short").addEventListener("click", qrFromShort);

  loadPresets();

  // Deep links: /?tab=short opens the shortener; /?prefill=<url> preloads a URL QR.
  if (params.get("tab") === "short") switchTab("short");
  const prefill = params.get("prefill");
  if (prefill) {
    switchTab("qr");
    const input = $("fields").querySelector('[data-name="url"]');
    if (input) input.value = prefill;
  }

  render();
}

wire();
