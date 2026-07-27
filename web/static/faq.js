"use strict";

(function () {
  const FAQS = [
    // Getting Started
    { c: "Getting Started", q: "What is Qrunex and what can I do with it?", a: "Qrunex is a free QR code generator and URL shortener. You can create scannable QR codes for 9 content types (websites, WiFi, business cards, contacts, email, phone, SMS, text, and locations), style them with colors, gradients, dots, and a center logo, then download as PNG or SVG. You can also shorten long links and track how many times they're scanned or clicked." },
    { c: "Getting Started", q: "Do I really not need to create an account?", a: "Correct — there's no signup and no login. Everything runs in your browser and our stateless API. Links you create are remembered locally on your own device under \"My Links\"." },
    { c: "Getting Started", q: "How much does Qrunex cost?", a: "It's free. There are no watermarks on your QR codes and no limits on downloads for normal use." },
    { c: "Getting Started", q: "I'm new to QR codes. Where should I start?", a: "Head to the home page, pick a content type, fill in the form, and watch the live preview update. When you're happy, download it. For background on each type, browse the Learn QR guides." },

    // QR Code Types
    { c: "QR Code Types", q: "What types of QR codes can I create?", a: "URL/website links, plain text, WiFi network access, business cards (vCard), personal contacts (vCard), pre-filled email, phone dial, pre-filled SMS, and map locations." },
    { c: "QR Code Types", q: "How do business card QR codes work?", a: "They encode a vCard containing your name, title, company, phone, email, and website. Scanning shows a one-tap prompt to save you straight into the phone's contacts." },
    { c: "QR Code Types", q: "How do WiFi QR codes work?", a: "They store your network name, password, and security type. Scanning with the built-in camera on iOS 11+ or modern Android shows a \"join network\" prompt — no password typing." },
    { c: "QR Code Types", q: "Can I create a QR code that opens a map location?", a: "Yes. Enter decimal latitude and longitude; scanning opens the point in the phone's maps app, ready for directions." },

    // Business Use Cases
    { c: "Business Use Cases", q: "Can I use Qrunex QR codes on printed materials?", a: "Absolutely. Download the SVG format for print — it's a vector, so it stays crisp at any size, from a business card to a billboard." },
    { c: "Business Use Cases", q: "Can I put my logo in the middle of the QR code?", a: "Yes. Open Customize and upload a PNG or JPEG logo. Use a higher error-correction level (Q or H) so the code still scans reliably with the logo covering part of it." },
    { c: "Business Use Cases", q: "What's the best QR code for a restaurant menu?", a: "A URL QR code pointing to your online menu. Shorten the URL first for a smaller, faster-scanning code, and you'll also get scan counts." },
    { c: "Business Use Cases", q: "Can multiple people scan the same code?", a: "Yes — there's no limit on scans. The same printed code works for everyone, forever, as long as the destination stays valid." },

    // Technical
    { c: "Technical", q: "What's the difference between PNG and SVG downloads?", a: "PNG is a raster image ideal for screens and quick sharing. SVG is a vector that scales infinitely without blurring — best for print and large formats." },
    { c: "Technical", q: "What is error correction and which level should I pick?", a: "Error correction lets a code still scan when partly damaged or covered. Low (7%) keeps codes simplest; High (30%) is most robust and required if you add a center logo. Medium is a good default." },
    { c: "Technical", q: "What image sizes can I generate?", a: "Anywhere from 128px up to 1000px for PNG. SVG is resolution-independent, so size is just the base canvas." },
    { c: "Technical", q: "Is there an API I can call directly?", a: "Yes. POST /api/v1/qr-codes creates a QR code and POST /api/v1/short-links creates a short link. Responses follow a consistent { entity, error, status } envelope." },
    { c: "Technical", q: "Why won't my QR code scan?", a: "Common causes: too little contrast between foreground and background, too much data packed in (long text/URLs), a logo that's too large, or printing it too small. Increase contrast, shorten the content, raise error correction, and print at least 2cm × 2cm." },

    // Privacy & Security
    { c: "Privacy & Security", q: "Is my QR code content stored on your servers?", a: "No. QR content is encoded on the fly and never logged. Request logging records only method, path, status, and timing — never the content." },
    { c: "Privacy & Security", q: "Are the QR codes tracked?", a: "Plain QR codes are not tracked at all. Only shortened URLs count clicks, and that's just an anonymous counter — no personal data." },
    { c: "Privacy & Security", q: "Is it safe to put a WiFi password in a QR code?", a: "The password is encoded in the code itself, so treat the image like the password — anyone who can scan it can join. Share it only where you'd share the password, and prefer a guest network." },
    { c: "Privacy & Security", q: "Do you use HTTPS?", a: "Yes. The public site is served over HTTPS, so content in transit is encrypted." },

    // URL Shortening
    { c: "URL Shortening", q: "How does URL shortening work?", a: "Paste a long URL and Qrunex returns a short link like /s/abc123. Anyone visiting it is redirected to your original URL, and the click is counted." },
    { c: "URL Shortening", q: "Do short links expire?", a: "Links persist for the life of the service instance. For guaranteed long-term persistence, database-backed storage can be enabled." },
    { c: "URL Shortening", q: "Can I see how many clicks my link got?", a: "Yes. Open My Links to see click counts for every link you've created on this device, or use the Refresh button on a link after shortening." },
    { c: "URL Shortening", q: "Where are my shortened links saved?", a: "In your browser's local storage on this device, under My Links. They aren't tied to an account, so clearing your browser data removes the list (the links themselves keep working)." },

    // Troubleshooting
    { c: "Troubleshooting", q: "My Links page is empty — where did my links go?", a: "My Links reads from this browser's local storage. If you cleared browsing data, switched browsers, or used private mode, the local list is gone even though the short links still redirect." },
    { c: "Troubleshooting", q: "The logo upload isn't showing in the preview.", a: "Make sure the file is a PNG or JPEG under the size limit. Very large images are rejected; try one around 512px square with a transparent or white background." },
  ];

  const CATS = ["All", ...Array.from(new Set(FAQS.map((f) => f.c)))];
  let activeCat = "All";
  let query = "";

  const listEl = document.getElementById("faq-list");
  const catsEl = document.getElementById("faq-cats");
  const countEl = document.getElementById("faq-count");
  const searchEl = document.getElementById("faq-search");

  catsEl.innerHTML = CATS.map(
    (c) => `<button class="faq-chip${c === "All" ? " active" : ""}" data-cat="${c}">${c}</button>`
  ).join("");
  catsEl.addEventListener("click", (e) => {
    const btn = e.target.closest(".faq-chip");
    if (!btn) return;
    activeCat = btn.dataset.cat;
    document.querySelectorAll(".faq-chip").forEach((c) => c.classList.toggle("active", c === btn));
    render();
  });

  searchEl.addEventListener("input", () => {
    query = searchEl.value.trim().toLowerCase();
    render();
  });

  function filtered() {
    return FAQS.filter((f) => {
      if (activeCat !== "All" && f.c !== activeCat) return false;
      if (query && !(f.q.toLowerCase().includes(query) || f.a.toLowerCase().includes(query))) return false;
      return true;
    });
  }

  function render() {
    const items = filtered();
    countEl.textContent = `${items.length} question${items.length === 1 ? "" : "s"} found`;
    if (!items.length) {
      listEl.innerHTML = `<p class="faq-empty">No questions match your search.</p>`;
      return;
    }
    listEl.innerHTML = items
      .map(
        (f) => `
        <details class="faq-item">
          <summary>
            <div class="faq-q">
              <b>${escapeHtml(f.q)}</b>
              <span class="faq-cat">${escapeHtml(f.c)}</span>
            </div>
            <span class="faq-chev">⌄</span>
          </summary>
          <div class="faq-a">${escapeHtml(f.a)}</div>
        </details>`
      )
      .join("");
  }

  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
  }

  render();
})();
