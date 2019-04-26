package license

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
)

// EncryptedPayloadKeyLabel is responsible for identifying an asymmetrically
// encrypted key used to perform symmetric encryption/decryption for encrypted
// payload objects.
const EncryptedPayloadKeyLabel = "Lithium Encrypted Payload Key"

// EncryptedPayload represents an encrypted license definition. It is encrypted
// using AES256 making use of the IV and Key provided. The Key itself is encrypted
// using RSA.
type EncryptedPayload struct {
	Data      []byte `json:"data"`
	Key       []byte `json:"key"`
	IV        []byte `json:"iv"`
	Algorithm string `json:"algorithm"`
}

// Encrypt will encrypt the provided data in a reversible manner. The data is
// first serialized using JSON, following which it is encrypted using a symmetric
// encryption algorithm, adopting a cryptographically random key and initialization
// vector. The vector is stored alongside the data, and the key is encrypted using
// an asymmetric algorithm and the provided public key.
// Only someone in posession of the corresponding private key will be able to decrypt
// the symmetric encryption key, and thereby decrypt the contents of the data.
func (p *EncryptedPayload) Encrypt(data interface{}, pubKey *rsa.PublicKey) error {
	rawData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	symmetricKey := make([]byte, 32)
	_, err = rand.Read(symmetricKey)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		return err
	}

	iv := make([]byte, block.BlockSize())
	_, err = rand.Read(iv)
	if err != nil {
		return err
	}

	encryptedData := make([]byte, len(rawData))

	encryptionStream := cipher.NewCFBEncrypter(block, iv)
	encryptionStream.XORKeyStream(encryptedData, rawData)

	asymmetricKey, err := rsa.EncryptOAEP(crypto.SHA256.New(), rand.Reader, pubKey, symmetricKey, []byte(EncryptedPayloadKeyLabel))
	if err != nil {
		return err
	}

	p.Algorithm = "aes256"
	p.Data = encryptedData
	p.IV = iv
	p.Key = asymmetricKey
	return nil
}

// Decrypt will transform an encrypted payload into the data variable
// you specify. This assumes that your provided private key matches the
// public key used to encrypt the symmetric key for the encrypted payload.
func (p *EncryptedPayload) Decrypt(data interface{}, privKey *rsa.PrivateKey) error {
	if p.Algorithm != "aes256" {
		return errors.New("unsupported encryption algorithm type, expected aes256")
	}

	symmetricKey, err := rsa.DecryptOAEP(crypto.SHA256.New(), rand.Reader, privKey, p.Key, []byte(EncryptedPayloadKeyLabel))
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(symmetricKey)
	if err != nil {
		return err
	}

	decryptedData := make([]byte, len(p.Data))

	decryptionStream := cipher.NewCFBDecrypter(block, p.IV)
	decryptionStream.XORKeyStream(decryptedData, p.Data)

	return json.Unmarshal(decryptedData, data)
}
