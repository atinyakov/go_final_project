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

	dbFileEnv := os.Getenv("TODO_DBFILE")
	if dbFileEnv == "" {
		dbFileEnv = "scheduler.db"
	}

	dbFile := filepath.Join(filepath.Dir(appPath), dbFileEnv)
	_, err = os.Stat(dbFile)

	var install bool
	if err != nil {
		fmt.Println("Файл БД не найден")

		install = true
	}
	fmt.Println("Запускаем БД")

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

	envPort := os.Getenv("TODO_PORT")
	if envPort == "" {
		envPort = "7540"
	}

	fmt.Println("Got app port:", envPort)

	taskService := services.NewTaskService(db)
	taskController := controllers.NewTaskController(taskService)
	jwtc := controllers.JwtController{}
	authController := controllers.NewAuthController(&jwtc)

	fmt.Println("Запускаем сервер")
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", handleNextDate)
	http.HandleFunc("/api/task", authController.Auth(taskController.HandleTask))
	http.HandleFunc("/api/tasks", authController.Auth(taskController.HandleAllTasks))
	http.HandleFunc("/api/task/done", authController.Auth(taskController.HandleDoneTask))

	http.HandleFunc("/api/signin", authController.HandleAuth)
	err = http.ListenAndServe(":"+envPort, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Завершаем работу")
}
