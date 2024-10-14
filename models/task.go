package models

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/atinyakov/go_final_project/nextdate"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date,omitempty"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

type TaskResponceSuccess struct {
	ID string `json:"id"`
}

type TaskResponceError struct {
	Error string `json:"error"`
}

type GetAllTaksksResponceSuccess struct {
	Tasks []*Task `json:"tasks"`
}

// hasDuplicates проверяет наличие дубликатов в списке строк
func hasDuplicates(arr []string) bool {
	seen := make(map[string]bool)
	for _, v := range arr {
		if seen[v] {
			return true
		}
		seen[v] = true
	}
	return false
}

// validateRepeat проверяет значение repeat по описанным паттернам
func validateRepeat(repeat string) error {
	// Определение регулярных выражений для каждого паттерна

	// d <число> — от 1 до 400
	dPattern := regexp.MustCompile(`^d\s([1-9][0-9]{0,2}|[1-3][0-9]{2}|400)$`)

	// y — ежегодно
	yPattern := regexp.MustCompile(`^y$`)

	// w <1-7> — через запятую, числа от 1 до 7
	wPattern := regexp.MustCompile(`^w\s([1-7](,[1-7])*)$`)

	// m <1-31,-1,-2> [через запятую от 1 до 12]
	// mPattern := regexp.MustCompile(`^m\s([1-9]|[12][0-9]|3[01]|-1|-2)(,([1-9]|[12][0-9]|3[01]|-1|-2))*\s?([1-9]|1[0-2])(,([1-9]|1[0-2]))*$`)

	// Проверка соответствия каждому паттерну
	switch {
	case dPattern.MatchString(repeat):
		return nil
	case yPattern.MatchString(repeat):
		return nil
	case wPattern.MatchString(repeat):
		// Проверка на дубликаты дней недели (например, "w 1,1" недопустимо)
		days := strings.Split(repeat[2:], ",")
		if hasDuplicates(days) {
			return fmt.Errorf("некорректные дни недели: дубликаты значений")
		}
		return nil
	// case mPattern.MatchString(repeat):
	// 	return nil
	default:
		fmt.Println("Не подеерживаемый формат", repeat)
		return errors.New("некорректный формат repeat")
	}
}

func ValidateTask(t *Task) error {
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

			t2, err := nextdate.Get(time.Now(), t.Date, t.Repeat)

			if err != nil {
				return err
			}

			t.Date = t2
			return nil
		}

	}

	return nil
}
