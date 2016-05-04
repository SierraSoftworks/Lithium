package application

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/SierraSoftworks/Lithium/src/license"
	"github.com/codegangsta/cli"
)

func newCertCommand() cli.Command {
	cwd, err := filepath.Abs("./")
	if err != nil {
		cwd = os.ExpandEnv("$HOME")
	}

	return cli.Command{
		Name:        "root",
		Usage:       "create a new root certificate for an application",
		Description: "This will create a new root certificate with the details of the application it will sign licenses for.",
		ArgsUsage:   "ID NAME",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "id",
				EnvVar: "APP_ID",
				Usage:  "the ID of the application you wish to create a certificate for",
			},
			cli.StringFlag{
				Name:   "name",
				EnvVar: "APP_NAME",
				Usage:  "the name of the application you wish to create a certificate for",
			},
			cli.StringFlag{
				Name:   "org",
				EnvVar: "APP_ORGANIZATION",
				Usage:  "the name of the organization who manages the application you're generating a certificate for",
			},
			cli.StringFlag{
				Name:   "path",
				EnvVar: "LITHIUM_LICENSE_PATH",
				Usage:  "the folder within which your Lithium licenses are stored",
				Value:  filepath.Join(cwd, "licenses"),
			},
			cli.IntFlag{
				Name:   "keySize",
				EnvVar: "LITHIUM_KEY_SIZE",
				Usage:  "the length of the secure key used for the certificate",
				Value:  4096,
			},
		},
		Action: func(c *cli.Context) error {
			id := c.String("id")
			name := c.String("name")
			org := c.String("org")
			path := c.String("path")

			if id == "" || name == "" || org == "" {
				return fmt.Errorf("expected you to provide a name and ID for the application as well as a name for your organization")
			}

			product := license.Product{
				ID:           id,
				Name:         name,
				Organization: org,
			}

			privKey, err := rsa.GenerateKey(rand.Reader, c.Int("keySize"))
			if err != nil {
				return err
			}

			cm := license.NewCertManager(&product)

			if path != "" {
				cm.Path = path
			}

			cert, err := cm.CreateRoot(privKey)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(filepath.Join(cm.Path, fmt.Sprintf("%s.key", product.ID)), pem.EncodeToMemory(&pem.Block{
				Type:  license.PrivateKeyType,
				Bytes: x509.MarshalPKCS1PrivateKey(privKey),
			}), os.ModePerm)

			if err != nil {
				return err
			}

			err = ioutil.WriteFile(filepath.Join(cm.Path, fmt.Sprintf("%s.crt", product.ID)), pem.EncodeToMemory(&pem.Block{
				Type:  license.CertificateType,
				Bytes: cert.Raw,
			}), os.ModePerm)

			if err != nil {
				return err
			}

			return nil
		},
	}
}
