package infrastructure

import (
	"database/sql"
	"fmt"

	"github.com/PKSonlem/testtask-secunda-api/internal/config"
	_ "github.com/go-sql-driver/mysql"
)

func NewMySQL(cfg config.MySQLConfig) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)

	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	return db, nil
}
