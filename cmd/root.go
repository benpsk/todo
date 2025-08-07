package cmd

import (
	"database/sql"
	"log"
	"os"

	"github.com/benpsk/todo/cmd/notification"
	"github.com/benpsk/todo/cmd/ui"
	"github.com/benpsk/todo/db"
)

const dbPath = "./db/todos.db"

type handler struct {
	db *sql.DB
}

func new(db *sql.DB) *handler {
	return &handler{db: db}
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
	handler := new(db)

	cmd := os.Args[1]
	switch cmd {
	case "add":
		handler.add()
	case "list", "ls":
		handler.list()
	case "delete":
		handler.delete()
	case "update":
		handler.update()
	case "--help", "-h":
		ui.Usage()
	case "cron":
		notification.Handle()
	default:
		ui.Usage()
	}
}
