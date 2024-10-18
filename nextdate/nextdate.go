package nextdate

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

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

func weeklyPattern(now time.Time, repeat string) (string, error) {
	// Split the days from the pattern (e.g., "w 1,4,5")
	daysStr := strings.Split(strings.TrimPrefix(repeat, "w "), ",")
	var days []int
	for _, dayStr := range daysStr {
		day, err := strconv.Atoi(dayStr)
		if err != nil || day < 1 || day > 7 {
			return "", fmt.Errorf("invalid day of the week: %s", dayStr)
		}
		days = append(days, day)
	}

	// Find the next valid weekday
	weekday := int(now.Weekday())
	if weekday == 0 { // Sunday is 7 in your scheme
		weekday = 7
	}

	// Calculate the nearest valid weekday
	minDaysToAdd := 8
	for _, day := range days {
		daysToAdd := (day - weekday + 7) % 7
		if daysToAdd == 0 {
			daysToAdd = 7 // Move to next week if it's today
		}
		if daysToAdd < minDaysToAdd {
			minDaysToAdd = daysToAdd
		}
	}

	// Add the calculated days to the current date
	nextDate := now.AddDate(0, 0, minDaysToAdd)
	return nextDate.Format("20060102"), nil
}

func Get(now time.Time, date string, repeat string) (string, error) {
	// Parse task date
	startDate, err := time.Parse("20060102", date)
	if err != nil {
		log.Printf("Task date is not in valid format: %s", err)
		return "", err
	}

	switch {
	case strings.HasPrefix(repeat, "d "):
		return dailyPattern(now, startDate, repeat)
	case repeat == "y":
		return yearlyPattern(now, startDate)
	case strings.HasPrefix(repeat, "w "):
		return weeklyPattern(now, repeat)
	case strings.HasPrefix(repeat, "m "):
		err = errors.New("repetition pattern m is not supported")
		return "", err
	case repeat == "":
		err = errors.New("no repetition range set")
		return "", err
	default:
		err = errors.New("repetition pattern is not supported")
		return "", err
	}
}
