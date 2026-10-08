package deliveroo

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// diagnoseTimeout bounds each diagnostic read. They run after a failure,
// often after the scrape context has expired, so they use their own deadline.
const diagnoseTimeout = 10 * time.Second

// bodyTextLimit caps how much visible page text goes into the logs.
const bodyTextLimit = 600

// Diagnostics describes the page at the moment something went wrong.
type Diagnostics struct {
	URL   string
	Title string
	// Messages holds headings, alert and error text, and any captcha or
	// challenge iframes: whatever the page is most likely telling us.
	Messages []string
	// BodyText is the start of the page's visible text.
	BodyText string
}

// pageMessagesJS collects the page's headings, alerts, error text and
// challenge iframes, plus its full visible text.
const pageMessagesJS = `() => {
	const selector = [
		'h1', 'h2', '[role="alert"]', '[aria-live]',
		'[class*="error" i]', '[data-testid*="error" i]',
		'iframe[src*="captcha" i]', 'iframe[title*="captcha" i]', 'iframe[title*="challenge" i]',
		'iframe[src*="arkoselabs"]', 'iframe[src*="funcaptcha"]',
	].join(',');
	const seen = new Set();
	const messages = [];
	for (const el of document.querySelectorAll(selector)) {
		const text = el.tagName === 'IFRAME'
			? 'iframe: ' + (el.title || el.src)
			: (el.innerText || '').trim().replace(/\s+/g, ' ');
		if (!text || seen.has(text)) continue;
		seen.add(text);
		messages.push(text.slice(0, 200));
		if (messages.length >= 15) break;
	}
	const body = document.body ? document.body.innerText : '';
	return { messages, body: body.trim().replace(/\s+/g, ' ') };
}`

var emailPattern = regexp.MustCompile(`[\w.+-]+@[\w-]+(\.[\w-]+)+`)

// Diagnose reads the current page's URL, title and visible messages.
// Email addresses are redacted, since the result is logged.
func (c *Client) Diagnose() (Diagnostics, error) {
	p := c.page.Context(context.Background()).Timeout(diagnoseTimeout)

	info, err := p.Info()
	if err != nil {
		return Diagnostics{}, fmt.Errorf("reading page info: %w", err)
	}
	d := Diagnostics{URL: info.URL, Title: info.Title}

	res, err := p.Eval(pageMessagesJS)
	if err != nil {
		return d, fmt.Errorf("reading page text: %w", err)
	}
	var text struct {
		Messages []string `json:"messages"`
		Body     string   `json:"body"`
	}
	if err := res.Value.Unmarshal(&text); err != nil {
		return d, fmt.Errorf("decoding page text: %w", err)
	}

	for _, m := range text.Messages {
		d.Messages = append(d.Messages, redactEmails(m))
	}
	d.BodyText = redactEmails(truncate(text.Body, bodyTextLimit))
	return d, nil
}

// SaveScreenshot writes the current page to path, for debugging failures.
func (c *Client) SaveScreenshot(path string) error {
	img, err := c.page.Context(context.Background()).Timeout(diagnoseTimeout).Screenshot(true, nil)
	if err != nil {
		return err
	}
	return os.WriteFile(path, img, 0o644)
}

// SaveHTML writes the current page's HTML to path, for debugging failures.
func (c *Client) SaveHTML(path string) error {
	html, err := c.page.Context(context.Background()).Timeout(diagnoseTimeout).HTML()
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(html), 0o644)
}

func redactEmails(s string) string {
	return emailPattern.ReplaceAllString(s, "<email>")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	// Back off to a rune boundary so the result stays valid UTF-8.
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return strings.TrimSpace(s[:n]) + "…"
}
