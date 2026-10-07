package deliveroo

import (
	"encoding/json"
	"errors"
	"fmt"
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
	if err := c.page.Navigate("https://deliveroo.co.uk/orders"); err != nil {
		return nil, fmt.Errorf("opening orders page: %w", err)
	}
	if err := c.page.Timeout(stepTimeout).WaitStable(time.Second); err != nil {
		return nil, fmt.Errorf("waiting for orders page: %w", err)
	}

	html, err := c.page.HTML()
	if err != nil {
		return nil, fmt.Errorf("reading orders page: %w", err)
	}
	return extractOrders(strings.NewReader(html))
}

func extractOrders(r io.Reader) ([]Order, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var orders []Order
	found := false

	doc.Find("script").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		text := s.Text()

		_, after, ok := strings.Cut(text, `"orders":`)
		if !ok {
			return true
		}
		dec := json.NewDecoder(strings.NewReader(after))
		if err := dec.Decode(&orders); err != nil {
			return true // wrong match, keep looking
		}

		found = true
		return false
	})

	if !found {
		return nil, errors.New(`no "orders" data on page (not logged in, or page layout changed?)`)
	}
	return orders, nil
}
