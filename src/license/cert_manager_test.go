package license

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"
)

var testProduct = &Product{
	Name:         "Lithium Testing",
	ID:           "testing",
	Organization: "Sierra Softworks",
}

func TestCertManCreateRoot(t *testing.T) {
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
	cm := NewCertManager(testProduct)
	cm.Path = testPath

	key, err := km.GetPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	cert, err := cm.CreateRoot(key)
	if err != nil {
		t.Fatal(err)
	}

	if cert == nil {
		t.Fatal("expected certificate to be defined")
	}

	if cert.Subject.CommonName != "Lithium Testing (testing)" {
		t.Errorf("expected CN='Lithium Testing (testing)', got '%s'", cert.Subject.CommonName)
	}

	if cert.Subject.SerialNumber != "Root Certificate" {
		t.Errorf("expected SN='Root Certificate', got '%s'", cert.Subject.SerialNumber)
	}
}

func TestCertManPrepareCSR(t *testing.T) {
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
	cm := NewCertManager(testProduct)
	cm.Path = testPath

	key, err := km.GetPrivateKey()
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

	csrData, err := x509.CreateCertificateRequest(rand.Reader, &reqTemplate, key)
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

	if cert.Subject.SerialNumber != "test" {
		t.Errorf("expected SN='test', got '%s'", cert.Subject.SerialNumber)
	}
}

func TestCertManSignCert(t *testing.T) {
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
	cm := NewCertManager(testProduct)
	cm.Path = testPath

	key, err := km.GetPrivateKey()
	if err != nil {
		t.Fatal(err)
	}

	rootCert, err := cm.CreateRoot(key)
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

	key2, err := km.GetPrivateKey()
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

	csrData, err := x509.CreateCertificateRequest(rand.Reader, &reqTemplate, key2)
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

	signedCert, err := cm.Sign(cert, key)
	if err != nil {
		t.Fatal(err)
	}

	if signedCert == nil {
		t.Fatal("expected signed certificate to exist")
	}

	err = signedCert.CheckSignatureFrom(rootCert)
	if err != nil {
		t.Fatal(err)
	}
}
