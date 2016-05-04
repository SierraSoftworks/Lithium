package application

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/SierraSoftworks/Lithium/src/license"
	"github.com/codegangsta/cli"
)

func listCommand() cli.Command {
	return cli.Command{
		Name:  "list",
		Usage: "retrieve a list of applications you are in posession of certificates for",
		Action: func(c *cli.Context) error {
			products, err := productsList(c.GlobalString("licensePath"))
			if err != nil {
				return cli.NewExitError("could not read license directory", 1)
			}

			fmt.Printf("%-20s %-25s %-25s\n", "ID", "Name", "Organization")

			for _, p := range products {
				fmt.Printf("%-20s %-25s %-25s\n", p.ID, p.Name, p.Organization)
			}

			return nil
		},
	}
}

func productsList(licensePath string) ([]*license.Product, error) {
	productFiles, err := filepath.Glob(filepath.Join(licensePath, "*.json"))
	if err != nil {
		return nil, err
	}

	products := make([]*license.Product, len(productFiles))
	for i := 0; i < len(products); i++ {
		product, err := parseProductFile(productFiles[i])
		if err != nil {
			return nil, err
		}

		products[i] = product
	}

	return products, nil
}

func parseProductFile(file string) (*license.Product, error) {
	var product license.Product

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &product)
	if err != nil {
		return nil, err
	}

	return &product, nil
}
