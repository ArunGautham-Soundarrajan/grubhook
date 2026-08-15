package parser

import (
	"io"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func deliverooHTMLParser(r io.Reader) (OrderDetails, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return OrderDetails{}, err
	}
	// Find the store name
	store := strings.TrimSuffix(strings.TrimSpace(doc.Find("h2").First().Text()), " has your order!")
	totalStr := strings.TrimSpace(doc.Find("p.total").Eq(1).Text())
	totalStr = strings.TrimPrefix(totalStr, "£")
	totalStr = strings.TrimSpace(totalStr)
	total, err := strconv.ParseFloat(totalStr, 64)
	if err != nil {
		return OrderDetails{}, err
	}
	return OrderDetails{
		store: store,
		total: total,
	}, nil
}
