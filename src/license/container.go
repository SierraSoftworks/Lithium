package license

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// LicenseKeyType is used to armour the Lithium license encryption key after it
// has been encrypted and encoded.
const LicenseKeyType = "LITHIUM LICENSE KEY"

// LicenseType is used to armour the Lithium license data when encrypted and encoded.
const LicenseType = "LITHIUM LICENSE"

// SignatureType is used to armour the Lithium license data's signature data.
const SignatureType = "LITHIUM SIGNATURE"

// CertificateType is used to armour the public certificates of the server
// chain which issued the license.
const CertificateType = "LITHIUM CERTIFICATE"

// EncryptedLicenseLabel is used to identify a key which is used for symmetric
// cryptographic operations, and which has been protected by an asymmetric
// encryption layer.
const EncryptedLicenseLabel = "Lithium License Encryption Key"

// Container represents
type Container struct {
	Payload      EncryptedPayload
	Signature    *Signature
	Certificates []*x509.Certificate
}

// Signature represents the signature of a bundle of data.
type Signature struct {
	Data      []byte
	Algorithm string
}

// EncodeContainer will encode a license container into its binary format.
func EncodeContainer(container *Container) ([]byte, error) {
	d := []byte{}

	d = append(d, pem.EncodeToMemory(&pem.Block{
		Type:    LicenseKeyType,
		Headers: map[string]string{},
		Bytes:   container.Payload.Key,
	})...)

	d = append(d, pem.EncodeToMemory(&pem.Block{
		Type: LicenseType,
		Headers: map[string]string{
			"algorithm": container.Payload.Algorithm,
			"iv":        base64.StdEncoding.EncodeToString(container.Payload.IV),
		},
		Bytes: container.Payload.Data,
	})...)

	if container.Signature == nil {
		return nil, errors.New("no signature has been provided for the license data")
	}

	if container.Signature.Algorithm == "" {
		return nil, errors.New("no signature algorithm has been specified")
	}

	d = append(d, pem.EncodeToMemory(&pem.Block{
		Type: SignatureType,
		Headers: map[string]string{
			"algorithm": container.Signature.Algorithm,
		},
		Bytes: container.Signature.Data,
	})...)

	for i := 0; i < len(container.Certificates); i++ {
		d = append(d, pem.EncodeToMemory(&pem.Block{
			Type:  CertificateType,
			Bytes: container.Certificates[i].Raw,
		})...)
	}

	return d, nil
}

// ParseContainer is responsible for parsing a license file into its structured representation.
func ParseContainer(licenseData []byte) (*Container, error) {
	c := Container{
		Certificates: make([]*x509.Certificate, 0),
	}
	var d []byte
	d = licenseData

	for {
		block, rest := pem.Decode(d)
		if block == nil {
			break
		}

		d = rest

		switch block.Type {
		case LicenseKeyType:
			c.Payload.Key = block.Bytes

		case LicenseType:
			c.Payload.Data = block.Bytes
			c.Payload.Algorithm = block.Headers["algorithm"]

			iv, err := base64.StdEncoding.DecodeString(block.Headers["iv"])
			if err != nil {
				return nil, err
			}

			c.Payload.IV = iv

		case SignatureType:
			algorithm, exists := block.Headers["algorithm"]
			if !exists {
				algorithm = "sha256"
			}

			c.Signature = &Signature{
				Algorithm: algorithm,
				Data:      block.Bytes,
			}

		case CertificateType:
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			}

			c.Certificates = append(c.Certificates, cert)
		}
	}

	if c.Signature == nil {
		return nil, errors.New("no license signature block was present in the container")
	}

	return &c, nil
}

// License will extract and decode the license data from the encrypted license block
// in this container.
func (c *Container) License(privKey *rsa.PrivateKey, rootCert *x509.Certificate) (*Data, error) {
	isValid, err := c.IsValid(rootCert)
	if !isValid {
		return nil, err
	}

	var d Data
	err = c.Payload.Decrypt(&d, privKey)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

// SetLicense will set the license data for this container. You will need to sign that
// data using your private key once you are finished. The public key you provide is used
// to ensure that the encryption key used to protect the license data is accessible by
// the intended client.
func (c *Container) SetLicense(data *Data, pubKey *rsa.PublicKey) error {
	return c.Payload.Encrypt(data, pubKey)
}

// Sign will populate the signature structure with the correct signature and algorithm
// for the license data provided.
func (c *Container) Sign(privateKey *rsa.PrivateKey, algorithm string) error {
	hash, err := hashByName(algorithm)
	if err != nil {
		return err
	}

	hashedData, err := computeHash(c.Payload.Data, hash)
	if err != nil {
		return err
	}

	signature, err := rsa.SignPSS(rand.Reader, privateKey, hash, hashedData, nil)
	if err != nil {
		return err
	}

	c.Signature = &Signature{
		Data:      signature,
		Algorithm: strings.ToLower(algorithm),
	}

	return nil
}

// IsValid is responsible for determining whether a container is valid by validating
// the signature of its data and the certification chain of that signature.
func (c *Container) IsValid(rootCertificate *x509.Certificate) (bool, error) {
	if len(c.Certificates) == 0 {
		return false, errors.New("expected at least one certificate to be present")
	}

	if !c.Certificates[0].Equal(rootCertificate) {
		return false, errors.New("expected first certificate in list to match known root")
	}

	for i := 1; i < len(c.Certificates); i++ {
		err := c.Certificates[i].CheckSignatureFrom(c.Certificates[i-1])
		if err != nil {
			return false, err
		}
	}

	hash, err := hashByName(c.Signature.Algorithm)
	if err != nil {
		return false, err
	}

	hashedData, err := computeHash(c.Payload.Data, hash)
	if err != nil {
		return false, err
	}

	parentCert := c.Certificates[len(c.Certificates)-1]
	switch parentCert.PublicKeyAlgorithm {
	case x509.RSA:
		err := rsa.VerifyPSS(parentCert.PublicKey.(*rsa.PublicKey), hash, hashedData, c.Signature.Data, nil)
		if err != nil {
			return false, errors.New("signature did not match the expected signed value")
		}

	default:
		return false, errors.New("unsupported public key algorithm for certificate, required RSA")
	}

	return true, nil
}

func computeHash(data []byte, algorithm crypto.Hash) ([]byte, error) {
	h := algorithm.New()
	_, err := h.Write(data)
	if err != nil {
		return nil, err
	}

	var hash []byte
	return h.Sum(hash), nil
}

func hashByName(algorithm string) (crypto.Hash, error) {
	switch strings.ToLower(algorithm) {
	case "sha1":
		return crypto.SHA1, nil
	case "sha256":
		return crypto.SHA256, nil
	case "sha512":
		return crypto.SHA512, nil
	default:
		return crypto.SHA256, fmt.Errorf("unsupported hash function '%s'", algorithm)
	}
}
