package cmd

import (
	"database/sql"
	"log"
	"os"

	"github.com/benpsk/todo/cmd/daemon"
	"github.com/benpsk/todo/cmd/ui"
	"github.com/benpsk/todo/db"
)

const dbPath = "./todos.db"

type App struct {
	db *sql.DB
}

func new(db *sql.DB) *App {
	return &App{db: db}
}

func Execute() {
	if len(os.Args) < 2 {
		ui.Usage()
		return
	}
	db, err := db.Connect(dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	app := new(db)
	d := daemon.New(db)

	cmd := os.Args[1]
	switch cmd {
	case "add":
		app.add()
	case "list", "ls":
		app.list()
	case "delete":
		app.delete()
	case "update":
		app.update()
	case "--help", "-h":
		ui.Usage()
	case "daemon":
		d.Handle()
	default:
		ui.Usage()
	}
}
