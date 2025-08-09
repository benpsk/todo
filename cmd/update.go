package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/benpsk/todo/cmd/service"
)

type updateFlag struct {
	ids      []int
	text     string
	status   string
	priority string
	due      *string
	tag      *string
}

func (a *updateFlag) GetStatus() string    { return a.status }
func (a *updateFlag) SetStatus(s string)   { a.status = s }
func (a *updateFlag) GetPriority() string  { return a.priority }
func (a *updateFlag) SetPriority(p string) { a.priority = p }
func (a *updateFlag) GetDue() *string       { return a.due }
func (a *updateFlag) SetDue(d string)      { a.due = &d }

func parseUpdate() *updateFlag {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
  parse := service.Parse(fs, "update <id>")

  // Extract IDs and text
	idList := make([]int, 0, len(parse.NonFlagArgs))
	var text string
	for _, arg := range parse.NonFlagArgs {
		id, err := strconv.Atoi(arg)
		if err == nil {
			idList = append(idList, id)
		} else {
			if text != "" {
				fmt.Fprintf(os.Stderr, "error: multiple text arguments provided: %q\n", arg)
				fs.Usage()
				os.Exit(1)
			}
			text = arg
		}
	}

	if len(idList) == 0 {
		fmt.Fprintf(os.Stderr, "error: at least one ID required\n")
		fs.Usage()
		os.Exit(1)
	}

  var due string
  if parse.Due != nil {
    due = strings.ToLower(*parse.Due)
  }
	return &updateFlag{
		ids:      idList,
		text:     text,
		status:   strings.ToLower(*parse.Status),
		priority: strings.ToLower(*parse.Priority),
		due:      &due,
		tag:      parse.Tag,
	}
}

func (app *App) updateTodo(cmd *updateFlag) error {
	query := `
    update todos 
    set updated_at = CURRENT_TIMESTAMP
  `
	var args []interface{}
	if cmd.text != "" {
		query += ", text=?"
		args = append(args, cmd.text)
	}
	if cmd.status != "" {
		query += ", status=?"
		args = append(args, cmd.status)
	}
	if cmd.priority != "" {
		query += ", priority=?"
		args = append(args, cmd.priority)
	}
	if *cmd.due != "" {
		query += ", due=?"
		args = append(args, cmd.due)
	}
	if cmd.tag != nil {
		query += ", tag=?"
		args = append(args, cmd.tag)
	}
	placeholders := make([]string, len(cmd.ids))
	for i, id := range cmd.ids {
		placeholders[i] = "?"
		args = append(args, id)
	}
	query += " where id in (" + strings.Join(placeholders, ",") + ")"

	_, err := app.db.Exec(query, args...)
	return err
}

func (app *App) update() {
	cmd := parseUpdate()
	if isValid := service.Validate(cmd); !isValid {
		os.Exit(1)
	}
	if err := app.updateTodo(cmd); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Success: Todo Updated!")
}
