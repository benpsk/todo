package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/benpsk/todo/cmd/service"
	_ "github.com/mattn/go-sqlite3"
)

type todo struct {
	id        int
	text      string
	priority  string    // <high, medium, low>
	status    string    // <pending, processing, done>
	due       *time.Time // datetime
	tag       *string
	createdAt time.Time // 2025-01-01 13:12:12
}

type listFlag struct {
	status   string
	priority string
	due      *string
	tag      *string
	find     string
	created  string
}

func (l *listFlag) GetStatus() string    { return l.status }
func (l *listFlag) SetStatus(s string)   { l.status = s }
func (l *listFlag) GetPriority() string  { return l.priority }
func (l *listFlag) SetPriority(p string) { l.priority = p }
func (l *listFlag) GetDue() *string       { return l.due }
func (l *listFlag) SetDue(d string)      { l.due = &d }

func parseList() *listFlag {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	parse := service.Parse(fs, "list")
  var due string
  if parse.Due != nil {
    due = strings.ToLower(*parse.Due)
  }
	return &listFlag{
		status:   strings.ToLower(*parse.Status),
		priority: strings.ToLower(*parse.Priority),
		due:      &due,
		tag:      parse.Tag,
		created:  *parse.Created,
	}
}

func (app *App) get(cmd *listFlag) ([]todo, error) {
	query := `
    SELECT id, text, priority, status, due, tag, created_at 
    FROM todos 
    WHERE 1=1  -- Always true to make conditional appending easier
  `
	var args []interface{} // To hold arguments for the query

	if cmd.status != "" {
		query += " AND status=?"
		args = append(args, cmd.status)
	}
	if cmd.priority != "" {
		query += " AND priority=?"
		args = append(args, cmd.priority)
	}
	if cmd.due != nil && *cmd.due != "" {
		query += " AND due <= ?"
		args = append(args, *cmd.due)
	}
	if cmd.tag != nil {
		query += " AND tag like ?"
		args = append(args, "%"+*cmd.tag+"%")
	}
	if cmd.find != "" {
		query += " AND find LIKE ?"
		args = append(args, "%"+cmd.find+"%") // Adding wildcard for LIKE search
	}
	if cmd.created != "" {
		query += " AND created_at=?"
		args = append(args, cmd.created)
	}
	// Order the results
	query += " ORDER BY priority DESC"

	rows, err := app.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []todo
	for rows.Next() {
		var t todo
		if err := rows.Scan(&t.id, &t.text, &t.priority, &t.status,
			&t.due, &t.tag, &t.createdAt); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, rows.Err()
}

func isValidCreated(cmd *listFlag) bool {
	if cmd.created == "" {
		return true
	}
	dueFormat1 := "2006-01-02" // for "2025-08-20"
	currentTime := time.Now()

	// Check if the due date is a specific datetime format (e.g., "2025-08-20")
	if _, err := time.Parse(dueFormat1, cmd.created); err == nil {
		return true
	}
	// Check if the due date is a weekday with time (e.g. "wed")
	weekday := cmd.created
	weekdayMap := map[string]int{
		"sun": 0, "mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6,
	}
	weekdayIndex, exists := weekdayMap[weekday]
	if !exists {
		return false // invalid weekday
	}
	// Find the difference in days between today and the target weekday
	currentWeekday := int(currentTime.Weekday())
	daysDiff := (currentWeekday - weekdayIndex + 7) % 7
	if daysDiff == 0 {
		daysDiff = -7 // If it's the same weekday, subtract a full week to get the last one
	}
	// Get the date for the previous weekday
	weekdayDate := currentTime.AddDate(0, 0, -daysDiff)
	cmd.created = weekdayDate.Format("2006-01-02")
	return true
}

func (app *App) list() {
	cmd := parseList()
	if isValid := service.Validate(cmd); !isValid {
		os.Exit(1)
	}
	if isValid := isValidCreated(cmd); !isValid {
		fmt.Fprintf(os.Stderr, "Invalid created date %v\n", cmd.created)
		os.Exit(1)
	}
	todos, err := app.get(cmd)
	if err != nil {
		log.Fatal(err)
	}
	statuses := map[string]string{
		"1": "pending",
		"2": "processing",
		"3": "done",
	}
	priorities := map[string]string{
		"1": "low",
		"2": "medium",
		"3": "high",
	}
	fmt.Println("=======================================")
	fmt.Println("          📋  Todo List 📋")
	fmt.Println("=======================================")
	fmt.Printf("%-2s | %-10s | %-8s | %-20s | %-10s | %s \n",
		"id", "status", "priority", "due", "tag", "task")
	for _, t := range todos {
		status, _ := statuses[t.status]
		priority, _ := priorities[t.priority]
		var due string
		if t.due != nil {
			due = t.due.Format("2006-01-02 15:04:05")
		}
		fmt.Printf("%-2d | %-10s | %-8s | %-20s | %-10s | %s \n",
			t.id, status, priority, due, *t.tag, t.text)
	}
	fmt.Println("=======================================")
}
