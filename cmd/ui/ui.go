package ui

import "fmt"

func Usage() {
	fmt.Println(`Todo CLI - Task Management Tool

Usage:
  todo <command> [options]

Commands:
  add       Add a new task
  list      List tasks
  update    Update existing tasks
  delete    Delete tasks

Examples:

  Add task:
    todo add "new task" --priority=high --status=processing --due=fri --tag=project1,ui
    todo add "new task" -p high -s processing -d fri-18:00 -t project1

  List tasks: [filter by last 7 due days]
    todo list --status=done --priority=high --due=wed-20:19 --created=wed --find=task1
    todo ls -s done -p high -d wed-20:19 -c wed -f task1

  Update tasks:
    todo update 1 2 3 --status=done --priority=high --due=wed-20:19
    todo update -s done -p high -d wed-20:19

  Delete tasks:
    todo delete 1 2 3

Options:
  -p, --priority   Set task priority (low|medium|high)
  -s, --status     Set task status (pending|processing|done)
  -d, --due        Set due date (e.g. 2025, 2025-01, fri, 2025-01-01)
  -t, --tag        Add one or more tags (eg. "p1,ui")
  -c, --created    Filter by creation date (eg. 2025, 2025-01, fri, 2025-01-01) 
  -f, --find       Search for keyword in task 

Enjoy!`)
}
