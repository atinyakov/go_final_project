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
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
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

	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(date))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

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

	fmt.Println("Got a port:", envPort)

	taskService := services.NewTaskService(db)
	taskController := controllers.NewTaskController(taskService)
	jwtc := controllers.JwtController{}
	authController := controllers.NewAuthController(&jwtc)

	fmt.Println("Starting a server")
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", handleNextDate)
	http.HandleFunc("/api/task", authController.Auth(taskController.HandleTask))
	http.HandleFunc("/api/tasks", authController.Auth(taskController.HandleAllTasks))
	http.HandleFunc("/api/task/done", authController.Auth(taskController.HandleDoneTask))

	http.HandleFunc("/api/signin", authController.HandleAuth)

	err = http.ListenAndServe("0.0.0.0:"+envPort, nil)
	fmt.Println("Server started on port=", envPort)

	if err != nil {
		panic(err)
	}
	fmt.Println("Stopping App")
}
