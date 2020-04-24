package internal

import (
	"github.com/fatih/color"
	"os/exec"
	"strings"
)

var (
	commandLogger *Logger
)

// SetCommandLogger is to set default command logger.
func SetCommandLogger(logger *Logger) {
	commandLogger = logger
}

// CommandExec executes command
func CommandExec(command string, args []string) error {
	cmd := exec.Command(command, args...)
	if err := cmd.Run(); err != nil {
		commandLogger.Error(color.RedString("[FAIL-CMD] %s %s [Reason]: %s",
			command, strings.Join(args, " ")), err.Error())
		return err
	} else {
		commandLogger.Info(color.GreenString("[SUCCESS-CMD] %s %s", command, strings.Join(args, " ")))
	}
	return nil
}

func init() {
	commandLogger, _ = NewLogger("")
}
