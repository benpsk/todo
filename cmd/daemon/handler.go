package daemon

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/robfig/cron/v3"
)

type App struct {
	pidFile string
	cron    *cron.Cron
	db      *sql.DB
}

func New(db *sql.DB) *App {
	homeDir, _ := os.UserHomeDir()
	return &App{
		pidFile: filepath.Join(homeDir, ".todo.pid"),
		cron:    cron.New(),
		db:      db,
	}
}

func (app *App) Handle() {
	if len(os.Args) < 3 {
		fmt.Println("daemon status | start | stop")
		os.Exit(1)
	}
	cmd := os.Args[2]
	switch cmd {
	case "status":
		if app.isDaemonRunning() {
			fmt.Println("Daemon is running")
		} else {
			fmt.Println("Daemon is not running")
		}
	case "start":
		app.runDaemon()
	case "stop":
		app.stopDaemon()
	default:
		os.Exit(1)
	}
}
