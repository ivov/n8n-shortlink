package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file" // file source
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Setup connects to the DB, sets PRAGMAs, and runs migrations.
func Setup(
	env string,
) (*sqlx.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	dbFilePath := filepath.Join(home, ".n8n-shortlink", "n8n-shortlink.sqlite")

	db, connErr := connect(ctx, dbFilePath)
	if connErr != nil {
		return nil, connErr
	}

	pragmaErr := setPragmas(ctx, db)
	if pragmaErr != nil {
		return nil, pragmaErr
	}

	if err := RunMigrations(db, env); err != nil {
		return nil, err
	}

	var tableExists bool
	err = db.QueryRow(`
		SELECT EXISTS (SELECT name FROM sqlite_master WHERE type='table' AND name='shortlinks');
	`).Scan(&tableExists)
	if err != nil {
		return nil, fmt.Errorf("error checking for shortlinks table: %w", err)
	}

	if !tableExists {
		return nil, fmt.Errorf("failed to find shortlinks table, did you forget to run migrations?")
	}

	return db, nil
}

func connect(
	ctx context.Context,
	filePath string,
) (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite DB: %s", err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func setPragmas(
	ctx context.Context,
	db *sqlx.DB,
) error {
	pragmas := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA cache_size = -20000;",
		"PRAGMA temp_store = memory;",
	}

	for _, pragma := range pragmas {
		_, err := db.ExecContext(ctx, pragma)
		if err != nil {
			return fmt.Errorf("failed to set PRAGMA: %s", err)
		}
	}

	return nil
}

// RunMigrations applies up migrations to the DB.
func RunMigrations(dbConn *sqlx.DB, env string) error {
	driver, err := sqlite3.WithInstance(dbConn.DB, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create sqlite driver: %w", err)
	}

	migrationsDirPath, err := getMigrationsDirPath(env)
	if err != nil {
		return fmt.Errorf("failed to get migrations path: %w", err)
	}
	migrationsURL := fmt.Sprintf("file://%s", migrationsDirPath)

	m, err := migrate.NewWithDatabaseInstance(migrationsURL, "sqlite3", driver)
	if err != nil {
		return fmt.Errorf("failed to create `migrate` instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run up migrations: %w", err)
	}

	return nil
}

// SetupTestDB creates an in-memory SQLite DB and runs migrations.
func SetupTestDB() (*sqlx.DB, error) {
	db, err := sqlx.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to open in-memory database: %w", err)
	}

	err = RunMigrations(db, "testing")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func getMigrationsDirPath(env string) (string, error) {
	var basePath string

	if _, err := os.Stat("/.dockerenv"); err == nil {
		return "/root/n8n-shortlink/internal/db/migrations", nil
	}

	basePath, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current working dir: %w", err)
	}

	if env == "testing" {
		basePath = filepath.Join(basePath, "..", "..") // root dir
	}

	migrationsDirPath := filepath.Join(basePath, "internal", "db", "migrations")

	if _, err := os.Stat(migrationsDirPath); os.IsNotExist(err) {
		return "", fmt.Errorf("migrations dir not found at %s", migrationsDirPath)
	}

	return migrationsDirPath, nil
}
