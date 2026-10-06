package deliveroo

import (
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
)

func (c *Client) Login(email, password string) {
	handleCookies(c.page)

	humanType(c.page, c.page.MustElement("#inline-email-address"), email)
	pause(500, 1500)
	c.page.MustElement("#continue-with-email").MustClick()

	humanType(c.page, c.page.MustElement("#login-password"), password)
	pause(500, 1500)
	c.page.MustElement("#email-password-submit").MustClick()

	time.Sleep(time.Second * 5)
}

func handleCookies(p *rod.Page) {
	p.MustElement("#__next > div.CustomCookieBanner-39382a7d9c5bd12f > div > div.CustomCookieBanner-6bcaa32076f4e46c > span:nth-child(2) > button").MustClick()
}

func pause(minMs, maxMs int) {
	time.Sleep(time.Duration(minMs+rand.Intn(maxMs-minMs)) * time.Millisecond)
}

func humanType(p *rod.Page, el *rod.Element, text string) {
	el.MustClick() // focus first
	for _, r := range text {
		p.Keyboard.MustType(input.Key(r))
		pause(80, 220)
	}
}
