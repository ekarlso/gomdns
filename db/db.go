package db

import (
	log "code.google.com/p/log4go"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

// Connect and return a connection
func Connect(dsn string) *sql.DB {
	log.Info("Connecting to %s", dsn)
	conn, err := sql.Open("mysql", dsn)

	if err != nil {
		log.Crash(err)
	}

	return conn
}

// Check that the DB is valid.
func CheckDB(dsn string) bool {
	conn := Connect(dsn)

	_, err := conn.Exec("SELECT * FROM domains")
	if err != nil {
		log.Crash(err)
	}

	return true
}
