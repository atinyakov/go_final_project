package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/atinyakov/go_final_project/controllers"
	"github.com/atinyakov/go_final_project/nextdate"
	"github.com/atinyakov/go_final_project/services"
	_ "modernc.org/sqlite" // Modernc SQLite driver without CGO
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

func initDb() (*sql.DB, error) {
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
		return nil, err
	}

	if install {
		err := createDb(db)
		if err != nil {
			fmt.Println(err) // TODO обработать ошибку красиво
		}
	}

	return db, nil
}

func handleNextDate(w http.ResponseWriter, req *http.Request) {
	repeat := req.FormValue("repeat")
	reqDate := req.FormValue("date")
	now, err := time.Parse("20060102", req.FormValue("now"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	date, err := nextdate.Get(now, reqDate, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// так как все успешно, то статус OK
	w.WriteHeader(http.StatusOK)
	// записываем сериализованные в JSON данные в тело ответа
	_, err = w.Write([]byte(date))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {

	db, err := initDb()
	if err != nil {
		panic(err)
	}

	defer db.Close()
	webDir := "./web"

	taskService := services.NewTaskService(db)
	taskController := controllers.NewTaskController(taskService)

	fmt.Println("Запускаем сервер")
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", handleNextDate)
	http.HandleFunc("/api/task", taskController.HandleTask)
	http.HandleFunc("/api/tasks", taskController.HandleAllTasks)
	http.HandleFunc("/api/task/done", taskController.HandleDoneTask)
	err = http.ListenAndServe(":7540", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Завершаем работу")
}
