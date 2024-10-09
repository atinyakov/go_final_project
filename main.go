package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite" // Modernc SQLite driver without CGO
)

func dailyPattern(now time.Time, startDate time.Time, repeat string) (string, error) {
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
	for now.After(nextDate) || nextDate == now {
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

func handleNextDate(w http.ResponseWriter, req *http.Request) {
	repeat := req.FormValue("repeat")
	reqDate := req.FormValue("date")
	now, err := time.Parse("20060102", req.FormValue("now"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	date, err := NextDate(now, reqDate, repeat)
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

	initDb()
	webDir := "./web"

	fmt.Println("Запускаем сервер")
	http.Handle("/", http.FileServer(http.Dir(webDir)))
	http.HandleFunc("/api/nextdate", handleNextDate)
	err := http.ListenAndServe(":7540", nil)
	if err != nil {
		panic(err)
	}
	fmt.Println("Завершаем работу")
}

// func NextDate(now time.Time, date string, repeat string) (string, error) {
// 	// Validate the input date format (YYYYMMDD)
// 	if len(date) != 8 {
// 		return "", fmt.Errorf("NextDate: invalid date format %s", date)
// 	}
// 	t1, err := time.Parse("20060102", date)
// 	if err != nil {
// 		return "", fmt.Errorf("NextDate: failed to parse date %s: %v", date, err)
// 	}

// 	combinedRule := strings.Split(repeat, " ")
// 	rule := combinedRule[0]

// 	// Daily repeat (d)
// 	if rule == "d" {
// 		if len(combinedRule) < 2 {
// 			return "", fmt.Errorf("NextDate: incorrect repeat rule %s", repeat)
// 		}

// 		days := combinedRule[1]
// 		numberOfDays, err := strconv.Atoi(days)
// 		if err != nil {
// 			return "", fmt.Errorf("NextDate: can't parse %s", days)
// 		}

// 		if numberOfDays > 400 {
// 			return "", fmt.Errorf("NextDate: number of days %d exceed limit", numberOfDays)
// 		}

// 		// Add days until we are after the 'now' date
// 		for !t1.After(now) {
// 			t1 = t1.AddDate(0, 0, numberOfDays)
// 		}

// 		return t1.Format("20060102"), nil
// 	}

// 	// Yearly repeat (y)
// 	if rule == "y" {
// 		// Handle far back dates by resetting the date to the current year
// 		if t1.Before(now) {
// 			for !t1.After(now) {
// 				t1 = t1.AddDate(1, 0, 0)
// 			}
// 		} else {
// 			t1 = t1.AddDate(1, 0, 0)
// 		}

// 		return t1.Format("20060102"), nil
// 	}

// 	return "", fmt.Errorf("NextDate: not supported format")
// }
