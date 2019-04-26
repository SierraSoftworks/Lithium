package license

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"
)

var demoLicenseData = []byte(demoData)

func TestSignContainer(t *testing.T) {
	KeySize = 1024
	machineCode := []byte("test")
	tempDir := os.TempDir()
	testPath, err := ioutil.TempDir(tempDir, "lithium")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testPath)

	km := NewKeyManager(machineCode)
	km.Path = testPath

	key, err := km.GetPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	cm := NewCertManager(testProduct)
	cert, err := cm.CreateRoot(key)
	if err != nil {
		t.Fatal(err)
	}

	c := Container{
		Certificates: []*x509.Certificate{cert},
	}

	err = c.SetLicense(&Data{
		Meta: &Metadata{
			ID:          "0",
			ActivatesOn: time.Now(),
			ExpiresOn:   time.Now(),
		},
		Payload: map[string]interface{}{
			"x": 1,
		},
	}, &key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	err = c.Sign(key, "sha256")
	if err != nil {
		t.Fatal(err)
	}

	if c.Signature == nil {
		t.Fatal("expected signature to be set")
	}

	if c.Signature.Algorithm != "sha256" {
		t.Errorf("expected signature algorithm to be 'sha256', got '%s'", c.Signature.Algorithm)
	}

	isValid, err := c.IsValid(cert)
	if !isValid {
		t.Error("expected certificate to be valid: ", err)
	}
}

func TestEncodeContainer(t *testing.T) {
	KeySize = 1024
	machineCode := []byte("test")
	tempDir := os.TempDir()
	testPath, err := ioutil.TempDir(tempDir, "lithium")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testPath)

	km := NewKeyManager(machineCode)
	km.Path = testPath

	rootKey, err := km.GetPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	cm := NewCertManager(testProduct)
	cm.Path = testPath
	rootCert, err := cm.CreateRoot(rootKey)
	if err != nil {
		t.Fatal(err)
	}

	err = cm.SetLocal(rootCert)
	if err != nil {
		t.Fatal(err)
	}

	err = km.ResetKeypair()
	if err != nil {
		t.Fatal(err)
	}

	childKey, err := km.GetPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	reqTemplate := x509.CertificateRequest{
		Subject:            pkix.Name{},
		Attributes:         []pkix.AttributeTypeAndValueSET{},
		SignatureAlgorithm: x509.SHA256WithRSA,
		Extensions:         []pkix.Extension{},
		DNSNames:           []string{"localhost"},
		EmailAddresses:     []string{},
		IPAddresses:        []net.IP{},
	}

	csrData, err := x509.CreateCertificateRequest(rand.Reader, &reqTemplate, childKey)
	if err != nil {
		t.Fatal(err)
	}

	csr, err := x509.ParseCertificateRequest(csrData)
	if err != nil {
		t.Fatal(err)
	}

	testTime, err := time.Parse("01/02/2006 15:04:05", "01/01/1970 00:00:00")
	if err != nil {
		t.Fatal(err)
	}

	cert := cm.Prepare(csr, &Data{
		Meta: &Metadata{
			ID:          "test",
			ActivatesOn: testTime,
			ExpiresOn:   testTime,
		},
	})

	signedCert, err := cm.Sign(cert, rootKey)
	if err != nil {
		t.Fatal(err)
	}

	c := Container{
		Certificates: []*x509.Certificate{rootCert, signedCert},
	}

	err = c.SetLicense(&Data{
		Meta: &Metadata{
			ID:          "0",
			ActivatesOn: time.Now(),
			ExpiresOn:   time.Now(),
		},
		Payload: map[string]interface{}{
			"x": 1,
		},
	}, &childKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	err = c.Sign(childKey, "sha256")
	if err != nil {
		t.Fatal(err)
	}

	encoded, err := EncodeContainer(&c)
	if err != nil {
		t.Fatal(err)
	}

	expectedBlocks := []string{
		LicenseKeyType,
		LicenseType,
		SignatureType,
		CertificateType,
		CertificateType,
	}

	data := encoded

	for i := 0; i < len(expectedBlocks); i++ {
		block, rest := pem.Decode(data)
		if block == nil {
			t.Fatal("expected another PEM block")
		}

		data = rest
		if block.Type != expectedBlocks[i] {
			t.Errorf("expected block type to be '%s', got '%s'", expectedBlocks[i], block.Type)
		}
	}

}
