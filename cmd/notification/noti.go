package notification

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/robfig/cron/v3"
)

func notify(title, message string) error {

	err := notify("Todo Reminder", "You have a task due!")
	if err != nil {
		log.Fatalf("Notification error: %v", err)
	}
	cmd := exec.Command("notify-send", title, message)
	return cmd.Run()
}

var cronScheduler *cron.Cron
var logFile *os.File

// Start the cron job and log to a file
func startCronJob(wg *sync.WaitGroup) {
	var err error
	// Open log file in append mode, creating it if it doesn't exist
	logFile, err = os.OpenFile("cron_output.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		os.Exit(1)
	}

	// Set up the logger to write to the file
	logger := log.New(logFile, "", log.LstdFlags)

	// Create a new cron scheduler
	cronScheduler = cron.New()

	// Add a job to the scheduler with the cron expression "* * * * *" (every minute)
	_, err = cronScheduler.AddFunc("* * * * *", func() {
		// Log a message each time the job runs
		logger.Printf("Job 'Yes' is running!\n")
	})
	if err != nil {
		fmt.Println("Error adding cron job:", err)
		os.Exit(1)
	}

	// Start the cron scheduler in a background goroutine
	go func() {
		cronScheduler.Start() // Start the cron scheduler
		wg.Done()              // Decrement the WaitGroup counter when cron job starts
	}()

	fmt.Println("Cron job scheduled and running in the background.")
}

// Stop the cron scheduler
func stopCronJob() {
	if cronScheduler != nil {
		cronScheduler.Stop() // Stop the cron scheduler
		fmt.Println("Cron scheduler stopped.")
	}
}

// Handle commands passed as arguments
func Handle() {
	args := os.Args[2:]
	if len(args) == 0 {
		return
	}
	if args[0] == "start" {
		var wg sync.WaitGroup
		wg.Add(1) // Add a counter for the goroutine running the cron job
		startCronJob(&wg)

		// Wait for the cron job goroutine to finish before allowing the program to exit
		wg.Wait()
	} else if args[0] == "stop" {
		stopCronJob()
	}
}
