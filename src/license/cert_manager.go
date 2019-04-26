package license

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

// CertManager is responsible for tracking the local node's certificates.
type CertManager struct {
	Path    string
	Product *Product
}

// NewCertManager is responsible for creating a new certificate manager
// instance. This instance will be responsible for creating, signing
// and persisting product certificates.
func NewCertManager(prod *Product) *CertManager {
	homeDir := os.ExpandEnv("$HOME")
	licenseFolder := filepath.Join(homeDir, ".lithium")

	return &CertManager{
		Path:    licenseFolder,
		Product: prod,
	}
}

// SetLocal will update the stored local certificate to match the
// certificate provided.
func (m *CertManager) SetLocal(cert *x509.Certificate) error {
	f, err := os.Create(m.getCertificateFilePath(fmt.Sprintf("%s.crt", m.Product.ID)))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(cert.Raw)
	return err
}

// GetLocal retrieves the certificate used to sign derivative
// licenses for the current server.
func (m *CertManager) GetLocal() (*x509.Certificate, error) {
	f, err := os.Open(m.getCertificateFilePath(fmt.Sprintf("%s.crt", m.Product.ID)))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}
	defer f.Close()

	certData, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certData)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// Sign is responsible for signing a provided certificate
// using the current product's certificate and private key. This then
// enables consumers of the resulting certificate to trace the authenticity
// of that certificate back to a single root certificate.
func (m *CertManager) Sign(csr *x509.Certificate, privKey interface{}) (*x509.Certificate, error) {
	ownCert, err := m.GetLocal()
	if err != nil {
		return nil, err
	}

	if csr == nil {
		return nil, errors.New("expected target certificate to exist")
	}

	if ownCert == nil {
		return nil, errors.New("no certificate available to sign request")
	}

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(0xffffffff))
	if err != nil {
		return nil, err
	}
	csr.SerialNumber = serialNumber

	certData, err := x509.CreateCertificate(rand.Reader, csr, ownCert, csr.PublicKey, privKey)
	if err != nil {
		return nil, err
	}

	cert, err := x509.ParseCertificate(certData)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

// Prepare is responsible for preparing an x509 certificate to match
// a specific license's constraints.
func (m *CertManager) Prepare(csr *x509.CertificateRequest, license *Data) *x509.Certificate {
	cert := x509.Certificate{
		Subject:               csr.Subject,
		DNSNames:              csr.DNSNames,
		Issuer:                m.getIssuer(),
		BasicConstraintsValid: true,
		MaxPathLen:            128,
		IsCA:                  false,
		NotBefore:             license.Meta.ActivatesOn,
		NotAfter:              license.Meta.ExpiresOn,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment,
		PublicKey:             csr.PublicKey,
		PublicKeyAlgorithm:    csr.PublicKeyAlgorithm,
	}

	if license.Meta.Pack != nil && len(license.Meta.Pack) > 0 {
		cert.KeyUsage = cert.KeyUsage | x509.KeyUsageCertSign
		cert.IsCA = true
	}

	cert.Subject.CommonName = fmt.Sprintf("%s (%s)", m.Product.Name, m.Product.ID)
	cert.Subject.SerialNumber = license.Meta.ID
	cert.Subject.Organization = []string{m.Product.Organization}
	cert.Subject.OrganizationalUnit = []string{"Lithium Licensing"}

	return &cert
}

// CreateRoot will create a new, self-signed, root certificate.
func (m *CertManager) CreateRoot(privKey *rsa.PrivateKey) (*x509.Certificate, error) {
	template := x509.Certificate{
		Subject: pkix.Name{
			CommonName:         fmt.Sprintf("%s (%s)", m.Product.Name, m.Product.ID),
			SerialNumber:       "Root Certificate",
			Organization:       []string{m.Product.Organization},
			OrganizationalUnit: []string{"Lithium Licensing"},
		},
		Issuer:                m.getIssuer(),
		BasicConstraintsValid: true,
		IsCA:               true,
		MaxPathLen:         128,
		SerialNumber:       big.NewInt(1),
		NotBefore:          time.Now(),
		NotAfter:           time.Now().Add(100 * 365 * 24 * time.Hour),
		KeyUsage:           x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature | x509.KeyUsageDataEncipherment | x509.KeyUsageKeyEncipherment | x509.KeyUsageContentCommitment,
		SignatureAlgorithm: x509.SHA256WithRSA,
		DNSNames:           []string{"localhost"},
		ExtKeyUsage:        []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	err := privKey.Validate()
	if err != nil {
		return nil, err
	}

	certData, err := x509.CreateCertificate(rand.Reader, &template, &template, privKey.Public(), privKey)
	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(certData)
}

func (m *CertManager) getIssuer() pkix.Name {
	return pkix.Name{
		CommonName:         "Sierra Softworks Lithium License Protocol",
		Organization:       []string{"Sierra Softworks"},
		OrganizationalUnit: []string{"Lithium Licensing"},
		Province:           []string{"Western Cape"},
		Country:            []string{"South Africa"},
	}
}

func (m *CertManager) getCertificateFilePath(file string) string {
	return filepath.Join(m.Path, file)
}
