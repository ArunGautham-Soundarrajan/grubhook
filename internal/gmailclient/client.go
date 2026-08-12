package gmailclient

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func New(ctx context.Context, token oauth2.Token, config oauth2.Config) (*gmail.Service, error) {
	client := config.Client(ctx, &token)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("creating gmail service: %w", err)
	}

	return srv, nil
}
