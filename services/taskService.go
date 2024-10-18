package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/atinyakov/go_final_project/models"
	"github.com/atinyakov/go_final_project/nextdate"
)

type TaskService struct {
	db *sql.DB
}

func NewTaskService(db *sql.DB) *TaskService {
	return &TaskService{db: db}
}

func (ts *TaskService) CreateTask(task *models.Task) (int64, error) {
	err := models.ValidateTask(task)
	if err != nil {
		return 0, err
	}

	res, err := ts.db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat))

	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (ts *TaskService) GetTask(id string) (*models.Task, *models.TaskResponceError) {
	row := ts.db.QueryRow("SELECT * FROM scheduler WHERE id = :id", sql.Named("id", id))
	task := models.Task{}

	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)

	if err != nil {
		log.Println(err)
		return nil, &models.TaskResponceError{
			Error: errors.New("Task not found").Error(),
		}
	}

	return &task, nil
}

func (ts *TaskService) DeleteTask(id string) *models.TaskResponceError {
	result, err := ts.db.Exec("DELETE FROM scheduler WHERE id = :id", sql.Named("id", id))
	if err != nil {
		fmt.Println(err)
		return &models.TaskResponceError{
			Error: "Failed to delete task",
		}
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		fmt.Println(err)
		return &models.TaskResponceError{
			Error: "Failed to retrieve affected rows",
		}
	}

	if rowsAffected == 0 {
		return &models.TaskResponceError{
			Error: fmt.Sprintf("Task with ID %s not found", id),
		}
	}

	return nil
}

func (ts *TaskService) MarkAsDone(id string) *models.TaskResponceError {
	task, err := ts.GetTask(id)
	if err != nil {
		return err
	}

	if task.Repeat == "" {
		return ts.DeleteTask(id)
	} else {
		nextdate, err := nextdate.Get(time.Now(), task.Date, task.Repeat)
		if err != nil {
			return &models.TaskResponceError{
				Error: err.Error(),
			}
		}

		result, err := ts.db.Exec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id",
			sql.Named("date", nextdate),
			sql.Named("title", task.Title),
			sql.Named("comment", task.Comment),
			sql.Named("repeat", task.Repeat),
			sql.Named("id", task.ID))

		if err != nil {
			return &models.TaskResponceError{
				Error: fmt.Errorf("failed to update task: %w", err).Error(),
			}
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return &models.TaskResponceError{
				Error: fmt.Errorf("failed to retrieve rows affected: %w", err).Error(),
			}
		}

		if rowsAffected == 0 {
			return &models.TaskResponceError{
				Error: fmt.Errorf("task with ID %s not found", task.ID).Error(),
			}
		}

		return nil
	}
}

func (ts *TaskService) UpdateTask(task *models.Task) *models.TaskResponceError {
	validationErr := models.ValidateTask(task)

	if validationErr != nil {
		return &models.TaskResponceError{
			Error: validationErr.Error(),
		}
	}

	result, err := ts.db.Exec("UPDATE scheduler SET date = :date, title = :title, comment = :comment, repeat = :repeat WHERE id = :id",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),
		sql.Named("comment", task.Comment),
		sql.Named("repeat", task.Repeat),
		sql.Named("id", task.ID))

	if err != nil {
		return &models.TaskResponceError{
			Error: fmt.Errorf("failed to update task: %w", err).Error(),
		}
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return &models.TaskResponceError{
			Error: fmt.Errorf("failed to retrieve rows affected: %w", err).Error(),
		}
	}

	if rowsAffected == 0 {
		return &models.TaskResponceError{
			Error: fmt.Errorf("task with ID %s not found", task.ID).Error(),
		}
	}

	return nil
}

func (ts *TaskService) GetAllTasks(search string) ([]*models.Task, *models.TaskResponceError) {

	var rows *sql.Rows
	var sqlError error
	if search != "" {
		t1, err := time.Parse("02.01.2006", search) // check if date
		if err != nil {
			rows, sqlError = ts.db.Query("SELECT * FROM scheduler WHERE title LIKE :search OR comment LIKE :search ORDER BY date LIMIT :limit",
				sql.Named("search", "%"+search+"%"),
				sql.Named("limit", 50))

		} else {
			search = t1.Format("20060102")
			rows, sqlError = ts.db.Query("SELECT * FROM scheduler WHERE date = :date LIMIT :limit ",
				sql.Named("date", search),
				sql.Named("limit", 50))

		}
	} else {
		rows, sqlError = ts.db.Query("SELECT * from scheduler ORDER BY date DESC LIMIT 50")
	}

	if sqlError != nil {
		log.Println(sqlError)
		return nil, &models.TaskResponceError{
			Error: sqlError.Error(),
		}
	}

	defer rows.Close()

	var allTasks []*models.Task

	for rows.Next() {
		task := models.Task{}

		err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err != nil {
			log.Println(err)
			return nil, &models.TaskResponceError{
				Error: err.Error(),
			}
		}

		allTasks = append(allTasks, &task)
	}

	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, &models.TaskResponceError{
			Error: err.Error(),
		}
	}

	// If no tasks found, return an empty slice instead of nil
	if len(allTasks) == 0 {
		return []*models.Task{}, nil
	}

	return allTasks, nil
}
