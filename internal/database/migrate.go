package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"os"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type gooseLogger struct{}

// Fatalf implements goose.Logger.
func (g gooseLogger) Fatalf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	var args []any
	for i, arg := range v {
		args = append(args, slog.Any(fmt.Sprintf("arg%d", i), arg))
	}
	slog.Error(msg, args...)
	os.Exit(1)
}

// Printf implements goose.Logger.
func (g gooseLogger) Printf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	var args []any
	for i, arg := range v {
		args = append(args, slog.Any(fmt.Sprintf("arg%d", i), arg))
	}
	slog.Info(msg, args...)
}

func Migrate(file string, remigrateCount int) error {
	db, err := openDb(file)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)
	goose.SetBaseFS(embedMigrations)
	goose.SetLogger(gooseLogger{})

	if err = goose.SetDialect(string(goose.DialectSQLite3)); err != nil {
		return fmt.Errorf("failed setting up database dialect: %w", err)
	}
	if remigrateCount > 0 {
		slog.Info("Running down migrations", slog.Int("remigrateCount", remigrateCount))
	}
	for range remigrateCount {
		if err := goose.Down(db, "migrations"); err != nil {
			return fmt.Errorf("down migration failed: %w", err)
		}
	}

	slog.Info("Running up migrations")
	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("up migrations failed: %w", err)
	}

	return nil
}
