package application

import (
	"github.com/codegangsta/cli"
)

func Command() cli.Command {
	return cli.Command{
		Name:      "application",
		ShortName: "app",
		HelpName:  "app",
		Aliases:   []string{"app"},
		Usage:     "manage configuration and licensing of an application",
		Subcommands: cli.Commands{
			listCommand(),
			newAppCommand(),
			newCertCommand(),
		},
	}
}
