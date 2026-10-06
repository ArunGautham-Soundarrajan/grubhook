package deliveroo

import (
	"time"

	"github.com/go-rod/rod"
)

type Client struct {
	browser *rod.Browser
	page    *rod.Page
}

func New() *Client {
	browser := rod.New().SlowMotion(800 * time.Millisecond).MustConnect()
	page := browser.MustPage("https://deliveroo.co.uk/login")

	return &Client{
		browser: browser,
		page:    page,
	}
}

func (c *Client) Close() error {
	return c.browser.Close()
}
