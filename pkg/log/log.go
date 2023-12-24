package log

import (
	"fmt"
	"log"
	"os"

	"github.com/wulfixyt/anubis-monitors/pkg/tasks/structs"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

func Init() {
	buffer := os.Stdout

	InfoLogger = log.New(buffer, "", log.LstdFlags)
	WarningLogger = log.New(buffer, "", log.LstdFlags)
	ErrorLogger = log.New(buffer, "", log.LstdFlags)
}

func Format(task *structs.Task, text string, colour string) string {
	return fmt.Sprintf("- [%s][%s] - %s", task.Id, task.Site, text)
}

// Example to show how to use
//func example() {
//InfoLogger.Println("Starting the application...")
//InfoLogger.Println("Something noteworthy happened")
//WarningLogger.Println("There is something you should know about")
//ErrorLogger.Println("Something went wrong")
//}
