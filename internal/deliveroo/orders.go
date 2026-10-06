package deliveroo

import (
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Order struct {
	ID           string    `json:"id"`
	Submitted    time.Time `json:"legacySubmittedDate"`
	RestaurantID string    `json:"restaurantId"`
	Restaurant   string    `json:"restaurantName"`
	Balance      string    `json:"balance"`
	Status       string    `json:"status"`
}

func (c *Client) FetchOrders() ([]Order, error) {
	c.page.MustNavigate("https://deliveroo.co.uk/orders")
	c.page.MustWaitStable()

	html := c.page.MustHTML()
	return extractOrders(strings.NewReader(html))
}

func extractOrders(r io.Reader) ([]Order, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var orders []Order

	doc.Find("script").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		text := s.Text()

		_, after, found := strings.Cut(text, `"orders":`)
		if !found {
			return true
		}
		dec := json.NewDecoder(strings.NewReader(after))
		if err := dec.Decode(&orders); err != nil {
			return true // wrong match, keep looking
		}

		return false
	})

	return orders, nil
}
