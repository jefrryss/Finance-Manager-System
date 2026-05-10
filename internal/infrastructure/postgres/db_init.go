package postgres

import (
	"Finance-Manager-System/configs"
	"fmt"

	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func NewDB(cfg *configs.Config) (*sqlx.DB, error) {
	var typeDB string = cfg.TypeDB
	var driverName string
	var url string

	switch typeDB {
	case "postgres":
		driverName = "pgx"
		url = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s",
			cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User,
			cfg.Postgres.Password, cfg.Postgres.DBName)

	default:
		return nil, fmt.Errorf("unknown db type: %s", cfg.TypeDB)
	}

	db, err := sqlx.Connect(driverName, url)
	if err != nil {
		return nil, fmt.Errorf("fatal to connect to db: %w", err)
	}

	return db, nil
}
