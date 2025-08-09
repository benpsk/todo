package daemon

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func (app *App) runDaemon() {
	if app.isDaemonRunning() {
		fmt.Println("Daemon is already running")
		return
	}

	fmt.Println("Starting todo daemon...")

	// Write PID file
	if err := app.writePID(); err != nil {
		log.Fatalf("Failed to write PID file: %v", err)
	}
	defer app.removePID()

	// Setup cron jobs
	app.setupCronJobs()
	app.cron.Start()
	defer app.cron.Stop()

	fmt.Println("Daemon started. Press Ctrl+C to stop.")

	// Wait for interrupt signal
	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nShutting down daemon...")
		cancel()
	}()

	<-ctx.Done()
	fmt.Println("Daemon stopped.")
}

func (app *App) isDaemonRunning() bool {
	data, err := os.ReadFile(app.pidFile)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return false
	}

	// Check if process is still running
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix systems, sending signal 0 checks if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

func (app *App) writePID() error {
	pid := os.Getpid()
	return os.WriteFile(app.pidFile, []byte(strconv.Itoa(pid)), 0644)
}

func (app *App) removePID() error {
	return os.Remove(app.pidFile)
}

func (app *App) stopDaemon() {
	if !app.isDaemonRunning() {
		fmt.Println("Daemon is not running")
		return
	}

	data, err := os.ReadFile(app.pidFile)
	if err != nil {
		fmt.Printf("Failed to read PID file: %v\n", err)
		return
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		fmt.Printf("Invalid PID in file: %v\n", err)
		return
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("Failed to find process: %v\n", err)
		return
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		fmt.Printf("Failed to stop daemon: %v\n", err)
		return
	}

	fmt.Println("Daemon stopped.")
}

func (app *App) setupCronJobs() {
	app.cron.AddFunc("*/5 * * * *", func() {
		fmt.Println("todo cron: run schedule!")
		app.execute()
	})
}
