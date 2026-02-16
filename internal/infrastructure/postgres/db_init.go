package postgres

import (
	"Finance-Manager-System/internal/configs"
	"database/sql"
	"fmt"
)

func New(cfg *configs.Config) (*sql.DB, error) {
	var driverName string = cfg.TypeDB
	var url string

	switch driverName {
	case "postgres":
		url = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s",
			cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User,
			cfg.Postgres.Password, cfg.Postgres.DBName)

	default:
		return nil, fmt.Errorf("unknown db type: %s", cfg.TypeDB)
	}

	db, err := sql.Open(driverName, url)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
