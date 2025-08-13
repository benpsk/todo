package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/benpsk/todo/cmd/service"
)

type addFlag struct {
	text     string
	status   string
	priority string
	due      *string
	tag      *string
}

func (a *addFlag) GetStatus() string    { return a.status }
func (a *addFlag) SetStatus(s string)   { a.status = s }
func (a *addFlag) GetPriority() string  { return a.priority }
func (a *addFlag) SetPriority(p string) { a.priority = p }
func (a *addFlag) GetDue() *string       { return a.due }
func (a *addFlag) SetDue(d *string)      { a.due = d }

func parseAdd() *addFlag {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	parse := service.Parse(fs, "add")

	if len(parse.NonFlagArgs) == 0 {
		fmt.Println("usage: todo add \"task text\" [--status=STATUS] [--priority=PRIORITY] [--due=DATE] [--tag=TAG]")
		os.Exit(1)
	}
	text := parse.NonFlagArgs[0]
  var due string
  if parse.Due != nil {
    due = strings.ToLower(*parse.Due)
  }
	return &addFlag{
		text:     text,
		status:   strings.ToLower(*parse.Status),
		priority: strings.ToLower(*parse.Priority),
		due:      &due,
		tag:      parse.Tag,
	}
}

func (app *App) save(cmd *addFlag) error {
	if cmd.status == "" {
		cmd.status = "1" // pending
	}
	if cmd.priority == "" {
		cmd.priority = "2" // medium
	}
	_, err := app.db.Exec(`
    insert into todos(text, status, priority, due, tag) values(?,?,?,?,?)
  `, cmd.text, cmd.status, cmd.priority, &cmd.due, cmd.tag)
	return err
}

func (app *App) add() {
	cmd := parseAdd()
	if isValid := service.Validate(cmd); !isValid {
		os.Exit(1)
	}
	if err := app.save(cmd); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Success: Todo Save!")
}
