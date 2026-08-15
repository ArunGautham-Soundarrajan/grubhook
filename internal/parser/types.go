package parser

import "time"

type OrderDetails struct {
	Partner string
	Store   string
	Total   float64
	Date    time.Time
}
