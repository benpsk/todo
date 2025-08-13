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
	priority  string     // <high, medium, low>
	status    string     // <pending, processing, done>
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
func (l *listFlag) GetDue() *string      { return l.due }
func (l *listFlag) SetDue(d *string)     { l.due = d }

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
	if cmd.due != nil {
		q, argv := service.DateQuery(*cmd.due, "due", "<=")
		query += q
		for _, v := range argv {
			args = append(args, v)
		}
	}
	if *cmd.tag != "" {
		query += " AND tag like ?"
		args = append(args, "%"+*cmd.tag+"%")
	}
	if cmd.find != "" {
		query += " AND find LIKE ?"
		args = append(args, "%"+cmd.find+"%") // Adding wildcard for LIKE search
	}
	if cmd.created != "" {
		q, argv := service.DateQuery(cmd.created, "created_at", "=")
		query += q
		for _, v := range argv {
			args = append(args, v)
		}
	}
	// default filter last 7 days
	if len(args) == 0 {
		last7 := time.Now().AddDate(0, 0, -7)
		query += " and created_at>=?"
		args = append(args, last7)
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
	daysDiff, isSuccess := service.IsValidateDate(cmd.created)
	if !isSuccess {
		return false
	}
	// default no value change
	if daysDiff == 99 {
		return true
	}
	// Calculate days to previous weekday
	daysToLast := (7 - daysDiff) % 7
	if daysToLast == 0 {
		daysToLast = 7 // Previous occurrence was a week ago if it's the same day
	}
	weekdayDate := time.Now().AddDate(0, 0, -daysToLast)
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
