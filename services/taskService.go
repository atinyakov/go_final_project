package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/atinyakov/go_final_project/models"
)

type TaskService struct {
	db *sql.DB
}

func NewTaskService(db *sql.DB) *TaskService {
	return &TaskService{db: db}
}

func (ts *TaskService) CreateTask(task *models.Task) (int64, error) {
	err := ValidateTask(task)
	if err != nil {
		return 0, err
	}

	res, err := ts.db.Exec("INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)",
		sql.Named("date", task.Date),
		sql.Named("title", task.Title),     // This should map correctly to the "title" field
		sql.Named("comment", task.Comment), // This should map correctly to the "comment" field
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
		nextdate, err := NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			return &models.TaskResponceError{
				Error: err.Error(),
			}
		}
		// return ts.UpdateTask(&models.Task{ID: task.ID, Date: nextdate, Title: task.Title, Comment: task.Comment, Repeat: task.Repeat})
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
	validationErr := ValidateTask(task)

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

// hasDuplicates проверяет наличие дубликатов в списке строк
// func hasDuplicates(arr []string) bool {
// 	seen := make(map[string]bool)
// 	for _, v := range arr {
// 		if seen[v] {
// 			return true
// 		}
// 		seen[v] = true
// 	}
// 	return false
// }

// validateRepeat проверяет значение repeat по описанным паттернам
func validateRepeat(repeat string) error {
	// Определение регулярных выражений для каждого паттерна

	// d <число> — от 1 до 400
	dPattern := regexp.MustCompile(`^d\s([1-9][0-9]{0,2}|[1-3][0-9]{2}|400)$`)

	// y — ежегодно
	yPattern := regexp.MustCompile(`^y$`)

	// w <1-7> — через запятую, числа от 1 до 7
	// wPattern := regexp.MustCompile(`^w\s([1-7](,[1-7])*)$`)

	// m <1-31,-1,-2> [через запятую от 1 до 12]
	// mPattern := regexp.MustCompile(`^m\s([1-9]|[12][0-9]|3[01]|-1|-2)(,([1-9]|[12][0-9]|3[01]|-1|-2))*\s?([1-9]|1[0-2])(,([1-9]|1[0-2]))*$`)

	// Проверка соответствия каждому паттерну
	switch {
	case dPattern.MatchString(repeat):
		return nil
	case yPattern.MatchString(repeat):
		return nil
	// case wPattern.MatchString(repeat):
	// 	// Проверка на дубликаты дней недели (например, "w 1,1" недопустимо)
	// 	days := strings.Split(repeat[2:], ",")
	// 	if hasDuplicates(days) {
	// 		return fmt.Errorf("некорректные дни недели: дубликаты значений")
	// 	}
	// 	return nil
	// case mPattern.MatchString(repeat):
	// 	return nil
	default:
		fmt.Println("Не подеерживаемый формат", repeat)
		return errors.New("некорректный формат repeat")
	}
}

func dailyPattern(now time.Time, startDate time.Time, repeat string) (string, error) {
	currentDate := now.Format("20060102")
	today, _ := time.Parse("20060102", currentDate)

	// Parse numeric value from repetition pattern
	days, err := strconv.Atoi(strings.TrimPrefix(repeat, "d "))
	if err != nil {
		log.Printf("Error parsing pattern value as int: %s\n", err)
		return "", err
	}
	// Define repetition range
	if days <= 0 || days > 400 {
		log.Printf("Invalid repetition range for 'd' pattern: %s\n", err)
		err = errors.New("invalid repetition range for 'd' pattern")
		return "", err
	}

	nextDate := startDate

	// Calculate repetition date if task date in the future
	nextDate = nextDate.AddDate(0, 0, days)

	// Calculate repetition date if task date in the past
	for today.After(nextDate) {
		nextDate = nextDate.AddDate(0, 0, days)
	}

	return nextDate.Format("20060102"), nil
}

// yearlyPattern takes repetition rule in "y" pattern, task date and now time and return repetition date of a task such as an error
func yearlyPattern(now time.Time, startDate time.Time) (string, error) {
	// Calculate repetition date if task date in the future
	nextDate := startDate.AddDate(1, 0, 0)
	// Calculate repetition date if task date in the past
	for now.After(nextDate) || nextDate == now {
		nextDate = nextDate.AddDate(1, 0, 0)
	}
	return nextDate.Format("20060102"), nil
}

// NextDate takes repetition rule, task date as string and now time and return repetition date of a task such as an error
func NextDate(now time.Time, date string, repeat string) (string, error) {
	// Parse task date
	startDate, err := time.Parse("20060102", date)
	if err != nil {
		log.Printf("Task date is not in valid format: %s", err)
		return "", err
	}

	// Chose suitable calculation func
	switch {
	case strings.HasPrefix(repeat, "d "):
		return dailyPattern(now, startDate, repeat)
	case repeat == "y":
		return yearlyPattern(now, startDate)
	case repeat == "":
		err = errors.New("no repetition range set")
		return "", err
	default:
		err = errors.New("repetition pattern is not supported")
		return "", err
	}
}

func ValidateTask(t *models.Task) error {
	if t.Title == "" {
		return errors.New("Title is not provided")
	}

	// Date missing
	now := time.Now().Format("20060102")
	if t.Date == "" {
		t.Date = now
	}
	t1, err := time.Parse("20060102", t.Date) // check form
	if err != nil {
		return errors.New("invalid date")
	}

	if t1.Before(time.Now()) { // check it is in the future or today
		if t.Repeat == "" {
			t.Date = now
		} else {
			err := validateRepeat(t.Repeat)
			if err != nil {
				fmt.Println("Rule not supported", t.Repeat, err.Error())
				return err
			}

			t2, err := NextDate(time.Now(), t.Date, t.Repeat)

			if err != nil {
				return err
			}

			t.Date = t2
			return nil
		}

	}

	return nil
}
