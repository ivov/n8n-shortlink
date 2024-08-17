package services

import (
	"fmt"

	"github.com/ivov/n8n-shortlink/internal/db/entities"
	"github.com/ivov/n8n-shortlink/internal/log"
	"github.com/jmoiron/sqlx"
)

// VisitService manages visits.
type VisitService struct {
	DB     *sqlx.DB
	Logger *log.Logger
}

// SaveVisit writes a visit to the DB.
func (vs *VisitService) SaveVisit(slug string, kind string, referer string, userAgent string) error {
	visit := entities.Visit{
		Slug:      slug,
		Referer:   referer,
		UserAgent: userAgent,
	}

	query := `
		INSERT INTO visits (slug, referer, user_agent)
		VALUES (:slug, :referer, :user_agent);
	`

	_, err := vs.DB.NamedExec(query, visit)
	if err != nil {
		return fmt.Errorf("failed to save visit: %w", err)
	}

	vs.Logger.Info(
		"user visited shortlink",
		log.Str("kind", kind),
		log.Str("slug", slug),
		log.Str("referer", referer),
		log.Str("user_agent", userAgent),
	)

	return nil
}
