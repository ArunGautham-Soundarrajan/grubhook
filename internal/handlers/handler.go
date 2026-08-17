package handlers

import (
	"github.com/ArunGautham-Soundarrajan/grubhook/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/gmail/v1"
)

type Handler struct {
	Queries  *db.Queries
	GmailSrv *gmail.Service
	Pool     *pgxpool.Pool
}
