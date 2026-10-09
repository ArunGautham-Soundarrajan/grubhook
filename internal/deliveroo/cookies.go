package deliveroo

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-rod/rod/lib/proto"
)

// exportedCookie is one cookie as written by browser cookie-export extensions
// (the chrome.cookies API shape, e.g. "Get cookies.txt LOCALLY" JSON export).
type exportedCookie struct {
	Name           string  `json:"name"`
	Value          string  `json:"value"`
	Domain         string  `json:"domain"`
	Path           string  `json:"path"`
	HostOnly       bool    `json:"hostOnly"`
	Secure         bool    `json:"secure"`
	HTTPOnly       bool    `json:"httpOnly"`
	SameSite       string  `json:"sameSite"`
	Session        bool    `json:"session"`
	ExpirationDate float64 `json:"expirationDate"`
}

// LoadCookies reads a JSON cookie export so a saved login session can be
// reused instead of logging in, which Deliveroo may challenge with 2FA.
func LoadCookies(path string) ([]*proto.NetworkCookieParam, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading cookies: %w", err)
	}
	var exported []exportedCookie
	if err := json.Unmarshal(data, &exported); err != nil {
		return nil, fmt.Errorf("parsing cookies %s: %w", path, err)
	}

	params := make([]*proto.NetworkCookieParam, 0, len(exported))
	for _, c := range exported {
		p := &proto.NetworkCookieParam{
			Name:     c.Name,
			Value:    c.Value,
			Path:     c.Path,
			Secure:   c.Secure,
			HTTPOnly: c.HTTPOnly,
			SameSite: sameSite(c.SameSite),
		}
		if c.HostOnly {
			// Setting Domain would make it a domain cookie; a URL keeps it host-only.
			p.URL = "https://" + strings.TrimPrefix(c.Domain, ".") + c.Path
		} else {
			p.Domain = c.Domain
		}
		if !c.Session && c.ExpirationDate > 0 {
			p.Expires = proto.TimeSinceEpoch(c.ExpirationDate)
		}
		params = append(params, p)
	}
	return params, nil
}

func sameSite(s string) proto.NetworkCookieSameSite {
	switch strings.ToLower(s) {
	case "strict":
		return proto.NetworkCookieSameSiteStrict
	case "lax":
		return proto.NetworkCookieSameSiteLax
	case "no_restriction", "none":
		return proto.NetworkCookieSameSiteNone
	default:
		return ""
	}
}
