package license

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

// PrivateKeyType is used to armour the machine's private key when
// PEM encoded and encrypted.
const PrivateKeyType = "LITHIUM PRIVATE KEY"

// PrivateKeyName is the name of the private key file. It may be
// changed if you wish to enable multiple side-by-side installations
// with different keys. This is usually not necessary.
var PrivateKeyName = "machine"

// PublicKeyType is used to armour the machine's public key when
// PEM encoded.
const PublicKeyType = "LITHIUM PUBLIC KEY"

// PublicKeyName is the name of your public key file. It may be
// changed if you wish to enable multiple side-by-side installations
// with different keys. This is usually not necessary.
var PublicKeyName = "machine.pub"

// KeySize is the size of the RSA key used for license encryption
// and decryption. 2048 is probably secure enough for any general
// use, however you can reduce this to 1024 or increase to 4096
// to balance generation speed and security.
var KeySize = 2048

// KeyManager provides a set of tools for accessing your machine's
// local key. This is tied to your machine and allows you to decrypt
// your license packs.
type KeyManager struct {
	MachineCode []byte
	Path        string
}

// NewKeyManager returns a new KeyManager for your local machine using
// the given machine code to encrypt and decrypt local machine keys.
func NewKeyManager(machineCode []byte) *KeyManager {
	homeDir := os.ExpandEnv("$HOME")
	licenseFolder := filepath.Join(homeDir, ".lithium")

	return &KeyManager{
		MachineCode: machineCode,
		Path:        licenseFolder,
	}
}

// GetPublicKey retrieves the public key for your local machine. This
// is used by upstream servers to identify and encrypt keys for your
// machine.
func (m *KeyManager) GetPublicKey() (*rsa.PublicKey, error) {
	err := m.ensureKeypair()
	if err != nil {
		return nil, err
	}

	data, err := m.readKeyFile(PublicKeyName)
	if err != nil {
		return nil, err
	}

	pubBlock, _ := pem.Decode(data)
	if pubBlock == nil {
		return nil, errors.New("machine key was not a valid PEM block")
	}

	if pubBlock.Type != PublicKeyType {
		return nil, errors.New("machine key was not of the correct type")
	}

	pub, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub.(type) {
	case *rsa.PublicKey:
		return pub.(*rsa.PublicKey), nil
	default:
		return nil, errors.New("only ECDSA public keys supported")
	}
}

// GetPrivateKey retrieves the private key for your local machine. This
// is used to decrypt license packs and sign child license files for
// later verification.
func (m *KeyManager) GetPrivateKey() (*rsa.PrivateKey, error) {
	err := m.ensureKeypair()
	if err != nil {
		return nil, err
	}

	data, err := m.readKeyFile(PrivateKeyName)
	if err != nil {
		return nil, err
	}

	privBlock, _ := pem.Decode(data)
	if privBlock == nil {
		return nil, errors.New("machine key was not a valid PEM block")
	}

	if privBlock.Type != PrivateKeyType {
		return nil, errors.New("machine key was not of the correct type")
	}

	if !x509.IsEncryptedPEMBlock(privBlock) {
		return nil, errors.New("expected machine key to be encrypted")
	}

	privData, err := x509.DecryptPEMBlock(privBlock, m.MachineCode)
	if err != nil {
		return nil, err
	}

	priv, err := x509.ParsePKCS1PrivateKey(privData)
	if err != nil {
		return nil, err
	}

	return priv, nil
}

// ResetKeypair will generate a new keypair for this machine, replacing
// the existing keypair and invalidating any licenses which were created
// for it.
func (m *KeyManager) ResetKeypair() error {
	return m.createKeypair()
}

func (m *KeyManager) ensureKeypair() error {
	if !m.keyFileExists(PrivateKeyName) || !m.keyFileExists(PublicKeyName) {
		return m.createKeypair()
	}

	return nil
}

func (m *KeyManager) keyFileExists(file string) bool {
	f, err := os.Open(m.getLicenseFilePath(file))
	if err == nil {
		defer f.Close()
	}

	if err != nil && os.IsNotExist(err) {
		return false
	} else if err != nil {
		return false
	}

	return true
}

func (m *KeyManager) createKeypair() error {
	priv, err := rsa.GenerateKey(rand.Reader, KeySize)
	if err != nil {
		return err
	}

	derPrivateKey := x509.MarshalPKCS1PrivateKey(priv)

	encryptedPrivateKey, err := x509.EncryptPEMBlock(
		rand.Reader,
		PrivateKeyType,
		derPrivateKey,
		m.MachineCode,
		x509.PEMCipherAES256,
	)

	err = m.writeKeyFile(PrivateKeyName, pem.EncodeToMemory(encryptedPrivateKey))
	if err != nil {
		return err
	}

	pub := priv.Public()
	pubKeyData, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return err
	}

	pubKeyBytes := pem.EncodeToMemory(&pem.Block{
		Type:  PublicKeyType,
		Bytes: pubKeyData,
	})

	return m.writeKeyFile(PublicKeyName, pubKeyBytes)
}

func (m *KeyManager) writeKeyFile(file string, data []byte) error {
	f, err := os.Create(m.getLicenseFilePath(file))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}

func (m *KeyManager) readKeyFile(file string) ([]byte, error) {
	f, err := os.Open(m.getLicenseFilePath(file))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ioutil.ReadAll(f)
}

func (m *KeyManager) getLicenseFilePath(file string) string {
	return filepath.Join(m.Path, file)
}
