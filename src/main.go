package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/SierraSoftworks/Lithium/src/commands"
	"github.com/codegangsta/cli"
)

var version = "1.0.0-dev"

func main() {
	app := cli.NewApp()
	app.Name = "litmus"
	app.Usage = "Manage your Lithium licenses"
	app.Commands = commands.Commands

	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("error determining working directory: %s", err)
		os.Exit(1)
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "licensePath",
			Usage:  "the `path` under which licenses are stored",
			EnvVar: "LITHIUM_LICENSE_PATH",
			Value:  filepath.Join(cwd, "./.lithium"),
		},
	}

	app.Author = "Benjamin Pannell"
	app.Copyright = "Copyright © 2016 Sierra Softworks"
	app.Email = "admin@sierrasoftworks.com"
	app.Version = version

	app.Run(os.Args)
}
