package commands

import (
	// Import the application commands list
	"github.com/SierraSoftworks/Lithium/src/commands/application"
	"github.com/codegangsta/cli"
)

// Commands is the list of registered commands for use within
// the lithium command line tool.
var Commands = cli.Commands{}

// RegisterCommand will register a new command line option
// for use within lithium's command line tool.
func RegisterCommand(c cli.Command) {
	Commands = append(Commands, c)
}

func init() {
	RegisterCommand(application.Command())
}
