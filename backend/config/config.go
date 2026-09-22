package config

import (
	"os"
)

type Config struct {
	ServerPort string
	DBDriver   string
	DBDSN      string
}

func LoadConfig() *Config {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	dbDriver := os.Getenv("DB_DRIVER")
	if dbDriver == "" {
		dbDriver = "mysql"
	}

	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		dbDSN = "root:@tcp(127.0.0.1:3306)/case_study_2?parseTime=true&multiStatements=true"
	}

	return &Config{
		ServerPort: port,
		DBDriver:   dbDriver,
		DBDSN:      dbDSN,
	}
}
