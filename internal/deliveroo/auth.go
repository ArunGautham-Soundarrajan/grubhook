package deliveroo

import (
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
)

const (
	cookieBannerButton = "#__next > div.CustomCookieBanner-39382a7d9c5bd12f > div > div.CustomCookieBanner-6bcaa32076f4e46c > span:nth-child(2) > button"
	cookieBannerWait   = 5 * time.Second
	loginWait          = 30 * time.Second
	passwordInput      = "#login-password"
	enterPasswordText  = "Enter password"
)

func (c *Client) Login(email, password string) error {
	// With a valid saved session, Deliveroo redirects away from /login.
	if err := c.page.Timeout(stepTimeout).WaitStable(time.Second); err != nil {
		return fmt.Errorf("login: waiting for login page: %w", err)
	}
	if !c.onLoginPage() {
		slog.Info("login: already logged in from saved cookies", "url", c.currentURL())
		return nil
	}

	c.acceptCookies()

	emailField, err := c.element("#inline-email-address")
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	if err := humanType(c.page, emailField, email); err != nil {
		return fmt.Errorf("login: typing email: %w", err)
	}
	pause(500, 1500)
	if err := c.click("#continue-with-email"); err != nil {
		return fmt.Errorf("login: %w", err)
	}
	slog.Info("login: email submitted")

	passwordField, err := c.passwordField()
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	if err := humanType(c.page, passwordField, password); err != nil {
		return fmt.Errorf("login: typing password: %w", err)
	}
	pause(500, 1500)
	if err := c.click("#email-password-submit"); err != nil {
		return fmt.Errorf("login: %w", err)
	}
	slog.Info("login: password submitted")

	// A successful login redirects away from /login.
	err = c.page.Timeout(loginWait).Wait(rod.Eval(`() => !location.pathname.startsWith("/login")`))
	if err != nil {
		return fmt.Errorf("login: still on login page after %s (wrong credentials, captcha or 2FA?): %w", loginWait, err)
	}
	slog.Info("login: succeeded", "url", c.currentURL())
	return nil
}

// acceptCookies dismisses the cookie banner if it shows up. A missing banner
// is not an error: it may already be accepted, or the page may have changed.
func (c *Client) acceptCookies() {
	el, err := c.page.Timeout(cookieBannerWait).Element(cookieBannerButton)
	if err != nil {
		slog.Info("cookie banner not found, continuing")
		return
	}
	if err := el.CancelTimeout().Click(proto.InputMouseButtonLeft, 1); err != nil {
		slog.Warn("clicking cookie banner", "err", err)
	}
}

// passwordField returns the password input. Deliveroo sometimes shows a
// one-time code page first; its "Enter password" button switches to the
// password form.
func (c *Client) passwordField() (*rod.Element, error) {
	el, err := c.page.Timeout(stepTimeout).Race().
		Element(passwordInput).
		ElementR("button, a", enterPasswordText).
		Do()
	if err != nil {
		return nil, fmt.Errorf("finding password field or %q button: %w", enterPasswordText, err)
	}
	el = el.CancelTimeout()

	if isInput, err := el.Matches(passwordInput); err != nil || isInput {
		return el, err
	}

	slog.Info("login: one-time code page shown, switching to password")
	if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return nil, fmt.Errorf("clicking %q: %w", enterPasswordText, err)
	}
	return c.element(passwordInput)
}

func (c *Client) onLoginPage() bool {
	res, err := c.page.Eval(`() => location.pathname.startsWith("/login")`)
	return err != nil || res.Value.Bool()
}

// currentURL returns the page URL for logging, or "" if it can't be read.
func (c *Client) currentURL() string {
	info, err := c.page.Info()
	if err != nil {
		return ""
	}
	return info.URL
}

func (c *Client) click(selector string) error {
	el, err := c.element(selector)
	if err != nil {
		return err
	}
	if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("clicking %q: %w", selector, err)
	}
	return nil
}

func pause(minMs, maxMs int) {
	time.Sleep(time.Duration(minMs+rand.Intn(maxMs-minMs)) * time.Millisecond)
}

func humanType(p *rod.Page, el *rod.Element, text string) error {
	if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil { // focus first
		return err
	}
	for _, r := range text {
		if err := p.Keyboard.Type(input.Key(r)); err != nil {
			return err
		}
		pause(80, 220)
	}
	return nil
}
