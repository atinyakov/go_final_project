package internal

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/atinyakov/go_final_project/controllers"
	"github.com/atinyakov/go_final_project/services"
)

func InitServer(webDir string, db *sql.DB) {
	taskService := services.NewTaskService(db)
	nextDateService := services.NewNextDateService()
	taskController := controllers.NewTaskController(taskService)
	jwtc := controllers.JwtController{}
	authController := controllers.NewAuthController(&jwtc)

	fmt.Println("Starting a server")
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", nextDateService.HandleNextDate)
	http.HandleFunc("/api/task", authController.Auth(taskController.HandleTask))
	http.HandleFunc("/api/tasks", authController.Auth(taskController.HandleAllTasks))
	http.HandleFunc("/api/task/done", authController.Auth(taskController.HandleDoneTask))

	http.HandleFunc("/api/signin", authController.HandleAuth)
}
