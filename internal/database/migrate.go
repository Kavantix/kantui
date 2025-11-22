package database

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func OpenDb() (*sql.DB, error) {
	return sql.Open("sqlite", "kantui.sqlite3")
}

func Migrate() error {
	db, err := OpenDb()
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

	if err = goose.SetDialect(string(goose.DialectSQLite3)); err != nil {
		return fmt.Errorf("failed setting up database dialect: %w", err)
	}
	// if err := goose.Down(db, "migrations"); err != nil {
	// 	return fmt.Errorf("down migrations failed: %w", err)
	// }

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("up migrations failed: %w", err)
	}

	return nil
}
