package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/atinyakov/go_final_project/models"
	"github.com/atinyakov/go_final_project/services"
)

type TaskController struct {
	ts *services.TaskService
}

func NewTaskController(ts *services.TaskService) *TaskController {
	return &TaskController{
		ts: ts,
	}
}

func (t *TaskController) HandleTask(w http.ResponseWriter, req *http.Request) {
	method := req.Method

	if method == http.MethodPost {
		var task models.Task
		var buf bytes.Buffer

		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			http.Error(w, "Error: Cannot deserialize JSON", http.StatusBadRequest)
			return
		}

		id, err := t.ts.CreateTask(&task)

		if err != nil {
			errorResponce := &models.TaskResponceError{Error: err.Error()}

			errorJSON, jsonErr := json.Marshal(errorResponce)
			if jsonErr != nil {
				http.Error(w, "Failed to encode task", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(errorJSON), http.StatusInternalServerError)
			return
		}

		fmt.Println("Created task id=", id)

		response := map[string]interface{}{
			"id": id,
		}

		resp, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	if method == http.MethodGet {
		id := req.URL.Query().Get("id")
		task, taskError := t.ts.GetTask(id)

		if taskError != nil {
			errorJSON, jsonErr := json.Marshal(taskError)
			if jsonErr != nil {
				http.Error(w, "Failed to create error response", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(errorJSON), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	if method == http.MethodPut {
		var task models.Task
		var buf bytes.Buffer

		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &task); err != nil {
			http.Error(w, "Error: Cannot deserialize JSON", http.StatusInternalServerError)
			return
		}

		taskError := t.ts.UpdateTask(&task)

		if taskError != nil {
			errorJSON, jsonErr := json.Marshal(taskError)
			if jsonErr != nil {
				http.Error(w, "Failed to update task", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(errorJSON), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(struct{}{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	if method == http.MethodDelete {
		id := req.URL.Query().Get("id")

		deleteError := t.ts.DeleteTask(id)

		if deleteError != nil {
			errorJSON, jsonErr := json.Marshal(deleteError)
			if jsonErr != nil {
				http.Error(w, "Failed to delete task", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(errorJSON), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(struct{}{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}

}

func (t *TaskController) HandleAllTasks(w http.ResponseWriter, req *http.Request) {
	method := req.Method

	if method == http.MethodGet {
		querryParams := req.URL.Query()
		searchString := querryParams.Get("search")

		tasks, taskError := t.ts.GetAllTasks(searchString)

		if taskError != nil {
			errorJSON, jsonErr := json.Marshal(taskError)
			if jsonErr != nil {
				http.Error(w, "Failed to create error response", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(errorJSON), http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(models.GetAllTaksksResponceSuccess{Tasks: tasks})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (t *TaskController) HandleDoneTask(w http.ResponseWriter, req *http.Request) {
	method := req.Method
	id := req.URL.Query().Get("id")

	if method == http.MethodPost {
		markAsDoneErr := t.ts.MarkAsDone(id)

		if markAsDoneErr != nil {
			errorJSON, jsonErr := json.Marshal(markAsDoneErr)
			if jsonErr != nil {
				http.Error(w, "Failed to mark as done task", http.StatusInternalServerError)
				return
			}

			http.Error(w, string(errorJSON), http.StatusBadRequest)
			return
		}

		resp, err := json.Marshal(struct{}{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
