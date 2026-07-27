"use strict";

// Shared QR-type content that powers the Learn grid (/learn) and the per-type
// guide pages (/qr/{slug}). Kept as a plain global so pages can use it without
// a module bundler. The `slug` matches the content-type id used by the API and
// by app.js on the home generator.

window.QR_ICONS = {
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

window.QR_GUIDES = [
  {
    slug: "business",
    label: "Business Card",
    icon: window.QR_ICONS.business,
    popular: true,
    tagline: "Scan to save your details as a contact",
    headerDesc: "Create a vCard QR code that saves your details to a phone's contacts in one scan.",
    whatIs:
      "A business card QR code encodes a vCard — a standard contact format. When someone scans it with their phone camera, they get a one-tap prompt to save your name, phone, email, company, and website straight into their address book. No typing, no transcription errors.",
    benefits: [
      "One scan saves your details — no typing, no transcription errors",
      "Works on printed cards, in email signatures, or on screen",
      "Eco-friendly alternative to paper business cards",
      "Always up-to-date contact information",
    ],
    commonUses: [
      "Networking events and conferences",
      "Email signatures and digital communications",
      "Social media profiles and websites",
      "Trade shows and exhibitions",
      "Business presentations and proposals",
    ],
    sample: {
      type: "business",
      data: {
        company: "Dojah Inc.",
        name: "Ada Lovelace",
        title: "CTO",
        phone: "+1 555 123 4567",
        email: "hello@dojah.io",
        website: "https://dojah.io",
      },
    },
  },
  {
    slug: "wifi",
    label: "WiFi",
    icon: window.QR_ICONS.wifi,
    popular: true,
    tagline: "Join a network without typing the password",
    headerDesc: "Create a QR code guests scan to join your WiFi without typing the password.",
    whatIs:
      "A WiFi QR code stores your network name, password, and security type. Scanning it with a phone camera shows a \"join network\" prompt; one tap connects. Works on iPhone (iOS 11+) and Android with the built-in camera, no app needed.",
    commonUses: [
      "Guest WiFi access in homes and offices",
      "Cafes, restaurants, and hotels",
      "Co-working spaces and event venues",
      "Airbnb and short-term rentals",
    ],
    sample: {
      type: "wifi",
      data: { ssid: "Cafe-Guest", password: "welcome123", auth: "WPA", hidden: "false" },
    },
  },
  {
    slug: "url",
    label: "Website Links",
    icon: window.QR_ICONS.url,
    popular: false,
    tagline: "Open any webpage from a scan",
    headerDesc: "Create a QR code that opens a website, landing page, or any link.",
    whatIs:
      "A URL QR code encodes a web address. Scanning it opens the page in the phone's browser instantly. Shorten the link first for a smaller, denser-free code that scans faster and lets you track clicks.",
    commonUses: [
      "Product packaging and marketing materials",
      "Posters, flyers, and billboards",
      "Restaurant menus and table tents",
      "Landing pages and campaigns",
    ],
    sample: { type: "url", data: { url: "https://qr.moshcore.com.ng" } },
  },
  {
    slug: "email",
    label: "Email Contact",
    icon: window.QR_ICONS.email,
    popular: false,
    tagline: "Open a pre-filled email",
    headerDesc: "Create a QR code that opens a new email with the recipient, subject, and body filled in.",
    whatIs:
      "An email QR code opens the phone's mail app with the recipient — and optionally the subject and message — already filled in. The user just reviews and hits send.",
    commonUses: [
      "Support and contact pages",
      "Feedback and enquiry forms",
      "Event RSVPs",
      "Print ads with a call-to-action",
    ],
    sample: {
      type: "email",
      data: { to: "hello@dojah.io", subject: "Hello!", body: "I scanned your QR code." },
    },
  },
  {
    slug: "phone",
    label: "Phone Numbers",
    icon: window.QR_ICONS.phone,
    popular: false,
    tagline: "Open the dialer, ready to call",
    headerDesc: "Create a QR code that opens the phone dialer pre-loaded with your number.",
    whatIs:
      "A phone QR code encodes a telephone number. Scanning it opens the dialer with the number ready — the user just taps call. Great for reducing friction on print materials.",
    commonUses: [
      "Business cards and storefronts",
      "Service and support hotlines",
      "Real-estate signs",
      "Emergency and helpline posters",
    ],
    sample: { type: "phone", data: { number: "+1 555 123 4567" } },
  },
  {
    slug: "sms",
    label: "Text Messages",
    icon: window.QR_ICONS.sms,
    popular: false,
    tagline: "Open a pre-filled text message",
    headerDesc: "Create a QR code that opens a new SMS to your number with a message pre-filled.",
    whatIs:
      "An SMS QR code opens the messaging app with the recipient number and a pre-written message. Handy for opt-in keywords, competitions, and quick enquiries.",
    commonUses: [
      "SMS marketing opt-ins and keywords",
      "Competitions and voting",
      "Quick enquiries and bookings",
      "Two-way support channels",
    ],
    sample: { type: "sms", data: { number: "+1 555 123 4567", message: "JOIN" } },
  },
  {
    slug: "location",
    label: "Locations",
    icon: window.QR_ICONS.location,
    popular: false,
    tagline: "Open a spot in the maps app",
    headerDesc: "Create a QR code that opens a set of coordinates in the phone's maps app.",
    whatIs:
      "A location QR code encodes latitude and longitude. Scanning it opens the point directly in Google Maps or Apple Maps, ready for directions.",
    commonUses: [
      "Event and venue directions",
      "Store and office locations",
      "Property listings",
      "Meeting points and tours",
    ],
    sample: { type: "location", data: { latitude: "37.7749", longitude: "-122.4194" } },
  },
  {
    slug: "contact",
    label: "Contacts",
    icon: window.QR_ICONS.contact,
    popular: false,
    tagline: "Save name, phone, and email",
    headerDesc: "Create a vCard QR code so anyone can save your contact in one tap.",
    whatIs:
      "A contact QR code encodes a vCard with a person's name, phone, email, organization, and title. Scanning saves the contact to the phone's address book instantly.",
    commonUses: [
      "Personal networking",
      "Team directory cards",
      "Speaker and staff badges",
      "Social profiles",
    ],
    sample: {
      type: "contact",
      data: {
        firstName: "Ada",
        lastName: "Lovelace",
        phone: "+1 555 123 4567",
        email: "ada@example.com",
        organization: "Analytical Engines",
        title: "Mathematician",
      },
    },
  },
  {
    slug: "text",
    label: "Text",
    icon: window.QR_ICONS.text,
    popular: false,
    tagline: "Show a message, works offline",
    headerDesc: "Create a QR code that displays any plain text when scanned.",
    whatIs:
      "A text QR code stores raw text that appears on screen when scanned — no internet needed. Keep it short so the code stays easy to scan.",
    commonUses: [
      "Serial numbers and asset tags",
      "Short instructions and notes",
      "Product details on packaging",
      "Offline messages",
    ],
    sample: { type: "text", data: { text: "Hello from Qrunex!" } },
  },
];

window.QR_GUIDE_BY_SLUG = window.QR_GUIDES.reduce((m, g) => {
  m[g.slug] = g;
  return m;
}, {});
