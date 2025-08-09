package daemon

import (
	"fmt"
	"log"
	"os/exec"
	"time"
)

type task struct {
	text string
}

func (app *App) execute() {
	title := "Reminder!"
	msg := "Your tasks: \n"
	var tasks []task
	var err error
	now := time.Now()

	t := fmt.Sprintf("%02d%02d", now.Hour(), now.Minute())
	switch t {
	case "0900":
		title = "Good Morning!"
		tasks, err = app.getTasks()
	case "1700":
		title = "Good Evening!"
		tasks, err = app.getTasks()
	default:
		tasks, err = app.getScheduleTasks()
	}
	if err != nil {
		log.Fatalf("Tasks query error: %v", err)
		return
	}
	if len(tasks) == 0 {
		return
	}
	for _, v := range tasks {
		msg += v.text + "\n"
	}
	err = notify(title, msg)
	if err != nil {
		log.Fatalf("Notification error: %v", err)
	}
}

func notify(title, message string) error {
	cmd := exec.Command("zenity", "--info", "--title="+title, "--text="+message)
	return cmd.Run()
}

func (app *App) getScheduleTasks() ([]task, error) {
	now := time.Now()
	next := now.Add(time.Minute * 4)
	formattedNow := now.Format("2006-01-02 15:04")
	formattedNext := next.Format("2006-01-02 15:04")
	query := `
    select text
    from todos 
    where due <= ? and due >= ?
  `
	rows, err := app.db.Query(query, formattedNext, formattedNow)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []task
	for rows.Next() {
		var t task
		if err := rows.Scan(&t.text); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (app *App) getTasks() ([]task, error) {
	now := time.Now()
	formattedTime := now.Format("2006-01-02 15:04")
	query := `
    select text
    from todos 
    where due <= ? and status != ?
  `
	rows, err := app.db.Query(query, formattedTime, 3)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []task
	for rows.Next() {
		var t task
		if err := rows.Scan(&t.text); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
