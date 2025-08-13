package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/ivov/n8n-shortlink/internal/db/entities"
	"github.com/ivov/n8n-shortlink/internal/errors"
	"github.com/ivov/n8n-shortlink/internal/log"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// ShortlinkService manages shortlinks.
type ShortlinkService struct {
	DB     *sqlx.DB
	Logger *log.Logger
}

// SaveShortlink writes a shortlink to the DB.
func (ss *ShortlinkService) SaveShortlink(shortlink *entities.Shortlink) (*entities.Shortlink, error) {
	query := `
		INSERT INTO shortlinks (slug, kind, content, creator_ip, expires_at, password, allowed_visits)
		VALUES (:slug, :kind, :content, :creator_ip, :expires_at, :password, :allowed_visits)
		RETURNING slug, kind, content, creator_ip, created_at, expires_at, password, allowed_visits;
	`

	rows, err := ss.DB.NamedQuery(query, shortlink)
	if err != nil {
		return nil, fmt.Errorf("failed to save shortlink: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.StructScan(shortlink)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.ErrShortlinkNotFound
	}

	ss.Logger.Info(
		"user created shortlink",
		log.Str("slug", shortlink.Slug),
		log.Str("kind", shortlink.Kind),
		log.Str("creator_ip", shortlink.CreatorIP),
		log.Str("with_password", fmt.Sprint(shortlink.Password != "")),
	)

	return shortlink, nil
}

const (
	defaultSlugLength = 4 // 64^4 = ~16.7 million possible slugs
	maxUserSlugLength = 512
)

// GenerateSlug generates a random URL-encoded shortlink slug, ensuring uniqueness with DB.
func (ss *ShortlinkService) GenerateSlug() (string, error) {
	for {
		bytes := make([]byte, defaultSlugLength)

		_, err := rand.Read(bytes[:])
		if err != nil {
			return "", err
		}

		slug := base64.RawURLEncoding.EncodeToString(bytes[:])[:defaultSlugLength]

		isUnique, err := ss.isSlugUnique(slug)
		if err != nil {
			return "", err
		}

		if isUnique && !isReserved(slug) {
			return slug, nil
		}
	}
}

func (ss *ShortlinkService) isSlugUnique(slug string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM shortlinks WHERE slug = $1);"

	err := ss.DB.Get(&exists, query, slug)
	if err != nil {
		return false, err
	}

	return !exists, nil
}

// GetBySlug retrieves the main parts of a shortlink by its slug.
func (ss *ShortlinkService) GetBySlug(slug string) (*entities.Shortlink, error) {
	var shortlink entities.Shortlink
	query := "SELECT kind, content, password FROM shortlinks WHERE slug = $1;"

	err := ss.DB.Get(&shortlink, query, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrShortlinkNotFound
		}

		return nil, err
	}

	return &shortlink, nil
}

// ValidateUserSlug checks if a user-provided slug meets all requirements.
func (ss *ShortlinkService) ValidateUserSlug(slug string) error {
	if len(slug) < defaultSlugLength {
		return errors.ErrSlugTooShort
	}

	if len(slug) > maxUserSlugLength {
		return errors.ErrSlugTooLong
	}

	if !regexp.MustCompile(`^[A-Za-z0-9_-]+$`).MatchString(slug) {
		return errors.ErrSlugMisformatted
	}

	isUnique, err := ss.isSlugUnique(slug)
	if err != nil {
		return fmt.Errorf("error checking for slug uniqueness: %w", err)
	}

	if !isUnique {
		return errors.ErrSlugTaken
	}

	if isReserved(slug) {
		return errors.ErrSlugReserved
	}

	return nil
}

var reservedSlugs = []string{"static", "health", "metrics", "docs", "spec", "challenge"}

func isReserved(path string) bool {
	for _, deny := range reservedSlugs {
		if path == deny {
			return true
		}
	}

	return false
}

// HashPassword generates a bcrypt hash of a plaintext password.
func (ss *ShortlinkService) HashPassword(plaintextPassword string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	if err != nil {
		ss.Logger.Error(err)
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// ValidateKind checks if a kind is supported.
func (ss *ShortlinkService) ValidateKind(kind string) error {
	switch kind {
	case "workflow", "url":
		return nil
	default:
		return fmt.Errorf("found unsupported kind: %s", kind)
	}
}

const passwordMinLength = 8

// ValidatePasswordLength checks if a password's length is valid.
func (ss *ShortlinkService) ValidatePasswordLength(password string) error {
	if len(password) < passwordMinLength {
		return errors.ErrPasswordTooShort
	}

	return nil
}

// VerifyPassword compares a bcrypt hash with a plaintext password.
func (ss *ShortlinkService) VerifyPassword(hashedPassword, plaintextPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plaintextPassword))
	return err == nil
}

var suspiciousPatterns = []string{
	"cpanel.site",
	"screenconnect.com",
	"oauth-",
	"order-payment",
	"order-delivery",

	"signin", "sign-in", "login", "log-in",
	"auth-", "sso-", "verify-account",
	"confirm-account", "activate-account",
	"account-verification", "account-suspended",

	"payment-", "billing-", "invoice-",
	"paypal-", "stripe-", "bank-",
	"refund-", "chargeback", "creditcard",

	"delivery-", "package-", "shipment-",
	"fedex-", "ups-", "dhl-", "usps-",
	"tracking-", "delivered-",

	"support-", "helpdesk-", "tech-support",
	"microsoft-", "apple-", "google-",
	"virus-detected", "security-alert",

	"urgent-", "immediate-", "suspended-",
	"update-", "renew-", "expire-",
	"winner-", "congratulations-", "prize-",

	".tk/", ".ml/", ".ga/", ".cf/",

	"bit.ly", "tinyurl.com", "t.co", "goo.gl",
}

func (ss *ShortlinkService) ValidateContent(content string) error {
	contentLower := strings.ToLower(content)

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(contentLower, pattern) {
			return errors.ErrContentBlocked
		}
	}

	return nil
}
