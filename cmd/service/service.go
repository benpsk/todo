package service

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Flagger interface {
	GetStatus() string
	GetPriority() string
	GetDue() *string
	SetStatus(string)
	SetPriority(string)
	SetDue(string)
}

func Validate(cmd Flagger) bool {
	var msg []string
	statuses := map[string]string{
		"pending":    "1",
		"processing": "2",
		"done":       "3",
	}
	if cmd.GetStatus() != "" {
		if validStatus, exists := statuses[cmd.GetStatus()]; exists {
			cmd.SetStatus(validStatus)
		} else if !slices.Contains([]string{"1", "2", "3"}, cmd.GetStatus()) {
			msg = append(msg, fmt.Sprintf("Invalid status %v", cmd.GetStatus()))
		}
	}
	priorities := map[string]string{
		"low":    "1",
		"medium": "2",
		"high":   "3",
	}
	if cmd.GetPriority() != "" {
		if validPriority, exists := priorities[cmd.GetPriority()]; exists {
			cmd.SetPriority(validPriority)
		} else if !slices.Contains([]string{"1", "2", "3"}, cmd.GetPriority()) {
			msg = append(msg, fmt.Sprintf("Invalid priority %v", cmd.GetPriority()))
		}
	}
	if cmd.GetDue() != nil && *cmd.GetDue() != "" {
		if !isValidDueDate(cmd) {
			msg = append(msg, fmt.Sprintf("Invalid due date %v", cmd.GetDue()))
		}
	}
	if len(msg) > 0 {
		fmt.Println("Errors:")
		for _, v := range msg {
			fmt.Fprintf(os.Stderr, "    %v\n", v)
		}
		return false
	}
	return true
}

func isValidDueDate(cmd Flagger) bool {
	dueFormat1 := "2006-01-02"       // for "2025-08-20"
	dueFormat2 := "2006-01-02 15:04" // for "2025-08-20 19:10"
	dueFormat3 := "15:04"            // for "19:20"
	currentTime := time.Now()

	// Check if the due date is a specific datetime format (e.g., "2025-08-20")
	if _, err := time.Parse(dueFormat1, *cmd.GetDue()); err == nil {
		return true
	}
	// Check if the due date is a specific datetime format (e.g., "2025-08-20 19:10")
	if _, err := time.Parse(dueFormat2, *cmd.GetDue()); err == nil {
		return true
	}
	// Check if the due date is in the time-only format (e.g., "19:20")
	if _, err := time.Parse(dueFormat3, *cmd.GetDue()); err == nil {
		cmd.SetDue(currentTime.Format("2006-01-02") + " " + *cmd.GetDue())
		return true
	}
	// Check if the due date is a weekday with time (e.g., "wed-19:10", "wed")
	parts := strings.Split(*cmd.GetDue(), "-")
	timePart := "23:59" // default
	if len(parts) == 2 {
		timePart = parts[1]
	}
	weekday := parts[0]
	// Check if it's a valid time format
	if _, err := time.Parse(dueFormat3, timePart); err == nil {
		// Get the current week day
		weekdayMap := map[string]int{
			"sun": 0, "mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6,
		}
		// If the weekday is valid, calculate the date for that weekday
		weekdayIndex, exists := weekdayMap[weekday]
		if !exists {
			return false // invalid weekday
		}
		// Find the difference in days between today and the target weekday
		currentWeekday := int(currentTime.Weekday())
		daysDiff := (weekdayIndex - currentWeekday + 7) % 7 // handle week wraparound
		// Get the date for that weekday
		weekdayDate := currentTime.AddDate(0, 0, daysDiff)
		cmd.SetDue(weekdayDate.Format("2006-01-02") + " " + timePart)
		return true
	}
	return false
}

func ValidateIds(ids []string) []int {
	idList := make([]int, 0, len(ids))
	for _, arg := range ids {
		id, err := strconv.Atoi(arg)
		if err != nil {
			fmt.Printf("invalid id %q: must be an integer\n", arg)
			os.Exit(1)
		}
		idList = append(idList, id)
	}
	return idList
}

type ParseRes struct {
	Status      *string
	Priority    *string
	Due         *string
	Tag         *string
	Created     *string
	FlagArgs    []string
	NonFlagArgs []string
}

func Parse(fs *flag.FlagSet, cmd string) *ParseRes {
	guide := struct {
		status   string
		priority string
		due      string
		tag      string
		created  string
	}{
		status:   "Status of the task (e.g., done, pending)",
		priority: "Priority of the task (e.g., high, low)",
		due:      "Due date of the task (e.g., 2025-08-06)",
		tag:      "Tag of the task (e.g., Project 01)",
		created:  "created date of the task (eg. 2025-08-01)",
	}
	status := fs.String("status", "", guide.status)
	priority := fs.String("priority", "", guide.priority)
	due := fs.String("due", "", guide.due)
	tag := fs.String("tag", "", guide.tag)
	created := fs.String("created", "", guide.created)

	// Shortcuts
	fs.StringVar(status, "s", *status, guide.status)
	fs.StringVar(priority, "p", *priority, guide.priority)
	fs.StringVar(due, "d", *due, guide.due)
	fs.StringVar(tag, "t", *tag, guide.tag)
	fs.StringVar(created, "c", *created, guide.created)

	// Custom usage function to include all flags
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: todo %s ... [text] [flags]\n", cmd)
		fmt.Fprintf(os.Stderr, "Flags:\n")
		fs.VisitAll(func(f *flag.Flag) {
			switch f.Name {
			case "status":
				fmt.Fprintf(os.Stderr, "  -s, --status\t\t%s\n", f.Usage)
			case "priority":
				fmt.Fprintf(os.Stderr, "  -p, --priority\t%s\n", f.Usage)
			case "due":
				fmt.Fprintf(os.Stderr, "  -d, --due\t\t%s\n", f.Usage)
			case "tag":
				fmt.Fprintf(os.Stderr, "  -t, --tag\t\t%s\n", f.Usage)
			case "created":
				fmt.Fprintf(os.Stderr, "  -c, --created\t\t%s\n", f.Usage)
			}
		})
	}
	var flagArgs, nonFlagArgs []string
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "--") {
			flagArgs = append(flagArgs, arg)
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") && !strings.HasPrefix(args[i+1], "--") && !strings.Contains(arg, "=") {
				flagArgs = append(flagArgs, args[i+1])
				i++
			}
		} else {
			nonFlagArgs = append(nonFlagArgs, arg)
		}
	}
	fs.Parse(flagArgs)

	return &ParseRes{
		Status:      status,
		Priority:    priority,
		Due:         due,
		Tag:         tag,
		Created:     created,
		FlagArgs:    flagArgs,
		NonFlagArgs: nonFlagArgs,
	}
}
