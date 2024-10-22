package internal

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func createDb(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY,
		date VARCHAR(8) NOT NULL,
		title TEXT NOT NULL,
		comment TEXT,
		repeat VARCHAR(128)
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	indexQuery := `CREATE INDEX IF NOT EXISTS task_date ON scheduler (date);`
	_, err = db.Exec(indexQuery)
	return err
}

func InitDb() (*sql.DB, error) {
	fmt.Println("Checking if DB file exists")
	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	dbFileEnv := os.Getenv("TODO_DBFILE")
	if dbFileEnv == "" {
		dbFileEnv = "scheduler.db"
	}

	dbFile := filepath.Join(filepath.Dir(appPath), dbFileEnv)
	_, err = os.Stat(dbFile)

	var install bool
	if err != nil {
		fmt.Println("DB file not found")

		install = true
	}
	fmt.Println("Starting DB", dbFileEnv)

	db, err := sql.Open("sqlite", dbFileEnv)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if install {
		err := createDb(db)
		if err != nil {
			fmt.Println(err)
		}
	}

	return db, nil
}
