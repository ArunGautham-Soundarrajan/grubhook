package parser

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime/quotedprintable"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
)

func GetLast30dMessageIDs(ctx context.Context, srv *gmail.Service, label string) ([]*gmail.Message, error) {
	query := "label:" + label + " newer_than:30d"
	r, err := srv.Users.Messages.List("me").Q(query).Do()
	if err != nil {
		return nil, err
	}

	return r.Messages, nil
}

func findHTMLPart(part *gmail.MessagePart) *gmail.MessagePart {
	if part.MimeType == "text/html" {
		return part
	}
	for _, p := range part.Parts {
		if found := findHTMLPart(p); found != nil {
			return found
		}
	}
	return nil
}

func GetOrderTotalFromMsg(ctx context.Context, srv *gmail.Service, id string) (OrderDetails, error) {
	msg, err := srv.Users.Messages.Get("me", id).Format("full").Do()
	if err != nil {
		return OrderDetails{}, fmt.Errorf("fetching message: %w", err)
	}

	header := msg.Payload.Headers
	var partner string
	for _, h := range header {
		if h.Name == "From" {
			partner = h.Value
			break
		}
	}

	if partner == "" {
		return OrderDetails{}, fmt.Errorf("no From header found in message %s", id)
	}

	htmlPart := findHTMLPart(msg.Payload)
	if htmlPart == nil {
		return OrderDetails{}, fmt.Errorf("no text/html part found in message %s", id)
	}

	body, err := base64.URLEncoding.DecodeString(htmlPart.Body.Data)
	if err != nil {
		return OrderDetails{}, fmt.Errorf("base64 decoding body: %w", err)
	}

	decoded, err := io.ReadAll(quotedprintable.NewReader(bytes.NewReader(body)))
	if err != nil {
		return OrderDetails{}, fmt.Errorf("decoding quoted-printable: %w", err)
	}

	var data OrderDetails
	switch {
	case strings.Contains(partner, "deliveroo"):
		data, err = deliverooHTMLParser(bytes.NewReader(decoded))
		if err != nil {
			return OrderDetails{}, fmt.Errorf("parsing deliveroo html: %w", err)
		}
	case strings.Contains(partner, "uber"):

		// os.WriteFile("uberEats.html", decoded, 0o644)
		// data, err = uberEatsHTMLParser(bytes.NewReader(decoded))
		// data.partner = partner
		// if err != nil {
		// 	return OrderDetails{}, fmt.Errorf("parsing uber eats html: %w", err)
		// }
	default:
		return OrderDetails{}, fmt.Errorf("unrecognized sender: %q", partner)
	}

	data.Date = time.UnixMilli(msg.InternalDate)
	data.Partner = partner

	return data, nil
}
