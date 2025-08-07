package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/benpsk/todo/cmd/service"
)

func (h *handler) delete() {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	fs.Parse(os.Args[2:])
	ids := fs.Args()
	if len(ids) == 0 {
		fmt.Println("usage: todo delete <id>...")
		os.Exit(1)
	}
  idList := service.ValidateIds(ids)
	if err := h.deleteTodo(idList); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted id:", ids)
}

func (h *handler) deleteTodo(ids []int) error {
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	query := "DELETE FROM todos WHERE id IN (" + strings.Join(placeholders, ",") + ")"

	_, err := h.db.Exec(query, args...)
	return err
}
