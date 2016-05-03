package license

import (
	"encoding/pem"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNewKeyManager(t *testing.T) {
	KeySize = 1024
	machineCode := []byte("test")

	m := NewKeyManager(machineCode)
	if m == nil {
		t.Fatal("expected NewKeyManager to return a key manager")
	}

	if !reflect.DeepEqual(m.MachineCode, machineCode) {
		t.Errorf("expected machine code to be %#v but got %#v instead.", machineCode, m.MachineCode)
	}

	if m.Path == "" {
		t.Errorf("expected path to be a defined value")
	}
}

func TestGetPublicKey(t *testing.T) {
	KeySize = 1024
	machineCode := []byte("test")
	tempDir := os.TempDir()
	testPath, err := ioutil.TempDir(tempDir, "lithium")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testPath)

	m := NewKeyManager(machineCode)
	m.Path = testPath

	key, err := m.GetPublicKey()
	if err != nil {
		t.Fatal(err)
	}

	if key == nil {
		t.Fatalf("expected public key to be defined")
	}

	f, err := os.Open(filepath.Join(testPath, PublicKeyName))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		t.Fatal("expected key file to be a valid PEM block")
	}

	if block.Type != PublicKeyType {
		t.Errorf("expected PEM block to be of type '%s'", PublicKeyType)
	}
}

func TestPublicKeyPersistence(t *testing.T) {
	KeySize = 1024
	machineCode := []byte("test")
	tempDir := os.TempDir()
	testPath, err := ioutil.TempDir(tempDir, "lithium")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testPath)

	m := NewKeyManager(machineCode)
	m.Path = testPath

	key, err := m.GetPublicKey()
	if err != nil {
		t.Fatal(err)
	}

	key2, err := m.GetPublicKey()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(key, key2) {
		t.Error("expected to receive the same keys")
	}
}

func TestGetPrivateKey(t *testing.T) {
	KeySize = 1024
	machineCode := []byte("test")
	tempDir := os.TempDir()
	testPath, err := ioutil.TempDir(tempDir, "lithium")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testPath)

	m := NewKeyManager(machineCode)
	m.Path = testPath

	key, err := m.GetPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	if key == nil {
		t.Fatalf("expected private key to be defined")
	}

	err = key.Validate()
	if err != nil {
		t.Error("expected key to be valid,", err)
	}

	if key.D == nil {
		t.Errorf("expected key.D to be defined")
	}

	f, err := os.Open(filepath.Join(testPath, PrivateKeyName))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		t.Fatal("expected key file to be a valid PEM block")
	}

	if block.Type != PrivateKeyType {
		t.Errorf("expected PEM block to be of type '%s'", PrivateKeyType)
	}
}

func TestPrivateKeyPersistence(t *testing.T) {
	KeySize = 1024
	machineCode := []byte("test")
	tempDir := os.TempDir()
	testPath, err := ioutil.TempDir(tempDir, "lithium")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testPath)

	m := NewKeyManager(machineCode)
	m.Path = testPath

	key, err := m.GetPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	key2, err := m.GetPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(key, key2) {
		t.Error("expected to receive the same keys")
	}
}

func TestResetKeys(t *testing.T) {
	KeySize = 1024
	machineCode := []byte("test")
	tempDir := os.TempDir()
	testPath, err := ioutil.TempDir(tempDir, "lithium")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testPath)

	m := NewKeyManager(machineCode)
	m.Path = testPath

	key, err := m.GetPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	if key == nil {
		t.Fatalf("expected private key to be defined")
	}

	err = m.ResetKeypair()
	if err != nil {
		t.Fatal(err)
	}

	newKey, err := m.GetPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(key, newKey) {
		t.Error("expected a new public key to have been generated")
	}
}
