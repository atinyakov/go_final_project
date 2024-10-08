package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // Modernc SQLite driver without CGO
)

// func NextDate(now time.Time, date string, repeat string) (string, error) {

// }

func createDb(db *sql.DB) error {
	query := `
	CREATE TABLE scheduler (
		id INTEGER PRIMARY KEY,
		date VARCHAR(8) NOT NULL,
		title TEXT NOT NULL,
		comment TEXT,
		repeat VARCHAR(128)
	);
	CREATE INDEX task_date ON scheduler (date);
	`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	indexQuery := `CREATE INDEX IF NOT EXISTS task_date ON scheduler (date);`
	_, err = db.Exec(indexQuery)
	return err
}

func initDb() {
	fmt.Println("Проверяем наличие файла БД")
	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")
	_, err = os.Stat(dbFile)

	var install bool
	if err != nil {
		fmt.Println("Файл БД не найден")

		install = true
	}
	// если install равен true, после открытия БД требуется выполнить
	// sql-запрос с CREATE TABLE и CREATE INDEX

	fmt.Println("Запускаем БД")

	db, err := sql.Open("sqlite", "scheduler.db")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	if install {
		err := createDb(db)
		if err != nil {
			fmt.Println(err) // TODO обработать ошибку красиво
		}
	}
}

func main() {

	initDb()
	webDir := "./web"

	fmt.Println("Запускаем сервер")
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	err := http.ListenAndServe(":7540", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Завершаем работу")
}
