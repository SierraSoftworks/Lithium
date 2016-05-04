package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/SierraSoftworks/Lithium/src/license"
	"github.com/codegangsta/cli"
)

func newAppCommand() cli.Command {
	return cli.Command{
		Name:        "create",
		Aliases:     []string{"c"},
		Usage:       "create a new application to accept licenses",
		ArgsUsage:   "ID NAME ORGANIZATION",
		Description: "This will create a new application description file which is used by the Lithium command line tool to track details about the application for licensing purposes.",
		Action: func(c *cli.Context) error {
			if c.NArg() < 3 {
				return errors.New("expected you to provide the ID, name and organization of the product")
			}

			args := c.Args()

			product := license.Product{
				ID:           args[0],
				Name:         args[1],
				Organization: args[2],
			}

			err := saveProduct(&product, c.GlobalString("licensePath"))
			if err != nil {
				return cli.NewExitError(err.Error(), 1)
			}

			return nil
		},
	}
}

func saveProduct(prod *license.Product, licensePath string) error {
	err := os.MkdirAll(licensePath, os.ModePerm|os.ModeDir)
	if err != nil {
		return err
	}

	data, err := json.Marshal(prod)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath.Join(licensePath, fmt.Sprintf("./%s.json", prod.ID)), data, os.ModePerm)
}
