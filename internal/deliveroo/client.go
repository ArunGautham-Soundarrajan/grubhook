package deliveroo

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

// stepTimeout bounds how long any single wait (finding an element, a page
// settling) may take before it fails instead of hanging.
const stepTimeout = 20 * time.Second

type Client struct {
	browser *rod.Browser
	page    *rod.Page
}

// New launches a browser and opens the login page. Every browser operation
// is bound to ctx, so cancelling it or hitting its deadline stops the scrape.
func New(ctx context.Context) (*Client, error) {
	// Prefer an installed Chrome/Chromium; rod only downloads its own if none is found.
	l := launcher.New().Context(ctx)
	if bin, ok := launcher.LookPath(); ok {
		l = l.Bin(bin)
	}
	controlURL, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("launching browser: %w", err)
	}

	browser := rod.New().Context(ctx).ControlURL(controlURL).SlowMotion(400 * time.Millisecond)
	if err := browser.Connect(); err != nil {
		return nil, fmt.Errorf("connecting to browser: %w", err)
	}

	page, err := browser.Page(proto.TargetCreateTarget{URL: "https://deliveroo.co.uk/login"})
	if err != nil {
		browser.Close()
		return nil, fmt.Errorf("opening login page: %w", err)
	}

	return &Client{
		browser: browser,
		page:    page,
	}, nil
}

func (c *Client) Close() error {
	// Detach from ctx so the browser still closes after a timeout.
	return c.browser.Context(context.Background()).Close()
}

// element waits up to stepTimeout for selector to appear.
func (c *Client) element(selector string) (*rod.Element, error) {
	el, err := c.page.Timeout(stepTimeout).Element(selector)
	if err != nil {
		return nil, fmt.Errorf("finding %q: %w", selector, err)
	}
	return el.CancelTimeout(), nil
}
