// Package qrcontent builds the encoded payload string for the various QR code
// content types (URL, text, WiFi, contact, etc.) from structured input fields.
package qrcontent

import (
	"fmt"
	"regexp"
	"strings"
)

// BuildError is a user-facing content-validation error with a stable code.
type BuildError struct {
	Code    string
	Message string
}

func (e *BuildError) Error() string { return e.Message }

func newErr(code, message string) *BuildError { return &BuildError{Code: code, Message: message} }

var schemePattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.\-]*:`)

// Types lists the supported content type identifiers.
var Types = []string{"url", "text", "wifi", "business", "contact", "email", "phone", "sms", "location"}

// Build returns the encoded QR payload for the given content type and fields.
func Build(kind string, data map[string]string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "url":
		return buildURL(data)
	case "text":
		return buildText(data)
	case "wifi":
		return buildWiFi(data)
	case "email":
		return buildEmail(data)
	case "phone":
		return buildPhone(data)
	case "sms":
		return buildSMS(data)
	case "location":
		return buildLocation(data)
	case "contact":
		return buildContact(data)
	case "business":
		return buildBusiness(data)
	default:
		return "", newErr("UNKNOWN_CONTENT_TYPE", "Unsupported content type")
	}
}

func buildURL(d map[string]string) (string, error) {
	raw := strings.TrimSpace(d["url"])
	if raw == "" {
		return "", newErr("MISSING_FIELD", "A URL is required")
	}
	if !schemePattern.MatchString(raw) {
		raw = "https://" + raw
	}
	return raw, nil
}

func buildText(d map[string]string) (string, error) {
	text := d["text"]
	if strings.TrimSpace(text) == "" {
		return "", newErr("MISSING_FIELD", "Text is required")
	}
	return text, nil
}

// buildWiFi produces the standard WIFI:...; join string understood by phone
// camera apps. Auth defaults to WPA; set auth=nopass for an open network.
func buildWiFi(d map[string]string) (string, error) {
	ssid := strings.TrimSpace(d["ssid"])
	if ssid == "" {
		return "", newErr("MISSING_FIELD", "Network name (SSID) is required")
	}
	auth := strings.ToUpper(strings.TrimSpace(d["auth"]))
	switch auth {
	case "", "WPA", "WPA2":
		auth = "WPA"
	case "WEP":
		// keep
	case "NOPASS":
		// keep
	default:
		return "", newErr("INVALID_FIELD", "WiFi security must be WPA, WEP, or nopass")
	}

	hidden := "false"
	if isTrue(d["hidden"]) {
		hidden = "true"
	}

	var b strings.Builder
	b.WriteString("WIFI:T:")
	b.WriteString(auth)
	b.WriteString(";S:")
	b.WriteString(escapeWiFi(ssid))
	if auth != "NOPASS" {
		b.WriteString(";P:")
		b.WriteString(escapeWiFi(d["password"]))
	}
	b.WriteString(";H:")
	b.WriteString(hidden)
	b.WriteString(";;")
	return b.String(), nil
}

func buildEmail(d map[string]string) (string, error) {
	to := strings.TrimSpace(d["to"])
	if to == "" {
		return "", newErr("MISSING_FIELD", "Recipient email is required")
	}
	params := []string{}
	if s := strings.TrimSpace(d["subject"]); s != "" {
		params = append(params, "subject="+queryEscape(s))
	}
	if body := d["body"]; body != "" {
		params = append(params, "body="+queryEscape(body))
	}
	out := "mailto:" + to
	if len(params) > 0 {
		out += "?" + strings.Join(params, "&")
	}
	return out, nil
}

func buildPhone(d map[string]string) (string, error) {
	num := strings.TrimSpace(d["number"])
	if num == "" {
		return "", newErr("MISSING_FIELD", "Phone number is required")
	}
	return "tel:" + sanitizePhone(num), nil
}

func buildSMS(d map[string]string) (string, error) {
	num := strings.TrimSpace(d["number"])
	if num == "" {
		return "", newErr("MISSING_FIELD", "Phone number is required")
	}
	msg := d["message"]
	if msg == "" {
		return "SMSTO:" + sanitizePhone(num), nil
	}
	return "SMSTO:" + sanitizePhone(num) + ":" + msg, nil
}

func buildLocation(d map[string]string) (string, error) {
	lat := strings.TrimSpace(d["latitude"])
	lng := strings.TrimSpace(d["longitude"])
	if lat == "" || lng == "" {
		return "", newErr("MISSING_FIELD", "Latitude and longitude are required")
	}
	if !isNumeric(lat) || !isNumeric(lng) {
		return "", newErr("INVALID_FIELD", "Latitude and longitude must be numeric")
	}
	return fmt.Sprintf("geo:%s,%s", lat, lng), nil
}

func buildContact(d map[string]string) (string, error) {
	first := strings.TrimSpace(d["firstName"])
	last := strings.TrimSpace(d["lastName"])
	if first == "" && last == "" {
		return "", newErr("MISSING_FIELD", "A first or last name is required")
	}
	return vCard(vCardFields{
		First: first, Last: last,
		Phone: d["phone"], Email: d["email"],
		Org: d["organization"], Title: d["title"], URL: d["website"],
	}), nil
}

func buildBusiness(d map[string]string) (string, error) {
	company := strings.TrimSpace(d["company"])
	name := strings.TrimSpace(d["name"])
	if company == "" && name == "" {
		return "", newErr("MISSING_FIELD", "A company or contact name is required")
	}
	// Treat the full name as a single "last name" component for simplicity.
	return vCard(vCardFields{
		Last:    name,
		Org:     company,
		Title:   d["title"],
		Phone:   d["phone"],
		Email:   d["email"],
		URL:     d["website"],
		Address: d["address"],
	}), nil
}

type vCardFields struct {
	First, Last       string
	Org, Title        string
	Phone, Email, URL string
	Address           string
}

func vCard(f vCardFields) string {
	var b strings.Builder
	b.WriteString("BEGIN:VCARD\nVERSION:3.0\n")
	fmt.Fprintf(&b, "N:%s;%s;;;\n", escapeVCard(f.Last), escapeVCard(f.First))
	fmt.Fprintf(&b, "FN:%s\n", escapeVCard(strings.TrimSpace(f.First+" "+f.Last)))
	if f.Org != "" {
		fmt.Fprintf(&b, "ORG:%s\n", escapeVCard(f.Org))
	}
	if f.Title != "" {
		fmt.Fprintf(&b, "TITLE:%s\n", escapeVCard(f.Title))
	}
	if f.Phone != "" {
		fmt.Fprintf(&b, "TEL;TYPE=CELL:%s\n", escapeVCard(f.Phone))
	}
	if f.Email != "" {
		fmt.Fprintf(&b, "EMAIL:%s\n", escapeVCard(f.Email))
	}
	if f.URL != "" {
		fmt.Fprintf(&b, "URL:%s\n", escapeVCard(f.URL))
	}
	if f.Address != "" {
		fmt.Fprintf(&b, "ADR;TYPE=WORK:;;%s;;;;\n", escapeVCard(f.Address))
	}
	b.WriteString("END:VCARD")
	return b.String()
}

// --- helpers ---------------------------------------------------------------

func escapeWiFi(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `;`, `\;`, `,`, `\,`, `:`, `\:`, `"`, `\"`)
	return r.Replace(s)
}

func escapeVCard(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `;`, `\;`, `,`, `\,`, "\n", `\n`)
	return r.Replace(s)
}

func queryEscape(s string) string {
	// Encode a minimal set so mailto params stay valid without over-escaping.
	r := strings.NewReplacer(" ", "%20", "&", "%26", "?", "%3F", "\n", "%0A", "#", "%23")
	return r.Replace(s)
}

func sanitizePhone(s string) string {
	var b strings.Builder
	for _, c := range s {
		if c == '+' || (c >= '0' && c <= '9') {
			b.WriteRune(c)
		}
	}
	return b.String()
}

func isTrue(s string) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes", "on":
		return true
	}
	return false
}

func isNumeric(s string) bool {
	dot := false
	for i, c := range s {
		switch {
		case c >= '0' && c <= '9':
		case c == '-' && i == 0:
		case c == '.' && !dot:
			dot = true
		default:
			return false
		}
	}
	return true
}
